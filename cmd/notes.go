package cmd

import (
	"fmt"
	"github.com/furisto/gog/repo"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var FailedToResolveTemplate = "failed to resolve '%s' as a valid ref"

func SetupNotesCmd(context CommandContext) *cobra.Command {
	noteCmd := &cobra.Command{
		Use:   "notes",
		Short: "Add or inspect object notes",
	}
	handler := NewNotesCmd(context.Logger)

	addCmd := setupNoteAddCmd(handler)
	copyCmd := setupNoteCopyCmd(handler)
	appendCmd := setupNotesAppendCmd(handler)
	removeCmd := setupNotesRemoveCmd(handler)
	listCmd := setupNotesListCmd(handler)
	showCmd := setupNotesShowCmd(handler)

	noteCmd.AddCommand(addCmd, copyCmd, appendCmd, removeCmd, listCmd, showCmd)

	noteCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return handler.ExecuteList(NotesListCmdOptions{})
	}

	return noteCmd
}

func setupNoteAddCmd(handler NotesCommand) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add notes for a given object",
	}

	addCmd.Args = cobra.ExactArgs(1)
	addOptions := NotesAddCmdOptions{}
	addCmd.LocalFlags().BoolVarP(&addOptions.Force, "force", "f", false,
		"When adding notes to an object that already has notes, overwrite the existing notes")
	addCmd.LocalFlags().StringVarP(&addOptions.Message, "message", "m", "",
		"Use the given note message")

	addCmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		addOptions.Path, err = os.Getwd()
		if err != nil {
			return err
		}
		addOptions.TargetObject = args[0]
		return handler.ExecuteAdd(addOptions)
	}

	return addCmd
}

func setupNoteCopyCmd(handler NotesCommand) *cobra.Command {
	copyCmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy the notes for the first object onto the second object",
	}

	copyCmd.Args = cobra.ExactArgs(2)
	copyOptions := NotesCopyCmdOptions{}
	copyCmd.LocalFlags().BoolVarP(&copyOptions.Force, "force", "f", false,
		"When copying notes to an object that already has notes, overwrite the existing notes")

	copyCmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		copyOptions.Path, err = os.Getwd()
		if err != nil {
			return err
		}
		copyOptions.FromObject = args[0]
		copyOptions.ToObject = args[1]
		return handler.ExecuteCopy(copyOptions)
	}

	return copyCmd
}

func setupNotesAppendCmd(handler NotesCommand) *cobra.Command {
	appendCmd := &cobra.Command{
		Use:   "append",
		Short: "Append to the notes of an existing object",
	}

	appendCmd.Args = cobra.ExactArgs(1)
	appendOptions := NotesAppendCmdOptions{}
	appendCmd.LocalFlags().StringVarP(&appendOptions.Message, "message", "m", "",
		"Use the given note message")

	appendCmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		appendOptions.Path, err = os.Getwd()
		if err != nil {
			return err
		}
		return handler.ExecuteAppend(appendOptions)
	}

	return appendCmd
}

func setupNotesRemoveCmd(handler NotesCommand) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Removes the notes for a given object",
	}

	removeCmd.Args = cobra.ExactArgs(1)
	removeOptions := NotesRemoveCmdOptions{}
	removeCmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		removeOptions.ObjectRef = args[0]
		removeOptions.Path, err = os.Getwd()
		if err != nil {
			return err
		}
		return handler.ExecuteRemove(removeOptions)
	}

	return removeCmd
}

func setupNotesListCmd(handler NotesCommand) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the notes for a given object. If no object is given, show a list of all note objects",
	}

	listCmd.Args = cobra.MaximumNArgs(1)
	listOptions := NotesListCmdOptions{}
	listCmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		listOptions.Path, err = os.Getwd()
		if err != nil {
			return err
		}

		if len(args) > 0 {
			listOptions.ObjectRef = args[0]
		}

		return handler.ExecuteList(listOptions)
	}

	return listCmd
}

