package lib

import (
	"fmt"
	"io"
	"strings"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(libFmt, `
declare namespace fmt {
    export function print(...n: any[]): void
    export function println(...n: any[]): void
    export function printf(format: string, ...params: any[]): void
    export function sprintf(format: string, ...params: any[]): string
	export function fprintf(w: io.Writer, format: string, ...params: any[]): void
	
    export function errorf(format: string, ...params: any[]): errors.Error
}	
	`)
}

var libFmt = []dune.NativeFunction{
	{
		Name:      "fmt.errorf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return dune.NullValue, fmt.Errorf("expected at least 1 parameter, got %d", len(args))
			}
			v := args[0]
			if v.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a string, got %s", v.Type)
			}

			values := make([]interface{}, argsLen-1)
			for i, a := range args[1:] {
				v := a.Export(0)
				if t, ok := v.(*dune.Error); ok {
					// wrap errors showing only the message
					v = t.Message()
				}
				values[i] = v
			}

			key := Translate(v.String(), vm)

			tokens, s := FormatTemplateTokens(key, values...)

			err := dune.NewPublicError(s)

			tkIndex := 0
			for _, tk := range tokens {
				if tk.Type == Parameter {
					if tk.Value == "wrap" {
						if tkIndex < argsLen {
							// +1 because the first arg is the format string
							e, ok := args[tkIndex+1].ToObjectOrNil().(*dune.Error)
							if ok {
								err.Wrapped = append(err.Wrapped, e)
							}
						}
					}
					tkIndex++
				}
			}

			return dune.NewObject(err), nil
		},
	},
	{
		Name:      "fmt.print",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			for _, v := range args {
				fmt.Fprint(vm.GetStdout(), fmtValue(v))
			}
			return dune.NullValue, nil
		},
	},
	{
		Name:      "fmt.println",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			for i, v := range args {
				if i > 0 {
					fmt.Fprint(vm.GetStdout(), " ")
				}
				fmt.Fprint(vm.GetStdout(), fmtValue(v))
			}
			fmt.Fprint(vm.GetStdout(), "\n")
			return dune.NullValue, nil
		},
	},
	{
		Name:      "fmt.printf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l == 0 {
				return dune.NullValue, fmt.Errorf("expected at least 1 parameter, got %d", len(args))
			}
			v := args[0]
			if v.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a string, got %s", v.Type)
			}

			values := make([]interface{}, l-1)
			for i, v := range args[1:] {
				values[i] = fmtValue(v)
			}

			fmt.Fprintf(vm.GetStdout(), v.String(), values...)
			return dune.NullValue, nil
		},
	},
	{
		Name:      "fmt.fprintf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l < 2 {
				return dune.NullValue, fmt.Errorf("expected at least 2 parameters, got %d", len(args))
			}

			w, ok := args[0].ToObjectOrNil().(io.Writer)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a io.Writer, got %s", args[0].TypeName())
			}

			v := args[1]
			if v.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 2 to be a string, got %s", v.TypeName())
			}

			values := make([]interface{}, l-2)
			for i, v := range args[2:] {
				values[i] = fmtValue(v)
			}

			fmt.Fprintf(w, v.String(), values...)
			return dune.NullValue, nil
		},
	},
	{
		Name:      "fmt.sprintf",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l == 0 {
				return dune.NullValue, fmt.Errorf("expected at least 1 parameter, got %d", len(args))
			}
			v := args[0]
			if v.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a string, got %s", v.Type)
			}

			values := make([]interface{}, l-1)
			for i, v := range args[1:] {
				values[i] = fmtValue(v)
			}

			s := fmt.Sprintf(v.String(), values...)
			return dune.NewString(s), nil
		},
	},
}

func fmtValue(v dune.Value) interface{} {
	switch v.Type {
	case dune.Object:
		return v.String()
	default:
		return v.Export(0)
	}
}

const (
	Text      = 0
	Parameter = 1
)

type Token struct {
	Type  int
	Value string
}

func Parse(text string) ([]Token, error) {
	var tokens []Token

	var buf []byte
	var inKey bool

	for i, l := 0, len(text); i < l; i++ {
		c := text[i]

		switch c {
		case '{':
			if !inKey && len(buf) > 0 {
				tokens = append(tokens, Token{Type: Text, Value: string(buf)})
				buf = nil
			}

			if i < l-1 && text[i+1] == '{' {
				inKey = true
				i++
				continue
			}

			buf = append(buf, c)

		case '}':
			if inKey && i < l-1 && text[i+1] == '}' {
				value := strings.Trim(string(buf), " ")
				tokens = append(tokens, Token{Type: Parameter, Value: value})
				buf = nil
				inKey = false
				i++
				continue
			}

			buf = append(buf, c)

		default:
			buf = append(buf, c)
		}
	}

	if inKey {
		return nil, fmt.Errorf(("unclosed key"))
	}

	if len(buf) > 0 {
		tokens = append(tokens, Token{Type: Text, Value: string(buf)})
	}

	return tokens, nil
}

func FormatTemplate(template string, args ...interface{}) string {
	_, s := FormatTemplateTokens(template, args...)
	return s
}

func FormatTemplateTokens(template string, args ...interface{}) ([]Token, string) {
	tokens, err := Parse(template)
	if err != nil {
		return nil, fmt.Sprintf("[ParseError] %s: %v", template, err)
	}

	result := make([]string, len(tokens))
	argCount := len(args)
	var argIndex int

	for i, tk := range tokens {
		switch tk.Type {
		case Text:
			result[i] = tk.Value

		case Parameter:
			if argCount < argIndex {
				result[i] = ""
				continue
			}

			var v interface{}

			if argIndex >= argCount {
				break
			}

			v = args[argIndex]
			argIndex++

			switch t := v.(type) {
			case string:
				result[i] = t
			case int, int32, int64:
				result[i] = fmt.Sprintf("%d", t)
			case float32, float64:
				result[i] = fmt.Sprintf("%.2f", t)
			default:
				result[i] = fmt.Sprintf("%v", t)
			}
		}
	}

	return tokens, strings.Join(result, "")
}
