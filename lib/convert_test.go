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

	if v.ToString() != "L" {
		t.Fatal(v.ToString())
	}
}

func TestRuneToInt2(t *testing.T) {
	v := runTest(t, `
		let s = "会"
		let r = s.runeAt(0)		
		return convert.toString(r)
	`)

	if v.ToString() != "会" {
		t.Fatal(v.ToString())
	}
}
