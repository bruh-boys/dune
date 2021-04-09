package lib

import (
	"strings"
	"testing"

	"github.com/dunelang/dune/lib/templates"
)

func Test7(t *testing.T) {
	source := `<%@ 	
var a = 1
var b = 2
%>
`

	expect := `var a = 1
var b = 2

`
	p, _, err := templates.CompileHtml(source)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(p), expect) {
		t.Fatal(string(p))
	}
}
