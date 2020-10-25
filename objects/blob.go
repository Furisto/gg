package objects

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/furisto/gog/storage"
	hasher "github.com/furisto/gog/util"
	"io/ioutil"
	"os"
	"strconv"
)

var BlobType = []byte("blob")

type Blob struct {
	oid     string
	size    uint32
	Content []byte
}

func NewBlob(content []byte) *Blob {
	blob := Blob{}
	blob.SetSize(uint32(len(content)))
	blob.Content = content

	header := blob.getHeader()
	blob.SetOID(hasher.Hash(header, content))

	return &blob
}

func NewBlobFromFile(filePath string) (*Blob, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	blob := Blob{}
	blob.SetSize(uint32(len(content)))
	blob.Content = content

	header := blob.getHeader()
	blob.SetOID(hasher.Hash(header, content))

	return &blob, nil
}

func LoadBlob(blobData []byte) (*Blob, error) {
	if !IsBlob(blobData) {
		return nil, errors.New("not of type blob")
	}

	start := bytes.IndexByte(blobData, sizeStartMarker)
	if start == -1 {
		return nil, errors.New("malformed object")
	}

	end := bytes.IndexByte(blobData, sizeEndMarker)
	if end == -1 {
		return nil, errors.New("malformed object")
	}

	if start > end {
		return nil, errors.New("malformed object")
	}

	blob := Blob{
		Content: blobData[end+1:],
	}

	size, err := strconv.Atoi(string(blobData[start+1 : end]))
	if err != nil {
		return nil, err
	}

	blob.SetSize(uint32(size))
	blob.SetOID(hasher.Hash(blobData))

	return &blob, nil
}

func IsBlob(data []byte) bool {
	return bytes.HasPrefix(data, BlobType)
}

func (b *Blob) Bytes() []byte {
	by := append(b.getHeader(), b.Content...)
	return by
}

func (b *Blob) OID() string {
	return b.oid
}

func (b *Blob) SetOID(oid string) {
	b.oid = oid
}

func (b *Blob) Size() uint32 {
	return b.size
}

func (b *Blob) SetSize(size uint32) {
	b.size = size
}

func (b *Blob) Type() string {
	return "Blob"
}

func (b *Blob) Save(store storage.ObjectStore) error {
	by := b.Bytes()
	return store.Put(b.OID(), by)
}

func (b *Blob) getHeader() []byte {
	return []byte(fmt.Sprintf("blob %d\x00", b.Size()))
}
