package lib

import (
	"testing"
)

func TestArraySum(t *testing.T) {
	v := runTest(t, `	
		let a = [2, 3]

		function main() {
			return a.sum()
		}
	`)

	if v.ToInt() != 5 {
		t.Fatal(v)
	}
}

func TestArrayWhere(t *testing.T) {
	v := runTest(t, `	
		let a = [1, 2, 3, 4]

		function main() {
			return a.where(t => t > 2).sum()
		}
	`)

	if v.ToInt() != 7 {
		t.Fatal(v)
	}
}
