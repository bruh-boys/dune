package dbx

import (
	"database/sql"
)

type Stmt struct {
	*DB
	*sql.Stmt
	query string
}

func (stmt *Stmt) Close() error {
	return stmt.Stmt.Close()
}

func (stmt *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	return stmt.Stmt.Exec(args...)
}

func (stmt *Stmt) Query(args ...interface{}) (*sql.Rows, error) {
	return stmt.Stmt.Query(args...)
}

func (stmt *Stmt) QueryRow(args ...interface{}) *sql.Row {
	return stmt.Stmt.QueryRow(args...)
}
