package lib

import (
	"io"

	"github.com/dunelang/dune"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

func init() {
	dune.RegisterLib(Encoding, `
	
declare namespace encoding {
    export interface Decoder {
        reader(r: io.Reader): io.Reader
    }
    export interface Encoder {
		writer(r: io.Writer): io.Writer
		string(s: string): string
    }

    export function newDecoderISO8859_1(): Decoder
    export function newEncoderISO8859_1(): Encoder
    export function newDecoderWindows1252(): Decoder
    export function newEncoderWindows1252(): Encoder
    export function newDecoderUTF16_LE(): Decoder
    export function newEncoderUTF16_LE(): Encoder
}

`)
}

var Encoding = []dune.NativeFunction{
	{
		Name:      "encoding.newDecoderISO8859_1",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			d := charmap.ISO8859_1.NewDecoder()
			return dune.NewObject(&decoder{d}), nil
		},
	},
	{
		Name:      "encoding.newEncoderISO8859_1",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			d := charmap.ISO8859_1.NewEncoder()
			return dune.NewObject(&encoder{d}), nil
		},
	},
	{
		Name:      "encoding.newDecoderWindows1252",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			d := charmap.Windows1252.NewDecoder()
			return dune.NewObject(&decoder{d}), nil
		},
	},
	{
		Name:      "encoding.newEncoderWindows1252",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			d := charmap.Windows1252.NewEncoder()
			return dune.NewObject(&encoder{d}), nil
		},
	},
	{
		Name:      "encoding.newDecoderUTF16_LE",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			d := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
			return dune.NewObject(&decoder{d}), nil
		},
	},
	{
		Name:      "encoding.newEncoderUTF16_LE",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			e := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
			return dune.NewObject(&encoder{e}), nil
		},
	},
}

type decoder struct {
	d *encoding.Decoder
}

func (d *decoder) Type() string {
	return "encoding.Decoder"
}

func (d *decoder) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "reader":
		return d.reader
	}
	return nil
}

func (d *decoder) reader(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	r, ok := args[0].ToObject().(io.Reader)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	rd := &reader{d.d.Reader(r)}

	return dune.NewObject(rd), nil
}

type encoder struct {
	d *encoding.Encoder
}

func (d *encoder) Type() string {
	return "encoding.Encoder"
}

func (d *encoder) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "string":
		return d.string
	case "writer":
		return d.writer
	}
	return nil
}

func (d *encoder) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	s, err := d.d.String(args[0].String())
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewString(s), nil
}

func (d *encoder) writer(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	w, ok := args[0].ToObject().(io.Writer)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	rd := &writer{d.d.Writer(w)}

	return dune.NewObject(rd), nil
}
