package filesystem

import (
	"testing"
)

func TestLayerFS(t *testing.T) {
	fs := NewLayerFS(NewVirtualFS())

	if err := fs.Write("test.txt", []byte("test")); err != nil {
		t.Fatal(err)
	}

	if err := fs.WriteLayer("test2.txt", []byte("test2")); err != nil {
		t.Fatal(err)
	}

	b, err := ReadAll(fs, "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "test" {
		t.Fatal(string(b))
	}

	b, err = ReadAll(fs, "test2.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "test2" {
		t.Fatal(string(b))
	}

	if err := fs.WriteLayer("test.txt", []byte("test3")); err != nil {
		t.Fatal(err)
	}

	b, err = ReadAll(fs, "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "test3" {
		t.Fatal(string(b))
	}
}
