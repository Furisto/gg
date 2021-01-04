package cmd

import (
	"errors"
	"fmt"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
)

var (
	ErrBranchRequired     = errors.New("branch name required")
	ErrBranchNotFound     = errors.New("branch not found")
	ErrRepositoryRequired = errors.New("")
)

func SetupBranchCmd(context CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch",
		Short: "List, create, or delete branches",
	}

	options := BranchCmdOptions{}
	//cmd.Flags().StringVarP(options.)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		handler := NewBranchCommand(context.Logger)
		return handler.Execute(options)
	}

	return cmd
}

type BranchCmdOptions struct {
	Path          string
	BranchName    string
	Delete        bool
	Rename        bool
	NewBranchName string
}

type BranchCommand struct {
	writer io.Writer
}

func NewBranchCommand(writer io.Writer) BranchCommand {
	return BranchCommand{
		writer: writer,
	}
}

func (cmd *BranchCommand) Execute(options BranchCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return ErrRepositoryRequired
	}

	if options.Delete && options.BranchName == "" {
		return ErrBranchRequired
	}

	if options.Delete {
		if _, err := ry.Branches.Get(options.BranchName); err != nil {
			return ErrBranchNotFound
		}

		if err := ry.Branches.Delete(options.BranchName); err != nil {
			return err
		}

		return nil
	}

	if options.BranchName == "" {
		branches := ry.Branches.List()
		for _, branch := range branches {
			fmt.Fprintln(cmd.writer, branch.Name)
		}
		return nil
	}

	headRef, err := ry.Head(true)
	if err != nil {
		return err
	}

	_, err = ry.Branches.Create(options.BranchName, headRef.RefValue)
	return err
}
