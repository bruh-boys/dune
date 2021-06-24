package lib

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(libArray, `
interface Array<T> {
    [n: number]: T
    slice(start?: number, count?: number): Array<T>
    range(start?: number, end?: number): Array<T>
    append(v: T[]): T[]
    push(...v: T[]): void
    pushRange(v: T[]): void
    copyAt(i: number, v: T[]): void
    length: number
    insertAt(i: number, v: T): void
    removeAt(i: number): void
    removeAt(from: number, to: number): void
    indexOf(v: T): number
    join(sep?: string): T
    sort(comprarer: (a: T, b: T) => boolean): void
    equals(other: Array<T>): boolean;
    any(func: (t: T) => any): boolean;
    all(func: (t: T) => any): boolean;
    contains<T>(t: T): boolean;
    remove<T>(t: T): void;
    first(): T;
	last(): T;
	clear(): void
    first(func?: (t: T) => any): T;
    last(func?: (t: T) => any): T;
    firstIndex(func: (t: T) => any): number;
    select<K>(func: (t: T) => K): Array<K>;
    selectMany<K>(func: (t: T) => K): K;
    distinct<K>(func?: (t: any) => K): Array<K>;
    where(func: (t: T) => any): Array<T>;
    groupBy(func: (t: T) => any): KeyIndexer<T[]>;
    sum<K extends number>(): number;
    sum<K extends number>(func: (t: T) => K): number;
    min(func: (t: T) => number): number;
    max(func: (t: T) => number): number;
    count(func: (t: T) => any): number;
}

declare namespace array {
    /**
     * Create a new array with size.
     */
    export function make<T>(size: number, capacity?: number): Array<T>

    /**
     * Create a new array of bytes with size.
     */
    export function bytes(size: number, capacity?: number): byte[]
}
	`)
}

