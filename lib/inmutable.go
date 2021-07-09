package lib

import (
	"fmt"
	"sync"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Inmutable, `

declare namespace inmutable {
	export function newObject(canAddNew: boolean, values: any): any 
}

`)
}

var Inmutable = []dune.NativeFunction{
	{
		Name:      "inmutable.newObject",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bool, dune.Map); err != nil {
				return dune.NullValue, err
			}

			p := &InmutableObject{
				canAddNew: args[0].ToBool(),
				values:    make(map[string]dune.Value),
			}

			for k, v := range args[1].ToMap().Map {
				p.values[k.String()] = v
			}

			return dune.NewObject(p), nil
		},
	},
}

type InmutableObject struct {
	sync.RWMutex
	canAddNew bool
	values    map[string]dune.Value
}

func (*InmutableObject) Type() string {
	return "inmutable.InmutableObject"
}

func (p *InmutableObject) GetField(key string, vm *dune.VM) (dune.Value, error) {
	p.RLock()
	v, ok := p.values[key]
	p.RUnlock()

	if !ok {
		return dune.UndefinedValue, nil
	}

	return v, nil
}

func (p *InmutableObject) SetField(key string, v dune.Value, vm *dune.VM) error {
	if !p.canAddNew {
		return fmt.Errorf("can't assign to an inmutable object")
	}

	p.Lock()

	if _, ok := p.values[key]; ok {
		return fmt.Errorf("can't assign to an inmutable object")
	}
	p.values[key] = v

	p.Unlock()

	return nil
}
