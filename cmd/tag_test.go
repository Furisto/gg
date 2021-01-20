package cmd

import (
	"bytes"
	"github.com/furisto/gog/repo"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	lightweightName = "lightweight"
	annotatedName   = "annotated"
	tagMessage      = "tag test"
)

func TestCreateLightweightTag(t *testing.T) {
	ry := createTestRepository(t)

	options := NewTagCmdOptionsBuilder(ry.Info.WorkingDirectory()).
		WithTagName(lightweightName).
		WithAnnotated(false).
		Build()

	var buffer bytes.Buffer
	cmd := NewTagCmd(&buffer)
	if _, err := cmd.ExecuteCreate(options); err != nil {
		t.Errorf("could not execute tagging operation: %v", err)
	}

	tagRef, err := ry.Tags.Get(lightweightName)
	if err != nil {
		t.Errorf("could not retrieve tag %s", lightweightName)
		return
	}

	if tagRef.Name != "refs/tags/lightweight" {
		t.Errorf("expected tag name of %s, but was %s", lightweightName, tagRef.Name)
	}

	if tagRef.RefValue != "" {
		t.Errorf("expected tag target value of %s, but was %s", "", tagRef.RefValue)
	}
}

func TestCreateAnnotatedTag(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetCommit := commits[0]

	options := NewTagCmdOptionsBuilder(ry.Info.WorkingDirectory()).
		WithTarget(targetCommit.OID()).
		Build()

	var buffer bytes.Buffer
	cmd := NewTagCmd(&buffer)
	tag, err := cmd.ExecuteCreate(options)
	if err != nil {
		t.Errorf("could not execute tagging operation: %v", err)
		return
	}

	assert.Equal(t, "Commit", tag.TargetType(), "tag target type")
	assert.Equal(t, targetCommit.OID(), tag.TargetOID(), "target object id")
	assert.Equal(t, "furisto", tag.Tagger().Name, "tagger name")
	assert.Equal(t, "furisto@test.com", tag.Tagger().Email, "tagger email")
	assert.Equal(t, tagMessage, tag.Message(), "tag message")
}

func TestCreateAnnotatedTagToTree(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetObject := commits[0].Tree

	options := NewTagCmdOptionsBuilder(ry.Info.WorkingDirectory()).
		WithTarget(targetObject).
		Build()

	var buffer bytes.Buffer
	cmd := NewTagCmd(&buffer)
	tag, err := cmd.ExecuteCreate(options)
	if err != nil {
		t.Errorf("could not execute tagging operation: %v", err)
		return
	}

	assert.Equal(t, "Tree", tag.TargetType(), "tag target type")
	assert.Equal(t, targetObject, tag.TargetOID(), "target object id")
	assert.Equal(t, "furisto", tag.Tagger().Name, "tagger name")
	assert.Equal(t, "furisto@test.com", tag.Tagger().Email, "tagger email")
	assert.Equal(t, tagMessage, tag.Message(), "tag message")
}

// todo: test
func TestCreateAnnotatedTagToBlob(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetObject := commits[0].Tree

	options := NewTagCmdOptionsBuilder(ry.Info.WorkingDirectory()).
		WithAnnotated(true).
		WithTarget(targetObject).
		Build()

	var buffer bytes.Buffer
	cmd := NewTagCmd(&buffer)
	tag, err := cmd.ExecuteCreate(options)
	if err != nil {
		t.Errorf("could not execute tagging operation: %v", err)
		return
	}

	assert.Equal(t, "Commit", tag.TargetType(), "tag target type")
	assert.Equal(t, targetObject, tag.TargetOID(), "target object id")
	assert.Equal(t, "furisto", tag.Tagger().Name, "tagger name")
	assert.Equal(t, "furisto@test.com", tag.Tagger().Email, "tagger email")
	assert.Equal(t, tagMessage, tag.Message(), "tag message")
}

func TestHandleInvalidTagTarget(t *testing.T) {
	ry, _ := prepareEnvWithCommitObjects(t)
	targetObject := "1234"

	options := NewTagCmdOptionsBuilder(ry.Info.WorkingDirectory()).
		WithTarget(targetObject).
		Build()

	var buffer bytes.Buffer
	cmd := NewTagCmd(&buffer)
	_, err := cmd.ExecuteCreate(options)

	assert.Error(t, repo.ErrInvalidTagTarget, err)
}

func TestHandleTagAlreadyExists(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetObject := commits[0].OID()
	if _, err := ry.Tags.CreateLightweight(lightweightName, targetObject, false); err != nil {
		t.Fatalf("could not create tag %s", lightweightName)
	}

	options := NewTagCmdOptionsBuilder(ry.Info.WorkingDirectory()).
		WithTagName(lightweightName).
		WithAnnotated(false).
		WithTarget(targetObject).
		Build()

	var buffer bytes.Buffer
	cmd := NewTagCmd(&buffer)
	_, err := cmd.ExecuteCreate(options)
	assert.EqualError(t, err, repo.ErrTagAlreadyExists.Error())
}

