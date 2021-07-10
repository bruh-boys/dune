package filesystem

import (
	"os"
	"path/filepath"
)

type LayerFS struct {
	fs  FS
	vfs *VirtualFS
}

// NewLayerFS allows to write layers on top of a regular filesystem.
// It is useful to write tests that need to be based on real code but with some
// additions or changes.
func NewLayerFS(fs FS) *LayerFS {
	vfs := NewVirtualFS()

	lf := &LayerFS{
		fs:  fs,
		vfs: vfs,
	}

	return lf
}

func (l *LayerFS) WriteLayer(name string, data []byte) error {
	return l.vfs.Write(name, data)
}

func (l *LayerFS) OpenForWriteLayer(name string) (File, error) {
	path, err := l.vfs.Abs(name)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := l.vfs.MkdirAll(dir); err != nil {
			return nil, err
		}
	}
	return l.vfs.OpenForWrite(name)
}

func (l *LayerFS) Abs(name string) (string, error) {
	if _, err := l.vfs.Stat(name); err == nil {
		return l.vfs.Abs(name)
	}

	return l.fs.Abs(name)
}

func (l *LayerFS) Rename(oldPath, newPath string) error {
	if _, err := l.vfs.Stat(oldPath); err == nil {
		return l.vfs.Rename(oldPath, newPath)
	}

	return l.fs.Rename(oldPath, newPath)
}

func (l *LayerFS) RemoveAll(path string) error {
	if _, err := l.vfs.Stat(path); err == nil {
		return l.vfs.RemoveAll(path)
	}
	return l.fs.RemoveAll(path)
}

func (l *LayerFS) Open(name string) (File, error) {
	if _, err := l.vfs.Stat(name); err == nil {
		return l.vfs.Open(name)
	}
	return l.fs.Open(name)
}

func (l *LayerFS) OpenIfExists(name string) (File, error) {
	if _, err := l.vfs.Stat(name); err == nil {
		return l.vfs.Open(name)
	}
	return l.fs.OpenIfExists(name)
}

func (l *LayerFS) OpenForWrite(name string) (File, error) {
	return l.fs.OpenForWrite(name)
}

func (l *LayerFS) OpenForAppend(name string) (File, error) {
	if _, err := l.vfs.Stat(name); err == nil {
		return l.vfs.OpenForAppend(name)
	}
	return l.fs.OpenForAppend(name)
}

func (l *LayerFS) Stat(name string) (os.FileInfo, error) {
	if fi, err := l.vfs.Stat(name); err == nil {
		return fi, nil
	}
	return l.fs.Stat(name)
}

func (l *LayerFS) Write(name string, data []byte) error {
	return l.fs.Write(name, data)
}

func (l *LayerFS) Append(name string, data []byte) error {
	return l.fs.Append(name, data)
}

func (l *LayerFS) AppendPath(path string, data []byte) error {
	return l.fs.AppendPath(path, data)
}

func (l *LayerFS) Mkdir(name string) error {
	return l.fs.Mkdir(name)
}

func (l *LayerFS) MkdirAll(name string) error {
	return l.fs.MkdirAll(name)
}

func (l *LayerFS) Chdir(dir string) error {
	return l.fs.Chdir(dir)
}

func (l *LayerFS) Getwd() (string, error) {
	return l.fs.Getwd()
}

func (l *LayerFS) SetHome(name string) error {
	return l.fs.SetHome(name)
}
