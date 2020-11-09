package objects

import "github.com/furisto/gog/storage"

type CommitIterator interface {
	MoveNext() bool
	Current() *Commit
}

type SimpleCommitIterator struct {
	start   *Commit
	current *Commit
	store   storage.ObjectStore
}

func NewCommitIterator(commit *Commit, store storage.ObjectStore) *SimpleCommitIterator {
	return &SimpleCommitIterator{
		start: commit,
		store: store,
	}
}

func (ci *SimpleCommitIterator) MoveNext() bool {
	if ci.current == nil {
		ci.current = ci.start
		return true
	}

	if len(ci.current.Parents) == 0 {
		return false
	}

	data, err := ci.store.Get(ci.current.Parents[0])
	if err != nil {
		return false
	}

	commit, err := DecodeCommit(ci.current.Parents[0], data)
	if err != nil {
		return false
	}

	ci.current = commit
	return true
}

func (ci *SimpleCommitIterator) Current() *Commit {
	return ci.current
}

type SkipCommitIterator struct {
	current uint64
	skips   uint64
	inner   CommitIterator
}

func NewSkipCommitIterator(iterator CommitIterator, skips uint64) *SkipCommitIterator {
	return &SkipCommitIterator{
		skips: skips,
		inner: iterator,
	}
}

func (sci *SkipCommitIterator) MoveNext() bool {
	for sci.current < sci.skips {
		if !sci.inner.MoveNext() {
			return false
		}
		sci.current++
	}

	return sci.inner.MoveNext()
}

func (si *SkipCommitIterator) Current() *Commit {
	return si.inner.Current()
}

type TakeCommitIterator struct {
	current uint64
	take    uint64
	inner   CommitIterator
}

func NewTakeCommitIterator(iterator CommitIterator, take uint64) *TakeCommitIterator {
	return &TakeCommitIterator{
		take:  take,
		inner: iterator,
	}
}

func (tci *TakeCommitIterator) MoveNext() bool {
	if tci.current < tci.take {
		if !tci.inner.MoveNext() {
			return false
		}
		tci.current++
		return true
	}

	return false
}

func (tci *TakeCommitIterator) Current() *Commit {
	return tci.inner.Current()
}

type FilterCommitIterator struct {
	filter func(commit *Commit) bool
	inner  CommitIterator
	commit *Commit
}

func NewFilterCommitIterator(iterator CommitIterator, filter func(commit *Commit) bool) *FilterCommitIterator {
	return &FilterCommitIterator{
		filter: filter,
		inner:  iterator,
	}
}

func (fci *FilterCommitIterator) MoveNext() bool {
	for fci.inner.MoveNext() {
		c := fci.inner.Current()
		if fci.filter(c) {
			fci.commit = c
			return true
		}
	}

	return false
}

func (fci *FilterCommitIterator) Current() *Commit {
	return fci.commit
}
