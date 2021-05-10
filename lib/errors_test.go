package lib

import (
	"os"
	"testing"

	"github.com/dunelang/dune"
)

func TestErrorIs(t *testing.T) {
	v := runTest(t, `
		let fs = io.newVirtualFS()

		try {
			fs.open("x")
		} catch(err) {
			return err
		}	
	`)

	err, ok := v.ToObjectOrNil().(*dune.VMError)
	if !ok {
		t.Fatalf("Type: %T", v)
	}

	if !err.Is(os.ErrNotExist.Error()) {
		t.Fatal("IS", err)
	}
}

func TestErrorWrap(t *testing.T) {
	v := runTest(t, `return fmt.errorf("ERROR %s", errors.newError("Snap!"))`)

	err, ok := v.ToObjectOrNil().(*dune.VMError)
	if !ok {
		t.Fatalf("Type: %T", v)
	}

	if err.ErrorMessage() != "ERROR Snap!" {
		t.Fatal("msg", err.ErrorMessage())
	}

	if err.Wrapped == nil {
		t.Fatal("wrap", err.Wrapped)
	}
}

func _TestErrorWrap2(t *testing.T) {
	v := runTest(t, `return fmt.typeErrorf("io", "ERROR %s", errors.newTypeError("io", "Snap!"))`)

	err, ok := v.ToObjectOrNil().(*dune.VMError)
	if !ok {
		t.Fatalf("Type: %T", v)
	}

	if err.ErrorMessage() != "ERROR Snap!" {
		t.Fatal("msg", err.ErrorMessage())
	}

	if err.Wrapped == nil {
		t.Fatal("wrap", err.Wrapped)
	}

	if !err.Is("io") {
		t.Fatal("IS", err)
	}
}
