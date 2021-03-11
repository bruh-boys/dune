package lib

import (
	"encoding/hex"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(HEX, `

declare namespace hex {
    export function encodeToString(b: byte[]): number
}


`)
}

var HEX = []dune.NativeFunction{
	{
		Name:      "hex.encodeToString",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			s := hex.EncodeToString(b)
			return dune.NewString(s), nil
		},
	},
}
