package lib

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dunelang/dune"
)

var ErrReadOnly = errors.New("readonly property")
var ErrReadOnlyOrUndefined = errors.New("undefined or readonly property")
var ErrUndefined = errors.New("undefined")
var ErrInvalidType = errors.New("invalid value type")
var ErrFileNotFound = errors.New("file not found")
var ErrUnauthorized = errors.New("unauthorized")
var ErrNoFileSystem = errors.New("there is no filesystem")

func init() {
	dune.RegisterLib(Errors, `
declare namespace errors {
	export function parse(err: string): Error
	export function newError(msg: string, ...args: any[]): Error
	export function newTypeError(type: string, msg: string, ...args: any[]): Error
	export function unwrap(err: Error): Error
	export function is(err: Error, type: string): Error
	export function rethrow(err: Error): void

	export interface Error {
		type: string
		message: string
		pc: number
		stackTrace: string
		string(): string
		is(error: string): boolean
	} 
}

`)
}

var Errors = []dune.NativeFunction{
	{
		Name:      "errors.parse",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			var e dune.VMError

			err := json.Unmarshal(args[0].ToBytes(), &e)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(&e), nil
		},
	},
	{
		Name:      "errors.is",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object, dune.String); err != nil {
				return dune.NullValue, err
			}
			e, ok := args[0].ToObjectOrNil().(*dune.VMError)
			if !ok {
				return dune.FalseValue, nil
			}
			return dune.NewBool(e.Is(args[1].String())), nil
		},
	},
	{
		Name:      "errors.unwrap",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			e, ok := args[0].ToObjectOrNil().(*dune.VMError)
			if !ok {
				return dune.FalseValue, nil
			}

			e = e.Wrapped

			if e == nil {
				return dune.NullValue, nil
			}

			return dune.NewObject(e), nil
		},
	},
	{
		Name:      "errors.rethrow",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			e, ok := args[0].ToObjectOrNil().(*dune.VMError)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected error, got %s", args[0].String())
			}

			e.IsRethrow = true

			return dune.NullValue, e
		},
	},
	{
		Name:      "errors.newError",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			argsLen := len(args)
			if argsLen < 1 {
				return dune.NullValue, fmt.Errorf("expected at least 1 parameter, got %d", len(args))
			}

			m := args[0]
			if m.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a string, got %s", m.Type)
			}

			return newTypeError("", m.String(), args[1:], vm)
		},
	},
	{
		Name:      "errors.newTypeError",
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

			m := args[1]
			if m.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 2 to be a string, got %s", m.Type)
			}

			return newTypeError(t.String(), m.String(), args[2:], vm)
		},
	},
}

func newTypeError(errorType, msg string, args []dune.Value, vm *dune.VM) (dune.Value, error) {
	argsLen := len(args)

	values := make([]interface{}, argsLen)
	for i, a := range args {
		v := a.Export(0)
		if t, ok := v.(*dune.VMError); ok {
			// wrap errors showing only the message
			v = t.ErrorMessage()
		}
		values[i] = v
	}

	msg = fmt.Sprintf(msg, values...)

	err := vm.NewError(msg)
	err.ErrorType = errorType
	return dune.NewObject(err), nil
}
