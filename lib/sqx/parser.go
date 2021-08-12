package sqx

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func Select(code string, params ...interface{}) (*SelectQuery, error) {
	q, err := Parse(code, params...)
	if err != nil {
		return nil, err
	}

	s, ok := q.(*SelectQuery)
	if !ok {
		return nil, fmt.Errorf("the query is not a Select")
	}

	return s, nil
}

// Where parses a select query filter
func Where(code string, params ...interface{}) (*SelectQuery, error) {
	q := &SelectQuery{}
	if err := q.Where(code, params...); err != nil {
		return nil, err
	}

	return q, nil
}

// Parse parses a single query.
func Parse(code string, params ...interface{}) (Query, error) {
	p := NewStrParser(code)

	p.SetParams(params)

	q, err := p.ParseQuery()
	if err != nil {
		return nil, err
	}

	if err := p.AssertParamsSet(); err != nil {
		return nil, err
	}

	return q, nil
}

type Parser struct {
	// If it replaces values with parameters
	ReplaceParams bool

	Params []interface{}

	lexer   *lexer
	lexIdex int
}

func NewStrParser(code string) *Parser {
	r := strings.NewReader(code)
	return NewParser(r)
}

func NewParser(r io.Reader) *Parser {
	return &Parser{lexer: newLexer(r)}
}

func newError(tok *Token, format string, args ...interface{}) *Error {
	return &Error{tok.Pos, fmt.Sprintf(format, args...), tok.Str}
}

func (p *Parser) SetParams(args []interface{}) {
	processParams(args)
	p.Params = args
}

func processParams(params []interface{}) {
	for i, v := range params {
		switch t := v.(type) {
		// convert all dates for the database into to UTC.
		// The mysql driver does this automatically but the sqlite not.
		case time.Time:
			params[i] = t.UTC()
			// save a string representation of duration.
			// TODO: Make this optional
		case time.Duration:
			params[i] = t.String()
		}
	}
}

func (p *Parser) AssertParamsSet() error {
	if len(p.Params) > 0 {
		return fmt.Errorf("too many params")
	}
	return nil
}

// ParseQuery parses a single query.
func (p *Parser) ParseQuery() (Query, error) {
	queries, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if len(queries) != 1 {
		return nil, fmt.Errorf("expected one query, got %d", len(queries))
	}

	return queries[0], nil
}

// Parse parses a sql script. It can contain one or many queries.
func (p *Parser) Parse() ([]Query, error) {
	if err := p.lexer.run(); err != nil {
		return nil, fmt.Errorf("sql Parser: %w", err)
	}

	var queries []Query

loop:
	for {
		t := p.peek()
		switch t.Type {
		case COMMENT:
			p.next()
			continue

		case SELECT:
			n, err := p.parseSelect()
			if err != nil {
				return nil, fmt.Errorf("sql Parser: %w", err)
			}
			queries = append(queries, n)

		case INSERT:
			n, err := p.parseInsert()
			if err != nil {
				return nil, fmt.Errorf("sql Parser: %w", err)
			}
			queries = append(queries, n)

		case UPDATE:
			n, err := p.parseUpdate()
			if err != nil {
				return nil, fmt.Errorf("sql Parser: %w", err)
			}
			queries = append(queries, n)

		case DELETE:
			n, err := p.parseDelete()
			if err != nil {
				return nil, fmt.Errorf("sql Parser: %w", err)
			}
			queries = append(queries, n)

		case CREATE:
			n, err := p.parseCreate()
			if err != nil {
				return nil, fmt.Errorf("sql Parser: %w", err)
			}
			queries = append(queries, n)

		case SHOW:
			n, err := p.parseShow()
			if err != nil {
				return nil, fmt.Errorf("sql Parser: %w", err)
			}
			queries = append(queries, n)

		case DROP:
			n, err := p.parseDrop()
			if err != nil {
				return nil, fmt.Errorf("sql Parser: %w", err)
			}
			queries = append(queries, n)

		case ALTER:
			n, err := p.parseAlter()
			if err != nil {
				return nil, fmt.Errorf("sql Parser: %w", err)
			}
			queries = append(queries, n)

		case EOF:
			break loop
		default:
			return nil, fmt.Errorf("sql Parser: %w", newError(t, "Unexpected '%s' (%v)", t.Str, t.Type))
		}

	inner:
		for {
			t = p.peek()
			switch t.Type {
			case COMMENT:
				p.next()
			case SEMICOLON:
				p.next()
				break inner
			case EOF:
				break loop
			default:
				return nil, fmt.Errorf("sql Parser: %w", newError(t, "Unexpected '%s' (%v)", t.Str, t.Type))
			}
		}

	}

	return queries, nil
}

func (p *Parser) parseDelete() (*DeleteQuery, error) {
	t, err := p.accept(DELETE)
	if err != nil {
		return nil, err
	}

	query := &DeleteQuery{
		Pos: t.Pos,
	}

	t = p.peek()
	if t.Type != FROM {
		for {
			a, err := p.accept(IDENT)
			if err != nil {
				return nil, err
			}

			query.Alias = append(query.Alias, a.Str)

			if p.peek().Type != COMMA {
				break
			}

			p.next()
		}
	}

	_, err = p.accept(FROM)
	if err != nil {
		return nil, err
	}

	t = p.peek()
	if t.Type != IDENT {
		return nil, newError(t, "Exepected ident, got %s", t.Str)
	}

	table, err := p.parseFromTable(t)
	if err != nil {
		return nil, err
	}

	query.Table = table

	where, err := p.parseWhere()
	if err != nil {
		return nil, err
	}
	query.WherePart = where

	limit, err := p.parseLimit()
	if err != nil {
		return nil, err
	}
	query.LimitPart = limit

	return query, nil
}

