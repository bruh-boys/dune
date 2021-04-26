//go:generate stringer -type=ColType

package dbx

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ColType int

const (
	String ColType = iota
	Int
	Decimal
	Bool
	Time
	Date
	DateTime
	Blob
	Unknown
)

type Column struct {
	Name     string       `json:"name"`
	Type     ColType      `json:"type"`
	ScanType reflect.Type `json:"-"`
}

func (c ColType) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteRune('"')
	buf.WriteString(strings.ToLower(c.String()))
	buf.WriteRune('"')
	return buf.Bytes(), nil
}

type Row struct {
	table  *Table
	Values []interface{} `json:"values"`
}

func (r *Row) Table() *Table {
	return r.table
}

// manually serialize
func (r *Row) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteRune('[')

	for i, v := range r.Values {
		if i > 0 {
			buf.WriteRune(',')
		}

		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(b); err != nil {
			return nil, err
		}
	}

	buf.WriteRune(']')

	return buf.Bytes(), nil
}

type Table struct {
	Columns []*Column `json:"columns"`
	Rows    []*Row    `json:"rows"`
}

func (t *Table) ColumnIndex(name string) int {
	for i, c := range t.Columns {
		if strings.EqualFold(c.Name, name) {
			return i
		}
	}
	return -1
}

func (t *Table) NewRow() *Row {
	v := make([]interface{}, len(t.Columns))
	r := &Row{table: t, Values: v}
	t.Rows = append(t.Rows, r)
	return r
}

func (r *Row) Columns() []*Column {
	return r.table.Columns
}

func (r *Row) ColumnIndex(column string) int {
	return r.table.ColumnIndex(column)
}

func (r *Row) Value(column string) (interface{}, bool) {
	i := r.table.ColumnIndex(column)
	if i == -1 {
		return nil, false
	}
	return r.Values[i], true
}

// prepare pointers for rows.Scan
func (r *Row) initScan() {
	for i := range r.Values {
		r.Values[i] = &r.Values[i]
	}
}

func ToTable(rows *sql.Rows) (*Table, error) {
	t, _, err := ToTableLimit(rows, 0)
	return t, err
}

// Returns a table with up to maxRows number of rows and returns also if
// there are more rows to read.
func ToTableLimit(rows *sql.Rows, maxRows int) (*Table, bool, error) {
	cols, err := getColumns(rows)
	if err != nil {
		return nil, false, err
	}
	t := &Table{Columns: cols}

	i := 0
	var moreRows bool

	for rows.Next() {
		if maxRows > 0 && i >= maxRows {
			moreRows = true
			break
		}
		i++

		if t.Columns == nil {
			cols, err := getColumns(rows)
			if err != nil {
				return nil, false, err
			}
			t.Columns = cols
		}

		r := t.NewRow()
		r.initScan()
		err := rows.Scan(r.Values...)
		if err != nil {
			return nil, false, err
		}

		err = convertValues(r)
		if err != nil {
			return nil, false, err
		}
	}

	if rows.Err() != nil {
		return nil, false, rows.Err()
	}

	return t, moreRows, nil
}

func getColumns(rows *sql.Rows) ([]*Column, error) {
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	cs := make([]*Column, len(types))

	for i, t := range types {
		cs[i] = &Column{
			Name:     t.Name(),
			Type:     getType(t.DatabaseTypeName()),
			ScanType: t.ScanType(),
		}
	}

	return cs, nil
}

func getType(s string) ColType {
	// remove the size
	i := strings.IndexRune(s, '(')
	if i != -1 {
		s = s[:i]
	}

	// Common type include "VARCHAR", "TEXT", "NVARCHAR", "DECIMAL", "BOOL", "INT", "BIGINT".

	switch strings.ToLower(s) {
	case "":
		return Unknown
	case "int", "integer", "bigint":
		return Int
	case "string", "text", "varchar", "nvarchar", "char", "mediumtext", "longtext":
		return String
	case "float", "real", "decimal", "double":
		return Decimal
	case "time":
		return Time
	case "date":
		return Date
	case "datetime":
		return DateTime
	case "bool", "boolean", "bit", "tinyint":
		return Bool
	case "blob":
		return Blob
	default:
		panic("invalid type: " + s)
	}
}

