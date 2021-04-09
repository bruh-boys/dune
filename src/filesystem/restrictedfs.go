package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RestrictedFS implements fileSystem that serves a directory as root
type RestrictedFS struct {
	fs         FS
	whitelist  []string
	blacklist  []string
	workingDir string
}

func NewRestrictedFS(fs FS) (*RestrictedFS, error) {
	return &RestrictedFS{fs: fs, workingDir: "/"}, nil
}

func (r *RestrictedFS) AddToWhitelist(name string) error {
	abs, err := r.fs.Abs(name)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	for _, v := range r.whitelist {
		if v == abs {
			return nil
		}
	}

	r.whitelist = append(r.whitelist, abs)
	return nil
}

func (r *RestrictedFS) AddToBlacklist(name string) error {
	abs, err := r.fs.Abs(name)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	for _, v := range r.blacklist {
		if v == abs {
			return nil
		}
	}

	r.blacklist = append(r.blacklist, abs)
	return nil
}

// Sets the home directory
func (r *RestrictedFS) SetHome(name string) error {
	return nil
}

func (r *RestrictedFS) Abs(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty path")
	}

	return r.fs.Abs(name)
}

func (r *RestrictedFS) checkPath(path string) error {
	abs, err := r.fs.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %v", err)
	}

	for _, v := range r.blacklist {
		if strings.HasPrefix(abs, v) {
			return fmt.Errorf("invalid path: not allowed: %s", abs)
		}
	}

	for _, v := range r.whitelist {
		if strings.HasPrefix(abs, v) {
			return nil
		}
	}

	return fmt.Errorf("invalid path: not allowed: %s", abs)
}

func (r *RestrictedFS) Rename(oldPath, newPath string) error {
	if err := r.checkPath(oldPath); err != nil {
		return err
	}
	if err := r.checkPath(newPath); err != nil {
		return err
	}
	return r.fs.Rename(oldPath, newPath)
}

func (r *RestrictedFS) RemoveAll(name string) error {
	if err := r.checkPath(name); err != nil {
		return err
	}

	return r.fs.RemoveAll(name)
}

func (r *RestrictedFS) OpenForAppend(name string) (File, error) {
	if err := r.checkPath(name); err != nil {
		return nil, err
	}

	return r.fs.OpenForAppend(name)
}

func (r *RestrictedFS) OpenForWrite(name string) (File, error) {
	if err := r.checkPath(name); err != nil {
		return nil, err
	}

	return r.fs.OpenForWrite(name)
}

func (r *RestrictedFS) Open(name string) (File, error) {
	if err := r.checkPath(name); err != nil {
		return nil, err
	}

	return r.fs.Open(name)
}

func (r *RestrictedFS) OpenIfExists(name string) (File, error) {
	if err := r.checkPath(name); err != nil {
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

func (r *RestrictedFS) Stat(name string) (os.FileInfo, error) {
	if err := r.checkPath(name); err != nil {
		return nil, err
	}

	return r.fs.Stat(name)
}

func (r *RestrictedFS) Write(name string, data []byte) error {
	if err := r.checkPath(name); err != nil {
		return err
	}

	return r.fs.Write(name, data)
}

func (r *RestrictedFS) Append(name string, data []byte) error {
	if err := r.checkPath(name); err != nil {
		return err
	}

	return r.fs.Append(name, data)
}

func (r *RestrictedFS) AppendPath(name string, data []byte) error {
	if err := r.checkPath(name); err != nil {
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

func (r *RestrictedFS) Mkdir(name string) error {
	if err := r.checkPath(name); err != nil {
		return err
	}

	return r.fs.Mkdir(name)
}

func (r *RestrictedFS) MkdirAll(name string) error {
	if err := r.checkPath(name); err != nil {
		return err
	}
	return r.fs.MkdirAll(name)
}

func (r *RestrictedFS) Chdir(name string) error {
	if err := r.checkPath(name); err != nil {
		return err
	}

	r.workingDir = name
	return nil
}

func (r *RestrictedFS) Getwd() (string, error) {
	return r.workingDir, nil
}
