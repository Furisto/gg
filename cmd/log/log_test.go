package log

import (
	"bytes"
	"fmt"
	"github.com/furisto/gog/cmd"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/repo"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestLogWithNoOptions(t *testing.T) {
	env := prepareEnvForLogTests(t)
	defer env.Cleanup()

	options := LogCmdOptions{
		Path: env.Repository.Location,
	}

	cmd := NewLogCmd(&env.Output, env.Formatter)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("log command did not execute successfully: %v", err)
	}

	compareLogOutput(t, env.Commits, env.Formatter.Commits)
}

func TestLogWithSkipOption(t *testing.T) {
	env := prepareEnvForLogTests(t)
	defer env.Cleanup()

	options := LogCmdOptions{
		Path:        env.Repository.Location,
		SkipCommits: 5,
	}

	cmd := NewLogCmd(&env.Output, env.Formatter)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("log command did not execute successfully: %v", err)
	}

	compareLogOutput(t, env.Commits[5:], env.Formatter.Commits)
}

func TestLogWithMaxOption(t *testing.T) {
	env := prepareEnvForLogTests(t)
	defer env.Cleanup()

	options := LogCmdOptions{
		Path:       env.Repository.Location,
		MaxCommits: 5,
	}

	cmd := NewLogCmd(&env.Output, env.Formatter)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("log command did not execute successfully: %v", err)
	}

	compareLogOutput(t, env.Commits[:5], env.Formatter.Commits)
}

func TestLogWithAuthorOption(t *testing.T) {
	env := prepareEnvForLogTests(t)
	defer env.Cleanup()

	re, err := regexp.Compile("fur[a-z]*")
	if err != nil {
		t.Fatalf("could not compile regex: %v", err)
	}

	options := LogCmdOptions{
		Path:   env.Repository.Location,
		Author: re,
	}

	cmd := NewLogCmd(&env.Output, env.Formatter)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("log command did not execute successfully: %v", err)
	}

	expected := []*objects.Commit{}
	for _, c := range env.Commits {
		if c.Author.Name == "furisto" {
			expected = append(expected, c)
		}
	}

	compareLogOutput(t, expected, env.Formatter.Commits)
}

func TestLogWithBeforeOption(t *testing.T) {
	env := prepareEnvForLogTests(t)
	defer env.Cleanup()

	limit := time.Date(2020, 11, 9, 2, 0, 0, 0, time.UTC)
	options := LogCmdOptions{
		Path:       env.Repository.Location,
		MaxCommits: 5,
		Before:     limit,
	}

	cmd := NewLogCmd(&env.Output, env.Formatter)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("log command did not execute successfully: %v", err)
	}

	var expected []*objects.Commit
	for _, c := range env.Commits {
		if c.Author.TimeStamp.Before(limit) {
			expected = append(expected, c)
		}
	}

	compareLogOutput(t, expected, env.Formatter.Commits)
}

func TestLogWithAfterOption(t *testing.T) {
	env := prepareEnvForLogTests(t)
	defer env.Cleanup()

	limit := time.Date(2020, 11, 9, 2, 0, 0, 0, time.UTC)
	options := LogCmdOptions{
		Path:  env.Repository.Location,
		After: limit,
	}

	cmd := NewLogCmd(&env.Output, env.Formatter)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("log command did not execute successfully: %v", err)
	}

	var expected []*objects.Commit
	for _, c := range env.Commits {
		if c.Author.TimeStamp.After(limit) {
			expected = append(expected, c)
		}
	}

	compareLogOutput(t, expected, env.Formatter.Commits)
}

func TestLogWithCombinedOptions(t *testing.T) {
	env := prepareEnvForLogTests(t)
	defer env.Cleanup()

	options := LogCmdOptions{
		Path:        env.Repository.Location,
		SkipCommits: 2,
		MaxCommits:  2,
	}

	cmd := NewLogCmd(&env.Output, env.Formatter)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("log command did not execute successfully: %v", err)
	}

	compareLogOutput(t, env.Commits[2:4], env.Formatter.Commits)
}

func prepareEnvForLogTests(t *testing.T) LogTestEnv {
	t.Helper()

	r, err := cmd.PrepareEnvWithNoCommmits()
	if err != nil {
		t.Fatalf("")
	}

	dates := []time.Time{
		time.Date(2020, 11, 8, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 8, 20, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 8, 22, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 8, 24, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 9, 2, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 9, 4, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 9, 6, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 9, 8, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 9, 10, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 9, 12, 0, 0, 0, time.UTC),
	}

	var commits []*objects.Commit
	for i := 0; i < 10; i++ {
		c, err := r.Commit(func(builder *objects.CommitBuilder) *objects.CommitBuilder {
			builder.WithMessage(fmt.Sprintf("test %v", i))
			builder.WithHook(func(commit *objects.Commit) {
				commit.Author.TimeStamp = dates[i]
				commit.Commiter.TimeStamp = dates[i]
			})
			if i%2 == 0 {
				builder.WithAuthor("log", "log@log.com")
			}
			return builder
		})

		if err != nil {
			t.Fatalf("could not create commit: %v", err)
		}
		commits = append(commits, c)
	}

	output := bytes.Buffer{}

	cleanup := func() {
		if err := os.RemoveAll(r.Location); err != nil {
			t.Fatalf("cleanup for test did not execute successfully: %v", err)
		}
	}

	return LogTestEnv{
		Repository: r,
		Commits:    reverse(commits),
		Dates:      dates,
		Output:     output,
		Formatter:  &testLogFormatter{},
		Cleanup:    cleanup,
	}
}

func compareLogOutput(t *testing.T, expected []*objects.Commit, actual []*objects.Commit) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("output of log was %v commits, but expected were %v commits", len(actual), len(expected))
		return
	}

	for i, e := range expected {
		if e.OID() != actual[i].OID() {
			t.Errorf("OIDs of commits did not match")
			formatter := newDefaultLogFormatter(os.Stdout)
			formatter.Write(e)
			formatter.Write(actual[i])
			break
		}
	}
}

type testLogFormatter struct {
	Commits []*objects.Commit
}

func (tlf *testLogFormatter) Write(commit *objects.Commit) error {
	tlf.Commits = append(tlf.Commits, commit)
	return nil
}

type LogTestEnv struct {
	Repository *repo.Repository
	Commits    []*objects.Commit
	Dates      []time.Time
	Output     bytes.Buffer
	Formatter  *testLogFormatter
	Cleanup    func()
}

func reverse(commits []*objects.Commit) []*objects.Commit {
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}
	return commits
}
