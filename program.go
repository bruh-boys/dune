//go:generate stringer -type=Opcode,AddressKind

package dune

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type AddressKind byte

const (
	AddrVoid AddressKind = iota
	AddrLocal
	AddrGlobal
	AddrConstant
	AddrClosure
	AddrEnum
	AddrFunc
	AddrNativeFunc
	AddrClass
	AddrData
	AddrUnresolved
)

type Address struct {
	Kind  AddressKind
	Value int32
}

func (a *Address) Copy() *Address {
	if a == Void {
		return a // the copy must be the same ref
	}

	return &Address{
		Kind:  a.Kind,
		Value: a.Value,
	}
}

func NewAddress(kind AddressKind, value int) *Address {
	return &Address{kind, int32(value)}
}

func (r *Address) Equal(b *Address) bool {
	return r.Kind == b.Kind && r.Value == b.Value
}

func (r *Address) String() string {
	switch r.Kind {
	case AddrEnum:
		return fmt.Sprintf("%dE", r.Value)
	case AddrFunc:
		return fmt.Sprintf("%dF", r.Value)
	case AddrNativeFunc:
		return fmt.Sprintf("%dN", r.Value)
	case AddrConstant:
		return fmt.Sprintf("%dK", r.Value)
	case AddrGlobal:
		return fmt.Sprintf("%dG", r.Value)
	case AddrLocal:
		return fmt.Sprintf("%dL", r.Value)
	case AddrClosure:
		return fmt.Sprintf("%dC", r.Value)
	case AddrClass:
		return fmt.Sprintf("%dA", r.Value)
	case AddrData:
		return fmt.Sprintf("%dD", r.Value)
	case AddrUnresolved:
		return fmt.Sprintf("%dU", r.Value)
	case AddrVoid:
		return "--"
	default:
		return fmt.Sprintf("%d-%d?", r.Kind, r.Value)
	}
}

var Void = NewAddress(AddrVoid, 0)

type Instruction struct {
	Opcode Opcode
	A      *Address
	B      *Address
	C      *Address
}

func (r *Instruction) Copy() *Instruction {
	copy := &Instruction{}
	copy.Opcode = r.Opcode
	copy.A = r.A.Copy()
	copy.B = r.B.Copy()
	copy.C = r.C.Copy()
	return copy
}

func NewInstruction(op Opcode, a, b, c *Address) *Instruction {
	return &Instruction{op, a, b, c}
}

func (i *Instruction) String() string {
	return i.Format(false)
}

func (i *Instruction) Format(padd bool) string {
	op := strings.Title(i.Opcode.String()[3:])
	return fmt.Sprintf("%-15s %6v %6v %6v", op, i.A, i.B, i.C)
}

type Register struct {
	Name     string
	Index    int
	StartPC  int
	EndPC    int
	Exported bool
	Module   string
	KAddress *Address
}

func (r *Register) Copy() *Register {
	copy := &Register{}
	copy.Name = r.Name
	copy.Index = r.Index
	copy.StartPC = r.StartPC
	copy.EndPC = r.EndPC
	copy.Exported = r.Exported
	copy.Module = r.Module
	copy.KAddress = r.KAddress
	return copy
}

func (r *Register) Equals(b *Register) bool {
	return r.Name == b.Name &&
		r.StartPC == b.StartPC &&
		r.EndPC == b.EndPC
}

type Class struct {
	Name       string
	Exported   bool
	Module     string
	Fields     []*Field
	Functions  []int
	Attributes []string
}

func (c *Class) Copy() *Class {
	copy := &Class{}

	copy.Name = c.Name
	copy.Exported = c.Exported

	copy.Fields = make([]*Field, len(c.Fields))
	for i, v := range c.Fields {
		copy.Fields[i] = &Field{
			Name:     v.Name,
			Exported: v.Exported,
		}
	}

	copy.Functions = make([]int, len(c.Functions))
	for i, v := range c.Functions {
		copy.Functions[i] = v
	}

	return copy
}

type Field struct {
	Name     string
	Exported bool
}

type EnumList struct {
	Name     string
	Module   string
	Exported bool
	Values   []*EnumValue
}

func (e *EnumList) ValueByName(name string) (*EnumValue, int) {
	for i, v := range e.Values {
		if v.Name == name {
			return v, i
		}
	}
	return nil, -1
}

type EnumValue struct {
	Name   string
	KIndex int
}

type FunctionKind byte

const (
	User FunctionKind = iota
	Init
	Main
	Global
)

