package benchmarks

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/dunelang/dune"
	_ "github.com/dunelang/dune/lib"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})

	// launch a Go server 9056
	go func() {
		http.ListenAndServe(":9056", nil)
	}()

	// launch a Dune server 9055
	go func() {
		p, err := dune.CompileStr(`
			let s = http.newServer()
			s.handler = (w, r) => w.write("Hello world")
			s.address = ":9055"
			s.start()`)
		if err != nil {
			log.Fatal(err)
		}

		p.AddPermission("trusted")

		if _, err := dune.NewVM(p).Run(); err != nil {
			log.Fatal(err)
		}
	}()
}

func BenchmarkParallelHTTP_Go(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			get(b, "http://localhost:9056")
		}
	})
}

func BenchmarkParallelHTTP(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			get(b, "http://localhost:9055")
		}
	})
}

func BenchmarkHTTP_Go(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		get(b, "http://localhost:9056")
	}
}
func BenchmarkHTTP(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		get(b, "http://localhost:9055")
	}
}

func get(b *testing.B, url string) {
	r, err := http.Get(url)
	if err != nil {
		b.Fatal(err)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		b.Fatal(err)
	}

	if string(data) != "Hello world" {
		b.Fatal(string(data))
	}
}
