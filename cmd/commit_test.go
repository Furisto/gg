package cmd

import (
	"bytes"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/repo"
	"os"
	"testing"
)

const CommitMessage = "Test"
const ParentCommit = "48743154a35f5751796d39ebceb615453abac8de"

func TestFirstCommitInRepository(t *testing.T) {
	r, err := prepareEnvWithNoCommmits()
	if err != nil {
		t.Fatalf("could not create test repository: %v", err)
	}
	defer os.RemoveAll(r.Location)

	commit := createCommit(r, t)
	checkCommit(commit, t, r)
}

func TestSubsequentCommitInRepository(t *testing.T) {
	r, err := prepareEnvWithCommits()
	if err != nil {
		t.Fatalf("could not create test repository: %v", err)
	}
	defer os.RemoveAll(r.Location)

	commit := createCommit(r, t)
	checkCommit(commit, t, r)
}

func prepareEnvWithNoCommmits() (r *repo.Repository, err error) {
	r, err = CreateTestRepository()
	if err != nil {
		return nil, err
	}

	if err := populateRepo(r.Location); err != nil {
		return nil, err
	}

	if err := r.Config.Set("user", "name", "furisto"); err != nil {
		return nil, err
	}
	if err := r.Config.Set("user", "email", "furisto@test.com"); err != nil {
		return nil, err
	}

	return r, nil
}

func prepareEnvWithCommits() (r *repo.Repository, err error) {
	r, err = prepareEnvWithNoCommmits()
	if err != nil {
		return nil, err
	}

	if _, err := r.Branches.Create("master", ParentCommit); err != nil {
		return nil, err
	}

	return r, nil
}

func createCommit(r *repo.Repository, t *testing.T) *objects.Commit {
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

	return commit
}

func checkCommit(commit *objects.Commit, t *testing.T, r *repo.Repository) {
	data, err := r.Storage.Get(commit.OID())
	if err != nil {
		t.Errorf("could not find expected commit %v", commit.OID())
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

	master, err := r.Branches.Get("master")
	if err != nil {
		t.Errorf("master branch has not been created")
		return
	}

	if master.RefValue != commit.OID() {
		t.Errorf("master branch does not have the expected commit of %v", commit.OID())
	}
}