type Function struct {
	Name              string
	Module            string
	Variadic          bool
	Exported          bool
	IsClass           bool
	Class             int
	IsGlobal          bool
	Index             int
	Arguments         int
	OptionalArguments int
	MaxRegIndex       int
	Anonimous         bool
	WrapClass         int
	Kind              FunctionKind
	Registers         []*Register
	Closures          []*Register
	Instructions      []*Instruction
	Positions         []Position
	Attributes        []string
	permissions       []string
}

func (f *Function) HasPermission(name string) bool {
	for _, v := range f.Permissions() {
		if v == "trusted" {
			return true
		}
		if v == name {
			return true
		}
	}
	return false
}

func (f *Function) Permissions() []string {
	if f.permissions == nil {
		for _, attribute := range f.Attributes {
			if strings.HasPrefix(attribute, "permissions ") {
				f.permissions = strings.Split(attribute, " ")[1:]
				break
			}
		}
	}

	return f.permissions
}

func (c *Function) Copy() *Function {
	copy := &Function{}

	copy.Name = c.Name
	copy.Variadic = c.Variadic
	copy.Exported = c.Exported
	copy.IsClass = c.IsClass
	copy.IsGlobal = c.IsGlobal
	copy.Index = c.Index
	copy.Arguments = c.Arguments
	copy.OptionalArguments = c.OptionalArguments
	copy.MaxRegIndex = c.MaxRegIndex
	copy.Kind = c.Kind

	copy.Registers = make([]*Register, len(c.Registers))
	for i, v := range c.Registers {
		copy.Registers[i] = v.Copy()
	}

	copy.Closures = make([]*Register, len(c.Closures))
	for i, v := range c.Closures {
		copy.Closures[i] = v.Copy()
	}

	copy.Instructions = make([]*Instruction, len(c.Instructions))
	for i, v := range c.Instructions {
		copy.Instructions[i] = v.Copy()
	}

	copy.Positions = make([]Position, len(c.Positions))
	for i, v := range c.Positions {
		copy.Positions[i] = v.Copy()
	}

	return copy
}

type Program struct {
	sync.Mutex
	Enums       []*EnumList
	Functions   []*Function
	Classes     []*Class
	Constants   []Value
	Files       []string
	Attributes  []string
	permissions []string
	Resources   map[string][]byte

	kSize   int // the memory for all constants
	funcMap map[string]*Function
}

func (p *Program) Permissions() []string {
	if p.permissions == nil {
		for _, attribute := range p.Attributes {
			if strings.HasPrefix(attribute, "permissions ") {
				p.permissions = strings.Split(attribute, " ")[1:]
				break
			}
		}
	}

	return p.permissions
}

func (p *Program) Attribute(name string) string {
	key := name + " "
	for _, attribute := range p.Attributes {
		if strings.HasPrefix(attribute, key) {
			return strings.TrimPrefix(attribute, key)
		}
	}
	return ""
}

func (p *Program) Copy() *Program {
	copy := &Program{
		kSize: p.kSize,
	}

	copy.Enums = make([]*EnumList, len(p.Enums))
	for i, v := range p.Enums {
		copy.Enums[i] = v
	}

	copy.Functions = make([]*Function, len(p.Functions))
	for i, v := range p.Functions {
		copy.Functions[i] = v.Copy()
	}

	copy.Classes = make([]*Class, len(p.Classes))
	for i, v := range p.Classes {
		copy.Classes[i] = v.Copy()
	}

	copy.Constants = make([]Value, len(p.Constants))
	for i, v := range p.Constants {
		copy.Constants[i] = v
	}

	copy.Files = make([]string, len(p.Files))
	for i, v := range p.Files {
		copy.Files[i] = v
	}

	for k, v := range p.Attributes {
		copy.Attributes[k] = v
	}

	copy.permissions = make([]string, len(p.permissions))
	for i, v := range p.permissions {
		copy.permissions[i] = v
	}

	if p.Resources != nil {
		copy.Resources = make(map[string][]byte)
		for k, v := range p.Resources {
			copy.Resources[k] = v
		}
	}

	if p.funcMap != nil {
		copy.funcMap = make(map[string]*Function)
		for k, v := range p.funcMap {
			copy.funcMap[k] = v
		}
	}

	return copy
}

func (p *Program) AddPermission(name string) {
	p.permissions = append(p.permissions, name)
}

