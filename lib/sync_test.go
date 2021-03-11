package lib

import (
	"testing"

	"github.com/dunelang/dune"
)

func TestMutex(t *testing.T) {
	p, err := dune.CompileStr(`	
		var mutex = sync.newMutex()

		function doStuff() {
			mutex.lock()
			defer(() => mutex.unlock())
		
			for (let i = 0; i < 10; i++) {
				throw "foo"
			}
		}
		
		export function main() {
			let wg = sync.newWaitGroup(10)
		
			for (let i = 0; i < 2; i++) {
				wg.go(doStuff)
			}
			
			wg.wait()
			return 1
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

	if v != dune.NewValue(1) {
		t.Fatalf("Returned: %v", v)
	}
}

func TestMutex2(t *testing.T) {
	runTest(t, `	
		function foo(m: Mutex) {
			m.lock()
			defer(() => m.unlock())
		}

		function main() {			
			let mutex = sync.newMutex()
			foo(mutex)
			foo(mutex)
		}
	`)
}

func TestMutex3(t *testing.T) {
	runTest(t, `
		class Device {
			private mutex = sync.newMutex()

			open() {
        		this.mutex.lock()
			}

			close() {
            	this.mutex.unlock()
			}
		}

		function foo(d: Device) {
			d.open()
			defer(() => d.close())
		}

		function main() {
			let d = new Device()
			foo(d)
			foo(d)
		}
	`)
}
