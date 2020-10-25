package objects

import "github.com/furisto/gog/storage"

const sizeStartMarker = byte(' ')
const sizeEndMarker = byte('\x00')

type Object interface {
	OID() string
	Size() uint32
	Type() string
	Save(store storage.ObjectStore) error
}
