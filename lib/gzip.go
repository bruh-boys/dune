package lib

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(GZIP, `

declare namespace gzip {
    export function newWriter(w: io.Writer): io.WriterCloser
    export function newReader(r: io.Reader): io.ReaderCloser
}


`)
}

var GZIP = []dune.NativeFunction{
	{
		Name:      "gzip.newWriter",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			w, ok := args[0].ToObjectOrNil().(io.Writer)
			if !ok {
				return dune.NullValue, fmt.Errorf("exepected a Writer, got %s", args[0].TypeName())
			}

			g := gzip.NewWriter(w)
			v := &gzipWriter{g}
			return dune.NewObject(v), nil
		},
	},
	{
		Name:      "gzip.newReader",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			r, ok := args[0].ToObjectOrNil().(io.Reader)
			if !ok {
				return dune.NullValue, fmt.Errorf("exepected a reader, got %s", args[0].TypeName())
			}

			gr, err := gzip.NewReader(r)
			if err != nil {
				return dune.NullValue, err
			}
			v := &gzipReader{gr}
			return dune.NewObject(v), nil
		},
	},
}

type gzipWriter struct {
	w *gzip.Writer
}

func (*gzipWriter) Type() string {
	return "gzip.Writer"
}

func (w *gzipWriter) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "write":
		return w.write
	case "close":
		return w.close
	}
	return nil
}

func (w *gzipWriter) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *gzipWriter) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
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

func (w *gzipWriter) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	err := w.w.Close()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

type gzipReader struct {
	r *gzip.Reader
}

func (*gzipReader) Type() string {
	return "gzip.Reader"
}

func (r *gzipReader) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "read":
		return r.read
	case "close":
		return r.close
	}
	return nil
}

func (r *gzipReader) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *gzipReader) read(args []dune.Value, vm *dune.VM) (dune.Value, error) {
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

func (r *gzipReader) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	err := r.r.Close()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}