var libArray = []dune.NativeFunction{
	{
		Name:      "array.make",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			ln := len(args)

			if ln > 0 {
				if args[0].Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected arg 1 to be int, got %s", args[0].TypeName())
				}
			}

			if ln > 1 {
				if args[1].Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected arg 2 to be int, got %s", args[1].TypeName())
				}
			}
			var size, cap int64
			switch len(args) {
			case 1:
				size = args[0].ToInt()
				return dune.NewArray(int(size)), nil

			case 2:
				size = args[0].ToInt()
				cap = args[1].ToInt()
				a := make([]dune.Value, size, cap)
				return dune.NewArrayValues(a), nil

			default:
				return dune.NullValue, fmt.Errorf("expected 1 or 2 arguments, got %d", len(args))
			}
		},
	},
	{
		Name:      "array.bytes",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			ln := len(args)

			if ln > 0 {
				if args[0].Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected arg 1 to be int, got %s", args[0].TypeName())
				}
			}

			if ln > 1 {
				if args[1].Type != dune.Int {
					return dune.NullValue, fmt.Errorf("expected arg 2 to be int, got %s", args[1].TypeName())
				}
			}

			var size, cap int64
			switch len(args) {
			case 1:
				size = args[0].ToInt()
				return dune.NewBytes(make([]byte, size)), nil

			case 2:
				size = args[0].ToInt()
				cap = args[1].ToInt()
				return dune.NewBytes(make([]byte, size, cap)), nil

			default:
				return dune.NullValue, fmt.Errorf("expected 1 or 2 arguments, got %d", len(args))
			}
		},
	},
	{
		Name:      "Array.prototype.copyAt",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be int, got %s", args[0].TypeName())
			}
			if args[1].Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected arg 2 to be array, got %s", args[1].TypeName())
			}

			a := this.ToArray()
			start := int(args[0].ToInt())
			b := args[1].ToArray()

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
		Name:      "Array.prototype.any",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			for _, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				var any bool

				switch r.Type {
				case dune.Int:
					any = r.ToInt() != 0

				case dune.Float:
					any = r.ToFloat() != 0

				case dune.Bool:
					any = r.ToBool()

				case dune.Null, dune.Undefined:
					any = false

				default:
					any = true
				}

				if any {
					return dune.TrueValue, nil
				}
			}

			return dune.FalseValue, nil
		},
	},

	{
		Name:      "Array.prototype.all",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			if len(a) == 0 {
				return dune.FalseValue, nil
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			for _, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				switch r.Type {
				case dune.Int:
					if r.ToInt() == 0 {
						return dune.FalseValue, nil
					}

				case dune.Float:
					if r.ToFloat() == 0 {
						return dune.FalseValue, nil
					}

				case dune.Bool:
					if !r.ToBool() {
						return dune.FalseValue, nil
					}

				case dune.Null, dune.Undefined:
					return dune.FalseValue, nil

				default:
					return dune.FalseValue, nil
				}
			}

			return dune.TrueValue, nil
		},
	},

	{
		Name:      "Array.prototype.contains",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}
			b := args[0]

			for _, v := range a {
				if v.Equals(b) {
					return dune.TrueValue, nil
				}
			}

			return dune.FalseValue, nil
		},
	},

	{
		Name:      "Array.prototype.remove",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := this.ToArray()
			b := args[0]

			for i, v := range a {
				if v.Equals(b) {
					obj := this.ToArrayObject()
					a := obj.Array
					copy(a[i:], a[i+1:])
					a[len(a)-1] = dune.NullValue
					obj.Array = a[:len(a)-1]
					break
				}
			}

			return dune.NullValue, nil
		},
	},

	{
		Name:      "Array.prototype.firstIndex",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			for i, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				var found bool

				switch r.Type {
				case dune.Int:
					found = r.ToInt() != 0

				case dune.Float:
					found = r.ToFloat() != 0

				case dune.Bool:
					found = r.ToBool()

				case dune.Null, dune.Undefined:
					found = false

				default:
					found = true
				}

				if found {
					return dune.NewInt(i), nil
				}
			}

			return dune.NewInt(-1), nil
		},
	},

	{
		Name:      "Array.prototype.first",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			items, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			if len(args) == 0 {
				if len(items) > 0 {
					return items[0], nil
				} else {
					return dune.NullValue, nil
				}
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			for i, item := range items {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, item)
				} else {
					r, err = vm.RunClosure(closure, item)
				}
				if err != nil {
					return dune.NullValue, err
				}

				var found bool

				switch r.Type {
				case dune.Int:
					found = r.ToInt() != 0

				case dune.Float:
					found = r.ToFloat() != 0

				case dune.Bool:
					found = r.ToBool()

				case dune.Null, dune.Undefined:
					found = false

				default:
					// any non null value is considered a match
					found = true
				}

				if found {
					return items[i], nil
				}
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:      "Array.prototype.last",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			if len(args) == 0 {
				if l := len(a); l > 0 {
					return a[l-1], nil
				} else {
					return dune.NullValue, nil
				}
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			last := dune.NullValue

			for i, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				var matches bool

				switch r.Type {
				case dune.Int:
					matches = r.ToInt() != 0

				case dune.Float:
					matches = r.ToFloat() != 0

				case dune.Bool:
					matches = r.ToBool()

				case dune.Null, dune.Undefined:
					matches = false

				default:
					matches = true
				}

				if matches {
					last = a[i]
				}
			}

			return last, nil
		},
	},

	{
		Name:      "Array.prototype.where",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			filtered := make([]dune.Value, 0)

			for _, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				var matches bool

				switch r.Type {
				case dune.Int:
					matches = r.ToInt() != 0

				case dune.Float:
					matches = r.ToFloat() != 0

				case dune.Bool:
					matches = r.ToBool()

				case dune.Null, dune.Undefined:
					matches = false

				default:
					matches = true
				}

				if matches {
					filtered = append(filtered, v)
				}
			}

			return dune.NewArrayValues(filtered), nil
		},
	},

	{
		Name:      "Array.prototype.sum",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			if len(args) == 0 {
				var ret float64
				var anyFloat bool

				for i, v := range a {
					switch v.Type {
					case dune.Int:
						ret += v.ToFloat()

					case dune.Float:
						ret += v.ToFloat()
						anyFloat = true

					default:
						return dune.NullValue, fmt.Errorf("invalid array value at index %d", i)
					}
				}

				if anyFloat {
					return dune.NewFloat(ret), nil
				}
				return dune.NewInt(int(ret)), nil
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			var ret float64
			var anyFloat bool

			for i, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				switch r.Type {
				case dune.Int:
					ret += r.ToFloat()

				case dune.Float:
					ret += r.ToFloat()
					anyFloat = true

				default:
					return dune.NullValue, fmt.Errorf("invalid array value at index %d", i)
				}
			}

			if anyFloat {
				return dune.NewFloat(ret), nil
			}
			return dune.NewInt(int(ret)), nil
		},
	},

	{
		Name:      "Array.prototype.min",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			if len(a) == 0 {
				return dune.UndefinedValue, err
			}

			var min, tmp float64
			var anyFloat bool

			if len(args) == 0 {
				for i, v := range a {
					switch v.Type {
					case dune.Int:
						tmp = v.ToFloat()

					case dune.Float:
						tmp = v.ToFloat()
						anyFloat = true

					default:
						return dune.NullValue, fmt.Errorf("invalid array value at index %d", i)
					}

					if i == 0 || tmp < min {
						min = tmp
					}
				}

				if anyFloat {
					return dune.NewFloat(min), nil
				}
				return dune.NewInt(int(min)), nil
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			for i, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				switch r.Type {
				case dune.Int:
					tmp = r.ToFloat()

				case dune.Float:
					tmp = r.ToFloat()
					anyFloat = true

				default:
					return dune.NullValue, fmt.Errorf("invalid array value at index %d", i)
				}

				if i == 0 || tmp < min {
					min = tmp
				}
			}

			if anyFloat {
				return dune.NewFloat(min), nil
			}
			return dune.NewInt(int(min)), nil
		},
	},

	{
		Name:      "Array.prototype.max",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			if len(a) == 0 {
				return dune.UndefinedValue, err
			}

			var max, tmp float64
			var anyFloat bool

			if len(args) == 0 {
				for i, v := range a {
					switch v.Type {
					case dune.Int:
						tmp = v.ToFloat()

					case dune.Float:
						tmp = v.ToFloat()
						anyFloat = true

					default:
						return dune.NullValue, fmt.Errorf("invalid array value at index %d", i)
					}

					if i == 0 || tmp > max {
						max = tmp
					}
				}

				if anyFloat {
					return dune.NewFloat(max), nil
				}
				return dune.NewInt(int(max)), nil
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			for i, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				switch r.Type {
				case dune.Int:
					tmp = r.ToFloat()

				case dune.Float:
					tmp = r.ToFloat()
					anyFloat = true

				default:
					return dune.NullValue, fmt.Errorf("invalid array value at index %d", i)
				}

				if i == 0 || tmp > max {
					max = tmp
				}
			}

			if anyFloat {
				return dune.NewFloat(max), nil
			}
			return dune.NewInt(int(max)), nil
		},
	},

	{
		Name:      "Array.prototype.count",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			var c int

			for _, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				var matches bool

				switch r.Type {
				case dune.Int:
					matches = r.ToInt() != 0

				case dune.Float:
					matches = r.ToFloat() != 0

				case dune.Bool:
					matches = r.ToBool()

				case dune.Null, dune.Undefined:
					matches = false

				default:
					matches = true
				}

				if matches {
					c++
				}
			}

			return dune.NewInt(c), nil
		},
	},

	{
		Name:      "Array.prototype.select",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			l := len(a)

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			items := make([]dune.Value, l)

			for i, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				items[i] = r
			}

			return dune.NewArrayValues(items), nil
		},
	},

	{
		Name:      "Array.prototype.selectMany",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			var items []dune.Value

			for i, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}
				if r.Type != dune.Array {
					return dune.NullValue, fmt.Errorf("the element in index %d is not an array", i)
				}
				items = append(items, r.ToArray()...)
			}

			return dune.NewArrayValues(items), nil
		},
	},

	{
		Name:      "Array.prototype.distinct",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			thisItems, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			items := make([]dune.Value, 0)

			if len(args) == 0 {
				for _, v := range thisItems {
					exists := false
					for _, w := range items {
						if v == w {
							exists = true
							break
						}
					}
					if !exists {
						items = append(items, v)
					}
				}
				return dune.NewArrayValues(items), nil
			}

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			for _, v := range thisItems {
				var vKey dune.Value
				if funcIndex != -1 {
					vKey, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					vKey, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				exists := false
				for _, w := range items {
					var existingKey dune.Value
					if funcIndex != -1 {
						existingKey, err = vm.RunFuncIndex(funcIndex, w)
					} else {
						existingKey, err = vm.RunClosure(closure, w)
					}
					if err != nil {
						return dune.NullValue, err
					}

					if existingKey == vKey {
						exists = true
						break
					}
				}

				if !exists {
					items = append(items, v)
				}
			}

			return dune.NewArrayValues(items), nil
		},
	},

	{
		Name:      "Array.prototype.groupBy",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}
			groups := make(map[dune.Value]dune.Value)

			funcIndex := -1
			var closure *dune.Closure

			b := args[0]
			switch b.Type {
			case dune.Func:
				funcIndex = b.ToFunction()

			case dune.Object:
				c, ok := b.ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				closure = c

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}

			for _, v := range a {
				var r dune.Value
				if funcIndex != -1 {
					r, err = vm.RunFuncIndex(funcIndex, v)
				} else {
					r, err = vm.RunClosure(closure, v)
				}
				if err != nil {
					return dune.NullValue, err
				}

				// no permitir agrupar por undefined
				if r.Type == dune.Undefined {
					r = dune.NullValue
				}

				tmp, ok := groups[r]

				if ok {
					g := tmp.ToArrayObject().Array
					g = append(g, v)
					groups[r] = dune.NewArrayValues(g)
				} else {
					g := []dune.Value{v}
					groups[r] = dune.NewArrayValues(g)
				}
			}

			return dune.NewMapValues(groups), nil
		},
	},

	{
		Name:      "Array.prototype.equals",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if len(args) == 0 || args[0].Type != dune.Array {
				return dune.FalseValue, nil
			}

			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}
			b := args[0].ToArray()

			if len(a) != len(b) {
				return dune.FalseValue, nil
			}

			for i := range a {
				if !a[i].Equals(b[i]) {
					return dune.FalseValue, nil
				}
			}

			return dune.TrueValue, nil
		},
	},

	{
		Name:      "Array.prototype.sort",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			switch this.Type {
			case dune.Null:
				return args[0], nil
			case dune.Array, dune.Bytes:
			default:
				return dune.NullValue, fmt.Errorf("expected array, called on %s", this.TypeName())
			}

			a := this.ToArray()

			b := args[0]
			switch b.Type {
			case dune.Func:
				c := &comparer{items: a, compFunc: b.ToFunction(), vm: vm}
				sort.Sort(c)
				return dune.NullValue, nil

			case dune.Object:
				cl, ok := b.ToObjectOrNil().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
				}
				c := &closureComparer{items: a, comp: cl, vm: vm}
				sort.Sort(c)
				return dune.NullValue, nil

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", b.TypeName())
			}
		},
	},
	{
		Name:      "Array.prototype.append",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			switch this.Type {
			case dune.Array, dune.Bytes:
			default:
				return dune.NullValue, fmt.Errorf("expected array, called on %s", this.TypeName())
			}

			a := this.ToArray()

			b := args[0]
			switch b.Type {
			case dune.Null:
				return this, nil
			case dune.Array, dune.Bytes:
			default:
				return dune.NullValue, fmt.Errorf("expected array, called on %s", b.TypeName())
			}

			c := append(a, b.ToArray()...)

			return dune.NewArrayValues(c), nil
		},
	},
	{
		Name:      "Array.prototype.pushRange",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			switch this.Type {
			case dune.Array, dune.Bytes:
			default:
				return dune.NullValue, fmt.Errorf("expected array, called on %s", this.TypeName())
			}

			b := args[0]
			switch b.Type {
			case dune.Null:
			case dune.Array, dune.Bytes:
				a := this.ToArrayObject()
				items := b.ToArray()
				a.Array = append(a.Array, items...)

				if vm.MaxAllocations > 0 {
					var allocs int
					for _, v := range items {
						allocs += v.Size()
					}
					if err := vm.AddAllocations(allocs); err != nil {
						return dune.NullValue, err
					}
				}

			default:
				return dune.NullValue, fmt.Errorf("expected array, called on %s", b.TypeName())
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:      "Array.prototype.push",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			switch this.Type {
			case dune.Array:
				a := this.ToArrayObject()
				a.Array = append(a.Array, args...)
				if vm.MaxAllocations > 0 {
					var allocs int
					for _, v := range a.Array {
						allocs += v.Size()
					}
					if err := vm.AddAllocations(allocs); err != nil {
						return dune.NullValue, err
					}
				}

			default:
				return dune.NullValue, fmt.Errorf("expected array, got %s", this.TypeName())
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:      "Array.prototype.insertAt",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected string array, got %s", this.TypeName())
			}
			if args[0].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be int, got %s", this.TypeName())
			}
			obj := this.ToArrayObject()
			i := int(args[0].ToInt())

			a := obj.Array
			a = append(a, dune.NullValue)
			copy(a[i+1:], a[i:])
			a[i] = args[1]
			obj.Array = a

			return dune.NullValue, nil
		},
	},
	{
		Name:      "Array.prototype.removeAt",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.Int {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be int, got %s", args[0].TypeName())
			}

			if this.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected string array, got %s", this.TypeName())
			}
			obj := this.ToArrayObject()
			i := int(args[0].ToInt())

			a := obj.Array
			copy(a[i:], a[i+1:])
			a[len(a)-1] = dune.NullValue
			obj.Array = a[:len(a)-1]

			return dune.NullValue, nil
		},
	},
	{
		Name:      "Array.prototype.removeRange",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected string array, got %s", this.TypeName())
			}
			obj := this.ToArrayObject()
			i := int(args[0].ToInt())
			j := int(args[0].ToInt())

			a := obj.Array
			copy(a[i:], a[j:])
			for k, n := len(a)-j+i, len(a); k < n; k++ {
				a[k] = dune.NullValue
			}
			obj.Array = a[:len(a)-j+i]

			return dune.NullValue, nil
		},
	},
	{
		Name:      "Array.prototype.indexOf",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected string array, got %s", this.TypeName())
			}
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}
			v := args[0]

			for i, j := range a {
				if j.Equals(v) {
					return dune.NewInt(i), nil
				}
			}

			return dune.NewInt(-1), nil
		},
	},
	{
		Name: "Array.prototype.reverse",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected string array, got %s", this.TypeName())
			}
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}
			l := len(a) - 1

			for i, k := 0, l/2; i <= k; i++ {
				a[i], a[l-i] = a[l-i], a[i]
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:      "Array.prototype.join",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected string array, got %s", this.TypeName())
			}
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}

			var sep string
			switch len(args) {
			case 0:
				sep = ""

			case 1:
				if args[0].Type != dune.String {
					return dune.NullValue, fmt.Errorf("expected string, got %s", args[0].TypeName())
				}
				sep = args[0].String()

			default:
				return dune.NullValue, fmt.Errorf("expected 0 or 1 args, got %d", len(args))
			}

			s := make([]string, len(a))
			for i, v := range a {
				switch v.Type {
				case dune.String, dune.Rune, dune.Int, dune.Float,
					dune.Bool, dune.Null, dune.Undefined:
					s[i] = v.String()

				default:
					return dune.NullValue, fmt.Errorf("invalid type at index %d, expected string, got %s", i, v.TypeName())
				}
			}
			r := strings.Join(s, sep)

			return dune.NewString(r), nil
		},
	},
	{
		Name:      "Array.prototype.slice",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected array, called on %s", this.TypeName())
			}
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}
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

			return dune.NewArrayValues(a), nil
		},
	},
	{
		Name:      "Array.prototype.range",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected array, called on %s", this.TypeName())
			}
			a, err := toArray(this)
			if err != nil {
				return dune.NullValue, err
			}
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

			return dune.NewArrayValues(a), nil
		},
	},
	{
		Name:      "Array.prototype.clear",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if this.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected array, called on %s", this.TypeName())
			}

			obj := this.ToArrayObject()
			obj.Array = obj.Array[:0]

			return dune.NullValue, nil
		},
	},
}

