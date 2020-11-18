package cmd

import (
	"bytes"
	"testing"
)

func TestAddSingleFile(t *testing.T) {
	r := PrepareEnvWithNoCommmits(t)
	var output bytes.Buffer

	options := AddCmdOptions{
		Path:     r.Info.WorkingDirectory(),
		DryRun:   false,
		Patterns: []string{"**/*"},
	}

	cmd := NewAddCmd(&output)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("error encountered during command execution: %v", err)
	}
}
