package objects

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/furisto/gog/config"
	"github.com/furisto/gog/storage"
	hasher "github.com/furisto/gog/util"
	"io"
	"io/ioutil"
	"strconv"
	"time"
)

//Git commit is represented on the file system as:
//commit [body_size]\x00[body] where body consists of
//tree [tree_oid]
//parent [parent_oid]
//author [name] <[email]> [author_date(in unix time)] [UTC_offset]
//committer [name] <[email]> [committer_date(in unix time)] [UTC_offset]
//
//[commit message]
//An example would be
//commit 256\x00tree 26806b616ec527a4e27ed1554a20931a872d39b0
//parent 667b37d33991249d2fbf89c7f3f30ea7026ce774
//author Furisto <24721048+Furisto@users.noreply.github.com> 1567343397 +0200
//committer Furisto <24721048+Furisto@users.noreply.github.com> 1567343397 +0200

//Hello from Git!

var CommitType = []byte("commit")
var ParentPrefix = []byte("parent")
var AuthorPrefix = []byte("author")
var CommitterPrefix = []byte("committer")

type Commit struct {
	oid      string
	size     uint32
	Tree     string
	Parents  []string
	Author   *Signature
	Commiter *Signature
	Message  string
}

func DecodeCommit(oid string, data []byte) (*Commit, error) {
	if !IsCommit(data) {
		return nil, errors.New("not of type commit")
	}

	br := bytes.NewReader(data[7:])
	reader := bufio.NewReader(br)

	sizeString, err := reader.ReadString(byte(0))
	if err != nil {
		return nil, err
	}

	sizeInt, err := strconv.Atoi(sizeString[:len(sizeString)-1])
	if err != nil {
		return nil, err
	}

	_, err = reader.ReadBytes(' ')
	if err != nil {
		return nil, err
	}

	treeOid, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	treeOid = treeOid[:len(treeOid)-1]

	var parentOids []string
	for {
		parentPrefix, err := reader.Peek(6)
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(parentPrefix, []byte("parent")) {
			break
		}

		_, err = reader.ReadBytes(' ')
		if err != nil {
			return nil, err
		}

		parentOid, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		parentOids = append(parentOids, parentOid[:len(parentOid)-1])
	}

	prefix, err := reader.ReadBytes(' ')
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(prefix[:len(prefix)-1], AuthorPrefix) {
		return nil, errors.New("commit does not have author")
	}

	authorBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	authorSig, err := DecodeSignature(authorBytes)
	if err != nil {
		return nil, err
	}

	prefix, err = reader.ReadBytes(' ')
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(prefix[:len(prefix)-1], CommitterPrefix) {
		return nil, errors.New("commit does not have author")
	}

	committerBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	committerSig, err := DecodeSignature(committerBytes)
	if err != nil {
		return nil, err
	}

	if _, err := reader.ReadBytes('\n'); err != nil {
		return nil, err
	}

	message, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &Commit{
		oid:      oid,
		size:     uint32(sizeInt),
		Tree:     treeOid,
		Parents:  parentOids,
		Author:   authorSig,
		Commiter: committerSig,
		Message:  string(message),
	}, nil

}

func EncodeCommit(commit *Commit) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	if err := commit.writeHeader(buf); err != nil {
		return nil, err
	}

	// tree
	if _, err := fmt.Fprintf(buf, "tree %s", commit.Tree); err != nil {
		return nil, err
	}

	// parents
	for _, parent := range commit.Parents {
		if _, err := fmt.Fprintf(buf, "\nparent %s", parent); err != nil {
			return nil, err
		}
	}

	// author
	if _, err := fmt.Fprintf(buf, "\nauthor "); err != nil {
		return nil, err
	}

	if err := commit.Author.Encode(buf); err != nil {
		return nil, err
	}

	// committer
	if _, err := fmt.Fprintf(buf, "\ncommitter "); err != nil {
		return nil, err
	}

	if err := commit.Commiter.Encode(buf); err != nil {
		return nil, err
	}

	// message
	if _, err := fmt.Fprintf(buf, "\n\n%v", commit.Message); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func IsCommit(data []byte) bool {
	return bytes.HasPrefix(data, CommitType)
}

func (c *Commit) OID() string {
	return c.oid
}

func (c *Commit) SetOID(oid string) {
	c.oid = oid
}

func (c *Commit) Size() uint32 {
	return c.size
}

func (c *Commit) SetSize(size uint32) {
	c.size = size
}

func (c *Commit) Type() string {
	return "Commit"
}

func (c *Commit) Save(store storage.ObjectStore) error {
	data, err := EncodeCommit(c)
	if err != nil {
		return err
	}

	err = store.Put(c.OID(), data)
	return err
}

func (c *Commit) writeHeader(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, "commit %d\x00", c.size)
	return err
}

