package util

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

func CloseFile(file *os.File, err error) {
	if cerr := file.Close(); cerr != nil && err == nil {
		err = cerr
	}
}

func ReadAndDecompressFile(path string) ([]byte, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("cannot decompress directory")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r, err := zlib.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, r); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