func toArray(v dune.Value) ([]dune.Value, error) {
	switch v.Type {
	case dune.Array:
		return v.ToArray(), nil

	case dune.Object:
		e, ok := v.ToObject().(dune.Enumerable)
		if !ok {
			return nil, fmt.Errorf("expected an enumerable, got %s", v.TypeName())
		}

		a, err := e.Values()
		if err != nil {
			return nil, fmt.Errorf("error enumerating values: %v", err)
		}

		return a, nil

	default:
		return nil, fmt.Errorf("expected an enumerable, got %s", v.TypeName())
	}
}

type comparer struct {
	items    []dune.Value
	compFunc int
	vm       *dune.VM
	err      error
}

func (c *comparer) Len() int {
	return len(c.items)
}
func (c *comparer) Swap(i, j int) {
	c.items[i], c.items[j] = c.items[j], c.items[i]
}
func (c *comparer) Less(i, j int) bool {
	if c.err != nil {
		return false
	}
	v, err := c.vm.RunFuncIndex(c.compFunc, c.items[i], c.items[j])
	if err != nil {
		c.err = err
		return false
	}

	if v.Type != dune.Bool {
		c.err = fmt.Errorf("the comparer function must return a boolean")
	}
	return v.ToBool()
}

type closureComparer struct {
	items []dune.Value
	comp  *dune.Closure
	vm    *dune.VM
	err   error
}

func (c *closureComparer) Len() int {
	return len(c.items)
}
func (c *closureComparer) Swap(i, j int) {
	c.items[i], c.items[j] = c.items[j], c.items[i]
}
func (c *closureComparer) Less(i, j int) bool {
	if c.err != nil {
		return false
	}
	v, err := c.vm.RunClosure(c.comp, c.items[i], c.items[j])
	if err != nil {
		c.err = err
		return false
	}

	if v.Type != dune.Bool {
		c.err = fmt.Errorf("the comparer function must return a boolean")
	}
	return v.ToBool()
}
