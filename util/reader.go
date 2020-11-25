package util

import (
	"encoding/binary"
	"io"
	"os"
	"time"
)

func ReadMultiple(reader io.Reader, items ...interface{}) error {
	for _, item := range items {
		if err := binary.Read(reader, binary.BigEndian, item); err != nil {
			return err
		}
	}

	return nil
}

func ReadUint16(reader io.Reader) (uint16, error) {
	var number uint16
	if err := binary.Read(reader, binary.BigEndian, &number); err != nil {
		return 0, err
	}

	return number, nil
}

func ReadUint32(reader io.Reader) (uint32, error) {
	var number uint32
	if err := binary.Read(reader, binary.BigEndian, &number); err != nil {
		return 0, err
	}

	return number, nil
}

func ReadTime(reader io.Reader) (time.Time, error) {
	changedTimeSeconds, err := ReadUint32(reader)
	if err != nil {
		return time.Time{}, err
	}

	changedTimeNano, err := ReadUint32(reader)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(changedTimeSeconds), int64(changedTimeNano)), nil
}

func ReadFileMode(reader io.Reader) (os.FileMode, error) {
	mode, err := ReadUint32(reader)
	if err != nil {
		return 0, err
	}

	return os.FileMode(mode), nil
}

func ReadString(reader io.Reader, length uint32) (string, error) {
	oid := make([]byte, length)
	if _, err := io.ReadFull(reader, oid); err != nil {
		return "", err
	}

	return string(oid), nil
}
