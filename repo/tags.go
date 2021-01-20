package repo

import (
	"errors"
	"fmt"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/plumbing/refs"
	"github.com/furisto/gog/storage"
)

var (
	ErrEmptyTag         = errors.New("empty tag name is not allowed")
	ErrTagAlreadyExists = errors.New("tag already exists")
	ErrInvalidTagTarget = errors.New("tag target is invalid")
)

type Tags struct {
	refMgr refs.RefManager
	store  storage.ObjectStore
	prefix string
}

func NewTags(refMgr refs.RefManager, store storage.ObjectStore) Tags {
	return Tags{
		refMgr: refMgr,
		store:  store,
		prefix: "refs/tags/",
	}
}

func (t *Tags) CreateAnnotated(tagName, target string, tagger *objects.Signature, message string, overwrite bool) (*objects.Tag, error) {
	if tagName == "" || tagger == nil || message == "" {
		fmt.Errorf("invalid paramters")
	}

	fullTagName := t.prefix + tagName
	if !overwrite {
		if _, err := t.Get(tagName); err == nil {
			return nil, ErrTagAlreadyExists
		}
	}

	data, err := t.store.Get(target)
	if err != nil {
		return nil, ErrInvalidTagTarget
	}

	targetType, err := objects.GetObjectType(data)
	if err != nil {
		return nil, err
	}

	tag, err := objects.NewTag(target, targetType, tagName, tagger, message)
	if err != nil {
		return nil, err
	}
	tagRef, err := t.refMgr.Set(fullTagName, tag.OID())
	if err != nil {
		return nil, err
	}

	if err := tag.Save(t.store); err != nil {
		_ = t.refMgr.Delete(tagRef.Name)
		return nil, err
	}

	return tag, nil
}

func (t *Tags) CreateLightweight(tagName, target string, overwrite bool) (*refs.Ref, error) {
	if tagName == "" || target == "" {
		return nil, fmt.Errorf("invalid parameters")
	}

	fullTagName := t.prefix + tagName
	if !overwrite {
		if _, err := t.Get(tagName); err == nil {
			return nil, fmt.Errorf("tag %s already exists", tagName)
		}
	}

	tagRef, err := t.refMgr.Set(fullTagName, target)
	if err != nil {
		return nil, err
	}

	return tagRef, nil
}

func (t *Tags) Get(tagName string) (*refs.Ref, error) {
	return t.refMgr.Get(t.prefix + tagName)
}

func (t *Tags) List() []*refs.Ref {
	return t.refMgr.List(t.prefix)
}

func (t *Tags) Delete(tagName string) error {
	_, err := t.refMgr.Get(t.prefix + tagName)
	if err != nil {
		return err
	}

	return t.refMgr.Delete(t.prefix + tagName)
}
