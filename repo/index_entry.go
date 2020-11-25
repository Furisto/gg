package repo

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/furisto/gog/util"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type IndexEntry struct {
	// Last time the file's metadata changed
	ChangedTime time.Time
	// Last time the file's data changed
	ModifiedTime time.Time
	//
	DeviceId uint32
	//
	Inode uint32
	//
	Mode os.FileMode
	//
	UID uint32
	//
	GID uint32
	//
	FileSize uint32
	//
	OID string
	//
	Flags Flags
	//
	ExtendedFlags uint16
	//
	Path string
}

func (ie *IndexEntry) Match(stat os.FileInfo) bool {
	if !ie.ModifiedTime.Equal(stat.ModTime()) {
		return false
	}

	if !ie.ChangedTime.Equal(stat.ModTime()) {
		return false
	}

	return true
}

func (ie *IndexEntry) Equals(other *IndexEntry) bool {
	if ie.OID != other.OID {
		return false
	}

	if ie.Path != other.Path {
		return false
	}

	if !ie.ChangedTime.Equal(other.ChangedTime) {
		return false
	}

	if !ie.ModifiedTime.Equal(other.ModifiedTime) {
		return false
	}

	if ie.FileSize != other.FileSize {
		return false
	}

	if ie.DeviceId != other.DeviceId {
		return false
	}

	if ie.Inode != other.Inode {
		return false
	}

	if ie.Mode != other.Mode {
		return false
	}

	if ie.UID != other.UID {
		return false
	}

	if ie.GID != other.GID {
		return false
	}

	if ie.Flags != other.Flags {
		return false
	}

	if ie.ExtendedFlags != other.ExtendedFlags {
		return false
	}

	return true
}

func (ie *IndexEntry) IsOurs() bool {
	return ie.Flags.StageType() == Ours
}

func (ie *IndexEntry) IsTheirs() bool {
	return ie.Flags.StageType() == Theirs
}

func (ie *IndexEntry) IsBase() bool {
	return ie.Flags.StageType() == Base
}

func (ie *IndexEntry) IsRegular() bool {
	return ie.Flags.StageType() == Regular
}

func (ie *IndexEntry) Encode(writer io.Writer) error {
	fields := []interface{}{
		ie.ChangedTime.Unix(),
		ie.ChangedTime.UnixNano(),
		ie.ModifiedTime.Unix(),
		ie.ModifiedTime.UnixNano(),
		ie.DeviceId,
		ie.Inode,
		ie.Mode,
		ie.UID,
		ie.GID,
		ie.FileSize,
	}

	if err := util.WriteMultiple(writer, binary.BigEndian, fields); err != nil {
		return err
	}

	oidBytes, err := hex.DecodeString(ie.OID)
	if err != nil {
		return err
	}

	if _, err := writer.Write(oidBytes); err != nil {
		return err
	}

	if err := binary.Write(writer, binary.BigEndian, ie.Flags); err != nil {
		return err
	}

	if _, err := writer.Write([]byte(ie.Path)); err != nil {
		return err
	}

	if _, err := writer.Write([]byte{0}); err != nil {
		return err
	}

	paddingLength := (indexEntryOffset + len(ie.Path) + 1) % 8
	if paddingLength > 0 {
		padding := make([]byte, 8-paddingLength)
		if _, err := writer.Write(padding); err != nil {
			return err
		}
	}

	return nil
}

func newIndexEntryFromFile(oid, path, workingDir string) (*IndexEntry, error) {
	if filepath.IsAbs(path) {
		path = strings.TrimPrefix(path, workingDir+string(os.PathSeparator))
	}

	stat, err := os.Stat(filepath.Join(workingDir, path))
	if err != nil {
		return nil, err
	}

	return &IndexEntry{
		ChangedTime:   stat.ModTime(),
		ModifiedTime:  stat.ModTime(),
		DeviceId:      0,
		Inode:         0,
		Mode:          0o100644,
		UID:           0,
		GID:           0,
		FileSize:      uint32(stat.Size()),
		OID:           oid,
		Flags:         Flags(0).WithLength(uint16(len(path))),
		ExtendedFlags: 0,
		Path:          filepath.ToSlash(path),
	}, nil
}

type indexEntrySorter []*IndexEntry

func (ies indexEntrySorter) Len() int {
	return len(ies)
}

func (ies indexEntrySorter) Less(i int, j int) bool {
	return ies[i].Path < ies[j].Path
}

func (ies indexEntrySorter) Swap(i int, j int) {
	ies[i], ies[j] = ies[j], ies[i]
}

type StageType uint8

const (
	Regular StageType = 0
	Base              = 1
	Ours              = 2
	Theirs            = 3
)

type Flags uint16

func (f Flags) AssumeValid() bool {
	return f&0x8000 != 0
}

func (f Flags) WithAssumeValid(valid bool) Flags {
	if valid {
		return f | 0x8000
	}

	return f & 0x7fff
}

func (f Flags) Extended() bool {
	return f&0x4000 != 0
}

func (f Flags) WithExtended(extended bool) Flags {
	if extended {
		return f | 0x4000
	}

	return f & 0xbfff
}

func (f Flags) Length() int16 {
	return int16(f & 0x0fff)
}

func (f Flags) WithLength(length uint16) Flags {
	if length > 0x2ff {
		length = 0x2ff
	}

	return f&0xf000 | Flags(length)
}

func (f Flags) StageType() StageType {
	return StageType(f & 0x3000)
}

func (f Flags) WithStageType(stage StageType) Flags {
	switch stage {
	case Regular:
		return f & 0xcfff
	case Base:
		return f&0xcfff | 0xdfff
	case Ours:
		return f&0xcfff | 0xefff
	case Theirs:
		return f&0xcfff | 0xffff
	default:
		panic("unknown stage type")
	}
}
