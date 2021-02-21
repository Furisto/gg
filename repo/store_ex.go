package repo

import (
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/storage"
)

func LoadTreeFromCommit(store storage.ObjectStore, commit *objects.Commit) (*objects.Tree, error) {
	treeData, err := store.Get(commit.Tree)
	if err != nil {
		return nil, err
	}

	return objects.LoadTree(treeData)
}

func LoadBlobFromTreeEntry(store storage.ObjectStore, treeEntry *objects.TreeEntry) (*objects.Blob, error) {
	blobData, err := store.Get(treeEntry.OID)
	if err != nil {
		return nil, err
	}

	return objects.LoadBlob(blobData)
}

func LoadBlob(store storage.ObjectStore, oid string) (*objects.Blob, error) {
	blobData, err := store.Get(oid)
	if err != nil {
		return nil, err
	}

	return objects.LoadBlob(blobData)
}

func LoadCommit(store storage.ObjectStore, oid string) (*objects.Commit, error) {
	commitData, err := store.Get(oid)
	if err != nil {
		return nil, err
	}

	return objects.DecodeCommit(oid, commitData)
}
