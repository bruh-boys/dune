package lib

import (
	"crypto/md5"
	"crypto/sha256"
	"hash"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(HASH, `

declare namespace hash {
    export function newMD5(): Hash
    export function newSHA256(): Hash

    export interface Hash {
        write(b: byte[]): number
        sum(): byte[]
    }
}


`)
}

var HASH = []dune.NativeFunction{
	{
		Name: "hash.newMD5",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			hash := md5.New()
			return dune.NewObject(hasher{hash}), nil
		},
	},
	{
		Name: "hash.newSHA256",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			hash := sha256.New()
			return dune.NewObject(hasher{hash}), nil
		},
	},
}

type hasher struct {
	h hash.Hash
}

func (hasher) Type() string {
	return "hash.Hash"
}

func (h hasher) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "write":
		return h.write
	case "sum":
		return h.sum
	}
	return nil
}

func (h hasher) Write(p []byte) (n int, err error) {
	return h.h.Write(p)
}

func (h hasher) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Bytes); err != nil {
		return dune.NullValue, err
	}

	buf := args[0].ToBytes()

	n, err := h.h.Write(buf)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewInt(n), nil
}

func (h hasher) sum(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	b := h.h.Sum(nil)
	return dune.NewBytes(b), nil
}
