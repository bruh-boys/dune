package sqx

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestStartOfDay(t *testing.T) {
	q, err := Parse("select START_OF_DAY()")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	tt := time.Now()
	u := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local)

	expect := fmt.Sprintf(`SELECT "%s"`, u.Format("2006-1-2 15:04:05"))
	if s != expect {
		t.Fatal(s, expect)
	}
}

func TestEndOfDay(t *testing.T) {
	q, err := Parse("select END_OF_DAY()")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	tt := time.Now()
	u := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.Local)
	u = u.AddDate(0, 0, 1).Add(-1 * time.Second)

	expect := fmt.Sprintf(`SELECT "%s"`, u.Format("2006-1-2 15:04:05"))
	if s != expect {
		t.Fatal(s, expect)
	}
}

func TestDateAdd(t *testing.T) {
	q, err := Parse("select DATE_ADD(UTC_TIMESTAMP(), INTERVAL 1 DAY)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT DATE_ADD(UTC_TIMESTAMP(), INTERVAL 1 DAY)` {
		t.Fatal(s)
	}
}

func TestGroupConcat(t *testing.T) {
	q, err := Parse("select group_concat(distinct v order by v asc separator ';') from t")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT GROUP_CONCAT(DISTINCT v ORDER BY v ASC SEPARATOR ';') FROM t` {
		t.Fatal(s)
	}
}

func TestSubquery(t *testing.T) {
	q, err := Parse("select x.foo from (select a, b from bar) as x")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT x.foo FROM (SELECT a, b FROM bar) AS x` {
		t.Fatal(s)
	}
}

func TestSubquery2(t *testing.T) {
	q, err := Parse("select a from (select * from bar)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT a FROM (SELECT * FROM bar)` {
		t.Fatal(s)
	}
}

func TestParseFunction(t *testing.T) {
	q, err := Parse("select max(22) from bar group by month(xx)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT MAX(22) FROM bar GROUP BY MONTH(xx)` {
		t.Fatal(s)
	}
}

func TestNullEquality(t *testing.T) {
	q, err := Parse("select * from foo where a != null")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE a IS NOT NULL` {
		t.Fatal(s)
	}
}

func TestNullEquality2(t *testing.T) {
	q, err := Parse("select * from foo where a != ?", nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE a IS NOT NULL` {
		t.Fatal(s)
	}
}

