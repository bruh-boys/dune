package lib

import (
	"fmt"
	"sync"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Secure, `

declare namespace secure {
    /**
     * 
     * @param read true if anyone can read it. false if only trusted code can read it.
     * @param write true if anyone can write it. false if only trusted code can write it.
     */
	export function newObject(read: boolean, write: boolean, values?: any): any 
}

`)
}

var Secure = []dune.NativeFunction{
	{
		Name:        "secure.newObject",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			ln := len(args)
			if ln < 2 || ln > 3 {
				return dune.NullValue, fmt.Errorf("expected 2 or 3 arguments, got %d", ln)
			}

			if args[0].Type != dune.Bool {
				return dune.NullValue, fmt.Errorf("expected argument 1 to be bool, got %s", args[0].Type)
			}

			if args[1].Type != dune.Bool {
				return dune.NullValue, fmt.Errorf("expected argument 2 to be bool, got %s", args[1].Type)
			}
			p := &SecureObject{
				values: make(map[string]dune.Value),
				read:   args[0].ToBool(),
				write:  args[1].ToBool(),
			}

			if ln == 3 {
				for k, v := range args[2].ToMap().Map {
					p.values[k.String()] = v
				}
			}

			return dune.NewObject(p), nil
		},
	},
}

type SecureObject struct {
	sync.RWMutex
	values map[string]dune.Value
	read   bool
	write  bool
}

func (*SecureObject) Type() string {
	return "secure.SecureObject"
}

func (p *SecureObject) GetField(key string, vm *dune.VM) (dune.Value, error) {
	if !p.read && !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	p.RLock()
	v, ok := p.values[key]
	p.RUnlock()

	if !ok {
		return dune.UndefinedValue, nil
	}

	return v, nil
}

func (p *SecureObject) SetField(key string, v dune.Value, vm *dune.VM) error {
	if !p.write && !vm.HasPermission("trusted") {
		return ErrUnauthorized
	}

	p.Lock()
	p.values[key] = v
	p.Unlock()

	return nil
}
