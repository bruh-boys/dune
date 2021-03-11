package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var OS = &osFS{}

// osFS implements FS using the local disk.
type osFS struct{}

// Abs returns the absolute path
func (osFS) Abs(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty path")
	}

	abs, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func (osFS) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (osFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (osFS) Open(name string) (File, error) {
	return os.Open(name)
}

func (osFS) OpenIfExists(name string) (File, error) {
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

func (osFS) OpenForWrite(name string) (File, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
}

func (osFS) OpenForAppend(name string) (File, error) {
	return os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
}

func (osFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (osFS) Write(name string, data []byte) error {
	return ioutil.WriteFile(name, data, 0644)
}

func (osFS) Append(name string, data []byte) error {
	f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	_, err = f.Write(data)

	err2 := f.Close()

	if err != nil {
		return err
	}

	if err2 != nil {
		return err2
	}

	return nil
}

func (fs *osFS) AppendPath(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	return fs.Append(path, data)
}

func (osFS) Mkdir(name string) error {
	return os.Mkdir(name, os.ModePerm)
}

func (osFS) MkdirAll(name string) error {
	return os.MkdirAll(name, os.ModePerm)
}

func (osFS) Chdir(dir string) error {
	return os.Chdir(dir)
}

func (osFS) Getwd() (string, error) {
	return os.Getwd()
}

// Sets the home directory
func (osFS) SetHome(name string) error {
	return fmt.Errorf("home is read only")
}
