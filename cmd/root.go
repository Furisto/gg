package cmd

import (
	"github.com/spf13/cobra"
)

func SetupRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "gog",
		Short: "The malevolent brother of git",
	}

	return rootCmd
}

type CommandOptions struct {
	Path string
}
