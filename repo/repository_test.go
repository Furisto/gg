package repo

import (
	"github.com/furisto/gog/util"
	"os"
	"path/filepath"
	"testing"
)

func TestIsRepositoryCreatedWithDefaultStorage(t *testing.T) {
	repoPath, err := util.CreateTemporaryDir()
	if err != nil {
		t.Fatal("Could not create temporary directory")
		return
	}
	defer os.RemoveAll(repoPath)

	InitDefault(repoPath, false)
	repoPath = filepath.Join(repoPath, ".git")

	RepoObjectsExist(t, repoPath)
}

func TestIsBareRepositoryCreatedWithDefaultStorage(t *testing.T) {
	repoPath, err := util.CreateTemporaryDir()
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





