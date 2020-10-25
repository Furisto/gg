package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestWriteTreeFromWorkingDirNoPrefix(t *testing.T) {
	r, err := CreateTestRepository()
	if err != nil {
		t.Fatal("")
		return
	}

	if err := populateRepo(r.Location); err != nil {
		t.Fatal("")
		return
	}

	options := WriteTreeOptions{
		Path:                r.Location,
		UseWorkingDirectory: true,
		Prefix:              "",
	}

	var output bytes.Buffer
	cmd := NewWriteTreeCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Errorf("")
		return
	}

	results := map[string][]byte{
		"f37c2df7d15983fffe5455764f3c97263a3c492e": []byte("blob 84\x00C:\\Users\\Furisto\\AppData\\Local\\Temp\\4631e64f-1bce-652c-d68c-6fd045d05e8a269543607\\01"),
	}

	for k, v := range results {
		data, err := r.Storage.Get(k)
		if err != nil || data == nil {
			t.Errorf("could not retrieve data for key %v", k)
		}

		if !bytes.Equal(v, data) {
			t.Errorf("expected data for key %v to be %v, but was %v", k, string(v), string(data))
		}
	}
}

func populateRepo(path string) error {
	for i := 0; i < 5; i++ {
		dirName := filepath.Join(path, strconv.Itoa(i))
		if err := os.Mkdir(dirName, os.ModeDir); err != nil {
			return err
		}

		for i := 0; i < 2; i++ {
			v := strconv.Itoa(i)
			if err := ioutil.WriteFile(filepath.Join(dirName, v), []byte(dirName+v), 0644); err != nil {
				return err
			}
		}
	}

	return nil
}
