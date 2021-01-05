package objects

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/furisto/gog/storage"
	hasher "github.com/furisto/gog/util"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var TreeType = []byte("tree")

type Tree struct {
	oid     string
	size    uint32
	Trees   []TreeEntry
	Blobs   []TreeEntry
	entries []TreeEntry
}

func NewTree(entries []TreeEntry) *Tree {
	tree := &Tree{
		entries: entries,
	}

	content := tree.getContent()
	tree.size = uint32(len(content))

	tree.oid = hasher.Hash(tree.getHeader(), content)
	return tree
}

func NewTreeFromDirectory(path string, prefix string) (*Tree, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	tree := Tree{}
	fileEntries, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range fileEntries {
		if f.Name() == ".git" {
			continue
		}

		if f.IsDir() {
			if !strings.HasPrefix(f.Name(), prefix) {
				continue
			}
			subTree, err := NewTreeFromDirectory(filepath.Join(path, f.Name()), "")
			if err != nil {
				return nil, err
			}

			tree.entries = append(tree.entries, TreeEntry{Mode: f.Mode(), Name: f.Name(), OID: subTree.OID(), Object: subTree})
		} else {
			if !strings.HasPrefix(f.Name(), prefix) {
				continue
			}
			blob, err := NewBlobFromFile(filepath.Join(path, f.Name()))
			if err != nil {
				return nil, err
			}

			tree.entries = append(tree.entries, TreeEntry{Mode: f.Mode(), Name: f.Name(), OID: blob.OID(), Object: blob})
		}
	}

	content := tree.getContent()
	tree.size = uint32(len(content))
	tree.oid = hasher.Hash(tree.getHeader(), content)

	return &tree, nil
}

func LoadTree(treeData []byte) (*Tree, error) {
	if !IsTree(treeData) {
		return nil, errors.New("not of type tree")
	}

	byteReader := bytes.NewReader(treeData[5:]) // start reading after "tree "
	reader := bufio.NewReader(byteReader)
	reader.Reset(byteReader)
	sizeSlice, err := reader.ReadString(byte(0))
	if err != nil {
		return nil, err
	}

	tree := new(Tree)

	sizeInt, err := strconv.Atoi(sizeSlice[:len(sizeSlice)-1])
	if err != nil {
		return nil, err
	}
	tree.size = uint32(sizeInt)

	for {
		modeString, err := reader.ReadString(' ')
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		modeString = modeString[:len(modeString)-1]
		modeNumber, err := strconv.ParseUint(modeString, 8, 32)
		if err != nil {
			return nil, err
		}

		mode := os.FileMode(modeNumber)

		name, err := reader.ReadString(byte(0))
		if err != nil {
			return nil, err
		}
		name = name[:len(name)-1] // remove null byte

		var oid = make([]byte, 20)
		_, err = io.ReadFull(reader, oid)
		if err != nil {
			return nil, err
		}

		treeEntry := TreeEntry{
			Mode: mode,
			Name: name,
			OID:  hex.EncodeToString(oid),
		}

		tree.entries = append(tree.entries, treeEntry)
	}

	return tree, nil
}

func IsTree(data []byte) bool {
	return bytes.HasPrefix(data, TreeType)
}

func (t *Tree) OID() string {
	return t.oid
}

func (t *Tree) SetOID(oid string) {
	t.oid = oid
}

func (t *Tree) Size() uint32 {
	return t.size
}

func (t *Tree) SetSize(size uint32) {
	t.size = size
}

func (t *Tree) Type() string {
	return "Tree"
}

func (t *Tree) Entries() []TreeEntry {
	sort.Sort(treeEntrySorter(t.entries))

	return t.entries
}

func (t *Tree) Bytes() []byte {
	header := t.getHeader()
	content := t.getContent()

	return append(header, content...)
}

func (t *Tree) Save(store storage.ObjectStore) error {
	entries := append(t.Trees, t.Blobs...)

	for _, entry := range entries {
		if err := entry.Object.Save(store); err != nil {
			return err
		}
	}

	return store.Put(t.OID(), t.Bytes())
}

func (t *Tree) getHeader() []byte {
	return []byte(fmt.Sprintf("tree %v\x00", t.Size()))
}

func (t *Tree) getContent() []byte {
	all := t.entries
	sort.Sort(treeEntrySorter(all))

	var b []byte
	for _, entry := range all {
		b = append(b, entry.Bytes()...)
	}

	return b
}

type TreeEntry struct {
	Mode   os.FileMode
	Name   string
	OID    string
	Object Object
}

func (t *TreeEntry) Bytes() []byte {
	var b []byte
	oid, _ := hex.DecodeString(t.OID)

	b = append(t.getHeader(), oid...)
	return b
}

func (t *TreeEntry) getHeader() []byte {
	return []byte(fmt.Sprintf("%o %v\x00", t.Mode, t.Name))
}

type treeEntrySorter []TreeEntry

func (tes treeEntrySorter) Len() int {
	return len(tes)
}

func (tes treeEntrySorter) Less(i int, j int) bool {
	return tes.adaptName(tes[i]) < tes.adaptName(tes[j])
}

func (tes treeEntrySorter) Swap(i int, j int) {
	tes[i], tes[j] = tes[j], tes[i]
}

func (*treeEntrySorter) adaptName(entry TreeEntry) string {
	if entry.Mode.IsDir() {
		return entry.Name + "/"
	}

	return entry.Name
}

type TreeBuilder struct {
	entries      []TreeEntry
	treeBuilders map[string]*TreeBuilder
}

func NewTreeBuilder() *TreeBuilder {
	return &TreeBuilder{
		treeBuilders: make(map[string]*TreeBuilder),
	}
}

func (tb *TreeBuilder) AddBlob(oid, name string, mode os.FileMode) {
	treeEntry := TreeEntry{
		OID:  oid,
		Name: name,
		Mode: mode,
	}

	tb.entries = append(tb.entries, treeEntry)
}

func (tb *TreeBuilder) AddSubTree(name string, treeBuilder *TreeBuilder) {
	tb.treeBuilders[name] = treeBuilder
}

func (tb *TreeBuilder) Build() *Tree {
	for name, treeBuilder := range tb.treeBuilders {
		tree := treeBuilder.Build()
		tb.entries = append(tb.entries, TreeEntry{
			OID:    tree.oid,
			Name:   name,
			Mode:   0o040000,
			Object: tree,
		})
	}

	sort.Sort(treeEntrySorter(tb.entries))
	return NewTree(tb.entries)
}
