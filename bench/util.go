package benchmarks

import (
	"testing"

	"github.com/dunelang/dune"
)

func initVM(b *testing.B, code string) *dune.VM {
	p, err := dune.CompileStr(code)
	if err != nil {
		b.Fatal(err)
	}

	p.AddPermission("trusted")

	vm := dune.NewVM(p)

	if err := vm.Initialize(); err != nil {
		b.Fatal(err)
	}

	return vm
}
