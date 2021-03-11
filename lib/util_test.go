package lib

import (
	"testing"

	"github.com/dunelang/dune"
)

func runTest(t *testing.T, code string, args ...dune.Value) dune.Value {
	p, err := dune.CompileStr(code)
	if err != nil {
		t.Fatal(err)
	}

	p.AddPermission("trusted")

	vm := dune.NewVM(p)
	vm.MaxSteps = 400

	// dune.Print(p)

	v, err := vm.Run(args...)
	if err != nil {
		//fmt.Println(code)
		t.Fatal(err)
	}

	return v
}

func runExpr(t *testing.T, code string, funcs ...dune.NativeFunction) (*dune.VM, error) {
	for _, f := range funcs {
		dune.AddNativeFunc(f)
	}

	p, err := dune.CompileStr(code)
	if err != nil {
		return nil, err
	}

	p.AddPermission("trusted")

	vm := dune.NewVM(p)

	for _, f := range funcs {
		dune.AddNativeFunc(f)
	}

	vm.MaxSteps = 1000

	_, err = vm.Run()
	return vm, err
}

func assertRegister(t *testing.T, register string, expected interface{}, code string) {
	vm, err := runExpr(t, code)
	if err != nil {
		t.Fatal(err)
	}

	ex := dune.NewValue(expected)

	v, _ := vm.RegisterValue(register)
	if v != ex {
		t.Fatalf("Expected %v, got %v", ex, v)
	}
}
