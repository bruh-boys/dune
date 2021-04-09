package lib

import "testing"

func TestHasPrefix(t *testing.T) {
	v := runTest(t, `	
		function main() {
			let a = "foo2"
			return a.hasPrefix("foo")
		}
	`)

	if v.ToBool() != true {
		t.Fatal(v)
	}
}
