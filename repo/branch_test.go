package repo

import (
	"strconv"
	"testing"
)

const (
	Prod  = "prod"
	Stage = "stage"
)

func TestCreateBranch(t *testing.T) {
	ry := prepareEnvWithCommits(t)

	for i := 0; i < 3; i++ {
		branchName := "branch" + strconv.Itoa(i)
		branchValue := "1234"

		b, err := ry.Branches.Create(branchName, branchValue)
		if err != nil {
			t.Errorf("could not create branch: %v", err)
		}

		if b.Name != "/refs/heads/"+branchName {
			t.Errorf("branch ref does not have the expected name. expected %v but was %v", branchName, b.Name)
		}

		if b.RefValue != branchValue {
			t.Errorf("branch ref does not have the expected value. expected %v but was %v", branchValue, b.RefValue)
		}
	}
}

func TestListBranch(t *testing.T) {
	ry := prepareEnvWithBranches(t)

	expected := []struct {
		BranchName  string
		BranchValue string
	}{
		{
			"/refs/heads/dev",
			"1234",
		},
		{
			"/refs/heads/master",
			"1234",
		},
		{
			"/refs/heads/prod",
			"1234",
		},
	}

	actual := ry.Branches.List()

	if len(expected) != len(actual) {
		t.Errorf("expected number of branches returned was %d, but actual number was %d", len(expected), len(actual))
	}

	for i := range expected {
		if expected[i].BranchName != actual[i].Name {
			t.Errorf("expected branch name of %s, but was %s", expected[i].BranchName, actual[i].Name)
		}

		if expected[i].BranchValue != actual[i].RefValue {
			t.Errorf("expected branch value of %s, but was %s", expected[i].BranchValue, actual[i].RefValue)
		}
	}
}

func TestDeleteBranch(t *testing.T) {
	ry := prepareEnvWithBranches(t)

	if err := ry.Branches.Delete("prod"); err != nil {
		t.Errorf("could not delete branch: %v", err)
		return
	}

	expected := []string{"/refs/heads/dev", "/refs/heads/master"}
	actual := ry.Branches.List()

	if len(expected) != len(actual) {
		t.Errorf("expected %d branches after deletion, but found %d branches", len(expected), len(actual))
		return
	}

	for i := range expected {
		if expected[i] != actual[i].Name {
			t.Errorf("expected branch name of %s, but got %s", expected[i], actual[i].Name)
		}
	}
}

func TestDeleteNonExistentBranch(t *testing.T) {
	ry := prepareEnvWithBranches(t)

	if err := ry.Branches.Delete(Stage); err == nil {
		t.Errorf("no error was returned while trying to delete non existent branch")
	}
}

func TestBranchForDeleteNotSpecified(t *testing.T) {
	ry := prepareEnvWithBranches(t)

	if err := ry.Branches.Delete(""); err == nil {
		t.Errorf("no error was returned while trying to delete non existent branch")
	}
}

func TestRenameBranch(t *testing.T) {
	// arrange
	ry := prepareEnvWithBranches(t)

	branchCount := len(ry.Branches.List())
	prodBranch, err := ry.Branches.Get(Prod)
	if err != nil {
		t.Fatalf("could not find branch %q", Prod)
	}

	// act
	if err := ry.Branches.Rename(Prod, Stage); err != nil {
		t.Errorf("could not rename branch: %v", err)
	}

	// assert
	if branchCount != len(ry.Branches.List()) {
		t.Errorf("number of branches has changed after rename operation")
	}

	stageBranch, err := ry.Branches.Get(Stage)
	if err != nil {
		t.Errorf("could not get renamed branch %q", Stage)
	}

	_, err = ry.Branches.Get(Prod)
	if err == nil {
		t.Errorf("branch %q can still be found after rename", Prod)
	}

	if prodBranch.RefValue != stageBranch.RefValue {
		t.Errorf("the renamed branch does have a different ref value than before the rename. old %s, new %s",
			prodBranch.RefValue, stageBranch.RefValue)
	}
}

func TestCopyBranch(t *testing.T) {
	// arrange
	ry := prepareEnvWithBranches(t)
	expectedBranches := len(ry.Branches.List()) + 1

	// act
	if err := ry.Branches.Copy(Prod, Stage); err != nil {
		t.Errorf("could not copy branch: %v", err)
	}

	// assert
	if expectedBranches != len(ry.Branches.List()) {
		t.Errorf("expected %d branches to exist, but %d branches exist", expectedBranches, len(ry.Branches.List()))
	}

	targetBranch, err := ry.Branches.Get(Stage)
	if err != nil {
		t.Errorf("could not find target branch %q after copy", Stage)
	}

	sourceBranch, err := ry.Branches.Get(Prod)
	if err != nil {
		t.Errorf("could not find source branch %q after copy", Prod)
	}

	if targetBranch.RefValue != sourceBranch.RefValue {
		t.Errorf("the ref value of the target branch %q does not match the ref value of the source branch %q",
			targetBranch.RefValue, sourceBranch.RefValue)
	}
}

func prepareEnvWithBranches(t *testing.T) *Repository {
	t.Helper()

	ry := prepareEnvWithCommits(t)
	branches := []struct {
		Name  string
		Value string
	}{
		{
			"master",
			"1234",
		},
		{
			"dev",
			"1234",
		},
		{
			"prod",
			"1234",
		},
	}

	for _, branch := range branches {
		if _, err := ry.Branches.Create(branch.Name, branch.Value); err != nil {
			t.Fatalf("failed to create branch %s", branch.Name)
		}
	}

	return ry
}
