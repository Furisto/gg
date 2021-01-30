package main

import (
	"fmt"
	"github.com/furisto/gog/cmd"
	"github.com/furisto/gog/cmd/log"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	rootCmd := setupCommands()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("fatal: %v", err)
		os.Exit(1)
	}
}

func setupCommands() *cobra.Command {
	cmdContext := cmd.CommandContext{
		Logger: os.Stdout,
	}
	rootCmd := cmd.SetupRootCmd()

	initCmd := cmd.SetupInitCmd(cmdContext)
	rootCmd.AddCommand(initCmd)

	hashObject := cmd.SetupHashObjectCmd(cmdContext)
	rootCmd.AddCommand(hashObject)

	catFile := cmd.SetupCatFileCmd(cmdContext)
	rootCmd.AddCommand(catFile)

	writeTree := cmd.SetupWriteTreeCmd(cmdContext)
	rootCmd.AddCommand(writeTree)

	commit := cmd.SetupCommitCmd(cmdContext)
	rootCmd.AddCommand(commit)

	log := log.SetupLogCmd(cmdContext)
	rootCmd.AddCommand(log)

	add := cmd.SetupAddCmd(cmdContext)
	rootCmd.AddCommand(add)

	branch := cmd.SetupBranchCmd(cmdContext)
	rootCmd.AddCommand(branch)

	tag := cmd.SetupTagCmd(cmdContext)
	rootCmd.AddCommand(tag)

	checkout := cmd.SetupCheckoutCmd(cmdContext)
	rootCmd.AddCommand(checkout)

	return rootCmd
}
