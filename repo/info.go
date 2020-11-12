package repo

import (
	"github.com/furisto/gog/config"
	"github.com/furisto/gog/plumbing/refs"
	"strconv"
)

type RepositoryInfo struct {
	repo *Repository
}

func NewRepositoryInfo(ry *Repository) *RepositoryInfo {
	return &RepositoryInfo{
		repo: ry,
	}
}

func (ryi *RepositoryInfo) GitDirectory() string {
	return ryi.repo.gitDir
}

func (ryi *RepositoryInfo) WorkingDirectory() string {
	return ryi.repo.workingDir
}

func (ryi *RepositoryInfo) IsBare() (bool, error) {
	bareStr, err := ryi.repo.Config.Get("core", "bare")
	if err != nil {
		if err == config.ErrUnknownKey {
			return false, nil
		}
		return false, err
	}

	return strconv.ParseBool(bareStr)
}

func (ryi *RepositoryInfo) IsHeadDetached() (bool, error) {
	head, err := ryi.repo.Head()
	if err != nil {
		return false, err
	}

	return head.IsRefType(refs.HashRef), nil
}

func (ryi *RepositoryInfo) IsHeadUnborn() (bool, error) {
	head, err := ryi.repo.Head()
	if err != nil {
		return false, err
	}

	if head.IsRefType(refs.HashRef) {
		return false, nil
	}

	_, err = ryi.repo.Refs.Resolve(head)
	return err == refs.ErrRefNotExist, nil
}
