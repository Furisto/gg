package cmd

import (
	"bytes"
	"github.com/furisto/gog/objects"
	"github.com/furisto/gog/repo"
	"os"
	"testing"
)

const CommitMessage = "Test"
const CommmitOID = ""
const CommitTreeOID = ""

func TestCommit(t *testing.T) {
	r, err := prepareEnvForCommitTest()
	if err != nil {
		t.Fatalf("could not create test repository: %v", err)
	}
	defer os.RemoveAll(r.Location)

	options := CommitOptions{
		Path:    r.Location,
		Message: "Test",
	}

	output := bytes.Buffer{}
	cmd := NewCommitCmd(&output)
	commit, err := cmd.Execute(options)
	if err != nil {
		t.Fatalf("error encountered durign command execution: %v", err)
	}

	data, err := r.Storage.Get(commit.OID())
	if err != nil {
		t.Errorf("could not find expected commit %v", CommmitOID)
	}

	c, err := objects.DecodeCommit(commit.OID(), data)
	if err != nil {
		t.Errorf("encountered error during decoding of commit: %v", err)
	}

	if c.Author.Name != "furisto" || c.Author.Email != "furisto@test.com" {
		t.Errorf("commit does not contain the expected author name and/or email")
	}

	if c.Commiter.Name != "furisto" || c.Commiter.Email != "furisto@test.com" {
		t.Errorf("commit does not contain the expected committer name and/or email")
	}

	if len(c.Parents) != 0 {
		t.Errorf("expected %v parents for commit, but was %v", 0, len(commit.Parents))
	}

	if c.Message != CommitMessage {
		t.Errorf(
			"commit does not contain the expected message '%v', the message was '%v'", CommitMessage, commit.Message)
	}
}

func prepareEnvForCommitTest() (r *repo.Repository, err error) {
	r, err = CreateTestRepository()
	if err != nil {
		return nil, err
	}

	if err := populateRepo(r.Location); err != nil {
		return nil, err
	}

	return r, nil
}
