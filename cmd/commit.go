package cmd

import (
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/plumbing/refs"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func SetupCommitCmd(context CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "",
	}

	options := CommitOptions{}
	cmd.Flags().StringVarP(&options.Message, "message", "m", "", "commit message")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if cwd, err := os.Getwd(); err == nil {
			options.Path = cwd
		} else {
			return err
		}

		handler := NewCommitCmd(context.Logger)
		_, err := handler.Execute(options)
		return err
	}

	return cmd
}

type CommitOptions struct {
	Path    string
	Message string
}

type CommitCommand struct {
	writer io.Writer
}

func NewCommitCmd(writer io.Writer) CommitCommand {
	return CommitCommand{
		writer: writer,
	}
}

func (cmd *CommitCommand) Execute(options CommitOptions) (*objects.Commit, error) {
	r, err := repo.FromExisting(options.Path)
	if err != nil {
		return nil, err
	}

	tree, err := objects.NewTreeFromDirectory(r.Info.WorkingDirectory(), "")
	if err != nil {
		return nil, err
	}
	if err := tree.Save(r.Storage); err != nil {
		return nil, err
	}

	headRef, err := r.Head(false)
	if err != nil {
		return nil, err
	}
	parentRef, err := r.Refs.Resolve(headRef)
	if err != nil {
		if err != refs.ErrRefNotExist {
			return nil, err
		}

		_, err := r.Branches.Get("master")
		if err != refs.ErrRefNotExist {
			return nil, err
		}

		master, err := r.Branches.Create("master", "")
		if err != nil {
			return nil, err
		}

		parentRef = master
	}

	commit, err := objects.NewCommitBuilder(tree.OID()).
		WithConfig(r.Config).
		WithParent(parentRef.RefValue).
		WithMessage(options.Message).
		Build()

	if err != nil {
		return nil, err
	}

	err = commit.Save(r.Storage)
	if err != nil {
		return nil, err
	}

	if _, err := r.Refs.Set(parentRef.Name, commit.OID()); err != nil {
		return nil, err
	}

	return commit, err
}
