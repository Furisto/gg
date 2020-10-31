package repo

import (
	"fmt"
	"github.com/furisto/gog/config"
	"github.com/furisto/gog/storage"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type Repository struct {
	Location string
	Storage  storage.ObjectStore
	Config   *config.RepoConfig
}

func Init(path string, bare bool, storage storage.ObjectStore) (*Repository, error) {
	var repoPath string
	if bare {
		repoPath = path
	} else {
		repoPath = filepath.Join(path, ".git")
	}

	directories := []string{"hooks", "info", "refs"}
	files := map[string][]byte{
		"description": []byte("Unnamed repository; edit this file 'description' to name the repository.\n"),
		"HEAD":        []byte("ref: refs/heads/master"),
	}

	for _, directory := range directories {
		if err := os.MkdirAll(filepath.Join(repoPath, directory), os.ModeDir); err != nil {
			return nil, err
		}
	}

	for k, v := range files {
		if err := ioutil.WriteFile(filepath.Join(repoPath, k), v, 0644); err != nil {
			return nil, err
		}
	}

	config, err := createConfig(
		filepath.Join(repoPath, "config"), map[string]string{"bare": strconv.FormatBool(bare)})
	if err != nil {
		return nil, err
	}

	repo := NewRepo(filepath.Dir(repoPath), storage, config)
	return repo, nil
}

func InitDefault(path string, bare bool) (*Repository, error) {
	repo, err := Init(path, bare, storage.NewFsStore(filepath.Join(path, ".git")))
	return repo, err
}

func NewRepo(path string, store storage.ObjectStore, config *config.RepoConfig) *Repository {
	return &Repository{
		Location: path,
		Storage:  store,
		Config:   config,
	}
}

func FromExisting(path string) (*Repository, error) {
	if path == filepath.Dir(path) {
		return nil, fmt.Errorf("fatal: not a git repository (or any of the parent directories)")
	}

	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return FromExisting(filepath.Dir(path))
	}

	return &Repository{
		Location: path,
		Storage:  storage.NewFsStore(gitPath),
		Config:   &config.RepoConfig{Location: path},
	}, nil
}

func createConfig(configPath string, values map[string]string) (*config.RepoConfig, error) {
	conf := ini.Empty()
	coreSection, err := conf.NewSection("core")
	if err != nil {
		return nil, err
	}

	coreSection.NewKey("repositoryformatversion", "0")
	coreSection.NewKey("filemode", "false")
	coreSection.NewKey("symlinks", "false")
	coreSection.NewKey("ignorecase", "true")

	for k, v := range values {
		if _, err := coreSection.NewKey(k, v); err != nil {
			return nil, err
		}
	}

	conf.SaveTo(configPath)
	repoConfig := config.RepoConfig{
		Location: configPath,
	}

	return &repoConfig, nil
}
