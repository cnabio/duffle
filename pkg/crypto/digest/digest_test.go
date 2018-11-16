package digest

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestOfReader(t *testing.T) {
	testString := "hello world!"
	buf := bytes.NewBufferString(testString)
	newBuf, tag, err := OfReader(buf)
	if err != nil {
		t.Error(err)
	}
	expectedTag := "7509e5bda0c762d2bac7f90d758b5b2263fa01cc"
	if tag != expectedTag {
		t.Errorf("expected '%s', got '%s'", expectedTag, tag)
	}

	// now check that the buffer we got back hasn't been tampered with
	endBuf, err := ioutil.ReadAll(newBuf)
	if err != nil {
		t.Error(err)
	}
	if string(endBuf) != testString {
		t.Errorf("expected '%s', got '%s'", testString, string(endBuf))
	}
}

func TestOfBuffer(t *testing.T) {
	testString := "hello world!"
	tag, err := OfBuffer([]byte(testString))
	if err != nil {
		t.Error(err)
	}
	expectedTag := "7509e5bda0c762d2bac7f90d758b5b2263fa01cc"
	if tag != expectedTag {
		t.Errorf("expected '%s', got '%s'", expectedTag, tag)
	}
}
