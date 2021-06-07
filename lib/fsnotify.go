package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dunelang/dune"

	"github.com/fsnotify/fsnotify"
)

func init() {
	dune.RegisterLib(FSNotify, `

declare namespace fsnotify {
    export function newWatcher(onEvent: EventHandler): Watcher

    export type EventHandler = (e: Event) => void

	export interface Watcher {
		add(path: string, recursive?: boolean): void
		close(): void
	}
 
	export interface Event {
		name: string
		operation: number
	}

	// const (
	// 	Create Op = 1 << iota
	// 	Write
	// 	Remove
	// 	Rename
	// 	Chmod
	// )
}

`)
}

var FSNotify = []dune.NativeFunction{
	{
		Name:        "fsnotify.newWatcher",
		Arguments:   1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0]
			switch v.Type {
			case dune.Func:
			case dune.Object:
			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", v.TypeName())
			}

			w, err := newFileWatcher(v, vm)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(w), nil
		},
	},
}

func newFileWatcher(fn dune.Value, vm *dune.VM) (*fsWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &fsWatcher{watcher: watcher}
	vm.SetGlobalFinalizer(w)

	w.start(fn, vm)

	return w, nil
}

type fsWatcher struct {
	watcher *fsnotify.Watcher
	closed  bool
}

func (w *fsWatcher) Type() string {
	return "fsnotify.Watcher"
}

func (w *fsWatcher) Close() error {
	return w.watcher.Close()
}

func (w *fsWatcher) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "add":
		return w.add
	case "close":
		return w.close
	}
	return nil
}

func (w *fsWatcher) start(fn dune.Value, vm *dune.VM) {
	var buf []fsEvent
	var timer *time.Timer

	go func() {
		for {
			if w.closed {
				break
			}

		OUTER:
			select {
			// watch for events
			case event := <-w.watcher.Events:
				if w.closed {
					return
				}

				op := int(event.Op)

				for _, e := range buf {
					if event.Name == e.name && op == e.operation {
						break OUTER
					}
				}

				e := fsEvent{
					name:      event.Name,
					operation: op,
				}

				buf = append(buf, e)

				if timer != nil {
					timer.Reset(100 * time.Millisecond)
				} else {
					timer = time.NewTimer(100 * time.Millisecond)
					go func() {
						<-timer.C
						for _, event := range buf {
							if err := runAsyncFuncOrClosure(vm, fn, dune.NewObject(event)); err != nil {
								fmt.Fprintln(vm.GetStdout(), err)
							}
						}
						buf = nil
						timer = nil
					}()
				}

			// watch for errors
			case err := <-w.watcher.Errors:
				if w.closed {
					return
				}
				fmt.Fprintln(vm.GetStdout(), "FsWatcher ERROR", err)
			}
		}
	}()
}

func (w *fsWatcher) add(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.String, dune.Bool); err != nil {
		return dune.NullValue, err
	}

	if err := ValidateArgRange(args, 1, 2); err != nil {
		return dune.NullValue, err
	}

	path := args[0].String()

	fi, err := os.Stat(path)
	if err != nil {
		return dune.NullValue, err
	}

	if !fi.Mode().IsDir() {
		err := w.watcher.Add(path)
		return dune.NullValue, err
	}

	var recursive bool
	if len(args) > 1 {
		recursive = args[1].ToBool()
	} else {
		recursive = true
	}

	if !recursive {
		err := w.watcher.Add(path)
		return dune.NullValue, err
	}

	if recursive {
		// if it is a directory add it recursively
		if err := filepath.Walk(path, w.watchDir); err != nil {
			return dune.NullValue, err
		}
	}

	return dune.NullValue, nil
}

func (w *fsWatcher) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	w.closed = true

	if err := w.watcher.Close(); err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (w *fsWatcher) watchDir(path string, fi os.FileInfo, err error) error {
	if fi.Mode().IsDir() {
		return w.watcher.Add(path)
	}
	return nil
}

type fsEvent struct {
	name      string
	operation int
}

func (e fsEvent) Type() string {
	return "fsnotify.Event"
}

func (e fsEvent) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "name":
		return dune.NewString(e.name), nil
	case "operation":
		return dune.NewInt(e.operation), nil
	}
	return dune.UndefinedValue, nil
}
