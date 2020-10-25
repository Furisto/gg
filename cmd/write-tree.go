package cmd

import (
	"errors"
	"github.com/furisto/gog/objects"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func SetupWriteTreeCmd(context CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write-tree",
		Short: "Create a tree object from the index or the working directory",
	}

	options := WriteTreeOptions{}
	cmd.Flags().BoolVarP(&options.UseWorkingDirectory, "working-dir", "w", false, "Use working directory")
	cmd.Flags().StringVar(&options.Prefix, "prefix", "", "write tree object for a subdirectory <prefix>")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		handler := NewWriteTreeCmd(context.Logger)
		return handler.Execute(options)
	}

	return cmd
}

type WriteTreeOptions struct {
	Path                string
	UseWorkingDirectory bool
	Prefix              string
}

type WriteTreeCommand struct {
	writer io.Writer
}

func NewWriteTreeCmd(writer io.Writer) WriteTreeCommand {
	return WriteTreeCommand{
		writer: writer,
	}
}

func (cmd *WriteTreeCommand) Execute(options WriteTreeOptions) error {
	if _, err := os.Stat(options.Path); os.IsNotExist(err) {
		return err
	}

	if options.UseWorkingDirectory {
		r, err := repo.FromExisting(options.Path)
		if err != nil {
			return err
		}

		tree, err := objects.NewTreeFromDirectory(options.Path, options.Prefix)
		if err != nil {
			return err
		}

		if err := tree.Save(r.Storage); err != nil {
			return err
		}
	} else {
		return errors.New("Not supported")
	}

	return nil
}
