package dune

import (
	"fmt"
	"sync"
)

func newInstance(a *Address, vm *VM) *instance {
	if a.Kind != AddrClass {
		panic(fmt.Sprintf("Invalid class address: %v", a))
	}

	class := vm.Program.Classes[a.Value]

	return &instance{
		iMap:  make(map[string]Value),
		class: class,
	}
}

type instance struct {
	sync.RWMutex
	iMap  map[string]Value
	class *Class
}

func (i *instance) String() string {
	return "[" + i.class.Name + "]"
}

func (i *instance) Function(name string, p *Program) (*Function, bool) {
	for _, i := range i.class.Functions {
		f := p.Functions[i]
		if f.Name == name {
			return f, true
		}
	}
	return nil, false
}

// returns true if the pc is class code
func (i *instance) isSelfPC(vm *VM) bool {
	frame := vm.callStack[vm.fp]
	f := vm.Program.Functions[frame.funcIndex]

	if f.Anonimous && f.WrapClass >= 0 && i.class == vm.Program.Classes[f.WrapClass] {
		// A lambda declared inside a class can access it's private methods
		return true
	}

	return f.IsClass && i.class == vm.Program.Classes[f.Class]
}

func (i *instance) GetProperty(name string, vm *VM) (Value, error) {
	// first look for a method
	f, ok := i.Function(name, vm.Program)
	if ok {
		if !f.Exported && !i.isSelfPC(vm) {
			return NullValue, vm.NewError("Attempted to access a private method: %s", name)
		}
		m := &Method{FuncIndex: f.Index, ThisObject: NewObject(i)}
		return NewObject(m), nil
	}

	if !i.isSelfPC(vm) {
		var ok bool
		for _, f := range i.class.Fields {
			if f.Name == name {
				ok = f.Exported
				break
			}
		}
		if !ok {
			return NullValue, vm.NewError("Attempted to access a private field: %s", name)
		}
	}

	// then look for a property
	var v Value
	i.RLock()
	v = i.iMap[name]
	i.RUnlock()
	return v, nil
}

func (i *instance) SetProperty(name string, v Value, vm *VM) error {
	if !i.isSelfPC(vm) {
		var ok bool
		for _, f := range i.class.Fields {
			if f.Name == name {
				ok = f.Exported
				break
			}
		}
		if !ok {
			return vm.NewError("Attempted to access a private field: %s", name)
		}
	}

	i.Lock()
	i.iMap[name] = v
	i.Unlock()
	return nil
}
