package lib

import (
	"fmt"
	"strings"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Reflect, `

declare namespace reflect {
    export const program: runtime.Program

	export interface Native {
		name: string
		permissions: string[]
	}

    export function nativeFunctions(): Native[]
    export function nativeProperties(): Native[]

    export function is<T>(v: any, name: string): v is T

    export function typeOf(v: any): string

    export function isValue(v: any): boolean
    export function isNativeObject(v: any): boolean
    export function isArray(v: any): boolean
    export function isMap(v: any): boolean

    export function getFunction(name: string): Function

    export function call(name: string, ...params: any[]): any

    export function runFunc(name: string, ...params: any[]): any
}


`)
}

var Reflect = []dune.NativeFunction{
	{
		Name:      "->reflect.program",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			p := vm.Program
			return dune.NewObject(&program{prog: p}), nil
		},
	},
	{
		Name:      "reflect.nativeFunctions",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := getNativeFuncions(false)
			return v, nil
		},
	},
	{
		Name:      "reflect.nativeProperties",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := getNativeFuncions(true)
			return v, nil
		},
	},
	{
		Name:      "reflect.is",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0].TypeName()
			b := args[1]
			if b.Type != dune.String {
				return dune.NullValue, fmt.Errorf("argument 2 must be a string, got %s", b.TypeName())
			}
			return dune.NewBool(a == b.String()), nil
		},
	},
	{
		Name:      "reflect.isValue",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			switch args[0].Type {
			case dune.Int, dune.Float, dune.Bool, dune.String:
				return dune.FalseValue, nil
			}
			return dune.TrueValue, nil
		},
	},
	{
		Name:      "reflect.isNativeObject",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0].Type == dune.Object
			return dune.NewBool(v), nil
		},
	},
	{
		Name:      "reflect.isArray",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0].Type == dune.Array
			return dune.NewBool(v), nil
		},
	},
	{
		Name:      "reflect.isMap",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0].Type == dune.Map
			return dune.NewBool(v), nil
		},
	},
	{
		Name:      "reflect.typeOf",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0]
			return dune.NewString(v.TypeName()), nil
		},
	},
	{
		Name:      "reflect.call",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if len(args) == 0 {
				return dune.NullValue, fmt.Errorf("expected the function name")
			}

			return vm.RunFunc(args[0].String(), args[1:]...)
		},
	},
	{
		Name:      "reflect.getFunction",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			name := args[0].String()
			fn, ok := vm.Program.Function(name)
			if !ok {
				return dune.NullValue, nil
			}

			v := dune.NewFunction(fn.Index)
			return v, nil
		},
	},
	{
		Name:      "reflect.runFunc",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l < 1 {
				return dune.NullValue, fmt.Errorf("expected at least 1 parameter, got %d", l)
			}

			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("argument must be a string, got %s", args[0].TypeName())
			}

			name := args[0].String()

			v, err := vm.RunFunc(name, args[1:]...)
			if err != nil {
				return dune.NullValue, err
			}

			return v, nil
		},
	},
}

func getNativeFuncions(properties bool) dune.Value {
	fns := dune.AllNativeFuncs()

	var v []dune.Value

	for _, f := range fns {
		name := f.Name

		if properties {
			if !strings.HasPrefix(name, "->") {
				continue
			}
			name = name[2:]
		} else {
			if strings.HasPrefix(name, "->") {
				continue
			}
		}

		m := make(map[dune.Value]dune.Value)
		m[dune.NewString("name")] = dune.NewString(name)

		perm := make([]dune.Value, len(f.Permissions))
		for i, p := range f.Permissions {
			perm[i] = dune.NewString(p)
		}

		m[dune.NewString("permissions")] = dune.NewArrayValues(perm)
		v = append(v, dune.NewMapValues(m))
	}

	return dune.NewArrayValues(v)
}
