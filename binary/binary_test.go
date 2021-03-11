package binary

import (
	"bytes"
	"testing"

	"github.com/dunelang/dune"
)

func TestClasses(t *testing.T) {
	p := compile(t, `
		function main() { 
			return new Foo().get()
		}

		// [classAttribute]
		class Foo {
			v
			constructor() {
				this.v = 5
			}
			get() {
				return this.v
			}
		}
	`)

	var buf bytes.Buffer

	err := Write(&buf, p)
	if err != nil {
		t.Fatal("Write: " + err.Error())
	}

	if p, err = Read(&buf); err != nil {
		t.Fatal("Read: " + err.Error())
	}

	if len(p.Classes) != 1 {
		t.Fatal("Expected classes")
	}

	if p.Classes[0].Attributes[0] != "classAttribute" {
		t.Fatal(p.Attributes)
	}

	assertValue(t, 5, p)
}

func TestEnum(t *testing.T) {
	p := compile(t, `
		function main() { 
			return Direction.Up
		}
		enum Direction {
			Up = 5
		}
	`)

	var buf bytes.Buffer

	err := Write(&buf, p)
	if err != nil {
		t.Fatal("Write: " + err.Error())
	}

	if p, err = Read(&buf); err != nil {
		t.Fatal("Read: " + err.Error())
	}

	if len(p.Enums) != 1 {
		t.Fatal("Expected enum")
	}

	assertValue(t, 5, p)
}

func TestBinary1(t *testing.T) {
	p := compile(t, `
		// [foo var]

		// [funcAttribute]
		function main() { 
			return 2 + 3
		}
	`)

	var buf bytes.Buffer

	err := Write(&buf, p)
	if err != nil {
		t.Fatal("Write: " + err.Error())
	}

	if p, err = Read(&buf); err != nil {
		t.Fatal("Read: " + err.Error())
	}

	if len(p.Attributes) != 1 {
		t.Fatal("Expected a attribute")
	}

	if p.Attributes[0] != "foo var" {
		t.Fatal(p.Attributes)
	}

	if p.Functions[1].Attributes[0] != "funcAttribute" {
		t.Fatal(p.Attributes)
	}

	assertValue(t, 5, p)
}

func TestBinaryResources(t *testing.T) {
	p := compile(t, `
		function main() {}
	`)

	p.Resources = map[string][]byte{"foo": []byte("bar")}

	var buf bytes.Buffer

	err := Write(&buf, p)
	if err != nil {
		t.Fatal("Write: " + err.Error())
	}

	if p, err = Read(&buf); err != nil {
		t.Fatal("Read: " + err.Error())
	}

	s := string(p.Resources["foo"])

	if s != "bar" {
		t.Fatal(s)
	}
}

func TestBinaryNativeLib(t *testing.T) {
	dune.AddNativeFunc(dune.NativeFunction{
		Name:      "math.square",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0].ToInt()
			return dune.NewInt64(v * v), nil
		},
	})

	p := compile(t, `
		function main() {
			return math.square(2)
		}
	`)

	var buf bytes.Buffer

	err := Write(&buf, p)
	if err != nil {
		t.Fatal("Write: " + err.Error())
	}

	if p, err = Read(&buf); err != nil {
		t.Fatal("Read: " + err.Error())
	}

	assertValue(t, 4, p)
}

func TestConstants(t *testing.T) {
	p := compile(t, `
		function main() { 
			return ["aaa", 1, 1.2, true, false, null, undefined, 'a']
		}
	`)

	var buf bytes.Buffer

	err := Write(&buf, p)
	if err != nil {
		t.Fatal("Write: " + err.Error())
	}

	if p, err = Read(&buf); err != nil {
		t.Fatal("Read: " + err.Error())
	}

	v, err := dune.NewVM(p).Run()
	if err != nil {
		t.Fatal(err)
	}

	a := v.ToArray()
	if a[0].ToString() != "aaa" {
		t.Fail()
	}
	if a[1].ToInt() != 1 {
		t.Fail()
	}
	if a[2].ToFloat() != 1.2 {
		t.Fail()
	}
	if !a[3].ToBool() {
		t.Fail()
	}
	if a[4].ToBool() {
		t.Fail()
	}
	if a[5] != dune.NullValue {
		t.Fail()
	}
	if a[6] != dune.UndefinedValue {
		t.Fail()
	}
	if a[7].ToRune() != 'a' {
		t.Fail()
	}
}

func compile(t *testing.T, code string) *dune.Program {
	p, err := dune.CompileStr(code)
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func assertValue(t *testing.T, expected interface{}, p *dune.Program) {
	vm := dune.NewVM(p)

	// vm.MaxSteps = 50

	ret, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}

	v := dune.NewValue(expected)

	if ret != v {
		t.Fatalf("Expected %v %T, got %v %T", expected, expected, ret, ret)
	}
}
