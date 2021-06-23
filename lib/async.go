package lib

import (
	"fmt"
	"time"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Async, `
	
declare function go(f: Function): void

declare namespace async {
    export function withDeadline(d: time.Duration, fn: (dl: Deadline) => void): void

	export interface Deadline {
		extend(d: time.Duration): void
		cancel(): void
	}
}


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
	{
		Name:        "async.withDeadline",
		Arguments:   2,
		Permissions: []string{"async"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			d, err := ToDuration(args[0])
			if err != nil {
				return dune.NullValue, err
			}

			v := args[1]
			switch v.Type {
			case dune.Func:
			case dune.Object:
			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", v.TypeName())
			}

			WithDeadline(d, func(dl *Deadline) {
				obj := dune.NewObject(dl)
				if err := runAsyncFuncOrClosure(vm, v, obj); err != nil {
					fmt.Fprintln(vm.GetStderr(), err)
				}
			})

			return dune.NullValue, err
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
		if c, ok := a.ToObjectOrNil().(*dune.Closure); ok {
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
		} else if c, ok := a.ToObjectOrNil().(*dune.Method); ok {
			if t != nil {
				t.w.Add(1)
			}
			go func() {
				_, err := m.RunMethod(c)
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
		} else {
			return dune.NullValue, fmt.Errorf("%v is not a function", a.TypeName())
		}

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
		if c, ok := fn.ToObject().(*dune.Closure); ok {
			_, err := m.RunClosure(c, args...)
			return err
		}

		if c, ok := fn.ToObject().(*dune.Method); ok {
			_, err := m.RunMethod(c, args...)
			return err
		}

		return fmt.Errorf("%v is not a function", fn.TypeName())

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

type Deadline struct {
	limit  time.Time
	ticker *time.Ticker
	done   chan bool
}

func (dl *Deadline) Type() string {
	return "async.Deadline"
}

func (dl *Deadline) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "extend":
		return dl.extend
	case "cancel":
		return dl.cancel
	}
	return nil
}

func (dl *Deadline) extend(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	dd, err := ToDuration(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	dl.Extend(dd)
	return dune.NullValue, nil
}

func (dl *Deadline) cancel(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	dl.Cancel()
	return dune.NullValue, nil
}

func (dl *Deadline) Extend(d time.Duration) {
	dl.limit = dl.limit.Add(d)
}

func (dl *Deadline) Cancel() {
	dl.done <- true
}

func WithDeadline(d time.Duration, fn func(dl *Deadline)) {
	ticker := time.NewTicker(d)

	dl := &Deadline{
		limit:  time.Now().Add(d),
		ticker: ticker,
		done:   make(chan bool),
	}

	go func() {
		fn(dl)
		dl.done <- true
	}()

	for {
		select {
		case <-dl.done:
			return
		case t := <-ticker.C:
			if !t.Before(dl.limit) {
				ticker.Stop()
				return
			}
		}
	}
}
