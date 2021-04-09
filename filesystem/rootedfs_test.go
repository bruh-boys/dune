package filesystem

import (
	"testing"
)

func TestRootedFSAbs(t *testing.T) {
	v := NewMemFS()
	assertErr(WritePath(v, "/demo/test/test.txt", []byte{}), t)

	fs, err := NewRootedFS("/demo/", v)
	if err != nil {
		t.Fatal(err)
	}

	abs, err := fs.Abs("test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if abs != "/test.txt" {
		t.Fatalf("Got %s", abs)
	}

	if err := fs.Chdir("test"); err != nil {
		t.Fatal(err)
	}

	abs, err = fs.Abs("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if abs != "/test/test.txt" {
		t.Fatalf("Got %s", abs)
	}
}

func TestRootedFS(t *testing.T) {
	v := NewMemFS()
	assertErr(WritePath(v, "/demo/test/foo/test.txt", []byte{}), t)
	assertErr(WritePath(v, "/demo/test/test.txt", []byte{}), t)
	assertErr(WritePath(v, "/demo/users/foo/data.txt", []byte{}), t)

	fs, err := NewRootedFS("/demo/test", v)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Stat("/demo/test/foo/test.txt"); err == nil {
		t.Fatal("Should fail")
	}

	if _, err := fs.Stat("/demo/test/test.txt"); err == nil {
		t.Fatal("Should fail")
	}

	if _, err := fs.Stat("//demo/users/foo/data.txt"); err == nil {
		t.Fatal("Should fail")
	}

	if _, err := fs.Stat("test.txt"); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Stat("foo/test.txt"); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Stat("/foo/test.txt"); err != nil {
		t.Fatal(err)
	}
}

func assertErr(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}