func TestNullEquality3(t *testing.T) {
	q, err := Parse("select * from foo where a != ? and b = ? and c = ?", 1, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE a != ? AND b = ? AND c IS NULL` {
		t.Fatal(s)
	}
}

func TestBitwiseOperator(t *testing.T) {
	q, err := Parse("select * from foo where (b >> ?) & 1", nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE (b >> ?) & 1` {
		t.Fatal(s)
	}
}

func TestBetween(t *testing.T) {
	q, err := Parse("select * from foo where id between ? and ?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE id BETWEEN ? AND ?` {
		t.Fatal(s)
	}
}

func TestBasicSelect(t *testing.T) {
	q, err := Parse("select f.*, bar from foo f")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT f.*, bar FROM foo AS f` {
		t.Fatal(s)
	}
}

func TestReplaceEmptyIN(t *testing.T) {
	var v []interface{}

	q, err := Parse("select * from foo where id in ?", v)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "foo", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo.foo WHERE 1=0` {
		t.Fatal(s)
	}
}

func TestReplaceIN20(t *testing.T) {
	q, err := Parse("select * from foo where id in ?", []interface{}{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "foo", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo.foo WHERE id IN (1, 2, 3)` {
		t.Fatal(s)
	}
}

func TestReplaceIN0(t *testing.T) {
	q, err := Parse("select * from foo where id in ?", []interface{}{"1", "2", "3"})
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "foo", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo.foo WHERE id IN (1, 2, 3)` {
		t.Fatal(s)
	}
}

func TestReplaceIN1(t *testing.T) {
	q, err := Parse("select * from foo where id in ?", []interface{}{1, true, 3, "1-2_33"})
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = toSQL(false, q, "foo", "")
	if err == nil {
		t.Fatal("Expected error. Only ints are valid")
	}
}

func TestReplaceIN11(t *testing.T) {
	q, err := Parse("select * from foo where id in ?", []interface{}{1, true, 3, "1'2"})
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = toSQL(false, q, "foo", "")
	if err == nil {
		t.Fatal("Expected error. Only ints are valid")
	}
}

func TestReplaceIN2(t *testing.T) {
	params := []interface{}{
		10,
		[]interface{}{1, 2, 3},
	}

	q, err := Parse("select * from foo where id > ? AND id in ?", params...)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "foo", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo.foo WHERE id > ? AND id IN (1, 2, 3)` {
		t.Fatal(s)
	}
}

func TestReplaceIN3(t *testing.T) {
	params := []interface{}{
		[]interface{}{1, 2, 3},
		10,
	}

	q, err := Parse("select * from foo where id in ? AND id > ?", params...)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "foo", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo.foo WHERE id IN (1, 2, 3) AND id > ?` {
		t.Fatal(s)
	}
}

func TestReplaceIN4(t *testing.T) {
	params := []interface{}{
		20,
		[]interface{}{1, 2, 3},
		10,
	}

	q, err := Parse("select * from foo where id < ? and id in ? AND id > ?", params...)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "foo", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo.foo WHERE id < ? AND id IN (1, 2, 3) AND id > ?` {
		t.Fatal(s)
	}
}

func TestDbExplicit(t *testing.T) {
	q, err := Parse("select * from cars")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "foo", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo.cars` {
		t.Fatal(s)
	}
}

func TestDbExplicit2(t *testing.T) {
	q, err := Parse("select id from cars")
	if err != nil {
		t.Fatal(err)
	}

	w := NewWriter(q, "foo", "", "", "mysql")
	w.EscapeIdents = true

	s, _, err := w.Write()
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT `id` FROM `foo`.`cars`" {
		t.Fatal(s)
	}
}

func TestDbExplicit3(t *testing.T) {
	q, err := Parse("select id from cars")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "foo", "mysql")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT id FROM foo.cars` {
		t.Fatal(s)
	}
}

func TestDbPrefix(t *testing.T) {
	q, err := Parse("select id from cars")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQLPrefix(q, "foo", "test_", "mysql")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT id FROM test_foo.cars` {
		t.Fatal(s)
	}
}

func TestParseSelectCount(t *testing.T) {
	q, err := Parse("select count(*) from cars")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT COUNT(*) FROM cars` {
		t.Fatal(s)
	}
}

func TestParseSelectOR(t *testing.T) {
	q, err := Parse("select * from foo where a LIKE ? or b LIKE ?", 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE a LIKE ? OR b LIKE ?` {
		t.Fatal(s)
	}
}

func TestParseSelect(t *testing.T) {
	q, err := Parse("select * from cars")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM cars` {
		t.Fatal(s)
	}
}

func TestParseSelect1(t *testing.T) {
	q, err := Parse("select 1 as num,true, false, null, 'te\"st'")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 AS num, true, false, null, "te\"st"` {
		t.Fatal(s)
	}
}

func TestParseSelect2(t *testing.T) {
	q, err := Parse("select (1+2)*-5")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT (1 + 2) * -5` {
		t.Fatal(s)
	}
}

func TestParseSelect3(t *testing.T) {
	q, err := Parse("select 1 from (select 1)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM (SELECT 1)` {
		t.Fatal(s)
	}
}

func TestParseSelect4(t *testing.T) {
	q, err := Parse("select id from c order by name, age desc limit 3,4")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT id FROM c ORDER BY name, age DESC LIMIT 3, 4` {
		t.Fatal(s)
	}
}

func TestParseSelect5(t *testing.T) {
	q, err := Parse("select id from c limit ?", nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT id FROM c LIMIT ?` {
		t.Fatal(s)
	}
}

func TestParseSelect6(t *testing.T) {
	q, err := Parse("select id from c limit ?,?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT id FROM c LIMIT ?, ?` {
		t.Fatal(s)
	}
}

func TestParseJoin(t *testing.T) {
	q, err := Parse("select id from a join b on a.id = b.id")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT id FROM a JOIN b ON a.id = b.id` {
		t.Fatal(s)
	}
}

func TestParseJoin2(t *testing.T) {
	q, err := Parse("select id from a left join b on a.id = b.id")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT id FROM a LEFT JOIN b ON a.id = b.id` {
		t.Fatal(s)
	}
}

func TestParseJoin3(t *testing.T) {
	q, err := Parse("select a.id, b.* from a left join b on a.id = b.id")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT a.id, b.* FROM a LEFT JOIN b ON a.id = b.id` {
		t.Fatal(s)
	}
}

func TestParseWhere(t *testing.T) {
	q, err := Parse("select 1 from foo where true and (id < 3)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM foo WHERE true AND (id < 3)` {
		t.Fatal(s)
	}
}

func TestParseWhereIs(t *testing.T) {
	q, err := Parse("select 1 is null, 1 is not null")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 IS null, 1 IS NOT null` {
		t.Fatal(s)
	}
}

func TestParseLike(t *testing.T) {
	q, err := Parse("select * from foo where name like ?", nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE name LIKE ?` {
		t.Fatal(s)
	}
}

func TestParseLike2(t *testing.T) {
	q, err := Parse("select * from foo where name not like ?", nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE name NOT LIKE ?` {
		t.Fatal(s)
	}
}

func TestParseWhere2(t *testing.T) {
	q, err := Parse("select 1 from foo where (select id from x) > 1")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "z", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM z.foo WHERE (SELECT id FROM z.x) > 1` {
		t.Fatal(s)
	}
}

func TestParseSelectDbPreffix1(t *testing.T) {
	q, err := Parse("select a from customers c, payments p")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "foo", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT a FROM foo.customers AS c, foo.payments AS p" {
		t.Fatal(s)
	}
}

func TestParseSelectDbPreffix2(t *testing.T) {
	q, err := Parse("select a from xx.customers c")
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = toSQL(false, q, "foo", "")
	if err == nil {
		t.Fatal("Should fail because it has a database already set")
	}
}

func TestParseSelectDbPreffix3(t *testing.T) {
	q, err := Parse("select (select id from foo) from bar")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "db", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT (SELECT id FROM db.foo) FROM db.bar" {
		t.Fatal(s)
	}
}

