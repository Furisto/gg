package repo

import (
	"gopkg.in/ini.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type ObjectStore interface {
	Get(oid string)  ([]byte, error)
	Put(oid string, data []byte) error
	Stat(oid string) error
}

type FilesystemStore struct {
	location string
}

func NewFsStore(path string) *FilesystemStore {
	path = filepath.Join(path, "objects")
	return &FilesystemStore{
		location: path,
	}
}

func(store *FilesystemStore) Get(oid string) ([]byte, error) {
	return nil, nil
}

func(store *FilesystemStore) Put(oid string, data []byte) error{
	return nil
}

func(store *FilesystemStore) Stat(oid string) error{
	return nil
}

type RepoConfig struct {
	location string
}

func (config *RepoConfig) Set(section string, key string, value string) error {
	return nil
}

func (config *RepoConfig) Get(section string, key string) (string, error) {
	return "", nil
}

type Repository struct {
	location string
	storage ObjectStore
	config *RepoConfig
}

func Init(path string, bare bool, storage ObjectStore) (*Repository, error){
	var repoPath string
	if bare {
		repoPath = path
	} else {
		repoPath = filepath.Join(path, ".git")
	}

	directories := []string {"hooks", "info", "refs"}
	files := map[string][]byte{
		"description": []byte ("Unnamed repository; edit this file 'description' to name the repository.\n"),
		"HEAD": []byte ("ref: refs/heads/master"),
	}

	for _, directory := range directories {
		if err := os.MkdirAll(filepath.Join(repoPath, directory), os.ModeDir); err != nil {
			return nil, err
		}
	}

	for k,v := range files {
		if err := ioutil.WriteFile(filepath.Join(repoPath, k), v, 0644); err != nil {
			return nil, err
		}
	}

	config, err := createConfig(
		filepath.Join(repoPath, "config"), map[string]string { "bare": strconv.FormatBool(bare)})
	if err != nil {
		return nil, err
	}

	repo:= NewRepo(repoPath,storage, config)
	return repo, nil
}

func InitDefault(path string, bare bool) (*Repository, error) {
	repo, err := Init(path, bare, NewFsStore(path))
	return repo, err
}

func NewRepo(path string, store ObjectStore, config *RepoConfig) *Repository{
	return &Repository{
		location: path,
		storage: store,
		config: config,
	}
}

func createConfig(configPath string, values map[string]string) (*RepoConfig, error) {
	config:= ini.Empty()
	coreSection, err := config.NewSection("core")
	if err!= nil {
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

	config.SaveTo(configPath)
	repoConfig :=RepoConfig{
		location: configPath,
	}

	return &repoConfig, nil
}

