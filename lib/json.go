package lib

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(JSON, `

declare namespace json {
    export function marshal(v: any, indent?: boolean, escapeHTML?: boolean): string
    export function unmarshal(str: string | byte[]): any

}
`)
}

var JSON = []dune.NativeFunction{
	{
		Name:      "json.marshal",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			ln := len(args)
			if ln == 0 || ln > 3 {
				return dune.NullValue, fmt.Errorf("expected 1, 2 or 3 arguments, got %d", len(args))
			}

			var indent bool
			var escapeHTML bool

			if ln > 1 {
				v := args[1]
				if v.Type != dune.Bool {
					return dune.NullValue, fmt.Errorf("expected arg 2 to be boolean, got %s", v.TypeName())
				}
				indent = v.ToBool()
			}

			if ln > 2 {
				v := args[2]
				if v.Type != dune.Bool {
					return dune.NullValue, fmt.Errorf("expected arg 3 to be boolean, got %s", v.TypeName())
				}
				escapeHTML = v.ToBool()
			}

			obj := args[0].ExportMarshal(0)

			buf := &bytes.Buffer{}

			encoder := json.NewEncoder(buf)

			if indent {
				encoder.SetIndent("", "    ")
			}

			encoder.SetEscapeHTML(escapeHTML)

			if err := encoder.Encode(obj); err != nil {
				return dune.NullValue, err
			}

			return dune.NewString(buf.String()), nil
		},
	},
	{
		Name:      "json.unmarshal",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if len(args) != 1 {
				return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
			}

			a := args[0]

			switch a.Type {
			case dune.String, dune.Bytes:
			default:
				return dune.NullValue, fmt.Errorf("expected argument to be string or byte[], got %v", args[0].Type)
			}

			if a.String() == "" {
				return dune.NullValue, nil
			}

			v, err := unmarshal(a.ToBytes())
			if err != nil {
				return dune.NullValue, err
			}

			return v, nil
		},
	},
}

func unmarshal(buf []byte) (dune.Value, error) {
	if len(buf) == 0 {
		return dune.NullValue, nil
	}

	var o interface{}
	err := json.Unmarshal(buf, &o)
	if err != nil {
		return dune.NullValue, err
	}

	return unmarshalObject(o)
}

func unmarshalObject(value interface{}) (dune.Value, error) {
	switch t := value.(type) {
	case nil:
		return dune.NullValue, nil
	case float32: // is this possible?
		i := int(t)
		if t == float32(i) {
			return dune.NewInt(i), nil
		}
		return dune.NewFloat(float64(t)), nil
	case float64:
		i := int(t)
		if t == float64(i) {
			return dune.NewInt(i), nil
		}
		return dune.NewFloat(t), nil
	case int, int32, int64, bool, string:
		return dune.NewValue(t), nil
	case []interface{}:
		s := make([]dune.Value, len(t))
		for i, v := range t {
			o, err := unmarshalObject(v)
			if err != nil {
				return dune.NullValue, err
			}
			s[i] = o
		}
		return dune.NewArrayValues(s), nil
	case map[string]interface{}:
		m := make(map[dune.Value]dune.Value, len(t))
		for k, v := range t {
			o, err := unmarshalObject(v)
			if err != nil {
				return dune.NullValue, err
			}
			m[dune.NewString(k)] = o
		}
		return dune.NewMapValues(m), nil

	default:
		return dune.NullValue, fmt.Errorf("invalid serialized type %T", value)
	}

}