func TestParseGroupBy(t *testing.T) {
	q, err := Parse("select 1 from foo group by a,b")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM foo GROUP BY a, b` {
		t.Fatal(s)
	}
}

func TestParseWhereIN(t *testing.T) {
	q, err := Parse("select 1 from foo where id in (1,2)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM foo WHERE id IN (1, 2)` {
		t.Fatal(s)
	}
}

func TestParseWhereIN1(t *testing.T) {
	q, err := Parse("select 1 from foo where id in ?", []interface{}{9})
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM foo WHERE id IN (9)` {
		t.Fatal(s)
	}
}

func TestParseWhereIN2(t *testing.T) {
	q, err := Parse("select 1 from foo where id in ('aa', 'bb')")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM foo WHERE id IN ("aa", "bb")` {
		t.Fatal(s)
	}
}

func TestParseWhereIN3(t *testing.T) {
	q, err := Parse("select 1 from foo where id in (1+2)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM foo WHERE id IN (1 + 2)` {
		t.Fatal(s)
	}
}

func TestParseWhereIN4(t *testing.T) {
	q, err := Parse("select 1 from foo where id in (select 1)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM foo WHERE id IN (SELECT 1)` {
		t.Fatal(s)
	}
}

func TestParseWhereIN5(t *testing.T) {
	q, err := Parse("select 1 from foo where id in ((select id from foo))")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM foo WHERE id IN ((SELECT id FROM foo))` {
		t.Fatal(s)
	}
}

func TestParseWhereIN6(t *testing.T) {
	q, err := Parse("select 1 from foo where id not in (2,3)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT 1 FROM foo WHERE id NOT IN (2, 3)` {
		t.Fatal(s)
	}
}

func TestParseSelectFunc(t *testing.T) {
	q, err := Parse("select now()")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT NOW()` {
		t.Fatal(s)
	}
}

func TestParseSelectFunc2(t *testing.T) {
	q, err := Parse("select * from foo where d >= now()")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo WHERE d >= NOW()` {
		t.Fatal(s)
	}
}

