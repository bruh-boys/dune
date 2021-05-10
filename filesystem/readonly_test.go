package filesystem

import (
	"testing"
)

func TestReadonly(t *testing.T) {
	v := NewVirtualFS()
	assertErr(WritePath(v, "test.txt", []byte("test")), t)
	fs := NewReadOnlyFS(v)

	b, err := ReadAll(fs, "test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "test" {
		t.Fatal(string(b))
	}

	if err := WritePath(fs, "test.txt", []byte{}); err != ErrReadOnly {
		t.Fatal(err)
	}
}
