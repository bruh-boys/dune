package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/dunelang/dune/filesystem"
	_ "github.com/dunelang/dune/lib"

	"github.com/dunelang/dune"
)

func TestTypescript(t *testing.T) {
	var verbose bool
	for _, a := range os.Args {
		if strings.Contains(a, "-test.v=true") {
			verbose = true
			break
		}
	}

	// pass a filter to run only tests that match it. Example:
	//
	//  $ f=While go test -v
	filter := os.Getenv("f")

	files, err := ioutil.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	var fail bool

	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, "_test.ts") {
			continue
		}

		p, err := dune.Compile(filesystem.OS, name)
		if err != nil {
			t.Fatal(err)
		}

		p.AddPermission("trusted")

		for _, fn := range p.Functions {
			if !strings.HasPrefix(fn.Name, "test") {
				continue
			}

			if name != "" && !strings.Contains(fn.Name, filter) {
				continue
			}

			vm := dune.NewVM(p)

			// dune.Print(p)

			if _, err = vm.RunFunc(fn.Name); err != nil {
				fmt.Printf("    FAIL  %s: %v", fn.Name, err)
				fail = true
			}

			if verbose {
				fmt.Printf("    PASS:  %s\n", fn.Name)
			}
		}
	}

	if fail {
		t.Fail()
	}
}