func setupNotesShowCmd(handler NotesCommand) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show the notes for a given object",
	}

	showCmd.Args = cobra.MaximumNArgs(1)
	showCmdOptions := NotesShowCmdOptions{}
	showCmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		showCmdOptions.Path, err = os.Getwd()
		if err != nil {
			return err
		}

		if len(args) > 0 {
			showCmdOptions.ObjectRef = args[0]
		}
		return handler.ExecuteShow(showCmdOptions)
	}

	return showCmd
}

type NotesAddCmdOptions struct {
	CommandOptions
	// Force overwrites an already existing note for the target commit
	Force bool
	// Message contains the value for the note
	Message string
	// TargetObject is the object id of the commit object to which the note will be added
	TargetObject string
}

type NotesCopyCmdOptions struct {
	CommandOptions
	// Force overwrites an already existing note for the target commit
	Force bool
	// FromObject is the object id of the commit object from which the note will be copied
	FromObject string
	// ToObject is the object id of the commit object to which the note will be copied
	ToObject string
}

type NotesAppendCmdOptions struct {
	CommandOptions
	// Message contains the value that will be appended to the note
	Message string
	// ObjectRef is the ref of the commit to which note the message will be appended
	ObjectRef string
}

type NotesRemoveCmdOptions struct {
	CommandOptions
	// ObjectRef is the ref of the note that will be deleted
	ObjectRef string
}

type NotesListCmdOptions struct {
	CommandOptions
	// ObjectRef is the ref of the commit for which the notes will be listed. If this is empty all notes
	// and commits will be listed
	ObjectRef string
}

type NotesShowCmdOptions struct {
	CommandOptions
	// ObjectRef is the ref of the commit object whose note will be shown
	ObjectRef string
}

type NotesCommand struct {
	writer io.Writer
}

func NewNotesCmd(writer io.Writer) NotesCommand {
	return NotesCommand{
		writer: writer,
	}
}

func (cmd *NotesCommand) ExecuteAdd(options NotesAddCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	_, err = ry.Notes.Create(options.Message, options.TargetObject, options.Force)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *NotesCommand) ExecuteCopy(options NotesCopyCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	if options.Force {
		fmt.Fprintf(cmd.writer, "Overwriting existing notes for object %s", options.ToObject)
	}

	if _, err := ry.Notes.Copy(options.FromObject, options.ToObject, options.Force); err != nil {
		fmt.Fprintf(cmd.writer, "Cannot copy notes. Found existing notes for object %s."+
			"Use '-f' to overwrite existing notes", options.ToObject)
		return err
	}

	return nil
}

func (cmd *NotesCommand) ExecuteAppend(options NotesAppendCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	if _, err := ry.Notes.Append(options.ObjectRef, options.Message); err != nil {
		return err
	}

	return nil
}

func (cmd *NotesCommand) ExecuteRemove(options NotesRemoveCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	ok, _ := ry.Storage.Stat(options.ObjectRef)
	if !ok {
		fmt.Fprintf(cmd.writer, FailedToResolveTemplate, options.ObjectRef)
		return nil
	}

	n, _ := ry.Notes.Find(options.ObjectRef)
	if n == nil {
		fmt.Fprintf(cmd.writer, "object %s has no note", options.ObjectRef)
		return nil
	}

	fmt.Fprintf(cmd.writer, "Removing note for object %s", options.ObjectRef)
	if err := ry.Notes.Remove(options.ObjectRef); err != nil {
		return err
	}

	return nil
}

func (cmd *NotesCommand) ExecuteList(options NotesListCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	notes, err := ry.Notes.List(options.ObjectRef)
	if err != nil {
		return err
	}

	for _, n := range notes {
		fmt.Fprintf(cmd.writer, "%s %s\n", n.MessageOID, n.CommitOID)
	}

	return nil
}

func (cmd *NotesCommand) ExecuteShow(options NotesShowCmdOptions) error {
	ry, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	note, _ := ry.Notes.Find(options.ObjectRef)
	if note == nil {
		return fmt.Errorf(FailedToResolveTemplate, options.ObjectRef)
	}

	message, err := note.Message()
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(cmd.writer, message)
	return err
}
