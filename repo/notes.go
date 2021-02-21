package repo

import (
	"fmt"
	"github.com/furisto/gog/config"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/plumbing/refs"
	"github.com/furisto/gog/storage"
)

const commitMessageForNotes = "Notes added by 'git notes add'"

type Notes struct {
	noteRef string
	refMgr  refs.RefManager
	store   storage.ObjectStore
	config  config.Config
}

func NewNotes(refMgr refs.RefManager, store storage.ObjectStore, config config.Config, noteRef string) Notes {
	return Notes{
		noteRef: "refs/notes/" + noteRef,
		refMgr:  refMgr,
		store:   store,
		config:  config,
	}
}

func (n *Notes) Create(message, commitOid string, force bool) (*Note, error) {
	commitData, err := n.store.Get(commitOid)
	if err != nil {
		return nil, err
	}

	if !objects.IsCommit(commitData) {
		return nil, fmt.Errorf("notes only support commits")
	}

	noteRef, err := n.refMgr.Get(n.noteRef)
	if err != nil && err != refs.ErrRefNotExist {
		return nil, err
	}

	var oldNoteHead *objects.Commit
	treeBuilder := objects.NewTreeBuilder()
	if noteRef != nil {
		noteRef, err = n.refMgr.Resolve(noteRef)
		if err != nil {
			return nil, err
		}

		oldNoteTree, err := n.retrieveNoteHeadTree()
		if err != nil {
			return nil, err
		}

		for _, entry := range oldNoteTree.Entries() {
			if entry.Name == commitOid && !force {
				return nil, fmt.Errorf("cannot overwrite existing note")
			}

			treeBuilder.AddBlob(entry.OID, entry.Name, entry.Mode)
		}
	}

	messageBlob := objects.NewBlob([]byte(message))
	treeBuilder.AddBlob(messageBlob.OID(), commitOid, 0o100644)
	noteTree := treeBuilder.Build()

	if err := messageBlob.Save(n.store); err != nil {
		return nil, err
	}
	if err := noteTree.Save(n.store); err != nil {
		return nil, err
	}

	cb := objects.NewCommitBuilder(noteTree.OID()).
		WithConfig(n.config).
		WithMessage(commitMessageForNotes)
	if oldNoteHead != nil {
		cb.WithParent(oldNoteHead.OID())
	}

	newNoteHead, err := cb.Build()
	if err != nil {
		return nil, err
	}

	if err := newNoteHead.Save(n.store); err != nil {
		return nil, err
	}

	_, err = n.refMgr.Set(n.noteRef, newNoteHead.OID())
	if err != nil {
		return nil, err
	}

	note := newNote(commitOid, messageBlob.OID(), n.store)
	return &note, nil
}

func (n *Notes) Append(commitOid, message string) (*Note, error) {
	oldNoteHead, err := n.refMgr.Get(n.noteRef)
	if err != nil {
		return nil, fmt.Errorf("note for commit %s does not exist", commitOid)
	}

	noteTree, err := n.retrieveNoteHeadTree()
	if err != nil {
		return nil, err
	}

	treeBuilder := objects.NewTreeBuilder()
	var targetEntry *objects.TreeEntry
	for _, entry := range noteTree.Entries() {
		if entry.Name == commitOid {
			targetEntry = &entry
			continue
		}

		treeBuilder.AddBlob(entry.OID, entry.Name, entry.Mode)
	}

	if targetEntry == nil {
		return nil, fmt.Errorf("")
	}

	oldBlobData, err := n.store.Get(targetEntry.OID)
	if err != nil {
		return nil, err
	}

	oldBlob, err := objects.LoadBlob(oldBlobData)
	if err != nil {
		return nil, err
	}

	modifiedBlob := objects.NewBlob(append(oldBlob.Content, "\n\n"+message...))
	if err := modifiedBlob.Save(n.store); err != nil {
		return nil, err
	}

	treeBuilder.AddBlob(modifiedBlob.OID(), commitOid, 0o100644)
	modifiedNodeTree := treeBuilder.Build()
	if err := modifiedNodeTree.Save(n.store); err != nil {
		return nil, err
	}

	newNoteHead, err := objects.NewCommitBuilder(modifiedNodeTree.OID()).
		WithConfig(n.config).
		WithMessage(commitMessageForNotes).
		WithParent(oldNoteHead.RefValue).
		Build()

	if err != nil {
		return nil, err
	}

	if err := newNoteHead.Save(n.store); err != nil {
		return nil, err
	}

	_, err = n.refMgr.Set(n.noteRef, newNoteHead.OID())
	note := newNote(commitOid, modifiedBlob.OID(), n.store)
	return &note, err
}

