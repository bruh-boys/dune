package dbx

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/dunelang/dune/lib/sqx"
)

var ErrReadOnly = errors.New("error 1299. Can't write changes in Read-Only mode")

type DB struct {
	sync.Mutex
	*sql.DB
	Driver    string
	Database  string
	Namespace string
	Prefix    string
	ReadOnly  bool
	NestedTx  bool
	nestedTx  int // to keep track of the number of nested transactions
	tx        *sql.Tx
}

func Open(driver, dsn string) (*DB, error) {
	return OpenDatabase("", driver, dsn)
}

// OpenDatabase opens a new database handle.
func OpenDatabase(database, driver, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	dx := &DB{
		Driver:   driver,
		Database: database,
		DB:       db,
	}

	return dx, nil
}

// Open returns a new db handler for the database.
func (db *DB) Open(database string) *DB {
	return &DB{
		Database: database,
		Driver:   db.Driver,
		Prefix:   db.Prefix,
		ReadOnly: db.ReadOnly,
		NestedTx: db.NestedTx,
		DB:       db.DB,
	}
}

// Copies the database but without transaction information
func (db *DB) Clone() *DB {
	return db.Open(db.Database)
}

func (db *DB) HasTransaction() bool {
	db.Lock()
	v := db.tx != nil
	db.Unlock()
	return v
}

type connection interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func (db *DB) connection() connection {
	if db.tx != nil {
		return db.tx
	}
	return db.DB
}

func (db *DB) Begin() error {
	db.Lock()
	defer db.Unlock()

	if db.NestedTx {
		if db.tx == nil && db.nestedTx > 0 {
			return fmt.Errorf("sqx: Previous transaction still open")
		}

		if db.tx != nil {
			db.nestedTx++
			return nil
		}

		db.nestedTx++
	} else {
		if db.tx != nil {
			return fmt.Errorf("there is a transaction open")
		}
	}

	t, err := db.DB.Begin()
	if err != nil {
		return err
	}

	db.tx = t
	return nil
}

