package repo

import (
	"fmt"
	"github.com/furisto/gog/config"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/plumbing/refs"
	"github.com/furisto/gog/storage"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

var defaultConfigValues = map[string]string{
	"repositoryformatversion": "0",
	"filemode":                "false",
	"symlinks":                "false",
	"ignorecase":              "true",
}

type Repository struct {
	Location string
	Storage  storage.ObjectStore
	Config   config.Config
	Refs     refs.RefManager
	Branches Branches
}

func Init(path string, bare bool, storage storage.ObjectStore, refs refs.RefManager) (*Repository, error) {
	var repoPath string
	if bare {
		repoPath = path
	} else {
		repoPath = filepath.Join(path, ".git")
	}

	directories := []string{"hooks", "info", "refs/heads", "refs/tags"}
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

	cfg, err := createConfig(
		filepath.Join(repoPath, "config"), map[string]string{"bare": strconv.FormatBool(bare)})
	if err != nil {
		return nil, err
	}

	repo := NewRepo(filepath.Dir(repoPath), storage, cfg, refs)
	return repo, nil
}

func InitDefault(path string, bare bool) (*Repository, error) {
	repo, err := Init(path, bare, storage.NewFsStore(filepath.Join(path, ".git")), refs.NewGitRefManager(""))
	return repo, err
}

func NewRepo(path string, store storage.ObjectStore, cfg config.Config, refs refs.RefManager) *Repository {
	return &Repository{
		Location: path,
		Storage:  store,
		Config:   cfg,
		Refs:     refs,
		Branches: NewBranches(refs),
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

	cfg, err := config.CreateDefaultConfigBuilder(filepath.Join(path, ".git", "config"))
	if err != nil {
		return nil, err
	}

	refs := refs.NewGitRefManager(gitPath)

	return &Repository{
		Location: path,
		Storage:  storage.NewFsStore(gitPath),
		Config:   cfg.Build(),
		Refs:     refs,
		Branches: NewBranches(refs),
	}, nil
}

func createConfig(configPath string, values map[string]string) (config.Config, error) {
	cb, err := config.CreateDefaultConfigBuilder(configPath)
	if err != nil {
		return nil, err
	}
	cfg := cb.Build()

	for k, v := range defaultConfigValues {
		if err := cfg.Set("core", k, v); err != nil {
			return nil, err
		}
	}

	for k, v := range values {
		if err := cfg.Set("core", k, v); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (ry *Repository) Head() (*refs.Ref, error) {
	ref, err := ry.Refs.Get("head")
	if err != nil {
		return nil, err
	}

	return ref, nil
}

func (ry *Repository) SetHead(ref *refs.Ref) error {
	_, err := ry.Refs.Set("head", ref.Name)
	return err
}

func (ry *Repository) Commit(configure func(builder *objects.CommitBuilder) *objects.CommitBuilder) (*objects.Commit, error) {
	tree, err := objects.NewTreeFromDirectory(ry.Location, "")
	if err != nil {
		return nil, err
	}
	if err := tree.Save(ry.Storage); err != nil {
		return nil, err
	}

	headRef, err := ry.Head()
	if err != nil {
		return nil, err
	}
	parentRef, err := ry.Refs.Resolve(headRef)
	if err != nil {
		if err != refs.ErrRefNotExist {
			return nil, err
		}

		_, err := ry.Branches.Get("master")
		if err != refs.ErrRefNotExist {
			return nil, err
		}

		master, err := ry.Branches.Create("master", "")
		if err != nil {
			return nil, err
		}

		parentRef = master
	}

	builder := objects.NewCommitBuilder(tree.OID()).
		WithConfig(ry.Config).
		WithParent(parentRef.RefValue)

	builder = configure(builder)
	commit, err := builder.Build()

	if err != nil {
		return nil, err
	}

	err = commit.Save(ry.Storage)
	if err != nil {
		return nil, err
	}

	if _, err := ry.Refs.Set(parentRef.Name, commit.OID()); err != nil {
		return nil, err
	}

	return commit, nil
}
