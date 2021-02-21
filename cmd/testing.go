package cmd

import (
	"crypto/rand"
	"fmt"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/plumbing/refs"
	"github.com/furisto/gog/repo"
	"github.com/furisto/gog/storage"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"
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

func prepareEnvWithCommitObjects(t *testing.T) (*repo.Repository, []*objects.Commit) {
	t.Helper()

	ry := createTestRepository(t)
	commits := make([]*objects.Commit, 0, 5)

	for i := 0; i < 5; i++ {
		dirName := filepath.Join(ry.Info.WorkingDirectory(), strconv.Itoa(i))
		if err := os.Mkdir(dirName, os.ModeDir); err != nil {
			t.Fatalf("could not create test directory: %v", err)
		}

		for j := 0; j < 2; j++ {
			v := strconv.Itoa(j)
			if err := ioutil.WriteFile(filepath.Join(dirName, v), []byte(strconv.Itoa(i)+v), 0644); err != nil {
				t.Fatalf("could not create test blob: %v", err)
			}
		}

		commit, err := ry.Commit(func(builder *objects.CommitBuilder) *objects.CommitBuilder {
			builder.WithMessage("commit1").
				WithHook(func(commit *objects.Commit) {
					commit.Commiter.TimeStamp = time.Unix(1611606380+int64(i)*10000, 0)
					commit.Author.TimeStamp = time.Unix(1611606380+int64(i)*10000, 0)
				})
			return builder
		})

		if err != nil {
			t.Fatalf("could not create commit object: %v", err)
		}

		commits = append(commits, commit)
	}

	return ry, commits
}

func prepareEnvWithTags(t *testing.T) (*repo.Repository, []*objects.Commit, []*refs.Ref) {
	t.Helper()

	ry, commits := prepareEnvWithCommitObjects(t)
	tags := make([]*refs.Ref, 4)

	for i := 0; i < 4; i++ {
		tagName := "lightweight" + strconv.Itoa(i)
		var tagValue string
		if i%2 == 0 {
			tagValue = commits[0].OID()
		} else {
			tagValue = "abcd"
		}
		r, err := ry.Tags.CreateLightweight(tagName, tagValue, false)
		if err != nil {
			t.Fatalf("could not create %s", tagName)
		}

		tags[i] = r
	}

	return ry, commits, tags
}

func prepareEnvWithAnnotatedTag(t *testing.T) (*repo.Repository, *objects.Tag) {
	t.Helper()

	ry, commits := prepareEnvWithCommitObjects(t)

	tagger := objects.Signature{
		Name:      "furisto",
		Email:     "furisto@test.com",
		TimeStamp: time.Unix(1611348418, 0),
	}
	tag, err := ry.Tags.CreateAnnotated("annotated", commits[0].OID(), &tagger, "annotated tag", false)
	if err != nil {
		t.Fatalf("could not create annotated tag: %v", err)
	}

	return ry, tag
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

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

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
