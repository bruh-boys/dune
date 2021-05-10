package filesystem

import (
	"errors"
	"os"
)

var ErrReadOnly = errors.New("this file is readonly")

// ReadOnlyFS implements fileSystem that serves a directory as root
type ReadOnlyFS struct {
	fs FS
}

func NewReadOnlyFS(fs FS) *ReadOnlyFS {
	return &ReadOnlyFS{fs: fs}
}

// Sets the home directory
func (r *ReadOnlyFS) SetHome(name string) error {
	return r.fs.SetHome(name)
}

func (r *ReadOnlyFS) Abs(name string) (string, error) {
	return r.fs.Abs(name)
}

func (r *ReadOnlyFS) Rename(oldPath, newPath string) error {
	return ErrReadOnly
}

func (r *ReadOnlyFS) RemoveAll(name string) error {
	return ErrReadOnly
}

func (r *ReadOnlyFS) OpenForAppend(name string) (File, error) {
	return nil, ErrReadOnly
}

func (r *ReadOnlyFS) OpenForWrite(name string) (File, error) {
	return nil, ErrReadOnly
}

func (r *ReadOnlyFS) Open(name string) (File, error) {
	return r.fs.Open(name)
}

func (r *ReadOnlyFS) OpenIfExists(name string) (File, error) {
	return r.fs.OpenIfExists(name)
}

func (r *ReadOnlyFS) Stat(name string) (os.FileInfo, error) {
	return r.fs.Stat(name)
}

func (r *ReadOnlyFS) Write(name string, data []byte) error {
	return ErrReadOnly
}

func (r *ReadOnlyFS) Append(name string, data []byte) error {
	return ErrReadOnly
}

func (r *ReadOnlyFS) AppendPath(name string, data []byte) error {
	return ErrReadOnly
}

func (r *ReadOnlyFS) Mkdir(name string) error {
	return ErrReadOnly
}

func (r *ReadOnlyFS) MkdirAll(name string) error {
	return ErrReadOnly
}

func (r *ReadOnlyFS) Chdir(name string) error {
	return r.fs.Chdir(name)
}

func (r *ReadOnlyFS) Getwd() (string, error) {
	return r.fs.Getwd()
}
