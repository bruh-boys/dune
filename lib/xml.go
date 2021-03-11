package lib

import (
	"github.com/dunelang/dune"

	"github.com/beevik/etree"
)

func init() {
	dune.RegisterLib(XML, `

declare namespace xml {
    export function newDocument(): XMLDocument

    export function readString(s: string): XMLDocument

    export interface XMLDocument {
        createElement(name: string): XMLElement
        selectElement(name: string): XMLElement
        toString(): string
    }

    export interface XMLElement {
        tag: string
        selectElements(name: string): XMLElement[]
        selectElement(name: string): XMLElement
        createElement(name: string): XMLElement
        createAttribute(name: string, value: string): XMLElement
        getAttribute(name: string): string
        setValue(value: string | number | boolean): void
        getValue(): string
    }
}


`)
}

var XML = []dune.NativeFunction{
	{
		Name: "xml.newDocument",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewObject(newXMLDoc()), nil
		},
	},
	{
		Name:      "xml.readString",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			xml := etree.NewDocument()
			if err := xml.ReadFromString(args[0].ToString()); err != nil {
				return dune.NullValue, err
			}
			return dune.NewObject(&xmlDoc{xml: xml}), nil
		},
	},
}

func newXMLDoc() *xmlDoc {
	xml := etree.NewDocument()
	xml.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	return &xmlDoc{xml: xml}
}

type xmlDoc struct {
	xml *etree.Document
}

func (t *xmlDoc) Type() string {
	return "XMLDocument"
}

func (t *xmlDoc) Size() int {
	return 1
}

func (t *xmlDoc) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "createElement":
		return t.createElement
	case "toString":
		return t.toString
	case "selectElement":
		return t.selectElement
	}
	return nil
}

func (t *xmlDoc) selectElement(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	e := t.xml.SelectElement(args[0].ToString())
	if e == nil {
		return dune.NullValue, nil
	}
	return dune.NewObject(&xmlElement{e}), nil
}

func (t *xmlDoc) toString(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}

	t.xml.Indent(2)

	s, err := t.xml.WriteToString()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewString(s), nil
}

func (t *xmlDoc) createElement(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	e := t.xml.CreateElement(args[0].ToString())

	return dune.NewObject(&xmlElement{e}), nil
}

type xmlElement struct {
	element *etree.Element
}

func (t *xmlElement) Type() string {
	return "XMLElement"
}

func (t *xmlElement) Size() int {
	return 1
}

func (t *xmlElement) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "tag":
		return dune.NewString(t.element.Tag), nil
	}
	return dune.UndefinedValue, nil
}

func (t *xmlElement) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "createAttribute":
		return t.createAttribute

	case "createElement":
		return t.createElement

	case "setValue":
		return t.setValue

	case "getAttribute":
		return t.getAttribute

	case "getValue":
		return t.getValue

	case "selectElement":
		return t.selectElement

	case "selectElements":
		return t.selectElements
	}
	return nil
}

func (t *xmlElement) selectElement(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	e := t.element.SelectElement(args[0].ToString())
	if e == nil {
		return dune.NullValue, nil
	}
	return dune.NewObject(&xmlElement{e}), nil
}

func (t *xmlElement) selectElements(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	elements := t.element.SelectElements(args[0].ToString())

	items := make([]dune.Value, len(elements))

	for i, v := range elements {
		items[i] = dune.NewObject(&xmlElement{v})
	}

	return dune.NewArrayValues(items), nil
}

func (t *xmlElement) getValue(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	v := t.element.Text()
	return dune.NewString(v), nil
}

func (t *xmlElement) getAttribute(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	a := args[0]
	v := t.element.SelectAttrValue(a.ToString(), "")
	return dune.NewString(v), nil
}

func (t *xmlElement) setValue(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 1, 1); err != nil {
		return dune.NullValue, err
	}

	a := args[0]

	switch a.Type {
	case dune.Int, dune.Float, dune.String, dune.Bool:
	default:
		return dune.NullValue, ErrInvalidType
	}

	t.element.SetText(a.ToString())
	return dune.NewObject(t), nil
}

func (t *xmlElement) createElement(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	e := t.element.CreateElement(args[0].ToString())

	return dune.NewObject(&xmlElement{e}), nil
}

func (t *xmlElement) createAttribute(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}

	t.element.CreateAttr(args[0].ToString(), args[1].ToString())
	return dune.NewObject(t), nil
}
