package repo

import (
	"errors"
	"github.com/furisto/gog/plumbing/refs"
	"path/filepath"
)

var (
	ErrEmptyBranchNameNotAllowed = errors.New("empty branch name not allowed")
	ErrBranchAlreadyExists       = errors.New("branch already exists")
)

type Branches struct {
	prefix string
	refs   refs.RefManager
}

func NewBranches(refs refs.RefManager) Branches {
	return Branches{
		prefix: "/refs/heads",
		refs:   refs,
	}
}

func (b *Branches) Create(name string, value string) (*refs.Ref, error) {
	if name == "" {
		return nil, ErrEmptyBranchNameNotAllowed
	}

	return b.refs.Set(filepath.Join(b.prefix, name), value)
}

func (b *Branches) Get(name string) (*refs.Ref, error) {
	if name == "" {
		return nil, ErrEmptyBranchNameNotAllowed
	}

	return b.refs.Get(filepath.Join(b.prefix, name))
}

func (b *Branches) List() []*refs.Ref {
	return b.refs.List(b.prefix)
}

func (b *Branches) Update(name string, value *refs.Ref) error {
	if name == "" {
		return ErrEmptyBranchNameNotAllowed
	}

	_, err := b.refs.Set(filepath.Join(b.prefix, name), value.Name)
	return err
}

func (b *Branches) Delete(name string) error {
	if name == "" {
		return ErrEmptyBranchNameNotAllowed
	}

	return b.refs.Delete(filepath.Join(b.prefix, name))
}
