package cmd

import (
	"bytes"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

const merge_file_pattern = ".merge_file_*"

func TestUnpackBlobFile(t *testing.T) {
	ry := createTestRepository(t)

	blob := objects.NewBlob([]byte("test unpack-file"))
	if err := ry.Storage.Put(blob.OID(), blob.Bytes()); err != nil {
		t.Fatalf("could not store blob in object database")
	}

	options := UnpackFileCmdOptions{
		Path: ry.Info.WorkingDirectory(),
		OID:  blob.OID(),
	}

	output := new(bytes.Buffer)
	cmd := NewUnpackFileCommand(output)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("error occured while executing unpack-file command: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(ry.Info.WorkingDirectory(), merge_file_pattern))
	if err != nil {
		t.Fatalf("error occured while trying to match %q", merge_file_pattern)
	}

	if len(matches) != 1 {
		t.Errorf("expected one merge_file_ but was %d", len(matches))
	}

	assert.Equal(t, filepath.Base(matches[0]), output.String())
}
