package cmd

import (
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func SetupAddCmd(context CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "",
	}

	options := AddCmdOptions{}
	cmd.Flags().BoolVarP(&options.DryRun, "dry-run", "n", false, "")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		handler := NewAddCmd(context.Logger)
		return handler.Execute(options)
	}

	return cmd
}

type AddCmdOptions struct {
	Path     string
	DryRun   bool
	Patterns []string
}

type AddCommand struct {
	writer io.Writer
}

func NewAddCmd(writer io.Writer) AddCommand {
	return AddCommand{
		writer: writer,
	}
}

func (cmd *AddCommand) Execute(options AddCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	for _, pattern := range options.Patterns {
		matches, err := filepath.Glob(filepath.Join(ry.Info.WorkingDirectory(), pattern))
		if err != nil {
			return err
		}

		for _, match := range matches {
			if strings.HasPrefix(match, ry.Info.GitDirectory()) {
				continue
			}

			if stat, err := os.Stat(match); err != nil || stat.IsDir() {
				continue
			}

			entry := ry.Index.Find(match)
			if entry != nil {
				stat, err := os.Stat(match)
				if err != nil {
					return err
				}

				if entry.Match(stat) {
					continue
				}
			}

			blob, err := objects.NewBlobFromFile(match)
			if err != nil {
				return err
			}

			if err := ry.Index.Set(blob.OID(), match); err != nil {
				return err
			}
		}
	}

	if err := ry.Index.Flush(); err != nil {
		return err
	}

	return nil
}
