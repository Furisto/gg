package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/furisto/gog/config"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/plumbing/refs"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrToManyParams     = errors.New("too many params")
	ErrTagAlreadyExists = errors.New("tag already exists")
)

func SetupTagCmd(context CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Create, list, delete or verify a tag object ",
	}

	cmd.Args = validateTagArgs

	createOptions := TagCmdCreateOptions{}
	cmd.Flags().BoolVarP(&createOptions.IsAnnotated, "annotate", "a", false, "annotated tag, needs a message")
	cmd.Flags().StringVarP(&createOptions.Message, "message", "m", "", "tag message")
	cmd.Flags().StringP("file", "F", "", "read message from a file")
	cmd.Flags().BoolVarP(&createOptions.Force, "force", "f", false, "replace the tag if exists")

	listOptions := TagCmdListOptions{}
	cmd.Flags().StringVar(&listOptions.PointsAt, "points-at", "", "print only tags of the object")

	deleteOptions := TagCmdDeleteOptions{}
	cmd.Flags().StringVarP(&deleteOptions.TagName, "delete", "d", "", "delete tags")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		handler := NewTagCmd(context.Logger)
		if len(args) == 0 {
			return handler.ExecuteList(listOptions)
		}

		_, err := handler.ExecuteCreate(createOptions)
		return err
	}

	return cmd
}

// same as cobra.MaximumArgs but returns the same error as git
func validateTagArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 2 {
		return ErrToManyParams
	}

	return nil
}

type TagCmdOptions struct {
	Path string
}

type TagCmdCreateOptions struct {
	Path        string
	TagName     string
	Target      string
	Message     string
	IsAnnotated bool
	Force       bool
}

type TagCmdListOptions struct {
	TagCmdOptions
	PointsAt string
}

type TagCmdDeleteOptions struct {
	TagCmdOptions
	TagName string
}

type TagCommand struct {
	writer io.Writer
}

func NewTagCmd(writer io.Writer) TagCommand {
	return TagCommand{
		writer: writer,
	}
}

func (cmd *TagCommand) ExecuteList(options TagCmdListOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	if len(options.PointsAt) != 0 {
		exists, err := ry.Storage.Stat(options.PointsAt)
		if !exists {
			return fmt.Errorf("malformed object name %s", options.PointsAt)
		}
		if err != nil {
			return err
		}
	}

	tags := ry.Tags.List()
	for _, tag := range tags {
		// todo: handle symbolic refs and incomplete hash refs
		if len(options.PointsAt) != 0 && tag.RefValue != options.PointsAt {
			continue
		}

		fmt.Fprintln(cmd.writer, refs.ShortTagname(tag.Name))
	}

	return nil
}

func (cmd *TagCommand) ExecuteDelete(options TagCmdDeleteOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	err = ry.Tags.Delete(options.TagName)
	if err != nil {
		if err == refs.ErrRefNotExist {
			return fmt.Errorf("tag '%s' not found", options.TagName)
		} else {
			return err
		}
	}
	return nil
}

func (cmd *TagCommand) ExecuteCreate(options TagCmdCreateOptions) (*objects.Tag, error) {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return nil, err
	}

	//if err = refs.ValidateRef(options.TagName); err != nil {
	//	return err
	//}

	if options.IsAnnotated {
		if len(options.Message) == 0 {
			options.Message, err = cmd.requestMessage(ry)
			if err != nil {
				return nil, err
			}
		}

		tagger, err := cmd.retrieveTagger(ry.Config)
		if err != nil {
			return nil, err
		}

		annotated, err := ry.Tags.CreateAnnotated(options.TagName, options.Target, tagger, options.Message, options.Force)
		return annotated, err
	}

	if _, err := ry.Tags.CreateLightweight(options.TagName, options.Target, options.Force); err != nil {
		return nil, err
	}

	return nil, nil
}

func (cmd *TagCommand) requestMessage(ry *repo.Repository) (string, error) {
	editorCmd, err := ry.Config.Get("core", "editor")
	if err != nil {
		return "", fmt.Errorf("editor is not defined")
	}

	// create empty file for tag message
	tagMsgFilePath := filepath.Join(ry.Info.GitDirectory(), "TAG_EDITMSG")
	tagMsgFile, err := os.Create(tagMsgFilePath)
	if err != nil {
		return "", err
	}
	tagMsgFile.Close()

	// open tag message file with configured editor
	editor := exec.Command(editorCmd + " " + tagMsgFilePath)
	if err = editor.Run(); err != nil {
		return "", err
	}

	// retrieve user specified message
	tagMsgFile, err = os.Open(tagMsgFilePath)
	if err != nil {
		return "", err
	}
	defer tagMsgFile.Close()

	builder := strings.Builder{}
	scanner := bufio.NewScanner(tagMsgFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		builder.WriteString(line)
	}

	return builder.String(), nil
}

func (cmd *TagCommand) retrieveTagger(cfg config.Config) (*objects.Signature, error) {
	name, err := cfg.Get("user", "name")
	if err != nil {
		return nil, err
	}

	email, err := cfg.Get("user", "email")
	if err != nil {
		return nil, err
	}

	return &objects.Signature{
		Name:      name,
		Email:     email,
		TimeStamp: time.Now(),
	}, nil
}
