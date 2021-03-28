package sqx

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var WhitelistFuncs = []string{
	"HOUR",
	"MINUTE",
	"SECOND",
	"DAY",
	"MONTH",
	"YEAR",
	"DAYOFWEEK",
	"NOW",
	"UTC_TIMESTAMP",
	"DATEDIFF",
	"DATE_ADD",
	"DATE",
	"TIME",
	"DATE_FORMAT",
	"MONTHNAME",
	"CONVERT_TZ",
	"SUM",
	"MAX",
	"MIN",
	"AVG",
	"COUNT",
	"ROUND",
	"CONCAT",
	"CONCAT_WS",
	"LENGTH",
	"LCASE",
	"REPLACE",
	"SUBSTR",
	"TRIM",
	"UCASE",
	"IF",
	"GROUP_CONCAT",
	"COALESCE",
	"DISTINCT",

	// Custom sqx funcs
	"START_OF_DAY",
	"END_OF_DAY",
}

type writer struct {
	// A database forces the query to specify all tables with it.
	// Returns an error if a table is prefixed with a different database.
	Database string

	Prefix string

	Namespace string

	Format       bool
	EscapeIdents bool

	ReplaceParams bool

	// whitelist of allowed functions. It it is nil everything allowed.
	WhitelistFuncs []string

	buf    *bytes.Buffer
	params []interface{}
	driver string

	query        Query // the main query
	currentQuery Query // the current query that can be a subquery

	Location *time.Location
}

func NewWriter(q Query, database, prefix, namespace, driver string) *writer {
	if driver == "" {
		driver = "mysql"
	}

	return &writer{
		buf:          new(bytes.Buffer),
		query:        q,
		currentQuery: q,
		Database:     database,
		Prefix:       prefix,
		Namespace:    namespace,
		driver:       driver,
		EscapeIdents: true,
	}
}

// ToSql returns the parsed query
func ToSql(q Query, database, prefix, namespace, driver string) (string, []interface{}, error) {
	return NewWriter(q, database, prefix, namespace, driver).Write()
}

func (p *writer) Write() (string, []interface{}, error) {
	switch t := p.query.(type) {
	case nil:
		return "", nil, fmt.Errorf("empty query")
	case *SelectQuery:
		err := p.writeSelect(t)
		if err != nil {
			return "", nil, err
		}
	case *InsertQuery:
		err := p.writeInsert(t)
		if err != nil {
			return "", nil, err
		}
	case *UpdateQuery:
		err := p.writeUpdate(t)
		if err != nil {
			return "", nil, err
		}
	case *DeleteQuery:
		err := p.writeDelete(t)
		if err != nil {
			return "", nil, err
		}
	case *CreateTableQuery:
		err := p.writeCreateTable(t)
		if err != nil {
			return "", nil, err
		}
	case *CreateDatabaseQuery:
		err := p.writeCreateDatabase(t)
		if err != nil {
			return "", nil, err
		}
	case *RenameColumnQuery:
		err := p.writeRenameColumnQuery(t)
		if err != nil {
			return "", nil, err
		}
	case *ModifyColumnQuery:
		err := p.writeModifyColumnQuery(t)
		if err != nil {
			return "", nil, err
		}
	case *AddColumnQuery:
		err := p.writeAddColumnQuery(t)
		if err != nil {
			return "", nil, err
		}
	case *ShowQuery:
		err := p.writeShow(t)
		if err != nil {
			return "", nil, err
		}
	case *DropDatabaseQuery:
		err := p.writeDropDatabase(t)
		if err != nil {
			return "", nil, err
		}
	case *DropTableQuery:
		err := p.writeDropTable(t)
		if err != nil {
			return "", nil, err
		}
	case *AlterDropQuery:
		err := p.writeAlterDropQuery(t)
		if err != nil {
			return "", nil, err
		}
	case *AddConstraintQuery:
		err := p.writeAddContraint(t)
		if err != nil {
			return "", nil, err
		}
	case *AddFKQuery:
		err := p.writeAddFK(t)
		if err != nil {
			return "", nil, err
		}
	default:
		panic(fmt.Sprintf("not implemented %T", t))
	}

	return p.buf.String(), p.params, nil
}

func (p *writer) writeRenameColumnQuery(q *RenameColumnQuery) error {
	p.buf.WriteString("ALTER TABLE ")

	if err := p.writeTable(q.Database, q.Table); err != nil {
		return err
	}

	p.buf.WriteString(" CHANGE ")

	if err := p.writeIdentifier(q.Name); err != nil {
		return err
	}

	p.buf.WriteString(" ")

	p.currentQuery = q
	return p.writeCreateColumn(q.Column)
}

func (p *writer) writeAddColumnQuery(q *AddColumnQuery) error {
	p.buf.WriteString("ALTER TABLE ")

	if err := p.writeTable(q.Database, q.Table); err != nil {
		return err
	}

	p.buf.WriteString(" ADD COLUMN ")

	p.currentQuery = q

	return p.writeCreateColumn(q.Column)
}

func (p *writer) writeAlterDropQuery(q *AlterDropQuery) error {
	p.buf.WriteString("ALTER TABLE ")

	if err := p.writeTable(q.Database, q.Table); err != nil {
		return err
	}

	p.buf.WriteString(" DROP ")

	switch q.Type {
	case "COLUMN", "INDEX":
		p.buf.WriteString(q.Type)
	default:
		return fmt.Errorf("invalid drop type: %s", q.Type)
	}

	p.buf.WriteString(" ")

	p.currentQuery = q

	return p.writeIdentifier(q.Item)
}

