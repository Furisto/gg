package storage

import (
	"fmt"
	hasher "github.com/furisto/gog/util"
	"io/ioutil"
	"os"
)

type Blob struct {
	OID string
	Size int
	Content []byte
}

func BlobFromFile(filePath string) (*Blob, error){
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	content, err := ioutil.ReadFile(filePath);
	if err != nil {
		return nil, err
	}

	blob := Blob {
		Size: len(content),
		Content: content,
	}
	
	header:= blob.getHeader()
	blob.OID = hasher.Hash(header, content)

	return &blob, nil
}

func BlobFromBytes(content []byte) (*Blob, error) {
	return nil, nil
}

func (b *Blob) Bytes()[]byte {
	bytes := append(b.getHeader(), b.Content...)
	return bytes
}

func (b *Blob) getHeader() []byte {
	return []byte(fmt.Sprintf("blob %v\x00", b.Size))
}
