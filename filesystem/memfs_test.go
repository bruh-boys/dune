package filesystem

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestOpenForWrite(t *testing.T) {
	fs := NewMemFS()

	f, err := fs.OpenForWrite("foo.txt")
	if err != nil {
		t.Fatal(err)
	}

	n, err := io.WriteString(f, "lalala")
	if err != nil {
		t.Fatal(err)
	}
	if n != 6 {
		t.Fatal(n)
	}
	f.Close()

	f, err = fs.Open("foo.txt")
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	if string(b) != "lalala" {
		t.Fatal(string(b))
	}

}

func TestDelete(t *testing.T) {
	fs := NewMemFS()

	err := fs.Write("foo.txt", []byte("whatever1"))
	if err != nil {
		t.Fatal(err)
	}

	err = fs.RemoveAll("foo.txt")
	if err != nil {
		t.Fatal(err)
	}

	f, err := fs.Open(".")
	if err != nil {
		t.Fatal(err)
	}

	files, err := f.Readdir(-1)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatal(len(files))
	}
}

func TestDelete2(t *testing.T) {
	fs := NewMemFS()

	err := WritePath(fs, "foo/bar1.txt", []byte("whatever1"))
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Write("foo/bar2.txt", []byte("whatever1"))
	if err != nil {
		t.Fatal(err)
	}

	err = fs.RemoveAll("foo")
	if err != nil {
		t.Fatal(err)
	}

	f, err := fs.Open(".")
	if err != nil {
		t.Fatal(err)
	}

	files, err := f.Readdir(-1)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatal(len(files))
	}
}

func TestRename(t *testing.T) {
	fs := NewMemFS()

	err := fs.Write("foo.txt", []byte("whatever1"))
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Rename("foo.txt", "foo2.txt")
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Open("foo2.txt")
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Open("foo.txt")
	if err != os.ErrNotExist {
		t.Fatal(err)
	}
}

func TestVF1(t *testing.T) {
	fs := NewMemFS()

	fi, err := fs.Stat("/")
	if err != nil {
		t.Fatal(err)
	}

	if !fi.IsDir() {
		t.Fail()
	}
}

func TestVF2(t *testing.T) {
	fs := NewMemFS()

	err := fs.Mkdir("/foo")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Write("/bar.txt", []byte("whatever"))
	if err != nil {
		t.Fatal(err)
	}

	fi, err := fs.Stat("/bar.txt")
	if err != nil {
		t.Fatal(err)
	}

	if fi.IsDir() {
		t.Fail()
	}
}

func TestVF22(t *testing.T) {
	fs := NewMemFS()

	if err := fs.MkdirAll("/foo/bar"); err != nil {
		t.Fatal(err)
	}

	if err := fs.MkdirAll("/foo/bar/duck"); err != nil {
		t.Fatal(err)
	}

	fi, err := fs.Stat("/foo/bar/duck")
	if err != nil {
		t.Fatal(err)
	}

	if !fi.IsDir() {
		t.Fail()
	}
}

func TestVF3(t *testing.T) {
	fs := NewMemFS()

	err := fs.Mkdir("/foo/")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Mkdir("/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Write("/foo/bar/bar.txt", []byte("whatever"))
	if err != nil {
		t.Fatal(err)
	}

	fi, err := fs.Stat("/foo/bar/bar.txt")
	if err != nil {
		t.Fatal(err)
	}

	if fi.IsDir() {
		t.Fail()
	}
}

func TestVF4(t *testing.T) {
	fs := NewMemFS()

	err := fs.Write("bar.txt", []byte("whatever"))
	if err != nil {
		t.Fatal(err)
	}

	// read it more than once
	for i := 0; i < 2; i++ {
		f, err := fs.Open("bar.txt")
		if err != nil {
			t.Fatal(err)
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}

		// need to close it for the next read
		f.Close()

		if string(b) != "whatever" {
			t.Fatalf("got: %s", string(b))
		}
	}
}

func TestVF5(t *testing.T) {
	fs := NewMemFS()

	err := fs.Write("foo.txt", []byte("whatever1"))
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Write("bar.txt", []byte("whatever2"))
	if err != nil {
		t.Fatal(err)
	}

	f, err := fs.Open(".")
	if err != nil {
		t.Fatal(err)
	}

	files, err := f.Readdir(-1)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatal(len(files))
	}
}

func TestVF6(t *testing.T) {
	fs := NewMemFS()

	err := fs.Mkdir("/foo/")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Write("/foo/foo.txt", []byte("whatever1"))
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Write("/foo/bar.txt", []byte("whatever2"))
	if err != nil {
		t.Fatal(err)
	}

	f, err := fs.Open("/foo/")
	if err != nil {
		t.Fatal(err)
	}

	files, err := f.Readdir(-1)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatal(len(files))
	}
}