func (p *Parser) parseUpdate() (*UpdateQuery, error) {
	t, err := p.accept(UPDATE)
	if err != nil {
		return nil, err
	}

	update := &UpdateQuery{
		Pos: t.Pos,
	}

	t = p.peek()
	if t.Type != IDENT {
		return nil, newError(t, "Exepected ident, got %s", t.Str)
	}

	update.Table, err = p.parseFromTable(t)
	if err != nil {
		return nil, err
	}

	t = p.peek()
	switch t.Type {
	case SET:
		_, err = p.accept(SET)
		if err != nil {
			return nil, err
		}

	case EOF:
		return update, nil

	default:
		return nil, newError(t, "Unexpected %s", t.Str)
	}

	var paren bool
	if p.peek().Type == LPAREN {
		p.next()
		paren = true
	}

	cols, err := p.parseColumnValues()
	if err != nil {
		return nil, err
	}
	update.Columns = cols

	if paren {
		if _, err = p.accept(RPAREN); err != nil {
			return nil, err
		}
	}

	where, err := p.parseWhere()
	if err != nil {
		return nil, err
	}
	update.WherePart = where

	limit, err := p.parseLimit()
	if err != nil {
		return nil, err
	}
	update.LimitPart = limit

	return update, nil
}

func (p *Parser) parseInsert() (*InsertQuery, error) {
	_, err := p.accept(INSERT)
	if err != nil {
		return nil, err
	}

	_, err = p.accept(INTO)
	if err != nil {
		return nil, err
	}

	t := p.peek()
	db, table, err := p.parseSelectorIdent()
	if err != nil {
		return nil, err
	}

	insert := &InsertQuery{
		Pos:   t.Pos,
		Table: &TableName{Pos: t.Pos, Database: db, Name: table},
	}

	if p.peek().Type == LPAREN {
		p.next()
		cols, err := p.parseColumnNames()
		if err != nil {
			return nil, err
		}
		insert.Columns = cols

		if _, err := p.accept(RPAREN); err != nil {
			return nil, err
		}
	}

	if p.peek().Type == SELECT {
		sel, err := p.parseSelect()
		if err != nil {
			return nil, err
		}
		insert.Select = sel
		return insert, nil
	}

	_, err = p.accept(VALUES)
	if err != nil {
		return nil, err
	}

	_, err = p.accept(LPAREN)
	if err != nil {
		return nil, err
	}

	values, err := p.parseExpressionList()
	if err != nil {
		return nil, err
	}
	insert.Values = values

	_, err = p.accept(RPAREN)
	if err != nil {
		return nil, err
	}
	return insert, nil
}

func (p *Parser) parseColumnValues() ([]ColumnValue, error) {
	var columns []ColumnValue

loop:
	for {
		t := p.peek()
		table, column, err := p.parseSelectorIdent()
		if err != nil {
			return nil, err
		}

		if _, err = p.accept(EQL); err != nil {
			return nil, err
		}

		exp, err := p.parseNotFactor()
		if err != nil {
			return nil, err
		}

		columns = append(columns, ColumnValue{Pos: t.Pos, Table: table, Name: column, Expr: exp})

		if p.peek().Type != COMMA {
			break loop
		}

		p.next()
	}

	return columns, nil
}

func (p *Parser) parseColumnNames() ([]*ColumnNameExpr, error) {

	var columns []*ColumnNameExpr

loop:
	for {
		col, err := p.parseColumnNameExpr()
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)

		if p.peek().Type != COMMA {
			break loop
		}

		p.next()
	}

	return columns, nil
}

func (p *Parser) parseAlter() (Query, error) {
	t, err := p.accept(ALTER)
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(TABLE); err != nil {
		return nil, err
	}

	db, table, err := p.parseSelectorIdent()
	if err != nil {
		return nil, err
	}

	tt := p.peek()
	switch strings.ToUpper(tt.Str) {
	case "MODIFY":
		p.next()
		return p.parseModifyColumn(t.Pos, db, table)
	case "CHANGE":
		p.next()
		return p.parseRenameColumn(t.Pos, db, table)
	case "ADD":
		p.next()
		return p.parseAlterTableAdd(t, db, table)
	case "DROP":
		p.next()
		return p.parseAlterDrop(t.Pos, db, table)
	default:
		return nil, newError(t, "Invalid alter type %s", tt.Str)
	}
}

func (p *Parser) parseAddColumn(pos Position, db, table string) (*AddColumnQuery, error) {
	c, err := p.parseCreateColumn()
	if err != nil {
		return nil, err
	}

	q := &AddColumnQuery{
		Pos:      pos,
		Database: db,
		Table:    table,
		Column:   c,
	}

	return q, nil
}

func (p *Parser) parseRenameColumn(pos Position, db, table string) (*RenameColumnQuery, error) {
	n, err := p.parseColumnName()
	if err != nil {
		return nil, err
	}

	c, err := p.parseCreateColumn()
	if err != nil {
		return nil, err
	}

	q := &RenameColumnQuery{
		Pos:      pos,
		Database: db,
		Table:    table,
		Name:     n,
		Column:   c,
	}

	return q, nil
}

func (p *Parser) parseModifyColumn(pos Position, db, table string) (*ModifyColumnQuery, error) {
	c, err := p.parseCreateColumn()
	if err != nil {
		return nil, err
	}

	q := &ModifyColumnQuery{
		Pos:      pos,
		Database: db,
		Table:    table,
		Column:   c,
	}

	return q, nil
}

func (p *Parser) parseAlterTableAdd(t *Token, db, table string) (Query, error) {
	tt := p.peek()
	switch strings.ToUpper(tt.Str) {
	case "CONSTRAINT":
		p.next()
		return p.parseAddConstraint(t, db, table)
	case "COLUMN":
		p.next()
		return p.parseAddColumn(t.Pos, db, table)
	default:
		return p.parseAddColumn(t.Pos, db, table)
	}
}

func (p *Parser) parseAddConstraint(t *Token, db, table string) (Query, error) {
	name, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}

	tp, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}

	tpStr := strings.ToUpper(tp.Str)

	switch tpStr {
	case "UNIQUE":
		return p.parseAddUniqueConstraint(t.Pos, db, table, name.Str)
	case "FOREIGN":
		if _, err := p.acceptString("KEY"); err != nil {
			return nil, err
		}
		return p.parseAddFKConstraint(t.Pos, db, table, name.Str)
	default:
		return nil, newError(tp, "Unexpected %s", tp.Str)
	}
}

