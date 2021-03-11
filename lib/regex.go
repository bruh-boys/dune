package lib

import (
	"regexp"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Regex, `

declare namespace regex {
    export function match(pattern: string, value: string): boolean
    export function split(pattern: string, value: string): string[]
    export function findAllString(pattern: string, value: string, count?: number): string[]
    export function findAllStringSubmatch(pattern: string, value: string, count?: number): string[][]
    export function findAllStringSubmatchIndex(pattern: string, value: string, count?: number): number[][]
    export function replaceAllString(pattern: string, source: string, replace: string): string
}

`)
}

var Regex = []dune.NativeFunction{
	{
		Name:      "regex.match",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			ok, err := regexp.MatchString(args[0].ToString(), args[1].ToString())
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewBool(ok), nil
		},
	},
	{
		Name:      "regex.split",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			r, err := regexp.Compile(args[0].ToString())
			if err != nil {
				return dune.NullValue, err
			}

			matches := r.Split(args[1].ToString(), -1)

			ln := len(matches)
			result := make([]dune.Value, ln)
			for i := 0; i < ln; i++ {
				result[i] = dune.NewString(matches[i])
			}

			return dune.NewArrayValues(result), nil
		},
	},
	{
		Name:      "regex.findAllStringSubmatchIndex",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgRange(args, 2, 3); err != nil {
				return dune.NullValue, err
			}
			if err := ValidateOptionalArgs(args, dune.String, dune.String, dune.Int); err != nil {
				return dune.NullValue, err
			}

			r, err := regexp.Compile(args[0].ToString())
			if err != nil {
				return dune.NullValue, err
			}

			var i int
			if len(args) == 3 {
				i = int(args[2].ToInt())
			} else {
				i = -1
			}

			matches := r.FindAllStringSubmatchIndex(args[1].ToString(), i)

			var result []dune.Value
			for _, v := range matches {
				a := []dune.Value{dune.NewInt(v[0]), dune.NewInt(v[1])}
				result = append(result, dune.NewArrayValues(a))
			}

			return dune.NewArrayValues(result), nil
		},
	},
	{
		Name:      "regex.findAllString",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgRange(args, 2, 3); err != nil {
				return dune.NullValue, err
			}
			if err := ValidateOptionalArgs(args, dune.String, dune.String, dune.Int); err != nil {
				return dune.NullValue, err
			}

			r, err := regexp.Compile(args[0].ToString())
			if err != nil {
				return dune.NullValue, err
			}

			var i int
			if len(args) == 3 {
				i = int(args[2].ToInt())
			} else {
				i = -1
			}

			matches := r.FindAllString(args[1].ToString(), i)

			var result []dune.Value

			for _, v := range matches {
				result = append(result, dune.NewString(v))
			}

			return dune.NewArrayValues(result), nil
		},
	},
	{
		Name:      "regex.findAllStringSubmatch",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgRange(args, 2, 3); err != nil {
				return dune.NullValue, err
			}
			if err := ValidateOptionalArgs(args, dune.String, dune.String, dune.Int); err != nil {
				return dune.NullValue, err
			}

			r, err := regexp.Compile(args[0].ToString())
			if err != nil {
				return dune.NullValue, err
			}

			var i int
			if len(args) == 3 {
				i = int(args[2].ToInt())
			} else {
				i = -1
			}

			matches := r.FindAllStringSubmatch(args[1].ToString(), i)

			var result []dune.Value

			for _, v := range matches {
				var subResult []dune.Value
				for _, sv := range v {
					subResult = append(subResult, dune.NewString(sv))
				}
				result = append(result, dune.NewArrayValues(subResult))
			}

			return dune.NewArrayValues(result), nil
		},
	},
	{
		Name:      "regex.replaceAllString",
		Arguments: 3,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			r, err := regexp.Compile(args[0].ToString())
			if err != nil {
				return dune.NullValue, err
			}

			result := r.ReplaceAllString(args[1].ToString(), args[2].ToString())

			return dune.NewString(result), nil
		},
	},
}
