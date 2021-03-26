package lib

import "testing"

func TestRuneToInt(t *testing.T) {
	v := runTest(t, `
		let s = "L"
		let r = s.runeAt(0)
		let i = convert.toInt(r)
		r = convert.toRune(i)
		return convert.toString(r)
	`)

	if v.String() != "L" {
		t.Fatal(v.String())
	}
}

func TestRuneToInt2(t *testing.T) {
	v := runTest(t, `
		let s = "会"
		let r = s.runeAt(0)		
		return convert.toString(r)
	`)

	if v.String() != "会" {
		t.Fatal(v.String())
	}
}
