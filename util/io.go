package util

import "os"

func CloseFile(file *os.File, err error) {
	if cerr := file.Close(); cerr != nil && err == nil {
		err = cerr
	}
}
