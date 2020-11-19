package repo

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"github.com/furisto/gog/util"
	"io"
	"os"
	"path/filepath"
	"sort"
)

var dirCacheMarker = []byte("DIRC")

var (
	ErrEntryDoesNotExist = errors.New("index entry does not exist")
	ErrCorruptIndex      = errors.New("index file is corrupt")
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
	defer indexFile.Close()

	bufReader := bufio.NewReader(indexFile)

	var dirCache []byte
	if _, err := io.ReadFull(bufReader, dirCache); err != nil {
		return nil, err
	}

	if !bytes.Equal(dirCache, dirCacheMarker) {
		return nil, ErrCorruptIndex
	}

	hasher := sha1.New()
	hasher.Write(dirCache)

	var version uint32
	if err := binary.Read(bufReader, binary.BigEndian, &version); err != nil {
		return nil, err
	}

	var entryLength uint32
	if err := binary.Read(bufReader, binary.BigEndian, &entryLength); err != nil {
		return nil, err
	}

	const offset int8 = 62
	var entries map[string]*IndexEntry

	for i := uint32(0); i < entryLength; i++ {
		var entry IndexEntry

		entry.ChangedTime, err = util.ReadTime(bufReader)
		if err != nil {
			return nil, err
		}

		entry.ModifiedTime, err = util.ReadTime(bufReader)
		if err != nil {
			return nil, err
		}

		entry.DeviceId, err = util.ReadUint32(bufReader)
		if err != nil {
			return nil, err
		}

		entry.Inode, err = util.ReadUint32(bufReader)
		if err != nil {
			return nil, err
		}

		entry.Mode, err = util.ReadFileMode(bufReader)
		if err != nil {
			return nil, err
		}

		entry.UID, err = util.ReadUint32(bufReader)
		if err != nil {
			return nil, err
		}

		entry.GID, err = util.ReadUint32(bufReader)
		if err != nil {
			return nil, err
		}

		entry.FileSize, err = util.ReadUint32(bufReader)
		if err != nil {
			return nil, err
		}

		entry.OID, err = util.ReadString(bufReader, 20)
		if err != nil {
			return nil, err
		}

		flagValue, err := util.ReadUint16(bufReader)
		if err != nil {
			return nil, err
		}
		entry.Flags = Flags(flagValue)

		entry.Path, err = bufReader.ReadString(byte(0))
		if err != nil {
			return nil, err
		}
		entry.Path = entry.Path[:len(entry.Path)-1]

		paddingLength := 62 + len(entry.Path) + 1%8
		if paddingLength > 0 {
			throwaway := make([]byte, 8-paddingLength)
			io.ReadFull(bufReader, throwaway)
		}

		entries[entry.Path] = &entry
	}

	return &Index{
		version: version,
		entries: entries,
	}, nil
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

func (ix *Index) Find(path string) *IndexEntry {
	entry, ok := ix.entries[path]
	if !ok {
		return nil
	}

	return entry
}

func (ix *Index) Flush() error {
	indexFile, err := os.Create(filepath.Join(ix.gitDir, "index"))
	if err != nil {
		return err
	}
	defer indexFile.Close()
	return ix.EncodeIndex(indexFile)
}