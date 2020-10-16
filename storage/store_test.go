package storage

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/furisto/gog/util"
	"os"
	"path/filepath"
	"testing"
)

const oid = "12345"
var fileContent = []byte("blob 25\x00The content of a blob")

func TestFsStorePutObjectDoesNotExist(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)

	store := NewFsStore(dir)
	const objectId = "abcd"
	if err:= store.Put(objectId, fileContent); err != nil {
		t.Errorf("encountered error writing content to store: %v", err)
	}

	objectDir:= filepath.Join(dir, "objects")
	bucketPath := filepath.Join(objectDir, objectId[:2])
	if _, err := os.Stat(bucketPath); err != nil {
		t.Errorf("expected bucket %v does not exist", bucketPath)
	}

	objectPath := filepath.Join(objectDir, objectId[:2], objectId[2:])
	if _, err := os.Stat(objectPath); err != nil {
		t.Errorf("expected file %v does nost exist", objectPath)
	}
}

func TestFsStoreGetExistingObject(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)

	_, err = createStorageObject(dir)
	if err != nil {
		t.Fatalf("could not create test file: %v", err)
		return
	}

	store :=NewFsStore(dir)
	data, err := store.Get(oid)
	if err!= nil {
		t.Errorf("could not find expected oid %v", oid)
		return
	}

	if !bytes.Equal(data, fileContent) {
		t.Errorf("expected value of %v, but was %v", fileContent, data)
	}
}

func TestFsStoreGetNonExistingObject(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)

	store := NewFsStore(dir)
	data, err := store.Get(oid)
	if err == nil {
		t.Errorf("expected error retrieving non existent oid, but no error was returned")
	}

	if data != nil {
		t.Errorf("data for non existent oid has been returned unexpectedly")
	}
}

func TestFsStoreStatObjectExists(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)

	objectPath, err := createStorageObject(dir)
	if err != nil {
		t.Fatalf("could not create test file: %v", err)
		return
	}
	defer os.Remove(objectPath)

	store := NewFsStore(dir)
	if exists, _ := store.Stat(oid); !exists{
		t.Errorf("could not find oid %v in store", oid)
	}
}

func TestFsStoreStatObjectDoesNotExist(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)

	store := NewFsStore(dir)
	if exists, _ := store.Stat(oid); exists{
		t.Errorf("found oid %v, oid should not exist", oid)
	}
}

func TestFsStoreDeleteObjectExists(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)
	_, err = createStorageObject(dir)
	if err != nil {
		t.Fatalf("could not create test file: %v", err)
		return
	}

	store:= NewFsStore(dir)
	if err := store.Delete(oid); err != nil {
		t.Errorf("could not delete oid %v: %v", oid, err)
	}
}

func createStorageObject(dir string) (string, error){
	dirName := filepath.Join(dir, "objects", "12")
	if err := os.MkdirAll(dirName, 0644); err != nil {
		return "", fmt.Errorf("could not create bucket directory: %v", err)
	}

	fileName := filepath.Join(dirName, "345")
	objectFile, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer objectFile.Close()

	writer:= zlib.NewWriter(objectFile)
	defer writer.Close()
	_, err = writer.Write(fileContent)
	if err != nil {
		return "", fmt.Errorf("could not create test file %v: %v", fileName, err)
	}

	return fileName, nil
}


