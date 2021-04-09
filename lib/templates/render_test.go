package templates

import (
	"strings"
	"testing"

	"github.com/dunelang/dune"
)

func TestCodeRemovesNewline(t *testing.T) {
	buf, _, err := Compile("<% %>\nFoo")
	if err != nil {
		t.Fatal(err)
	}

	result := string(buf)

	if !strings.Contains(result, "w.write(`Foo`)\n") {
		t.Fatal(result)
	}
}

func TestAttributes(t *testing.T) {
	buf, _, err := Compile(`
	<%@ // [foo] %>
	
	<%@ // [foo bar] %>

	`)
	if err != nil {
		t.Fatal(err)
	}

	result := string(buf)

	p, err := dune.CompileStr(result)
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Attributes) != 2 {
		t.Fatal(p.Attributes)
	}

	if p.Attributes[0] != "foo" {
		t.Fatal(p.Attributes[0])
	}

	if p.Attributes[1] != "foo bar" {
		t.Fatal(p.Attributes[1])
	}
}
