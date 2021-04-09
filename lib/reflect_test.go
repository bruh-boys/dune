package lib

import "testing"

func TestGenericIsType(t *testing.T) {
	assertRegister(t, "x", true, `
		let x = reflect.is(3, "int");
	`)
}