func (p *Parser) parseAddUniqueConstraint(pos Position, db, table, name string) (*AddConstraintQuery, error) {
	q := &AddConstraintQuery{
		Pos:      pos,
		Type:     "UNIQUE",
		Database: db,
		Table:    table,
		Name:     name,
	}

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

	cols, err := p.parseColumnNames()
	if err != nil {
		return nil, err
	}
	q.Columns = cols

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	return q, nil
}

func (p *Parser) parseAddFKConstraint(pos Position, db, table, name string) (*AddFKQuery, error) {
	q := &AddFKQuery{
		Pos:      pos,
		Type:     "FOREIGN_KEY",
		Database: db,
		Table:    table,
		Name:     name,
	}

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

	col, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}

	q.Column = col.Str

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	if _, err := p.acceptString("REFERENCES"); err != nil {
		return nil, err
	}

	refDB, refTable, err := p.parseSelectorIdent()
	if err != nil {
		return nil, err
	}

	q.RefDatabase = refDB
	q.RefTable = refTable

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

	refCol, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}

	q.RefColumn = refCol.Str

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	if p.peek().Type == ON {
		p.next()
		if _, err := p.accept(DELETE); err != nil {
			return nil, err
		}
		if _, err := p.acceptString("CASCADE"); err != nil {
			return nil, err
		}
		q.DeleteCascade = true
	}

	return q, nil
}

func (p *Parser) parseDrop() (Query, error) {
	t, err := p.accept(DROP)
	if err != nil {
		return nil, err
	}

	switch p.next().Type {
	case DATABASE:
		return p.parseDropDatabase(t.Pos)
	case TABLE:
		return p.parseDropTable(t.Pos)
	default:
		return nil, newError(t, "Unexpected %s", t.Type)
	}
}

func (p *Parser) parseDropDatabase(pos Position) (*DropDatabaseQuery, error) {
	q := &DropDatabaseQuery{Pos: pos}

	if strings.ToUpper(p.peek().Str) == "IF" {
		p.next()
		if _, err := p.accept(EXISTS); err != nil {
			return nil, err
		}
		q.IfExists = true
	}

	t, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}
	q.Database = t.Str

	return q, nil
}

func (p *Parser) parseDropTable(pos Position) (*DropTableQuery, error) {
	q := &DropTableQuery{Pos: pos}

	if strings.ToUpper(p.peek().Str) == "IF" {
		p.next()
		if _, err := p.accept(EXISTS); err != nil {
			return nil, err
		}
		q.IfExists = true
	}

	db, table, err := p.parseSelectorIdent()
	if err != nil {
		return nil, err
	}
	q.Database = db
	q.Table = table

	return q, nil
}

func (p *Parser) parseAlterDrop(pos Position, db, table string) (*AlterDropQuery, error) {
	q := &AlterDropQuery{Pos: pos}
	q.Database = db
	q.Table = table

	t := p.next()
	switch strings.ToUpper(t.Str) {
	case "COLUMN":
		q.Type = "COLUMN"
	case "INDEX":
		q.Type = "INDEX"
	default:
		return nil, newError(t, "Unexpected %s", t.Str)
	}

	c, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}

	q.Item = c.Str

	return q, nil
}

func (p *Parser) parseShow() (*ShowQuery, error) {
	t, err := p.accept(SHOW)
	if err != nil {
		return nil, err
	}

	q := &ShowQuery{Pos: t.Pos}

	t, err = p.accept(IDENT)
	if err != nil {
		return nil, err
	}
	q.Type = t.Str

	switch strings.ToLower(q.Type) {
	case "databases":
		return q, nil
	case "tables":
		// from what database is optional
		if p.peek().Type == FROM {
			p.next()
			t, err = p.accept(IDENT)
			if err != nil {
				return nil, err
			}
			q.Database = t.Str
		}
	case "columns", "index":
		// from what table is required
		if _, err = p.accept(FROM); err != nil {
			return nil, err
		}
		db, tableName, err := p.parseSelectorIdent()
		if err != nil {
			return nil, err
		}
		q.Database = db
		q.Table = tableName
	default:
		return nil, newError(t, "Unexpected %s", q.Type)
	}

	return q, nil
}

func (p *Parser) parseCreate() (Query, error) {
	t, err := p.accept(CREATE)
	if err != nil {
		return nil, err
	}

	q := p.peek()
	switch p.peek().Type {
	case DATABASE:
		return p.parseCreateDatabase(t)
	case TABLE:
		return p.parseCreateTable(t)
	default:
		return nil, newError(t, "Unexpected %s", q.Type)
	}
}

func (p *Parser) parseCreateDatabase(t *Token) (*CreateDatabaseQuery, error) {
	if _, err := p.accept(DATABASE); err != nil {
		return nil, err
	}

	s := &CreateDatabaseQuery{Pos: t.Pos}

	if strings.ToUpper(p.peek().Str) == "IF" {
		p.next()
		if _, err := p.accept(NOT); err != nil {
			return nil, err
		}
		if _, err := p.accept(EXISTS); err != nil {
			return nil, err
		}
		s.IfNotExists = true
	}

	t, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}
	s.Name = t.Str

	return s, nil
}

func (p *Parser) parseCreateTable(t *Token) (*CreateTableQuery, error) {
	if _, err := p.accept(TABLE); err != nil {
		return nil, err
	}

	s := &CreateTableQuery{Pos: t.Pos}

	if strings.ToUpper(p.peek().Str) == "IF" {
		p.next()
		if _, err := p.accept(NOT); err != nil {
			return nil, err
		}
		if _, err := p.accept(EXISTS); err != nil {
			return nil, err
		}
		s.IfNotExists = true
	}

	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	if p.peek().Type == PERIOD {
		p.next()
		table, err := p.parseIdent()
		if err != nil {
			return nil, err
		}
		s.Database = name
		s.Name = table
	} else {
		s.Name = name
	}

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

	cols, err := p.parseCreateColumns()
	if err != nil {
		return nil, err
	}
	s.Columns = cols

	t = p.peek()
	if strings.EqualFold(t.Str, "PRIMARY") {
		p.next()
		if _, err := p.acceptString("KEY"); err != nil {
			return nil, err
		}
		if _, err := p.accept(LPAREN); err != nil {
			return nil, err
		}
		if _, err := p.acceptString("id"); err != nil {
			return nil, err
		}
		if _, err := p.accept(RPAREN); err != nil {
			return nil, err
		}
		if _, err := p.accept(COMMA); err != nil {
			return nil, err
		}
	}

loop:
	for {
		t = p.peek()
		switch t.Type {
		case IDENT:
			switch strings.ToUpper(t.Str) {
			case "UNIQUE":
				c, err := p.parseUniqueKey()
				if err != nil {
					return nil, err
				}
				s.Constraints = append(s.Constraints, c)
			default:
				return nil, newError(t, "Unexpected %s", t.Str)
			}
		case CONSTRAINT:
			c, err := p.parseConstraint()
			if err != nil {
				return nil, err
			}
			s.Constraints = append(s.Constraints, c)
		default:
			break loop
		}

		if p.peek().Type != COMMA {
			break loop
		}

		p.next()
	}

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	return s, nil
}

