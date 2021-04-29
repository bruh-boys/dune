package lib

import (
	"testing"

	"github.com/dunelang/dune"

	_ "github.com/mattn/go-sqlite3"
)

func TestQuery1(t *testing.T) {
	v := runTest(t, `
		function main() {
			let db = sql.open("sqlite3", ":memory:")
			db.exec("CREATE TABLE foo (id key, name varchar(10))")
			db.exec("INSERT INTO foo VALUES (1, 'a')")
			db.exec("INSERT INTO foo VALUES (2, 'b')")
			return db.queryValue("SELECT count(*) FROM foo WHERE id in ?", [1,2])
		}
	`)

	if v != dune.NewValue(2) {
		t.Fatal(v)
	}
}

func TestSelectAlternateJoin(t *testing.T) {
	v := runTest(t, `
		function main() {
			let db = sql.open("sqlite3", ":memory:")
			db.exec("CREATE TABLE foo (id key, name varchar(10))")
			db.exec("INSERT INTO foo VALUES (1, 'a')")
			db.exec("INSERT INTO foo VALUES (2, 'b')")
			db.exec("CREATE TABLE bar (id key, name varchar(10))")
			db.exec("INSERT INTO bar VALUES (1, 'c')")
			db.exec("INSERT INTO bar VALUES (2, 'a')")
			db.exec("INSERT INTO bar VALUES (3, 'a')")

			let s = "select count(*) from foo f, bar b"
			let q = sql.parse(s)
			q.where("f.name = b.name")

			return db.queryValue(q)
		}
	`)

	if v != dune.NewValue(2) {
		t.Fatal(v)
	}
}

func TestSQLUpdate(t *testing.T) {
	v := runTest(t, `
		function main() {
			let db = sql.open("sqlite3", ":memory:")
			db.exec("CREATE TABLE foo (id key, name varchar(10))")
			db.exec("INSERT INTO foo VALUES (1, 'a')")
			db.exec("INSERT INTO foo VALUES (2, 'b')")
			db.exec("UPDATE foo SET name = 'c' WHERE id = 1")
			return db.queryValue("SELECT name FROM foo WHERE id = 1")
		}
	`)

	if v != dune.NewValue("c") {
		t.Fatal(v)
	}
}

func TestBuilderUpdate(t *testing.T) {
	v := runTest(t, `
		function main() {
			let db = sql.open("sqlite3", ":memory:")
			db.exec("CREATE TABLE foo (id key, name varchar(10))")
			db.exec("INSERT INTO foo VALUES (1, 'a')")
			db.exec("INSERT INTO foo VALUES (2, 'b')")

			let q = sql.parse("UPDATE foo")
			q.addColumns("name = 'c'")
			q.where('id = 1')

			let s = q.toSQL()
			db.exec(s)

			return db.queryValue("SELECT name FROM foo WHERE id = 1")
		}
	`)

	if v != dune.NewValue("c") {
		t.Fatal(v)
	}
}

func TestBuilderDelete(t *testing.T) {
	v := runTest(t, `
		function main() {
			let db = sql.open("sqlite3", ":memory:")
			db.exec("CREATE TABLE foo (id key, name varchar(10))")
			db.exec("INSERT INTO foo VALUES (1, 'a')")
			db.exec("INSERT INTO foo VALUES (2, 'b')")

			let q = sql.parse("DELETE FROM foo")
			q.where('id = 1')

			let s = q.toSQL()

			let rowsAffected = db.exec(s).rowsAffected
			if(rowsAffected != 1) {
				throw "RowsAffected " + rowsAffected
			}
			
			return db.queryValue("SELECT name FROM foo WHERE id = 1")
		}
	`)

	if v.Type != dune.Null {
		t.Fatal(v)
	}
}

func TestSqliteMultiDB(t *testing.T) {
	runTest(t, `
		function main() {
			let db = sql.open("sqlite3", ":memory:")
			db.initMultiDB()

			db.exec("CREATE DATABASE foo")
			db.database = "foo"
			db.exec("CREATE TABLE product (name varchar(10))")
			db.exec("INSERT INTO product VALUES ('a')")

			
			db.exec("CREATE DATABASE bar")
			db.database = "bar"
			db.exec("CREATE TABLE product (name varchar(10))")
			db.exec("INSERT INTO product VALUES ('b')")

			db.database = null
			assert.equal(["a"], db.queryValues("SELECT name FROM foo.product"))
			assert.equal(["b"], db.queryValues("SELECT name FROM bar.product"))
		}
	`)
}

func TestSqliteShowDatabases(t *testing.T) {
	runTest(t, `
		function main() {
			let db = sql.open("sqlite3", ":memory:")
			db.initMultiDB()

			db.exec("CREATE DATABASE foo")			
			db.exec("CREATE DATABASE bar")
			
			let result = db.query("SHOW DATABASES")
			assert.equal(2, result.length)
			assert.equal("foo", result[0].Database)
			assert.equal("bar", result[1].Database)
		}
	`)
}

func TestSqliteShowColumns(t *testing.T) {
	runTest(t, `
		function main() {
			let db = sql.open("sqlite3", ":memory:")
			db.initMultiDB()

			db.exec("CREATE DATABASE foo")		
			db.exec("CREATE TABLE foo.product (nameA varchar(10))")

			db.exec("CREATE DATABASE bar")
			db.exec("CREATE TABLE bar.product (nameB varchar(10))")
			
			let result = db.query("SHOW COLUMNS FROM foo.product")
			assert.equal(1, result.length)
			assert.equal("nameA", result[0].name)
			
			result = db.query("SHOW COLUMNS FROM bar.product")
			assert.equal(1, result.length)
			assert.equal("nameB", result[0].name)
		}
	`)
}
