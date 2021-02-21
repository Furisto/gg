package cmd

import (
	"bytes"
	"fmt"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/repo"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNoteAddNoPredecessor(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)

	options := NotesAddCmdOptions{
		Force:        false,
		Message:      "notes test",
		TargetObject: commits[0].OID(),
	}
	options.Path = ry.Info.WorkingDirectory()

	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteAdd(options); err != nil {
		t.Fatalf("error occured while adding note to commit: %v", err)
	}

	noteDir := filepath.Join(ry.Info.GitDirectory(), "refs/notes")
	assert.DirExists(t, noteDir)
	assert.FileExists(t, filepath.Join(noteDir, "commits"))

	noteHead := ry.Notes.Head()
	if noteHead == nil {
		t.Fatal("created note but no note head does not have a commit")
	}
}

func TestNoteAddWithOverwrite(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	oldNote, err := ry.Notes.Create("first note", commits[0].OID(), false)
	if err != nil {
		t.Fatalf("could not create note: %v", err)
	}

	options := NotesAddCmdOptions{
		Force:        true,
		Message:      "notes test",
		TargetObject: commits[0].OID(),
	}
	options.Path = ry.Info.WorkingDirectory()

	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteAdd(options); err != nil {
		t.Fatalf("error occured while adding note to commit: %v", err)
	}

	noteDir := filepath.Join(ry.Info.GitDirectory(), "refs/notes")
	assert.DirExists(t, noteDir)
	assert.FileExists(t, filepath.Join(noteDir, "commits"))

	assert.NotEqual(t, oldNote.CommitOID, ry.Notes.Head().OID())
}

func TestNoteAddToExistingWithoutOverwrite(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	if _, err := ry.Notes.Create("first note", commits[0].OID(), false); err != nil {
		t.Fatalf("could not create note: %v", err)
	}

	options := NotesAddCmdOptions{
		Force:        false,
		Message:      "notes test",
		TargetObject: commits[0].OID(),
	}
	options.Path = ry.Info.WorkingDirectory()

	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteAdd(options); err == nil {
		t.Fatalf("no error occured while adding note to commit: %v", err)
	}
}

func TestCopyNote(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	oldNote, err := ry.Notes.Create("first note", commits[0].OID(), false)
	if err != nil {
		t.Fatalf("could not create note")
	}

	options := NotesCopyCmdOptions{
		Force:      false,
		FromObject: commits[0].OID(),
		ToObject:   commits[1].OID(),
	}
	options.Path = ry.Info.WorkingDirectory()

	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteCopy(options); err != nil {
		t.Fatalf("error occured while copying note from one commit to another: %v", err)
	}

	assertCopyNote(t, ry, oldNote.CommitOID, commits[1].OID())
}

func assertCopyNote(t *testing.T, ry *repo.Repository, from, to string) {
	assert.NotEqual(t, from, ry.Notes.Head().OID())
	noteTree, err := repo.LoadTreeFromCommit(ry.Storage, ry.Notes.Head())
	if err != nil {
		t.Fatalf("could not load note tree: %v", err)
	}

	assert.Equal(t, 2, len(noteTree.Entries()))
	messageEntry, ok := noteTree.GetEntryByName(to)
	assert.True(t, ok)

	messageBlob, err := repo.LoadBlobFromTreeEntry(ry.Storage, &messageEntry)
	if err != nil {
		t.Fatalf("could not load message blob %s: %v", messageEntry.OID, err)
	}

	assert.Equal(t, []byte("first note"), messageBlob.Content)
}

func TestCopyNoteWithOverwrite(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	firstNote, err := ry.Notes.Create("first note", commits[0].OID(), false)
	if err != nil {
		t.Fatalf("could not create note")
	}

	secondNote, err := ry.Notes.Create("second note", commits[1].OID(), false)
	if err != nil {
		t.Fatalf("could not create note")
	}

	options := NotesCopyCmdOptions{
		Force:      true,
		FromObject: commits[0].OID(),
		ToObject:   commits[1].OID(),
	}
	options.Path = ry.Info.WorkingDirectory()

	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteCopy(options); err != nil {
		t.Fatalf("error occured while copying note from one commit to another: %v", err)
	}

	assertCopyNote(t, ry, firstNote.CommitOID, secondNote.CommitOID)
}

func TestNoteAppend(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetCommit := commits[0]

	_, err := ry.Notes.Create("notes test", targetCommit.OID(), false)
	if err != nil {
		t.Fatalf("could not create note for commit %s: %v", targetCommit.OID(), err)
	}
	oldNoteHead := ry.Notes.Head()

	options := NotesAppendCmdOptions{
		Message:   "notes append test",
		ObjectRef: targetCommit.OID(),
	}
	options.Path = ry.Info.WorkingDirectory()

	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteAppend(options); err != nil {
		t.Fatalf("could not append note to commit %s: %v", targetCommit.OID(), err)
	}

	assert.NotEqual(t, oldNoteHead.OID(), ry.Notes.Head().OID())
	targetNote, err := ry.Notes.Find(targetCommit.OID())
	if err != nil || targetNote == nil {
		t.Fatalf("could not find note for commit %s", targetCommit.OID())
	}

	expected := "notes test\n\nnotes append test"
	actual, err := targetNote.Message()
	if err != nil {
		t.Fatalf("could not retrieve message from note: %v", err)
	}
	assert.Equal(t, expected, actual)
}

