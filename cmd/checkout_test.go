package cmd

import (
	"bytes"
	"testing"
)

func TestCheckout(t *testing.T) {
	ry, _ := prepareEnvWithCommitObjects(t)

	options := CheckoutCmdOptions{
		Path:         ry.Info.WorkingDirectory(),
		CreateBranch: false,
		Ref:          "refs/heads/master",
	}

	output := new(bytes.Buffer)
	cmd := NewCheckoutCmd(output)
	if err := cmd.Execute(options); err != nil {
		t.Fatalf("error occured during execution of checkout command: %v", err)
	}
}
