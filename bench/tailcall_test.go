package benchmarks

import (
	"log"
	"testing"

	"github.com/dunelang/dune"
	"github.com/dunelang/dune/parser"
)

func BenchmarkNoTailCall(b *testing.B) {
	parser.Optimizations = false

	vm := initVM(b, `
			function fact(n, a?) {
				if(a == null) {
					a = 1
				}

				if(n == 0) {
					return a
				}

				return fact(n - 1, n * a);
			}	
		`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		v, err := vm.RunFunc("fact", dune.NewInt(20))
		if err != nil {
			log.Fatal(err)
		}

		if v.ToInt() != 2432902008176640000 {
			log.Fatal(v)
		}
	}
}

func BenchmarkTailCall(b *testing.B) {
	parser.Optimizations = true

	vm := initVM(b, `
			function fact(n, a?) {
				if(a == null) {
					a = 1
				}

				if(n == 0) {
					return a
				}

				return fact(n - 1, n * a);
			}	
		`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		v, err := vm.RunFunc("fact", dune.NewInt(20))
		if err != nil {
			log.Fatal(err)
		}

		if v.ToInt() != 2432902008176640000 {
			log.Fatal(v)
		}
	}
}