func TestForUpdate(t *testing.T) {
	q, err := Parse("select * from foo for update")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `SELECT * FROM foo FOR UPDATE` {
		t.Fatal(s)
	}
}

func TestParseDelete(t *testing.T) {
	q, err := Parse("delete from foo where x = 'foo' and r = 'bar' limit 3")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "z", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `DELETE FROM z.foo WHERE x = "foo" AND r = "bar" LIMIT 3` {
		t.Fatal(s)
	}
}

func TestParseUpdate1(t *testing.T) {
	q, err := Parse("update foo set x=3")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE foo SET x = 3" {
		t.Fatal(s)
	}
}

func TestParseUpdate2(t *testing.T) {
	q, err := Parse("update foo set x = (3+2) where id >= 10 limit 2")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "UPDATE foo SET x = (3 + 2) WHERE id >= 10 LIMIT 2" {
		t.Fatal(s)
	}
}

func TestParseUpdate3(t *testing.T) {
	q, err := Parse("update post set title = concat(title, '-Z')")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE post SET title = CONCAT(title, "-Z")` {
		t.Fatal(s)
	}
}

func TestParseUpdate4(t *testing.T) {
	q, err := Parse("UPDATE Employee SET password=?,webPunch=?,status=? WHERE id=?", 1, 2, 3, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE Employee SET password = ?, webPunch = ?, status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateJoin(t *testing.T) {
	q, err := Parse("UPDATE a JOIN b ON a.id = b.ida SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE a JOIN b ON a.id = b.ida SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateJoinLeft(t *testing.T) {
	q, err := Parse("UPDATE a left JOIN b ON a.id = b.ida SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE a LEFT JOIN b ON a.id = b.ida SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateJoinRight(t *testing.T) {
	q, err := Parse("UPDATE a right JOIN b ON a.id = b.ida SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE a RIGHT JOIN b ON a.id = b.ida SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateJoinOuter(t *testing.T) {
	q, err := Parse("UPDATE a outer JOIN b ON a.id = b.ida SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE a OUTER JOIN b ON a.id = b.ida SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateJoinInner(t *testing.T) {
	q, err := Parse("UPDATE a INNER JOIN b SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE a INNER JOIN b SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateJoinCross(t *testing.T) {
	q, err := Parse("UPDATE a CROSS JOIN b SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE a CROSS JOIN b SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateJoinDouble(t *testing.T) {
	q, err := Parse("UPDATE a JOIN b ON a.id = b.ida JOIN c ON b.id = c.idb SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE a JOIN b ON a.id = b.ida JOIN c ON b.id = c.idb SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateJoinMixed(t *testing.T) {
	q, err := Parse("UPDATE a RIGHT JOIN b ON a.id = b.ida OUTER JOIN c ON b.id = c.idb SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE a RIGHT JOIN b ON a.id = b.ida OUTER JOIN c ON b.id = c.idb SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateAlias(t *testing.T) {
	q, err := Parse("UPDATE aa a JOIN bb b ON a.id = b.ida SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE aa AS a JOIN bb AS b ON a.id = b.ida SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseUpdateAliasAs(t *testing.T) {
	q, err := Parse("UPDATE aa AS a JOIN bb AS b ON a.id = b.ida SET status=? WHERE id=?", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `UPDATE aa AS a JOIN bb AS b ON a.id = b.ida SET status = ? WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseDelete2(t *testing.T) {
	q, err := Parse("DELETE FROM Employee WHERE id=?", nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `DELETE FROM Employee WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseDeleteJoin(t *testing.T) {
	q, err := Parse("DELETE a, b FROM a JOIN b ON a.id = bd.id")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `DELETE a, b FROM a JOIN b ON a.id = bd.id` {
		t.Fatal(s)
	}
}

func TestParseDeleteJoinLeft(t *testing.T) {
	q, err := Parse("DELETE a,b FROM a left JOIN b ON a.id = b.ida WHERE id=?", nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `DELETE a, b FROM a LEFT JOIN b ON a.id = b.ida WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseDeleteJoinRight(t *testing.T) {
	q, err := Parse("DELETE a,b FROM a right JOIN b ON a.id = b.ida WHERE id=?", nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `DELETE a, b FROM a RIGHT JOIN b ON a.id = b.ida WHERE id IS NULL` {
		t.Fatal(s)
	}
}

func TestParseDeleteJoinDouble(t *testing.T) {
	q, err := Parse("DELETE a,b,c FROM a JOIN b ON a.id = b.ida JOIN c ON b.id = c.idb WHERE id>5")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `DELETE a, b, c FROM a JOIN b ON a.id = b.ida JOIN c ON b.id = c.idb WHERE id > 5` {
		t.Fatal(s)
	}
}

func TestParseDeleteJoinMixed(t *testing.T) {
	q, err := Parse("DELETE a,b,c FROM a LEFT JOIN b ON a.id = b.ida RIGHT JOIN c ON b.id = c.idb WHERE id>5")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != `DELETE a, b, c FROM a LEFT JOIN b ON a.id = b.ida RIGHT JOIN c ON b.id = c.idb WHERE id > 5` {
		t.Fatal(s)
	}
}

func TestParseInsert1(t *testing.T) {
	q, err := Parse("insert into foo values (3, 4)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "INSERT INTO foo VALUES (3, 4)" {
		t.Fatal(s)
	}
}

func TestParseInsert2(t *testing.T) {
	q, err := Parse("insert into foo (id, id2) values (3, 4)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "INSERT INTO foo (id, id2) VALUES (3, 4)" {
		t.Fatal(s)
	}
}

func TestParseInsert3(t *testing.T) {
	q, err := Parse("insert into foo values (3, 4)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "x", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "INSERT INTO x.foo VALUES (3, 4)" {
		t.Fatal(s)
	}
}

func TestParseInsert4(t *testing.T) {
	q, err := Parse("insert into foo values (?, ?)", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "INSERT INTO foo VALUES (?, ?)" {
		t.Fatal(s)
	}
}

func TestParseInsert5(t *testing.T) {
	q, err := Parse("insert into foo values (default)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQL(false, q, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "INSERT INTO foo VALUES (default)" {
		t.Fatal(s)
	}
}

func TestNamespace1(t *testing.T) {
	q, err := Parse("insert into foo values (default)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQLNamespace(false, q, "", "admin", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "INSERT INTO admin_foo VALUES (default)" {
		t.Fatal(s)
	}
}

func TestNamespace2(t *testing.T) {
	q, err := Parse("create table if not exists cars (weight int)")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQLNamespace(false, q, "", "admin", "")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(s, "CREATE TABLE IF NOT EXISTS admin_cars") {
		t.Fatal(s)
	}
}
func TestNamespace3(t *testing.T) {
	q, err := Parse("select id from a join b on a.id = b.id")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQLNamespace(false, q, "", "admin", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT id FROM admin_a JOIN admin_b ON a.id = b.id" {
		t.Fatal(s)
	}
}

func TestNamespace4(t *testing.T) {
	q, err := Parse("select id from a join foo_b on a.id = b.id")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQLNamespace(false, q, "", "admin", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT id FROM admin_a JOIN foo_b ON a.id = b.id" {
		t.Fatal(s)
	}
}

func TestNamespace5(t *testing.T) {
	q, err := Parse("select x.foo from (select a, b from bar) as x")
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := toSQLNamespace(false, q, "", "admin", "")
	if err != nil {
		t.Fatal(err)
	}

	if s != "SELECT x.foo FROM (SELECT a, b FROM admin_bar) AS x" {
		t.Fatal(s)
	}
}

func toSQLPrefix(q Query, database, prefix, driver string) (string, []interface{}, error) {
	w := NewWriter(q, database, prefix, "", driver)
	w.EscapeIdents = false
	return w.Write()
}

func toSQLNamespace(format bool, q Query, database, namespace, driver string) (string, []interface{}, error) {
	w := NewWriter(q, database, "", namespace, driver)
	w.EscapeIdents = false
	w.Format = format
	return w.Write()
}

func toSQL(format bool, q Query, database, driver string) (string, []interface{}, error) {
	w := NewWriter(q, database, "", "", driver)
	w.EscapeIdents = false
	w.Format = format
	return w.Write()
}
