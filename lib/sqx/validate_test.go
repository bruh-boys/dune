package sqx

import (
	"strings"
	"testing"
)

func TestValidateIN(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
		},
	}

	expected := ""

	query := `SELECT a FROM foo WHERE a IN (1,2,3)`

	assertSelect(t, opt, expected, query)
}

func TestValidateIN2(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
		},
	}

	expected := "invalid IN expression: subquery"

	query := `SELECT a FROM foo WHERE a IN (SELECT * FROM bar)`

	assertSelect(t, opt, expected, query)
}

func TestValidateDatabase(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
		},
	}

	expected := "should not specify a database"

	query := `SELECT a FROM db.foo`

	assertSelect(t, opt, expected, query)
}

func TestValidateDatabase2(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
		},
	}

	expected := "should not specify a database"

	query := `SELECT (SELECT z.a FROM db.foo z) FROM foo`

	assertSelect(t, opt, expected, query)
}

func TestValidateDatabase3(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
		},
	}

	expected := "should not specify a database"

	query := `SELECT x.a FROM foo x WHERE x.a = (SELECT z.a FROM db.foo z)`

	assertSelect(t, opt, expected, query)
}

func TestValidateSelectTable(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name: "foo",
			},
		},
	}

	assertSelect(t, opt, "invalid table", `SELECT a FROM bar`)
}

func TestValidateColumn1(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name: "foo",
			},
		},
	}

	assertSelect(t, opt, "invalid table", `SELECT bar.b FROM foo`)
}
func TestValidateColumn2(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
		},
	}

	assertSelect(t, opt, "", `SELECT a FROM foo`)
}

func TestValidateColumn3(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"b"},
			},
		},
	}

	assertSelect(t, opt, "not found", `SELECT a FROM foo`)
}

func TestValidateColumn4(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
			{
				Name:    "bar",
				Columns: []string{"b"},
			},
		},
	}

	expected := "not found 'bar.a'"

	query := `SELECT a.a, 
					b.a 
				FROM foo a 
				JOIN bar b`

	assertSelect(t, opt, expected, query)
}

func TestValidateSubquery(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"aaa"},
			},
			{
				Name:    "bar",
				Columns: []string{"bbb"},
			},
		},
	}

	expected := "not found 'bar.xx'"

	query := `SELECT (SELECT p.xx FROM bar p) FROM foo a`

	assertSelect(t, opt, expected, query)
}

func TestValidateFunction(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
		},
	}

	expected := ""

	query := `SELECT SUM(a, 12, "asdf", true) FROM foo`

	assertSelect(t, opt, expected, query)
}

func TestValidateFunction2(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
		},
	}

	expected := "invalid argument: not found 'foo.b'"

	query := `SELECT SUM(x.a, (SELECT SUM(z.b) FROM foo z)) FROM foo x`

	assertSelect(t, opt, expected, query)
}
func TestValidateFunction3(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"a"},
			},
		},
	}

	expected := "invalid function"

	query := `SELECT FILE() FROM foo`

	assertSelect(t, opt, expected, query)
}

func TestValidateWhere(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"aaa"},
			},
			{
				Name:    "bar",
				Columns: []string{"bbb"},
			},
		},
	}

	assertSelect(t, opt, "all columns must be qualified", `SELECT a.aaa, 
									b.bbb 
							FROM foo a 
							JOIN bar b
							WHERE aaa = 3`)
}

func TestValidateWhere2(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"aaa"},
			},
			{
				Name:    "bar",
				Columns: []string{"bbb"},
			},
		},
	}

	assertSelect(t, opt, "not found 'bar.aaa'", `SELECT a.aaa, 
									b.bbb 
							FROM foo a 
							JOIN bar b
							WHERE b.aaa = 3`)
}

func TestValidateOrder1(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"aaa"},
			},
			{
				Name:    "bar",
				Columns: []string{"bbb"},
			},
		},
	}

	assertSelect(t, opt, "not found 'bar.aaa'", `SELECT a.aaa, 
									b.bbb 
							FROM foo a 
							JOIN bar b
							ORDER BY b.aaa`)
}

func TestValidateLimit(t *testing.T) {
	opt := &ValidateOptions{
		Tables: []*ValidateTable{
			{
				Name:    "foo",
				Columns: []string{"*"},
			},
		},
	}

	assertSelect(t, opt, "invalid range", `SELECT a FROM foo LIMIT 1000000`)
}

func assertSelect(t *testing.T, options *ValidateOptions, expectedErr string, query string) {
	q, err := Parse(query)
	if err != nil {
		t.Fatal(err)
	}

	s, ok := q.(*SelectQuery)
	if !ok {
		t.Fatalf("not a select query: %T", q)
	}

	err = ValidateSelect(s, options)
	if err == nil {
		if expectedErr != "" {
			t.Fatalf(`expected an error "%s" but got Success`, expectedErr)
		}
		return
	}

	if expectedErr == "" && err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(err.Error(), expectedErr) {

		t.Fatalf("expected error %s, got %v", expectedErr, err.Error())
	}
}
