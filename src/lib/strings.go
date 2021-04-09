package lib

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dunelang/dune"

	"golang.org/x/exp/utf8string"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	dune.RegisterLib(Strings, `
	
declare namespace strings {
    export function newReader(a: string): io.Reader
}

interface String {
    runeAt(i: number): string
}

declare namespace strings {
    export function equalFold(a: string, b: string): boolean
    export function isChar(value: string): boolean
    export function isDigit(value: string): boolean
    export function isIdent(value: string): boolean
    export function isAlphanumeric(value: string): boolean
    export function isAlphanumericIdent(value: string): boolean
    export function isNumeric(value: string): boolean
	export function sort(a: string[]): void
	export function repeat(value: string, count: number): string
}
	  
interface String {
    [n: number]: string 

    /**
     * Gets the length of the string.
     */
    length: number

    /**
     * The number of bytes oposed to the number of runes returned by length.
     */
    runeCount: number

    toLower(): string

    toUpper(): string

    toTitle(): string

    toUntitle(): string

    replace(oldValue: string, newValue: string, times?: number): string

    hasPrefix(prefix: string): boolean
    hasSuffix(prefix: string): boolean

    trim(cutset?: string): string
    trimLeft(cutset?: string): string
    trimRight(cutset?: string): string
    trimPrefix(prefix: string): string
    trimSuffix(suffix: string): string

    rightPad(pad: string, total: number): string
    leftPad(pad: string, total: number): string

    take(to: number): string
    substring(from: number, to?: number): string
    runeSubstring(from: number, to?: number): string

    split(s: string): string[]
    splitEx(s: string): string[]

    contains(s: string): boolean
    equalFold(s: string): boolean

    indexOf(s: string, start?: number): number
    lastIndexOf(s: string, start?: number): number


	/**
	 * Replace with regular expression.
	 * The syntax is defined: https://golang.org/pkg/regexp/syntax
	 */
    replaceRegex(expr: string, replace: string): string
}

	`)
}

