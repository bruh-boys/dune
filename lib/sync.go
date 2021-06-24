package lib

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Sync, `
	
declare namespace sync {
    export function newMutex(): Mutex
	export function newWaitGroup(concurrency?: number): WaitGroup
	
	export function execLocked(key: string, func: Function): any

    export interface WaitGroup {
        go(f: Function): void
        wait(): void
    }

    export interface Mutex {
        lock(): void
        unlock(): void
    }

    export function newChannel(buffer?: number): Channel

    export function select(channels: Channel[], defaultCase?: boolean): { index: number, value: any, receivedOK: boolean }

    export interface Channel {
        send(v: any): void
        receive(): any
        close(): void
    }

	export function withDeadline(d: time.Duration | number, fn: (dl: Deadline) => void): void

	export interface Deadline {
		extend(d: time.Duration): void
		cancel(): void
	}
}

	`)
}

var Sync = []dune.NativeFunction{
	{
		Name:        "sync.withDeadline",
		Arguments:   2,
		Permissions: []string{"sync"},
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

			err = WithDeadline(d, func(dl *Deadline) error {
				obj := dune.NewObject(dl)
				return runAsyncFuncOrClosure(vm, v, obj)
			})

			return dune.NullValue, err
		},
	},
	{
		Name:        "sync.newWaitGroup",
		Arguments:   -1,
		Permissions: []string{"sync"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}

			wg := &waitGroup{w: &sync.WaitGroup{}}

			if len(args) == 1 {
				concurrency := int(args[0].ToInt())
				wg.limit = make(chan bool, concurrency)
			}

			return dune.NewObject(wg), nil
		},
	},
	{
		Name:        "sync.newChannel",
		Arguments:   -1,
		Permissions: []string{"sync"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}

			var ch chan dune.Value
			var b int
			if len(args) > 0 {
				b = int(args[0].ToInt())
				ch = make(chan dune.Value, b)
			} else {
				ch = make(chan dune.Value)
			}

			c := &channel{buffer: b, c: ch}
			return dune.NewObject(c), nil
		},
	},
	{
		Name:        "sync.select",
		Arguments:   -1,
		Permissions: []string{"sync"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			argLen := len(args)
			if argLen == 0 || argLen > 2 {
				return dune.NullValue, fmt.Errorf("expected 1 or 2 args, got %d", argLen)
			}

			a := args[0]
			if a.Type != dune.Array {
				return dune.NullValue, fmt.Errorf("expected arg 1 to be an array of channels, got %s", a.TypeName())
			}

			chans := a.ToArray()
			l := len(chans)
			cases := make([]reflect.SelectCase, l)
			for i, c := range chans {
				ch := c.ToObjectOrNil().(*channel)
				if ch == nil {
					return dune.NullValue, fmt.Errorf("invalid channel at index %d", i)
				}
				cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch.c)}
			}

			if argLen == 2 {
				b := args[1]
				if b.Type != dune.Bool {
					return dune.NullValue, fmt.Errorf("expected arg 2 to be a bool, got %s", b.TypeName())
				}
				if b.ToBool() {
					cases = append(cases, reflect.SelectCase{Dir: reflect.SelectDefault})
				}
			}

			i, value, ok := reflect.Select(cases)

			m := make(map[dune.Value]dune.Value, 3)
			m[dune.NewString("index")] = dune.NewInt(i)

			// case default will send an invalid value and will panic if read
			if value.IsValid() {
				m[dune.NewString("value")] = value.Interface().(dune.Value)
			}

			m[dune.NewString("receivedOK")] = dune.NewBool(ok)

			return dune.NewMapValues(m), nil
		},
	},
	{
		Name:        "sync.newMutex",
		Arguments:   0,
		Permissions: []string{"sync"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			m := &mutex{mutex: &sync.Mutex{}}
			return dune.NewObject(m), nil
		},
	},

	{
		Name:        "sync.execLocked",
		Arguments:   -1,
		Permissions: []string{"sync"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l < 2 {
				return dune.NullValue, fmt.Errorf("expected at least 2 parameter, got %d", l)
			}

			keyVal := args[0]
			if keyVal.Type != dune.String {
				return dune.NullValue, fmt.Errorf("key must be a string, got %s", keyVal.TypeName())
			}

			m := globalKeyMutex.getMutex(keyVal.String())

			m.Lock()
			defer m.Unlock()

			var retVal dune.Value
			var err error

			switch args[1].Type {
			case dune.Func:
				f := int(args[1].ToFunction())
				retVal, err = vm.RunFuncIndex(f, args[2:]...)
			case dune.Object:
				closure, ok := args[1].ToObject().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", args[1].TypeName())
				}
				retVal, err = vm.RunClosure(closure, args[2:]...)

			default:
				return dune.NullValue, fmt.Errorf("argument 1 must be a string (function name), got %s", args[1].TypeName())
			}

			if err != nil {
				// return the error with the stacktrace included in the message
				// because the caller in the program will have it's own stacktrace.
				return dune.NullValue, vm.WrapError(err)
			}

			return retVal, nil
		},
	},
}

