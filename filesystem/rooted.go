package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RootedFS implements fileSystem that serves a directory as root
type RootedFS struct {
	fs         FS
	root       string
	workingDir string
}

func NewRootedFS(root string, fs FS) (*RootedFS, error) {
	a, err := fs.Abs(root)
	if err != nil {
		return nil, err
	}
	return &RootedFS{fs: fs, root: a, workingDir: "/"}, nil
}

// Sets the home directory
func (r *RootedFS) SetHome(name string) error {
	return nil
}

func (r *RootedFS) Abs(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty path")
	}

	var abs string
	switch name[0] {
	case '/':
		abs = name
	default:
		abs = filepath.Join(r.workingDir, name)
	}

	return abs, nil
}

func (r *RootedFS) innerFSAbs(name string) (string, error) {
	abs, err := r.Abs(name)
	if err != nil {
		return "", err
	}

	if abs[0] == '/' {
		abs = abs[1:]
	}

	abs = filepath.Join(r.root, abs)

	// this is redundant. Remove?
	if !strings.HasPrefix(abs, r.root) {
		return "", fmt.Errorf("invalid path")
	}

	return abs, nil
}

func (r *RootedFS) Rename(oldPath, newPath string) error {
	var err error
	oldPath, err = r.innerFSAbs(oldPath)
	if err != nil {
		return err
	}
	newPath, err = r.innerFSAbs(newPath)
	if err != nil {
		return err
	}
	return r.fs.Rename(oldPath, newPath)
}

func (r *RootedFS) RemoveAll(name string) error {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return err
	}

	return r.fs.RemoveAll(name)
}

func (r *RootedFS) OpenForAppend(name string) (File, error) {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return nil, err
	}

	return r.fs.OpenForAppend(name)
}

func (r *RootedFS) OpenForWrite(name string) (File, error) {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return nil, err
	}

	return r.fs.OpenForWrite(name)
}

func (r *RootedFS) Open(name string) (File, error) {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return nil, err
	}

	return r.fs.Open(name)
}

func (r *RootedFS) OpenIfExists(name string) (File, error) {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return nil, err
	}

	f, err := r.fs.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

func (r *RootedFS) Stat(name string) (os.FileInfo, error) {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return nil, err
	}
	return r.fs.Stat(name)
}

func (r *RootedFS) Write(name string, data []byte) error {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return err
	}
	return r.fs.Write(name, data)
}

func (r *RootedFS) Append(name string, data []byte) error {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return err
	}

	return r.fs.Append(name, data)
}

func (r *RootedFS) AppendPath(name string, data []byte) error {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return err
	}

	dir := filepath.Dir(name)
	if dir != "." {
		if err := r.fs.MkdirAll(dir); err != nil {
			return err
		}
	}
	return r.fs.Append(name, data)
}

func (r *RootedFS) Mkdir(name string) error {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return err
	}

	return r.fs.Mkdir(name)
}

func (r *RootedFS) MkdirAll(name string) error {
	var err error
	name, err = r.innerFSAbs(name)
	if err != nil {
		return err
	}
	return r.fs.MkdirAll(name)
}

func (r *RootedFS) Chdir(name string) error {
	var err error
	name, err = r.Abs(name)
	if err != nil {
		return err
	}

	r.workingDir = name
	return nil
}

func (r *RootedFS) Getwd() (string, error) {
	return r.workingDir, nil
}
