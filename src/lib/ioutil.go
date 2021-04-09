package lib

import (
	"bytes"
	"fmt"
	"io"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(IOUtil, `

declare namespace ioutil {
    export function readAll(r: io.Reader): byte[]
}

`)
}

var IOUtil = []dune.NativeFunction{
	{
		Name:      "ioutil.readAll",
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

			b, err := ReadAll(reader, vm)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewBytes(b), nil
		},
	},
}

func ReadAll(reader io.Reader, vm *dune.VM) ([]byte, error) {
	//return ioutil.ReadAll(reader)

	b := make([]byte, bytes.MinRead)
	var buf bytes.Buffer

	for {
		n, err := reader.Read(b)
		buf.Write(b[:n])

		if n < bytes.MinRead || err == io.EOF {
			return buf.Bytes(), nil
		}

		if err != nil {
			return nil, err
		}

		if err := vm.AddAllocations(n); err != nil {
			return nil, err
		}
	}
}
