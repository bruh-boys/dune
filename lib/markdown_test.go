package lib

import (
	"testing"
)

func TestMarkdown(t *testing.T) {
	v := runTest(t, `
		function main() {
			return markdown.toHTML("# Hello world!\n")
		}
	`)

	if v.String() != "<h1>Hello world!</h1>\n" {
		t.Fatal(v.String())
	}
}
