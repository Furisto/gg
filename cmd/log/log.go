package log

import (
	"github.com/furisto/gog/cmd"
	"github.com/furisto/gog/plumbing/objects"
	"github.com/furisto/gog/repo"
	"github.com/furisto/gog/storage"
	"github.com/spf13/cobra"
	"io"
	"regexp"
	"time"
)

func SetupLogCmd(context cmd.CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Show commit logs",
	}

	options := LogCmdOptions{}
	cmd.Flags().Uint64VarP(&options.MaxCommits, "max-count", "n", 10, "Limit the number of commits to output")
	cmd.Flags().Uint64Var(
		&options.SkipCommits, "skip", 0, "Skip number commits before starting to show the commit output")
	//cmd.Flags().StringVar(&options.Author, "author", "",
	//	"Limit the commits output to ones with author/committer header lines that match the specified pattern")
	cmd.Flags().String("author", "", "Limit the commits output to ones with author/committer header lines that match the specified pattern")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		handler := NewLogCmd(context.Logger, &defaultLogFormatter{})
		return handler.Execute(options)
	}

	return cmd
}

type LogCmdOptions struct {
	Path        string
	MaxCommits  uint64
	SkipCommits uint64
	Before      time.Time
	After       time.Time
	Author      *regexp.Regexp
}

type LogCommand struct {
	writer    io.Writer
	formatter logFormatter
}

func NewLogCmd(writer io.Writer, formatter logFormatter) LogCommand {
	return LogCommand{
		writer:    writer,
		formatter: formatter,
	}
}

func (cmd *LogCommand) Execute(options LogCmdOptions) error {
	r, err := repo.FromExisting(options.Path)
	if err != nil {
		return err
	}

	headRef, err := r.Head()
	if err != nil {
		return err
	}

	resolvedRef, err := r.Refs.Resolve(headRef)
	if err != nil {
		return err
	}

	data, err := r.Storage.Get(resolvedRef.RefValue)
	if err != nil {
		return err
	}

	lastCommit, err := objects.DecodeCommit(resolvedRef.RefValue, data)
	if err != nil {
		return err
	}

	iterator := cmd.createCommitIterator(lastCommit, r.Storage, options)
	for iterator.MoveNext() {
		commit := iterator.Current()
		cmd.formatter.Write(commit)
	}

	return nil
}

func (cmd *LogCommand) createCommitIterator(commit *objects.Commit, store storage.ObjectStore, options LogCmdOptions) objects.CommitIterator {
	var iter objects.CommitIterator = objects.NewCommitIterator(commit, store)
	if options.Author != nil {
		iter = objects.NewFilterCommitIterator(iter, func(commit *objects.Commit) bool {
			return options.Author.MatchString(commit.Author.Name)
		})
	}

	if !options.Before.IsZero() {
		iter = objects.NewFilterCommitIterator(iter, func(commit *objects.Commit) bool {
			return commit.Author.TimeStamp.Before(options.Before)
		})
	}

	if !options.After.IsZero() {
		iter = objects.NewFilterCommitIterator(iter, func(commit *objects.Commit) bool {
			return commit.Author.TimeStamp.After(options.After)
		})
	}

	if options.SkipCommits > 0 {
		iter = objects.NewSkipCommitIterator(iter, options.SkipCommits)
	}

	if options.MaxCommits > 0 {
		iter = objects.NewTakeCommitIterator(iter, options.MaxCommits)
	}

	return iter
}
