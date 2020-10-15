package util

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
)

func Hash(data ...[]byte) string {
	sha := sha1.New()
	for _, d := range data {
		sha.Write(d)
	}

	return fmt.Sprintf("%x", sha.Sum(nil))
}

func HashFile(fileName string) (string, error) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return Hash(content), nil
}


