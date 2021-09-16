package ast

import (
	"fmt"
	"strings"
	"testing"
)

func TestBasicLex(t *testing.T) {
	data := []struct {
		s string
		t []Type
	}{
		{"+1", []Type{INT}},
		{"-1", []Type{INT}},
		{"-1.1", []Type{FLOAT}},
		{"1-1", []Type{INT, SUB, INT}},
		{"a-1", []Type{IDENT, SUB, INT}},
		{"-1 // foo", []Type{INT, COMMENT}},
		{"0x33FA3eEE", []Type{HEX}},
		{"0xEE ^ 0xFF", []Type{HEX, XOR, HEX}},
		{"^= 0xFF", []Type{XOR_ASSIGN, HEX}},
		{"|= 0xFF", []Type{BOR_ASSIGN, HEX}},
		{"<< 0xFF", []Type{LSH, HEX}},
		{">> 0xFF", []Type{RSH, HEX}},
		{"~a", []Type{BNT, IDENT}},
		{`print "\""`, []Type{IDENT, STRING}},
		{"test \\ test", []Type{IDENT, RUNE, IDENT}},
		{`"qx\"aa"`, []Type{STRING}},
		{"foo 2344 345.44 true false", []Type{IDENT, INT, FLOAT, TRUE, FALSE}},
		{"for i := 0; i < 10", []Type{FOR, IDENT, DECL, INT, SEMICOLON, IDENT, LSS, INT}},
		{"b++", []Type{IDENT, INC}},
		{"a**b", []Type{IDENT, EXP, IDENT}},
		{"x + --b", []Type{IDENT, ADD, DEC, IDENT}},
		{"i := 1 + b", []Type{IDENT, DECL, INT, ADD, IDENT}},
		{"\"bar \\n  foo\"", []Type{STRING}},
		{"`xxxxx \n  qqqqq`", []Type{STRING}},
		{"// [foo]", []Type{ATTRIBUTE}},
		{`a := 0 // bla bla bla
		  // this is a comment
		  b := 0`, []Type{IDENT, DECL, INT, COMMENT,
			COMMENT, IDENT, DECL, INT}},
		{`/* 
				comment
		  */
		  b := 0`, []Type{MULTILINE_COMMENT, IDENT, DECL, INT}},
		{"1 /* foo */ 2", []Type{INT, MULTILINE_COMMENT, INT}},
	}

	for i, d := range data {
		if err := test(d.s, d.t); err != nil {
			t.Fatalf("test [%d] %v", i, err)
		}
	}
}

func TestLexQuotes(t *testing.T) {
	s := `"\""`
	l := New(strings.NewReader(s), "")
	if err := l.Run(); err != nil {
		t.Fatal(err)
	}

	if len(l.Tokens) != 1 {
		t.Fatal(len(l.Tokens))
	}

	if k := l.Tokens[0]; k.Str != "\"" {
		t.Fatal(k)
	}
}

func TestLexRune(t *testing.T) {
	s := `'a'`
	l := New(strings.NewReader(s), "")
	if err := l.Run(); err != nil {
		t.Fatal(err)
	}

	if len(l.Tokens) != 1 {
		t.Fatal(len(l.Tokens))
	}

	k := l.Tokens[0]

	if k.Type != RUNE {
		t.Fatal(k.Type)
	}

	if k.Str != string('a') {
		t.Fatal(k)
	}
}
func TestLexSingleQuotes(t *testing.T) {
	s := `'\''`
	l := New(strings.NewReader(s), "")
	if err := l.Run(); err != nil {
		t.Fatal(err)
	}

	if len(l.Tokens) != 1 {
		t.Fatal(len(l.Tokens))
	}

	if k := l.Tokens[0]; k.Str != string('\'') {
		t.Fatal(k)
	}
}

func TestLexNewLine(t *testing.T) {
	s := `"\n"`
	l := New(strings.NewReader(s), "")
	if err := l.Run(); err != nil {
		t.Fatal(err)
	}

	if len(l.Tokens) != 1 {
		t.Fatal(len(l.Tokens))
	}

	k := l.Tokens[0]
	if k.Str != "\n" {
		t.Fatal(k)
	}
}

func TestLexEscape(t *testing.T) {
	s := "\\"
	l := New(strings.NewReader(s), "")
	if err := l.Run(); err != nil {
		t.Fatal(err)
	}

	if len(l.Tokens) != 1 {
		t.Fatal(len(l.Tokens))
	}

	k := l.Tokens[0]
	if k.Str != s {
		t.Fatal(k)
	}
}

func TestLexEscape2(t *testing.T) {
	s := `"\\"`
	l := New(strings.NewReader(s), "")
	if err := l.Run(); err != nil {
		t.Fatal(err)
	}

	if len(l.Tokens) != 1 {
		t.Fatal(len(l.Tokens))
	}

	k := l.Tokens[0]
	if k.Str != "\\" {
		t.Fatal(k)
	}
}

func test(s string, types []Type) error {
	l := New(strings.NewReader(s), "")

	if err := l.Run(); err != nil {
		return err
	}

	if len(types) != len(l.Tokens) {
		return fmt.Errorf("found %d tokens, expected %d", len(l.Tokens), len(types))
	}

	for i, t := range types {
		lt := l.Tokens[i]
		if lt.Type != t {
			return fmt.Errorf("%d. found %v, expected %v", i, lt.Type, t)
		}
	}

	return nil
}
