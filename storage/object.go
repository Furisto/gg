package storage

const sizeStartMarker = byte(' ')
const sizeEndMarker = byte('\x00')

type Object interface {
	OID() string
	Size() uint32
	Type() string
}




