package cmd

import "github.com/spf13/cobra"

func SetupHashObjectCmd() *cobra.Command {
	hashObjectCmd := &cobra.Command{

	}

	return hashObjectCmd
}


type HashObjectCommand struct {

}

func NewHashObject() *HashObjectCommand {
	return nil
}