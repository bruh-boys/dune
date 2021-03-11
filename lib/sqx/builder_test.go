package sqx

import (
	"strings"
	"testing"
)

func TestSelectTables(t *testing.T) {
	q, err := Select("SELECT * FROM foo a")
	if err != nil {
		t.Fatal(err)
	}

	tables := GetTables(q)

	if len(tables) != 1 {
		t.Fatal(len(tables))
	}

	table := tables[0]
	if table.Name != "foo" || table.Alias != "a" {
		t.Fatal(table.Name, table.Alias)
	}
}

func TestSelectTables2(t *testing.T) {
	q, err := Select(`SELECT * 
					  FROM foo a,
					  (SELECT x FROM bar b)`)

	if err != nil {
		t.Fatal(err)
	}

	tables := GetTables(q)

	if len(tables) != 2 {
		t.Fatal(len(tables))
	}

	table := tables[0]
	if table.Name != "foo" || table.Alias != "a" {
		t.Fatal(table.Name, table.Alias)
	}

	table = tables[1]
	if table.Name != "bar" || table.Alias != "b" {
		t.Fatal(table.Name, table.Alias)
	}
}

func TestSelectTables3(t *testing.T) {
	q, err := Select(`SELECT (SELECT x 
							  FROM bar b)
					  FROM (SELECT * 
							FROM foo a
							WHERE y IN (SELECT x 
										FROM test))`)

	if err != nil {
		t.Fatal(err)
	}

	tables := GetTables(q)

	if len(tables) != 3 {
		t.Fatal(len(tables))
	}

	table := tables[0]
	if table.Name != "bar" || table.Alias != "b" {
		t.Fatal(table.Name, table.Alias)
	}

	table = tables[1]
	if table.Name != "foo" || table.Alias != "a" {
		t.Fatal(table.Name, table.Alias)
	}

	table = tables[2]
	if table.Name != "test" || table.Alias != "" {
		t.Fatal(table.Name, table.Alias)
	}
}

func TestInsertTables(t *testing.T) {
	q, err := Parse("INSERT INTO foo SELECT * FROM bar")
	if err != nil {
		t.Fatal(err)
	}

	tables := GetTables(q)

	if len(tables) != 2 {
		t.Fatal(len(tables))
	}

	table := tables[0]
	if table.Name != "foo" || table.Alias != "" {
		t.Fatal(table.Name, table.Alias)
	}

	table = tables[1]
	if table.Name != "bar" || table.Alias != "" {
		t.Fatal(table.Name, table.Alias)
	}
}

func TestUpdateTables(t *testing.T) {
	q, err := Parse("UPDATE foo SET x = 1 WHERE id > (SELECT id FROM bar)")
	if err != nil {
		t.Fatal(err)
	}

	tables := GetTables(q)

	if len(tables) != 2 {
		t.Fatal(len(tables))
	}

	table := tables[0]
	if table.Name != "foo" || table.Alias != "" {
		t.Fatal(table.Name, table.Alias)
	}

	table = tables[1]
	if table.Name != "bar" || table.Alias != "" {
		t.Fatal(table.Name, table.Alias)
	}
}

func TestDeleteTables(t *testing.T) {
	q, err := Parse("DELETE FROM foo WHERE id > (SELECT id FROM bar)")
	if err != nil {
		t.Fatal(err)
	}

	tables := GetTables(q)

	if len(tables) != 2 {
		t.Fatal(len(tables))
	}

	table := tables[0]
	if table.Name != "foo" || table.Alias != "" {
		t.Fatal(table.Name, table.Alias)
	}

	table = tables[1]
	if table.Name != "bar" || table.Alias != "" {
		t.Fatal(table.Name, table.Alias)
	}
}

func TestParams1(t *testing.T) {
	q, err := Select("select id from users where id = ?", 1)
	if err != nil {
		t.Fatal(err)
	}

	s, params, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.EqualFold(s, "select id from users where id = ?") {
		t.Fatal(s)
	}

	assertParams(t, params, 1)
}

func assertParams(t *testing.T, params []interface{}, expected ...interface{}) {
	if len(params) != len(expected) {
		t.Fatalf("Expected %v, got %v", expected, params)
	}
}

func TestAddColumns(t *testing.T) {
	q, err := Select("select id from users")
	if err != nil {
		t.Fatal(err)
	}

	q.AddColumns("name, age")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT id, name, age FROM users" {
		t.Fatal(s)
	}
}

