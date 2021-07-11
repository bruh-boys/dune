package lib

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Multipart, `
	
declare namespace multipart {
    export function newWriter(w: io.Writer): Writer
	
    export interface Writer {
        writeField(key: string, value: string): void
        createFormFile(fieldName: string, fileName: string): io.Writer
		formDataContentType(): string
		close(): void
    }
}

	`)
}

var Multipart = []dune.NativeFunction{
	{
		Name:      "multipart.newWriter",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			r := args[0].ToObject()

			w, ok := r.(io.Writer)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected a io.Writer, got %v", args[0])
			}

			m := multipart.NewWriter(w)
			return dune.NewObject(&multipartWriter{m}), nil
		},
	},
}

func newMultipartWriter(w io.Writer) *multipart.Writer {
	return multipart.NewWriter(w)
}

type multipartWriter struct {
	w *multipart.Writer
}

func (*multipartWriter) Type() string {
	return "multipart.Writer"
}

func (w *multipartWriter) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "createFormFile":
		return w.createFormFile
	case "writeField":
		return w.writeField
	case "formDataContentType":
		return w.formDataContentType
	case "close":
		return w.close
	}
	return nil
}

func (w *multipartWriter) createFormFile(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}

	f, err := w.w.CreateFormFile(args[0].ToString(), args[1].ToString())
	return dune.NewObject(&writer{f}), err
}

func (w *multipartWriter) formDataContentType(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}
	return dune.NewString(w.w.FormDataContentType()), nil
}

func (w *multipartWriter) writeField(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}

	err := w.w.WriteField(args[0].ToString(), args[1].ToString())
	return dune.NullValue, err
}

func (w *multipartWriter) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	err := w.w.Close()
	return dune.NullValue, err
}

func (w *multipartWriter) Close() error {
	return w.w.Close()
}
