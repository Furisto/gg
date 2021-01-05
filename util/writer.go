package util

import (
	"encoding/binary"
	"io"
)

func WriteMultiple(writer io.Writer, items []interface{}) error {
	for _, item := range items {
		if err := binary.Write(writer, binary.BigEndian, item); err != nil {
			return err
		}
	}

	return nil
}
