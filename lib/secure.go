package lib

import (
	"sync"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Secure, `

declare namespace secure {
	export function newSecureObject(read: boolean, write: boolean): any 
}

`)
}

var Secure = []dune.NativeFunction{
	{
		Name:        "secure.newSecureObject",
		Arguments:   2,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bool, dune.Bool); err != nil {
				return dune.NullValue, err
			}

			p := &SecureObject{
				values: make(map[string]dune.Value),
				read:   args[0].ToBool(),
				write:  args[1].ToBool(),
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

func (p *SecureObject) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
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

func (p *SecureObject) SetProperty(key string, v dune.Value, vm *dune.VM) error {
	if !p.write && !vm.HasPermission("trusted") {
		return ErrUnauthorized
	}

	p.Lock()
	p.values[key] = v
	p.Unlock()

	return nil
}