func (p *Parser) parseConstraint() (CreateTableConstraint, error) {
	if _, err := p.accept(CONSTRAINT); err != nil {
		return nil, err
	}

	t, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}

	name := t.Str

	t, err = p.accept(IDENT)
	if err != nil {
		return nil, err
	}

	switch strings.ToUpper(t.Str) {
	case "FOREIGN":
		return p.parseFKConstraint(name)
	case "UNIQUE":
		return p.parseUniqueConstraint(name)
	default:
		return nil, newError(t, "Unexpected %s", t.Str)
	}
}

func (p *Parser) parseFKConstraint(name string) (*ForeginKey, error) {
	c := &ForeginKey{Name: name}

	if _, err := p.acceptString("KEY"); err != nil {
		return nil, err
	}

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

	name, err := p.parseColumnName()
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	c.Column = name

	if _, err := p.acceptString("REFERENCES"); err != nil {
		return nil, err
	}

	refTable, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	c.RefTable = refTable

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

	name, err = p.parseColumnName()
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	c.RefColumn = name

	t := p.peek()
	if t.Type == ON {
		p.next()
		if _, err := p.acceptString("DELETE"); err != nil {
			return nil, err
		}
		if _, err := p.acceptString("CASCADE"); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (p *Parser) parseUniqueConstraint(name string) (*Constraint, error) {
	c := &Constraint{Name: name, Type: "UNIQUE"}

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

loop:
	for {
		name, err := p.parseColumnName()
		if err != nil {
			return nil, err
		}
		c.Columns = append(c.Columns, name)

		if p.peek().Type != COMMA {
			break loop
		}

		p.next()
	}

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	return c, nil
}

func (p *Parser) parseUniqueKey() (*Constraint, error) {
	if _, err := p.acceptString("UNIQUE"); err != nil {
		return nil, err
	}
	if _, err := p.acceptString("KEY"); err != nil {
		return nil, err
	}

	t, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}

	c := &Constraint{Name: t.Str, Type: "UNIQUE"}

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

loop:
	for {
		name, err := p.parseColumnName()
		if err != nil {
			return nil, err
		}
		c.Columns = append(c.Columns, name)

		if p.peek().Type != COMMA {
			break loop
		}

		p.next()
	}

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	return c, nil
}

func (p *Parser) parseCreateColumns() ([]*CreateColumn, error) {
	var columns []*CreateColumn

loop:
	for {
		t := p.peek()
		switch t.Type {
		case CONSTRAINT:
			break loop
		case IDENT:
			switch t.Str {
			case "PRIMARY", "UNIQUE":
				break loop
			}
		}

		col, err := p.parseCreateColumn()
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)

		if p.peek().Type != COMMA {
			break
		}

		p.next()
	}

	return columns, nil
}

func (p *Parser) parseColumnName() (string, error) {
	t := p.next()

	switch t.Type {
	case IDENT, STRING, TEXT, ORDER:
		return t.Str, nil

	default:
		return "", newError(t, "Invalid column name, got %v", t.Type)
	}
}

func (p *Parser) parseCreateColumn() (*CreateColumn, error) {
	name, err := p.parseColumnName()
	if err != nil {
		return nil, err
	}

	c := &CreateColumn{Name: name}

	// Key is a shortcut for the primary key as an autoincrement int.
	k := p.peek()
	if k.Type == IDENT && strings.ToUpper(k.Str) == "KEY" {
		p.next()
		c.Type = Int
		c.Key = true
		c.Unsigned = true
		return c, nil
	}

	if err := p.parseColumnType(c); err != nil {
		return nil, err
	}

	if err := p.parseColumnSize(c); err != nil {
		return nil, err
	}

	if p.peek().Type == UNSIGNED {
		p.next()
		c.Unsigned = true
	}

	t := p.peek()
	if t.Str == "AUTO_INCREMENT" {
		if !strings.EqualFold(name, "id") {
			return nil, newError(t, "AUTO_INCREMENT is only supported on ID columns: %s", name)
		}
		p.next()
	}

	nullable := true
	if p.peek().Type == NOT {
		p.next()
		nullable = false
	}
	if p.peek().Type == NULL {
		p.next()
		c.Nullable = nullable
	}

	t = p.peek()
	if t.Type == DEFAULT {
		p.next()
		t = p.next()
		switch t.Type {
		case STRING:
			c.Default = "'" + t.Str + "'"

		case INT, FLOAT, TRUE, FALSE:
			c.Default = t.Str

		case IDENT:
			c.Default = t.Str
			if p.peek().Type == LPAREN {
				p.next()
				if _, err := p.accept(RPAREN); err != nil {
					return nil, err
				}
				c.Default += "()"
			}

		default:
			return nil, newError(t, "Invalid default value type %s", t.Str)
		}
	}

	return c, nil
}

func (p *Parser) parseColumnSize(c *CreateColumn) error {
	if p.peek().Type != LPAREN {
		return nil
	}

	p.next()

	t, err := p.accept(INT)
	if err != nil {
		return err
	}

	c.Size = t.Str

	if p.peek().Type == COMMA {
		p.next()

		t, err = p.accept(INT)
		if err != nil {
			return err
		}
		c.Decimals = t.Str
	}

	if _, err := p.accept(RPAREN); err != nil {
		return err
	}

	return nil
}