var globalKeyMutex = newKeyMutex()

func newKeyMutex() *keyMutex {
	return &keyMutex{mutexes: make(map[string]*sync.Mutex)}
}

type keyMutex struct {
	sync.RWMutex
	mutexes map[string]*sync.Mutex
}

func (m *keyMutex) getMutex(key string) *sync.Mutex {
	m.Lock()
	v, ok := m.mutexes[key]
	if !ok {
		v = &sync.Mutex{}
		m.mutexes[key] = v
	}
	m.Unlock()
	return v
}

type mutex struct {
	mutex *sync.Mutex
}

func (mutex) Type() string {
	return "sync.Mutex"
}

func (m *mutex) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "lock":
		return m.lock
	case "unlock":
		return m.unlock
	}
	return nil
}

func (m *mutex) lock(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	m.mutex.Lock()
	return dune.NullValue, nil
}

func (m *mutex) unlock(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	m.mutex.Unlock()
	return dune.NullValue, nil
}

type channel struct {
	buffer int
	c      chan dune.Value
}

func (c *channel) Type() string {
	return "sync.Channel"
}

func (c *channel) Size() int {
	return 1
}

func (c *channel) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "send":
		return c.send
	case "receive":
		return c.receive
	case "close":
		return c.close
	}
	return nil
}

func (c *channel) send(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 arg")
	}
	c.c <- args[0]
	return dune.NullValue, nil
}

func (c *channel) receive(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 args")
	}
	v := <-c.c
	return v, nil
}

func (c *channel) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 args")
	}
	close(c.c)
	return dune.NullValue, nil
}

type waitGroup struct {
	w     *sync.WaitGroup
	limit chan bool
}

func (t *waitGroup) Type() string {
	return "sync.WaitGroup"
}

func (t *waitGroup) Size() int {
	return 1
}

func (t *waitGroup) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "go":
		return t.goRun
	case "wait":
		return t.wait
	}
	return nil
}

func (t *waitGroup) goRun(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("sync") {
		return dune.NullValue, ErrUnauthorized
	}

	if t.limit != nil {
		t.limit <- true
	}

	return launchGoroutine(args, vm, t)
}

func (t *waitGroup) wait(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	t.w.Wait()
	return dune.NullValue, nil
}

type Deadline struct {
	limit  time.Time
	ticker *time.Ticker
	done   chan bool
}

func (dl *Deadline) Type() string {
	return "sync.Deadline"
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

func WithDeadline(d time.Duration, fn func(dl *Deadline) error) error {
	ticker := time.NewTicker(d)

	dl := &Deadline{
		limit:  time.Now().Add(d),
		ticker: ticker,
		done:   make(chan bool),
	}

	var err error

	go func() {
		err = fn(dl)
		dl.done <- true
	}()

	for {
		select {
		case <-dl.done:
			return err
		case t := <-ticker.C:
			if !t.Before(dl.limit) {
				ticker.Stop()
				return err
			}
		}
	}
}
