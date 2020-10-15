package storage

import (
	"github.com/furisto/gog/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFsStorePutObjectDoesNotExist(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)

	store := NewFsStore(dir)
	const objectId = "abcd"
	store.Put(objectId, []byte("blob 25\x00The content of a blob"))

	bucketPath := filepath.Join(dir, objectId[:2])
	if _, err := os.Stat(bucketPath); err != nil {
		t.Errorf("expected bucket %v does not exist", bucketPath)
	}

	objectPath := filepath.Join(dir, objectId[:2], objectId[2:])
	if _, err := os.Stat(objectPath); err != nil {
		t.Errorf("expected file %v does nost exist", objectPath)
	}
}

func TestFsStorePutObjectAlreadyExists(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)



	const objectId = "abcd"
	ioutil.WriteFile(filepath.Join(dir, objectId[:2], objectId[2:]), []byte("content"), 644)

	store := NewFsStore(dir)
	store.Put(objectId, []byte("blob 25\x00The content of a blob"))


}

func TestFsStoreGetExistingObject(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)
}

func TestFsStoreGetNonExistingObject(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)
}

func TestFsStoreStatObjectExists(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)
}

func TestFsStoreStatObjectDoesNotExist(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)
}

func TestFsStoreDeleteObjectExists(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)
}

func TestFsStoreDeleteObjectDoesNotExist(t *testing.T) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(dir)
}


