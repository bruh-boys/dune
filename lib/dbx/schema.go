package dbx

import (
	"database/sql"
	"fmt"
)

func (db *DB) Databases() (*Table, error) {
	query := "show databases"

	rows, err := db.queryable().Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ToTable(rows)
}

func (db *DB) HasDatabase(name string) (bool, error) {
	switch db.Driver {
	case "mysql":
		return db.queryBoolean(`SELECT 1
							FROM information_schema.tables
							WHERE table_schema = ?`, name)

	default:
		return false, fmt.Errorf("invalid driver: %s", db.Driver)
	}
}

func (db *DB) HasTable(name string) (bool, error) {
	switch db.Driver {

	case "mysql":
		if db.Database == "" {
			return false, fmt.Errorf("no database specified")
		}
		return db.queryBoolean(`SELECT 1
							FROM information_schema.tables
							WHERE table_schema = ?
							AND table_name = ?`, db.Database, name)

	default:
		return false, fmt.Errorf("invalid driver: %s", db.Driver)
	}
}

func (db *DB) Tables() ([]string, error) {
	switch db.Driver {

	case "mysql":
		q := `SHOW TABLES`
		name := db.Database
		if name != "" {
			q += " FROM " + db.Database
		}
		return db.dbTables(q)

	default:
		return nil, fmt.Errorf("invalid driver: %s", db.Driver)
	}
}

func (db *DB) Columns(table string) ([]SchemaColumn, error) {
	switch db.Driver {
	case "mysql":
		return db.mysqlColumns(table)

	default:
		return nil, fmt.Errorf("invalid driver: %s", db.Driver)
	}
}

type SchemaColumn struct {
	Name     string
	Type     string
	Nullable bool
}

func (db *DB) mysqlColumns(table string) ([]SchemaColumn, error) {
	rows, err := db.queryable().Query("SHOW COLUMNS FROM " + table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []SchemaColumn

	var null string
	var dummy interface{}
	for rows.Next() {
		c := SchemaColumn{}
		err = rows.Scan(&c.Name, &c.Type, &null, &dummy, &dummy, &dummy)
		if err != nil {
			return nil, err
		}
		columns = append(columns, c)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return columns, nil
}

func (db *DB) queryBoolean(query string, args ...interface{}) (bool, error) {
	row := db.queryable().QueryRow(query, args...)

	var i int

	err := row.Scan(&i)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return i > 0, nil
}

func (db *DB) dbTables(query string) ([]string, error) {
	rows, err := db.QueryRaw(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	var name string

	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return tables, nil
}

func IsIdent(s string) bool {
	for i, c := range s {
		if !isIdent(c, i) {
			return false
		}
	}
	return true
}

func isIdent(ch rune, pos int) bool {
	return ch == '_' ||
		'A' <= ch && ch <= 'Z' ||
		'a' <= ch && ch <= 'z' ||
		isDecimal(ch) && pos > 0
}

func isDecimal(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