func TestHandleOverwriteTag(t *testing.T) {
	ry, commits := prepareEnvWithCommitObjects(t)
	targetObject := commits[0].OID()
	if _, err := ry.Tags.CreateLightweight(lightweightName, targetObject, false); err != nil {
		t.Fatalf("could not create tag %s", lightweightName)
	}

	options := NewTagCmdOptionsBuilder(ry.Info.WorkingDirectory()).
		WithTagName(lightweightName).
		WithAnnotated(false).
		WithOverwrite(true).
		WithTarget(targetObject).
		Build()

	var buffer bytes.Buffer
	cmd := NewTagCmd(&buffer)
	_, err := cmd.ExecuteCreate(options)
	if err != nil {
		t.Errorf("error creating annotated tag %q with overwrite", annotatedName)
	}
}

func TestListTagsUnfiltered(t *testing.T) {
	ry, _, _ := prepareEnvWithTags(t)

	options := TagCmdListOptions{
		PointsAt: "",
	}
	options.Path = ry.Info.WorkingDirectory()

	var actual bytes.Buffer
	cmd := NewTagCmd(&actual)
	if err := cmd.ExecuteList(options); err != nil {
		t.Errorf("encountered error while executing tag list operation: %v", err)
		return
	}

	expected := []byte(
		"lightweight0\n" +
			"lightweight1\n" +
			"lightweight2\n" +
			"lightweight3\n")

	assert.Equal(t, expected, actual.Bytes())
}

func TestListTagsFilterdByPointedAt(t *testing.T) {
	ry, commits, _ := prepareEnvWithTags(t)
	targetObject := commits[0].OID()

	options := TagCmdListOptions{
		PointsAt: targetObject,
	}
	options.Path = ry.Info.WorkingDirectory()

	var actual bytes.Buffer
	cmd := NewTagCmd(&actual)
	if err := cmd.ExecuteList(options); err != nil {
		t.Errorf("encountered error while executing tag list operation: %v", err)
		return
	}

	expected := []byte(
		"lightweight0\n" +
			"lightweight2\n")

	assert.Equal(t, expected, actual.Bytes())
}

func TestDeleteTag(t *testing.T) {
	ry, _, originalTags := prepareEnvWithTags(t)

	options := TagCmdDeleteOptions{
		TagName: "lightweight0",
	}
	options.Path = ry.Info.WorkingDirectory()

	var actual bytes.Buffer
	cmd := NewTagCmd(&actual)
	if err := cmd.ExecuteDelete(options); err != nil {
		t.Errorf("error occured during tag delete operation: %v", err)
		return
	}

	tags := ry.Tags.List()
	expected := originalTags[1:]
	assert.Equal(t, expected, tags)
}

func TestDeleteUnknownTag(t *testing.T) {
	ry, _, _ := prepareEnvWithTags(t)

	options := TagCmdDeleteOptions{
		TagName: "doesnotexist",
	}
	options.Path = ry.Info.WorkingDirectory()

	var actual bytes.Buffer
	cmd := NewTagCmd(&actual)
	err := cmd.ExecuteDelete(options)
	if err == nil {
		t.Errorf("did not report an error while trying to delete a tag that does not exist")
		return
	}

	assert.EqualError(t, err, "tag 'doesnotexist' not found")
}

type TagCmdOptionsBuilder struct {
	options TagCmdCreateOptions
}

func NewTagCmdOptionsBuilder(path string) *TagCmdOptionsBuilder {
	options := TagCmdCreateOptions{
		Path:        path,
		TagName:     annotatedName,
		Target:      "",
		Message:     tagMessage,
		IsAnnotated: true,
		Force:       false,
	}

	return &TagCmdOptionsBuilder{
		options: options,
	}
}

func (b *TagCmdOptionsBuilder) WithTagName(tagName string) *TagCmdOptionsBuilder {
	b.options.TagName = tagName
	return b
}

func (b *TagCmdOptionsBuilder) WithTarget(target string) *TagCmdOptionsBuilder {
	b.options.Target = target
	return b
}

func (b *TagCmdOptionsBuilder) WithAnnotated(flag bool) *TagCmdOptionsBuilder {
	b.options.IsAnnotated = flag
	return b
}

func (b *TagCmdOptionsBuilder) WithOverwrite(flag bool) *TagCmdOptionsBuilder {
	b.options.Force = flag
	return b
}

func (b *TagCmdOptionsBuilder) Build() TagCmdCreateOptions {
	return b.options
}
