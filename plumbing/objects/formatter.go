package objects

import (
	"fmt"
	"os"
	"strings"
)

func FormatObject(o Object) (string, error) {
	switch o.Type() {
	case "Blob":
		blob, ok := o.(*Blob)
		if !ok {
			panic("object is of type blob, but cannot be cast to blob")
		}

		return string(blob.Content), nil
	case "Tree":
		tree, ok := o.(*Tree)
		if !ok {
			panic("object is of type tree, but cannot be cast to tree")
		}

		builder := strings.Builder{}
		for _, entry := range tree.Entries() {
			_, err := builder.Write(
				[]byte(fmt.Sprintf("%o %v %v %5v\n", entry.Mode, mapModeToType(entry.Mode), entry.OID, entry.Name)))
			if err != nil {
				return "", err
			}
		}
		return builder.String(), nil
	case "Commit":
		commit, ok := o.(*Commit)
		if !ok {
			panic("object is of type commit, but cannot be cast to commit")
		}

		return fmt.Sprintf("tree %s\nauthor %s\ncommitter %s\n\n%s",
			commit.Tree, commit.Author, commit.Commiter, commit.Message), nil
	case "Tag":
		tag, ok := o.(*Tag)
		if !ok {
			panic("object is of type tag, but cannot be cast to tag")
		}

		return fmt.Sprintf("object %s\n"+
			"type %s\n"+
			"tag %s\n"+
			"tagger %s\n\n"+
			"%s", tag.TargetOID(), strings.ToLower(tag.TargetType()), tag.Name(), tag.Tagger().String(), tag.Message()), nil
	}

	return "", fmt.Errorf("unknown object type")
}

func mapModeToType(mode os.FileMode) string {
	switch mode {
	case 16384:
		return "tree"
	case 0644:
		return "blob"
	default:
		return "unknown"
	}
}
