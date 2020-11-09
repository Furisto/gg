package cmd

import (
	"github.com/furisto/gog/plumbing/refs"
	"github.com/furisto/gog/repo"
	"github.com/furisto/gog/storage"
	"github.com/furisto/gog/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const ParentCommit = "48743154a35f5751796d39ebceb615453abac8de"

func PrepareEnvWithCommits() (r *repo.Repository, err error) {
	r, err = PrepareEnvWithNoCommmits()
	if err != nil {
		return nil, err
	}

	if _, err := r.Branches.Create("master", ParentCommit); err != nil {
		return nil, err
	}

	return r, nil
}

func PrepareEnvWithNoCommmits() (r *repo.Repository, err error) {
	r, err = CreateTestRepository()
	if err != nil {
		return nil, err
	}

	if err := PopulateRepo(r.Location); err != nil {
		return nil, err
	}

	if err := r.Config.Set("user", "name", "furisto"); err != nil {
		return nil, err
	}
	if err := r.Config.Set("user", "email", "furisto@test.com"); err != nil {
		return nil, err
	}

	return r, nil
}

func CreateTestRepository() (*repo.Repository, error) {
	dir, err := util.CreateTemporaryDir()
	if err != nil {
		return nil, err
	}

	gitDir := filepath.Join(dir, ".git")
	store := storage.NewFsStore(gitDir)
	refs := refs.NewGitRefManager(gitDir)
	r, err := repo.Init(dir, false, store, refs)
	if err != nil {
		return nil, err
	}

	return r, err
}

func PopulateRepo(path string) error {
	for i := 0; i < 5; i++ {
		dirName := filepath.Join(path, strconv.Itoa(i))
		if err := os.Mkdir(dirName, os.ModeDir); err != nil {
			return err
		}

		for j := 0; j < 2; j++ {
			v := strconv.Itoa(j)
			if err := ioutil.WriteFile(filepath.Join(dirName, v), []byte(strconv.Itoa(i)+v), 0644); err != nil {
				return err
			}
		}
	}

	return nil
}
