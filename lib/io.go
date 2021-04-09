package lib

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/dunelang/dune"
	"github.com/dunelang/dune/filesystem"
)

func init() {
	dune.RegisterLib(IO, `

declare namespace io {
    export interface Reader {
        read(b: byte[]): number
    }

    export interface ReaderAt {
		ReadAt(p: byte[], off: number): number
    }
	
    export interface ReaderCloser extends Reader {
        close(): void
    }

    export interface Writer {
        write(v: string | byte[]): number | void
    }

    export interface WriterCloser extends Writer {
        close(): void
    }

    export function copy(dst: Writer, src: Reader): number

    export function newMemFS(): FileSystem

    export function newRootedFS(root: string, baseFS: FileSystem): FileSystem

    export function newRestrictedFS(baseFS: FileSystem): RestrictedFS

    /** 
     * Sets the default data file system that will be returned by io.dataFS()
     */
    export function setDataFS(fs: FileSystem): void

    export function newBuffer(): Buffer

    export interface Buffer {
        length: number
        cap: number
        read(b: byte[]): number
        write(v: any): void
        string(): string
        toBytes(): byte[]
	}

    export interface FileSystem {
		getWd(): string
        abs(path: string): string
        open(path: string): File
        openIfExists(path: string): File
        openForWrite(path: string): File
        openForAppend(path: string): File
        chdir(dir: string): void
        exists(path: string): boolean
        rename(source: string, dest: string): void
        removeAll(path: string): void
        readAll(path: string): byte[]
        readAllIfExists(path: string): byte[]
        readString(path: string): string
        readStringIfExists(path: string): string
        write(path: string, data: string | io.Reader | byte[]): void
        append(path: string, data: string | byte[]): void
        mkdir(path: string): void
        stat(path: string): FileInfo
        readDir(path: string): FileInfo[]
        readNames(path: string, recursive?: boolean): string[]
	}
	
	export interface RestrictedFS extends FileSystem {
		addToWhitelist(path: string): void
		addToBlacklist(path: string): void
	}

    export interface File {
        read(b: byte[]): number
        write(v: string | byte[] | io.Reader): number
        writeAt(v: string | byte[] | io.Reader, offset: number): number
        close(): void
    }

    export interface FileInfo {
        name: string
        modTime: time.Time
        isDir: boolean
        size: number
    }
}
`)
}

var IO = []dune.NativeFunction{
	{
		Name:      "io.newBuffer",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewObject(NewBuffer()), nil
		},
	},
	{
		Name:      "io.copy",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object, dune.Object); err != nil {
				return dune.NullValue, err
			}

			dst, ok := args[0].ToObject().(io.Writer)
			if !ok {
				return dune.NullValue, ErrInvalidType
			}

			src, ok := args[1].ToObject().(io.Reader)
			if !ok {
				return dune.NullValue, ErrInvalidType
			}

			i, err := io.Copy(dst, src)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewInt64(i), nil
		},
	},
	{
		Name:      "io.newRootedFS",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.Object); err != nil {
				return dune.NullValue, err
			}
			fs, ok := args[1].ToObject().(*FileSystemObj)
			if !ok {
				return dune.NullValue, fmt.Errorf("invalid filesystem argument, got %v", args[1])
			}
			root := args[0].String()
			rFS, err := filesystem.NewRootedFS(root, fs.FS)
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewObject(NewFileSystem(rFS)), nil
		},
	},
	{
		Name:      "io.newRestrictedFS",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			fs, ok := args[0].ToObject().(*FileSystemObj)
			if !ok {
				return dune.NullValue, fmt.Errorf("invalid filesystem argument, got %v", args[1])
			}

			rFS, err := filesystem.NewRestrictedFS(fs.FS)
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewObject(NewFileSystem(rFS)), nil
		},
	},
	{
		Name: "io.newMemFS",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args); err != nil {
				return dune.NullValue, err
			}

			fs := filesystem.NewMemFS()
			return dune.NewObject(NewFileSystem(fs)), nil
		},
	},
}

type readerCloser struct {
	r io.ReadCloser
}

