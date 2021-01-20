package objects

import (
	"bytes"
	"github.com/furisto/gog/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

const tagFilePath = "./testdata/decode_tag"

func TestDecodeTag(t *testing.T) {
	data, err := util.DecompressFile(tagFilePath)
	if err != nil {
		t.Fatalf("could not read golden file at %s", tagFilePath)
	}

	tag, err := DecodeTag(bytes.NewReader(data))
	if err != nil {
		t.Errorf("")
		return
	}

	assert.Equal(t, uint32(161), tag.Size(), "tag object size")
	assert.Equal(t, "404ab0364d9ca3f06936d7c7c97c1d2de1e696f3", tag.TargetOID(), "target oid")
	assert.Equal(t, "commit", tag.TargetType(), "target type")
	assert.Equal(t, "annotated", tag.Name(), "name")
	assert.Equal(t, "annotated", tag.Message(), "message")
}

// todo: needs timezone handling
//func TestEncodeTag(t *testing.T) {
//	tagger := Signature{
//		Name: "Furisto",
//		Email: "24721048+Furisto@users.noreply.github.com",
//		TimeStamp: time.Unix(int64(1609334881),int64(0)),
//	}
//	tag, err := NewTag(
//		"404ab0364d9ca3f06936d7c7c97c1d2de1e696f3", "commit","annotated",&tagger, "annotated")
//	if err != nil {
//		t.Fatalf("")
//	}
//
//	var actual bytes.Buffer
//	if err := tag.EncodeTag(&actual); err != nil {
//		t.Errorf("")
//		return
//	}
//
//	expected, err := util.DecompressFile(tagFilePath)
//	if err != nil {
//		t.Fatalf("")
//	}
//
//	assert.Equal(t, expected, actual.Bytes())
//}
