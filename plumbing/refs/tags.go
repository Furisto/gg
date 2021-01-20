package refs

import (
	"fmt"
	"strings"
)

var genericTagError = "%s is not a valid ref name"

// see https://git-scm.com/docs/git-check-ref-format
func ValidateRef(tagName string) error {
	if tagName == "" {
		return fmt.Errorf("empty tag name")
	}

	// must contain at least one /
	if !strings.Contains(tagName, "/") {
		return fmt.Errorf(genericTagError, tagName)
	}

	// slash-separated components cannot start with . or end with .lock
	components := strings.Split(tagName, "/")
	for _, c := range components {
		if strings.HasPrefix(c, ".") || strings.HasSuffix(c, ".lock") {
			return fmt.Errorf(genericTagError, tagName)
		}
	}

	// cannot be @
	if tagName == "@" {
		return fmt.Errorf(genericTagError, tagName)
	}

	// cannot contain
	for _, p := range []string{"@{", "\\", "..", " ", "~", "^", ":", "?", "*", "["} {
		if strings.Contains(tagName, p) {
			return fmt.Errorf(genericTagError, tagName)
		}
	}

	// cannot end with .
	if strings.HasSuffix(tagName, ".") {
		return fmt.Errorf(genericTagError, tagName)
	}

	return nil
}

func ShortTagname(tagName string) string {
	return strings.TrimPrefix(tagName, "refs/tags/")
}
