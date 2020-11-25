package repo

import (
	"fmt"
	"testing"
	"time"
)

func TestDecodeIndex(t *testing.T) {
	ix, err := DecodeIndex("./testdata/index")
	if err != nil {
		t.Errorf("error occured during decoding of index: %v", err)
		return
	}

	if len(ix.Entries()) != 4 {
		t.Errorf("")
	}

	expected := []*IndexEntry{
		{
			ChangedTime:   time.Unix(0x5FBD0CA0, 0x08C09F38),
			ModifiedTime:  time.Unix(0x5FB42031, 0x30A0FE94),
			DeviceId:      0,
			Inode:         0,
			Mode:          0o100644,
			UID:           0,
			GID:           0,
			FileSize:      2,
			OID:           "857f065e4154176c98f4274d223066861e8e3d80",
			Flags:         3,
			ExtendedFlags: 0,
			Path:          "0/0",
		},
		{
			ChangedTime:   time.Unix(0x5FBD0CA0, 0x08D78ACC),
			ModifiedTime:  time.Unix(0x5FB42031, 0x30A0FE94),
			DeviceId:      0,
			Inode:         0,
			Mode:          0o100644,
			UID:           0,
			GID:           0,
			FileSize:      2,
			OID:           "a616ad491b179d212b8a78f2067b361980fffc54",
			Flags:         3,
			ExtendedFlags: 0,
			Path:          "0/1",
		},
		{
			ChangedTime:   time.Unix(0x5FBD0CA0, 0x90548F4),
			ModifiedTime:  time.Unix(0x5FB42031, 0x30B04458),
			DeviceId:      0,
			Inode:         0,
			Mode:          0o100644,
			UID:           0,
			GID:           0,
			FileSize:      2,
			OID:           "9a037142aa3c1b4c490e1a38251620f113465330",
			Flags:         3,
			ExtendedFlags: 0,
			Path:          "1/0",
		},
		{
			ChangedTime:   time.Unix(0x5FBD0CA0, 0x09149494),
			ModifiedTime:  time.Unix(0x5FB42031, 0x30B04458),
			DeviceId:      0,
			Inode:         0,
			Mode:          0o100644,
			UID:           0,
			GID:           0,
			FileSize:      2,
			OID:           "9d607966b721abde8931ddd052181fae905db503",
			Flags:         3,
			ExtendedFlags: 0,
			Path:          "1/1",
		},
	}

	entries := ix.Entries()
	for i, expct := range expected {

		t1 := entries[i].ModifiedTime.Unix()
		t2 := entries[i].ModifiedTime.UnixNano()
		t3 := entries[i].ChangedTime.Unix()
		t4 := entries[i].ChangedTime.UnixNano()
		fmt.Printf("%v %v %v %v", t1, t2, t3, t4)
		if !expct.Equals(entries[i]) {
			t.Errorf("expected %+v but got %+v", expct, entries[i])
		}
	}
}
