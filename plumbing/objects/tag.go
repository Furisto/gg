package objects

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/furisto/gog/plumbing/refs"
	"github.com/furisto/gog/storage"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

// The on disk representation of a git tag is
// tag [body_size]\x00object [oid of git object that the tag points to]
// type [type of git object pointed to]
// tag [name of tag]
// tagger [signature]
//
// [commit message]
//
// An example would be
// tag

var tagType = []byte("Tag")

type Tag struct {
	ref        refs.Ref
	size       uint32
	oid        string
	targetOID  string
	targetType string
	name       string
	tagger     *Signature
	message    string
}

func (t *Tag) Size() uint32 {
	return t.size
}

func (t *Tag) OID() string {
	return t.oid
}

func (t *Tag) Type() string {
	return "Tag"
}

func (t *Tag) Save(store storage.ObjectStore) error {
	var buf bytes.Buffer
	if err := t.EncodeTag(&buf); err != nil {
		return err
	}
	return store.Put(t.oid, buf.Bytes())
}

func (t *Tag) Name() string {
	return t.name
}

func (t *Tag) Tagger() *Signature {
	return t.tagger
}

func (t *Tag) TargetOID() string {
	return t.targetOID
}

func (t *Tag) TargetType() string {
	return t.targetType
}

func (t *Tag) Message() string {
	return t.message
}

func IsTag(data []byte) bool {
	return bytes.HasPrefix(data, tagType)
}

func NewTag(targetOID, targetType, name string, tagger *Signature, message string) (*Tag, error) {
	tag := &Tag{
		targetOID:  targetOID,
		targetType: targetType,
		name:       name,
		tagger:     tagger,
		message:    message,
	}

	var buf bytes.Buffer
	if err := tag.writeBody(&buf); err != nil {
		return nil, err
	}

	hasher := sha1.New()
	_, err := hasher.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	tag.oid = hex.EncodeToString(hasher.Sum(nil))
	tag.size = uint32(buf.Len())

	return tag, nil
}

func (t *Tag) EncodeTag(writer io.Writer) error {
	if err := t.writeHeader(writer); err != nil {
		return err
	}

	if err := t.writeBody(writer); err != nil {
		return err
	}

	return nil
}

func (t *Tag) writeHeader(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, "tag %d\x00", t.size)
	return err
}

func (t *Tag) writeBody(writer io.Writer) error {
	if _, err := fmt.Fprintf(writer, "object %s\n", t.targetOID); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "type %s\n", t.targetType); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "tag %s\n", t.name); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "tagger %s\n\n", t.tagger.String()); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "%s\n", t.message); err != nil {
		return err
	}

	return nil
}

func DecodeTag(reader io.Reader) (*Tag, error) {
	buf := bufio.NewReader(reader)

	objectType, err := buf.ReadString(byte(' '))
	if err != nil {
		return nil, err
	}

	if objectType != "tag " {
		return nil, fmt.Errorf("object is not of type tag")
	}

	sizeBytes, err := buf.ReadString(byte(0))
	if err != nil {
		return nil, err
	}

	size, err := strconv.Atoi(sizeBytes[:len(sizeBytes)-1])
	if err != nil {
		return nil, err
	}

	targetOID, err := readEntry(buf, "object")
	if err != nil {
		return nil, err
	}

	targetType, err := readEntry(buf, "type")
	if err != nil {
		return nil, err
	}

	tagName, err := readEntry(buf, "tag")
	if err != nil {
		return nil, err
	}

	tagger, err := readEntry(buf, "tagger")
	if err != nil {
		return nil, err
	}

	signature, err := DecodeSignature([]byte(tagger))
	if err != nil {
		return nil, err
	}

	if _, err := buf.ReadBytes('\n'); err != nil {
		return nil, err
	}

	messageBytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	message := string(messageBytes)
	message = strings.TrimSuffix(message, "\n")

	return &Tag{
		oid:        "",
		size:       uint32(size),
		targetOID:  targetOID,
		targetType: targetType,
		name:       tagName,
		tagger:     signature,
		message:    message,
	}, nil
}

func readEntry(reader *bufio.Reader, name string) (string, error) {
	entryName, err := reader.ReadString(' ')
	if err != nil {
		return "", err
	}

	if entryName[:len(entryName)-1] != name {
		return "", fmt.Errorf("tag is corrupted, could not find %s", name)
	}

	entryValue, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return entryValue[:len(entryValue)-1], err
}