func (db *DB) Rollback() error {
	db.Lock()
	defer db.Unlock()

	if db.NestedTx {
		if db.tx != nil {
			err := db.tx.Rollback()
			db.tx = nil
			if err != nil {
				return err
			}
		}

		if db.nestedTx == 0 {
			// there is an error. It should have an nested value
			return fmt.Errorf("no transaction open.3")
		}

		db.nestedTx--
	} else {
		if db.tx == nil {
			return fmt.Errorf("no transaction open")
		}

		err := db.tx.Rollback()
		db.tx = nil
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) Commit() error {
	db.Lock()
	defer db.Unlock()

	if db.NestedTx {
		if db.nestedTx == 0 {
			return fmt.Errorf("no transaction open")
		}

		db.nestedTx--

		if db.nestedTx > 0 {
			return nil
		}
	}

	if db.tx == nil {
		return fmt.Errorf("no transaction open")
	}

	err := db.tx.Commit()
	db.tx = nil
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) Prepare(query string) (*Stmt, error) {
	stmt, err := db.connection().Prepare(query)
	if err != nil {
		return nil, err
	}

	return &Stmt{DB: db, Stmt: stmt, query: query}, nil
}

func (db *DB) ToSql(q sqx.Query) (string, []interface{}, error) {
	s, params, err := sqx.ToSql(q, db.Database, db.Prefix, db.Namespace, db.Driver)
	if err != nil {
		return "", nil, err
	}
	return s, params, err
}

func (db *DB) ExecRaw(query string, args ...interface{}) (sql.Result, error) {
	if db.ReadOnly {
		return nil, ErrReadOnly
	}

	q := db.connection()
	r, err := q.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (db *DB) QueryRaw(query string, args ...interface{}) (*sql.Rows, error) {
	r, err := db.connection().Query(query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return r, nil
}

func (db *DB) QueryRowRaw(query string, args ...interface{}) *sql.Row {
	return db.connection().QueryRow(query, args...)
}

func (db *DB) ScanValueRaw(v interface{}, query string, args ...interface{}) error {
	return db.connection().QueryRow(query, args...).Scan(v)
}

type Reader struct {
	columns []*Column
	rows    *sql.Rows
	values  []interface{}
}

func (r *Reader) Columns() ([]*Column, error) {
	if r.columns == nil {
		cols, err := getColumns(r.rows)
		if err != nil {
			return nil, err
		}
		r.columns = cols
	}
	return r.columns, nil
}

func (r *Reader) Next() bool {
	return r.rows.Next()
}

func (r *Reader) Read() ([]interface{}, error) {
	if r.values == nil {
		cols, err := r.rows.Columns()
		if err != nil {
			return nil, err
		}
		r.values = make([]interface{}, len(cols))
	}

	for i := range r.values {
		r.values[i] = &r.values[i]
	}

	if err := r.rows.Scan(r.values...); err != nil {
		return nil, err
	}

	cols, err := r.Columns()
	if err != nil {
		return nil, err
	}

	for i, v := range r.values {
		val, err := Convert(v, cols[i].Type)
		if err != nil {
			return nil, fmt.Errorf("error converting %s: %v", cols[i].Name, err)
		}
		r.values[i] = val
	}

	return r.values, nil
}

func (r *Reader) Err() error {
	return r.rows.Err()
}

func (r *Reader) Close() error {
	return r.rows.Close()
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

func (db *DB) ShowQuery(query string) (*Table, error) {
	q, err := sqx.Parse(query)
	if err != nil {
		return nil, err
	}

	sq, ok := q.(*sqx.ShowQuery)
	if !ok {
		return nil, fmt.Errorf("not a show query")
	}

	return db.ShowQueryEx(sq)
}

func (db *DB) ShowQueryEx(query *sqx.ShowQuery) (*Table, error) {
	s, _, err := db.ToSql(query)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryRaw(s)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ToTable(rows)
}

func (db *DB) ShowReader(query string) (*Reader, error) {
	q, err := sqx.Parse(query)
	if err != nil {
		return nil, err
	}

	sq, ok := q.(*sqx.ShowQuery)
	if !ok {
		return nil, fmt.Errorf("not a show query")
	}

	return db.ShowReaderEx(sq)
}

func (db *DB) ShowReaderEx(query *sqx.ShowQuery) (*Reader, error) {
	s, _, err := db.ToSql(query)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryRaw(s)
	if err != nil {
		return nil, err
	}

	return &Reader{rows: rows}, nil
}

func (db *DB) ReaderEx(query *sqx.SelectQuery) (*Reader, error) {
	rows, err := db.QueryRowsEx(query)
	if err != nil {
		return nil, err
	}

	return &Reader{rows: rows}, nil
}

func (db *DB) Reader(query string, args ...interface{}) (*Reader, error) {
	rows, err := db.QueryRows(query, args...)
	if err != nil {
		return nil, err
	}

	return &Reader{rows: rows}, nil
}

func (db *DB) ReaderRaw(query string, args ...interface{}) (*Reader, error) {
	rows, err := db.QueryRaw(query, args...)
	if err != nil {
		return nil, err
	}

	return &Reader{rows: rows}, nil
}

func (db *DB) QueryEx(query *sqx.SelectQuery) (*Table, error) {
	rows, err := db.QueryRowsEx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ToTable(rows)
}

func (db *DB) QueryRowsEx(q *sqx.SelectQuery) (*sql.Rows, error) {
	s, params, err := db.ToSql(q)
	if err != nil {
		return nil, err
	}

	return db.QueryRaw(s, params...)
}

func (db *DB) Query(query string, args ...interface{}) (*Table, error) {
	rows, err := db.QueryRows(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ToTable(rows)
}

func (db *DB) QueryRows(query string, args ...interface{}) (*sql.Rows, error) {
	q, err := sqx.Parse(query, args...)
	if err != nil {
		return nil, err
	}

	sq, ok := q.(*sqx.SelectQuery)
	if !ok {
		return nil, fmt.Errorf("not a select query")
	}

	s, params, err := db.ToSql(sq)
	if err != nil {
		return nil, err
	}

	return db.QueryRaw(s, params...)
}

func (db *DB) QueryRow(query string, args ...interface{}) (*Row, error) {
	t, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	switch len(t.Rows) {
	case 0:
		return nil, nil
	case 1:
		return t.Rows[0], nil
	default:
		return nil, fmt.Errorf("the query returned %d results", len(t.Rows))
	}
}

func (db *DB) QueryRowEx(query *sqx.SelectQuery) (*Row, error) {
	t, err := db.QueryEx(query)
	if err != nil {
		return nil, err
	}

	switch len(t.Rows) {
	case 0:
		return nil, nil
	case 1:
		return t.Rows[0], nil
	default:
		return nil, fmt.Errorf("the query returned %d results", len(t.Rows))
	}
}

func (db *DB) QueryValue(query string, args ...interface{}) (interface{}, error) {
	r, err := db.QueryRow(query, args...)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, nil
	}

	if len(r.Values) != 1 {
		return nil, fmt.Errorf("the query returned %d values", len(r.Values))
	}

	return r.Values[0], nil
}

func (db *DB) QueryValueRaw(query string, args ...interface{}) (interface{}, error) {
	rows, err := db.QueryRaw(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	t, more, err := ToTableLimit(rows, 1)
	if err != nil {
		return nil, err
	}
	if more {
		return nil, fmt.Errorf("the query returned more than one row")
	}

	if len(t.Rows) == 0 {
		return nil, nil
	}

	r := t.Rows[0]

	if len(r.Values) != 1 {
		return nil, fmt.Errorf("the query returned %d values", len(r.Values))
	}

	return r.Values[0], nil
}

func (db *DB) QueryValueEx(query *sqx.SelectQuery) (interface{}, error) {
	r, err := db.QueryRowEx(query)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, nil
	}

	if len(r.Values) != 1 {
		return nil, fmt.Errorf("the query returned %d values", len(r.Values))
	}

	return r.Values[0], nil
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	if db.ReadOnly {
		return nil, ErrReadOnly
	}

	q, err := sqx.Parse(query, args...)
	if err != nil {
		return nil, err
	}

	return db.ExecEx(q, args...)
}

func (db *DB) ExecEx(q sqx.Query, args ...interface{}) (sql.Result, error) {
	if db.ReadOnly {
		return nil, ErrReadOnly
	}

	s, params, err := db.ToSql(q)
	if err != nil {
		return nil, err
	}

	r, err := db.ExecRaw(s, params...)
	if err != nil {
		return nil, err
	}

	return r, nil
}
