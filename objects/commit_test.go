package objects

import (
	hasher "github.com/furisto/gog/util"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestCommitBuilderWithExplicitValues(t *testing.T) {
	cb := NewCommitBuilder(hasher.Hash())
	commit, err := cb.WithAuthor("author", "author@test.com").
		WithCommitter("committer", "committer@test.com").
		WithMessage("Test message").
		Build()

	if err != nil {
		t.Errorf("encountered error while building commit: %v", err)
	}

	assert.Equal(t, commit.Author.Name, "author")
	assert.Equal(t, commit.Author.Email, "author@test.com")
	assert.Equal(t, commit.Commiter.Name, "committer")
	assert.Equal(t, commit.Commiter.Email, "committer@test.com")
	assert.Equal(t, commit.Message, "Test message")
}
