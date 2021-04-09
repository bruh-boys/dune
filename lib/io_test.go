package lib

import (
	"testing"
)

func TestIOBufferWrite(t *testing.T) {
	assertRegister(t, "x", 2, `
		let b = io.newBuffer()
		b.write([0xaa,0xaa])
		let x = b.length
	`)
}
