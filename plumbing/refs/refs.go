package refs

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// see https://github.com/git/git/blob/7f7ebe054af6d831b999d6c2241b9227c4e4e08d/refs.c#L520-L528
const (
	RefPattern    = "refs/"
	TagPattern    = RefPattern + "tags"
	BranchPattern = RefPattern + "heads"
	RemotePattern = RefPattern + "remotes"
)

var (
	ErrRefNotExist = errors.New("ref does not exist")
)

const refMarker = "ref: "

const (
	Head          = ""
	FetchHead     = ""
	OrigHead      = ""
	MergeHead     = ""
	CheryPickHead = ""
)

const (
	SymbolicRef RefType = iota
	HashRef
)

type RefType int8

type Ref struct {
	Name     string
	refType  RefType
	RefValue string
}

func DecodeRefFromFile(name, path string) (*Ref, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, ErrRefNotExist
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewRef(name, string(content))
}

func NewRef(name, value string) (*Ref, error) {
	if strings.HasPrefix(value, refMarker) {
		value := strings.TrimPrefix(value, refMarker)
		return NewSymbolicRef(name, value), nil
	}

	return NewHashRef(name, value), nil
}

func NewSymbolicRef(name, value string) *Ref {
	return &Ref{
		Name:     name,
		refType:  SymbolicRef,
		RefValue: value,
	}
}

func NewHashRef(name, value string) *Ref {
	return &Ref{
		Name:     name,
		refType:  HashRef,
		RefValue: value,
	}
}

func (r *Ref) IsRefType(refType RefType) bool {
	return r.refType == refType
}

type RefManager interface {
	Set(name string, target string) (*Ref, error)
	Get(name string) (*Ref, error)
	List(prefix string) []*Ref
	Delete(name string) error
	Find(ref string) (*Ref, error)
	Resolve(ref *Ref) (*Ref, error)
}

type GitRefManager struct {
	gitDir  string
	refDirs []string
}

func NewGitRefManager(gitDir string) *GitRefManager {
	return &GitRefManager{
		gitDir: gitDir,
		refDirs: []string{
			gitDir,
			path.Join(gitDir, RefPattern),
			path.Join(gitDir, TagPattern),
			path.Join(gitDir, BranchPattern),
			path.Join(gitDir, RemotePattern),
		},
	}
}

func (grm *GitRefManager) Set(name, target string) (*Ref, error) {
	ref, err := NewRef(name, target)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filepath.Join(grm.gitDir, ref.Name), []byte(ref.RefValue), 0644)
	if err != nil {
		return nil, err
	}

	return ref, nil
}

func (grm *GitRefManager) Get(name string) (*Ref, error) {
	refPath := filepath.Join(grm.gitDir, name)
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		return nil, ErrRefNotExist
	}

	return DecodeRefFromFile(name, refPath)
}

func (grm *GitRefManager) List(suffix string) []*Ref {
	var refs []*Ref
	suffix = strings.TrimSuffix(suffix, "/")

	for _, dir := range grm.refDirs {
		if !strings.HasSuffix(dir, suffix) {
			continue
		}

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, f := range files {
			ref, err := DecodeRefFromFile(path.Join(suffix, f.Name()), path.Join(dir, f.Name()))
			if err != nil {
				continue
			}

			refs = append(refs, ref)
		}
	}

	return refs
}

func (grm *GitRefManager) Delete(name string) error {
	refPath := filepath.Join(grm.gitDir, name)
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		return ErrRefNotExist
	}

	if err := os.Remove(refPath); err != nil {
		return err
	}

	return nil
}

// see https://mirrors.edge.kernel.org/pub/software/scm/git/docs/gitrevisions.html
func (grm *GitRefManager) Find(ref string) (*Ref, error) {
	for _, refDir := range append([]string{grm.gitDir}, grm.refDirs...) {
		_, err := os.Stat(filepath.Join(refDir, ref))
		if err == nil {
			return DecodeRefFromFile(ref, filepath.Join(grm.gitDir, ref))
		}
	}

	return nil, ErrRefNotExist
}

// follows symbolic refs until it reaches a hash ref
func (grm *GitRefManager) Resolve(ref *Ref) (*Ref, error) {
	if ref.IsRefType(HashRef) {
		return ref, nil
	}

	refPath := filepath.Join(grm.gitDir, ref.RefValue)
	newRef, err := DecodeRefFromFile(ref.RefValue, refPath)
	if err != nil {
		return nil, err
	}

	return grm.Resolve(newRef)
}
