package cmd

import (
	"errors"
	"fmt"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func SetupCatFileCmd(context CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cat-file",
		Short: "Provide content or type and size information for repository objects",
	}

	cmd.Args = cobra.ExactArgs(1)

	options := CatFileOptions{}
	cmd.Flags().BoolVarP(&options.Pretty, "", "p", false, "pretty-print object's content")
	cmd.Flags().BoolVarP(&options.Size, "", "s", false, "")
	cmd.Flags().BoolVarP(&options.Type, "", "t", false, "")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		options.OID = args[0]
		if cwd, err := os.Getwd(); err == nil {
			options.Path = cwd
		} else {
			fmt.Fprintf(context.Logger, "")
			return nil
		}

		handler := NewCatFileCmd(context.Logger)
		return handler.Execute(options)
	}

	return cmd
}

type CatFileOptions struct {
	OID    string
	Path   string
	Type   bool
	Size   bool
	Pretty bool
}

type CatFileCmd struct {
	writer io.Writer
}

func NewCatFileCmd(writer io.Writer) CatFileCmd {
	return CatFileCmd{
		writer: writer,
	}
}

func (cmd *CatFileCmd) Execute(options CatFileOptions) error {
	repo, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	oids, err := repo.Storage.Find(options.OID)
	if err != nil {
		return err
	}

	if len(oids) > 1 {
		var compOids string
		for _, oid := range oids {
			compOids += oid + "\n"
		}
		return fmt.Errorf("cannot decide between: %v", compOids)
	}

	data, err := repo.Storage.Get(oids[0])
	if err != nil {
		return err
	}

	var o objects.Object
	// todo: implement other object types
	if objects.IsBlob(data) {
		o, err = objects.LoadBlob(data)
		if err != nil {
			return err
		}
	} else if objects.IsTree(data) {
		o, err = objects.LoadTree(data)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Invalid object type")
	}

	if options.Size {
		fmt.Fprintf(cmd.writer, "%v", o.Size())
	} else if options.Type {
		fmt.Fprintf(cmd.writer, "%v", o.Type())
	} else if options.Pretty {
		output, err := objects.FormatObject(o)
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.writer, output)
	}

	return nil
}
