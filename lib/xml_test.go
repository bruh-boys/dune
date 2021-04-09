package lib

import (
	"strings"
	"testing"
)

func TestXMLWrite(t *testing.T) {
	vm, err := runExpr(t, `
		let doc = xml.newDocument()
		let people = doc.createElement("People")
		let person = people.createElement("Person")
		person.createAttribute("age", "33");
		person.setValue("John")
		let x = doc.string()
	`)
	if err != nil {
		t.Fatal(err)
	}

	v, _ := vm.RegisterValue("x")
	expected := `<People>
  <Person age="33">John</Person>
</People>`

	if !strings.Contains(v.String(), expected) {
		t.Fatalf("Unexpected XML:\n%v", v)
	}
}
