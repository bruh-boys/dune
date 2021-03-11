package lib

import (
	"fmt"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(libMap, `	
declare interface StringMap {
    [key: string]: string
}

declare interface KeyIndexer<T> {
    [key: string]: T
}

declare type Map<T> = KeyIndexer<T>
 
declare namespace Object {
    export function len(v: any): number
    export function keys(v: any): string[]
    export function values<T>(v: Map<T>): T[]
    export function values<T>(v: any): T[]
    export function deleteKey(v: any, key: string | number): void
    export function deleteKeys(v: any): void
    export function hasKey(v: any, key: any): boolean
    export function clone<T>(v: T): T
}
	`)
}

var libMap = []dune.NativeFunction{
	{
		Name:      "Object.isMap",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0].Type == dune.Map
			return dune.NewBool(a), nil
		},
	},
	{
		Name:      "Object.clone",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			if a.Type != dune.Map {
				return dune.NullValue, fmt.Errorf("expected a map or object, got %s", a.TypeName())
			}

			m := a.ToMap().Map

			clone := make(map[dune.Value]dune.Value, len(m))
			for k, v := range m {
				clone[k] = v
			}

			return dune.NewMapValues(clone), nil
		},
	},
	{
		Name:      "Object.len",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			if a.Type != dune.Map {
				return dune.NullValue, fmt.Errorf("expected a map or object, got %s", a.TypeName())
			}
			m := a.ToMap()
			m.RLock()
			l := len(m.Map)
			m.RUnlock()
			return dune.NewInt(l), nil
		},
	},
	{
		Name:      "Object.keys",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			var keys []dune.Value

			switch a.Type {
			case dune.Enum:
				e := vm.Program.Enums[args[0].ToEnum()]
				keys = make([]dune.Value, len(e.Values))
				for i, v := range e.Values {
					keys[i] = dune.NewString(v.Name)
				}

			case dune.Map:
				m := a.ToMap()
				m.RLock()
				keys = make([]dune.Value, len(m.Map))
				var i int
				for k := range m.Map {
					keys[i] = k
					i++
				}
				m.RUnlock()

			default:
				return dune.NullValue, fmt.Errorf("invalid type %s", a.TypeName())
			}

			return dune.NewArrayValues(keys), nil
		},
	},
	{
		Name:      "Object.values",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			if a.Type != dune.Map {
				return dune.NullValue, fmt.Errorf("expected a map or object, got %s", a.TypeName())
			}

			m := a.ToMap()
			m.RLock()
			values := make([]dune.Value, len(m.Map))
			var i int
			for k := range m.Map {
				values[i] = m.Map[k]
				i++
			}
			m.RUnlock()
			return dune.NewArrayValues(values), nil
		},
	},
	{
		Name:      "Object.deleteKey",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			if a.Type != dune.Map {
				return dune.NullValue, fmt.Errorf("expected a map or object, got %s", a.TypeName())
			}

			b := args[1]
			switch b.Type {
			case dune.String, dune.Int:
			default:
				return dune.NullValue, fmt.Errorf("invalid key type: %s", b.TypeName())
			}

			m := a.ToMap()
			m.Lock()
			delete(m.Map, b)
			m.Unlock()
			return dune.NullValue, nil
		},
	},
	{
		Name:      "Object.deleteKeys",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			if a.Type != dune.Map {
				return dune.NullValue, fmt.Errorf("expected a map or object, got %s", a.TypeName())
			}

			m := a.ToMap()
			m.Lock()
			for k := range m.Map {
				delete(m.Map, k)
			}
			m.Unlock()
			return dune.NullValue, nil
		},
	},
	{
		Name:      "Object.hasKey",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			b := args[1]

			a := args[0]
			switch a.Type {
			case dune.Map:
				m := a.ToMap()
				m.RLock()
				_, ok := m.Map[b]
				m.RUnlock()
				return dune.NewBool(ok), nil

			case dune.Object:
				if o, ok := a.ToObject().(dune.PropertyGetter); ok {
					v, err := o.GetProperty(b.ToString(), vm)
					if err != nil {
						return dune.NullValue, err
					}
					return dune.NewBool(v.Type != dune.Undefined), nil
				}
			}

			return dune.NullValue, fmt.Errorf("expected a map or object, got %s", a.TypeName())
		},
	},
}