func (p *Program) HasPermission(name string) bool {
	for _, v := range p.Permissions() {
		if v == "trusted" {
			return true
		}
		if v == name {
			return true
		}
	}
	return false
}

func (p *Program) addConstant(v Value) *Address {
	for i, k := range p.Constants {
		if k.Type == v.Type && k.object == v.object {
			return NewAddress(AddrConstant, i)
		}
	}

	i := len(p.Constants)
	p.Constants = append(p.Constants, v)
	return NewAddress(AddrConstant, i)
}

func (p *Program) Strip() {
	for i := range p.Functions {
		f := p.Functions[i]
		f.Positions = nil
		if strings.Contains(f.Name, ".prototype.") {
			continue
		}
		if f.Name != "main" {
			f.Name = ""
		}
		for j := range f.Registers {
			r := f.Registers[j]
			if !r.Exported {
				r.Name = ""
			}
		}
	}
	p.Files = nil
}

func (p *Program) FileIndex(file string) int {
	for i, f := range p.Files {
		if file == f {
			return i
		}
	}
	return -1
}

func (p *Program) ToTraceLine(f *Function, pc int) TraceLine {
	ln := len(f.Positions)

	if ln == 0 || ln <= pc {
		return TraceLine{Function: f.Name}
	}

	// Instructions that have an empty position belong to the last source line in code.
	var file string
	var pos Position
	for {
		pos = f.Positions[pc]
		if pc > 0 && pos.Line == 0 { // pos.Line is in base 1 so this is an empty position
			pc--
			continue
		}

		if len(p.Files) > 0 {
			file = p.Files[pos.File]
		}
		break
	}

	return TraceLine{Function: f.Name, File: file, Line: pos.Line}
}

func (p *Program) initFuncMap() {
	if p.funcMap != nil {
		return
	}

	funcMap := make(map[string]*Function, len(p.Functions))
	p.funcMap = funcMap
	for _, f := range p.Functions {
		if !f.IsClass {
			funcMap[f.Name] = f
		}
	}
}

func (p *Program) Function(name string) (*Function, bool) {
	p.Lock()
	p.initFuncMap()
	f, ok := p.funcMap[name]
	p.Unlock()
	return f, ok
}

type TraceLine struct {
	Function string
	File     string
	Line     int
}

func (p TraceLine) String() string {
	var buf bytes.Buffer

	switch p.File {
	case "", ".":
		if p.Line > 0 {
			fmt.Fprintf(&buf, "line %d", p.Line)
		} else {
			fmt.Fprint(&buf, p.Function)
		}
	default:
		fmt.Fprintf(&buf, "%s:%d", p.File, p.Line)
	}

	return buf.String()
}

func (p TraceLine) SameLine(o TraceLine) bool {
	return p.File == o.File && p.Line == o.Line
}

type Position struct {
	File   int
	Line   int
	Column int
}

func (p Position) Copy() Position {
	return Position{
		File:   p.File,
		Line:   p.Line,
		Column: p.Column,
	}
}

func Print(p *Program) {
	Fprint(os.Stdout, p)
}

func PrintFunction(f *Function, p *Program) {
	FprintFunction(os.Stdout, f, p)
}

func Sprint(p *Program) (string, error) {
	var b bytes.Buffer
	Fprint(&b, p)
	return b.String(), nil
}

const separator1 = "\n==============================================================="
const separator2 = "\n---------------------------------------------------------------"

func Fprint(w io.Writer, p *Program) {
	if len(p.Attributes) > 0 {
		fmt.Fprint(w, separator1)
		fmt.Fprint(w, "\nAttributes")
		fmt.Fprint(w, separator1)
		for _, d := range p.Attributes {
			fmt.Fprintf(w, "\n %s", d)
		}
		fmt.Fprint(w, "\n")
	}

	if len(p.Functions) > 0 {
		fmt.Fprint(w, separator1)
		fmt.Fprint(w, "\nFunctions")
		fmt.Fprint(w, separator1)
		for _, f := range p.Functions {
			if f.IsClass {
				continue
			}
			FprintFunction(w, f, p)
		}
	}

	if len(p.Classes) > 0 {
		fmt.Fprint(w, separator1)
		fmt.Fprint(w, "\nClasses")
		fmt.Fprint(w, separator1)
		for i, c := range p.Classes {
			fmt.Fprintf(w, "\n%dC Class %s", i, c.Name)
			for _, f := range c.Functions {
				FprintFunction(w, p.Functions[f], p)
			}
		}
	}

	if len(p.Enums) > 0 {
		fmt.Fprint(w, separator1)
		fmt.Fprint(w, "\nEnums")
		fmt.Fprint(w, separator1)
		for i, enum := range p.Enums {
			fmt.Fprintf(w, "\n%dE %s", i, enum.Name)
			for ii, v := range enum.Values {
				k := p.Constants[v.KIndex]
				fmt.Fprintf(w, "\n  %-5d %s=%v", ii, v.Name, k.String())
			}
		}
		fmt.Fprint(w, "\n")
	}

	if len(p.Constants) > 0 {
		fmt.Fprint(w, separator1)
		fmt.Fprint(w, "\nConstants")
		fmt.Fprint(w, separator1)
		fmt.Fprintln(w)
		FprintConstants(w, p)
	}

	fmt.Fprint(w, "\n")
}

