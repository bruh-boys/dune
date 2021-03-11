package ast

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

func Print(v interface{}) error {
	p := Printer{}
	if err := p.Printf(os.Stdout, v); err != nil {
		return err
	}
	return nil
}

func Sprint(v interface{}) (string, error) {
	var b bytes.Buffer

	p := Printer{}

	if err := p.Printf(&b, v); err != nil {
		return "", err
	}

	return b.String(), nil
}

// A helper to inspect the Module
type Printer struct {
	Positions bool
}

func (p *Printer) Print(v interface{}) error {
	return p.Printf(os.Stdout, v)
}

func (p *Printer) Printf(w io.Writer, v interface{}) error {
	return p.printf(w, v, 0)
}

func (p *Printer) printf(w io.Writer, v interface{}, indentLevel int) error {
	t := reflect.TypeOf(v)
	if t == nil {
		fmt.Fprint(w, " nil\n")
		return nil
	}

	val := reflect.ValueOf(v)

	if t.Kind() == reflect.Ptr {
		if val.IsZero() {
			fmt.Fprintf(w, "%v", t.String())
			fmt.Fprint(w, " nil\n")
			return nil
		}

		fmt.Fprint(w, "*")
		elm := val.Elem()
		v = elm.Interface()
		val = reflect.ValueOf(v)
		t = val.Type()
	}

	fmt.Fprintf(w, "%v", t.String())

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Bool:
		fmt.Fprintf(w, " %v\n", v)

	case reflect.String:
		s := v.(string)
		if len(s) > 70 {
			s = s[:70] + " ..."
		}
		s = strings.ReplaceAll(s, "\n", "\\n")
		fmt.Fprintf(w, " \"%v\"\n", s)

	case reflect.Struct:
		ln := val.NumField()
		if ln == 0 {
			fmt.Fprint(w, "{}\n")
		} else {
			fmt.Fprint(w, " {\n")
			for i := 0; i < ln; i++ {
				f := val.Field(i)
				if !f.CanInterface() {
					// don't inspect internal fields, in a slice for example
					continue
				}

				fv := f.Interface()
				if p.isIgnored(fv) {
					continue
				}

				indentf(w, indentLevel+1)
				fmt.Fprint(w, t.Field(i).Name)

				fmt.Fprint(w, " ")

				if err := p.printf(w, fv, indentLevel+1); err != nil {
					return err
				}
			}
			indentf(w, indentLevel)
			fmt.Fprint(w, "}\n")
		}

	case reflect.Array, reflect.Slice:
		ln := val.Len()
		if ln == 0 {
			fmt.Fprint(w, "\n")
		} else {
			fmt.Fprint(w, "[\n")
			for i := 0; i < ln; i++ {
				indentf(w, indentLevel+1)
				av := val.Index(i).Interface()
				if err := p.printf(w, av, indentLevel+1); err != nil {
					return err
				}
			}
			indentf(w, indentLevel)
			fmt.Fprint(w, "]\n")
		}

	case reflect.Map:
		ln := val.Len()
		if ln == 0 {
			fmt.Fprint(w, "\n")
		} else {
			fmt.Fprint(w, "{\n")
			for _, k := range val.MapKeys() {
				indentf(w, indentLevel+1)
				fmt.Fprint(w, "\""+k.String()+"\": ")
				mv := val.MapIndex(k).Interface()
				if err := p.printf(w, mv, indentLevel+1); err != nil {
					return err
				}
			}
			indentf(w, indentLevel)
			fmt.Fprint(w, "}\n")
		}

	default:
		fmt.Fprint(w, " [UNKNOWN]\n")
	}

	return nil
}

func (p *Printer) isIgnored(v interface{}) bool {
	if !p.Positions {
		if _, ok := v.(Position); ok {
			return true
		}
	}
	return false
}

func indentf(w io.Writer, n int) {
	for i := 0; i < n; i++ {
		fmt.Fprint(w, "\t")
	}
}
