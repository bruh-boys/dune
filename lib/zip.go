package lib

import (
	"archive/zip"
	"fmt"
	"io"

	"github.com/dunelang/dune"
	"github.com/dunelang/dune/filesystem"
)

func init() {
	dune.RegisterLib(ZIP, `


declare namespace zip {
    export function newWriter(w: io.Writer): Writer
    export function newReader(r: io.Reader, size: number): io.ReaderCloser
    export function open(path: string, fs?: io.FileSystem): Reader

    export interface Writer {
        create(name: string): io.Writer
        flush(): void
        close(): void
    }

    export interface Reader {
        files(): File[]
        close(): void
    }

    export interface File {
        name: string
        compressedSize: number
        uncompressedSize: number
        open(): io.ReaderCloser
    }
}


`)
}

var ZIP = []dune.NativeFunction{
	{
		Name:      "zip.newWriter",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			w, ok := args[0].ToObjectOrNil().(io.Writer)
			if !ok {
				return dune.NullValue, fmt.Errorf("exepected a Writer, got %s", args[0].TypeName())
			}

			g := zip.NewWriter(w)
			v := &zipWriter{g}
			return dune.NewObject(v), nil
		},
	},
	{
		Name:      "zip.newReader",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object, dune.Int); err != nil {
				return dune.NullValue, err
			}

			r, ok := args[0].ToObjectOrNil().(io.ReaderAt)
			if !ok {
				return dune.NullValue, fmt.Errorf("exepected a reader, got %s", args[0].TypeName())
			}

			size := args[1].ToInt()

			gr, err := zip.NewReader(r, size)
			if err != nil {
				return dune.NullValue, err
			}

			v := &zipReader{r: gr}
			return dune.NewObject(v), nil
		},
	},
	{
		Name:      "zip.open",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String, dune.Object); err != nil {
				return dune.NullValue, err
			}

			if err := ValidateArgRange(args, 1, 2); err != nil {
				return dune.NullValue, err
			}

			var fs filesystem.FS
			if len(args) == 2 {
				fsObj, ok := args[1].ToObjectOrNil().(*FileSystemObj)
				if !ok {
					return dune.NullValue, fmt.Errorf("exepected a FileSystem, got %s", args[1].TypeName())
				}
				fs = fsObj.FS
			} else {
				fs = vm.FileSystem
			}

			f, err := fs.Open(args[0].String())
			if err != nil {
				return dune.NullValue, err
			}

			fi, err := f.Stat()
			if err != nil {
				return dune.NullValue, err
			}

			size := fi.Size()

			gr, err := zip.NewReader(f, size)
			if err != nil {
				return dune.NullValue, err
			}

			v := &zipReader{gr, f}
			return dune.NewObject(v), nil
		},
	},
}

type zipWriter struct {
	w *zip.Writer
}

func (*zipWriter) Type() string {
	return "zip.Writer"
}

func (w *zipWriter) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "create":
		return w.create
	case "flush":
		return w.flush
	case "close":
		return w.close
	}
	return nil
}

func (w *zipWriter) flush(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	if err := w.w.Flush(); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (w *zipWriter) create(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	f, err := w.w.Create(name)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(&writer{f}), nil
}

func (w *zipWriter) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	err := w.w.Close()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

type zipReader struct {
	r *zip.Reader
	c io.Closer
}

func (*zipReader) Type() string {
	return "zip.Reader"
}
func (r *zipReader) Close() error {
	if r.c == nil {
		return nil
	}
	return r.c.Close()
}

func (r *zipReader) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "files":
		return r.files
	case "close":
		return r.close
	}
	return nil
}

func (r *zipReader) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	if err := r.Close(); err != nil {
		return dune.NullValue, err
	}
	return dune.NullValue, nil
}

func (r *zipReader) files(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	files := r.r.File
	a := make([]dune.Value, len(files))

	for i, f := range files {
		a[i] = dune.NewObject(&zipFile{f})
	}
	return dune.NewArrayValues(a), nil
}

type zipFile struct {
	f *zip.File
}

func (*zipFile) Type() string {
	return "zip.File"
}

func (f *zipFile) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "name":
		return dune.NewString(f.f.Name), nil
	case "compressedSize":
		return dune.NewInt64(int64(f.f.CompressedSize64)), nil
	case "uncompressedSize":
		return dune.NewInt64(int64(f.f.UncompressedSize64)), nil
	}

	return dune.UndefinedValue, nil
}

func (f *zipFile) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "open":
		return f.open
	}
	return nil
}

func (f *zipFile) open(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	rc, err := f.f.Open()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(&readerCloser{rc}), nil
}