func (c *Commit) writeContent(buf io.Writer) error {
	// tree
	if _, err := fmt.Fprintf(buf, "tree %s", c.Tree); err != nil {
		return err
	}

	// parents
	for _, parent := range c.Parents {
		if _, err := fmt.Fprintf(buf, "\nparent %s", parent); err != nil {
			return err
		}
	}

	// author
	if _, err := fmt.Fprintf(buf, "\nauthor "); err != nil {
		return err
	}

	if err := c.Author.Encode(buf); err != nil {
		return err
	}

	// committer
	if _, err := fmt.Fprintf(buf, "\ncommitter "); err != nil {
		return err
	}

	if err := c.Commiter.Encode(buf); err != nil {
		return err
	}

	// message
	if _, err := fmt.Fprintf(buf, "\n\n%v", c.Message); err != nil {
		return err
	}

	return nil
}

type CommitBuilder struct {
	tree           string
	authorName     string
	authorEmail    string
	committerName  string
	committerEmail string
	parentOids     []string
	message        string
	config         config.Config
	hook           func(*Commit)
}

func NewCommitBuilder(treeOid string) *CommitBuilder {
	return &CommitBuilder{
		tree:   treeOid,
		config: &config.NilConfig{},
	}
}

func (cb *CommitBuilder) WithParent(parentOid string) *CommitBuilder {
	if parentOid != "" {
		cb.parentOids = append(cb.parentOids, parentOid)
	}
	return cb
}

func (cb *CommitBuilder) WithAuthor(authorName, authorEmail string) *CommitBuilder {
	cb.authorName = authorName
	cb.authorEmail = authorEmail
	return cb
}

func (cb *CommitBuilder) WithCommitter(commiterName, commiterEmail string) *CommitBuilder {
	cb.committerName = commiterName
	cb.committerEmail = commiterEmail
	return cb
}

func (cb *CommitBuilder) WithMessage(message string) *CommitBuilder {
	cb.message = message
	return cb
}

func (cb *CommitBuilder) WithConfig(cfg config.Config) *CommitBuilder {
	cb.config = cfg
	return cb
}

func (cb *CommitBuilder) WithHook(hook func(*Commit)) {
	cb.hook = hook
}

func (cb *CommitBuilder) Build() (*Commit, error) {
	var err error

	if cb.authorName == "" && cb.authorEmail == "" {
		if cb.authorName, err = cb.config.Get("user", "name"); err != nil {
			cb.authorName = "unknown"
		}

		if cb.authorEmail, err = cb.config.Get("user", "email"); err != nil {
			cb.authorEmail = "unknown"
		}
	}

	if cb.committerName == "" && cb.committerEmail == "" {
		if cb.committerName, err = cb.config.Get("user", "name"); err != nil {
			cb.committerName = "unknown"
		}

		if cb.committerEmail, err = cb.config.Get("user", "email"); err != nil {
			cb.committerEmail = "unknown"
		}
	}

	c := Commit{
		Tree:     cb.tree,
		Parents:  cb.parentOids,
		Author:   &Signature{Name: cb.authorName, Email: cb.authorEmail, TimeStamp: time.Now()},
		Commiter: &Signature{Name: cb.committerName, Email: cb.committerEmail, TimeStamp: time.Now()},
		Message:  cb.message,
	}

	if cb.hook != nil {
		cb.hook(&c)
	}

	var content bytes.Buffer
	if err := c.writeContent(&content); err != nil {
		return nil, err
	}

	c.size = uint32(len(content.Bytes()))

	var header bytes.Buffer
	if err := c.writeHeader(&header); err != nil {
		return nil, err
	}

	c.oid = hasher.Hash(header.Bytes(), content.Bytes())
	return &c, nil
}

type Signature struct {
	Name      string
	Email     string
	TimeStamp time.Time
}

func DecodeSignature(data []byte) (*Signature, error) {
	splits := bytes.Split(data, []byte(" "))
	if len(splits) < 2 {
		return nil, errors.New("signature is corrupt")
	}

	name := splits[0]
	mail := bytes.Trim(splits[1], "<>")

	n, err := strconv.ParseInt(string(splits[2]), 10, 64)
	if err != nil {
		return nil, err
	}
	timestamp := time.Unix(n, 0)

	return &Signature{
		Name:      string(name),
		Email:     string(mail),
		TimeStamp: timestamp,
	}, nil
}

func (t *Signature) Encode(writer io.Writer) error {
	if _, err := fmt.Fprintf(writer, "%s <%s>", t.Name, t.Email); err != nil {
		return err
	}

	_, err := fmt.Fprintf(writer, " %d %s", t.TimeStamp.Unix(), t.TimeStamp.Format("+0000"))
	return err
}

func (s Signature) String() string {
	return fmt.Sprintf("%s <%s> %d %s",
		s.Name, s.Email, s.TimeStamp.Unix(), s.TimeStamp.Format("+0000")) // todo: handle timezones
}
