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

type RepositoryType int8

const (
	NoRepo RepositoryType = iota
	BareRepo
	NonBareRepo
)

type Repository struct {
	gitDir     string
	workingDir string
	Storage    storage.ObjectStore
	Index      *Index
	Config     config.Config
	Info       *RepositoryInfo
	Refs       refs.RefManager
	Branches   Branches
}

func Init(path string, bare bool, storage storage.ObjectStore, refs refs.RefManager) (*Repository, error) {
	var gitDir, workingDir string
	if bare {
		gitDir = path
		workingDir = ""
	} else {
		gitDir = filepath.Join(path, ".git")
		workingDir = path
	}

	directories := []string{"hooks", "info", "refs/heads", "refs/tags"}
	files := map[string][]byte{
		"description": []byte("Unnamed repository; edit this file 'description' to name the repository.\n"),
		"HEAD":        []byte("ref: refs/heads/master"),
	}

	for _, directory := range directories {
		if err := os.MkdirAll(filepath.Join(gitDir, directory), os.ModeDir); err != nil {
			return nil, err
		}
	}

	for k, v := range files {
		if err := ioutil.WriteFile(filepath.Join(gitDir, k), v, 0644); err != nil {
			return nil, err
		}
	}

	cfg, err := createConfig(
		filepath.Join(gitDir, "config"), map[string]string{"bare": strconv.FormatBool(bare)})
	if err != nil {
		return nil, err
	}

	repo := NewRepo(workingDir, gitDir, storage, cfg, refs)
	return repo, nil
}

func InitDefault(path string, bare bool) (*Repository, error) {
	gitPath := createGitPaths(path, bare)
	repo, err := Init(path, bare, storage.NewFsStore(gitPath), refs.NewGitRefManager(gitPath))
	return repo, err
}

func NewRepo(workingDir string, gitDir string, store storage.ObjectStore, cfg config.Config, refs refs.RefManager) *Repository {
	index := NewIndex(workingDir, gitDir)

	r := &Repository{
		gitDir:     gitDir,
		workingDir: workingDir,
		Storage:    store,
		Index:      index,
		Config:     cfg,
		Refs:       refs,
		Branches:   NewBranches(refs),
	}

	r.Info = NewRepositoryInfo(r)
	return r
}

func FromExisting(path string) (*Repository, error) {
	if path == filepath.Dir(path) {
		return nil, fmt.Errorf("fatal: not a git repository (or any of the parent directories)")
	}

	repoType := isGitRepository(path)
	if repoType == NoRepo {
		return FromExisting(filepath.Dir(path))
	}

	var workingDir, gitDir string
	if repoType == BareRepo {
		workingDir = ""
		gitDir = path
	} else {
		workingDir = path
		gitDir = filepath.Join(path, ".git")
	}

	cfg, err := config.CreateDefaultConfigBuilder(filepath.Join(gitDir, "config"))
	if err != nil {
		return nil, err
	}

	refs := refs.NewGitRefManager(gitDir)

	return NewRepo(workingDir, gitDir, storage.NewFsStore(gitDir), cfg.Build(), refs), nil
}

func isGitRepository(path string) RepositoryType {
	candidate := filepath.Join(path, ".git")
	if _, err := os.Stat(candidate); err == nil {
		if isGitDirectory(candidate) {
			return NonBareRepo
		}
	}

	if isGitDirectory(path) {
		return BareRepo
	}

	return NoRepo
}

// see https://github.com/git/git/blob/e31aba42fb12bdeb0f850829e008e1e3f43af500/setup.c#L328-L394
func isGitDirectory(path string) bool {
	headPath := filepath.Join(path, "HEAD")
	if _, err := os.Stat(headPath); os.IsNotExist(err) {
		return false
	}

	objectsPath := filepath.Join(path, "objects")
	if _, err := os.Stat(objectsPath); os.IsNotExist(err) {
		return false
	}

	refsPath := filepath.Join(path, "refs")
	if _, err := os.Stat(refsPath); os.IsNotExist(err) {
		return false
	}

	return true
}

func createGitPaths(path string, bare bool) string {
	var gitDir string
	if bare {
		gitDir = path
	} else {
		gitDir = filepath.Join(path, ".git")
	}

	return gitDir
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
	ref, err := ry.Refs.Get("HEAD")
	if err != nil {
		return nil, err
	}

	return ref, nil
}

func (ry *Repository) SetHead(ref string) error {
	_, err := ry.Refs.Set("HEAD", ref)
	return err
}

func (ry *Repository) Commit(configure func(builder *objects.CommitBuilder) *objects.CommitBuilder) (*objects.Commit, error) {
	tree, err := objects.NewTreeFromDirectory(ry.gitDir, "")
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