var Strings = []dune.NativeFunction{
	{
		Name:      "strings.newReader",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			r := strings.NewReader(args[0].String())

			return dune.NewObject(&reader{r}), nil
		},
	},
	{
		Name:      "String.prototype.runeAt",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			if a.Type != dune.Int {
				return dune.NullValue, fmt.Errorf("expected int, got %s", a.Type)
			}

			i := int(a.ToInt())

			if i < 0 {
				return dune.NullValue, vm.NewError("Index out of range in string")
			}

			// TODO: prevent this in every call
			v := utf8string.NewString(this.String())

			if int(i) >= v.RuneCount() {
				return dune.NullValue, vm.NewError("Index out of range in string")
			}

			return dune.NewRune(v.At(int(i))), nil
		},
	},
	{
		Name:      "strings.equalFold",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			b := args[1]

			switch a.Type {
			case dune.Null, dune.Undefined:
				switch b.Type {
				case dune.Null, dune.Undefined:
					return dune.TrueValue, nil
				case dune.String:
					return dune.FalseValue, nil
				default:
					return dune.NullValue, fmt.Errorf("expected argument 2 to be string got %v", b.Type)
				}
			case dune.String:
			default:
				return dune.NullValue, fmt.Errorf("expected argument 1 to be string got %v", a.Type)
			}

			switch b.Type {
			case dune.Null, dune.Undefined:
				// a cant be null at this point
				return dune.FalseValue, nil
			case dune.String:
			default:
				return dune.NullValue, fmt.Errorf("expected argument 2 to be string got %v", b.Type)
			}

			eq := strings.EqualFold(a.String(), b.String())
			return dune.NewBool(eq), nil
		},
	},
	{
		Name:      "strings.isIdent",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0]
			switch v.Type {
			case dune.String, dune.Rune:
			default:
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", v.TypeName())
			}

			b := IsIdent(v.String())
			return dune.NewBool(b), nil
		},
	},
	{
		Name:      "strings.isAlphanumeric",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0]
			switch v.Type {
			case dune.String, dune.Rune:
			default:
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", v.TypeName())
			}
			b := IsAlphanumeric(v.String())
			return dune.NewBool(b), nil
		},
	},
	{
		Name:      "strings.isAlphanumericIdent",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0]
			switch v.Type {
			case dune.String, dune.Rune:
			default:
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", v.TypeName())
			}
			b := IsAlphanumericIdent(v.String())
			return dune.NewBool(b), nil
		},
	},
	{
		Name:      "strings.isNumeric",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0]
			switch v.Type {
			case dune.String, dune.Rune:
			case dune.Int, dune.Float:
				return dune.TrueValue, nil
			default:
				return dune.FalseValue, nil
			}

			b := IsNumeric(v.String())
			return dune.NewBool(b), nil
		},
	},
	{
		Name:      "strings.isChar",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0]
			switch v.Type {
			case dune.String, dune.Rune:
			default:
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}

			s := v.String()
			if len(s) != 1 {
				return dune.FalseValue, nil
			}

			r := rune(s[0])
			if 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z' {
				return dune.TrueValue, nil
			}

			return dune.FalseValue, nil
		},
	},
	{
		Name:      "strings.isDigit",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0]
			switch v.Type {
			case dune.String, dune.Rune:
			default:
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}

			s := v.String()
			if len(s) != 1 {
				return dune.FalseValue, nil
			}

			r := rune(s[0])
			if '0' <= r && r <= '9' {
				return dune.TrueValue, nil
			}

			return dune.FalseValue, nil
		},
	},
	{
		Name:      "strings.sort",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be array, got %s", args[0].TypeName())
			}

			a := args[0].ToArray()

			s := make([]string, len(a))

			for i, v := range a {
				s[i] = v.String()
			}

			sort.Strings(s)

			for i, v := range s {
				a[i] = dune.NewString(v)
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:      "strings.repeat",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.Int); err != nil {
				return dune.NullValue, err
			}

			a := args[0].String()
			b := int(args[1].ToInt())

			values := make([]string, b)

			for i := 0; i < b; i++ {
				values[i] = a
			}

			s := strings.Join(values, "")
			return dune.NewString(s), nil
		},
	},
	{
		Name:      "String.prototype.replaceRegex",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			exp := args[0].String()
			repl := args[1].String()
			s := this.String()
			r, err := regexp.Compile(exp)
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewString(r.ReplaceAllString(s, repl)), nil
		},
	},
	{
		Name: "String.prototype.toLower",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			s := this.String()
			return dune.NewString(strings.ToLower(s)), nil
		},
	},
	{
		Name: "String.prototype.toUpper",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			s := this.String()
			return dune.NewString(strings.ToUpper(s)), nil
		},
	},
	{
		Name: "String.prototype.toTitle",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			s := this.String()
			if len(s) > 0 {
				s = strings.ToUpper(s[:1]) + s[1:]
			}
			return dune.NewString(s), nil
		},
	},
	{
		Name: "String.prototype.toUntitle",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			s := this.String()
			if len(s) > 0 {
				s = strings.ToLower(s[:1]) + s[1:]
			}
			return dune.NewString(s), nil
		},
	},
	{
		Name:      "String.prototype.replace",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l < 2 || l > 3 {
				return dune.NullValue, fmt.Errorf("expected 2 or 3 arguments, got %d", len(args))
			}

			oldStr := args[0].String()
			newStr := args[1].String()

			times := -1
			if l > 2 {
				times = int(args[2].ToInt())
			}

			s := this.String()
			return dune.NewString(strings.Replace(s, oldStr, newStr, times)), nil
		},
	},
	{
		Name:      "String.prototype.split",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			sep := args[0].String()

			s := this.String()

			parts := Split(s, sep)
			res := make([]dune.Value, len(parts))

			for i, v := range parts {
				res[i] = dune.NewString(v)
			}
			return dune.NewArrayValues(res), nil
		},
	},
	{
		Name:      "String.prototype.splitEx",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			sep := args[0].String()

			s := this.String()

			parts := strings.Split(s, sep)
			res := make([]dune.Value, len(parts))

			for i, v := range parts {
				res[i] = dune.NewString(v)
			}
			return dune.NewArrayValues(res), nil
		},
	},
	{
		Name:      "String.prototype.trim",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			var cutset string
			switch len(args) {
			case 0:
				cutset = " \t\r\n"
			case 1:
				cutset = args[0].String()
			default:
				return dune.NullValue, fmt.Errorf("expected 0 or 1 arguments, got %d", len(args))
			}
			s := this.String()
			return dune.NewString(strings.Trim(s, cutset)), nil
		},
	},
	{
		Name:      "String.prototype.trimLeft",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			var cutset string
			switch len(args) {
			case 0:
				cutset = " \t\r\n"
			case 1:
				cutset = args[0].String()
			default:
				return dune.NullValue, fmt.Errorf("expected 0 or 1 arguments, got %d", len(args))
			}
			s := this.String()
			return dune.NewString(strings.TrimLeft(s, cutset)), nil
		},
	},
	{
		Name:      "String.prototype.trimRight",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			var cutset string
			switch len(args) {
			case 0:
				cutset = " \t\r\n"
			case 1:
				cutset = args[0].String()
			default:
				return dune.NullValue, fmt.Errorf("expected 0 or 1 arguments, got %d", len(args))
			}
			s := this.String()
			return dune.NewString(strings.TrimRight(s, cutset)), nil
		},
	},
	{
		Name:      "String.prototype.trimPrefix",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}
			s := this.String()
			prefix := args[0].String()
			s = strings.TrimPrefix(s, prefix)
			return dune.NewString(s), nil
		},
	},
	{
		Name:      "String.prototype.trimSuffix",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}
			s := this.String()
			prefix := args[0].String()
			s = strings.TrimSuffix(s, prefix)
			return dune.NewString(s), nil
		},
	},
	{
		Name:      "String.prototype.substring",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			s := this.String()

			switch len(args) {
			case 1:
				v1 := args[0]
				if v1.Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected int, got %s", v1.Type)
				}
				a := int(v1.ToInt())
				return dune.NewString(s[a:]), nil
			case 2:
				v1 := args[0]
				if v1.Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected int, got %s", v1.Type)
				}
				v2 := args[1]
				if v2.Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected int, got %s", v2.Type)
				}
				l := len(s)
				a := int(v1.ToInt())
				b := int(v2.ToInt())
				if a < 0 || a > l {
					return dune.NullValue, fmt.Errorf("start out of range")
				}
				if b < a || b > l {
					return dune.NullValue, fmt.Errorf("end out of range")
				}
				return dune.NewString(s[a:b]), nil
			}

			return dune.NullValue, fmt.Errorf("expected 1 or 2 parameters")
		},
	},
	{
		Name:      "String.prototype.runeSubstring",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			s := this.String()

			switch len(args) {
			case 1:
				v1 := args[0]
				if v1.Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected int, got %s", v1.Type)
				}
				a := int(v1.ToInt())
				return dune.NewString(substring(s, a, -1)), nil
			case 2:
				v1 := args[0]
				if v1.Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected int, got %s", v1.Type)
				}
				v2 := args[1]
				if v2.Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected int, got %s", v2.Type)
				}
				l := len(s)
				a := int(v1.ToInt())
				b := int(v2.ToInt())
				if a < 0 || a > l {
					return dune.NullValue, fmt.Errorf("start out of range")
				}
				if b < a || b > l {
					return dune.NullValue, fmt.Errorf("end out of range")
				}
				return dune.NewString(substring(s, a, b)), nil
			}

			return dune.NullValue, fmt.Errorf("expected 1 or 2 parameters")
		},
	},
	{
		Name:      "String.prototype.take",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be int, got %s", args[0].TypeName())
			}

			s := this.String()
			i := int(args[0].ToInt())

			if len(s) > i {
				s = s[:i]
			}
			return dune.NewString(s), nil
		},
	},
	{
		Name:      "String.prototype.hasPrefix",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}
			v := args[0].String()
			s := this.String()
			return dune.NewBool(strings.HasPrefix(s, v)), nil
		},
	},
	{
		Name:      "String.prototype.hasSuffix",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}
			v := args[0].String()
			s := this.String()
			return dune.NewBool(strings.HasSuffix(s, v)), nil
		},
	},
	{
		Name:      "String.prototype.indexOf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			ln := len(args)

			if ln > 0 {
				if args[0].Type != dune.String {
					return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
				}
			}

			if ln > 1 {
				if args[1].Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected arg 2 to be int, got %s", args[1].TypeName())
				}
			}

			sep := args[0].String()
			s := this.String()

			var i int
			if len(args) > 1 {
				i = int(args[1].ToInt())
				if i > len(s) {
					return dune.NullValue, fmt.Errorf("index out of range")
				}
				s = s[i:]
			}
			return dune.NewInt(strings.Index(s, sep) + i), nil
		},
	},
	{
		Name:      "String.prototype.lastIndexOf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			ln := len(args)

			if ln > 0 {
				if args[0].Type != dune.String {
					return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
				}
			}

			if ln > 1 {
				if args[1].Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected arg 2 to be int, got %s", args[1].TypeName())
				}
			}

			sep := args[0].String()
			s := this.String()

			if len(args) > 1 {
				i := int(args[1].ToInt())
				if i > len(s) {
					return dune.NullValue, fmt.Errorf("index out of range")
				}
				s = s[i:]
			}
			return dune.NewInt(strings.LastIndex(s, sep)), nil
		},
	},
	{
		Name:      "String.prototype.contains",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}

			sep := args[0].String()
			s := this.String()
			return dune.NewBool(strings.Contains(s, sep)), nil
		},
	},
	{
		Name:      "String.prototype.rightPad",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}
			if args[1].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("expected arg 2 to be int, got %s", args[1].TypeName())
			}
			pad := args[0].String()
			if len(pad) != 1 {
				return dune.NullValue, fmt.Errorf("invalid pad size. Must be one character")
			}
			total := int(args[1].ToInt())
			s := this.String()
			return dune.NewString(rightPad(s, rune(pad[0]), total)), nil
		},
	},
	{
		Name:      "String.prototype.leftPad",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}
			if args[1].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("expected arg 2 to be int, got %s", args[1].TypeName())
			}

			pad := args[0].String()
			if len(pad) != 1 {
				return dune.NullValue, fmt.Errorf("invalid pad size. Must be one character")
			}
			total := int(args[1].ToInt())
			s := this.String()
			return dune.NewString(leftPad(s, rune(pad[0]), total)), nil
		},
	},
	{
		Name:      "String.prototype.equalFold",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be string, got %s", args[0].TypeName())
			}
			eq := strings.EqualFold(this.String(), args[0].String())
			return dune.NewBool(eq), nil
		},
	},
}

