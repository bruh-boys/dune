package lib

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dunelang/dune"
)

func TestDefer001(t *testing.T) {
	vm, err := runExpr(t, `		
		let x = 0

		function main() {
			defer(() => { x++ })
			foo()
		}
		
		function foo() {
			throw "snap!"
		}
	`)

	if err == nil || !strings.Contains(err.Error(), "snap!") {
		t.Fatal(err)
	}

	v, _ := vm.RegisterValue("x")
	expected := dune.NewValue(1)

	if v != expected {
		t.Fatalf("Expected %v, got %v", 1, v)
	}
}

func TestDefer01(t *testing.T) {
	p, err := dune.CompileStr(`			
		export function main() {
			let a = 0

			defer(() => { a = 2 })
		
			a = 1
			return a
		}
	`)

	if err != nil {
		t.Fatal(err)
	}

	p.AddPermission("trusted")

	// dune.Print(p)

	vm := dune.NewVM(p)

	v, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}

	if v != dune.NewValue(2) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestDefer10(t *testing.T) {
	code := `		
		let x			
		function main() {
			 defer(() => { x = 3 })
		}
	`

	vm := assertFinalized(t, code)

	v, _ := vm.RegisterValue("x")
	if v != dune.NewValue(3) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestDefer1(t *testing.T) {
	v := runTest(t, `	
		let a = 0

		function foo() {
			 defer(() => {a = 33})
		}

		function main() {
			foo()	
			return a
		}
	`)

	if v.ToInt() != 33 {
		t.Fatal(v)
	}
}

func TestDefer2(t *testing.T) {
	v := runTest(t, `	
		let a = 0

		function foo() {
			defer(() => {a = 33})
			throw "aa"
		}

		function main() {
			try {
				foo()	
			} catch { 
				
			}
			return a
		}
	`)

	if v != dune.NewValue(33) {
		t.Fatal(v)
	}
}

func TestDefer3(t *testing.T) {
	v := runTest(t, `	
		let a = 0

		function foo() {
			defer(() => {a = 33})
			bar()
		}

		function bar() {
			throw "aa"
		}

		function main() {
			try {
				foo()	
			} catch { 
				
			}
			return a
		}
	`)

	if v.ToInt() != 33 {
		t.Fatal(v)
	}
}

func TestDefer11(t *testing.T) {
	code := `	
		let x = 0;
		
		function foo() {
			 defer(() => { x++ })
		}
		
		function main() {
			foo()
		}
	`

	vm := assertFinalized(t, code)

	v, _ := vm.RegisterValue("x")
	if v != dune.NewValue(1) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestDefer12(t *testing.T) {
	code := `	
		let x = 0
		
		function foo() {
			let f = runtime.newFinalizable(() => { x++ })
			runtime.setFinalizer(f)
		}
		
		function main() {
			foo()
		}
	`

	vm := assertFinalized(t, code)

	v, _ := vm.RegisterValue("x")
	if v != dune.NewValue(1) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestDefer13(t *testing.T) {
	code := `	
		let x = 0
		
		function bar() {
		     defer(() => { x += 10 })
		    throw "ERRR"
		}
		
		function foo() {
		     defer(() => { x += 5 })
		    bar()
		}
		
		export function main() {
			try {
				foo()
			}
			catch {
		    		x += 1;
			}
			x += 2
		}
	`

	vm := assertFinalized(t, code)

	v, _ := vm.RegisterValue("x")
	if v != dune.NewValue(18) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestDeferWithoutMain(t *testing.T) {
	var closed bool
	dune.AddNativeFunc(dune.NativeFunction{
		Name: "deferTests.close0",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			closed = true
			return dune.NullValue, nil
		},
	})

	p, err := dune.CompileStr(`		
		defer(() => deferTests.close0())	
	`)

	if err != nil {
		t.Fatal(err)
	}

	// dune.Print(p)

	vm := dune.NewVM(p)

	if _, err = vm.Run(); err != nil {
		t.Fatal(err)
	}

	if !closed {
		t.Fatalf("Defer not executed")
	}
}

func TestDeferSyntax(t *testing.T) {
	var closed bool

	dune.AddNativeFunc(dune.NativeFunction{
		Name: "deferTests.close1",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			closed = true
			return dune.NullValue, nil
		},
	})

	p, err := dune.CompileStr(`		
		defer(deferTests.close1)	
	`)

	if err != nil {
		t.Fatal(err)
	}

	vm := dune.NewVM(p)

	if _, err = vm.Run(); err != nil {
		t.Fatal(err)
	}

	if !closed {
		t.Fatalf("Defer not executed")
	}
}

func TestDeferNativeError(t *testing.T) {
	// add a dummy function to test if defer is executed after a native error
	dune.AddNativeFunc(dune.NativeFunction{
		Name: "deferTests.dummyError",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NullValue, fmt.Errorf("error!")
		},
	})

	var closed bool
	dune.AddNativeFunc(dune.NativeFunction{
		Name: "deferTests.close2",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			closed = true
			return dune.NullValue, nil
		},
	})

	p, err := dune.CompileStr(`		
			defer(() => deferTests.close2())		
			deferTests.dummyError()
	`)

	if err != nil {
		t.Fatal(err)
	}

	p.AddPermission("trusted")

	vm := dune.NewVM(p)

	if _, err = vm.Run(); err == nil {
		t.Fatal("Expected to fail")
	}

	if !closed {
		t.Fatalf("Defer not executed")
	}
}

