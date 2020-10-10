package repo

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestIsRepositoryCreatedWithDefaultStorage(t *testing.T) {
	repoPath, err := createTempRepoPath()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	t.Logf("Repo directory is %v", repoPath)
	defer os.RemoveAll(repoPath)

	InitDefault(repoPath, false)
	repoPath = filepath.Join(repoPath, ".git")

	RepoObjectsExist(t, repoPath)
}

func TestIsBareRepositoryCreatedWithDefaultStorage(t *testing.T) {
	repoPath, err := createTempRepoPath()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(repoPath)

	InitDefault(repoPath, true)
	RepoObjectsExist(t, repoPath)
}

func RepoObjectsExist(t *testing.T, repoPath string) {
	AssertFsObjectExists(t, filepath.Join(repoPath, "hooks"))
	AssertFsObjectExists(t, filepath.Join(repoPath, "info"))
	AssertFsObjectExists(t, filepath.Join(repoPath, "refs"))

	AssertFsObjectExists(t, filepath.Join(repoPath, "config"))
	AssertFsObjectExists(t, filepath.Join(repoPath, "description"))
	AssertFsObjectExists(t, filepath.Join(repoPath, "HEAD"))
}

func AssertFsObjectExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Does not exist %v", path)
	}
}

func createTempRepoPath() (string, error){
	uuid, err := generateUUID()
	if err != nil {
		return "", err
	}

	repoPath, err := ioutil.TempDir("", uuid)
	if err != nil {
		return "", err
	}

	return repoPath, nil
}

func generateUUID() (string, error) {
	buffer := make([]byte, 16)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		buffer[0:4], buffer[4:6], buffer[6:8], buffer[8:10], buffer[10:])

	return uuid, nil
}


