package repo

import (
	"testing"
)

func TestBareRepository(t *testing.T) {
	ry := setupInfoTestEnv(t, true)

	bare, err := ry.Info.IsBare()
	if err != nil {
		t.Fatalf("could not determine if repository is bare: %v", err)
	}

	if !bare {
		t.Errorf("repository is not considered bare")
	}

	detached, err := ry.Info.IsHeadDetached()
	if err != nil {
		t.Fatalf("could not determine if head is detached: %v", err)
	}

	if detached {
		t.Errorf("head is in detached state")
	}

	unborn, err := ry.Info.IsHeadUnborn()
	if err != nil {
		t.Fatalf("could not determine if head is unborn: %v", err)
	}

	if !unborn {
		t.Errorf("head is not considered unborn, even though no commits have been made")
	}
}

func TestEmptyRepository(t *testing.T) {
	ry := setupInfoTestEnv(t, false)

	bare, err := ry.Info.IsBare()
	if err != nil {
		t.Fatalf("could not determine if repository is bare: %v", err)
	}

	if bare {
		t.Errorf("repository is considered bare")
	}

	detached, err := ry.Info.IsHeadDetached()
	if err != nil {
		t.Fatalf("could not determine if head is detached: %v", err)
	}

	if detached {
		t.Errorf("head is in detached state")
	}

	unborn, err := ry.Info.IsHeadUnborn()
	if err != nil {
		t.Fatalf("could not determine if head is unborn: %v", err)
	}

	if !unborn {
		t.Errorf("head is not considered unborn, even though no commits have been made")
	}
}

func TestNonEmptyRepository(t *testing.T) {
	ry := prepareEnvWithCommits(t)

	bare, err := ry.Info.IsBare()
	if err != nil {
		t.Fatalf("could not determine if repository is bare: %v", err)
	}

	if bare {
		t.Errorf("repository is considered bare")
	}

	detached, err := ry.Info.IsHeadDetached()
	if err != nil {
		t.Fatalf("could not determine if head is detached: %v", err)
	}

	if detached {
		t.Errorf("head is in detached state")
	}

	unborn, err := ry.Info.IsHeadUnborn()
	if err != nil {
		t.Fatalf("could not determine if head is unborn: %v", err)
	}

	if unborn {
		t.Errorf("head is unborn, even though commits have been made")
	}
}

func TestDetachedRepository(t *testing.T) {
	ry := prepareEnvWithCommits(t)
	if err := ry.SetHead(ParentCommit); err != nil {
		t.Fatalf("could not set head: %v", err)
	}

	detached, err := ry.Info.IsHeadDetached()
	if err != nil {
		t.Fatalf("could not determine if head is detached: %v", err)
	}

	if !detached {
		t.Errorf("head should be in detached state")
	}

	unborn, err := ry.Info.IsHeadUnborn()
	if err != nil {
		t.Fatalf("could not determine if head is unborn: %v", err)
	}

	if unborn {
		t.Errorf("head is unborn, even though commits have been made")
	}
}

func setupInfoTestEnv(t *testing.T, bare bool) *Repository {
	t.Helper()

	dir := createTemporaryDir(t)
	ry, err := InitDefault(dir, bare)
	if err != nil {
		t.Fatalf("could not initialize repository for test")
	}

	return ry
}
