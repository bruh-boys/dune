package lib

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Convert, `

declare namespace convert {
    export function toInt(v: string | number): number
    export function toFloat(v: string | number): number
    export function toString(v: any): string
    export function toRune(v: any): string
    export function toBool(v: string | number | boolean): boolean
    export function toBytes(v: string | byte[]): byte[]
}

`)
}

var Convert = []dune.NativeFunction{
	{
		Name:      "convert.toByte",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			var r dune.Value

			switch a.Type {
			case dune.String:
			default:
				return dune.NullValue, fmt.Errorf("can't convert %v to byte", a.Type)
			}

			s := a.String()
			if len(s) != 1 {
				return dune.NullValue, fmt.Errorf("can't convert %v to int", a.Type)
			}

			return r, nil
		},
	},
	{
		Name:      "convert.toRune",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]

			switch a.Type {
			case dune.String:
				s := a.String()
				if len(s) != 1 {
					return dune.NullValue, fmt.Errorf("can't convert %v to rune", s)
				}
				return dune.NewRune(rune(s[0])), nil
			case dune.Int:
				i := a.ToInt()
				if i > 255 {
					return dune.NullValue, fmt.Errorf("can't convert %v to rune", i)
				}
				return dune.NewRune(rune(i)), nil
			default:
				return dune.NullValue, fmt.Errorf("can't convert %v to byte", a.Type)
			}
		},
	},
	{
		Name:      "convert.toInt",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			var r dune.Value

			switch a.Type {
			case dune.Int:
				r = a
			case dune.Float:
				r = dune.NewInt64(a.ToInt())
			case dune.Rune:
				r = dune.NewInt64(a.ToInt())
			case dune.String:
				s := strings.Trim(a.String(), " ")
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return dune.NullValue, err
				}
				r = dune.NewInt64(i)
			default:
				return dune.NullValue, fmt.Errorf("can't convert %v to int", a.Type)
			}

			return r, nil
		},
	},
	{
		Name:      "convert.toFloat",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			switch a.Type {
			case dune.Int:
				return dune.NewFloat(a.ToFloat()), nil
			case dune.Float:
				return a, nil
			case dune.String:
				s := strings.Trim(a.String(), " ")
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return dune.NullValue, err
				}
				return dune.NewFloat(f), nil
			default:
				return dune.NullValue, fmt.Errorf("can't convert %v to int", a.Type)
			}
		},
	},
	{
		Name:      "convert.toBool",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			var r dune.Value

			switch a.Type {

			case dune.Bool:
				r = a

			case dune.Int:
				switch a.ToInt() {
				case 0:
					r = dune.FalseValue
				case 1:
					r = dune.TrueValue
				default:
					return dune.NullValue, fmt.Errorf("can't convert %v to bool", a.Type)
				}

			case dune.String:
				s := a.String()
				s = strings.Trim(s, " ")
				switch s {
				case "true", "1":
					r = dune.TrueValue
				case "false", "0":
					r = dune.FalseValue
				default:
					return dune.NullValue, fmt.Errorf("can't convert %v to bool", s)
				}

			default:
				return dune.NullValue, fmt.Errorf("can't convert %v to bool", a.Type)

			}

			return r, nil
		},
	},
	{
		Name:      "convert.toString",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			return dune.NewString(a.String()), nil
		},
	},
	{
		Name:      "convert.toBytes",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			var r dune.Value

			switch a.Type {
			case dune.String:
				r = dune.NewBytes(a.ToBytes())
			case dune.Bytes:
				r = a
			default:
				return dune.NullValue, fmt.Errorf("can't convert %v to int", a.Type)
			}

			return r, nil
		},
	},
}
