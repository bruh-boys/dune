package lib

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(CSV, `

declare namespace csv {
    export function newReader(r: io.Reader): Reader
    export interface Reader {
		comma: string
		lazyQuotes: boolean
        read(): string[]
    }

    export function newWriter(r: io.Writer): Writer
    export interface Writer {
        comma: string
        write(v: (string | number)[]): void
        flush(): void
    }
}
`)
}

var CSV = []dune.NativeFunction{
	{
		Name:      "csv.newReader",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			r, ok := args[0].ToObject().(io.Reader)
			if !ok {
				return dune.NullValue, ErrInvalidType
			}

			reader := csv.NewReader(r)

			return dune.NewObject(&csvReader{reader}), nil
		},
	},
	{
		Name:      "csv.newWriter",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			w, ok := args[0].ToObject().(io.Writer)
			if !ok {
				return dune.NullValue, ErrInvalidType
			}

			writer := csv.NewWriter(w)

			return dune.NewObject(&csvWriter{writer}), nil
		},
	},
}

type csvReader struct {
	r *csv.Reader
}

func (r *csvReader) Type() string {
	return "csv.Reader"
}

func (r *csvReader) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "comma":
		return dune.NewString(string(r.r.Comma)), nil
	case "lazyQuotes":
		return dune.NewBool(r.r.LazyQuotes), nil
	}
	return dune.UndefinedValue, nil
}

func (r *csvReader) SetProperty(key string, v dune.Value, vm *dune.VM) error {
	switch key {
	case "comma":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		s := v.ToString()
		if len(s) != 1 {
			return fmt.Errorf("invalid comma: %s", s)
		}
		r.r.Comma = rune(s[0])
		return nil
	case "lazyQuotes":
		if v.Type != dune.Bool {
			return ErrInvalidType
		}
		r.r.LazyQuotes = v.ToBool()
		return nil
	}
	return ErrReadOnlyOrUndefined
}

func (r *csvReader) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "read":
		return r.read
	}
	return nil
}

func (r *csvReader) read(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	records, err := r.r.Read()
	if err != nil {
		if err == io.EOF {
			return dune.NullValue, nil
		}
		return dune.NullValue, err
	}

	result := make([]dune.Value, len(records))
	for i, v := range records {
		result[i] = dune.NewString(v)
	}

	return dune.NewArrayValues(result), nil
}

type csvWriter struct {
	w *csv.Writer
}

func (*csvWriter) Type() string {
	return "csv.Writer"
}

func (w *csvWriter) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "comma":
		return dune.NewString(string(w.w.Comma)), nil
	}
	return dune.UndefinedValue, nil
}

func (w *csvWriter) SetProperty(key string, v dune.Value, vm *dune.VM) error {
	switch key {
	case "comma":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		s := v.ToString()
		if len(s) != 1 {
			return fmt.Errorf("invalid comma: %s", s)
		}
		w.w.Comma = rune(s[0])
		return nil
	}
	return ErrReadOnlyOrUndefined
}

func (w *csvWriter) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "write":
		return w.write
	case "flush":
		return w.flush
	}
	return nil
}

func (w *csvWriter) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Array); err != nil {
		return dune.NullValue, err
	}

	a := args[0].ToArray()

	values := make([]string, len(a))

	for i, v := range a {
		values[i] = v.ToString()
	}

	if err := w.w.Write(values); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (w *csvWriter) flush(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	w.w.Flush()

	if err := w.w.Error(); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}
