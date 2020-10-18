package storage

import "fmt"

func FormatObject(o Object) (string, error) {
	switch o.Type() {
	case "Blob":
		blob, ok := o.(*Blob)
		if !ok {
			panic("object is of type blob, but cannot be cast to blob")
		}

		return string(blob.Content), nil
		break
	}

	return "", fmt.Errorf("unknown object type")
}
