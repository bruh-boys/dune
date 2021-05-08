package lib

import (
	"bytes"
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
				if i > 0 {
					fmt.Fprint(vm.GetStdout(), " ")
				}
				var s string
				switch v.Type {
				case dune.String, dune.Int, dune.Float, dune.Bool:
					s = v.String()
					fmt.Fprint(vm.GetStdout(), s)
					lastInline = true
				default:
					obj := v.ExportMarshal(0)

					str, ok := obj.(fmt.Stringer)
					if ok {
						fmt.Fprintln(vm.GetStdout(), str.String())
						continue
					}

					if m, ok := obj.(json.Marshaler); ok {
						if err := marshal(m, vm); err != nil {
							return dune.NullValue, err
						}
						continue
					}

					if err := marshal(obj, vm); err != nil {
						return dune.NullValue, err
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

func marshal(obj interface{}, vm *dune.VM) error {
	buf := &bytes.Buffer{}

	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "    ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(obj); err != nil {
		return err
	}

	// trim the last byte because encoder.Encode adds a '\n' at the end.
	buf.Truncate(buf.Len() - 1)

	fmt.Fprintln(vm.GetStdout(), buf.String())
	return nil
}