func (r *readerCloser) Type() string {
	return "io.Reader"
}

func (r *readerCloser) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "read":
		return r.read
	case "close":
		return r.close
	}
	return nil
}

func (r *readerCloser) Close() error {
	return r.r.Close()
}

func (r *readerCloser) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	err := r.r.Close()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (r *readerCloser) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *readerCloser) read(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Bytes); err != nil {
		return dune.NullValue, err
	}

	buf := args[0].ToBytes()

	n, err := r.r.Read(buf)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewInt(n), nil
}

func NewReader(r io.Reader) *reader {
	return &reader{r}
}

type reader struct {
	r io.Reader
}

func (r *reader) Type() string {
	return "io.Reader"
}

func (r *reader) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "read":
		return r.read
	}
	return nil
}

func (r *reader) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *reader) read(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Bytes); err != nil {
		return dune.NullValue, err
	}

	buf := args[0].ToBytes()

	n, err := r.r.Read(buf)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewInt(n), nil
}

func NewWriter(w io.Writer) *writer {
	return &writer{w}
}

type writer struct {
	w io.Writer
}

func (*writer) Type() string {
	return "io.Writer"
}

func (w *writer) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "write":
		return w.write
	}
	return nil
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *writer) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Bytes); err != nil {
		return dune.NullValue, err
	}

	buf := args[0].ToBytes()

	n, err := w.w.Write(buf)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewInt(n), nil
}

func NewBuffer() Buffer {
	var b bytes.Buffer
	return Buffer{&b}
}

type Buffer struct {
	Buf *bytes.Buffer
}

func (b Buffer) Type() string {
	return "io.Buffer"
}

func (b Buffer) Read(p []byte) (n int, err error) {
	return b.Buf.Read(p)
}

func (b Buffer) Write(p []byte) (n int, err error) {
	return b.Buf.Write(p)
}

func (b Buffer) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "length":
		return dune.NewInt(b.Buf.Len()), nil
	case "cap":
		return dune.NewInt(b.Buf.Cap()), nil
	}

	return dune.UndefinedValue, nil
}

func (b Buffer) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "write":
		return b.write
	case "string":
		return b.string
	case "toBytes":
		return b.toBytes
	}
	return nil
}

func (b Buffer) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	if err := Write(b.Buf, args[0], vm); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func WriteAt(w io.WriterAt, v dune.Value, off int64, vm *dune.VM) error {
	var d []byte

	switch v.Type {
	case dune.Null, dune.Undefined:
		return nil
	case dune.String, dune.Bytes:
		d = v.ToBytes()
	case dune.Int:
		i := v.ToInt()
		if i < 0 || i > 255 {
			return fmt.Errorf("invalid byte value %d", i)
		}
		d = []byte{byte(i)}
	case dune.Array:
		a := v.ToArray()
		d = make([]byte, len(a))
		for i, b := range a {
			switch b.Type {
			case dune.Int:
				x := b.ToInt()
				if x < 0 || x > 255 {
					return fmt.Errorf("invalid byte value %d at %d", x, i)
				}
				d[i] = byte(x)
			}
		}
	case dune.Object:
		r, ok := v.ToObject().(io.Reader)
		if !ok {
			return ErrInvalidType
		}
		var err error
		d, err = ioutil.ReadAll(r)
		if err != nil {
			return err
		}
	default:
		return ErrInvalidType
	}

	if err := vm.AddAllocations(len(d)); err != nil {
		return err
	}
	_, err := w.WriteAt(d, off)
	return err
}