func (p *Parser) parseColumnType(c *CreateColumn) error {
	t := p.next()
	switch t.Type {
	case INTEGER:
		c.Type = Int
	case BIGINT:
		c.Type = BigInt
	case DECIMAL:
		c.Type = Decimal
	case CHAR:
		c.Type = Char
	case VARCHAR:
		c.Type = Varchar
	case TEXT:
		c.Type = Text
	case MEDIUMTEXT:
		c.Type = MediumText
	case BOOL:
		c.Type = Bool
	case BLOB:
		c.Type = Blob
	case MEDIUMBLOB:
		c.Type = MediumBlob
	case DATETIME:
		c.Type = DatTime
	default:
		return newError(t, "Invalid column type %s", t.Str)
	}

	return nil
}

func (p *Parser) parseSelect() (*SelectQuery, error) {
	t, err := p.accept(SELECT)
	if err != nil {
		return nil, err
	}

	s := &SelectQuery{
		Pos: t.Pos,
	}

	if p.peek().Type == DISTINCT {
		s.Distinct = true
		p.next()
	}

	columns, err := p.parseSelectColumns()
	if err != nil {
		return nil, err
	}
	s.Columns = columns

	// a simple select without from part
	switch p.peek().Type {
	case EOF, RPAREN:
		return s, nil
	}

	if p.peek().Type == FROM {
		from, err := p.parseFrom()
		if err != nil {
			return nil, err
		}
		s.From = from

		where, err := p.parseWhere()
		if err != nil {
			return nil, err
		}
		s.WherePart = where

		group, err := p.parseGroupBy()
		if err != nil {
			return nil, err
		}
		s.GroupByPart = group

		having, ok, err := p.parseHaving()
		if err != nil {
			return nil, err
		}
		if ok {
			s.HavingPart = having
		}
	}

	order, err := p.parseOrderBy()
	if err != nil {
		return nil, err
	}
	s.OrderByPart = order

	limit, err := p.parseLimit()
	if err != nil {
		return nil, err
	}
	s.LimitPart = limit

	union, err := p.parseUnion()
	if err != nil {
		return nil, err
	}
	s.UnionPart = union

	ok, err := p.parseForUpdate()
	if err != nil {
		return nil, err
	}
	s.ForUpdate = ok

	return s, nil
}

func (p *Parser) parseUnion() ([]*SelectQuery, error) {
	if p.peek().Type != UNION {
		return nil, nil
	}

	p.next()

	var queries []*SelectQuery

loop:
	for {
		q, err := p.parseSelect()
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)

		switch p.peek().Type {
		case UNION:
			p.next()
			continue
		default:
			break loop
		}
	}

	return queries, nil
}

func (p *Parser) parseForUpdate() (bool, error) {
	if p.peek().Type != FOR {
		return false, nil
	}

	p.next()

	if _, err := p.accept(UPDATE); err != nil {
		return false, nil
	}

	return true, nil
}

func (p *Parser) parseLimit() (*Limit, error) {
	if p.peek().Type != LIMIT {
		return nil, nil
	}

	t := p.next()

	var e Expr
	rows := p.peek()
	switch rows.Type {
	case INT:
		p.next()
		e = &ConstantExpr{rows.Pos, INT, rows.Str}

	case QUESTION:
		p.next()
		v, err := p.nextParameter()
		if err != nil {
			return nil, newError(t, err.Error())
		}
		e = &ParameterExpr{Pos: rows.Pos, Name: "", Value: v}
	default:
		return nil, newError(t, "Unexpected %s after LIMIT", t.Str)
	}

	limit := &Limit{Pos: t.Pos, RowCount: e}

	if p.peek().Type == COMMA {
		p.next()

		var e2 Expr
		offset := p.peek()
		switch rows.Type {
		case INT:
			p.next()
			e2 = &ConstantExpr{offset.Pos, INT, offset.Str}

		case QUESTION:
			p.next()
			v, err := p.nextParameter()
			if err != nil {
				return nil, newError(t, err.Error())
			}
			e2 = &ParameterExpr{Pos: offset.Pos, Name: "", Value: v}
		default:
			return nil, newError(t, "Unexpected %s after LIMIT", t.Str)
		}

		// when both values are specified then the first one is the offset
		// and the second is the number of rows.
		limit.Offset = e
		limit.RowCount = e2
	}

	return limit, nil
}

func (p *Parser) parseOrderBy() ([]*OrderColumn, error) {
	if p.peek().Type != ORDER {
		return nil, nil
	}

	p.next()
	_, err := p.accept(BY)
	if err != nil {
		return nil, err
	}

	var columns []*OrderColumn

loop:
	for {
		col, err := p.parseColumnNameExpr()
		if err != nil {
			return nil, err
		}

		orderCol := &OrderColumn{Expr: col}
		columns = append(columns, orderCol)

		t := p.peek()
		switch t.Type {
		case ASC, DESC, RANDOM:
			p.next()
			orderCol.Type = t.Type
		}

		switch p.peek().Type {
		case COMMA:
			p.next()
			continue
		default:
			break loop
		}
	}

	return columns, nil
}

func (p *Parser) parseHaving() (*WherePart, bool, error) {
	if p.peek().Type != HAVING {
		return nil, false, nil
	}
	p.next()

	t, err := p.parseHavingPart()
	if err != nil {
		return nil, false, err
	}

	return t, true, nil
}

func (p *Parser) parseHavingPart() (*WherePart, error) {
	exp, err := p.parseBooleanExpr()
	if err != nil {
		return nil, err
	}
	return &WherePart{Pos: exp.Position(), Expr: exp}, nil
}

func (p *Parser) parseGroupBy() ([]Expr, error) {
	if p.peek().Type != GROUP {
		return nil, nil
	}

	p.next()
	_, err := p.accept(BY)
	if err != nil {
		return nil, err
	}

	var columns []Expr

	for {
		col, err := p.parseSelectColumnExpr()
		if err != nil {
			return nil, err
		}

		columns = append(columns, col)
		if p.peek().Type != COMMA {
			break
		}
		p.next()
	}

	return columns, nil
}

func (p *Parser) parseWhere() (*WherePart, error) {
	if p.peek().Type != WHERE {
		return nil, nil
	}
	t := p.next()
	exp, err := p.parseBooleanExpr()
	if err != nil {
		return nil, err
	}
	return &WherePart{Pos: t.Pos, Expr: exp}, nil
}