func TestDeferNativeError2(t *testing.T) {
	// add a dummy function to test if defer is executed after a native error
	dune.AddNativeFunc(dune.NativeFunction{
		Name: "deferTests.dummyError",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NullValue, fmt.Errorf("error!")
		},
	})

	p, err := dune.CompileStr(`		
			let a
			defer(() => { a = 2 })		
			deferTests.dummyError()
	`)

	if err != nil {
		t.Fatal(err)
	}

	p.AddPermission("trusted")

	// dune.Print(p)

	vm := dune.NewVM(p)

	if _, err := vm.Run(); err == nil {
		t.Fatal("Expected to fail")
	}

	v, _ := vm.RegisterValue("a")

	if v != dune.NewValue(2) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestDeferClosure(t *testing.T) {
	p, err := dune.CompileStr(`	
		let ret = 0

		function foo() {
			let a = 1
			defer(() => { ret += a })
			a++
		}

		export function main() {
			foo()
			return ret
		}
	`)

	if err != nil {
		t.Fatal(err)
	}

	v, err := dune.NewVM(p).Run()
	if err != nil {
		t.Fatal(err)
	}

	if v != dune.NewValue(2) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestClassPrivateDefer(t *testing.T) {
	p, err := dune.CompileStr(`
		class Foo {
			i = 0
			bar() {
				defer(() => this.sum())
			}
			private sum() {
				this.i++
			}
		}

		let f = new Foo()
		f.bar()
		return f.i
	`)

	if err != nil {
		t.Fatal(err)
	}

	v, err := dune.NewVM(p).Run()
	if err != nil {
		t.Fatal(err)
	}

	if v != dune.NewValue(1) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestClassLambda(t *testing.T) {
	p, err := dune.CompileStr(`
		class Foo {
			sum() {
				let x = () => this.bar()
				return x()
			}
			private bar() {
				return 1
			}
		}

		return new Foo().sum()
	`)

	if err != nil {
		t.Fatal(err)
	}

	v, err := dune.NewVM(p).Run()
	if err != nil {
		t.Fatal(err)
	}

	if v != dune.NewValue(1) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestOutOfClassLambda(t *testing.T) {
	p, err := dune.CompileStr(`
		class Foo {
			private sum() {
				return 1
			}
		}

		let f = new Foo()
		let x = () => f.sum()
		return x()
	`)

	if err != nil {
		t.Fatal(err)
	}

	_, err = dune.NewVM(p).Run()
	if err == nil {
		t.Fatal("expected to fail")
	}

	if !strings.Contains(err.Error(), "access a private method") {
		t.Fatal(err)
	}
}

type finalizableObj struct {
	finalized bool
}

func (f *finalizableObj) Close() error {
	f.finalized = true
	return nil
}

func assertFinalized(t *testing.T, code string) *dune.VM {
	var items []*finalizableObj

	var libs = []dune.NativeFunction{
		{
			Name: "test.newFinalizable",
			Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
				v := &finalizableObj{}
				vm.SetFinalizer(v)
				items = append(items, v)
				return dune.NewObject(v), nil
			},
		},
		{
			Name: "test.newGlobalFinalizable",
			Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
				v := &finalizableObj{}
				vm.SetGlobalFinalizer(v)
				items = append(items, v)
				return dune.NewObject(v), nil
			},
		},
	}

	vm, err := runExpr(t, code, libs...)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range items {
		if !v.finalized {
			t.Fatal("Not finalized")
		}
	}

	return vm
}

func TestRunFunc(t *testing.T) {
	v := runTest(t, `	
		function main() {
			return foo(1, 2)
		}

		function foo(a, b){ 
			return runtime.runFunc("sum", a, b)
		}

		function sum(a, b){ 
			return a + b 
		}
	`)

	if v.ToInt() != 3 {
		t.Fatal(v)
	}
}

func TestRunFunc2(t *testing.T) {
	v := runTest(t, `
		let v

		function main() {
			foo()
			return v
		}

		function foo(){ 
			try {
				runtime.runFunc("sum", 1, 2)
			} finally {
				v += 2
			}
		}

		function sum(a, b){ 
			v = a + b 
		}
	`)

	if v.ToInt() != 5 {
		t.Fatal(v)
	}
}

func TestRunFunc3(t *testing.T) {
	v := runTest(t, `
		let v

		function main() {
			let p = runtime.vm.program
			let vm = runtime.newVM(p)
			try {
				vm.runFunc("sum", 1, 2)
			} finally {
				v = vm.getValue("v")
				v += 2
			}
			return v
		}

		function sum(a, b){ 
			v = a + b 
		}
	`)

	if v.ToInt() != 5 {
		t.Fatal(v)
	}
}

func TestRunFunc4(t *testing.T) {
	v := runTest(t, `
		let v

		function main() {
			return foo()
		}

		function foo() {
			let p = runtime.vm.program
			let vm = runtime.newVM(p)
			try {
				vm.runFunc("sum", 1, 2)
			} finally {
				v = vm.getValue("v")
				v += 2
			}
			return v
		}

		function sum(a, b){ 
			v = a + b 
		}
	`)

	if v.ToInt() != 5 {
		t.Fatal(v)
	}
}

// Tests: finalize
func TestFinalize1(t *testing.T) {
	assertFinalized(t, `	
		function main() {	
			let f = test.newFinalizable();
			runtime.setFinalizer(f)
		}
	`)
}

func TestFinalize2(t *testing.T) {
	assertFinalized(t, `	
		function main() {	
			test.newFinalizable();
		}
	`)
}

func TestFinalizeFunc1(t *testing.T) {
	assertFinalized(t, `		
		function foo() {
			let f = test.newFinalizable();
			runtime.setFinalizer(f)
		}
		
		function main() {
			foo()
		}
	`)
}

func TestFinalizeFunc2(t *testing.T) {
	assertFinalized(t, `		
		function foo() {
			test.newFinalizable();
		}
		
		function main() {
			foo()
		}
	`)
}

func TestFinalizeGlobal(t *testing.T) {
	assertFinalized(t, `		
		function foo() {
			test.newGlobalFinalizable();
		}
		
		function main() {
			foo()
		}
	`)
}
