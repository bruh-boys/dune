package lib

import (
	"bytes"
	"fmt"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Bytes, `
declare namespace bytes {
	export function newReader(b: byte[]): io.Reader
}	
`)
}

var Bytes = []dune.NativeFunction{
	{
		Name:      "bytes.newReader",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			r := args[0].ToBytes()

			s := bytes.NewReader(r)

			reader := NewReader(s)

			return dune.NewObject(reader), nil
		},
	},
	{
		Name:      "Bytes.prototype.copyAt",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be int, got %s", args[0].TypeName())
			}

			switch args[1].Type {
			case dune.Bytes, dune.Array, dune.String:
			default:
				return dune.NullValue, fmt.Errorf("expected arg 2 to be bytes, got %s", args[1].TypeName())
			}

			a := this.ToBytes()
			start := int(args[0].ToInt())
			b := args[1].ToBytes()

			lenB := len(b)

			if lenB+start > len(a) {
				return dune.NullValue, fmt.Errorf("the array has not enough capacity")
			}

			for i := 0; i < lenB; i++ {
				a[i+start] = b[i]
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:      "Bytes.prototype.append",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Bytes {
				return dune.NullValue, fmt.Errorf("expected byte array, got %s", this.TypeName())
			}
			a := this.ToBytes()

			b := args[0]
			if b.Type != dune.Bytes {
				return dune.NullValue, fmt.Errorf("expected array, got %s", b.TypeName())
			}

			c := append(a, b.ToBytes()...)

			return dune.NewBytes(c), nil
		},
	},
	{
		Name:      "Bytes.prototype.indexOf",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Bytes {
				return dune.NullValue, fmt.Errorf("expected byte array, got %s", this.TypeName())
			}
			a := this.ToBytes()
			v := byte(args[0].ToInt())

			for i, j := range a {
				if j == v {
					return dune.NewInt(i), nil
				}
			}

			return dune.NewInt(-1), nil
		},
	},
	{
		Name: "Bytes.prototype.reverse",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Bytes {
				return dune.NullValue, fmt.Errorf("expected byte array, got %s", this.TypeName())
			}
			a := this.ToBytes()
			l := len(a) - 1

			for i, k := 0, l/2; i <= k; i++ {
				a[i], a[l-i] = a[l-i], a[i]
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:      "Bytes.prototype.slice",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Bytes {
				return dune.NullValue, fmt.Errorf("expected string array, got %s", this.TypeName())
			}
			a := this.ToBytes()
			l := len(a)

			switch len(args) {
			case 0:
				a = a[0:]
			case 1:
				a = a[int(args[0].ToInt()):]
			case 2:
				start := int(args[0].ToInt())
				if start < 0 || start > l {
					return dune.NullValue, fmt.Errorf("index out of range")
				}

				end := start + int(args[1].ToInt())
				if end < 0 || end > l {
					return dune.NullValue, fmt.Errorf("index out of range")
				}

				a = a[start:end]
			default:
				return dune.NullValue, fmt.Errorf("expected 0, 1 or 2 params, got %d", len(args))
			}

			return dune.NewBytes(a), nil
		},
	},
	{
		Name:      "Bytes.prototype.range",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Bytes {
				return dune.NullValue, fmt.Errorf("expected array, called on %s", this.TypeName())
			}
			a := this.ToBytes()
			l := len(a)

			switch len(args) {
			case 0:
				a = a[0:]
			case 1:
				a = a[int(args[0].ToInt()):]
			case 2:
				start := int(args[0].ToInt())
				if start < 0 || start > l {
					return dune.NullValue, fmt.Errorf("index out of range")
				}

				end := int(args[1].ToInt())
				if end < 0 || end > l {
					return dune.NullValue, fmt.Errorf("index out of range")
				}

				a = a[start:end]
			default:
				return dune.NullValue, fmt.Errorf("expected 0, 1 or 2 params, got %d", len(args))
			}

			return dune.NewBytes(a), nil
		},
	},
}
