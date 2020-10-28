package cmd

import (
	"bufio"
	"bytes"
	"github.com/furisto/gog/objects"
	"github.com/furisto/gog/repo"
	"github.com/furisto/gog/storage"
	"github.com/furisto/gog/util"
	"io"
	"path/filepath"
	"strconv"
	"testing"
)

const BlobContent = "Hello Git!"

func TestPrintSizeOfBlob(t *testing.T) {
	r, blob, err := prepareEnvForBlobTest()
	if err != nil {
		t.Fatalf("could not prepare test environment: %v", err)
	}

	options := CatFileOptions{
		OID:    blob.OID(),
		Path:   r.Location,
		Type:   false,
		Size:   true,
		Pretty: false,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
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
	r, blob, err := prepareEnvForBlobTest()
	if err != nil {
		t.Fatalf("could not prepare test environment: %v", err)
	}

	options := CatFileOptions{
		OID:    blob.OID(),
		Path:   r.Location,
		Type:   true,
		Size:   false,
		Pretty: false,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
	}

	if !bytes.Equal([]byte("Blob"), output.Bytes()) {
		t.Errorf("expected type of %v, but was of type %v", "Blob", output.String())
	}
}

func TestPrettyPrintOfBlob(t *testing.T) {
	r, blob, err := prepareEnvForBlobTest()
	if err != nil {
		t.Fatalf("could not prepare test environment: %v", err)
	}

	options := CatFileOptions{
		OID:    blob.OID(),
		Path:   r.Location,
		Type:   false,
		Size:   false,
		Pretty: true,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
	}

	if !bytes.Equal([]byte(BlobContent), output.Bytes()) {
		t.Errorf("content was expected to be '%v', but was '%v'", BlobContent, output.Bytes())
	}
}

func TestPrintSizeOfTree(t *testing.T) {
	r, tree, err := prepareEnvForTreeTest()
	if err != nil {
		t.Fatalf("could not prepare test environment: %v", err)
	}

	options := CatFileOptions{
		OID:    tree.OID(),
		Path:   r.Location,
		Type:   false,
		Size:   true,
		Pretty: false,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
		return
	}

	const expectedSize = "140"
	if output.String() != expectedSize {
		t.Errorf("expected size of %v, but was %v", expectedSize, output.String())
	}
}

func TestPrintTypeOfTree(t *testing.T) {
	r, tree, err := prepareEnvForTreeTest()
	if err != nil {
		t.Fatalf("could not prepare test environment: %v", err)
	}

	options := CatFileOptions{
		OID:    tree.OID(),
		Path:   r.Location,
		Type:   true,
		Size:   false,
		Pretty: false,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
		return
	}

	if output.String() != "Tree" {
		t.Errorf("expected type '%v' but got type '%v'", "Tree", output.String())
	}
}

func TestPrettyPrintOfTree(t *testing.T) {
	r, tree, err := prepareEnvForTreeTest()
	if err != nil {
		t.Fatalf("could not prepare test environment: %v", err)
	}

	options := CatFileOptions{
		OID:    tree.OID(),
		Path:   r.Location,
		Type:   false,
		Size:   false,
		Pretty: true,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
		return
	}

	treeContent := []string{
		"40000 tree 9aacd487c128e9d564997629c0c4257f44183aaf     0",
		"40000 tree 44f70e4f280f5641a30d69706500490032ccce59     1",
		"40000 tree a1ccacffd24f2c562e75f1fa9502eed3428e4aa2     2",
		"40000 tree ca2b251fcfd68d8453c594152521a246c249d8ef     3",
		"40000 tree 7be5f5c4d3cc7b3d007865832f5f00fc442d4075     4",
	}

	reader := bufio.NewReader(&output)
	lineCount := 0

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}

			t.Errorf("encountered error reading cat-file output")
			break
		}

		if lineCount >= len(treeContent) {
			t.Errorf("received more lines of output than the expected %v", len(treeContent))
			break
		}

		if treeContent[lineCount] != string(line) {
			t.Errorf("expected content was %v, but got %v", treeContent[lineCount], string(line))
		}

		lineCount++
	}

	if lineCount != len(treeContent) {
		t.Errorf("received less lines of output than the expected %v", len(treeContent))
	}
}

func prepareEnvForBlobTest() (r *repo.Repository, blob *objects.Blob, err error) {
	r, err = CreateTestRepository()
	if err != nil {
		return nil, nil, err
	}

	blob = objects.NewBlob([]byte(BlobContent))
	if err := r.Storage.Put(blob.OID(), blob.Bytes()); err != nil {
		return nil, nil, err
	}

	return r, blob, nil
}

func prepareEnvForTreeTest() (r *repo.Repository, tree *objects.Tree, err error) {
	r, err = CreateTestRepository()
	if err != nil {
		return nil, nil, err
	}

	if err := populateRepo(r.Location); err != nil {
		return nil, nil, err
	}

	tree, err = objects.NewTreeFromDirectory(r.Location, "")
	if err != nil {
		return nil, nil, err
	}
	if err := tree.Save(r.Storage); err != nil {
		return nil, nil, err
	}

	return r, tree, nil
}

func CreateTestRepository() (*repo.Repository, error) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		return nil, err
	}

	store := storage.NewFsStore(filepath.Join(dir, ".git"))
	r, err := repo.Init(dir, false, store)
	if err != nil {
		return nil, err
	}

	return r, err
}