func (n *Notes) List(objectRef string) ([]*Note, error) {
	noteTree, err := n.retrieveNoteHeadTree()
	if err != nil {
		return nil, err
	}

	var notes []*Note
	for _, entry := range noteTree.Entries() {
		if objectRef != "" && objectRef != entry.Name {
			continue
		}
		n := newNote(entry.Name, entry.OID, n.store)
		notes = append(notes, &n)
	}

	return notes, nil
}

func (n *Notes) Copy(sourceCommit, targetCommit string, force bool) (*Note, error) {
	if _, err := n.store.Get(targetCommit); err != nil {
		return nil, fmt.Errorf("failed to resolve %s as valid ref", targetCommit)
	}

	_, err := n.refMgr.Get(n.noteRef)
	if err != nil {
		return nil, err
	}

	noteTree, err := n.retrieveNoteHeadTree()
	if err != nil {
		return nil, err
	}

	sourceEntry, ok := noteTree.GetEntryByName(sourceCommit)
	if !ok {
		return nil, fmt.Errorf("missing notes on source object %s. cannot copy.", sourceCommit)
	}

	sourceBlobData, err := n.store.Get(sourceEntry.OID)
	if err != nil {
		return nil, err
	}

	sourceBlob, err := objects.LoadBlob(sourceBlobData)
	if err != nil {
		return nil, err
	}

	return n.Create(string(sourceBlob.Content), targetCommit, force)
}

func (n *Notes) Find(commitOid string) (*Note, error) {
	_, err := n.refMgr.Get(n.noteRef)
	if err != nil {
		return nil, nil
	}

	noteTree, err := n.retrieveNoteHeadTree()
	if err != nil {
		return nil, err
	}

	for _, entry := range noteTree.Entries() {
		if entry.Name == commitOid {
			note := newNote(entry.Name, entry.OID, n.store)
			return &note, nil
		}
	}

	return nil, nil
}

func (n *Notes) Remove(commitOid string) error {
	noteHeadRef, err := n.refMgr.Get(n.noteRef)
	if err != nil {
		return nil
	}

	noteTree, err := n.retrieveNoteHeadTree()
	if err != nil {
		return err
	}

	treeBuilder := objects.NewTreeBuilder()
	noteExists := false
	for _, entry := range noteTree.Entries() {
		if entry.Name == commitOid {
			noteExists = true
			continue
		}
		treeBuilder.AddBlob(entry.OID, entry.Name, entry.Mode)
	}

	if !noteExists {
		return fmt.Errorf("object %s has no note")
	}

	newNoteTree := treeBuilder.Build()
	newNoteHead, err := objects.NewCommitBuilder(newNoteTree.OID()).
		WithMessage("notes").
		WithParent(noteHeadRef.RefValue).
		Build()
	if err != nil {
		return err
	}

	if err := newNoteTree.Save(n.store); err != nil {
		return err
	}

	if err := newNoteHead.Save(n.store); err != nil {
		return err
	}

	_, err = n.refMgr.Set(noteHeadRef.Name, newNoteHead.OID())
	return err
}

func (n *Notes) Head() *objects.Commit {
	noteHead, err := n.retrieveNoteHead()
	if err != nil {
		return nil
	} else {
		return noteHead
	}
}

func (n *Notes) retrieveNoteHeadTree() (*objects.Tree, error) {
	noteHead, err := n.retrieveNoteHead()
	if err != nil {
		return nil, err
	}

	noteTreeData, err := n.store.Get(noteHead.Tree)
	if err != nil {
		return nil, err
	}

	return objects.LoadTree(noteTreeData)
}

func (n *Notes) retrieveNoteHead() (*objects.Commit, error) {
	noteHeadRef, err := n.refMgr.Get(n.noteRef)
	if err != nil {
		return nil, err
	}

	noteHeadData, err := n.store.Get(noteHeadRef.RefValue)
	if err != nil {
		return nil, err
	}

	noteHead, err := objects.DecodeCommit(noteHeadRef.RefValue, noteHeadData)
	if err != nil {
		return nil, err
	}

	return noteHead, nil
}

type Note struct {
	CommitOID  string
	MessageOID string
	store      storage.ObjectStore
}

func newNote(commit, messageBlob string, store storage.ObjectStore) Note {
	return Note{
		CommitOID:  commit,
		MessageOID: messageBlob,
		store:      store,
	}
}

func (n *Note) Commit() (*objects.Commit, error) {
	commit, err := LoadCommit(n.store, n.CommitOID)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func (n *Note) Message() (string, error) {
	blob, err := LoadBlob(n.store, n.MessageOID)
	if err != nil {
		return "", err
	}
	return string(blob.Content), nil
}
