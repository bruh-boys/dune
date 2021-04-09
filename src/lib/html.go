package lib

import (
	"fmt"
	"html"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(HTML, `

declare namespace html {
    export function encode(s: any): string
    export function decode(s: any): string
}


`)
}

var HTML = []dune.NativeFunction{
	{
		Name:      "html.encode",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			switch a.Type {
			case dune.Null, dune.Undefined:
				return dune.NullValue, nil
			case dune.String:
				return dune.NewString(html.EscapeString(a.String())), nil
			default:
				return dune.NewString(a.String()), nil
			}
		},
	},
	{
		Name:      "html.decode",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			switch a.Type {
			case dune.Null, dune.Undefined:
				return dune.NullValue, nil
			case dune.String:
				return dune.NewString(html.UnescapeString(a.String())), nil
			default:
				return dune.NullValue, fmt.Errorf("expected string, got %v", a.Type)
			}
		},
	},
}