func (p *Parser) parseSelectColumns() ([]Expr, error) {
	t := p.peek()

	if t.Type == MUL {
		p.next()
		return []Expr{&AllColumnsExpr{Pos: t.Pos}}, nil
	}

	var columns []Expr

loop:
	for {
		exp, err := p.parseSelectColumnExpr()
		if err != nil {
			return nil, err
		}
		columns = append(columns, exp)

		t = p.peek()
		if t.Type != COMMA {
			if t.Type == IDENT {
				return nil, newError(t, "ParseColumns: Unexpected IDENT '%s' at %s", t.Str, t.Pos)
			}
			break loop
		}

		p.next()

		// allow a trailing comma
		if p.peek().Type == FROM {
			break loop
		}
	}

	return columns, nil
}

func (p *Parser) parseSelectColumnExpr() (Expr, error) {
	exp, err := p.parseBooleanExpr()
	if err != nil {
		return nil, err
	}

	if p.peek().Type == AS {
		p.next()
		t := p.next()
		var alias string
		switch t.Type {
		case IDENT, STRING, TEXT:
			alias = t.Str
		default:
			return nil, newError(t, "Expecting alias, got %s", t.Str)
		}
		// if the expresion is a simple column name then don't create a SelectColumnExpr.
		if sel, ok := exp.(*ColumnNameExpr); ok {
			sel.Alias = alias
			return sel, nil
		}

		return &SelectColumnExpr{Expr: exp, Alias: t.Str}, nil
	}

	return exp, nil
}

func (p *Parser) parseFrom() ([]SqlFrom, error) {
	_, err := p.accept(FROM)
	if err != nil {
		return nil, err
	}

	var froms []SqlFrom

	for {
		f, err := p.parseSqlFrom()
		if err != nil {
			return nil, err
		}
		froms = append(froms, f)

		if p.peek().Type != COMMA {
			break
		}

		p.next()
	}

	return froms, nil
}

func (p *Parser) parseSqlFrom() (SqlFrom, error) {
	t := p.peek()
	// is a subquery
	if t.Type == LPAREN {
		sel, err := p.parseParenExpr()
		if err != nil {
			return nil, err
		}

		if p.peek().Type == AS {
			p.next()
			t = p.next()
			switch t.Type {
			case IDENT, STRING, TEXT:
				return &FromAsExpr{sel, t.Str}, nil
			default:
				return nil, newError(t, "Expecting alias, got %s", t.Str)
			}
		} else if p.peek().Type == IDENT {
			t, err = p.accept(IDENT)
			if err != nil {
				return nil, err
			}
			return &FromAsExpr{sel, t.Str}, nil
		} else {
			return sel, nil
		}
	}

	return p.parseFromTable(t)
}

func (p *Parser) parseFromTable(t *Token) (*Table, error) {
	table := &Table{Pos: t.Pos}

	db, tableName, err := p.parseSelectorIdent()
	if err != nil {
		return nil, err
	}

	table.Database = db
	table.Name = tableName

	switch p.peek().Type {
	case AS:
		p.next()
		t = p.next()
		switch t.Type {
		case IDENT, STRING, TEXT:
			table.Alias = t.Str
		default:
			return nil, newError(t, "Expecting alias, got %s", t.Str)
		}
	case IDENT:
		t, err = p.accept(IDENT)
		if err != nil {
			return nil, err
		}
		table.Alias = t.Str
	}

	switch p.peek().Type {
	case LEFT, RIGHT, INNER, OUTER, CROSS, JOIN:
		joins, err := p.parseJoins()
		if err != nil {
			return nil, err
		}
		table.Joins = joins
	}

	return table, nil
}

func (p *Parser) parseJoins() ([]*Join, error) {
	var joins []*Join

loop:
	for {
		var tp Type

		t := p.peek()
		switch t.Type {
		case LEFT, RIGHT, INNER, OUTER, CROSS:
			tp = t.Type
			p.next()
			if _, err := p.accept(JOIN); err != nil {
				return nil, err
			}
		case JOIN:
			tp = JOIN
			p.next()
		default:
			break loop
		}

		expr, err := p.parseSqlFrom()
		if err != nil {
			return nil, err
		}
		join := &Join{Pos: t.Pos, TableExpr: expr, Type: tp}

		// parse the alias part
		switch p.peek().Type {
		case AS:
			p.next()
			t = p.next()
			switch t.Type {
			case IDENT, STRING, TEXT:
				join.Alias = t.Str
			default:
				return nil, newError(t, "Expecting alias, got %s", t.Str)
			}

		case IDENT:
			t, err := p.accept(IDENT)
			if err != nil {
				return nil, err
			}
			join.Alias = t.Str
		}

		if p.peek().Type == ON {
			p.next()
			exp, err := p.parseBooleanExpr()
			if err != nil {
				return nil, err
			}
			join.On = exp
		}

		joins = append(joins, join)
	}

	return joins, nil
}

// parses a 'name' or a 'name.selector' expression
func (p *Parser) parseSelectorIdent() (string, string, error) {
	a, err := p.parseIdent()
	if err != nil {
		return "", "", err
	}

	if p.peek().Type == PERIOD {
		p.next()
		b, err := p.parseIdent()
		if err != nil {
			return "", "", err
		}
		return a, b, nil
	}

	return "", a, nil
}

func (p *Parser) parseIdent() (string, error) {
	a, err := p.accept(IDENT)
	if err != nil {
		return "", err
	}

	return a.Str, nil
}

func (p *Parser) parseColumnNameExpr() (*ColumnNameExpr, error) {
	t := p.peek()

	table, name, err := p.parseSelectorIdent()
	if err != nil {
		return nil, err
	}

	return &ColumnNameExpr{Pos: t.Pos, Table: table, Name: name}, nil
}

func (p *Parser) parseColumnExpr() (Expr, error) {
	t := p.peek()

	table, name, err := p.parseSelecColumnIdent()
	if err != nil {
		return nil, err
	}

	if name == "*" {
		return &AllColumnsExpr{Pos: t.Pos, Table: table}, nil
	}

	return &ColumnNameExpr{Pos: t.Pos, Table: table, Name: name}, nil
}