func Write(w io.Writer, v dune.Value, vm *dune.VM) error {
	var d []byte

	switch v.Type {
	case dune.Null, dune.Undefined:
		return nil
	case dune.String, dune.Bytes:
		d = v.ToBytes()
	case dune.Int:
		i := v.ToInt()
		if i < 0 || i > 255 {
			return fmt.Errorf("invalid byte value %d", i)
		}
		d = []byte{byte(i)}
	case dune.Array:
		a := v.ToArray()
		d = make([]byte, len(a))
		for i, b := range a {
			switch b.Type {
			case dune.Int:
				x := b.ToInt()
				if x < 0 || x > 255 {
					return fmt.Errorf("invalid byte value %d at %d", x, i)
				}
				d[i] = byte(x)
			}
		}
	case dune.Object:
		r, ok := v.ToObject().(io.Reader)
		if !ok {
			return ErrInvalidType
		}
		// we are not worrying here about allocations.
		// TODO: Make it safe??
		if _, err := io.Copy(w, r); err != nil {
			return err
		}
	default:
		return ErrInvalidType
	}

	if err := vm.AddAllocations(len(d)); err != nil {
		return err
	}

	_, err := w.Write(d)
	return err
}

func (b Buffer) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 argument, got %d", len(args))
	}
	return dune.NewString(b.Buf.String()), nil
}

func (b Buffer) toBytes(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 argument, got %d", len(args))
	}
	return dune.NewBytes(b.Buf.Bytes()), nil
}

type FileSystemObj struct {
	FS filesystem.FS
}

func NewFileSystem(fs filesystem.FS) *FileSystemObj {
	return &FileSystemObj{fs}
}

func (f *FileSystemObj) Type() string {
	return "os.FileSystem"
}

func (f *FileSystemObj) GetMethod(name string) dune.NativeMethod {
	if f == nil {
		return nil
	}

	switch name {
	case "stat":
		return f.stat
	case "readAll":
		return f.readAll
	case "readString":
		return f.readString
	case "readAllIfExists":
		return f.readAllIfExists
	case "readStringIfExists":
		return f.readStringIfExists
	case "write":
		return f.write
	case "append":
		return f.append
	case "mkdir":
		return f.mkdir
	case "readDir":
		return f.readDir
	case "readNames":
		return f.readNames
	case "exists":
		return f.exists
	case "rename":
		return f.rename
	case "removeAll":
		return f.removeAll
	case "abs":
		return f.abs
	case "chdir":
		return f.chdir
	case "open":
		return f.open
	case "openIfExists":
		return f.openIfExists
	case "openForWrite":
		return f.openForWrite
	case "openForAppend":
		return f.openForAppend
	case "addToWhitelist":
		return f.addToWhitelist
	case "addToBlacklist":
		return f.addToBlacklist
	case "getWd":
		return f.getWd
	}
	return nil
}

func (f *FileSystemObj) getWd(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	s, err := f.FS.Getwd()
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewString(s), nil
}

func (f *FileSystemObj) addToBlacklist(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	fs, ok := f.FS.(*filesystem.RestrictedFS)
	if !ok {
		return dune.NullValue, fmt.Errorf("invalid method")
	}

	path := args[0].String()

	if err := fs.AddToBlacklist(path); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (f *FileSystemObj) addToWhitelist(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	fs, ok := f.FS.(*filesystem.RestrictedFS)
	if !ok {
		return dune.NullValue, fmt.Errorf("invalid method")
	}

	path := args[0].String()

	if err := fs.AddToWhitelist(path); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (f *FileSystemObj) rename(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}

	source := args[0].String()
	dest := args[1].String()

	if err := f.FS.Rename(source, dest); err != nil {
		if os.IsNotExist(err) {
			return dune.NullValue, fmt.Errorf("rename %v to %v: %w", source, dest, err)
		}
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (f *FileSystemObj) removeAll(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	if err := f.FS.RemoveAll(name); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (f *FileSystemObj) openIfExists(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	fi, err := f.FS.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return dune.NullValue, nil
		}
		return dune.NullValue, err
	}

	return dune.NewObject(newFile(fi, vm)), nil
}

func (f *FileSystemObj) open(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	fi, err := f.FS.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return dune.NullValue, fmt.Errorf("error opening '%v': %w", name, err)
		}
		return dune.NullValue, err
	}

	return dune.NewObject(newFile(fi, vm)), nil
}

func (f *FileSystemObj) openForWrite(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	fi, err := f.FS.OpenForWrite(name)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(newFile(fi, vm)), nil
}

