package log

import (
	"fmt"
	"github.com/furisto/gog/plumbing/objects"
	"io"
)

type logFormatter interface {
	Write(commit *objects.Commit) error
}

type defaultLogFormatter struct {
	writer io.Writer
}

func newDefaultLogFormatter(writer io.Writer) *defaultLogFormatter {
	return &defaultLogFormatter{
		writer: writer,
	}
}

func (dlf *defaultLogFormatter) Write(commit *objects.Commit) error {
	if _, err := fmt.Fprintf(dlf.writer, "commit  %v\n", commit.OID()); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(dlf.writer, "Author: %v %v\n", commit.Author.Name, commit.Author.Email); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(dlf.writer, "Date:   %v\n\n", commit.Author.TimeStamp.Format("Mon Jan 2 15:04:05 2006 -0700")); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(dlf.writer, "%v\n\n", commit.Message); err != nil {
		return err
	}

	return nil
}
