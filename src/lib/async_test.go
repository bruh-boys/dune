package lib

import (
	"testing"

	"github.com/dunelang/dune"
)

func TestAsyncClosure(t *testing.T) {
	p, err := dune.CompileStr(`	
		function main() {
			let a = 0
			let wg = sync.newWaitGroup()			
			wg.go(() => {
				for(let i = 0; i < 10; i++) {
					a++
				}
			})
			wg.wait()
			return a
		}
	`)

	if err != nil {
		t.Fatal(err)
	}

	p.AddPermission("trusted")

	vm := dune.NewVM(p)

	v, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}

	if v != dune.NewValue(10) {
		t.Fatalf("Returned: %v", v)
	}
}