// parses a select column expression like: 'name', 'a.*' or 'a.name'
func (p *Parser) parseSelecColumnIdent() (string, string, error) {
	a, err := p.accept(IDENT)
	if err != nil {
		return "", "", err
	}

	if p.peek().Type == PERIOD {
		p.next()
		if p.peek().Type == MUL {
			p.next()
			return a.Str, "*", nil
		}

		b, err := p.accept(IDENT)
		if err != nil {
			return "", "", err
		}
		return a.Str, b.Str, nil
	}

	return "", a.Str, nil
}

func (p *Parser) parseBooleanExpr() (Expr, error) {
	lh, err := p.parseBooleanTerm()
	if err != nil {
		return nil, err
	}
	var e Expr = lh

loop:
	for {
		t := p.peek()
		switch t.Type {
		case OR:
			p.next()
			rh, err := p.parseBooleanTerm()
			if err != nil {
				return nil, err
			}

			e = &BinaryExpr{Left: e, Right: rh, Operator: t.Type}
		default:
			break loop
		}
	}

	return e, nil
}

func (p *Parser) parseBooleanTerm() (Expr, error) {
	lh, err := p.parseNotFactor()
	if err != nil {
		return nil, err
	}

	var e Expr = lh
loop:
	for {
		t := p.peek()
		switch t.Type {
		case AND:
			p.next()
			rh, err := p.parseNotFactor()
			if err != nil {
				return nil, err
			}

			e = &BinaryExpr{Left: e, Right: rh, Operator: t.Type}
		default:
			break loop
		}
	}

	return e, nil
}

func (p *Parser) parseNotFactor() (Expr, error) {
	t := p.peek()
	switch t.Type {
	case NT:
		p.next()
		exp, err := p.parseRelation()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Pos: t.Pos, Operator: t.Type, Operand: exp}, nil
	}

	return p.parseRelation()
}

func (p *Parser) parseRelation() (Expr, error) {
	lh, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	var e Expr = lh
loop:
	for {
		t := p.peek()
		switch t.Type {
		case NOT:
			p.next()
			t = p.peek()
			switch t.Type {
			case LIKE:
				p.next()
				rh, err := p.parseExpr()
				if err != nil {
					return nil, err
				}

				e = &BinaryExpr{
					Left:     e,
					Right:    rh,
					Operator: NOTLIKE,
				}
			case IN:
				rh, err := p.parseInExpr(e)
				if err != nil {
					return nil, err
				}

				e = &BinaryExpr{
					Left:     e,
					Right:    rh,
					Operator: NOTIN,
				}
			default:
				return nil, newError(t, "Unexpected %s after NOT", t.Str)
			}
		case EQL, NEQ, LSS, LEQ, GTR, GEQ, LIKE:
			p.next()
			rh, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			e = &BinaryExpr{
				Left:     e,
				Right:    rh,
				Operator: t.Type,
			}
		case IS:
			p.next()
			tp := IS

			if p.peek().Type == NOT {
				tp = ISNOT
				p.next()
			}

			rh, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			e = &BinaryExpr{
				Left:     e,
				Right:    rh,
				Operator: tp,
			}
		case IN:
			rh, err := p.parseInExpr(e)
			if err != nil {
				return nil, err
			}

			e = &BinaryExpr{
				Left:     e,
				Right:    rh,
				Operator: t.Type,
			}
		case BETWEEN:
			rh, err := p.parseBetweenExpr(e)
			if err != nil {
				return nil, err
			}

			e = &BinaryExpr{
				Left:     e,
				Right:    rh,
				Operator: t.Type,
			}
		default:
			break loop
		}
	}

	return e, nil
}

func (p *Parser) parseBetweenExpr(e Expr) (*BetweenExpr, error) {
	p.next()

	left, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(AND); err != nil {
		return nil, err
	}

	right, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return &BetweenExpr{LExpr: left, RExpr: right}, nil
}

func (p *Parser) parseInExpr(e Expr) (*InExpr, error) {
	p.next()

	if p.peek().Type != LPAREN {
		exp, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &InExpr{LParen: exp.Position(), Values: []Expr{exp}, RParen: exp.Position()}, nil
	}

	lp, err := p.accept(LPAREN)
	if err != nil {
		return nil, err
	}

	var values []Expr
	if p.peek().Type == SELECT {
		s, err := p.parseSelect()
		if err != nil {
			return nil, err
		}
		values = append(values, s)
	} else {
		values, err = p.parseExpressionList()
		if err != nil {
			return nil, err
		}
	}

	rp, err := p.accept(RPAREN)
	if err != nil {
		return nil, err
	}

	return &InExpr{LParen: lp.Pos, Values: values, RParen: rp.Pos}, nil
}

func (p *Parser) parseExpr() (Expr, error) {
	lh, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	var e Expr = lh

loop:
	for {
		t := p.peek()
		switch t.Type {
		case ADD, SUB:
			p.next()
			rh, err := p.parseTerm()
			if err != nil {
				return nil, err
			}

			e = &BinaryExpr{Left: e, Right: rh, Operator: t.Type}
		default:
			break loop
		}
	}

	return e, nil
}

func (p *Parser) parseTerm() (Expr, error) {
	lh, err := p.parseSignedFactor()
	if err != nil {
		return nil, err
	}

	var e Expr = lh

loop:
	for {
		t := p.peek()
		switch t.Type {
		case MUL, DIV, MOD, LSF, ANB:
			p.next()
			rh, err := p.parseSignedFactor()
			if err != nil {
				return nil, err
			}

			e = &BinaryExpr{Left: e, Right: rh, Operator: t.Type}
		default:
			break loop
		}
	}

	return e, nil
}

func (p *Parser) parseSignedFactor() (Expr, error) {
	t := p.peek()
	switch t.Type {
	case ADD, SUB:
		p.next()
		exp, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Pos: t.Pos, Operator: t.Type, Operand: exp}, nil
	default:
		return p.parseFactor()
	}
}

func parseValue(t Type, s string) (interface{}, error) {
	switch t {
	case INT:
		return strconv.Atoi(s)
	case FLOAT:
		return strconv.ParseFloat(s, 64)
	case STRING:
		return s, nil
	case NULL:
		return nil, nil
	case TRUE:
		return true, nil
	case FALSE:
		return false, nil
	default:
		return nil, fmt.Errorf("invalid value type: %v", t)
	}
}

