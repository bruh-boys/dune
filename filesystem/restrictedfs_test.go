package filesystem

import (
	"testing"
)

func TestRestrictedFSAbs(t *testing.T) {
	v := NewMemFS()
	assertErr(WritePath(v, "/demo/test/test.txt", []byte{}), t)
	assertErr(WritePath(v, "/demo/public/foo.txt", []byte{}), t)
	assertErr(WritePath(v, "/demo/private/bar.txt", []byte{}), t)
	assertErr(WritePath(v, "/demo/foo.txt", []byte{}), t)

	fs, err := NewRestrictedFS(v)
	if err != nil {
		t.Fatal(err)
	}

	assertErr(fs.AddToWhitelist("/demo/test"), t)
	assertErr(fs.AddToWhitelist("/demo/public"), t)

	if _, err := fs.Stat("/demo/test/test.txt"); err != nil {
		t.Fatal()
	}

	if _, err := fs.Stat("/demo/public/foo.txt"); err != nil {
		t.Fatal()
	}

	if _, err := fs.Stat("/demo/private/bar.txt"); err == nil {
		t.Fatal(err)
	}

	if _, err := fs.Stat("/demo/foo.txt"); err == nil {
		t.Fatal(err)
	}
}

func TestRestrictedFSBlacklist(t *testing.T) {
	v := NewMemFS()
	assertErr(WritePath(v, "/demo/test/test.txt", []byte{}), t)
	assertErr(WritePath(v, "/demo/foo.txt", []byte{}), t)

	fs, err := NewRestrictedFS(v)
	if err != nil {
		t.Fatal(err)
	}

	fs.AddToWhitelist("/")

	assertErr(fs.AddToBlacklist("/demo/test"), t)

	if _, err := fs.Stat("/demo/test/test.txt"); err == nil {
		t.Fatal()
	}

	if _, err := fs.Stat("/demo/foo.txt"); err != nil {
		t.Fatal(err)
	}
}
