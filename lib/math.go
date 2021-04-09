package lib

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/dunelang/dune"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	dune.RegisterLib(Math, `

declare namespace math {
    /**
     * returns, as an int, a non-negative pseudo-random number in (0,n)
     */
    export function rand(n: number): number

    export function abs(n: number): number

    export function pow(n: number, exp: number): number

    export function min(nums: number[]): number

	export function floor(n: number): number
	
    export function ceil(n: number): number

    export function round(n: number, decimals?: number): number

    export function median(nums: number[]): number

    export function standardDev(nums: number[]): number
}

`)
}

var Math = []dune.NativeFunction{
	{
		Name:      "math.pow",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Float, dune.Float); err != nil {
				return dune.NullValue, err
			}
			v := math.Pow(args[0].ToFloat(), args[1].ToFloat())
			return dune.NewFloat(v), nil
		},
	},
	{
		Name:      "math.abs",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}
			v := math.Abs(args[0].ToFloat())
			return dune.NewFloat(v), nil
		},
	},
	{
		Name:      "math.floor",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}
			v := math.Floor(args[0].ToFloat())
			return dune.NewInt64(int64(v)), nil
		},
	},
	{
		Name:      "math.ceil",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}
			v := math.Ceil(args[0].ToFloat())
			return dune.NewInt64(int64(v)), nil
		},
	},
	{
		Name:      "math.round",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {

			l := len(args)
			if l > 2 {
				return dune.NullValue, fmt.Errorf("expected 1 or 2 params, got %d", l)
			}

			f := args[0]
			switch f.Type {
			case dune.Float, dune.Int:
			default:
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a number, got %s", f.TypeName())
			}

			if l == 1 {
				v := math.Round(f.ToFloat())
				return dune.NewInt64(int64(v)), nil
			}

			d := args[1]
			if d.Type != dune.Int {
				return dune.NullValue, fmt.Errorf("expected parameter 2 to be int, got %s", d.TypeName())
			}

			i := math.Pow10(int(d.ToInt()))
			v := math.Round(f.ToFloat()*i) / i
			return dune.NewFloat(v), nil
		},
	},
	{
		Name:      "math.rand",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}
			v := rand.Intn(int(args[0].ToInt()))
			return dune.NewInt(v), nil
		},
	},
	{
		Name:      "math.median",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Array); err != nil {
				return dune.NullValue, err
			}

			a := args[0].ToArray()

			values := make([]float64, len(a))

			for i, v := range a {
				switch v.Type {
				case dune.Int, dune.Float:
					values[i] = v.ToFloat()
				default:
					return dune.NullValue, fmt.Errorf("element at %d is not a number: %s", i, v.TypeName())
				}
			}

			r := median(values)
			return dune.NewFloat(r), nil
		},
	},
	{
		Name:      "math.min",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Array); err != nil {
				return dune.NullValue, err
			}

			a := args[0].ToArray()
			var min float64

			for i, v := range a {
				switch v.Type {
				case dune.Int, dune.Float:
					k := v.ToFloat()
					if i == 0 {
						min = k
					} else if k < min {
						min = k
					}
				default:
					return dune.NullValue, fmt.Errorf("element at %d is not a number: %s", i, v.TypeName())
				}
			}

			return dune.NewFloat(min), nil
		},
	},
	{
		Name:      "math.standardDev",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Array); err != nil {
				return dune.NullValue, err
			}

			a := args[0].ToArray()

			values := make([]float64, len(a))
			for i, v := range a {
				switch v.Type {
				case dune.Int, dune.Float:
					values[i] = v.ToFloat()
				default:
					return dune.NullValue, fmt.Errorf("element at %d is not a number: %s", i, v.TypeName())
				}
			}

			m := median(values)
			d := stdDev(values, m)
			return dune.NewFloat(d), nil
		},
	},
}

func median(numbers []float64) float64 {
	middle := len(numbers) / 2
	result := numbers[middle]
	if len(numbers)%2 == 0 {
		result = (result + numbers[middle-1]) / 2
	}
	return result
}

func stdDev(numbers []float64, mean float64) float64 {
	total := 0.0
	for _, number := range numbers {
		total += math.Pow(number-mean, 2)
	}
	variance := total / float64(len(numbers)-1)
	return math.Sqrt(variance)
}
