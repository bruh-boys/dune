package parser

import (
	"strings"
	"testing"

	"github.com/dunelang/dune/ast"
)

func TestParseFuncAttributes(t *testing.T) {
	a, err := ParseStr(`
		// [attribute1]
		function bar() {

		}
	`)

	if err != nil {
		t.Fatal(err)
	}

	// ast.Print(a.File)

	fn, ok := a.File.Stms[0].(*ast.FuncDeclStmt)
	if !ok {
		t.Fatalf("Expected FuncDeclStmt, got %T", a.File.Stms[0])
	}

	if len(fn.Attributes) != 1 {
		t.Fatal(fn.Attributes)
	}
	if fn.Attributes[0] != "attribute1" {
		t.Fatal(fn.Attributes)
	}
}

func TestParseFuncAttributes2(t *testing.T) {
	a, err := ParseStr(`
		// [attribute ignore]

		// [attribute1]
		// [attribute2 foo]
		function bar() {

		}
	`)

	if err != nil {
		t.Fatal(err)
	}

	// ast.Print(a.File)

	if len(a.File.Attributes) != 1 {
		t.Fatal(a.File.Attributes)
	}

	fn, ok := a.File.Stms[0].(*ast.FuncDeclStmt)
	if !ok {
		t.Fatalf("Expected FuncDeclStmt, got %T", a.File.Stms[0])
	}

	if len(fn.Attributes) != 2 {
		t.Fatal(fn.Attributes)
	}
	if fn.Attributes[0] != "attribute1" || fn.Attributes[1] != "attribute2 foo" {
		t.Fatal(fn.Attributes)
	}
}

func TestParseClassAttributes(t *testing.T) {
	a, err := ParseStr(`
		// [attribute1]
		class bar { }
	`)

	if err != nil {
		t.Fatal(err)
	}

	// ast.Print(a.File)

	class, ok := a.File.Stms[0].(*ast.ClassDeclStmt)
	if !ok {
		t.Fatalf("Expected FuncDeclStmt, got %T", a.File.Stms[0])
	}

	if len(class.Attributes) != 1 {
		t.Fatal(class.Attributes)
	}
	if class.Attributes[0] != "attribute1" {
		t.Fatal(class.Attributes)
	}
}
func TestParseSelector1(t *testing.T) {
	a, err := ParseStr(`let a = b.c.d`)
	if err != nil {
		t.Fatal(err)
	}

	// ast.Print(a.File)

	exp, ok := a.File.Stms[0].(*ast.VarDeclStmt)
	if !ok {
		t.Fatalf("Expected VarDeclStmt, got %T", a.File.Stms[0])
	}

	sel, ok := exp.Value.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("Expected SelectorExpr, got %T", exp.Value)
	}

	if !sel.First {
		t.Fatal("Not first")
	}
}

func TestParseSelector2(t *testing.T) {
	a, err := ParseStr(`let a = b?.()`)
	if err != nil {
		t.Fatal(err)
	}

	// ast.Print(a.File)

	exp, ok := a.File.Stms[0].(*ast.VarDeclStmt)
	if !ok {
		t.Fatalf("Expected VarDeclStmt, got %T", a.File.Stms[0])
	}

	call, ok := exp.Value.(*ast.CallExpr)
	if !ok {
		t.Fatalf("Expected CallExpr, got %T", exp.Value)
	}

	if !call.First {
		t.Fatal("Not first")
	}
}

func TestParseSelector3(t *testing.T) {
	a, err := ParseStr(`let a = b?.[0]`)
	if err != nil {
		t.Fatal(err)
	}

	//ast.Print(a.File)

	exp, ok := a.File.Stms[0].(*ast.VarDeclStmt)
	if !ok {
		t.Fatalf("Expected VarDeclStmt, got %T", a.File.Stms[0])
	}

	i, ok := exp.Value.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("Expected IndexExpr, got %T", exp.Value)
	}

	if !i.First {
		t.Fatal("Not first")
	}
}

func TestParseSwitchFallthrough1(t *testing.T) {
	_, err := ParseStr(`
		switch(1) {
		case 1:
		case 2:
			let a = 3
		}
		
	`)

	if err != nil {
		t.Fatal(err)
	}
}

func TestParseSwitchFallthrough2(t *testing.T) {
	_, err := ParseStr(`
		switch(1) {
		case 1:
			let a = 3

		case 2:
		}
		
	`)

	if err == nil || !strings.Contains(err.Error(), "Fallthrough") {
		t.Fatal(err)
	}
}

func TestParseSwitchFallthrough3(t *testing.T) {
	_, err := ParseStr(`
		switch(1) {
		case 1:
			let a = 3

		default:

		case 2:
			let b = 3
		}
		
	`)

	if err == nil || !strings.Contains(err.Error(), "Fallthrough") {
		t.Fatal(err)
	}
}
func TestIgnoreTypeBugfix(t *testing.T) {
	_, err := ParseStr(`export type V = { [key: string]: X }`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIgnoreTypeBugfix2(t *testing.T) {
	_, err := ParseStr(`export type StringOrPlurals = string | string[]`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIgnoreTypeBugfix3(t *testing.T) {
	_, err := ParseStr(`export type Translation = StringOrPlurals | Map<StringOrPlurals>`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseClassProperty(t *testing.T) {
	a, err := ParseStr(`
		class foo {
			private get bar() {}
			private set bar2(v) {} 
		}
	`)
	if err != nil {
		t.Fatal(err)
	}

	class, ok := a.File.Stms[0].(*ast.ClassDeclStmt)
	if !ok {
		t.Fatalf("Expected VarDeclStmt, got %T", a.File.Stms[0])
	}

	if len(class.Getters) != 1 {
		t.Fatal("expected one getter")
	}

	if len(class.Setters) != 1 {
		t.Fatal("expected one setter")
	}
}

func TestParseClassProperty2(t *testing.T) {
	a, err := ParseStr(`
		class foo {
			get() {}
			set bar(v) {} 
		}
	`)
	if err != nil {
		t.Fatal(err)
	}

	class, ok := a.File.Stms[0].(*ast.ClassDeclStmt)
	if !ok {
		t.Fatalf("Expected VarDeclStmt, got %T", a.File.Stms[0])
	}

	if len(class.Functions) != 1 {
		t.Fatal("expected one func")
	}

	if len(class.Getters) != 0 {
		t.Fatal("expected 0 getter")
	}

	if len(class.Setters) != 1 {
		t.Fatal("expected one setter")
	}
}

func TestParseVariadicArg(t *testing.T) {
	_, err := ParseStr(`
		let foo: any
		foo.bar((...args: any[]) => { })
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseLambdaNullableArg(t *testing.T) {
	_, err := ParseStr(`
		type foo = (arg?: any) => void
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseEnumNegativeValue(t *testing.T) {
	_, err := ParseStr(`
		export enum Priority {
			high = 2,
			low = -1,
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
}
