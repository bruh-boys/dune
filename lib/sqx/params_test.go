package sqx

import "testing"

func TestParseSelectParams(t *testing.T) {
	p := NewStrParser("select * from foo where name like 'bar'")
	p.ReplaceParams = true

	q, err := p.ParseQuery()
	if err != nil {
		t.Fatal(err)
	}

	_, ok := q.(*SelectQuery)
	if !ok {
		t.Fatal("The query is not a Select")
	}

	s, params, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE name LIKE ?` {
		t.Fatal(s)
	}

	if len(params) != 1 || params[0] != "bar" {
		t.Fatal(params)
	}
}

func TestParseSelectParams2(t *testing.T) {
	p := NewStrParser("select id from a where b=\"\\\"\"")
	p.ReplaceParams = true

	q, err := p.ParseQuery()
	if err != nil {
		t.Fatal(err)
	}

	_, ok := q.(*SelectQuery)
	if !ok {
		t.Fatal("The query is not a Select")
	}

	s, params, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT id FROM a WHERE b = ?` {
		t.Fatal(s)
	}

	if len(params) != 1 || params[0] != "\"" {
		t.Fatal(params)
	}
}

func TestParseSelectParams3(t *testing.T) {
	p := NewStrParser(`select id from a where b="\""`)
	p.ReplaceParams = true

	q, err := p.ParseQuery()
	if err != nil {
		t.Fatal(err)
	}

	_, ok := q.(*SelectQuery)
	if !ok {
		t.Fatal("The query is not a Select")
	}

	s, params, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT id FROM a WHERE b = ?` {
		t.Fatal(s)
	}

	if len(params) != 1 || params[0] != "\"" {
		t.Fatal(params)
	}
}
