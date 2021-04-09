package lib

import (
	"fmt"
	"time"

	"github.com/dunelang/dune"

	"github.com/dunelang/dune/lib/cache"
)

func init() {
	dune.RegisterLib(Caching, `
	
declare namespace caching {
 
    export function newCache(d?: time.Duration | number): Cache

    export interface Cache {
        get(key: string): any | null
        save(key: string, v: any): void
        delete(key: string): void
        keys(): string[]
        items(): Map<any>
        clear(): void
    }
}

`)
}

var Caching = []dune.NativeFunction{
	{
		Name:      "caching.newCache",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)

			var d time.Duration

			switch l {
			case 0:
				d = 1 * time.Minute
			case 1:
				var a = args[0]
				switch a.Type {
				case dune.Int:
					dd, err := ToDuration(a)
					if err != nil {
						return dune.NullValue, err
					}
					d = dd
				case dune.Object:
					dur, ok := a.ToObject().(Duration)
					if !ok {
						return dune.NullValue, fmt.Errorf("expected duration, got %s", a.TypeName())
					}
					d = time.Duration(dur)
				}
			default:
				return dune.NullValue, fmt.Errorf("expected 0 or 1 arguments, got %d", l)
			}
			return dune.NewObject(newCacheObj(d)), nil
		},
	},
}

func newCacheObj(d time.Duration) *cacheObj {
	return &cacheObj{
		cache: cache.New(d, 30*time.Second),
	}
}

type cacheObj struct {
	cache *cache.Cache
}

func (*cacheObj) Type() string {
	return "caching.Cache"
}

func (c *cacheObj) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "get":
		return c.get
	case "save":
		return c.save
	case "delete":
		return c.delete
	case "clear":
		return c.clear
	case "keys":
		return c.keys
	case "items":
		return c.items
	}
	return nil
}

func (c *cacheObj) keys(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	keys := c.cache.Keys()

	m := make([]dune.Value, len(keys))

	for i, k := range keys {
		m[i] = dune.NewString(k)
	}

	return dune.NewArrayValues(m), nil
}

func (c *cacheObj) items(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	items := c.cache.Items()

	m := make(map[dune.Value]dune.Value, len(items))

	for k, v := range items {
		m[dune.NewString(k)] = v.Object.(dune.Value)
	}

	return dune.NewMapValues(m), nil
}

func (c *cacheObj) get(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	key := args[0].String()
	if i, ok := c.cache.Get(key); ok {
		return i.(dune.Value), nil
	}
	return dune.NullValue, nil
}

func (c *cacheObj) save(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 2 {
		return dune.NullValue, fmt.Errorf("expected 2 args, got %d", len(args))
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("invalid argument type, expected string, got %s", args[0].TypeName())
	}
	key := args[0].String()
	v := args[1]
	c.cache.Set(key, v, cache.DefaultExpiration)
	return dune.NullValue, nil
}

func (c *cacheObj) delete(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	key := args[0].String()
	c.cache.Delete(key)
	return dune.NullValue, nil
}

func (c *cacheObj) clear(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	c.cache.Flush()
	return dune.NullValue, nil
}
