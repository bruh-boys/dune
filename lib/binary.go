package lib

import (
	"encoding/binary"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Binary, `
	
declare namespace binary {
    export function putInt16LittleEndian(v: byte[], n: number): void
    export function putInt32LittleEndian(v: byte[], n: number): void
    export function putInt64LittleEndian(v: byte[], n: number): void

    export function putInt16BigEndian(v: byte[], n: number): void
    export function putInt32BigEndian(v: byte[], n: number): void
    export function putInt64BigEndian(v: byte[], n: number): void

    export function int16LittleEndian(v: byte[]): number
    export function int32LittleEndian(v: byte[]): number
    export function int64LittleEndian(v: byte[]): number

    export function int16BigEndian(v: byte[]): number
    export function int32BigEndian(v: byte[]): number
    export function int64BigEndian(v: byte[]): number
}
 
`)
}

var Binary = []dune.NativeFunction{
	{
		Name:      "binary.putInt16LittleEndian",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Int); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := args[1].ToInt()
			binary.LittleEndian.PutUint16(b, uint16(i))
			return dune.NullValue, nil
		},
	},
	{
		Name:      "binary.putInt32LittleEndian",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Int); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := args[1].ToInt()
			binary.LittleEndian.PutUint32(b, uint32(i))
			return dune.NullValue, nil
		},
	},
	{
		Name:      "binary.putInt64LittleEndian",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Int); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := args[1].ToInt()
			binary.LittleEndian.PutUint64(b, uint64(i))
			return dune.NullValue, nil
		},
	},
	{
		Name:      "binary.putInt16BigEndian",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Int); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := args[1].ToInt()
			binary.BigEndian.PutUint16(b, uint16(i))
			return dune.NullValue, nil
		},
	},
	{
		Name:      "binary.putInt32BigEndian",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Int); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := args[1].ToInt()
			binary.BigEndian.PutUint32(b, uint32(i))
			return dune.NullValue, nil
		},
	},
	{
		Name:      "binary.putInt64BigEndian",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Int); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := args[1].ToInt()
			binary.BigEndian.PutUint64(b, uint64(i))
			return dune.NullValue, nil
		},
	},
	{
		Name:      "binary.int16LittleEndian",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := binary.LittleEndian.Uint16(b)
			return dune.NewInt64(int64(i)), nil
		},
	},
	{
		Name:      "binary.int32LittleEndian",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := binary.LittleEndian.Uint32(b)
			return dune.NewInt64(int64(i)), nil
		},
	},
	{
		Name:      "binary.int64LittleEndian",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := binary.LittleEndian.Uint64(b)
			return dune.NewInt64(int64(i)), nil
		},
	},
	{
		Name:      "binary.int16BigEndian",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := binary.BigEndian.Uint16(b)
			return dune.NewInt64(int64(i)), nil
		},
	},
	{
		Name:      "binary.int32BigEndian",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := binary.BigEndian.Uint32(b)
			return dune.NewInt64(int64(i)), nil
		},
	},
	{
		Name:      "binary.int64BigEndian",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()
			i := binary.BigEndian.Uint64(b)
			return dune.NewInt64(int64(i)), nil
		},
	},
}
