package lib

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Strconv, `

declare namespace strconv {
    export function formatInt(i: number, base: number): string
    export function parseInt(s: string, base: number, bitSize: number): number
    export function formatCustomBase34(i: number): string
    export function parseCustomBase34(s: string): number


    export function formatRef(i: number, length?: number): string
    export function parseRef(ref: string): number

}

`)
}

var Strconv = []dune.NativeFunction{
	{
		Name:      "strconv.formatRef",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Int, dune.Int); err != nil {
				return dune.NullValue, err
			}

			s := EncodeCustomBase34(uint64(args[0].ToInt()))

			var size int
			if len(args) > 1 {
				size = int(args[1].ToInt())
				s += customBase34Delimiter + RandomAlphanumeric(size)
			}

			return dune.NewString(strings.ToUpper(s)), nil
		},
	},
	{
		Name:      "strconv.parseRef",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			s := args[0].ToString()

			i := strings.IndexRune(s, 'L')
			if i != -1 {
				s = s[:i]
			}

			v := DecodeCustomBase34(s)
			return dune.NewInt64(int64(v)), nil
		},
	},
	{
		Name:      "strconv.formatInt",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int, dune.Int); err != nil {
				return dune.NullValue, err
			}
			v := strconv.FormatInt(args[0].ToInt(), int(args[1].ToInt()))
			return dune.NewString(v), nil
		},
	},
	{
		Name:      "strconv.parseInt",
		Arguments: 3,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.Int, dune.Int); err != nil {
				return dune.NullValue, err
			}
			v, err := strconv.ParseInt(args[0].ToString(), int(args[1].ToInt()), int(args[2].ToInt()))
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewInt64(v), nil
		},
	},
	{
		Name:      "strconv.formatCustomBase34",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}
			v := EncodeCustomBase34(uint64(args[0].ToInt()))
			return dune.NewString(v), nil
		},
	},
	{
		Name:      "strconv.parseCustomBase34",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			v := DecodeCustomBase34(args[0].ToString())
			return dune.NewInt64(int64(v)), nil
		},
	},
}

// The L is not present so it can be used as delimiter
const customBase34Delimiter = "L"

var (
	base34 = [34]byte{
		'2', '3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H',
		'J', 'K', 'M', 'N', 'P', 'Q', 'R', 'S',
		'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'0', '1', 'I'}

	index = map[byte]int{
		'2': 0, '3': 1, '4': 2, '5': 3, '6': 4,
		'7': 5, '8': 6, '9': 7, 'A': 8, 'B': 9,
		'C': 10, 'D': 11, 'E': 12, 'F': 13, 'G': 14,
		'H': 15, 'J': 16, 'K': 17, 'M': 18, 'N': 19,
		'P': 20, 'Q': 21, 'R': 22, 'S': 23, 'T': 24,
		'U': 25, 'V': 26, 'W': 27, 'X': 28, 'Y': 29,
		'Z': 30, 'a': 8, 'b': 9, 'c': 10, 'd': 11,
		'e': 12, 'f': 13, 'g': 14, 'h': 15, 'j': 16,
		'k': 17, 'm': 18, 'n': 19, 'p': 20, 'q': 21,
		'r': 22, 's': 23, 't': 24, 'u': 25, 'v': 26,
		'w': 27, 'x': 28, 'y': 29, 'z': 30,
		'0': 31, '1': 32, 'I': 33}
)

// EncodeCustomBase34 encodes a uint64 value to string in base34 format
func EncodeCustomBase34(value uint64) string {
	var res [16]byte
	var i int
	for i = len(res) - 1; value != 0; i-- {
		res[i] = base34[value%34]
		value /= 34
	}
	return string(res[i+1:])
}

// DecodeCustomBase34 decodes a base34-encoded string back to uint64
func DecodeCustomBase34(s string) uint64 {
	res := uint64(0)
	l := len(s) - 1
	b34 := big.NewInt(34)
	bidx := big.NewInt(0)
	bpow := big.NewInt(0)
	for idx := range s {
		c := s[l-idx]
		byteOffset := index[c]
		bidx.SetUint64(uint64(idx))
		res += uint64(byteOffset) * bpow.Exp(b34, bidx, nil).Uint64()
	}
	return res
}