func (p *writer) writeModifyColumnQuery(q *ModifyColumnQuery) error {
	p.buf.WriteString("ALTER TABLE ")

	if err := p.writeTable(q.Database, q.Table); err != nil {
		return err
	}

	p.buf.WriteString(" MODIFY ")

	p.currentQuery = q

	return p.writeCreateColumn(q.Column)
}

func (p *writer) writeDelete(s *DeleteQuery) error {
	p.currentQuery = s

	p.buf.WriteString("DELETE")

	if len(s.Alias) > 1 {
		p.buf.WriteRune(' ')
		for i, a := range s.Alias {
			if i > 0 {
				p.buf.WriteString(", ")
			}
			if err := p.writeIdentifier(a); err != nil {
				return err
			}
		}
	}

	p.buf.WriteString(" FROM ")

	err := p.writeFromTable(s.Table)
	if err != nil {
		return err
	}

	if s.WherePart != nil {
		if p.Format {
			p.buf.WriteRune('\n')
		}

		p.buf.WriteString(" WHERE ")

		err := p.writeExpr(s.WherePart.Expr)
		if err != nil {
			return err
		}
	}

	if s.LimitPart != nil {
		err := p.writeLimit(s.LimitPart)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *writer) writeUpdate(s *UpdateQuery) error {
	p.currentQuery = s

	p.buf.WriteString("UPDATE ")

	err := p.writeFromTable(s.Table)
	if err != nil {
		return err
	}

	p.buf.WriteString(" SET ")

	for i, col := range s.Columns {
		if i > 0 {
			p.buf.WriteString(", ")
		}

		if p.Format {
			p.buf.WriteString("\n ")
		}

		err := p.writeColumnValue(col)
		if err != nil {
			return err
		}
	}

	if s.WherePart != nil {
		if p.Format {
			p.buf.WriteRune('\n')
		}

		p.buf.WriteString(" WHERE ")

		err := p.writeExpr(s.WherePart.Expr)
		if err != nil {
			return err
		}
	}

	if s.LimitPart != nil {
		err := p.writeLimit(s.LimitPart)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *writer) writeInsert(s *InsertQuery) error {
	p.currentQuery = s

	p.buf.WriteString("INSERT INTO ")

	err := p.writeTable(s.Table.Database, s.Table.Name)
	if err != nil {
		return err
	}

	if len(s.Columns) > 0 {
		p.buf.WriteString(" (")

		for i, col := range s.Columns {
			if i > 0 {
				p.buf.WriteString(", ")
			}

			err := p.writeColumnNameExpr(col)
			if err != nil {
				return err
			}
		}

		p.buf.WriteString(")")
	}

	if p.Format {
		p.buf.WriteRune('\n')
	} else {
		p.buf.WriteRune(' ')
	}

	if s.Select != nil {
		return p.writeSelect(s.Select)
	}

	p.buf.WriteString("VALUES (")

	for i, v := range s.Values {
		if i > 0 {
			p.buf.WriteString(", ")
		}

		err := p.writeExpr(v)
		if err != nil {
			return err
		}
	}

	p.buf.WriteString(")")
	return nil
}

func (p *writer) writeAddFK(s *AddFKQuery) error {
	p.buf.WriteString("ALTER TABLE ")

	if err := p.writeTable(s.Database, s.Table); err != nil {
		return err
	}

	p.buf.WriteString(" ADD CONSTRAINT ")

	if err := p.writeIdentifier(s.Name); err != nil {
		return err
	}

	p.buf.WriteString(" FOREIGN KEY(")

	if err := p.writeIdentifier(s.Column); err != nil {
		return err
	}

	p.buf.WriteString(") REFERENCES ")

	if err := p.writeTable(s.RefDatabase, s.RefTable); err != nil {
		return err
	}

	p.buf.WriteRune('(')

	if err := p.writeIdentifier(s.RefColumn); err != nil {
		return err
	}

	p.buf.WriteRune(')')

	if s.DeleteCascade {
		p.buf.WriteString(" ON DELETE CASCADE")
	}

	return nil
}

func (p *writer) writeAddContraint(s *AddConstraintQuery) error {
	p.buf.WriteString("ALTER TABLE ")

	if err := p.writeTable(s.Database, s.Table); err != nil {
		return err
	}

	p.buf.WriteString(" ADD CONSTRAINT ")

	if err := p.writeIdentifier(s.Name); err != nil {
		return err
	}

	switch s.Type {
	case "UNIQUE":
		p.buf.WriteString(" UNIQUE ")
	default:
		return fmt.Errorf("invalid cosntraint type %s at %v", s.Type, s.Pos)
	}

	p.buf.WriteString("(")

	for i, col := range s.Columns {
		if i > 0 {
			p.buf.WriteString(", ")
		}

		err := p.writeColumnNameExpr(col)
		if err != nil {
			return err
		}
	}

	p.buf.WriteString(")")

	return nil
}

func (p *writer) writeDropDatabase(s *DropDatabaseQuery) error {
	p.currentQuery = s
	p.buf.WriteString("DROP DATABASE ")

	if s.IfExists {
		p.buf.WriteString("IF EXISTS ")
	}

	if err := p.writeIdentifier(s.Database); err != nil {
		return err
	}

	return nil
}

func (p *writer) writeDropTable(s *DropTableQuery) error {
	p.currentQuery = s
	p.buf.WriteString("DROP TABLE ")

	if s.IfExists {
		p.buf.WriteString("IF EXISTS ")
	}

	if err := p.writeTable(s.Database, s.Table); err != nil {
		return err
	}

	return nil
}

func (p *writer) writeShow(s *ShowQuery) error {
	p.currentQuery = s

	switch strings.ToLower(s.Type) {
	case "databases":
		return p.writeShowDatabases(s)
	case "tables":
		return p.writeShowTables(s)
	case "columns":
		return p.writeShowColumns(s)
	case "index":
		return p.writeShowIndex(s)
	default:
		return fmt.Errorf("invalid identifier %s at %v", s.Type, s.Pos)
	}
}

func (p *writer) writeShowDatabases(s *ShowQuery) error {
	if p.Database != "" {
		return fmt.Errorf("invalid database in SHOW DATABASES at %v", s.Pos)
	}

	switch p.driver {
	default:
		p.buf.WriteString("SHOW DATABASES")
	}
	return nil
}

func (p *writer) validateDatabase(name string) bool {
	if p.Database != "" {
		if name != "" && name != p.Database {
			return false
		}
	}
	return true
}

func (p *writer) writeDatabase(database string) error {
	return p.writeIdentifier(p.Prefix + database)
}

func (p *writer) writeShowTables(s *ShowQuery) error {
	if !p.validateDatabase(s.Database) {
		return fmt.Errorf("invalid database %s at %v", s.Database, s.Pos)
	}

	database := s.Database
	if database == "" {
		database = p.Database
	}

	switch p.driver {
	default:
		p.buf.WriteString("SHOW TABLES")
		if database != "" {
			p.buf.WriteString(" FROM ")
			if err := p.writeDatabase(database); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *writer) writeShowColumns(q *ShowQuery) error {
	switch p.driver {
	default:
		p.buf.WriteString("SHOW COLUMNS FROM ")
		if err := p.writeTable(q.Database, q.Table); err != nil {
			return err
		}
	}
	return nil
}

func (p *writer) writeShowIndex(q *ShowQuery) error {
	switch p.driver {
	default:
		p.buf.WriteString("SHOW INDEX FROM ")
		if err := p.writeTable(q.Database, q.Table); err != nil {
			return err
		}
	}
	return nil
}

func (p *writer) writeCreateDatabase(s *CreateDatabaseQuery) error {
	p.currentQuery = s

	p.buf.WriteString("CREATE DATABASE ")

	if s.IfNotExists {
		p.buf.WriteString("IF NOT EXISTS ")
	}

	if err := p.writeDatabase(s.Name); err != nil {
		return err
	}
	return nil
}

func (p *writer) writeCreateTable(s *CreateTableQuery) error {
	p.currentQuery = s

	p.buf.WriteString("CREATE TABLE ")

	if s.IfNotExists {
		p.buf.WriteString("IF NOT EXISTS ")
	}

	if err := p.writeTable(p.Database, s.Name); err != nil {
		return err
	}

	p.buf.WriteString(" (")

	var key *CreateColumn

	for i, col := range s.Columns {
		if i > 0 {
			p.buf.WriteString(", ")
		}

		if p.Format {
			p.buf.WriteString("\n\t")
		}

		p.currentQuery = s
		err := p.writeCreateColumn(col)
		if err != nil {
			return err
		}

		if key == nil && col.Key {
			key = col
		}
	}

	if p.driver == "mysql" && key != nil {
		p.buf.WriteString(", ")
		if p.Format {
			p.buf.WriteString("\n\t")
		}

		p.buf.WriteString("PRIMARY KEY(")

		if err := p.writeIdentifier(key.Name); err != nil {
			return err
		}

		p.buf.WriteString(")")
	}

	for _, c := range s.Constraints {
		if p.Format {
			p.buf.WriteRune('\n')
		}

		switch t := c.(type) {
		case *Constraint:
			if err := p.writeConstraint(s, t); err != nil {
				return err
			}
		case *ForeginKey:
			if err := p.writeFKConstraint(s, t); err != nil {
				return err
			}
		default:
			panic(fmt.Sprintf("not implemented %T", c))
		}
	}

	if p.Format {
		p.buf.WriteRune('\n')
	}

	p.buf.WriteString(")")

	if p.driver == "mysql" {
		p.buf.WriteString(" ENGINE=InnoDb")
		p.buf.WriteString(" DEFAULT CHARACTER SET = utf8")
		p.buf.WriteString(" DEFAULT COLLATE = utf8_general_ci")
		return nil
	}

	return nil
}

func (p *writer) writeFKConstraint(s *CreateTableQuery, c *ForeginKey) error {

	p.buf.WriteString(", CONSTRAINT ")

	if err := p.writeIdentifier(c.Name); err != nil {
		return err
	}

	p.buf.WriteString(" FOREIGN KEY (")

	if err := p.writeIdentifier(c.Column); err != nil {
		return err
	}

	p.buf.WriteString(") REFERENCES ")

	if err := p.writeTable(p.Database, c.RefTable); err != nil {
		return err
	}

	p.buf.WriteString("(")

	if err := p.writeIdentifier(c.RefColumn); err != nil {
		return err
	}

	p.buf.WriteString(")")

	if c.DeleteCascade {
		p.buf.WriteString(" ON DELETE CASCADE")
	}

	return nil
}

func (p *writer) writeConstraint(s *CreateTableQuery, c *Constraint) error {
	p.buf.WriteString(", CONSTRAINT ")

	if err := p.writeIdentifier(c.Name); err != nil {
		return err
	}

	p.buf.WriteString(" ")

	if err := p.writeUnescapedAlphanumeric(strings.ToUpper(c.Type)); err != nil {
		return err
	}

	p.buf.WriteString(" (")
	for i, col := range c.Columns {
		if i > 0 {
			p.buf.WriteString(", ")
		}
		if err := p.writeIdentifier(col); err != nil {
			return err
		}
	}
	p.buf.WriteString(")")
	return nil
}

func (p *writer) writeCreateColumn(c *CreateColumn) error {
	switch p.driver {
	default:
		return p.writeCreateColumnMySQL(c)
	}
}

func (p *writer) writeCreateColumnMySQL(c *CreateColumn) error {
	if err := p.writeIdentifier(c.Name); err != nil {
		return err
	}

	switch c.Type {
	case Int:
		p.buf.WriteString(" int")
	case Decimal:
		p.buf.WriteString(" decimal")
	case Char:
		p.buf.WriteString(" char")
	case Varchar:
		p.buf.WriteString(" varchar")
	case Text:
		p.buf.WriteString(" text")
	case MediumText:
		p.buf.WriteString(" mediumtext")
	case Bool:
		p.buf.WriteString(" bool")
	case Blob:
		p.buf.WriteString(" blob")
	case MediumBlob:
		p.buf.WriteString(" mediumblob")
	case DatTime:
		p.buf.WriteString(" datetime")
	default:
		return fmt.Errorf("invalid type %v", c.Type)
	}

	if c.Size != "" {
		p.buf.WriteString("(")
		if err := p.writeUnescapedAlphanumeric(c.Size); err != nil {
			return err
		}

		if c.Decimals != "" {
			p.buf.WriteString(",")
			if err := p.writeUnescapedAlphanumeric(c.Decimals); err != nil {
				return err
			}

		}
		p.buf.WriteString(")")
	}

	if c.Unsigned {
		p.buf.WriteString(" UNSIGNED")
	}

	if c.Key {
		p.buf.WriteString(" AUTO_INCREMENT")
	}

	if !c.Nullable {
		p.buf.WriteString(" NOT")
	}
	p.buf.WriteString(" NULL")

	if c.Default != "" {
		p.buf.WriteString(" DEFAULT ")
		if isAlphanumeric(c.Default) {
			if err := p.writeQuotedAlphanumeric(c.Default); err != nil {
				return err
			}
		} else {
			p.buf.WriteString(c.Default)
		}
	}

	return nil
}

func (p *writer) writeSelect(s *SelectQuery) error {
	p.currentQuery = s

	p.buf.WriteString("SELECT ")

	if s.Distinct {
		p.buf.WriteString("DISTINCT ")
	}

	for i, col := range s.Columns {
		if i > 0 {
			p.buf.WriteString(", ")
			if p.Format {
				p.buf.WriteString("\n       ")
			}
		}

		err := p.writeExpr(col)
		if err != nil {
			return err
		}
	}

	if s.From != nil {
		if p.Format {
			p.buf.WriteString(" \n")
		} else {
			p.buf.WriteRune(' ')
		}

		p.buf.WriteString("FROM ")

		for i, from := range s.From {
			if i > 0 {
				p.buf.WriteString(", ")
			}

			if err := p.writeFrom(from); err != nil {
				return err
			}
		}
	}

	if s.WherePart != nil {
		if p.Format {
			p.buf.WriteRune('\n')
		} else {
			p.buf.WriteRune(' ')
		}

		p.buf.WriteString("WHERE ")

		err := p.writeExpr(s.WherePart.Expr)
		if err != nil {
			return err
		}
	}

	if len(s.GroupByPart) > 0 {
		if p.Format {
			p.buf.WriteRune('\n')
		} else {
			p.buf.WriteRune(' ')
		}

		p.buf.WriteString("GROUP BY ")

		for i, group := range s.GroupByPart {
			if i > 0 {
				p.buf.WriteString(", ")
			}

			err := p.writeExpr(group)
			if err != nil {
				return err
			}
		}
	}

	if s.HavingPart != nil {
		if p.Format {
			p.buf.WriteRune('\n')
		} else {
			p.buf.WriteRune(' ')
		}

		p.buf.WriteString("HAVING ")

		err := p.writeExpr(s.HavingPart.Expr)
		if err != nil {
			return err
		}
	}

	if len(s.OrderByPart) > 0 {
		if p.Format {
			p.buf.WriteRune('\n')
		} else {
			p.buf.WriteRune(' ')
		}

		if err := p.writeOrderBy(s.OrderByPart); err != nil {
			return err
		}
	}

	if s.LimitPart != nil {
		err := p.writeLimit(s.LimitPart)
		if err != nil {
			return err
		}
	}

	for _, u := range s.UnionPart {
		if p.Format {
			p.buf.WriteRune('\n')
		} else {
			p.buf.WriteRune(' ')
		}
		p.buf.WriteString("UNION ")
		err := p.writeSelect(u)
		if err != nil {
			return err
		}
	}

	if s.ForUpdate {
		p.buf.WriteString(" FOR UPDATE")
	}

	return nil
}

func (p *writer) writeLimit(s *Limit) error {
	if p.Format {
		p.buf.WriteRune('\n')
	} else {
		p.buf.WriteRune(' ')
	}

	p.buf.WriteString("LIMIT ")

	if s.Offset != nil {
		p.writeExpr(s.Offset)
		p.buf.WriteString(", ")
	}

	p.writeExpr(s.RowCount)
	return nil
}

func (p *writer) writeFrom(s SqlFrom) error {
	switch t := s.(type) {
	case *Table:
		return p.writeFromTable(t)
	case *ParenExpr:
		return p.writeParenExpr(t)
	case *FromAsExpr:
		if err := p.writeParenExpr(t.From); err != nil {
			return err
		}
		p.buf.WriteString(" AS ")
		return p.writeIdentifier(t.Alias)
	default:
		return fmt.Errorf("invalid from %T", t)
	}
}

func validateIdent(s string) error {
	for i, c := range s {
		if !isIdent(byte(c), i) {
			return fmt.Errorf("invalid identifier %s", s)
		}
	}
	return nil
}

func validateSeparator(s string) error {
	for _, c := range s {
		switch c {
		case ';', ' ', ',', '-', '_', '|':
			continue
		}

		return fmt.Errorf("invalid identifier %s", s)
	}
	return nil
}

func (p *writer) writeTable(database, table string) error {
	if !p.validateDatabase(database) {
		return fmt.Errorf("invalid database %s", database)
	}

	if database == "" {
		database = p.Database
	}

	if database != "" {
		if err := p.writeDatabase(database); err != nil {
			return err
		}
		p.buf.WriteString(".")
	}

	if p.Namespace != "" && !strings.ContainsRune(table, '_') {
		table = p.Namespace + "_" + table
	}

	if err := p.writeIdentifier(table); err != nil {
		return err
	}
	return nil
}

func (p *writer) writeFromTable(t *Table) error {
	if err := p.writeTable(t.Database, t.Name); err != nil {
		return err
	}

	if t.Alias != "" {
		p.buf.WriteString(" AS ")
		if err := p.writeIdentifier(t.Alias); err != nil {
			return err
		}
	}

	return p.writeJoins(t.Joins)
}

func (p *writer) writeIdentifier(s string) error {
	// validate that is an identifier
	for i, c := range s {
		if !isIdent(byte(c), i) {
			return fmt.Errorf("invalid identifier %s", s)
		}
	}

	if p.EscapeIdents {
		p.buf.WriteRune('`')
	}

	p.buf.WriteString(s)

	if p.EscapeIdents {
		p.buf.WriteRune('`')
	}

	return nil
}

func (p *writer) writeUnescapedAlphanumeric(s string) error {
	// validate that is an identifier
	for _, c := range s {
		if !isIdent(byte(c), 1) {
			return fmt.Errorf("invalid identifier %s", s)
		}
	}

	p.buf.WriteString(s)
	return nil
}

func (p *writer) writeQuotedAlphanumeric(s string) error {
	for i, l := 0, len(s)-1; i <= l; i++ {
		c := s[i]

		if i == 0 || i == l {
			switch c {
			case '\'', '"':
				continue
			}
		}

		if !isIdent(byte(c), 1) {
			return fmt.Errorf("invalid identifier %s", s)
		}
	}

	p.buf.WriteString(s)
	return nil
}

func (p *writer) writeJoins(joins []*Join) error {
	for _, j := range joins {
		if err := p.writeJoin(j); err != nil {
			return err
		}
	}
	return nil
}

func (p *writer) writeJoin(join *Join) error {
	if p.Format {
		p.buf.WriteRune('\n')
	} else {
		p.buf.WriteRune(' ')
	}

	switch join.Type {
	case LEFT:
		p.buf.WriteString("LEFT JOIN ")
	case RIGHT:
		p.buf.WriteString("RIGHT JOIN ")
	case INNER:
		p.buf.WriteString("INNER JOIN ")
	case OUTER:
		p.buf.WriteString("OUTER JOIN ")
	case CROSS:
		p.buf.WriteString("CROSS JOIN ")
	case JOIN:
		if p.Format {
			p.buf.WriteRune(' ')
		}
		p.buf.WriteString("JOIN ")
	default:
		return fmt.Errorf("invalid join type: %v", join.Type)
	}

	if err := p.writeFrom(join.TableExpr); err != nil {
		return err
	}

	if join.Alias != "" {
		p.buf.WriteString(" AS ")
		if err := p.writeIdentifier(join.Alias); err != nil {
			return err
		}
	}

	if join.On != nil {
		p.buf.WriteString(" ON ")
		if err := p.writeExpr(join.On); err != nil {
			return err
		}
	}

	return nil
}

func (p *writer) writeExpr(s Expr) error {
	switch t := s.(type) {
	case *ParameterExpr:
		return p.writeParameterExpr(t)
	case *AllColumnsExpr:
		return p.writeAllColumnsExpr(t)
	case *ConstantExpr:
		return p.writeConstantExpr(t)
	case *ParenExpr:
		return p.writeParenExpr(t)
	case *UnaryExpr:
		return p.writeUnaryExpr(t)
	case *BinaryExpr:
		return p.writeBinaryExpr(t)
	case *SelectColumnExpr:
		return p.writeSelectColumnExpr(t)
	case *ColumnNameExpr:
		return p.writeColumnNameExpr(t)
	case *SelectQuery:
		return p.writeSelect(t)
	case *CallExpr:
		return p.writeFuncCallExpr(t)
	case *BetweenExpr:
		return p.writeBetweenExpr(t)
	case *InExpr:
		return p.writeInExpr(t)
	case *GroupConcatExpr:
		return p.writeGroupConcat(t)
	case *DateIntervalExpr:
		return p.writeDateIntervalExpr(t)
	default:
		return fmt.Errorf("invalid expr %T", t)
	}
}

// replace the parameter with the actual value in the query
// because mysql doesn't allow parametrized IN's
func (p *writer) replaceInParameter(t *InExpr) (bool, error) {
	if len(t.Values) != 1 {
		return false, nil
	}

	param, ok := t.Values[0].(*ParameterExpr)
	if !ok {
		return false, nil
	}

	p.buf.WriteRune('(')

	switch x := param.Value.(type) {
	case []interface{}:
		for i, v := range x {
			if i > 0 {
				p.buf.WriteString(", ")
			}
			if err := p.writeInConstant(v); err != nil {
				return true, err
			}
		}
	case interface{}:
		if err := p.writeInConstant(x); err != nil {
			return true, err
		}

	default:
		panic("Invalid IN parameter")
	}

	p.buf.WriteRune(')')

	return true, nil
}

func (p *writer) writeInConstant(v interface{}) error {
	switch t := v.(type) {
	case int:
		p.buf.WriteString(strconv.FormatInt(int64(t), 10))
		return nil
	case int32:
		p.buf.WriteString(strconv.FormatInt(int64(t), 10))
		return nil
	case int64:
		p.buf.WriteString(strconv.FormatInt(t, 10))
		return nil
	case float64:
		i := int64(t)
		if t != float64(i) {
			return fmt.Errorf("invalid IN value %v", v)
		}
		p.buf.WriteString(strconv.FormatInt(i, 10))
		return nil
	case string:
		// if it is an int print it without quotes.
		// TODO: review if this is really what we want.
		if _, err := strconv.Atoi(t); err == nil {
			p.buf.WriteString(t)
			return nil
		}

		// only print safe strings
		if isValidIN(t) {
			p.buf.WriteRune('\'')
			p.buf.WriteString(t)
			p.buf.WriteRune('\'')
			return nil
		}

		return fmt.Errorf("invalid IN value %v", v)
	case time.Time:
		d := t.Format(`'2006-01-02 15:04:05'`)
		p.buf.WriteString(d)
		return nil

	default:
		return fmt.Errorf("invalid IN value: %v", v)
	}
}

func isValidIN(str string) bool {
	for i := range str {
		if !isValidINChar(str[i]) {
			return false
		}
	}

	return true
}

func isValidINChar(c byte) bool {
	if 'A' <= c && c <= 'Z' ||
		'a' <= c && c <= 'z' ||
		'0' <= c && c <= '9' {
		return true
	}

	switch c {
	case '.', '_', '-', '/', '*':
		return true
	}

	return false
}

func (p *writer) writeBetweenExpr(t *BetweenExpr) error {
	if err := p.writeExpr(t.LExpr); err != nil {
		return err
	}

	p.buf.WriteString(" AND ")

	if err := p.writeExpr(t.RExpr); err != nil {
		return err
	}

	return nil
}

func (p *writer) writeInExpr(t *InExpr) error {
	if ok, err := p.replaceInParameter(t); ok {
		return err
	}
	p.buf.WriteRune('(')
	for i, a := range t.Values {
		if i > 0 {
			p.buf.WriteString(", ")
		}
		if err := p.writeExpr(a); err != nil {
			return err
		}
	}
	p.buf.WriteRune(')')
	return nil
}

func (p *writer) isWhiteListedFunc(name string) bool {
	list := p.WhitelistFuncs
	if list == nil {
		list = WhitelistFuncs
		if list == nil {
			return true
		}
	}

	for _, v := range list {
		if strings.EqualFold(v, name) {
			return true
		}
	}

	return false
}

func (p *writer) writeDateIntervalExpr(t *DateIntervalExpr) error {
	p.buf.WriteString("INTERVAL ")
	p.buf.WriteString(strconv.Itoa(t.Value))
	p.buf.WriteString(" ")
	p.buf.WriteString(t.Type)
	return nil
}

func (p *writer) writeGroupConcat(t *GroupConcatExpr) error {
	p.buf.WriteString("GROUP_CONCAT(")

	if t.Distinct {
		p.buf.WriteString("DISTINCT ")
	}

	for i, exp := range t.Expressions {
		if i > 0 {
			p.buf.WriteRune(',')
		}
		if err := p.writeExpr(exp); err != nil {
			return err
		}
	}

	if len(t.OrderByPart) > 0 {
		p.buf.WriteRune(' ')
		if err := p.writeOrderBy(t.OrderByPart); err != nil {
			return err
		}
	}

	if t.Separator != "" {
		p.buf.WriteString(` SEPARATOR '`)
		if err := validateSeparator(t.Separator); err != nil {
			return err
		}
		p.buf.WriteString(t.Separator)
		p.buf.WriteRune('\'')
	}

	p.buf.WriteRune(')')
	return nil
}

func (p *writer) writeFuncCallExpr(t *CallExpr) error {
	name := strings.ToUpper(t.Name)

	if !p.isWhiteListedFunc(name) {
		return fmt.Errorf("the function %s is not allowed", name)
	}

	switch name {
	case "START_OF_DAY":
		return p.writeStartOfDay()
	case "END_OF_DAY":
		return p.writeEndOfDay()
	}

	p.buf.WriteString(name)

	p.buf.WriteRune('(')

	for i, a := range t.Args {
		if i > 0 {
			p.buf.WriteString(", ")
		}

		if err := p.writeExpr(a); err != nil {
			return err
		}
	}

	p.buf.WriteRune(')')

	return nil
}

func (p *writer) writeStartOfDay() error {
	tt := time.Now()

	loc := p.Location
	if loc == nil {
		loc = time.Local
	} else {
		tt = tt.In(p.Location)
	}

	u := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, loc)

	p.buf.WriteRune('"')
	p.buf.WriteString(u.Format("2006-1-2 15:04:05"))
	p.buf.WriteRune('"')

	return nil
}

func (p *writer) writeEndOfDay() error {
	tt := time.Now()

	loc := p.Location
	if loc == nil {
		loc = time.Local
	} else {
		tt = tt.In(p.Location)
	}

	u := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, loc)
	u = u.AddDate(0, 0, 1).Add(-1 * time.Second)

	p.buf.WriteRune('"')
	p.buf.WriteString(u.Format("2006-1-2 15:04:05"))
	p.buf.WriteRune('"')

	return nil
}

func (p *writer) writeOrderBy(t []*OrderColumn) error {
	p.buf.WriteString("ORDER BY ")

	for i, c := range t {
		if i > 0 {
			p.buf.WriteString(", ")
		}

		err := p.writeOrderColumn(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *writer) writeOrderColumn(t *OrderColumn) error {
	err := p.writeExpr(t.Expr)
	if err != nil {
		return err
	}

	switch t.Type {
	case NOTSET:
	case ASC:
		p.buf.WriteString(" ASC")
	case DESC:
		p.buf.WriteString(" DESC")
	default:
		return fmt.Errorf("invalid order type %s at %v", t.Type, t.Expr.Position())
	}

	return nil
}

func (p *writer) writeSelectColumnExpr(t *SelectColumnExpr) error {
	err := p.writeExpr(t.Expr)
	if err != nil {
		return err
	}

	if t.Alias != "" {
		p.buf.WriteString(" AS ")
		if err := p.writeIdentifier(t.Alias); err != nil {
			return err
		}
	}

	return nil
}

func (p *writer) writeColumnValue(t ColumnValue) error {
	if t.Table != "" {
		if err := p.writeIdentifier(t.Table); err != nil {
			return err
		}
		p.buf.WriteRune('.')
	}

	if err := p.writeIdentifier(t.Name); err != nil {
		return err
	}

	p.buf.WriteString(" = ")

	return p.writeExpr(t.Expr)
}

func (p *writer) writeColumnNameExpr(t *ColumnNameExpr) error {
	if t.Table != "" {
		if err := p.writeIdentifier(t.Table); err != nil {
			return err
		}
		p.buf.WriteRune('.')
	}

	if err := p.writeIdentifier(t.Name); err != nil {
		return err
	}

	if t.Alias != "" {
		p.buf.WriteString(" AS ")
		if err := p.writeIdentifier(t.Alias); err != nil {
			return err
		}
	}
	return nil
}

func (p *writer) writeParenExpr(t *ParenExpr) error {
	p.buf.WriteRune('(')

	err := p.writeExpr(t.X)
	if err != nil {
		return err
	}

	p.buf.WriteRune(')')
	return nil
}

func (p *writer) writeUnaryExpr(t *UnaryExpr) error {
	switch t.Operator {
	case ADD:
		p.buf.WriteRune('+')
	case SUB:
		p.buf.WriteRune('-')
	default:
		return fmt.Errorf("invalid unary operator %v", t.Operator)
	}

	err := p.writeExpr(t.Operand)
	if err != nil {
		return err
	}

	return nil
}

// IN is special: it is an error to do "IN ()". If it is empty then replace
// it with an imposible condition to prevent the database crashing:
func (p *writer) handleEmptyIN(t *BinaryExpr) (bool, error) {
	inExp, ok := t.Right.(*InExpr)
	if !ok {
		return false, fmt.Errorf("invalid IN expression")
	}

	if len(inExp.Values) != 1 {
		return false, nil
	}

	param, ok := inExp.Values[0].(*ParameterExpr)
	if !ok {
		return false, nil
	}

	var isEmpty bool
	switch t := param.Value.(type) {
	case nil:
		isEmpty = true

	case []interface{}:
		if len(t) == 0 {
			isEmpty = true
		}
	}

	if isEmpty {
		p.buf.WriteString("1=0")
		return true, nil
	}

	return false, nil
}

// NOT IN is special: it is an error to do "NOT IN ()". If it is empty then
// just an always true condition to prevent the database crashing:
func (p *writer) handleEmptyNOTIN(t *BinaryExpr) (bool, error) {
	inExp, ok := t.Right.(*InExpr)
	if !ok {
		return false, fmt.Errorf("invalid IN expression")
	}

	if len(inExp.Values) != 1 {
		return false, nil
	}

	param, ok := inExp.Values[0].(*ParameterExpr)
	if !ok {
		return false, nil
	}

	var isEmpty bool
	switch t := param.Value.(type) {
	case nil:
		isEmpty = true

	case []interface{}:
		if len(t) == 0 {
			isEmpty = true
		}
	}

	if isEmpty {
		p.buf.WriteString("1=1")
		return true, nil
	}

	return false, nil
}

func (p *writer) handleNullEquality(t *BinaryExpr, equals bool) (bool, error) {
	var isNull bool

	switch t := t.Right.(type) {
	case *ParameterExpr:
		if t.Value != nil {
			return false, nil
		}
		isNull = t.Value == nil

	case *ConstantExpr:
		if t.Kind == NULL {
			isNull = true
		}
	}

	if isNull {
		err := p.writeExpr(t.Left)
		if err != nil {
			return false, err
		}
		if equals {
			p.buf.WriteString(" IS NULL")
		} else {
			p.buf.WriteString(" IS NOT NULL")
		}
		return true, nil
	}

	return false, nil
}

func (p *writer) writeBinaryExpr(t *BinaryExpr) error {
	switch t.Operator {
	case IN:
		ok, err := p.handleEmptyIN(t)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	case NOTIN:
		ok, err := p.handleEmptyNOTIN(t)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	case EQL:
		ok, err := p.handleNullEquality(t, true)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	case NEQ:
		ok, err := p.handleNullEquality(t, false)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}

	err := p.writeExpr(t.Left)
	if err != nil {
		return err
	}

	p.buf.WriteRune(' ')

	switch t.Operator {
	case ADD:
		p.buf.WriteRune('+')
	case SUB:
		p.buf.WriteRune('-')
	case DIV:
		p.buf.WriteRune('/')
	case MUL:
		p.buf.WriteRune('*')
	case MOD:
		p.buf.WriteRune('%')
	case LSF:
		p.buf.WriteString(">>")
	case ANB:
		p.buf.WriteRune('&')
	case LSS:
		p.buf.WriteRune('<')
	case LEQ:
		p.buf.WriteString("<=")
	case GTR:
		p.buf.WriteRune('>')
	case GEQ:
		p.buf.WriteString(">=")
	case EQL:
		p.buf.WriteRune('=')
	case LIKE:
		p.buf.WriteString("LIKE")
	case NOTLIKE:
		p.buf.WriteString("NOT LIKE")
	case NEQ:
		p.buf.WriteString("!=")
	case AND:
		if p.Format {
			p.buf.WriteString("\n ")
		}
		if p.Format {
			p.buf.WriteString(" ")
		}
		p.buf.WriteString("AND")
	case OR:
		if p.Format {
			p.buf.WriteString("\n ")
		}

		if p.Format {
			p.buf.WriteString("  ")
		}
		p.buf.WriteString("OR")
	case BETWEEN:
		p.buf.WriteString("BETWEEN")
	case IN:
		p.buf.WriteString("IN")
	case NOTIN:
		p.buf.WriteString("NOT IN")
	case IS:
		p.buf.WriteString("IS")
	case ISNOT:
		p.buf.WriteString("IS NOT")
	default:
		return fmt.Errorf("invalid binary operator %v", t.Operator)
	}

	p.buf.WriteRune(' ')

	err = p.writeExpr(t.Right)
	if err != nil {
		return err
	}

	return nil
}

func (p *writer) writeAllColumnsExpr(t *AllColumnsExpr) error {
	if t.Table != "" {
		if err := p.writeIdentifier(t.Table); err != nil {
			return err
		}
		p.buf.WriteRune('.')
	}
	p.buf.WriteRune('*')
	return nil
}

func (p *writer) writeParameterExpr(t *ParameterExpr) error {
	if p.ReplaceParams {
		p.writeValue(t.Value)
	} else {
		p.buf.WriteRune('?')
		p.params = append(p.params, t.Value)
	}
	return nil
}

func (p *writer) writeValue(v interface{}) error {
	switch t := v.(type) {
	case int, int64, float64, bool:
		p.buf.WriteString(fmt.Sprintf("%v", t))
	case string:
		p.buf.WriteRune('"')
		p.buf.WriteString(sanitize(t))
		p.buf.WriteRune('"')
	case time.Time:
		p.buf.WriteRune('"')
		p.buf.WriteString(sanitize(t.Format("2006-01-02 15:04:05")))
		p.buf.WriteRune('"')
	case nil:
		p.buf.WriteString("null")
	default:
		p.buf.WriteRune('"')
		p.buf.WriteString(fmt.Sprintf("[%T]", t))
		p.buf.WriteRune('"')
	}

	return nil
}

func (p *writer) writeConstantExpr(t *ConstantExpr) error {
	switch t.Kind {
	case INT, FLOAT:
		p.buf.WriteString(t.Value)
	case STRING:
		p.buf.WriteRune('"')
		p.buf.WriteString(sanitize(t.Value))
		p.buf.WriteRune('"')
	case NULL:
		p.buf.WriteString("null")
	case TRUE:
		p.buf.WriteString("true")
	case FALSE:
		p.buf.WriteString("false")
	case DEFAULT:
		p.buf.WriteString("default")
	default:
		return fmt.Errorf("invalid constant type: %v %v at %v", t.Kind, t.Value, t.Position())
	}
	return nil
}

func sanitize(s string) string {
	return strings.Replace(s, "\"", "\\\"", -1)
}
