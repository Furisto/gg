package cmd

import (
	"fmt"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

func SetupInitCmd(context CommandContext) *cobra.Command {
	var initCmd = &cobra.Command{
		Use: "init",
		Short: "Initializes a git repository",
	}

	initCmd.Args = cobra.NoArgs
	initCmd.SuggestFor = []string {"new", "create"}

	options := InitCmdOptions{ }
	initCmd.Flags().BoolVar(&options.bare, "bare", false, "create a bare repository")
	initCmd.Flags().BoolVarP(&options.quiet, "quiet", "q", false, "be quiet")
	
	initCmd.Run = func(cmd *cobra.Command, args []string) {
		newInitCmd := NewInitCmd(context.Logger)
		if cwd, err:= os.Getwd(); err == nil {
			options.path = cwd
		} else {
			fmt.Fprint(context.Logger, "Could not determine working directory")
			return
		}
		newInitCmd.Execute(options)
	}

	return initCmd
}

type InitCommand struct {
	writer io.Writer
}

type InitCmdOptions struct {
	bare bool
	quiet bool
	path string
}

func NewInitCmd(logger io.Writer) InitCommand {
	return InitCommand{
		writer: logger,
	}
}

func (cmd *InitCommand) Execute(opt InitCmdOptions) {
	gitFolder := filepath.Join(opt.path, ".git")
	var message string

	if _, err := os.Stat(gitFolder); os.IsNotExist(err) {
		message = fmt.Sprintln("Initialized empty Git repository in %v", opt.path)
	} else {
		message = fmt.Sprintln("Reinitialized existing Git repository in %v", opt.path)
	}

	repo.InitDefault(opt.path, opt.bare)
	fmt.Fprintf(cmd.writer, message)
}
