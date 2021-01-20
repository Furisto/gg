package objects

import (
	"fmt"
	"github.com/furisto/gog/storage"
)

const sizeStartMarker = byte(' ')
const sizeEndMarker = byte('\x00')

type Object interface {
	OID() string
	Size() uint32
	Type() string
	Save(store storage.ObjectStore) error
}

func GetObjectType(data []byte) (string, error) {
	if IsBlob(data) {
		return "Blob", nil
	} else if IsTree(data) {
		return "Tree", nil
	} else if IsCommit(data) {
		return "Commit", nil
	} else if IsTag(data) {
		return "Tag", nil
	}

	return "", fmt.Errorf("unknown object")
}