func TestAddColumnsUpdate(t *testing.T) {
	q, err := Parse("UPDATE users")
	if err != nil {
		t.Fatal(err)
	}

	u := q.(*UpdateQuery)

	u.AddColumns("name = ?, age = ?", "bill", 22)

	u.Where("id = ?", 1)

	s, params, err := toSQL(false, u, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE users SET name = ?, age = ? WHERE id = ?" {
		t.Fatal(s)
	}

	if len(params) != 3 || params[0] != "bill" || params[1] != 22 || params[2] != 1 {
		t.Fatal(params)
	}
}

func TestSetColumns(t *testing.T) {
	q, err := Select("select id from users")
	if err != nil {
		t.Fatal(err)
	}

	q.SetColumns("count(*)")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT COUNT(*) FROM users" {
		t.Fatal(s)
	}
}

func TestColNames(t *testing.T) {
	q, err := Select("select id,name,true from users where id > 3 and name != null")
	if err != nil {
		t.Fatal(err)
	}

	n := NameExprColumns(q)
	if len(n) != 4 ||
		n[0].Name != "id" ||
		n[1].Name != "name" ||
		n[2].Name != "id" ||
		n[3].Name != "name" {
		t.Fatal(n)
	}
}

func TestSetColumns2(t *testing.T) {
	q, err := Select("select id from users")
	if err != nil {
		t.Fatal(err)
	}

	err = q.SetColumns("id test1")

	if err == nil {
		t.Error("Expected to fail")
	}

	if !strings.Contains(err.Error(), "Unexpected IDENT 'test1'") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestConcatLimit(t *testing.T) {
	q, err := Select("select * from users")
	if err != nil {
		t.Fatal(err)
	}

	q.Limit(20)

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM users LIMIT 20" {
		t.Fatal(s)
	}
}

func TestConcatOrder(t *testing.T) {
	q, err := Select("select * from users")
	if err != nil {
		t.Fatal(err)
	}

	if err := q.OrderBy("id asc, name desc"); err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM users ORDER BY id ASC, name DESC" {
		t.Fatal(s)
	}
}

func TestConcatLimitOffset(t *testing.T) {
	q, err := Select("select * from users")
	if err != nil {
		t.Fatal(err)
	}

	q.LimitOffset(20, 10)

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM users LIMIT 20, 10" {
		t.Fatal(s)
	}
}

func TestConcatWhere(t *testing.T) {
	q, err := Select("select * from users")
	if err != nil {
		t.Fatal(err)
	}

	if err := q.Where("id=2"); err != nil {
		t.Fatal(err)
	}

	if err := q.And("status=?", 2); err != nil {
		t.Fatal(err)
	}

	if err := q.Or("name=?", 2); err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM users WHERE id = 2 AND status = ? OR name = ?" {
		t.Fatal(s)
	}
}

func TestConcatWhere2(t *testing.T) {
	q, err := Select("select * from users")
	if err != nil {
		t.Fatal(err)
	}

	if err := q.Where("id=2 and (id > 0 and id > 1)"); err != nil {
		t.Fatal(err)
	}

	if err := q.And("status=?", 2); err != nil {
		t.Fatal(err)
	}

	if err := q.Or("name=?", 2); err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM users WHERE id = 2 AND (id > 0 AND id > 1) AND status = ? OR name = ?" {
		t.Fatal(s)
	}
}

func TestConcatjoin(t *testing.T) {
	q, err := Select("select * from users")
	if err != nil {
		t.Fatal(err)
	}

	if err := q.Join("invoices on user.id = invoice.iduser"); err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM users JOIN invoices ON user.id = invoice.iduser" {
		t.Fatal(s)
	}
}

func TestNestQuerys(t *testing.T) {
	q, err := Select("select * from users where id = 1")
	if err != nil {
		t.Fatal(err)
	}

	filter, err := Where("status=2")
	if err != nil {
		t.Fatal(err)
	}

	if err := filter.Or("status=3"); err != nil {
		t.Fatal(err)
	}

	q.AndQuery(filter)

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM users WHERE id = 1 AND (status = 2 OR status = 3)" {
		t.Fatal(s)
	}
}

func TestSelectAlternateJoin(t *testing.T) {
	q, err := Select("select * from t1, t2, t3 where t1.id = t2.id and t1.id != t3.id")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT * FROM t1, t2, t3 WHERE t1.id = t2.id AND t1.id != t3.id" {
		t.Fatal(s)
	}
}

