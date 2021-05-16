package benchmarks

import (
	"log"
	"testing"
	"time"

	"github.com/dunelang/dune"
)

func BenchmarkCall(b *testing.B) {
	vm := initVM(b, `
		function foo() {
			return time.now().hour
		}`)

	b.ResetTimer()
	b.ReportAllocs()

	var a dune.Value
	var err error
	for i := 0; i < b.N; i++ {
		a, err = vm.RunFunc("foo")
		if err != nil {
			log.Fatal(err)
		}
	}
	_ = a
}

func BenchmarkGoCall(b *testing.B) {
	fn := func() int {
		return time.Now().Hour()
	}

	b.ResetTimer()
	b.ReportAllocs()

	var a int
	for i := 0; i < b.N; i++ {
		a = fn()
	}
	_ = a
}

func BenchmarkLoop(b *testing.B) {
	vm := initVM(b, `
		function foo() {
			let a = 0
			for(let i = 0; i < 100; i++) {
				a += time.now().hour
			}
			return a
		}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := vm.RunFunc("foo")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func BenchmarkGoLoop(b *testing.B) {
	fn := func() int {
		a := 0
		for i := 0; i < 100; i++ {
			a += time.Now().Hour()
		}
		return a
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fn()
	}
}
