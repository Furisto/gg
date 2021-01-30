package cmd

import (
	"bytes"
	"fmt"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func SetupCheckoutCmd(context CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout",
		Short: "Switch branches or restore working tree files",
	}

	options := CheckoutCmdOptions{}
	// cmd.Flags().Bool("b", false, "create and checkout a new branch")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if cwd, err := os.Getwd(); err == nil {
			options.Path = cwd
		}
		handler := NewCheckoutCmd(context.Logger)
		return handler.Execute(options)
	}

	return cmd
}

type CheckoutCmdOptions struct {
	Path         string
	CreateBranch bool
	Ref          string
}

type CheckoutCommand struct {
	writer io.Writer
}

func NewCheckoutCmd(writer io.Writer) CheckoutCommand {
	return CheckoutCommand{
		writer: writer,
	}
}

func (cmd *CheckoutCommand) Execute(options CheckoutCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	ref, err := ry.Refs.Find(options.Ref)
	if err != nil {
		return err
	}

	checkoutRef, err := ry.Refs.Resolve(ref)
	if err != nil {
		return err
	}

	refObjectData, err := ry.Storage.Get(checkoutRef.RefValue)
	if err != nil {
		return err
	}

	if objects.IsBlob(refObjectData) {
		return fmt.Errorf("cannot checkout blob")
	}

	var tree *objects.Tree
	if objects.IsCommit(refObjectData) {
		commit, err := objects.DecodeCommit(checkoutRef.RefValue, refObjectData)
		if err != nil {
			return err
		}

		tree, err = cmd.commitToTree(commit, ry)
		if err != nil {
			return err
		}
	} else if objects.IsTree(refObjectData) {
		tree, err = objects.LoadTree(refObjectData)
		if err != nil {
			return err
		}
	} else if objects.IsTag(refObjectData) {
		tag, err := objects.DecodeTag(checkoutRef.RefValue, bytes.NewReader(refObjectData))
		if err != nil {
			return err
		}

		if tag.TargetType() == "Blob" {
			return fmt.Errorf("cannot checkout blob")
		} else if tag.TargetType() == "Commit" {
			commit, err := objects.DecodeCommit(checkoutRef.RefValue, refObjectData)
			if err != nil {
				return err
			}

			tree, err = cmd.commitToTree(commit, ry)
			if err != nil {
				return err
			}
		} else if tag.TargetType() == "Tree" {
			tree, err = objects.LoadTree(refObjectData)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("unknown object type")
	}

	err = filepath.Walk(ry.Info.WorkingDirectory(), func(path string, info os.FileInfo, err error) error {
		if path == ry.Info.WorkingDirectory() || strings.HasPrefix(path, ry.Info.GitDirectory()) {
			return nil
		}

		return os.RemoveAll(path)
	})

	if err != nil {
		return err
	}

	return cmd.checkoutTree(ry.Info.WorkingDirectory(), tree, ry)
}

func (cmd *CheckoutCommand) commitToTree(commit *objects.Commit, ry *repo.Repository) (*objects.Tree, error) {
	treeData, err := ry.Storage.Get(commit.Tree)
	if err != nil {
		return nil, err
	}

	return objects.LoadTree(treeData)
}

func (cmd *CheckoutCommand) checkoutTree(path string, tree *objects.Tree, ry *repo.Repository) error {
	for _, entry := range tree.Entries() {
		if entry.Mode != 0o40000 {
			data, err := ry.Storage.Get(entry.OID)
			if err != nil {
				return err
			}

			blob, err := objects.LoadBlob(data)
			if err != nil {
				return err
			}

			if err = ioutil.WriteFile(filepath.Join(path, entry.Name), blob.Content, entry.Mode); err != nil {
				return err
			}
		} else {
			dirPath := filepath.Join(path, entry.Name)
			if err := os.Mkdir(dirPath, os.ModeDir); err != nil {
				return err
			}

			treeData, err := ry.Storage.Get(entry.OID)
			if err != nil {
				return err
			}

			subTree, err := objects.LoadTree(treeData)
			if err != nil {
				return err
			}

			if err := cmd.checkoutTree(dirPath, subTree, ry); err != nil {
				return err
			}
		}
	}

	return nil
}
