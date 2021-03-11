package lib

import (
	"os"
	"testing"

	"github.com/dunelang/dune"
)

func TestErrorf(t *testing.T) {
	v := runTest(t, `return fmt.errorf("Day {{day}}: {{wrap}}", 2, errors.newError("Snap!"))`)

	err, ok := v.ToObjectOrNil().(*dune.Error)
	if !ok {
		t.Fatalf("Type: %T", v)
	}

	if err.Message() != "Day 2: Snap!" {
		t.Fatal("msg", err.Message())
	}

	if len(err.Wrapped) != 1 {
		t.Fatal("wrap", err.Wrapped)
	}

	if !err.Is("Snap!") {
		t.Fatal("IS", err)
	}
}

func TestErrorIs(t *testing.T) {
	v := runTest(t, `
		let fs = io.newMemFS()

		try {
			fs.open("x")
		} catch(err) {
			return err
		}	
	`)

	err, ok := v.ToObjectOrNil().(*dune.Error)
	if !ok {
		t.Fatalf("Type: %T", v)
	}

	if !err.Is(os.ErrNotExist.Error()) {
		t.Fatal("IS", err)
	}
}

func TestParse(t *testing.T) {
	var tests = []struct {
		text   string
		tokens []Token
	}{
		{"{{ name }}", []Token{{Parameter, "name"}}},
		{"Hi {{ name }}", []Token{{Text, "Hi "}, {Parameter, "name"}}},
		{"Hi {{ name | uppercase }}!", []Token{{Text, "Hi "}, {Parameter, "name | uppercase"}, {Text, "!"}}},
	}

	for i, test := range tests {
		tokens, err := Parse(test.text)
		if err != nil {
			t.Fatalf("Test %d: %s -> %s", i, test.text, err.Error())
		}

		if len(tokens) != len(test.tokens) {
			t.Fatalf("Test %d: expected %d tokens, got %d: %v", i, len(test.tokens), len(tokens), tokens)
		}

		for j, tk := range tokens {
			tt := test.tokens[j]
			if tk.Type != tt.Type || tk.Value != tt.Value {
				t.Fatalf("Test %d: token %d, different: %v", i, j, tk)
			}
		}
	}
}

func TestFormatTemplate(t *testing.T) {
	var tests = []struct {
		text     string
		args     []interface{}
		expected string
	}{
		{"{{ name }}", []interface{}{"Bill"}, "Bill"},
		{"Hi {{ name }}", []interface{}{"Bill"}, "Hi Bill"},
		{"Hi {{ name | uppercase }}!", []interface{}{"Bill"}, "Hi Bill!"},
		{"{{ name }} is {{ age}}", []interface{}{"Bill", 33}, "Bill is 33"},
	}

	for i, test := range tests {
		result := FormatTemplate(test.text, test.args...)
		if test.expected != result {
			t.Fatalf("Test %d: expected %s, got %s", i, test.expected, result)
		}
	}
}
