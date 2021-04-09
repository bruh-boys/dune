package lib

import (
	"bufio"
	"fmt"
	"io"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Bufio, `
	
declare namespace bufio {
    export function newWriter(w: io.Writer): Writer
    export function newScanner(r: io.Reader): Scanner
    export function newReader(r: io.Reader): Reader

    export interface Writer {
        write(data: byte[]): number
        writeString(s: string): number
        writeByte(b: byte): void
        writeRune(s: string): number
        flush(): void
    }

    export interface Scanner {
        scan(): boolean 
        text(): string
    }

    export interface Reader {
        readString(delim: byte): string
        readBytes(delim: byte): byte[]
        readByte(): byte
        readRune(): number
    }
}

`)
}

var Bufio = []dune.NativeFunction{
	{
		Name:      "bufio.newScanner",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			r := args[0].ToObject()

			reader, ok := r.(io.Reader)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected a io.Reader, got %v", args[0])
			}

			s := bufio.NewScanner(reader)

			return dune.NewObject(&scanner{s}), nil
		},
	},
	{
		Name:      "bufio.newReader",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			r := args[0].ToObject()

			reader, ok := r.(io.Reader)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected a io.Reader, got %v", args[0])
			}

			s := bufio.NewReader(reader)

			return dune.NewObject(&bufioReader{s}), nil
		},
	},
	{
		Name:      "bufio.newWriter",
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

			s := bufio.NewWriter(w)

			return dune.NewObject(&bufioWriter{s}), nil
		},
	},
}

type bufioWriter struct {
	w *bufio.Writer
}

func (*bufioWriter) Type() string {
	return "bufio.Writer"
}

func (w *bufioWriter) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "write":
		return w.write
	case "writeString":
		return w.writeString
	case "writeByte":
		return w.writeByte
	case "writeRune":
		return w.writeRune
	case "flush":
		return w.flush
	}
	return nil
}

func (w *bufioWriter) writeString(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	n, err := w.w.WriteString(args[0].String())
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewInt(n), nil
}

func (w *bufioWriter) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Bytes); err != nil {
		return dune.NullValue, err
	}

	n, err := w.w.Write(args[0].ToBytes())
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewInt(n), nil
}

func (w *bufioWriter) flush(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}
	err := w.w.Flush()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (w *bufioWriter) writeByte(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Int); err != nil {
		return dune.NullValue, err
	}

	err := w.w.WriteByte(byte(args[0].ToInt()))
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (w *bufioWriter) writeRune(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 1, 1); err != nil {
		return dune.NullValue, err
	}

	var r rune
	switch args[0].Type {
	case dune.Rune, dune.Int:
		r = args[0].ToRune()
	default:
		return dune.NullValue, fmt.Errorf("expected rune, got %v", args[0].TypeName())
	}

	n, err := w.w.WriteRune(r)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewInt(n), nil
}

type bufioReader struct {
	s *bufio.Reader
}

func (*bufioReader) Type() string {
	return "bufio.Reader"
}

func (s *bufioReader) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "readString":
		return s.readString
	case "readRune":
		return s.readRune
	case "readByte":
		return s.readByte
	case "readBytes":
		return s.readBytes
	}
	return nil
}

func (s *bufioReader) readRune(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	r, _, err := s.s.ReadRune()
	if err != nil {
		if err == io.EOF {
			return dune.NullValue, nil
		}
		return dune.NullValue, err
	}

	return dune.NewRune(r), nil
}

func (s *bufioReader) readByte(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	b, err := s.s.ReadByte()
	if err != nil {
		if err == io.EOF {
			return dune.NullValue, nil
		}
		return dune.NullValue, err
	}

	return dune.NewInt(int(b)), nil
}

func (s *bufioReader) readBytes(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	delim := args[0].String()
	if len(delim) != 1 {
		return dune.NullValue, fmt.Errorf("invalid delimiter lenght. Must be a byte: %v", delim)
	}

	v, err := s.s.ReadBytes(delim[0])

	if err != nil {
		if err == io.EOF {
			return dune.NullValue, nil
		}
		return dune.NullValue, err
	}

	return dune.NewBytes(v), nil
}

func (s *bufioReader) readString(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	delim := args[0].String()
	if len(delim) != 1 {
		return dune.NullValue, fmt.Errorf("invalid delimiter lenght. Must be a byte: %v", delim)
	}

	v, err := s.s.ReadString(delim[0])

	if err != nil {
		if err == io.EOF {
			return dune.NullValue, nil
		}
		return dune.NullValue, err
	}

	return dune.NewString(v), nil
}

type scanner struct {
	s *bufio.Scanner
}

func (s *scanner) Type() string {
	return "bufio.Scanner"
}

func (s *scanner) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "scan":
		return s.scan
	case "text":
		return s.text
	}
	return nil
}

func (s *scanner) text(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	v := s.s.Text()

	if err := s.s.Err(); err != nil {
		return dune.NullValue, err
	}

	return dune.NewString(v), nil
}

func (s *scanner) scan(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	v := s.s.Scan()

	if err := s.s.Err(); err != nil {
		return dune.NullValue, err
	}

	return dune.NewBool(v), nil
}
