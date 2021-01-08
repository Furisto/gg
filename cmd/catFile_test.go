package cmd

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/repo"
	"github.com/furisto/gog/util"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const BlobContent = "Hello Git!"
const CommitOID = "3ab896244757f512b43eb80384463d4dd9334384"

var RawBlobContent = []byte("blob 10\x00Hello Git!")

func TestPrintSizeOfBlob(t *testing.T) {
	r, blob := prepareEnvForBlobTest(t)

	options := CatFileOptions{
		OID:    blob.OID(),
		Path:   r.Info.WorkingDirectory(),
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
	r, blob := prepareEnvForBlobTest(t)

	options := CatFileOptions{
		OID:    blob.OID(),
		Path:   r.Info.WorkingDirectory(),
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
	r, blob := prepareEnvForBlobTest(t)

	options := CatFileOptions{
		OID:    blob.OID(),
		Path:   r.Info.WorkingDirectory(),
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

func TestRawPrintOfBlob(t *testing.T) {
	r, blob := prepareEnvForBlobTest(t)

	options := CatFileOptions{
		OID:    blob.OID(),
		Path:   r.Info.WorkingDirectory(),
		Type:   false,
		Size:   false,
		Pretty: false,
		Raw:    true,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
		return
	}

	if !bytes.Equal(RawBlobContent, output.Bytes()) {
		t.Errorf("content was expected to be %s, but was %s", RawBlobContent, output.Bytes())
	}
}

func TestPrintSizeOfTree(t *testing.T) {
	r, tree := prepareEnvForTreeTest(t)

	options := CatFileOptions{
		OID:    tree.OID(),
		Path:   r.Info.WorkingDirectory(),
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
	r, tree := prepareEnvForTreeTest(t)

	options := CatFileOptions{
		OID:    tree.OID(),
		Path:   r.Info.WorkingDirectory(),
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
	r, tree := prepareEnvForTreeTest(t)

	options := CatFileOptions{
		OID:    tree.OID(),
		Path:   r.Info.WorkingDirectory(),
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

func TestRawPrintOfTree(t *testing.T) {
	ry, tree := prepareEnvForTreeTest(t)

	expected := readGoldenTree(t)

	options := CatFileOptions{
		OID:    tree.OID(),
		Path:   ry.Info.WorkingDirectory(),
		Type:   false,
		Size:   false,
		Pretty: false,
		Raw:    true,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
		return
	}

	if !bytes.Equal(expected, output.Bytes()) {
		t.Errorf("content was expected to be %s, but was %s", expected, output.Bytes())
	}
}

func TestPrintSizeOfCommit(t *testing.T) {
	ry := prepareEnvForCommitTest(t)

	options := CatFileOptions{
		OID:    CommitOID,
		Path:   ry.Info.WorkingDirectory(),
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

	if size != 215 {
		t.Errorf("expected length of %v, but length is %v", 12, size)
	}
}

func TestPrintTypeOfCommit(t *testing.T) {
	ry := prepareEnvForCommitTest(t)

	options := CatFileOptions{
		OID:    CommitOID,
		Path:   ry.Info.WorkingDirectory(),
		Type:   true,
		Size:   false,
		Pretty: false,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
	}

	if !bytes.Equal([]byte("Commit"), output.Bytes()) {
		t.Errorf("expected type of %v, but was of type %v", "Blob", output.String())
	}
}

func TestPrettyPrintCommit(t *testing.T) {
	ry := prepareEnvForCommitTest(t)

	options := CatFileOptions{
		OID:    CommitOID,
		Path:   ry.Info.WorkingDirectory(),
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

	expected := "tree 80fa9593f3c3d03f011492504e5d877b97b1277f\n" +
		"author Furisto <24721048+Furisto@users.noreply.github.com> 1609952762 +0000\n" +
		"committer Furisto <24721048+Furisto@users.noreply.github.com> 1609952762 +0000\n\n" +
		"print commit\n"

	if expected != output.String() {
		t.Errorf("did not receive expected output. Got \n %s \n\n but expected \n %s", output.String(), expected)
	}
}

func TestRawPrintOfCommmit(t *testing.T) {
	ry := prepareEnvForCommitTest(t)

	options := CatFileOptions{
		OID:    CommitOID,
		Path:   ry.Info.WorkingDirectory(),
		Type:   false,
		Size:   false,
		Pretty: false,
		Raw:    true,
	}

	output := bytes.Buffer{}
	cmd := NewCatFileCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("error occured during command execution: %v", err)
		return
	}

	expected := "commit 215\u0000tree 80fa9593f3c3d03f011492504e5d877b97b1277f\n" +
		"author Furisto <24721048+Furisto@users.noreply.github.com> 1609952762 +0100\n" +
		"committer Furisto <24721048+Furisto@users.noreply.github.com> 1609952762 +0100\n\n" +
		"print commit\n"

	if expected != output.String() {
		t.Errorf("did not receive expected output. Got \n %s \n\n but expected \n %s", output.String(), expected)
	}
}

func readGoldenTree(t *testing.T) []byte {
	t.Helper()

	testFilePath := filepath.Join("./testdata/print_tree")
	fileContent, err := os.Open(testFilePath)
	if err != nil {
		t.Fatalf("could not read test file at %s: %v", testFilePath, err)
	}

	reader, err := zlib.NewReader(fileContent)
	if err != nil {
		t.Fatalf("could not create zlib reader for %s: %v", testFilePath, err)
	}

	var decompressed bytes.Buffer
	if _, err = io.Copy(&decompressed, reader); err != nil {
		t.Fatalf("could not copy")
	}

	return decompressed.Bytes()
}

func prepareEnvForBlobTest(t *testing.T) (*repo.Repository, *objects.Blob) {
	t.Helper()

	ry := createTestRepository(t)

	blob := objects.NewBlob([]byte(BlobContent))
	if err := ry.Storage.Put(blob.OID(), blob.Bytes()); err != nil {
		t.Fatalf("")
	}

	return ry, blob
}

func prepareEnvForTreeTest(t *testing.T) (*repo.Repository, *objects.Tree) {
	t.Helper()

	ry := createTestRepository(t)
	populateRepo(t, ry.Info.WorkingDirectory())

	if err := ry.Index.Add(ry.Info.WorkingDirectory()); err != nil {
		t.Fatalf("could not add working directory to index: %v", err)
	}

	index_converter := repo.NewIndexToTreeConverter(ry.Index)
	tree, err := index_converter.Convert()
	if err != nil {
		t.Fatalf("could not convert index to tree: %v", err)
	}

	if err := tree.Save(ry.Storage); err != nil {
		t.Fatalf("could not save tree: %v", err)
	}

	return ry, tree
}

func prepareEnvForCommitTest(t *testing.T) *repo.Repository {
	t.Helper()

	ry := createTestRepository(t)

	testFilePath := filepath.Join("./testdata/print_commit")
	testFileData, err := util.ReadAndDecompressFile(testFilePath)
	if err != nil {
		t.Fatalf("could not decompress %s: %v", testFilePath, err)
	}

	if err := ry.Storage.Put(CommitOID, testFileData); err != nil {
		t.Fatalf("could not store commit in object database: %v", err)
	}
	return ry
}
