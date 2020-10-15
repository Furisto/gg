package storage

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestCreateBlob(t *testing.T) {
	filePath, err := createBlobFile()
	if err != nil {
		t.Fatal("Could not create blob file")
	}

	blob, err := BlobFromFile(filePath)
	if err != nil {
		t.Error("Could not create blob")
	}

	const oid = "34c92c2d93d8cdb680b66118dd37551caa0b4a25"
	if blob.OID != oid {
		t.Errorf("Expected oid of %v, but was %v", oid, blob.OID)
	}

	const size = 11
	if blob.Size != 11 {
		t.Errorf("Expected blob size of %v, but was %v", 11, blob.Size)
	}

	if bytes.Compare(blob.Content, []byte("lorem ipsum")) != 0 {
		t.Errorf("Expcted blob content to be %v, but was %v", "", blob.Content)
	}
}

func createBlobFile() (string, error) {
	file, err := ioutil.TempFile("", "foo")
	if err != nil {
		return "", err
	}
	defer file.Close()

	file.Write([]byte("lorem ipsum"))
	return file.Name(), nil
}
