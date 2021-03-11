package sqx

import (
	"fmt"
	"strconv"
	"strings"
)

type ValidateOptions struct {
	Tables      []*ValidateTable
	queryTables []*Table
}

type ValidateTable struct {
	Name     string
	Database string
	Columns  []string
}

func ValidateSelect(q *SelectQuery, options *ValidateOptions) error {
	if options == nil {
		options = &ValidateOptions{}
	}

	for _, u := range q.From {
		from, ok := u.(*Table)
		if !ok {
			return fmt.Errorf("invalid FROM, expected a table")
		}
		if err := validateFromTable(from, options); err != nil {
			return fmt.Errorf("invalid FROM: %w", err)
		}

		options.queryTables = append(options.queryTables, from)

		for _, join := range from.Joins {
			if err := validateFromExpr(join.TableExpr, options); err != nil {
				return fmt.Errorf("invalid column: %w", err)
			}
		}
	}

	for _, expr := range q.Columns {
		if err := validateExpr(expr, options); err != nil {
			return fmt.Errorf("invalid column: %w", err)
		}
	}

	if q.WherePart != nil {
		if err := validateExpr(q.WherePart.Expr, options); err != nil {
			return fmt.Errorf("invalid where filter: %w", err)
		}
	}

	for _, expr := range q.GroupByPart {
		if err := validateExpr(expr, options); err != nil {
			return fmt.Errorf("invalid GroupBy: %w", err)
		}
	}

	if q.HavingPart != nil {
		if err := validateExpr(q.HavingPart.Expr, options); err != nil {
			return fmt.Errorf("invalid Having: %w", err)
		}
	}

	for _, order := range q.OrderByPart {
		if err := validateExpr(order.Expr, options); err != nil {
			return fmt.Errorf("invalid OrderBy: %w", err)
		}
	}

	if q.LimitPart != nil {
		if err := validateLimitExpr(q.LimitPart.Offset); err != nil {
			return fmt.Errorf("invalid Limit: %w", err)
		}
		if err := validateLimitExpr(q.LimitPart.RowCount); err != nil {
			return fmt.Errorf("invalid Limit: %w", err)
		}
	}

	for _, u := range q.UnionPart {
		if err := ValidateSelect(u, options); err != nil {
			return fmt.Errorf("invalid Union: %w", err)
		}
	}

	return nil
}

func validateLimitExpr(expr Expr) error {
	if expr == nil {
		return nil
	}

	c, ok := expr.(*ConstantExpr)
	if !ok {
		return fmt.Errorf("expected a number")
	}

	if c.Kind != INT {
		return fmt.Errorf("expected a int number")
	}

	i, err := strconv.Atoi(c.Value)
	if err != nil {
		return fmt.Errorf("expected a valid int number")
	}

	const MAX_ROWS = 20000
	if i < 1 || i > MAX_ROWS {
		return fmt.Errorf("invalid range")
	}

	return nil
}

func validateFromExpr(expr SqlFrom, options *ValidateOptions) error {
	switch t := expr.(type) {
	case *Table:
		options.queryTables = append(options.queryTables, t)
		return nil
	case *ParenExpr:
		return validateExpr(t, options)
	case *FromAsExpr:
		return validateExpr(t.From, options)
	default:
		return fmt.Errorf("invalid from %T", t)
	}
}

func validateExpr(expr Expr, options *ValidateOptions) error {
	//Print(expr)

	switch t := expr.(type) {

	case *ColumnNameExpr:
		return validateColumnName(t, options)

	case *BinaryExpr:
		switch t.Operator {
		case NOTIN, NOTLIKE:
			return fmt.Errorf("not_in is not allowed")
		}

		if err := validateExpr(t.Left, options); err != nil {
			return err
		}
		if err := validateExpr(t.Right, options); err != nil {
			return err
		}
		return nil

	case *ParenExpr:
		return validateExpr(t.X, options)

	case *SelectQuery:
		if err := ValidateSelect(t, options); err != nil {
			return err
		}
		return nil

	case *CallExpr:
		return validateCallExpr(t, options)

	case *ConstantExpr:
		switch t.Kind {
		case INT, FLOAT, STRING, TRUE, FALSE:
			return nil
		default:
			return fmt.Errorf("invalid constant: %s", t.Value)
		}

	case *SelectColumnExpr:
		return validateExpr(t.Expr, options)

	case *InExpr:
		return validateInExpr(t, options)

	case *ParameterExpr, *AllColumnsExpr:
		return nil

	default:
		return fmt.Errorf("invalid type: %T", t)
	}
}
func validateInExpr(in *InExpr, options *ValidateOptions) error {
	for _, v := range in.Values {
		switch t := v.(type) {

		case *ColumnNameExpr:
			return validateColumnName(t, options)

		case *SelectQuery:
			return fmt.Errorf("invalid IN expression: subquery")

			// if err := ValidateSelect(t, options); err != nil {
			// 	return fmt.Errorf("invalid IN expression: %w", err)
			// }
			// return nil

		case *ConstantExpr:
			switch t.Kind {
			case INT, STRING:
				return nil
			default:
				return fmt.Errorf("invalid IN constant: %s", t.Value)
			}

		default:
			return fmt.Errorf("invalid IN expression: %T", t)
		}
	}

	return nil
}

func validateCallExpr(expr *CallExpr, options *ValidateOptions) error {
	var found bool
	for _, f := range WhitelistFuncs {
		if strings.EqualFold(f, expr.Name) {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("invalid function: %s", expr.Name)
	}

	for _, arg := range expr.Args {
		if err := validateExpr(arg, options); err != nil {
			return fmt.Errorf("invalid argument: %w", err)
		}
	}

	return nil
}

func validateColumnName(expr *ColumnNameExpr, options *ValidateOptions) error {
	if len(options.Tables) == 0 {
		return nil
	}

	table, err := getValidateTable(expr.Table, options)
	if err != nil {
		return fmt.Errorf("not found '%s': %w", expr.FullName(), err)
	}

	switch len(table.Columns) {
	case 0:
		return fmt.Errorf("not found '%s.%s'", table.Name, expr.Name)
	case 1:
		if table.Columns[0] == "*" {
			return nil
		}
	}

	var found bool
	for _, c := range table.Columns {
		if expr.Name == c {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("not found '%s.%s'", table.Name, expr.Name)
	}

	return nil
}

func validateFromTable(t *Table, options *ValidateOptions) error {
	if err := validateDatabase(t.Database, options); err != nil {
		return err
	}

	if _, err := getValidateTable(t.Name, options); err != nil {
		return err
	}

	return nil
}

func validateDatabase(name string, options *ValidateOptions) error {
	if name != "" {
		return fmt.Errorf("should not specify a database")
	}

	return nil
}

func getValidateTable(name string, options *ValidateOptions) (*ValidateTable, error) {
	if name == "" {
		if len(options.queryTables) != 1 {
			return nil, fmt.Errorf("with more than two tables all columns must be qualified")
		}
		name = options.queryTables[0].Name
		if name == "" {
			return nil, fmt.Errorf("table name is required")
		}
	}

	for _, v := range options.Tables {
		if v.Name == name {
			return v, nil
		}
	}

	// if the table name is not found then search as alias
	var aliasTable string
	for _, t := range options.queryTables {
		if t.Alias == name {
			aliasTable = t.Name
			break
		}
	}

	if aliasTable != "" {
		for _, v := range options.Tables {
			if v.Name == aliasTable {
				return v, nil
			}
		}
	}

	return nil, fmt.Errorf("invalid table '%s'", name)
}