func (f *FileSystemObj) openForAppend(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	fi, err := f.FS.OpenForAppend(name)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(newFile(fi, vm)), nil
}

func newFile(fi filesystem.File, vm *dune.VM) *file {
	f := &file{f: fi}
	vm.SetGlobalFinalizer(f)
	return f
}

type file struct {
	io.ReaderAt
	f      filesystem.File
	closed bool
}

func (f *file) Type() string {
	return "os.File"
}

func (f *file) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true
	return f.f.Close()
}

func (f *file) Write(p []byte) (n int, err error) {
	return f.f.Write(p)
}

func (f *file) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

func (f *file) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	return f.f.Seek(offset, whence)
}

func (f *file) Stat() (os.FileInfo, error) {
	return f.f.Stat()
}

func (f *file) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "write":
		return f.write
	case "writeAt":
		return f.writeAt
	case "read":
		return f.read
	case "close":
		return f.close
	}
	return nil
}

func (f *file) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 parameter")
	}

	a := args[0]

	if err := Write(f.f, a, vm); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (f *file) writeAt(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 2 {
		return dune.NullValue, fmt.Errorf("expected 2 parameters")
	}

	a := args[0]

	offV := args[1]
	if offV.Type != dune.Int {
		return dune.NullValue, fmt.Errorf("expected parameter 2 to be int")
	}

	if err := WriteAt(f.f, a, offV.ToInt(), vm); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (f *file) read(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Bytes); err != nil {
		return dune.NullValue, err
	}

	buf := args[0].ToBytes()

	n, err := f.f.Read(buf)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewInt(n), nil
}

func (f *file) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("no parameters expected")
	}
	if !f.closed {
		f.closed = true
		f.f.Close()
	}
	return dune.NullValue, nil
}

func (f *FileSystemObj) abs(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	path := args[0].String()
	abs, err := f.FS.Abs(path)
	if err != nil {
		if os.IsNotExist(err) {
			return dune.NullValue, fmt.Errorf("error opening '%v': %w", path, err)
		}
		return dune.NullValue, err
	}

	return dune.NewString(abs), nil
}

func (f *FileSystemObj) chdir(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	dir := args[0].String()
	if err := f.FS.Chdir(dir); err != nil {
		if os.IsNotExist(err) {
			return dune.NullValue, fmt.Errorf("error opening '%v': %w", dir, err)
		}
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (f *FileSystemObj) exists(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	if _, err := f.FS.Stat(name); err != nil {
		return dune.FalseValue, nil
	}
	return dune.TrueValue, nil
}

func (f *FileSystemObj) readNames(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	var name string
	var recursive bool
	l := len(args)

	if l == 0 {
		name = "."
	} else {
		name = args[0].String()
	}

	if l > 2 {
		return dune.NullValue, fmt.Errorf("expected 2 arguments max, got %d", len(args))
	}

	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].TypeName())
	}

	if l == 2 {
		switch args[1].Type {
		case dune.Bool, dune.Undefined, dune.Null:
			recursive = args[1].ToBool()
		default:
			return dune.NullValue, fmt.Errorf("expected argument 2 to be a boolean, got %v", args[1].TypeName())
		}
	}

	fis, err := ReadNames(f.FS, name, recursive)
	if err != nil {
		return dune.NullValue, err
	}

	result := make([]dune.Value, len(fis))

	for i, fi := range fis {
		result[i] = dune.NewString(fi)
	}

	return dune.NewArrayValues(result), nil
}

// ReadNames reads the directory and file names contained in dirname.
func ReadNames(fs filesystem.FS, dirname string, recursive bool) ([]string, error) {
	n, err := readNames(fs, dirname, true, recursive)
	if err != nil {
		return nil, err
	}
	sort.Strings(n)
	return n, nil
}

