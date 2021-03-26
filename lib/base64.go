package lib

import (
	"encoding/base64"
	"fmt"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Base64, `
	
declare namespace base64 {
    export function encode(s: any): string
    export function encodeWithPadding(s: any): string
    export function decode(s: any): string
    export function decodeWithPadding(s: any): string
}

`)
}

var Base64 = []dune.NativeFunction{
	{
		Name:      "base64.encode",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			switch a.Type {
			case dune.Null, dune.Undefined:
				return dune.NullValue, nil
			case dune.Bytes, dune.String:
				encoder := base64.RawStdEncoding
				encoded := encoder.EncodeToString(a.ToBytes())
				return dune.NewString(encoded), nil
			default:
				return dune.NullValue, fmt.Errorf("expected string, got %v", a.Type)
			}
		},
	},
	{
		Name:      "base64.encodeWithPadding",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			switch a.Type {
			case dune.Null, dune.Undefined:
				return dune.NullValue, nil
			case dune.Bytes, dune.String:
				encoder := base64.StdEncoding.WithPadding(base64.StdPadding)
				encoded := encoder.EncodeToString(a.ToBytes())
				return dune.NewString(encoded), nil
			default:
				return dune.NullValue, fmt.Errorf("expected string, got %v", a.Type)
			}
		},
	},
	{
		Name:      "base64.decode",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			switch a.Type {
			case dune.Null, dune.Undefined:
				return dune.NullValue, nil
			case dune.String:
				encoder := base64.RawStdEncoding
				encoded, err := encoder.DecodeString(a.String())
				if err != nil {
					return dune.NullValue, err
				}
				return dune.NewBytes(encoded), nil
			default:
				return dune.NullValue, fmt.Errorf("expected string, got %v", a.Type)
			}
		},
	},
	{
		Name:      "base64.decodeWithPadding",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			switch a.Type {
			case dune.Null, dune.Undefined:
				return dune.NullValue, nil
			case dune.String:
				encoder := base64.StdEncoding.WithPadding(base64.StdPadding)
				encoded, err := encoder.DecodeString(a.String())
				if err != nil {
					return dune.NullValue, err
				}
				return dune.NewString(string(encoded)), nil
			default:
				return dune.NullValue, fmt.Errorf("expected string, got %v", a.Type)
			}
		},
	},
}
