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

	if v.ToString() != "<h1>Hello world!</h1>\n" {
		t.Fatal(v.ToString())
	}
}