func FprintFunction(w io.Writer, f *Function, p *Program) {
	name := f.Name

	if f.IsClass {
		cl := p.Classes[f.Class]
		name = fmt.Sprintf("%s.%s", cl.Name, name)
	}

	fmt.Fprintf(w, "\n%dF %s", f.Index, name)
	fmt.Fprint(w, separator2)

	if len(f.Attributes) > 0 {
		fmt.Fprint(w, "\nAttributes")
		fmt.Fprint(w, separator2)
		for _, d := range f.Attributes {
			fmt.Fprintf(w, "\n %s", d)
		}
		fmt.Fprint(w, separator2)
	}

	for i, v := range f.Instructions {
		printInstruction(w, p, f, i, v)
	}

	var regType string
	if f.Index == 0 {
		regType = "G"
	} else {
		regType = "L"
	}

	//fmt.Fprintf(w, "\n  MaxRegIndex %d", f.MaxRegIndex)
	fmt.Fprintln(w)
	for i, r := range f.Registers {
		fmt.Fprintf(w, "\n  %d%s %s %d-%d", i, regType, r.Name, r.StartPC, r.EndPC)
	}

	fmt.Fprint(w, "\n")
}

func printInstruction(w io.Writer, p *Program, f *Function, i int, instr *Instruction) {
	fmt.Fprintf(w, "\n  %-5d %s", i, instr.Format(true))

	if len(f.Positions) > i {
		t := p.ToTraceLine(f, i)
		fmt.Fprintf(w, "   ;   %s", t.String())
	}

	fmt.Fprint(w)
}

func FprintConstants(w io.Writer, p *Program) {
	for i, k := range p.Constants {
		switch k.Type {
		case String:
			s := k.String()
			if len(s) > 50 {
				s = s[:50]
			}
			s = strings.Replace(s, "\n", "\\n", -1)
			fmt.Fprintf(w, "%dK string %v\n", i, s)
		default:
			fmt.Fprintf(w, "%dK %v %v\n", i, k.Type, k.String())
		}
	}
}

func SprintNames(p *Program, registers bool) (string, error) {
	var b bytes.Buffer
	FprintNames(&b, p, registers)
	return b.String(), nil
}

func PrintNames(p *Program, registers bool) {
	FprintNames(os.Stdout, p, registers)
}

func FprintNames(w io.Writer, p *Program, registers bool) {
	if len(p.Classes) > 0 {
		for i, c := range p.Classes {
			fmt.Fprintf(w, "\n%dC %s", i, c.Name)

			functions := make([]*Function, len(c.Functions))
			for i, fIndex := range c.Functions {
				functions[i] = p.Functions[fIndex]
			}
			fprintFunctionNames(w, p, true, 1, functions, registers)
		}
		fmt.Fprint(w, "\n")
	}

	fprintFunctionNames(w, p, false, 0, p.Functions, registers)
	fmt.Fprint(w, "\n")
}

func fprintFunctionNames(w io.Writer, p *Program, isClass bool, indent int, functions []*Function, registers bool) {
	for i, f := range functions {
		if !isClass && f.IsClass {
			continue
		}

		fmt.Fprintf(w, "\n%s%dF %s", strings.Repeat("\t", indent), i, f.Name)

		fmt.Fprintf(w, "    %s", p.ToTraceLine(f, 0).String())

		var addrType string
		if f.IsGlobal {
			addrType = "G"
		} else {
			addrType = "L"
		}

		if registers {
			for j, r := range f.Registers {
				fmt.Fprintf(w, "\n%s%d%s %s", strings.Repeat("\t", indent+1), j, addrType, r.Name)
			}
		}
	}
}
