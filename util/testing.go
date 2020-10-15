package util

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
)

func CreateTemporaryDir() (string, error){
	uuid, err := GenerateUUID()
	if err != nil {
		return "", err
	}

	tempDir, err := ioutil.TempDir("", uuid)
	if err != nil {
		return "", err
	}

	return tempDir, nil
}

func CreateTemporaryFile() (*os.File, error) {
	uuid, err := GenerateUUID()
	if err != nil {
		return nil, err
	}

	tempFile, err := ioutil.TempFile("", uuid)
	if err != nil {
		return nil, err
	}

	return tempFile, nil
}

func GenerateUUID() (string, error) {
	buffer := make([]byte, 16)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		buffer[0:4], buffer[4:6], buffer[6:8], buffer[8:10], buffer[10:])

	return uuid, nil
}