func TestNoteShow(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetCommit := commits[0]
	if _, err := ry.Notes.Create("notes test", targetCommit.OID(), false); err != nil {
		t.Fatalf("could not create note for commit %s", targetCommit.OID())
	}

	options := NotesShowCmdOptions{
		ObjectRef: targetCommit.OID(),
	}
	options.Path = ry.Info.WorkingDirectory()

	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteShow(options); err != nil {
		t.Fatalf("could not show note for commit %s: %v", targetCommit.OID(), err)
	}

	assert.Equal(t, "notes test", output.String())
}

func TestListingOfNotes(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	addNotesToCommits(t, ry, commits)

	options := NotesListCmdOptions{
		ObjectRef: "",
	}
	options.Path = ry.Info.WorkingDirectory()
	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteList(options); err != nil {
		t.Fatalf("could not execute note list: %v", err)
	}

	actual := strings.FieldsFunc(output.String(), func(c rune) bool { return c == '\n' })
	expected := []string{
		"105bc203aa908f9bde05cdd3d83d4942ab876452 52337f8c0b9c3a405f61e774e110449667f02c48",
		"b58359108b45a89fda850ec75dac57d4d04f2a8a 8c2e614ca354d8d72a938d9f03720251bfdf99b9",
		"9eeccd72f72437225ecd8e19b12c4a3f2f2ed02a a86c9bb220ca126890e1e4e4ecd18d9fc79c057d",
		"c6706482e0a6e6b27709bc27873a83e5dfb9e601 b9ba344cda497e93f8598392e345eaa749dab27d",
		"4f01e2291ceb1724d11a9bac61ef5f43fc153267 ff6d4f1e76027c68a784f1af7ae1235aa40f8ad8",
	}

	assert.ElementsMatch(t, expected, actual)
}

func TestListingOfNotesWithFilter(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetCommit := commits[1].OID()
	addNotesToCommits(t, ry, commits)

	options := NotesListCmdOptions{
		ObjectRef: targetCommit,
	}
	options.Path = ry.Info.WorkingDirectory()

	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteList(options); err != nil {
		t.Fatalf("could not execute note list: %v", err)
	}

	actual := strings.FieldsFunc(output.String(), func(c rune) bool { return c == '\n' })
	expected := []string{
		"b58359108b45a89fda850ec75dac57d4d04f2a8a 8c2e614ca354d8d72a938d9f03720251bfdf99b9",
	}

	assert.ElementsMatch(t, expected, actual)
}

func TestRemoveExistingNote(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetCommit := commits[0]
	if _, err := ry.Notes.Create("note remove", targetCommit.OID(), false); err != nil {
		t.Fatalf("could not create note for commit %s", targetCommit.OID())
	}

	options := NotesRemoveCmdOptions{
		ObjectRef: targetCommit.OID(),
	}
	options.Path = ry.Info.WorkingDirectory()

	output := executeNoteRemove(t, options)
	assert.Equal(t, fmt.Sprintf("Removing note for object %s", targetCommit.OID()), output.String())
}

func TestRemoveFromObjectRefWithoutNote(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetCommit := commits[0]

	options := NotesRemoveCmdOptions{
		ObjectRef: targetCommit.OID(),
	}
	options.Path = ry.Info.WorkingDirectory()

	output := executeNoteRemove(t, options)
	assert.Equal(t, fmt.Sprintf("object %s has no note", targetCommit.OID()), output.String())
}

func TestRemoveFromNonExistantObjectRef(t *testing.T) {
	ry, _ := prepareEnvWithCommitObjects(t)
	targetCommit := "89szdhf"

	options := NotesRemoveCmdOptions{
		ObjectRef: targetCommit,
	}
	options.Path = ry.Info.WorkingDirectory()

	output := executeNoteRemove(t, options)
	assert.Equal(t, fmt.Sprintf("failed to resolve '%s' as a valid ref", targetCommit), output.String())
}

func executeNoteRemove(t *testing.T, options NotesRemoveCmdOptions) *bytes.Buffer {
	t.Helper()

	output := new(bytes.Buffer)
	cmd := NewNotesCmd(output)
	if err := cmd.ExecuteRemove(options); err != nil {
		t.Fatalf("could not execute note remove: %v", err)
	}
	return output
}

func addNotesToCommits(t *testing.T, ry *repo.Repository, commits []*objects.Commit) []*repo.Note {
	t.Helper()

	notes := make([]*repo.Note, 0, 5)
	for i, c := range commits {
		n, err := ry.Notes.Create("note"+strconv.Itoa(i), c.OID(), false)
		if err != nil {
			t.Fatalf("could not create note %d for commit %s", i, c.OID())
		}
		notes = append(notes, n)
	}

	return notes
}