func substring(s string, start int, end int) string {
	start_str_idx := 0
	i := 0
	for j := range s {
		if i == start {
			start_str_idx = j
		}
		if i == end {
			return s[start_str_idx:j]
		}
		i++
	}
	return s[start_str_idx:]
}

// IsNumeric returns true if s contains only digits
func IsNumeric(s string) bool {
	for _, r := range s {
		if !IsDecimal(r) {
			return false
		}
	}
	return true
}

// IsDecimal returns true if r is a digit
func IsDecimal(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsIdent returns if s is a valid identifier.
func IsIdent(s string) bool {
	for i, c := range s {
		if !isStrIdent(c, i) {
			return false
		}
	}
	return true
}

func isStrIdent(ch rune, pos int) bool {
	return ch == '_' ||
		'A' <= ch && ch <= 'Z' ||
		'a' <= ch && ch <= 'z' ||
		IsDecimal(ch) && pos > 0
}

func IsAlphanumericIdent(s string) bool {
	for i, c := range s {
		if c == '_' {
			continue
		}
		if !isAlphanumeric(c, i) {
			return false
		}
	}
	return true
}

func IsAlphanumeric(s string) bool {
	for _, c := range s {
		if !isAlphanumeric(c, 1) {
			return false
		}
	}
	return true
}

func isAlphanumeric(ch rune, pos int) bool {
	return 'A' <= ch && ch <= 'Z' ||
		'a' <= ch && ch <= 'z' ||
		IsDecimal(ch) && pos > 0
}

func Split(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		if p != "" {
			// only append non empty values
			result = append(result, p)
		}
	}
	return result
}

func rightPad(s string, pad rune, total int) string {
	l := total - utf8.RuneCountInString(s)
	if l < 1 {
		return s
	}
	return s + strings.Repeat(string(pad), l)
}

func leftPad(s string, pad rune, total int) string {
	l := total - utf8.RuneCountInString(s)
	if l < 1 {
		return s
	}
	return strings.Repeat(string(pad), l) + s
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// https://stackoverflow.com/a/31832326/4264
func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
