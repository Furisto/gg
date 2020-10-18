package storage

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ObjectStore interface {
	Get(oid string)  ([]byte, error)
	Put(oid string, data []byte) error
	Stat(oid string) (bool, error)
	Find(prefix string) ([]string, error)
}

type FilesystemStore struct {
	location string
}

func NewFsStore(path string) *FilesystemStore {
	path = filepath.Join(path, "objects")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0644)
	}

	return &FilesystemStore{
		location: path,
	}
}

func(store *FilesystemStore) Get(oid string) ([]byte, error) {
	if err:= checkObjectId(oid); err != nil {
		return nil, err
	}

	objectPath:= filepath.Join(store.location, oid[:2], oid[2:])
	exists, _ := store.Stat(objectPath)
	if !exists {
		return nil, fmt.Errorf("oid %v could not be found", oid)
	}

	objectFile, err := os.Open(objectPath)
	if err != nil {
		return nil, err
	}

	reader, err := zlib.NewReader(objectFile)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	io.Copy(&buffer, reader)
	reader.Close()

	return buffer.Bytes(), nil
}

func (store *FilesystemStore) Find(prefix string) ([]string, error) {
	if err:= checkObjectId(prefix); err != nil {
		return nil, err
	}

	bucket := filepath.Join(store.location, prefix[:2])
	if _, err:= os.Stat(bucket); os.IsNotExist(err) {
		return []string{}, nil
	}

	files, err := ioutil.ReadDir(bucket)
	if err != nil {
		return nil, err
	}

	oids := make([]string, len(files))
	for i, f := range files {
		oids[i] = prefix[:2] + filepath.Base(f.Name())
	}

	return oids, nil
}

func(store *FilesystemStore) Put(oid string, data []byte) error{
	if err:= checkObjectId(oid); err != nil {
		return err
	}

	if data == nil || len(data) == 0 {
		return fmt.Errorf("empty data cannot be stored")
	}

	bucket := oid[:2]
	bucketPath := filepath.Join(store.location, bucket)
	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		err := os.Mkdir(bucketPath, os.ModeDir)
		if err != nil {
			return err
		}
	}

	objectPath:= filepath.Join(bucketPath, oid[2:])
	_, err := os.Stat(objectPath)
	if err == nil {
		return nil // object already exists
	}

	objectFile, err := os.Create(objectPath)
	if err != nil {
		return err
	}

	writer := zlib.NewWriter(objectFile)
	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

func(store *FilesystemStore) Stat(oid string) (bool, error){
	if err:= checkObjectId(oid); err != nil {
		return false, err
	}

	objectPath := filepath.Join(store.location, oid[:2], oid[2:])
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

func(store *FilesystemStore) Delete(oid string) error {
	if err:= checkObjectId(oid); err != nil {
		return err
	}

	objectPath := filepath.Join(store.location, oid[:2], oid[2:])
	if err := os.Remove(objectPath); err != nil {
		return err
	}

	return nil
}

func checkObjectId(oid string) error {
	if len(oid) < 4 {
		fmt.Errorf("oid needs to be ast least 4 characters long")
	}

	return nil
}