func convertValues(r *Row) error {
	cols := r.table.Columns

	for i, v := range r.Values {
		val, err := Convert(v, cols[i].Type)
		if err != nil {
			return fmt.Errorf("error converting %s: %v", cols[i].Name, err)
		}
		r.Values[i] = val
	}

	return nil
}

func Convert(v interface{}, t ColType) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	switch t {
	case Unknown:
		switch v := v.(type) {
		case []byte:
			return string(v), nil
		case string:
			return v, nil
		case nil:
			return "", nil
		default:
			return v, nil
		}
	case String:
		switch v := v.(type) {
		case []byte:
			return string(v), nil
		case string:
			return v, nil
		case nil:
			return "", nil
		default:
			return fmt.Sprintf("%v", v), nil
		}
	case Blob:
		// no way to distinguish with the mysql driver from
		// text and blob
		switch v := v.(type) {
		case []byte:
			return string(v), nil
		case string:
			return v, nil
		case nil:
			return "", nil
		default:
			return fmt.Sprintf("%v", v), nil
		}
	case Bool:
		switch v := v.(type) {
		case []byte:
			switch string(v) {
			case "1", "true":
				return true, nil
			default:
				return false, nil
			}
		case bool:
			return v, nil
		case int64:
			return v == 1, nil
		}
		return nil, fmt.Errorf("can't convert type %T to bool", v)
	case Int:
		switch v := v.(type) {
		case []byte:
			f, err := strconv.Atoi(string(v))
			if err != nil {
				return nil, err
			}
			return f, nil
		case int64:
			return v, nil
		case float64:
			return int(v), nil
		default:
			return nil, fmt.Errorf("can't convert type %T to int", v)
		}
	case Decimal:
		switch v := v.(type) {
		case []byte:
			f, err := strconv.ParseFloat(string(v), 64)
			if err != nil {
				return nil, err
			}
			return f, nil
		case float32:
			return float64(v), nil
		case float64:
			return v, nil
		default:
			return nil, fmt.Errorf("can't convert type %T to float", v)
		}
	case Time, Date, DateTime:
		switch v := v.(type) {
		case []byte:
			t, err := ParseDateTime(string(v))
			if err != nil {
				return nil, err
			}
			return t, nil
		case time.Time:
			return v, nil
		default:
			return nil, fmt.Errorf("can't convert type %T to time", v)
		}
	default:
		return nil, fmt.Errorf("invalid type %v", t)
	}
}

func ParseInt(v interface{}) (int, error) {
	switch t := v.(type) {
	case nil:
		return 0, nil
	case int:
		return t, nil
	case int32:
		return int(t), nil
	case int64:
		return int(t), nil
	case string:
		return strconv.Atoi(t)
	default:
		panic("Invalid type for int")
	}
}

func ParseDateTime(v interface{}) (time.Time, error) {
	switch t := v.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return t, nil
	case []uint8:
		return ParseDateTimeStr(string(t))
	case string:
		return ParseDateTimeStr(t)
	default:
		panic(fmt.Sprintf("Invalid type for datetime %T", v))
	}
}

func ParseDateTimeStr(str string) (time.Time, error) {
	l := len(str)

	if l == 10 {
		if str == "0000-00-00" {
			return time.Time{}, nil
		}
		return time.Parse("2006-01-02", str)
	}

	var format string
	switch str[10] {
	case 'T':
		if str == "0000-00-00T00:00:00" {
			return time.Time{}, nil
		}
		format = "2006-01-02T15:04:05"
	case ' ':
		if str == "0000-00-00 00:00:00" {
			return time.Time{}, nil
		}
		format = "2006-01-02 15:04:05"
	}

	return time.Parse(format, str)
}
