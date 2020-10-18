package cmd

import (
	"bytes"
	"github.com/furisto/gog/repo"
	"github.com/furisto/gog/storage"
	"github.com/furisto/gog/util"
	"path/filepath"
	"strconv"
	"testing"
)

func TestPrintSizeOfBlob(t *testing.T) {
	output:= bytes.Buffer{}
	repo, err := CreateTestRepository()
	if err != nil {
		t.Fatalf("")
		return
	}

	blob := storage.NewBlob([]byte("Hello Git!"))
	if err:= repo.Storage.Put(blob.OID(), blob.Bytes()); err != nil {
		t.Fatalf("cannot store data: %v", err)
		return
	}

	options := CatFileOptions{
		OID:    blob.OID(),
		Path: repo.Location,
		Type:   false,
		Size:   true,
		Pretty: false,
	}

	cmd:= NewCatFileCmd(&output)
	if err:= cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
	}

	size, err := strconv.Atoi(output.String())
	if err != nil {
		t.Errorf("output could not be parsed as int: %v", err)
		return
	}

	if size != 10 {
		t.Errorf("expected length of %v, but length is %v", 12, size)
	}
}

func TestPrintTypeOfBlob(t *testing.T) {
	output:= bytes.Buffer{}
	repo, err := CreateTestRepository()
	if err != nil {
		t.Fatalf("")
	}

	blob := storage.NewBlob([]byte("Hello Git!"))
	if err:= repo.Storage.Put(blob.OID(), blob.Bytes()); err != nil {
		t.Fatalf("cannot store data: %v", err)
	}

	options := CatFileOptions{
		OID:    blob.OID(),
		Path: repo.Location,
		Type:   true,
		Size:   false,
		Pretty: false,
	}

	cmd:= NewCatFileCmd(&output)
	if err:= cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
	}

	if !bytes.Equal([]byte("Blob"), output.Bytes()) {
		t.Errorf("expected type of %v, but was of type %v", "Blob", output.String())
	}
}

func TestPrettyPrintOfBlob(t *testing.T) {
	output:= bytes.Buffer{}
	repo, err := CreateTestRepository()
	if err != nil {
		t.Fatalf("")
	}

	var content = []byte("Hello Git!")
	blob := storage.NewBlob(content)
	if err:= repo.Storage.Put(blob.OID(), blob.Bytes()); err != nil {
		t.Fatalf("cannot store data: %v", err)
	}

	options := CatFileOptions{
		OID:    blob.OID(),
		Path:   repo.Location,
		Type:   false,
		Size:   false,
		Pretty: true,
	}

	cmd:= NewCatFileCmd(&output)
	if err:= cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
	}

	if !bytes.Equal(content, output.Bytes()) {
		t.Errorf("content was expected to be '%v', but was '%v'", content, output.Bytes())
	}
}

func CreateTestRepository() (*repo.Repository, error) {
	dir, err := util.CreateTemporaryDir()
	if err!= nil {
		return nil, err
	}

	store := storage.NewFsStore(filepath.Join(dir, ".git"))
	repo, err := repo.Init(dir, false, store)
	if err!= nil {
		return nil, err
	}

	return repo, err
}
