package lib

import (
	"fmt"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Async, `
	
declare function go(f: Function): void


	`)
}

var Async = []dune.NativeFunction{
	{
		Name:        "go",
		Arguments:   1,
		Permissions: []string{"async"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return launchGoroutine(args, vm, nil)
		},
	},
}

func launchGoroutine(args []dune.Value, vm *dune.VM, t *waitGroup) (dune.Value, error) {
	m, err := cloneForAsync(vm)
	if err != nil {
		return dune.NullValue, err
	}

	a := args[0]
	switch a.Type {
	case dune.Func:
		if t != nil {
			t.w.Add(1)
		}
		go func() {
			_, err := m.RunFuncIndex(a.ToFunction())
			if err != nil {
				fmt.Fprintln(vm.GetStderr(), err)
			}
			if t != nil {
				t.w.Done()
				if t.limit != nil {
					<-t.limit
				}
			}
		}()

	case dune.Object:
		c, ok := a.ToObjectOrNil().(*dune.Closure)
		if !ok {
			return dune.NullValue, fmt.Errorf("%v is not a function", a.TypeName())
		}

		if t != nil {
			t.w.Add(1)
		}

		go func() {
			_, err := m.RunClosure(c)
			if err != nil {
				fmt.Fprintln(vm.GetStderr(), err)
			}
			if t != nil {
				t.w.Done()
				if t.limit != nil {
					<-t.limit
				}
			}
		}()

	default:
		return dune.NullValue, fmt.Errorf("%v is not a function", a.TypeName())
	}

	return dune.NullValue, nil
}

func runAsyncFuncOrClosure(vm *dune.VM, fn dune.Value, args ...dune.Value) error {
	m, err := cloneForAsync(vm)
	if err != nil {
		return err
	}

	switch fn.Type {
	case dune.Func:
		_, err := m.RunFuncIndex(fn.ToFunction(), args...)
		return err

	case dune.Object:
		c, ok := fn.ToObject().(*dune.Closure)
		if !ok {
			return fmt.Errorf("%v is not a function", fn.TypeName())
		}
		_, err := m.RunClosure(c, args...)
		return err

	default:
		return fmt.Errorf("%v is not a function", fn.TypeName())
	}
}

func cloneForAsync(vm *dune.VM) (*dune.VM, error) {
	m := dune.NewInitializedVM(vm.Program, vm.Globals())
	m.MaxAllocations = vm.MaxAllocations
	m.MaxFrames = vm.MaxFrames
	m.MaxSteps = vm.MaxSteps
	m.FileSystem = vm.FileSystem
	m.Context = vm.Context
	m.Language = vm.Language
	m.Localizer = vm.Localizer
	m.Location = vm.Location
	m.Now = vm.Now
	m.Stdin = vm.Stdin
	m.Stdout = vm.Stdout
	m.Stderr = vm.Stderr

	if err := m.AddSteps(vm.Steps()); err != nil {
		return nil, err
	}

	return m, nil
}
