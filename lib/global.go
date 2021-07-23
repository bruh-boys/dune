package lib

import (
	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Global, `
	
declare namespace global {
    export const value: any
}

`)
}

var globalValue = dune.NewMap(0)

var Global = []dune.NativeFunction{
	{
		Name:      "->global.value",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return globalValue, nil
		},
	},
}
