package lib

import (
	"image"
	"image/png"
	"io"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Png, `

declare namespace png {

    export function encode(w: io.Writer, img: Image): void

    export function decode(buf: byte[] | io.Reader): Image

    export interface Image { }
}


`)
}

var Png = []dune.NativeFunction{
	{
		Name:      "png.decode",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			r, ok := args[0].ToObject().(io.Reader)
			if !ok {
				return dune.NullValue, ErrInvalidType
			}

			img, err := png.Decode(r)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(imageObj{img}), nil
		},
	},
	{
		Name:      "png.encode",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object, dune.Object); err != nil {
				return dune.NullValue, err
			}

			w, ok := args[0].ToObject().(io.Writer)
			if !ok {
				return dune.NullValue, ErrInvalidType
			}

			i, ok := args[1].ToObject().(imageObj)
			if !ok {
				return dune.NullValue, ErrInvalidType
			}

			err := png.Encode(w, i.img)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NullValue, nil
		},
	},
}

type imageObj struct {
	img image.Image
}

func (i imageObj) Type() string {
	return "image"
}
