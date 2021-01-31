package cmd

import (
	"fmt"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func SetupUnpackFileCmd(context CommandContext) *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Args = cobra.ExactArgs(1)

	options := UnpackFileCmdOptions{}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		options.OID = args[0]
		path, err := os.Getwd()
		if err != nil {
			return err
		}
		options.Path = path

		handler := NewUnpackFileCommand(context.Logger)
		return handler.Execute(options)
	}

	return cmd
}

type UnpackFileCmdOptions struct {
	Path string
	OID  string
}

type UnpackFileCommand struct {
	writer io.Writer
}

func NewUnpackFileCommand(writer io.Writer) UnpackFileCommand {
	return UnpackFileCommand{
		writer: writer,
	}
}

func (cmd *UnpackFileCommand) Execute(options UnpackFileCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	data, err := ry.Storage.Get(options.OID)
	if err != nil {
		return err
	}

	if !objects.IsBlob(data) {
		return fmt.Errorf("%s is not a blob object", options.OID)
	}

	blob, err := objects.LoadBlob(data)
	if err != nil {
		return err
	}

	blobFile, err := ioutil.TempFile(ry.Info.WorkingDirectory(), ".merge_file_")
	if err != nil {
		return err
	}

	_, err = blobFile.Write(blob.Content)
	if err != nil {
		return err
	}

	fmt.Fprint(cmd.writer, filepath.Base(blobFile.Name()))
	return nil
}
