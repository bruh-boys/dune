package lib

import (
	"encoding/json"
	"fmt"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Console, `

declare namespace console {
	export function log(...v: any[]): void
}
`)
}

var Console = []dune.NativeFunction{
	{
		Name:      "console.log",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			var lastInline bool
			for i, v := range args {
				var s string
				switch v.Type {
				case dune.String, dune.Int, dune.Float, dune.Bool:
					if i > 0 {
						fmt.Fprint(vm.GetStdout(), " ")
					}
					s = v.String()
					fmt.Fprint(vm.GetStdout(), s)
					lastInline = true
				default:
					ojb := v.Export(0)
					str, ok := ojb.(fmt.Stringer)
					if ok {
						fmt.Fprintln(vm.GetStdout(), str.String())
					} else {
						b, err := json.MarshalIndent(ojb, "", "    ")
						if err != nil {
							return dune.NullValue, err
						}
						s = string(b)
						fmt.Fprintln(vm.GetStdout(), s)
					}
				}
			}

			if lastInline {
				fmt.Fprintln(vm.GetStdout())
			}

			return dune.NullValue, nil
		},
	},
}
