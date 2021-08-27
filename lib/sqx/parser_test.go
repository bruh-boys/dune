package sqx

import (
	"strings"
	"testing"
)

func TestParseNot(t *testing.T) {
	q, err := Parse("UPDATE type SET foo = not foo")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE type SET foo = NOT foo" {
		t.Fatal(s)
	}
}

func TestParseNot2(t *testing.T) {
	q, err := Parse("UPDATE type SET foo = !foo")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE type SET foo = !foo" {
		t.Fatal(s)
	}
}

func TestParseUpdateAND(t *testing.T) {
	_, err := Parse("UPDATE type SET weekDays = 0 AND holiday = ? WHERE id = ?")
	if err == nil {
		t.Fatal("Expetected to fail")
	}

	if !strings.Contains(err.Error(), "Unexpected 'AND'") {
		t.Fatal(err)
	}
}

func TestParseUpdateValues2(t *testing.T) {
	q, err := Parse("UPDATE type SET weekDays = (1 + 2)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE type SET weekDays = (1 + 2)" {
		t.Fatal(s)
	}
}

func TestParseFromSubquery(t *testing.T) {
	q, err := Parse("SELECT * FROM (SELECT 1) as a")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM (SELECT 1) AS a" {
		t.Fatal(s)
	}
}

func TestParseJoinSubquery(t *testing.T) {
	q, err := Parse("select * from foo a join (select name from bar) as b")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM foo AS a JOIN (SELECT name FROM bar) AS b" {
		t.Fatal(s)
	}
}