func (p *Parser) nextParameter() (interface{}, error) {
	if len(p.Params) == 0 {
		return nil, fmt.Errorf("parameter without a value")
	}
	v := p.Params[0]
	p.Params = p.Params[1:]
	return v, nil
}

func (p *Parser) parseFactor() (Expr, error) {
	t := p.peek()
	switch t.Type {
	case INT, FLOAT, NULL, TRUE, FALSE:
		p.next()
		return &ConstantExpr{t.Pos, t.Type, t.Str}, nil

	case STRING:
		p.next()
		if p.ReplaceParams {
			v, err := parseValue(t.Type, t.Str)
			if err != nil {
				return nil, newError(t, err.Error())
			}
			return &ParameterExpr{Pos: t.Pos, Name: "", Value: v}, nil
		}
		return &ConstantExpr{t.Pos, t.Type, t.Str}, nil

	case DEFAULT:
		p.next()
		return &ConstantExpr{t.Pos, t.Type, t.Str}, nil

	case MUL:
		p.next()
		return &AllColumnsExpr{Pos: t.Pos}, nil

	case QUESTION:
		v, err := p.nextParameter()
		if err != nil {
			return nil, newError(t, err.Error())
		}
		param := &ParameterExpr{Pos: t.Pos, Name: "", Value: v}
		p.next()
		return param, nil

	case IDENT, DISTINCT:
		switch strings.ToUpper(t.Str) {
		case "GROUP_CONCAT":
			return p.parseGroupConcat()
		case "INTERVAL":
			if p.peekTwo().Type == INT {
				return p.parseInterval()
			}
		}

		if p.peekTwo().Type == LPAREN {
			return p.parseCallExpr()
		}

		return p.parseColumnExpr()

	case LPAREN:
		return p.parseParenExpr()

	default:
		return nil, newError(t, "Expecting expression, got %s", t.Type)
	}
}

func (p *Parser) parseParenExpr() (*ParenExpr, error) {
	lparen, err := p.accept(LPAREN)
	if err != nil {
		return nil, err
	}

	var exp Expr
	if p.peek().Type == SELECT {
		exp, err = p.parseSelect()
	} else {
		exp, err = p.parseBooleanExpr()
	}

	if err != nil {
		return nil, err
	}

	rparen, err := p.accept(RPAREN)
	if err != nil {
		return nil, err
	}

	return &ParenExpr{lparen.Pos, exp, rparen.Pos}, nil
}

func (p *Parser) parseInterval() (*DateIntervalExpr, error) {
	p.next()

	n, err := p.accept(INT)
	if err != nil {
		return nil, err
	}

	i, err := strconv.Atoi(n.Str)
	if err != nil {
		return nil, err
	}

	tp, err := p.accept(IDENT)
	if err != nil {
		return nil, err
	}

	expr := &DateIntervalExpr{
		Value: i,
		Type:  tp.Str,
	}

	return expr, nil
}

func (p *Parser) parseGroupConcat() (*GroupConcatExpr, error) {
	t := p.next()
	switch t.Type {
	case IDENT:
	default:
		return nil, newError(t, "Expecting expression, got %s", t.Type)
	}

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

	gr := &GroupConcatExpr{}

	t = p.peek()
	if t.Type == DISTINCT {
		p.next()
		gr.Distinct = true
	}

	for {
		exp, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		gr.Expressions = append(gr.Expressions, exp)

		t = p.peek()
		if t.Type == COMMA {
			p.next()
			continue
		}

		break
	}

	orderBy, err := p.parseOrderBy()
	if err != nil {
		return nil, err
	}
	gr.OrderByPart = orderBy

	t = p.peek()
	if strings.ToUpper(t.Str) == "SEPARATOR" {
		p.next()
		t, err = p.accept(STRING)
		if err != nil {
			return nil, err
		}
		err := validateSeparator(t.Str)
		if err != nil {
			return nil, err
		}
		gr.Separator = t.Str
	}

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	return gr, nil
}

func (p *Parser) parseCallExpr() (*CallExpr, error) {
	t := p.next()
	switch t.Type {
	case IDENT, DISTINCT:
	default:
		return nil, newError(t, "Expecting expression, got %s", t.Type)
	}

	if _, err := p.accept(LPAREN); err != nil {
		return nil, err
	}

	var err error
	var args []Expr
	if p.peek().Type != RPAREN {
		args, err = p.parseExpressionList()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.accept(RPAREN); err != nil {
		return nil, err
	}

	call := &CallExpr{Pos: t.Pos, Name: t.Str, Args: args}

	return call, nil
}

func (p *Parser) parseExpressionList() ([]Expr, error) {
	var args []Expr

	for {
		exp, err := p.parseBooleanExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, exp)

		if p.peek().Type != COMMA {
			break
		}

		p.next()
	}

	return args, nil
}

func (p *Parser) peek() *Token {
	if p.lexIdex >= len(p.lexer.Tokens) {
		return &Token{Type: EOF}
	}
	return p.lexer.Tokens[p.lexIdex]
}

// peek two positions forward
func (p *Parser) peekTwo() *Token {
	i := p.lexIdex + 1
	if i >= len(p.lexer.Tokens) {
		return &Token{Type: EOF}
	}
	return p.lexer.Tokens[i]
}

func (p *Parser) next() *Token {
	if p.lexIdex >= len(p.lexer.Tokens) {
		return &Token{Type: EOF}
	}
	t := p.lexer.Tokens[p.lexIdex]
	p.lexIdex++
	return t
}

func (p *Parser) accept(k Type) (*Token, error) {
	t := p.next()
	if t.Type != k {
		return t, newError(t, "Expecting %v got %v (%s)", k, t.Type, t.Str)
	}
	return t, nil
}

func (p *Parser) acceptString(s string) (*Token, error) {
	t := p.next()

	// don't check if it is of type string. Maybe we want to check
	// if it is the str "ADD" but ADD is an ident for '+'

	if !strings.EqualFold(t.Str, s) {
		return t, newError(t, "Expecting %s, got %s", s, t.Str)
	}
	return t, nil
}
