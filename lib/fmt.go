package lib

import (
	"fmt"
	"io"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(libFmt, `
declare namespace fmt {
    export function print(...n: any[]): void
    export function println(...n: any[]): void
    export function printf(format: string, ...params: any[]): void
    export function sprintf(format: string, ...params: any[]): string
	export function fprintf(w: io.Writer, format: string, ...params: any[]): void
	
    export function errorf(format: string, ...params: any[]): errors.Error
    export function typeErrorf(type: string, format: string, ...params: any[]): errors.Error
}	
	`)
}

var libFmt = []dune.NativeFunction{
	{
		Name:      "fmt.errorf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			argsLen := len(args)
			if argsLen < 1 {
				return dune.NullValue, fmt.Errorf("expected at least 1 parameter, got %d", len(args))
			}

			msg := args[0]
			if msg.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a string, got %s", msg.Type)
			}

			return typeErrorf("", msg.String(), args[1:], vm)
		},
	},
	{
		Name:      "fmt.typeErrorf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			argsLen := len(args)
			if argsLen < 2 {
				return dune.NullValue, fmt.Errorf("expected at least 2 parameters, got %d", len(args))
			}

			t := args[0]
			if t.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a string, got %s", t.Type)
			}

			msg := args[1]
			if msg.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 2 to be a string, got %s", msg.Type)
			}

			return typeErrorf(t.String(), msg.String(), args[2:], vm)
		},
	},
	{
		Name:      "fmt.print",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			for _, v := range args {
				fmt.Fprint(vm.GetStdout(), fmtValue(v))
			}
			return dune.NullValue, nil
		},
	},
	{
		Name:      "fmt.println",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			for i, v := range args {
				if i > 0 {
					fmt.Fprint(vm.GetStdout(), " ")
				}
				fmt.Fprint(vm.GetStdout(), fmtValue(v))
			}
			fmt.Fprint(vm.GetStdout(), "\n")
			return dune.NullValue, nil
		},
	},
	{
		Name:      "fmt.printf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l == 0 {
				return dune.NullValue, fmt.Errorf("expected at least 1 parameter, got %d", len(args))
			}
			v := args[0]
			if v.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a string, got %s", v.Type)
			}

			values := make([]interface{}, l-1)
			for i, v := range args[1:] {
				values[i] = fmtValue(v)
			}

			fmt.Fprintf(vm.GetStdout(), v.String(), values...)
			return dune.NullValue, nil
		},
	},
	{
		Name:      "fmt.fprintf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l < 2 {
				return dune.NullValue, fmt.Errorf("expected at least 2 parameters, got %d", len(args))
			}

			w, ok := args[0].ToObjectOrNil().(io.Writer)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a io.Writer, got %s", args[0].TypeName())
			}

			v := args[1]
			if v.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 2 to be a string, got %s", v.TypeName())
			}

			values := make([]interface{}, l-2)
			for i, v := range args[2:] {
				values[i] = fmtValue(v)
			}

			fmt.Fprintf(w, v.String(), values...)
			return dune.NullValue, nil
		},
	},
	{
		Name:      "fmt.sprintf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l == 0 {
				return dune.NullValue, fmt.Errorf("expected at least 1 parameter, got %d", len(args))
			}
			v := args[0]
			if v.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a string, got %s", v.Type)
			}

			values := make([]interface{}, l-1)
			for i, v := range args[1:] {
				values[i] = fmtValue(v)
			}

			s := fmt.Sprintf(v.String(), values...)
			return dune.NewString(s), nil
		},
	},
}

func typeErrorf(errorType, msg string, args []dune.Value, vm *dune.VM) (dune.Value, error) {
	argsLen := len(args)

	var wrap *dune.VMError

	values := make([]interface{}, argsLen)
	for i, a := range args {
		v := a.Export(0)
		if t, ok := v.(*dune.VMError); ok {
			// wrap errors showing only the message
			v = t.ErrorMessage()
			wrap = t
		}
		values[i] = v
	}

	key := Translate(msg, vm)

	if len(values) > 0 {
		key = fmt.Sprintf(key, values...)
	}

	err := dune.NewTypeError(errorType, key)
	if wrap != nil {
		err.Wrapped = wrap
	}

	return dune.NewObject(err), nil
}

func fmtValue(v dune.Value) interface{} {
	switch v.Type {
	case dune.Object:
		return v.String()
	default:
		return v.Export(0)
	}
}
