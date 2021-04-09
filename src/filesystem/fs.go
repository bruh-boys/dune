package filesystem

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FS interface {
	Open(name string) (File, error)
	OpenIfExists(name string) (File, error)
	OpenForWrite(name string) (File, error)
	OpenForAppend(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
	Write(name string, data []byte) error
	Append(name string, data []byte) error
	AppendPath(name string, data []byte) error
	Rename(oldPath, newPath string) error
	RemoveAll(path string) error
	Mkdir(name string) error
	MkdirAll(name string) error
	Chdir(dir string) error
	Getwd() (string, error)
	Abs(name string) (string, error)
	SetHome(name string) error
}

type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.WriterAt
	Stat() (os.FileInfo, error)
	Readdir(n int) ([]os.FileInfo, error)
	io.Writer
}

func Exists(fs FS, path string) bool {
	if _, err := fs.Stat(path); err != nil {
		return false
	}
	return true
}

func ReadAll(fs FS, path string) ([]byte, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	return b, err
}

func ReadDir(fs FS, dirname string) ([]os.FileInfo, error) {
	f, err := fs.Open(dirname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	list, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func WritePath(fs FS, path string, data []byte) error {
	path, err := fs.Abs(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := fs.MkdirAll(dir); err != nil {
			return err
		}
	}

	return fs.Write(path, data)
}
