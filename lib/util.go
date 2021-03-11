package lib

import (
	"fmt"
	"strconv"

	"github.com/dunelang/dune"
)

// validate the number of args ant type
func ValidateOrNilArgs(args []dune.Value, t ...interface{}) error {
	exp := len(t)
	got := len(args)
	if exp != got {
		return fmt.Errorf("expected %d arguments, got %d", exp, got)
	}

	for i, v := range t {
		a := args[i]

		switch a.Type {
		case dune.Null, dune.Undefined:
			continue
		}

		if v != nil && !validateType(v.(dune.Type), a.Type) {
			return fmt.Errorf("expected argument %d to be %v, got %s", i, v, a.TypeName())
		}
	}

	return nil
}

// validate the number of args ant type
func ValidateArgs(args []dune.Value, t ...interface{}) error {
	exp := len(t)
	got := len(args)
	if exp != got {
		return fmt.Errorf("expected %d arguments, got %d", exp, got)
	}

	for i, v := range t {
		a := args[i]
		if v != nil && !validateType(v.(dune.Type), a.Type) {
			return fmt.Errorf("expected argument %d to be %v, got %s", i, v, a.TypeName())
		}
	}
	return nil
}

// validate that if present, args are of type t
func ValidateOptionalArgs(args []dune.Value, t ...dune.Type) error {
	exp := len(t)
	got := len(args)
	if got > exp {
		return fmt.Errorf("expected %d arguments max, got %d", exp, got)
	}

	for i, v := range args {
		a := t[i]
		t := v.Type
		if t == dune.Undefined || t == dune.Null {
			continue
		}
		if !validateType(t, a) {
			return fmt.Errorf("expected argument %d to be %v, got %s", i, a, v.TypeName())
		}
	}

	return nil
}

func validateType(a, b dune.Type) bool {
	if a == b {
		return true
	}

	switch a {
	case dune.Bytes:
		switch b {
		case dune.Bytes, dune.String, dune.Array:
			return true
		}

	case dune.String:
		switch b {
		case dune.String, dune.Bytes, dune.Rune:
			return true
		}

	case dune.Int:
		switch b {
		case dune.Int, dune.Float, dune.Bool:
			return true
		}

	case dune.Float:
		switch b {
		case dune.Int, dune.Float:
			return true
		}
	}

	return false
}

// func validateType(v, t dune.Type) bool {
// 	if v == t {
// 		return true
// 	}

// 	if v == dune.Bytes && t == dune.String {
// 		return true
// 	}
// 	if v == dune.String && t == dune.Bytes {
// 		return true
// 	}

// 	if v == dune.Int && t == dune.Float {
// 		return true
// 	}
// 	if v == dune.Float && t == dune.Int {
// 		return true
// 	}

// 	if v == dune.Int && t == dune.Bool {
// 		return true
// 	}

// 	return false
// }

func ValidateArgRange(args []dune.Value, counts ...int) error {
	l := len(args)
	for _, v := range counts {
		if l == v {
			return nil
		}
	}

	var s string
	j := len(counts) - 1
	for i, v := range counts {
		if i == j {
			s += " or "
		} else if i > 0 {
			s += ", "
		}
		s += strconv.Itoa(v)
	}

	return fmt.Errorf("expected %s arguments, got %d", s, l)
}

func runFuncOrClosure(vm *dune.VM, fn dune.Value, args ...dune.Value) error {
	switch fn.Type {
	case dune.Func:
		_, err := vm.RunFuncIndex(fn.ToFunction(), args...)
		return err

	case dune.Object:
		c, ok := fn.ToObject().(*dune.Closure)
		if !ok {
			return fmt.Errorf("%v is not a function", fn.TypeName())
		}
		_, err := vm.RunClosure(c, args...)
		return err

	default:
		return fmt.Errorf("%v is not a function", fn.TypeName())
	}
}
