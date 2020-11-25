package repo

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/furisto/gog/util"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var dirCacheMarker = []byte("DIRC")

const indexEntryOffset = 62

var (
	ErrEntryDoesNotExist       = errors.New("index entry does not exist")
	ErrCorruptIndex            = errors.New("index file is corrupt")
	ErrUnsupportedIndexVersion = errors.New("index version not supported")
)

type Index struct {
	workingDir string
	gitDir     string
	version    uint32
	entries    map[string]*IndexEntry
	store      io.ReadWriteCloser
}

func NewIndex(workingDir, gitDir string) *Index {
	return &Index{
		workingDir: workingDir,
		gitDir:     gitDir,
		version:    2,
		entries:    make(map[string]*IndexEntry),
	}
}

func DecodeIndex(path string) (*Index, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	indexFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer util.CloseFile(indexFile, err)

	hasher := sha1.New()
	reader := io.TeeReader(bufio.NewReader(indexFile), hasher)

	version, entryLength, err := readHeader(reader)
	if err != nil {
		return nil, err
	}

	entries, err := readEntries(reader, entryLength)
	if err != nil {
		return nil, err
	}

	if err := verifyFooter(reader, hasher.Sum(nil)); err != nil {
		return nil, err
	}

	return &Index{
		version: version,
		entries: entries,
	}, nil
}

func readHeader(reader io.Reader) (version uint32, entryLength uint32, err error) {
	dirCache := make([]byte, 4)
	if _, err := io.ReadFull(reader, dirCache); err != nil {
		return 0, 0, err
	}

	if !bytes.Equal(dirCache, dirCacheMarker) {
		return 0, 0, ErrCorruptIndex
	}

	if err := binary.Read(reader, binary.BigEndian, &version); err != nil {
		return 0, 0, err
	}

	if version != 2 {
		return 0, 0, ErrUnsupportedIndexVersion
	}

	if err := binary.Read(reader, binary.BigEndian, &entryLength); err != nil {
		return 0, 0, err
	}

	return version, entryLength, nil
}

func readEntries(reader io.Reader, entryLength uint32) (map[string]*IndexEntry, error) {
	entries := make(map[string]*IndexEntry)

	for i := uint32(0); i < entryLength; i++ {
		var entry IndexEntry
		var csec, cnano, msec, mnano uint32

		fields := []interface{}{
			&csec,
			&cnano,
			&msec,
			&mnano,
			&entry.DeviceId,
			&entry.Inode,
			&entry.Mode,
			&entry.UID,
			&entry.GID,
			&entry.FileSize,
		}

		if err := util.ReadMultiple(reader, fields...); err != nil {
			return nil, err
		}

		entry.ChangedTime = time.Unix(int64(csec), int64(cnano))
		entry.ModifiedTime = time.Unix(int64(msec), int64(mnano))

		oidBytes := make([]byte, 20)
		if _, err := io.ReadFull(reader, oidBytes); err != nil {
			return nil, err
		}

		entry.OID = hex.EncodeToString(oidBytes)

		flagValue, err := util.ReadUint16(reader)
		if err != nil {
			return nil, err
		}
		entry.Flags = Flags(flagValue)

		pathBytes := make([]byte, entry.Flags.Length()+1)
		if _, err := io.ReadFull(reader, pathBytes); err != nil {
			return nil, err
		}
		entry.Path = string(pathBytes[:len(pathBytes)-1])
		entries[entry.Path] = &entry

		if err = discardPadding(reader, &entry); err != nil {
			return nil, err
		}
	}

	return entries, nil
}

func discardPadding(reader io.Reader, entry *IndexEntry) error {
	paddingInverse := (indexEntryOffset + entry.Flags.Length() + 1) % 8
	if paddingInverse > 0 {
		discard := make([]byte, 8-paddingInverse)
		if _, err := io.ReadFull(reader, discard); err != nil {
			return err
		}
	}

	return nil
}

func verifyFooter(reader io.Reader, hash []byte) error {
	readHash := make([]byte, 20)
	if _, err := io.ReadFull(reader, readHash); err != nil {
		return err
	}

	if !bytes.Equal(readHash, hash) {
		return ErrCorruptIndex
	}

	return nil
}

func (ix *Index) EncodeIndex(writer io.Writer) error {
	hasher := sha1.New()
	mw := io.MultiWriter(writer, hasher)

	if err := ix.writeHeader(mw); err != nil {
		return err
	}

	if err := ix.writeEntries(mw); err != nil {
		return err
	}
	if err := ix.writeFooter(mw, hasher.Sum(nil)); err != nil {
		return err
	}

	return nil
}

func (ix *Index) writeHeader(mw io.Writer) error {
	if _, err := mw.Write(dirCacheMarker); err != nil {
		return err
	}

	if err := binary.Write(mw, binary.BigEndian, ix.version); err != nil {
		return err
	}

	if err := binary.Write(mw, binary.BigEndian, uint32(len(ix.entries))); err != nil {
		return err
	}

	return nil
}

func (ix *Index) writeEntries(writer io.Writer) error {
	for _, entry := range ix.Entries() {
		if err := entry.Encode(writer); err != nil {
			return err
		}
	}

	return nil
}

func (ix *Index) writeFooter(writer io.Writer, hash []byte) error {
	if _, err := writer.Write(hash); err != nil {
		return err
	}

	return nil
}

func (ix *Index) Version() uint32 {
	return ix.version
}

func (ix *Index) Entries() []*IndexEntry {
	entries := make([]*IndexEntry, len(ix.entries))
	j := 0
	for _, v := range ix.entries {
		entries[j] = v
		j++
	}
	sort.Sort(indexEntrySorter(entries))

	return entries
}

func (ix *Index) Set(oid, path string) error {
	entry, err := newIndexEntryFromFile(oid, path, ix.workingDir)
	if err != nil {
		return err
	}
	ix.entries[path] = entry
	return nil
}

func (ix *Index) Delete(path string) {
	delete(ix.entries, path)
}

func (ix *Index) Find(path string) (*IndexEntry, error) {
	entry, ok := ix.entries[path]
	if !ok {
		return nil, ErrEntryDoesNotExist
	}

	return entry, nil
}

func (ix *Index) Flush() (err error) {
	indexFile, err := os.Create(filepath.Join(ix.gitDir, "index"))
	if err != nil {
		return err
	}
	defer func() {
		cerr := indexFile.Close()
		if err == nil {
			err = cerr
		}
	}()

	if err := ix.EncodeIndex(indexFile); err != nil {
		return err
	}

	return indexFile.Sync()
}
