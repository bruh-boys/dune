package filesystem

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func NewVirtualFS() *VirtualFS {
	info := &vFileInfo{
		name:    "/",
		isDir:   true,
		modTime: time.Now(),
	}

	f := &vFile{info: info, files: make(map[string]*vFile)}
	return &VirtualFS{file: f, workdir: "/"}
}

// A tree of virtual files. A virtual file can be a directory or a individual file.
type VirtualFS struct {
	file    *vFile
	workdir string
}

func (v *VirtualFS) Abs(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty path")
	}

	if name == "." {
		return v.workdir, nil
	}

	name = trimLastSlash(name)
	if name[0] == '/' {
		return name, nil
	}

	name = filepath.Join(v.workdir, name)
	return name, nil
}

func (v *VirtualFS) SetHome(name string) error {
	return fmt.Errorf("not implemented")
}

func (v *VirtualFS) PrintPaths() {
	v.print(v.file.files, 1)
}

func (v *VirtualFS) print(files map[string]*vFile, indent int) {
	for k, f := range files {
		fmt.Print(strings.Repeat(" ", indent))
		fmt.Print(k)
		fmt.Print("\n")
		v.print(f.files, indent+1)
	}
}

// Copy copies recursively path src into dst.
func (v *VirtualFS) CopyAt(dst, src string, fs FS) error {
	fi, err := fs.Stat(src)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		b, err := ReadAll(fs, src)
		if err != nil {
			return err
		}

		v.Write(dst, b)
		return nil
	}

	err = v.MkdirAll(dst)
	if err != nil {
		return err
	}

	paths, err := ReadDir(fs, src)
	if err != nil {
		return err
	}

	for _, f := range paths {
		fDst := filepath.Join(dst, f.Name())
		fSrc := filepath.Join(src, f.Name())
		if err := v.CopyAt(fDst, fSrc, fs); err != nil {
			return err
		}
	}

	return nil
}

func (v *VirtualFS) Chdir(name string) error {
	name, err := v.Abs(name)
	if err != nil {
		return err
	}

	s, err := v.Stat(name)
	if err != nil {
		return err
	}

	if !s.IsDir() {
		return os.ErrInvalid
	}

	v.workdir = name
	return nil
}

func (v *VirtualFS) Getwd() (string, error) {
	// check that  still exists
	s, err := v.Stat(v.workdir)
	if err != nil {
		return "", err
	}

	// also that it has not been replaced by a file
	if !s.IsDir() {
		return "", os.ErrInvalid
	}

	return v.workdir, nil
}

func (v *VirtualFS) OpenForWrite(name string) (File, error) {
	f, err := v.open(name, true)
	if err != nil {
		return nil, err
	}

	vf := f.(*vFile)
	vf.data = nil // overwrite
	vf.forAppend = false
	return f, nil
}

func (v *VirtualFS) OpenForAppend(name string) (File, error) {
	f, err := v.open(name, true)
	if err != nil {
		return nil, err
	}
	f.(*vFile).forAppend = true
	return f, nil
}

func (v *VirtualFS) Open(name string) (File, error) {
	return v.open(name, false)
}