func TestSelectAlternateJoinAlias(t *testing.T) {
	q, err := Select("select t1.id, t2.name, t2.test from table1 t1, table2 t2, table3 t3 where t1.id = t2.id and t1.id != t3.id")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT t1.id, t2.name, t2.test FROM table1 AS t1, table2 AS t2, table3 AS t3 WHERE t1.id = t2.id AND t1.id != t3.id" {
		t.Fatal(s)
	}
}

func TestSelectAlternateJoinAliasAs(t *testing.T) {
	q, err := Select("select t1.id, t2.name, t2.test from table1 as t1, table2 as t2, table3 as t3 where t1.id = t2.id and t1.id != t3.id")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT t1.id, t2.name, t2.test FROM table1 AS t1, table2 AS t2, table3 AS t3 WHERE t1.id = t2.id AND t1.id != t3.id" {
		t.Fatal(s)
	}
}

func TestUpdateAddColumns(t *testing.T) {
	q, err := Parse("update users set a = 1")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := q.(*UpdateQuery)
	if !ok {
		t.Fatal("Not a update query")
	}

	u.AddColumns("b = 2")
	u.AddColumns("c = 3")

	s, _, err := toSQL(false, u, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE users SET a = 1, b = 2, c = 3" {
		t.Fatal(s)
	}
}

func TestUpdateSetColumns(t *testing.T) {
	q, err := Parse("update users set a = 1, b = 2, c = 3")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := q.(*UpdateQuery)
	if !ok {
		t.Fatal("Not a update query")
	}

	u.AddColumns("d = 4")
	u.SetColumns("all = 11")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE users SET all = 11" {
		t.Fatal(s)
	}
}

func TestUpdateWhere(t *testing.T) {
	q, err := Parse("update users set a = 1")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := q.(*UpdateQuery)
	if !ok {
		t.Fatal("Not a update query")
	}

	u.Where("b > 10")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE users SET a = 1 WHERE b > 10" {
		t.Fatal(s)
	}
}

func TestUpdateWhereAnd(t *testing.T) {
	q, err := Parse("update users set a = 1 where b > 10")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := q.(*UpdateQuery)
	if !ok {
		t.Fatal("Not a update query")
	}

	u.And("c < 0")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE users SET a = 1 WHERE b > 10 AND c < 0" {
		t.Fatal(s)
	}
}

func TestUpdateWhereOr(t *testing.T) {
	q, err := Parse("update users set a = 1")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := q.(*UpdateQuery)
	if !ok {
		t.Fatal("Not a update query")
	}
	u.Where("b > 10")
	u.Or("c < 10")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE users SET a = 1 WHERE b > 10 OR c < 10" {
		t.Fatal(s)
	}
}

func TestUpdateJoin(t *testing.T) {
	q, err := Parse("update a set a = 1")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := q.(*UpdateQuery)
	if !ok {
		t.Fatal("Not a update query")
	}

	u.Join("b on a.id = b.idb")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE a JOIN b ON a.id = b.idb SET a = 1" {
		t.Fatal(s)
	}
}

func TestUpdateJoinMultiple(t *testing.T) {
	q, err := Parse("update a left join b on a.id = b.idb set a = 1")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := q.(*UpdateQuery)
	if !ok {
		t.Fatal("Not a update query")
	}

	u.Join("right join c on a.id = c.idc")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE a LEFT JOIN b ON a.id = b.idb RIGHT JOIN c ON a.id = c.idc SET a = 1" {
		t.Fatal(s)
	}
}

func TestUpdateJoinNew(t *testing.T) {
	q, err := Parse("update a set a = 1")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := q.(*UpdateQuery)
	if !ok {
		t.Fatal("Not a update query")
	}

	u.Join("b on a.id = b.idb")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE a JOIN b ON a.id = b.idb SET a = 1" {
		t.Fatal(s)
	}
}

func TestUpdateComplex(t *testing.T) {
	q, err := Parse("update users u inner join b on u.id = b.idb set a = 1")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := q.(*UpdateQuery)
	if !ok {
		t.Fatal("Not a update query")
	}

	u.Join("c on u.id = c.idc")
	u.Join("right join d on c.id = d.id")
	u.Where("u.status = 'a'")
	u.Or("c.status = 'c'")
	u.And("d.id < 100")
	u.AddColumns("z = 4")

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE users AS u INNER JOIN b ON u.id = b.idb "+
		"JOIN c ON u.id = c.idc RIGHT JOIN d ON c.id = d.id "+
		`SET a = 1, z = 4 WHERE u.status = "a" `+
		`OR c.status = "c" AND d.id < 100` {
		t.Fatal(s)
	}
}
