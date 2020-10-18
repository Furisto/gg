package cmd

import (
	"fmt"
	"github.com/furisto/gog/repo"
	"github.com/furisto/gog/storage"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

func SetupHashObjectCmd(context CommandContext) *cobra.Command {
	hashObjectCmd := &cobra.Command{
		Use: "hash-object",
		Short: "Compute object ID and optionally creates a blob from a file",
	}

	hashObjectCmd.Args = cobra.ExactArgs(1)

	options := HashObjectOptions{}
	hashObjectCmd.Flags().BoolVar(&options.store, "w", false, "write the object into the object database")

	hashObjectCmd.RunE = func(cmd *cobra.Command, args []string) error {
		filePath, err := filepath.Abs(args[0])
		if err != nil {
			fmt.Fprintf(context.Logger, "%v is not a valid path", args[0])
		}
		options.file = filePath
		hashCmd := NewHashObjectCmd(context.Logger)
		return hashCmd.Execute(options)
	}

	return hashObjectCmd
}

type HashObjectOptions struct {
	file string
	store bool
}

type HashObjectCommand struct {
	writer io.Writer
}

func NewHashObjectCmd(writer io.Writer) HashObjectCommand {
	return HashObjectCommand{
		writer: writer,
	}
}

func (cmd *HashObjectCommand) Execute(options HashObjectOptions) error{
	stat, err := os.Stat(options.file)
	if err != nil {
		return fmt.Errorf("fatal: Cannot open '%v': No such file or directory", options.file)
	}

	if stat.IsDir() {
		return fmt.Errorf("cannot hash directory %v", options.file)
	}

	blob, err := storage.NewBlobFromFile(options.file)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.writer, "%v", blob.OID())

	if options.store {
		gitRepo, err := repo.FromExisting(options.file)
		if err != nil {
			return err
		}

		gitRepo.Storage.Put(blob.OID(), blob.Bytes())
	}

	return nil
}