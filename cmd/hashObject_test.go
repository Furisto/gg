package cmd

import (
	"bytes"
	"github.com/furisto/gog/repo"
	"github.com/furisto/gog/storage"
	"github.com/furisto/gog/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var fileContent = []byte("Hello git!")
var hashedContent = []byte("57ea241164ccfd0b63d58eb247d52a670514b370")

func TestHashObjectOnlyNoRepository(t *testing.T) {
	output := bytes.Buffer{}

	file, err := util.CreateTemporaryFile()
	if err != nil {
		t.Fatalf("Could not create temporary file: %v", err)
		return
	}
	defer os.Remove(file.Name())

	if err := ioutil.WriteFile(file.Name(), fileContent, 0644); err != nil {
		t.Fatalf("Could not write to test file: %v", err)
		return
	}

	cmd := NewHashObjectCmd(&output)
	options := HashObjectOptions{
		file: file.Name(),
		store: false,
	}

	cmd.Execute(options)

	if !bytes.Equal(output.Bytes(), hashedContent) {
		t.Errorf("Expected hash of %v, but was %v", hashedContent, output)
	}
}

func TestHashAndStore(t *testing.T) {
	repoDir, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatalf("Could not create temporary directory: %v", err)
		return
	}
	defer os.RemoveAll(repoDir)
	output:= bytes.Buffer{}

	repo, err := repo.Init(repoDir, false, storage.NewFsStore(filepath.Join(repoDir, ".git")))
	if err != nil {
		t.Fatalf("Could not create repository: %v", err)
		return
	}
	os.Chdir(repoDir)
	file := filepath.Join(repoDir, "foo")
	ioutil.WriteFile(file, fileContent, 0644)
	cmd:= NewHashObjectCmd(&output)
	options := HashObjectOptions{
		file:  file,
		store: true,
	}

	cmd.Execute(options)
	if !bytes.Equal(output.Bytes(), hashedContent) {
		t.Errorf("Expected hash of %v, but was %v", hashedContent, output)
	}

	exists, _ := repo.Storage.Stat(string(hashedContent))
	if !exists {
		t.Errorf("Could not find key %v in git storage", string(hashedContent))
	}
}