func (v *VirtualFS) OpenIfExists(name string) (File, error) {
	f, err := v.open(name, false)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

func (v *VirtualFS) open(name string, createIfNotExists bool) (File, error) {
	name, err := v.Abs(name)
	if err != nil {
		return nil, err
	}

	items := split(name, "/")
	if len(items) == 0 {
		return v.file, nil
	}

	file := v.file

	for i, l := 0, len(items); i < l; i++ {
		item := items[i]
		var ok bool
		nextFile, ok := file.files[item]
		if !ok {
			if createIfNotExists && i == l-1 {
				info := &vFileInfo{
					name:    item,
					modTime: time.Now(),
				}
				f := &vFile{info: info, files: make(map[string]*vFile)}
				file.files[item] = f
				return f, nil
			}
			return nil, os.ErrNotExist
		}
		file = nextFile
	}

	file.forAppend = false
	return file, nil
}
func (v *VirtualFS) Write(name string, data []byte) error {
	name, err := v.Abs(name)
	if err != nil {
		return err
	}

	// open de directory
	dir := filepath.Dir(name)
	f, err := v.Open(dir)
	if err != nil {
		return err
	}

	vf := f.(*vFile)

	if !vf.info.isDir {
		return os.ErrInvalid
	}

	name = base(name)

	info := &vFileInfo{
		name:    name,
		size:    len(data),
		modTime: time.Now(),
	}

	vf.files[name] = &vFile{data: data, info: info, files: make(map[string]*vFile)}
	return nil
}

func (v *VirtualFS) Append(name string, data []byte) error {
	name, err := v.Abs(name)
	if err != nil {
		return err
	}

	// open de directory
	dir := filepath.Dir(name)
	f, err := v.Open(dir)
	if err != nil {
		return err
	}

	vf := f.(*vFile)

	if !vf.info.isDir {
		return os.ErrInvalid
	}

	name = base(name)

	data = append(vf.data, data...)

	info := &vFileInfo{
		name:    name,
		size:    len(data),
		modTime: time.Now(),
	}

	vf.files[name] = &vFile{data: data, info: info, files: make(map[string]*vFile)}
	return nil
}

func (v *VirtualFS) AppendPath(path string, data []byte) error {
	path, err := v.Abs(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := v.MkdirAll(dir); err != nil {
			return err
		}
	}

	return v.Append(path, data)
}

func (v *VirtualFS) getDir(name string) (File, error) {
	// if we are opening a somthing in the root dir just return root
	dir := filepath.Dir(name)
	if dir == name {
		return v.file, nil
	}

	return v.Open(dir)
}

func (v *VirtualFS) Mkdir(name string) error {
	if name == "/" {
		return os.ErrExist
	}

	name, err := v.Abs(name)
	if err != nil {
		return err
	}

	// open de directory
	f, err := v.getDir(name)
	if err != nil {
		return err
	}
	vf := f.(*vFile)

	if !vf.info.isDir {
		return os.ErrInvalid
	}

	name = base(name)

	// check that it doesn't exist
	if _, ok := vf.files[name]; ok {
		return os.ErrExist
	}

	info := &vFileInfo{
		name:    name,
		modTime: time.Now(),
		isDir:   true,
	}

	vf.files[name] = &vFile{info: info, files: make(map[string]*vFile)}
	return nil
}

func (v *VirtualFS) MkdirAll(name string) error {
	name, err := v.Abs(name)
	if err != nil {
		return err
	}

	items := split(name, "/")
	if len(items) == 0 {
		return nil
	}

	for i := range items {
		dir := strings.Join(items[:i+1], "/")
		if name[0] == '/' {
			dir = "/" + dir
		}

		err := v.Mkdir(dir)
		if err != nil && err != os.ErrExist {
			return err
		}
	}

	return nil
}

func base(s string) string {
	s = trimFirstSlash(s)
	return filepath.Base(s)
}

func trimFirstSlash(s string) string {
	if len(s) > 0 && s[0] == '/' {
		return s[1:]
	}
	return s
}

func trimLastSlash(s string) string {
	l := len(s) - 1
	if l > 0 {
		if s[l] == '/' {
			return s[:l]
		}
	}
	return s
}

func (v *VirtualFS) Stat(name string) (os.FileInfo, error) {
	name, err := v.Abs(name)
	if err != nil {
		return nil, err
	}

	f, err := v.Open(name)
	if err != nil {
		return nil, err
	}
	return f.(*vFile).info, nil
}

func (v *VirtualFS) Rename(oldPath, newPath string) error {
	var err error
	oldPath, err = v.Abs(oldPath)
	if err != nil {
		return err
	}
	newPath, err = v.Abs(newPath)
	if err != nil {
		return err
	}

	if err := v.CopyAt(newPath, oldPath, v); err != nil {
		return err
	}

	return v.RemoveAll(oldPath)
}

func (v *VirtualFS) RemoveAll(name string) error {
	name, err := v.Abs(name)
	if err != nil {
		return err
	}

	dir := filepath.Dir(name)
	if dir == "" {
		dir = "."
	}

	f, err := v.Open(dir)
	if err != nil {
		return err
	}

	delete(f.(*vFile).files, filepath.Base(name))
	return nil
}

type vFile struct {
	data      []byte
	info      *vFileInfo
	files     map[string]*vFile
	reader    *bytes.Reader
	forAppend bool
}

func (f *vFile) Write(p []byte) (n int, err error) {
	f.data = append(f.data, p...)
	return len(p), nil
}

func (f *vFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("not implemented")
}

func (f *vFile) Close() error {
	f.reader = nil
	return nil
}

func (f *vFile) Read(b []byte) (int, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader(f.data)
	}
	return f.reader.Read(b)
}

func (f *vFile) ReadAt(b []byte, i int64) (int, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader(f.data)
	}
	return f.reader.ReadAt(b, i)
}

func (f *vFile) Seek(offset int64, whence int) (int64, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader(f.data)
	}
	return f.reader.Seek(offset, whence)
}

func (f *vFile) Stat() (os.FileInfo, error) {
	f.info.size = len(f.data)
	return f.info, nil
}

func (f *vFile) Readdir(n int) ([]os.FileInfo, error) {
	if n == -1 {
		n = len(f.files)
	}

	files := make([]os.FileInfo, n)

	i := 0
	for _, file := range f.files {
		files[i] = file.info
		i++

		if i >= n {
			break
		}
	}

	return files, nil
}

type vFileInfo struct {
	name    string
	size    int
	modTime time.Time
	isDir   bool
}

func (v *vFileInfo) Name() string {
	return v.name
}

func (v *vFileInfo) Size() int64 {
	return int64(v.size)
}

func (v *vFileInfo) Mode() os.FileMode {
	if v.isDir {
		return os.ModeDir
	}
	return os.ModePerm
}

func (v *vFileInfo) ModTime() time.Time {
	return v.modTime
}

func (v *vFileInfo) IsDir() bool {
	return v.isDir
}

func (v *vFileInfo) Sys() interface{} {
	return nil
}

func split(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