func readNames(fs filesystem.FS, dirname string, removeTopDir, recursive bool) ([]string, error) {
	f, err := fs.Open(dirname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("error opening '%v': %w", dirname, err)
		}
		return nil, err
	}

	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	var names []string

	for _, l := range list {
		name := filepath.Join(dirname, l.Name())
		names = append(names, name)

		if recursive && l.IsDir() {
			sub, err := readNames(fs, name, false, true)
			if err != nil {
				return nil, err
			}

			//			if removeTopDir {
			//				for i, v := range sub {
			//					j := strings.IndexRune(v, os.PathSeparator) + 1
			//					sub[i] = v[j:]
			//				}
			//			}

			names = append(names, sub...)
		}
	}

	return names, nil
}

func (f *FileSystemObj) stat(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	fi, err := f.FS.Stat(name)
	if err != nil {
		// ignore errors. Just return null if is invalid
		return dune.NullValue, nil
	}

	return dune.NewObject(fileInfo{fi}), nil
}

func (f *FileSystemObj) readDir(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	file, err := f.FS.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return dune.NullValue, fmt.Errorf("error opening '%v': %w", name, err)
		}
		return dune.NullValue, err
	}

	list, err := file.Readdir(-1)
	file.Close()
	if err != nil {
		return dune.NullValue, err
	}

	result := make([]dune.Value, len(list))

	for i, fi := range list {
		result[i] = dune.NewObject(fileInfo{fi})
	}

	return dune.NewArrayValues(result), nil
}

func (f *FileSystemObj) mkdir(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	name := args[0].String()

	if err := f.FS.MkdirAll(name); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (f *FileSystemObj) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 2 {
		return dune.NullValue, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument 1 to be a string, got %d", args[0].Type)
	}
	name := args[0].String()

	file, err := f.FS.OpenForWrite(name)
	if err != nil {
		return dune.NullValue, err
	}

	if err := Write(file, args[1], vm); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (f *FileSystemObj) append(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 2 {
		return dune.NullValue, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument 1 to be a string, got %s", args[0].TypeName())
	}
	name := args[0].String()

	b := args[1]
	switch b.Type {
	case dune.Bytes, dune.String:
	default:
		return dune.NullValue, fmt.Errorf("expected argument 2 to be a string or byte array, got %s", args[1].TypeName())
	}

	f.FS.AppendPath(name, b.ToBytes())
	return dune.NullValue, nil
}

func (f *FileSystemObj) readAll(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	b, err := f.read(false, args, vm)
	if err != nil {
		return dune.NullValue, err
	}
	if b == nil {
		return dune.NullValue, nil
	}
	return dune.NewBytes(b), nil
}

func (f *FileSystemObj) readString(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	b, err := f.read(false, args, vm)
	if err != nil {
		return dune.NullValue, err
	}
	if b == nil {
		return dune.NullValue, nil
	}
	return dune.NewString(string(b)), nil
}

func (f *FileSystemObj) readAllIfExists(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	b, err := f.read(true, args, vm)
	if err != nil {
		return dune.NullValue, err
	}
	if b == nil {
		return dune.NullValue, nil
	}
	return dune.NewBytes(b), nil
}

func (f *FileSystemObj) readStringIfExists(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	b, err := f.read(true, args, vm)
	if err != nil {
		return dune.NullValue, err
	}
	if b == nil {
		return dune.NullValue, nil
	}
	return dune.NewString(string(b)), nil
}

func (f *FileSystemObj) read(ifExists bool, args []dune.Value, vm *dune.VM) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	if args[0].Type != dune.String {
		return nil, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}
	name := args[0].String()

	file, err := f.FS.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			if ifExists {
				return nil, nil
			}
			return nil, fmt.Errorf("error opening '%v': %w", name, err)
		}
		return nil, err
	}
	b, err := ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		return nil, err
	}
	return b, nil
}

type fileInfo struct {
	fi os.FileInfo
}

func (f fileInfo) Type() string {
	return "os.FileInfo"
}

func (f fileInfo) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "name":
		return dune.NewString(f.fi.Name()), nil
	case "modTime":
		return dune.NewObject(TimeObj(f.fi.ModTime())), nil
	case "isDir":
		return dune.NewBool(f.fi.IsDir()), nil
	case "size":
		return dune.NewInt64(f.fi.Size()), nil
	}
	return dune.UndefinedValue, nil
}
