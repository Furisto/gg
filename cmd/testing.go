package cmd

import (
	"crypto/rand"
	"fmt"
	"github.com/furisto/gog/plumbing/refs"
	"github.com/furisto/gog/repo"
	"github.com/furisto/gog/storage"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
)

const ParentCommit = "48743154a35f5751796d39ebceb615453abac8de"

func createTemporaryDir(t *testing.T) string {
	t.Helper()

	uuid := generateUUID(t)
	tempDir, err := ioutil.TempDir("", uuid)
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}

	return tempDir
}

func createTemporaryFile(t *testing.T) (*os.File, error) {
	t.Helper()

	uuid := generateUUID(t)
	tempFile, err := ioutil.TempFile("", uuid)
	if err != nil {
		return nil, err
	}

	return tempFile, nil
}

func generateUUID(t *testing.T) string {
	t.Helper()

	buffer := make([]byte, 16)
	_, err := rand.Read(buffer)
	if err != nil {
		t.Fatalf("could not create random number: %v", err)
	}

	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		buffer[0:4], buffer[4:6], buffer[6:8], buffer[8:10], buffer[10:])
	return uuid
}

func prepareEnvWithCommits(t *testing.T) *repo.Repository {
	t.Helper()

	ry := PrepareEnvWithNoCommmits(t)

	if _, err := ry.Branches.Create("master", ParentCommit); err != nil {
		t.Fatalf("could not create branch 'master'")
	}

	return ry
}

func PrepareEnvWithNoCommmits(t *testing.T) *repo.Repository {
	t.Helper()

	ry := createTestRepository(t)
	populateRepo(t, ry.Info.WorkingDirectory())

	if err := ry.Config.Set("user", "name", "furisto"); err != nil {
		t.Fatalf("could not set user name")
	}
	if err := ry.Config.Set("user", "email", "furisto@test.com"); err != nil {
		t.Fatalf("could not set user email")
	}

	return ry
}

func createTestRepository(t *testing.T) *repo.Repository {
	t.Helper()

	dir := createTemporaryDir(t)
	gitDir := path.Join(dir, ".git")

	store := storage.NewFsStore(gitDir)
	refMgr := refs.NewGitRefManager(gitDir)
	r, err := repo.Init(dir, false, store, refMgr)
	if err != nil {
		t.Fatalf("could not initialize test repository: %v", err)
	}

	return r
}

func populateRepo(t *testing.T, path string) {
	t.Helper()

	for i := 0; i < 5; i++ {
		dirName := filepath.Join(path, strconv.Itoa(i))
		if err := os.Mkdir(dirName, os.ModeDir); err != nil {
			t.Fatalf("could not create test directory: %v", err)
		}

		for j := 0; j < 2; j++ {
			v := strconv.Itoa(j)
			if err := ioutil.WriteFile(filepath.Join(dirName, v), []byte(strconv.Itoa(i)+v), 0644); err != nil {
				t.Fatalf("could not create test blob: %v", err)
			}
		}
	}
}
