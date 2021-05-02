package lib

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dunelang/dune"

	"github.com/dunelang/dune/lib/dbx"
	"github.com/dunelang/dune/lib/sqx"
)

func init() {
	dune.RegisterLib(SQL, `


declare namespace sql {
    export type DriverType = "mysql" | "sqlite3"

    /**
     * If you specify a databaseName every query will be parsed and all tables will be
     * prefixed with the database name: "SELECT foo FROM bar" will automatically be converted 
     * to "SELECT databasename.foo FROM bar". 
     */
    export function open(driver: DriverType, connString: string, databaseName?: string): DB
	
	export function setWhitelistFuncs(funcs: string[]): void
	

    /**
     * DB is a handle to the database.
     */
    export interface DB {
		database: string
		prefix: string
		namespace: string
        readOnly: boolean
		locked: boolean
        driver: DriverType
		hasTransaction: boolean
		
		initMultiDB(): void

		setMaxOpenConns(v: number): void
		setMaxIdleConns(v: number): void
		setConnMaxLifetime(d: time.Duration | number): void

		onExecuting: (query: Query, ...params: any[]) => void
		
        open(name: string, namespace?: string): DB
        clone(): DB
        close(): void

        reader(query: string | SelectQuery, ...params: any[]): Reader
        query(query: string | SelectQuery, ...params: any[]): any[]
        queryRaw(query: string | SelectQuery, ...params: any[]): any[]
        queryFirst(query: string | SelectQuery, ...params: any[]): any
        queryFirstRaw(query: string | SelectQuery, ...params: any[]): any
        queryValues(query: string | SelectQuery, ...params: any[]): any[]
        queryValuesRaw(query: string | SelectQuery, ...params: any[]): any[]
        queryValue(query: string | SelectQuery, ...params: any[]): any
    	queryValueRaw(query: string | SelectQuery, ...params: any[]): any

        loadTable(query: string | SelectQuery, ...params: any[]): Table
        loadTableRaw(query: string | SelectQuery, ...params: any[]): Table

        exec(query: string | Query, ...params: any[]): Result
        execRaw(query: string, ...params: any[]): Result

        beginTransaction(): void
        commit(): void
        rollback(): void

        hasDatabase(name: string): boolean
        hasTable(name: string): boolean
        databases(): string[]
        tables(): string[]
        columns(table: string): SchemaColumn[]
    }

    export interface SchemaColumn {
        name: string
        type: string
        size: number
        decimals: number
        nullable: boolean
    }

    export interface Reader {
        next(): boolean
        read(): any
        readValues(): any[]
        close(): void
    }

    export interface Result {
        lastInsertId: number
        rowsAffected: number
    }

    export function parse(query: string, ...params: any[]): Query
	export function select(query: string, ...params: any[]): SelectQuery
	
	export interface ValidateOptions {
		tables: Map<string[]>
	}

	export interface QueryTable {
		name: string
		alias: string
		database: string
		leftJoin: boolean
	}

	export function getTables(q: Query): QueryTable[]
	export function getFilterColumns(q: Query): { name: string, table: string }[]
	
	export function validateSelect(q: SelectQuery, options: ValidateOptions): void
	
    export function newSelect(): SelectQuery

    export function where(filter: string, ...params: any[]): SelectQuery

    export function orderBy(s: string): SelectQuery

    export interface Query {
        parameters: any[]
        toSQL(format?: boolean, driver?: DriverType, escapeIdents?: boolean): string
    }

    export interface CRUDQuery extends Query {
        hasLimit: boolean
        hasWhere: boolean
        parameters: any[]
        where(s: string, ...params: any[]): SelectQuery
        and(s: string, ...params: any[]): SelectQuery
        and(filter: SelectQuery): SelectQuery
        or(s: string, ...params: any[]): SelectQuery
        limit(rowCount: number): SelectQuery
        limit(rowCount: number, offset: number): SelectQuery
    }

    export interface SelectQuery extends Query {
        columnsLength: number
        hasLimit: boolean
        hasFrom: boolean
        hasWhere: boolean
        hasDistinct: boolean
        hasOrderBy: boolean
        hasUnion: boolean
        hasGroupBy: boolean
        hasHaving: boolean
        parameters: any[]
        addColumns(s: string): SelectQuery
        setColumns(s: string): SelectQuery
        from(s: string): SelectQuery
        fromExpr(q: SelectQuery, alias: string): SelectQuery
        limit(rowCount: number): SelectQuery
        limit(rowCount: number, offset: number): SelectQuery
        groupBy(s: string): SelectQuery
        orderBy(s: string): SelectQuery
        where(s: string, ...params: any[]): SelectQuery
        having(s: string, ...params: any[]): SelectQuery
        and(s: string, ...params: any[]): SelectQuery
        and(filter: SelectQuery): SelectQuery
        or(s: string, ...params: any[]): SelectQuery
        or(filter: SelectQuery): SelectQuery
        join(s: string, ...params: any[]): SelectQuery

        /**
         * copies all the elements of the query from the Where part.
         */
        setFilter(q: SelectQuery): void

        getFilterColumns(): string[]
    }

    export interface InsertQuery extends Query {
        parameters: any[]
        addColumn(s: string, value: any): void
	}
	
    export interface UpdateQuery extends Query {
        hasLimit: boolean
        hasWhere: boolean
        parameters: any[]
        addColumns(s: string, ...params: any[]): UpdateQuery
        setColumns(s: string, ...params: any[]): UpdateQuery
        where(s: string, ...params: any[]): UpdateQuery
        and(s: string, ...params: any[]): UpdateQuery
        and(filter: UpdateQuery): UpdateQuery
        or(s: string, ...params: any[]): UpdateQuery
        limit(rowCount: number): UpdateQuery
        limit(rowCount: number, offset: number): UpdateQuery
    }

    export interface DeleteQuery extends Query {
        hasLimit: boolean
        hasWhere: boolean
        parameters: any[]
        where(s: string, ...params: any[]): DeleteQuery
        and(s: string, ...params: any[]): DeleteQuery
        and(filter: DeleteQuery): DeleteQuery
        or(s: string, ...params: any[]): DeleteQuery
        or(filter: SelectQuery): SelectQuery
        limit(rowCount: number): DeleteQuery
        limit(rowCount: number, offset: number): DeleteQuery
    }

    export interface Table {
        columns: Column[]
        rows: Row[]
    }

    export interface Row extends Array<any> {
        [index: number]: any
        [key: string]: any
        length: number
        columns: Array<Column>
    }

    export type ColumnType = "string" | "int" | "decimal" | "bool" | "datetime"

    export interface Column {
        name: string
        type: ColumnType
    }


}


`)
}

var SQL = []dune.NativeFunction{
	{
		Name:      "sql.open",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.HasPermission("trusted") {
				return dune.NullValue, ErrUnauthorized
			}

			l := len(args)
			if l < 2 || l > 3 {
				return dune.NullValue, fmt.Errorf("expected 2 or 3 parameters, got %d", l)
			}

			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("argument 1 must be a string, got %s", args[0].TypeName())
			}

			if args[1].Type != dune.String {
				return dune.NullValue, fmt.Errorf("argument 2 must be a string, got %s", args[1].TypeName())
			}

			driver := args[0].String()
			connString := args[1].String()

			db, err := dbx.Open(driver, connString)
			if err != nil {
				return dune.NullValue, err
			}

			db.SetMaxOpenConns(500)
			db.SetMaxIdleConns(250)
			db.SetConnMaxLifetime(5 * time.Minute)

			if l == 3 {
				if args[2].Type != dune.String {
					return dune.NullValue, fmt.Errorf("argument 3 must be a string, got %s", args[2].TypeName())
				}
				name := args[2].String()
				if err := validateDatabaseName(name); err != nil {
					return dune.NullValue, err
				}
				db = db.Open(name)
			}

			ldb := newDB(db)
			vm.SetGlobalFinalizer(ldb)
			return dune.NewObject(ldb), nil
		},
	},
	{
		Name:      "sql.setWhitelistFuncs",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.HasPermission("trusted") {
				return dune.NullValue, ErrUnauthorized
			}

			if err := ValidateArgs(args, dune.Array); err != nil {
				return dune.NullValue, err
			}

			a := args[0].ToArray()

			sqx.ValidFuncs = make([]string, len(a))

			for i, v := range a {
				if v.Type != dune.String {
					return dune.NullValue, fmt.Errorf("invalid value at index %d. It's a %s", i, v.TypeName())
				}
				sqx.ValidFuncs[i] = v.String()
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:      "sql.getTables",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			var q sqx.Query
			switch t := args[0].ToObjectOrNil().(type) {
			case selectQuery:
				q = t.query
			case deleteQuery:
				q = t.query
			case insertQuery:
				q = t.query
			case updateQuery:
				q = t.query
			default:
				return dune.NullValue, fmt.Errorf("expected argument to be Query, got %v", args[0].TypeName())
			}

			tables := sqx.GetTables(q)

			result := make([]dune.Value, len(tables))

			for i, t := range tables {
				r := make(map[dune.Value]dune.Value, len(tables))
				r[dune.NewString("name")] = dune.NewString(t.Name)
				if t.Database != "" {
					r[dune.NewString("database")] = dune.NewString(t.Database)
				}
				if t.Alias != "" {
					r[dune.NewString("alias")] = dune.NewString(t.Alias)
				}
				if t.LeftJoin {
					r[dune.NewString("leftJoin")] = dune.TrueValue
				}
				result[i] = dune.NewMapValues(r)
			}

			return dune.NewArrayValues(result), nil
		},
	},
	{
		Name:      "sql.getFilterColumns",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			var q sqx.Query
			switch t := args[0].ToObjectOrNil().(type) {
			case selectQuery:
				q = t.query
			case deleteQuery:
				q = t.query
			case insertQuery:
				q = t.query
			case updateQuery:
				q = t.query
			default:
				return dune.NullValue, fmt.Errorf("expected argument to be Query, got %v", args[0].TypeName())
			}

			columns := sqx.GetFilterColumns(q)

			result := make([]dune.Value, len(columns))

			for i, t := range columns {
				r := make(map[dune.Value]dune.Value, len(columns))
				r[dune.NewString("name")] = dune.NewString(t.Name)
				if t.Table != "" {
					r[dune.NewString("table")] = dune.NewString(t.Table)
				}
				result[i] = dune.NewMapValues(r)
			}

			return dune.NewArrayValues(result), nil
		},
	},
	{
		Name:      "sql.validateSelect",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object, dune.Map); err != nil {
				return dune.NullValue, err
			}

			s, ok := args[0].ToObjectOrNil().(selectQuery)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected argument to be SelectQuery, got %v", args[0].TypeName())
			}

			opt := &sqx.ValidateOptions{}

			optMap := args[1].ToMap().Map

			tablesVal, ok := optMap[dune.NewString("tables")]
			if !ok {
				return dune.NullValue, fmt.Errorf("invalid options: expected tables")
			}

			if tablesVal.Type != dune.Map {
				return dune.NullValue, fmt.Errorf("invalid tables value: %v", tablesVal)
			}

			for k, v := range tablesVal.ToMap().Map {
				if k.Type != dune.String {
					return dune.NullValue, fmt.Errorf("invalid table name: %v", k)
				}

				if v.Type != dune.Array {
					return dune.NullValue, fmt.Errorf("invalid tables value for %s: %v", k, v)
				}

				colValues := v.ToArray()

				cols := make([]string, len(colValues))

				for i, c := range colValues {
					if c.Type != dune.String {
						return dune.NullValue, fmt.Errorf("invalid column value for %s: %v", k, c)
					}
					cols[i] = c.String()
				}

				opt.Tables = append(opt.Tables, &sqx.ValidateTable{
					Name:    k.String(),
					Columns: cols,
				})
			}

			err := sqx.ValidateSelect(s.query, opt)

			return dune.NullValue, err
		},
	},
	{
		Name:      "sql.newSelect",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			s := selectQuery{&sqx.SelectQuery{}}
			return dune.NewObject(s), nil
		},
	},
	{
		Name:      "sql.parse",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l == 0 {
				return dune.NullValue, fmt.Errorf("expected at least one argument, got %d", l)
			}

			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
			}

			v := args[0].String()

			var params []interface{}
			if l > 1 {
				params = getSqlParams(args[1:])
			}

			q, err := sqx.Parse(v, params...)
			if err != nil {
				return dune.NullValue, err
			}
			obj, err := getQueryObject(q)
			if err != nil {
				return dune.NullValue, err
			}
			return obj, nil
		},
	},
	{
		Name:      "sql.select",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			switch l {
			case 0:
				s := selectQuery{&sqx.SelectQuery{}}
				return dune.NewObject(s), nil
			}

			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
			}

			v := args[0].String()
			var params []interface{}
			if l > 1 {
				params = getSqlParams(args[1:])
			}

			q, err := sqx.Select(v, params...)
			if err != nil {
				return dune.NullValue, err
			}

			s := selectQuery{q}
			return dune.NewObject(s), nil
		},
	},
	{
		Name:      "sql.where",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			switch l {
			case 0:
				s := selectQuery{&sqx.SelectQuery{}}
				return dune.NewObject(s), nil
			}

			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
			}

			v := args[0].String()

			var params []interface{}
			if l > 1 {
				params = make([]interface{}, l-1)
				for i, v := range args[1:] {
					params[i] = v.Export(0)
				}
			}

			q, err := sqx.Where(v, params...)
			if err != nil {
				return dune.NullValue, err
			}

			s := selectQuery{q}
			return dune.NewObject(s), nil
		},
	},
}

func validateDatabaseName(name string) error {
	if name == "" {
		return fmt.Errorf("invalid name: null")
	}

	if !IsIdent(name) {
		return fmt.Errorf("invalid name. It can only contain alphanumeric values")
	}

	l := len(name)
	if l < 3 {
		return fmt.Errorf("name too short. Min 3 chars")
	}

	if l > 40 {
		return fmt.Errorf("name too long. Max 40 chars")
	}

	switch name {
	case "mysql", "performance", "information":
		return fmt.Errorf("invalid database name")
	}

	return nil
}

type sqlResult struct {
	result sql.Result
}

func (t sqlResult) Type() string {
	return "sql.Result"
}

func (t sqlResult) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "lastInsertId":
		i, err := t.result.LastInsertId()
		if err != nil {
			panic(err)
		}
		return dune.NewInt64(i), nil
	case "rowsAffected":
		i, err := t.result.RowsAffected()
		if err != nil {
			panic(err)
		}
		return dune.NewInt64(i), nil
	}
	return dune.UndefinedValue, nil
}

func getRawQuery(driver string, v dune.Value) (string, []interface{}, error) {
	switch v.Type {
	case dune.String:
		return v.String(), nil, nil
	case dune.Object:
		q, ok := v.ToObject().(selectQuery)
		if !ok {
			return "", nil, fmt.Errorf("expected a string or sql.SelectQuery, got %s", v.TypeName())
		}

		w := sqx.NewWriter(q.query, "", "", "", driver)
		w.EscapeIdents = false
		v, params, err := w.Write()
		if err != nil {
			return "", nil, err
		}
		return v, params, nil
	default:
		return "", nil, fmt.Errorf("expected a string or sql.SelectQuery, got %v", v)
	}
}

func getQueryObject(q sqx.Query) (dune.Value, error) {
	var obj interface{}

	switch t := q.(type) {
	case *sqx.SelectQuery:
		obj = selectQuery{t}

	case *sqx.InsertQuery:
		obj = insertQuery{t}

	case *sqx.UpdateQuery:
		obj = updateQuery{t}

	case *sqx.DeleteQuery:
		obj = deleteQuery{t}

	case *sqx.DropTableQuery:
		obj = dropTableQuery{t}

	case *sqx.AlterDropQuery:
		obj = alterDropQuery{t}

	case *sqx.DropDatabaseQuery:
		obj = dropDatabaseQuery{t}

	case *sqx.AddConstraintQuery:
		obj = addConstraintQuery{t}

	case *sqx.AddFKQuery:
		obj = addFKQuery{t}

	case *sqx.AddColumnQuery:
		obj = addColumnQuery{t}

	case *sqx.RenameColumnQuery:
		obj = renameColumnQuery{t}

	case *sqx.ModifyColumnQuery:
		obj = modifyColumnQuery{t}

	case *sqx.CreateDatabaseQuery:
		obj = createDatabaseQuery{t}

	case *sqx.CreateTableQuery:
		obj = createTableQuery{t}

	case *sqx.ShowQuery:
		obj = showQuery{t}

	default:
		return dune.NullValue, fmt.Errorf("invalid query: %T", q)
	}

	return dune.NewObject(obj), nil
}

func getQuery(v dune.Value, params []dune.Value) (sqx.Query, error) {
	switch v.Type {
	case dune.String:
		return sqx.Parse(v.String(), getSqlParams(params)...)
	case dune.Object:
		if len(params) > 0 {
			return nil, fmt.Errorf("can't set params with a query object")
		}
		obj := v.ToObject()
		switch t := obj.(type) {
		case selectQuery:
			return t.query, nil
		case showQuery:
			return t.query, nil
		default:
			return nil, fmt.Errorf("invalid query object, got %s", v.TypeName())
		}
	default:
		return nil, fmt.Errorf("expected a string or sql.SelectQuery, got %v", v)
	}
}

func getSqlParams(args []dune.Value) []interface{} {
	params := make([]interface{}, len(args))
	for i, v := range args {
		switch v.Type {
		case dune.Null, dune.Undefined:
			// leave it nil
		default:
			params[i] = v.Export(0)
		}
	}

	return params
}

func newReader(r *dbx.Reader, vm *dune.VM) dbReader {
	rd := dbReader{r: r}
	vm.SetGlobalFinalizer(rd)
	return rd
}

type dbReader struct {
	r *dbx.Reader
}

func (r dbReader) Type() string {
	return "sql.Reader"
}

func (r dbReader) Close() error {
	return r.r.Close()
}

func (r dbReader) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "columns":
		cols, err := r.r.Columns()
		if err != nil {
			return dune.NullValue, err
		}
		return dune.NewObject(columns{cols}), nil
	}
	return dune.UndefinedValue, nil
}

func (r dbReader) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "next":
		return r.next
	case "read":
		return r.read
	case "readValues":
		return r.readValues
	case "close":
		return r.close
	}
	return nil
}

func (r dbReader) next(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}
	return dune.NewBool(r.r.Next()), nil
}

func (r dbReader) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}
	err := r.r.Close()
	return dune.NullValue, err
}

func (r dbReader) read(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	cols, err := r.r.Columns()
	if err != nil {
		return dune.NullValue, err
	}

	values, err := r.r.Read()
	if err != nil {
		return dune.NullValue, err
	}

	obj := make(map[dune.Value]dune.Value, len(cols))
	for i, col := range cols {
		obj[dune.NewString(col.Name)] = convertDBValue(values[i])
	}

	return dune.NewMapValues(obj), nil
}

func (r dbReader) readValues(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	values, err := r.r.Read()
	if err != nil {
		return dune.NullValue, err
	}

	vs := make([]dune.Value, len(values))
	for i, v := range values {
		vs[i] = convertDBValue(v)
	}

	return dune.NewArrayValues(vs), nil
}

type column struct {
	col *dbx.Column
}

func (t column) Type() string {
	return "sql.Column"
}

func (t column) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "name":
		return dune.NewString(t.col.Name), nil
	case "type":
		return dune.NewString(t.col.Type.String()), nil
	}
	return dune.UndefinedValue, nil
}

type columns struct {
	columns []*dbx.Column
}

func (t columns) Type() string {
	return "sql.Columns"
}

func (t columns) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "length":
		return dune.NewInt(len(t.columns)), nil
	}
	return dune.UndefinedValue, nil
}

func (t columns) GetIndex(i int) (dune.Value, error) {
	cols := t.columns
	if i >= len(cols) {
		return dune.NullValue, fmt.Errorf("index out of range")
	}

	return dune.NewObject(column{cols[i]}), nil
}

func (r columns) Values() ([]dune.Value, error) {
	vs := r.columns
	values := make([]dune.Value, len(vs))
	for i, v := range vs {
		values[i] = dune.NewObject(column{v})
	}
	return values, nil
}

func convertDBValue(v interface{}) dune.Value {
	switch t := v.(type) {
	case time.Time:
		return dune.NewObject(TimeObj(t))
	default:
		return dune.NewValue(v)
	}
}

func newDB(db *dbx.DB) *libDB {
	return &libDB{db: db}
}

type libDB struct {
	locked          bool
	onExecutingFunc dune.Value
	db              *dbx.DB
}

func (s *libDB) Close() error {
	if s.db.HasTransaction() {
		if err := s.db.Rollback(); err != nil {
			return err
		}
	}
	return s.db.Close()
}

func (s *libDB) Type() string {
	return "sql.DB"
}

func (s *libDB) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "hasTransaction":
		return dune.NewBool(s.db.HasTransaction()), nil
	case "database":
		if !vm.HasPermission("trusted") {
			return dune.NullValue, ErrUnauthorized
		}
		return dune.NewString(s.db.Database), nil
	case "prefix":
		if !vm.HasPermission("trusted") {
			return dune.NullValue, ErrUnauthorized
		}
		return dune.NewString(s.db.Prefix), nil
	case "namespace":
		return dune.NewString(s.db.Namespace), nil
	case "locked":
		return dune.NewBool(s.locked), nil
	case "readOnly":
		return dune.NewBool(s.db.ReadOnly), nil
	case "driver":
		return dune.NewString(s.db.Driver), nil
	case "onExecuting":
		if !vm.HasPermission("trusted") {
			return dune.NullValue, ErrUnauthorized
		}
		return s.onExecutingFunc, nil
	}
	return dune.UndefinedValue, nil
}

func (s *libDB) SetProperty(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "onExecuting":
		if !vm.HasPermission("trusted") {
			return ErrUnauthorized
		}
		switch v.Type {
		case dune.Func:
		case dune.Object:
			if _, ok := v.ToObject().(*dune.Closure); !ok {
				return fmt.Errorf("%v is not a function", v.TypeName())
			}
		default:
			return fmt.Errorf("%v is not a function", v.TypeName())
		}

		s.onExecutingFunc = v
		return nil

	case "database":
		if !vm.HasPermission("trusted") {
			return ErrUnauthorized
		}
		if s.locked {
			return fmt.Errorf(("the database is locked"))
		}

		switch v.Type {
		case dune.Undefined, dune.Null:
			s.db.Database = ""

		case dune.String:
			name := v.String()
			if err := validateDatabaseName(name); err != nil {
				return err
			}
			s.db.Database = name

		default:
			return fmt.Errorf("expected string, got %s", v.TypeName())
		}
		return nil

	case "prefix":
		if !vm.HasPermission("trusted") {
			return ErrUnauthorized
		}
		if s.locked {
			return fmt.Errorf(("the database is locked"))
		}
		if v.Type != dune.String {
			return fmt.Errorf("expected string, got %s", v.TypeName())
		}
		name := v.String()
		if !IsIdent(name) {
			return fmt.Errorf("invalid name. It can only contain alphanumeric values")
		}
		s.db.Prefix = name
		return nil

	case "namespace":
		if !vm.HasPermission("trusted") {
			return ErrUnauthorized
		}
		if s.locked {
			return fmt.Errorf(("the database is locked"))
		}
		if v.Type != dune.String {
			return fmt.Errorf("expected string, got %s", v.TypeName())
		}
		name := v.String()
		if !IsIdent(name) {
			return fmt.Errorf("invalid name. It can only contain alphanumeric values")
		}
		s.db.Namespace = name
		return nil

	case "locked":
		if !vm.HasPermission("dbAdministrator") {
			return ErrUnauthorized
		}
		switch v.Type {
		case dune.Bool, dune.Undefined, dune.Null:
		default:
			return fmt.Errorf("expected bool, got %s", v.TypeName())
		}
		s.locked = v.ToBool()
		return nil

	case "readOnly":
		if !vm.HasPermission("dbAdministrator") {
			return ErrUnauthorized
		}
		if s.locked {
			return fmt.Errorf(("the database is locked"))
		}
		switch v.Type {
		case dune.Bool, dune.Undefined, dune.Null:
		default:
			return fmt.Errorf("expected bool, got %s", v.TypeName())
		}
		readOnly := v.ToBool()
		s.db.ReadOnly = readOnly
		return nil

	default:
		return ErrReadOnlyOrUndefined
	}
}

func (s *libDB) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "open":
		return s.open
	case "clone":
		return s.clone
	case "close":
		return s.close
	case "initMultiDB":
		return s.initMultiDB
	case "exec":
		return s.exec
	case "reader":
		return s.reader
	case "query":
		return s.query
	case "queryFirst":
		return s.queryFirst
	case "queryValue":
		return s.queryValue
	case "queryValueRaw":
		return s.queryValueRaw
	case "queryValues":
		return s.queryValues
	case "queryValuesRaw":
		return s.queryValuesRaw
	case "loadTable":
		return s.loadTable
	case "loadTableRaw":
		return s.loadTableRaw
	case "execRaw":
		return s.execRaw
	case "queryRaw":
		return s.queryRaw
	case "queryFirstRaw":
		return s.queryFirstRaw
	case "beginTransaction":
		return s.beginTransaction
	case "commit":
		return s.commit
	case "rollback":
		return s.rollback
	case "tables":
		return s.tables
	case "databases":
		return s.databases
	case "hasDatabase":
		return s.hasDatabase
	case "hasTable":
		return s.hasTable
	case "columns":
		return s.columns
	case "toSQL":
		return s.toSQL
	case "setMaxOpenConns":
		return s.setMaxOpenConns
	case "setMaxIdleConns":
		return s.setMaxIdleConns
	case "setConnMaxLifetime":
		return s.setConnMaxLifetime
	}
	return nil
}

func (s *libDB) setMaxOpenConns(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := ValidateArgs(args, dune.Int); err != nil {
		return dune.NullValue, err
	}

	v := args[0].ToInt()
	s.db.SetMaxOpenConns(int(v))

	return dune.NullValue, nil
}

func (s *libDB) setMaxIdleConns(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := ValidateArgs(args, dune.Int); err != nil {
		return dune.NullValue, err
	}

	v := args[0].ToInt()
	s.db.SetMaxIdleConns(int(v))

	return dune.NullValue, nil
}

func (s *libDB) setConnMaxLifetime(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := ValidateArgRange(args, 1, 1); err != nil {
		return dune.NullValue, err
	}

	d, err := ToDuration(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	s.db.SetConnMaxLifetime(d)
	return dune.NullValue, nil
}

func (s *libDB) open(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := ValidateOptionalArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}

	if err := ValidateArgRange(args, 1, 2); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	if err := validateDatabaseName(name); err != nil {
		return dune.NullValue, err
	}

	db := s.db.Open(name)

	if len(args) == 2 {
		db.Namespace = args[1].String()
	}

	ldb := newDB(db)
	ldb.onExecutingFunc = s.onExecutingFunc
	return dune.NewObject(ldb), nil
}

func (s *libDB) clone(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	db := s.db.Clone()
	ldb := newDB(db)
	ldb.onExecutingFunc = s.onExecutingFunc
	return dune.NewObject(ldb), nil
}

func (s *libDB) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	if err := s.db.Close(); err != nil {
		return dune.NullValue, err
	}
	return dune.NullValue, nil
}

func (s *libDB) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	var q sqx.Query

	a := args[0]

	switch a.Type {
	case dune.Object:
		var ok bool
		q, ok = a.ToObject().(sqx.Query)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected a query, got %s", a.TypeName())
		}
	case dune.String:
		var err error
		q, err = sqx.Parse(a.String())
		if err != nil {
			return dune.NullValue, err
		}
	}

	sq, _, err := s.toSQLString(q, vm)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewString(sq), nil
}

func (s *libDB) toSQLString(q sqx.Query, vm *dune.VM) (string, []interface{}, error) {
	w := sqx.NewWriter(q, s.db.Database, s.db.Prefix, s.db.Namespace, s.db.Driver)
	w.Location = GetLocation(vm)
	w.EscapeIdents = true
	return w.Write()
}

func (s *libDB) beginTransaction(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	err := s.db.Begin()
	return dune.NullValue, err
}

func (s *libDB) commit(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	err := s.db.Commit()
	return dune.NullValue, err
}

func (s *libDB) rollback(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}
	err := s.db.Rollback()
	return dune.NullValue, err
}

func (s *libDB) execRaw(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := s.onExecutingRaw(args, vm); err != nil {
		return dune.NullValue, err
	}

	var query string
	var params []interface{}
	var err error

	l := len(args)

	if l == 0 {
		return dune.NullValue, errors.New("no query provided")
	}

	a := args[0]
	if a.Type != dune.String {
		return dune.NullValue, fmt.Errorf("invalid query, got %v", a)
	}
	query = a.String()

	if l > 1 {
		params = getSqlParams(args[1:])
	}

	res, err := s.db.ExecRaw(query, params...)
	if err != nil {
		if errors.Is(err, dbx.ErrReadOnly) {
			return dune.NullValue, dune.NewTypeError("sql", err.Error())
		}
		return dune.NullValue, err
	}

	return dune.NewObject(sqlResult{res}), nil
}

func getExecQuery(args []dune.Value, vm *dune.VM) (sqx.Query, error) {
	l := len(args)
	if l == 0 {
		return nil, errors.New("no query provided")
	}

	var q sqx.Query
	a := args[0]
	switch a.Type {
	case dune.String:
		var err error
		q, err = sqx.Parse(a.String(), getSqlParams(args[1:])...)
		if err != nil {
			return nil, err
		}
	case dune.Object:
		if len(args) > 1 {
			return nil, fmt.Errorf("invalid query type to pass parameters: %v", a.ToObject())
		}

		switch t := a.ToObject().(type) {
		case insertQuery:
			q = t.query
		case updateQuery:
			q = t.query
		case deleteQuery:
			q = t.query
		case createTableQuery:
			q = t.query
		case addConstraintQuery:
			q = t.query
		case addFKQuery:
			q = t.query
		case addColumnQuery:
			q = t.query
		case renameColumnQuery:
			q = t.query
		case modifyColumnQuery:
			q = t.query
		case dropTableQuery:
			q = t.query
		case alterDropQuery:
			q = t.query
		case createDatabaseQuery:
			q = t.query
		case dropDatabaseQuery:
			q = t.query
		default:
			return nil, fmt.Errorf("invalid query type: %v", t)
		}
	default:
		return nil, fmt.Errorf("expected a query, got %s", a.TypeName())
	}

	// check permissions
	switch q.(type) {
	case *sqx.InsertQuery:
	case *sqx.UpdateQuery:
	case *sqx.DeleteQuery:
	case *sqx.CreateTableQuery:
	case *sqx.AddConstraintQuery:
	case *sqx.AddFKQuery:
	case *sqx.AddColumnQuery:
	case *sqx.RenameColumnQuery:
	case *sqx.ModifyColumnQuery:
	case *sqx.DropTableQuery:
	case *sqx.AlterDropQuery:
		break
	default:
		if !vm.HasPermission("trusted") {
			return nil, ErrUnauthorized
		}
	}

	return q, nil
}

func (s *libDB) initMultiDB(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	err := s.db.InitMultiDB()
	return dune.NullValue, err
}

func (s *libDB) exec(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	q, err := getExecQuery(args, vm)
	if err != nil {
		return dune.NullValue, err
	}

	if err := s.onExecuting(q, args[1:], vm); err != nil {
		return dune.NullValue, err
	}

	if s.db.Driver == "sqlite3" {
		switch t := q.(type) {
		case *sqx.DropDatabaseQuery:
			err := s.dropSqliteDatabase(t.Database)
			return dune.NullValue, err
		}
	}

	sQuery, params, err := s.toSQLString(q, vm)
	if err != nil {
		return dune.NullValue, err
	}

	res, err := s.db.ExecRaw(sQuery, params...)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(sqlResult{res}), nil
}

func (s *libDB) dropSqliteDatabase(name string) error {
	_, err := s.db.ExecRaw("DELETE FROM dbx_internal_schema WHERE name = ?", name)
	if err != nil {
		return err
	}

	q := `SELECT name 
			FROM sqlite_master 
			WHERE type = 'table' 
			AND name LIKE '` + name + "_%'"

	rows, err := s.db.QueryRaw(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	var table string
	var tables []string
	for rows.Next() {
		err = rows.Scan(&table)
		if err != nil {
			return err
		}
		tables = append(tables, table)
	}

	if rows.Err() != nil {
		return err
	}

	for _, table := range tables {
		if _, err := s.db.ExecRaw("DROP TABLE " + table); err != nil {
			return err
		}
	}

	return nil
}

func (s *libDB) queryFirst(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	var q sqx.Query
	var err error

	switch len(args) {
	case 0:
		return dune.NullValue, errors.New("no query provided")
	default:
		if q, err = getQuery(args[0], args[1:]); err != nil {
			return dune.NullValue, err
		}
	}

	if err := s.onExecuting(q, args[1:], vm); err != nil {
		return dune.NullValue, err
	}

	var rows *sql.Rows

	switch q.(type) {
	case *sqx.SelectQuery:
		sQuery, sParams, err := s.toSQLString(q, vm)
		if err != nil {
			return dune.NullValue, err
		}
		rows, err = s.db.QueryRaw(sQuery, sParams...)
		if err != nil {
			return dune.NullValue, err
		}
	default:
		return dune.NullValue, fmt.Errorf("not a select query")
	}

	defer rows.Close()

	t, _, err := dbx.ToTableLimit(rows, 1)
	if err != nil {
		return dune.NullValue, err
	}

	switch len(t.Rows) {
	case 0:
		return dune.NullValue, nil
	case 1:
		r := t.Rows[0]
		m := make(map[dune.Value]dune.Value, len(t.Columns))
		for k, col := range t.Columns {
			m[dune.NewString(col.Name)] = convertDBValue(r.Values[k])
		}
		return dune.NewMapValues(m), nil
	default:
		panic(fmt.Sprintf("The table has more than 1 row: %d", len(t.Rows)))
	}
}

func (s *libDB) queryFirstRaw(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := s.onExecutingRaw(args, vm); err != nil {
		return dune.NullValue, err
	}

	l := len(args)
	if l == 0 {
		return dune.NullValue, errors.New("no query provided")
	}

	query, params, err := getRawQuery(s.db.Driver, args[0])
	if err != nil {
		return dune.NullValue, err
	}

	if l > 1 {
		params = append(params, getSqlParams(args[1:])...)
	}

	rows, err := s.db.QueryRaw(query, params...)
	if err != nil {
		return dune.NullValue, err
	}
	if rows == nil {
		return dune.NullValue, nil
	}

	defer rows.Close()

	t, err := dbx.ToTable(rows)
	if err != nil {
		return dune.NullValue, err
	}

	var r *dbx.Row
	switch len(t.Rows) {
	case 0:
		return dune.NullValue, nil
	case 1:
		r = t.Rows[0]
	default:
		return dune.NullValue, fmt.Errorf("the query returned %d results", len(t.Rows))
	}

	m := make(map[dune.Value]dune.Value, len(t.Columns))
	for k, col := range t.Columns {
		m[dune.NewString(col.Name)] = convertDBValue(r.Values[k])
	}

	return dune.NewMapValues(m), nil
}

func (s *libDB) queryValue(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	var q sqx.Query
	var err error

	switch len(args) {
	case 0:
		return dune.NullValue, errors.New("no query provided")
	default:
		if q, err = getQuery(args[0], args[1:]); err != nil {
			return dune.NullValue, err
		}
	}

	if err := s.onExecuting(q, args[1:], vm); err != nil {
		return dune.NullValue, err
	}

	var v interface{}

	switch q.(type) {
	case *sqx.SelectQuery:
		sQuery, sParams, err := s.toSQLString(q, vm)
		if err != nil {
			return dune.NullValue, err
		}

		v, err = s.db.QueryValueRaw(sQuery, sParams...)
		if err != nil {
			return dune.NullValue, err
		}
	default:
		return dune.NullValue, fmt.Errorf("not a select query")
	}

	if v == nil {
		return dune.NullValue, nil
	}
	return convertDBValue(v), nil
}

func (s *libDB) queryValueRaw(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := s.onExecutingRaw(args, vm); err != nil {
		return dune.NullValue, err
	}

	l := len(args)
	if l == 0 {
		return dune.NullValue, errors.New("no query provided")
	}

	query, params, err := getRawQuery(s.db.Driver, args[0])
	if err != nil {
		return dune.NullValue, err
	}

	if l > 1 {
		params = append(params, getSqlParams(args[1:])...)
	}

	row := s.db.QueryRowRaw(query, params...)

	var v interface{}
	if err := row.Scan(&v); err != nil {
		return dune.NullValue, err
	}

	a := convertDBValue(v)
	return a, nil
}

func toValueArgs(q sqx.Query, args []dune.Value) []dune.Value {
	var sq dune.Value
	switch t := q.(type) {
	case *sqx.SelectQuery:
		sq = dune.NewObject(selectQuery{t})
	case *sqx.InsertQuery:
		sq = dune.NewObject(insertQuery{t})
	case *sqx.UpdateQuery:
		sq = dune.NewObject(updateQuery{t})
	case *sqx.DeleteQuery:
		sq = dune.NewObject(deleteQuery{t})
	default:
		sq = dune.NewObject(query{t})
	}

	sArgs := []dune.Value{sq}
	if len(args) > 0 {
		sArgs = append(sArgs, dune.NewArrayValues(args))
	}

	return sArgs
}

func (s *libDB) onExecuting(q sqx.Query, args []dune.Value, vm *dune.VM) error {
	v := s.onExecutingFunc
	switch v.Type {
	case dune.Null, dune.Undefined:
		return nil

	case dune.Func, dune.Object:
		sArgs := toValueArgs(q, args)
		return s.onExecutingRaw(sArgs, vm)

	default:
		return fmt.Errorf("%v is not a function", v.TypeName())
	}
}

func (s *libDB) onExecutingRaw(args []dune.Value, vm *dune.VM) error {
	v := s.onExecutingFunc

	switch v.Type {
	case dune.Null, dune.Undefined:
		return nil

	case dune.Func:
		if _, err := vm.RunFuncIndex(s.onExecutingFunc.ToFunction(), args...); err != nil {
			return fmt.Errorf("error in onExecuting: %w", err)
		}
		return nil

	case dune.Object:
		c, ok := v.ToObject().(*dune.Closure)
		if !ok {
			return fmt.Errorf("%v is not a function", v.TypeName())
		}
		if _, err := vm.RunClosure(c, args...); err != nil {
			return fmt.Errorf("error in onExecuting: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("%v is not a function", v.TypeName())
	}
}

func (s *libDB) hasDatabase(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	exists, err := s.db.HasDatabase(args[0].String())
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewBool(exists), nil
}

func (s *libDB) hasTable(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	exists, err := s.db.HasTable(args[0].String())
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewBool(exists), nil
}

func (s *libDB) tables(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	tables, err := s.db.Tables()
	if err != nil {
		return dune.NullValue, err
	}

	result := make([]dune.Value, len(tables))

	for i, t := range tables {
		result[i] = dune.NewString(t)
	}

	return dune.NewArrayValues(result), nil
}

func (s *libDB) databases(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	table, err := s.db.Databases()
	if err != nil {
		return dune.NullValue, err
	}

	rows := table.Rows

	result := make([]dune.Value, len(rows))

	for i, row := range rows {
		result[i] = dune.NewString(row.Values[0].(string))
	}

	return dune.NewArrayValues(result), nil
}

func (s *libDB) columns(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	var dbName, table string

	parts := Split(args[0].String(), ".")

	switch len(parts) {
	case 1:
		dbName = s.db.Database
		table = parts[0]
		if !validateTable(table) {
			return dune.NullValue, fmt.Errorf("invalid table: %s", args[0])
		}

	case 2:
		if !vm.HasPermission("trusted") {
			return dune.NullValue, ErrUnauthorized
		}

		dbName = parts[0]
		table = parts[1]
		if !validateTable(parts[1]) {
			return dune.NullValue, fmt.Errorf("invalid table: %s", args[0])
		}

	default:
		return dune.NullValue, fmt.Errorf("invalid table: %s", args[0])

	}

	if dbName != "" {
		table = dbName + "." + table
	}

	columns, err := s.db.Columns(table)
	if err != nil {
		return dune.NullValue, err
	}

	result := make([]dune.Value, len(columns))

	for i, v := range columns {
		c, err := newColumn(v)
		if err != nil {
			return dune.NullValue, err
		}

		result[i] = dune.NewObject(c)
	}

	return dune.NewArrayValues(result), nil
}

func validateTable(table string) bool {
	parts := Split(table, ":")

	for _, p := range parts {
		if !IsIdent(p) {
			return false
		}
	}

	return true
}

func (s *libDB) query(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	tbl, err := s.getTable(args, vm)
	if err != nil {
		return dune.NullValue, err
	}

	result := make([]dune.Value, len(tbl.Rows))
	l := len(tbl.Columns)

	for i, r := range tbl.Rows {
		m := make(map[dune.Value]dune.Value, l)
		for k, col := range tbl.Columns {
			m[dune.NewString(col.Name)] = convertDBValue(r.Values[k])
		}
		result[i] = dune.NewMapValues(m)
	}

	return dune.NewArrayValues(result), nil
}

func (s *libDB) loadTable(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	tbl, err := s.getTable(args, vm)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(&table{dbxTable: tbl}), nil
}

func (s *libDB) loadTableRaw(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := s.onExecutingRaw(args, vm); err != nil {
		return dune.NullValue, err
	}

	l := len(args)
	if l == 0 {
		return dune.NullValue, errors.New("no query provided")
	}

	q, params, err := getRawQuery(s.db.Driver, args[0])
	if err != nil {
		return dune.NullValue, err
	}

	if l > 1 {
		params = append(params, getSqlParams(args[1:])...)
	}

	rows, err := s.db.QueryRaw(q, params...)
	if err != nil {
		return dune.NullValue, err
	}
	if rows == nil {
		return dune.NullValue, nil
	}

	defer rows.Close()

	tbl, err := dbx.ToTable(rows)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(&table{dbxTable: tbl}), nil
}

type schemaColumn struct {
	name     string
	typeName string
	size     int
	decimals int
	nullable bool
}

func (schemaColumn) Type() string {
	return "sql.SchemaColumn"
}

func (c schemaColumn) String() string {
	return c.name + ":sql.SchemaColumn"
}

func (c schemaColumn) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "name":
		return dune.NewString(c.name), nil
	case "type":
		return dune.NewString(c.typeName), nil
	case "nullable":
		return dune.NewBool(c.nullable), nil
	case "size":
		return dune.NewInt(c.size), nil
	case "decimals":
		return dune.NewInt(c.decimals), nil
	}
	return dune.UndefinedValue, nil
}

func newColumn(c dbx.SchemaColumn) (schemaColumn, error) {
	var size, decimals int

	t := c.Type
	i := strings.IndexRune(t, '(')

	if i != -1 {
		sizePart := t[i+1:]
		e := strings.IndexRune(sizePart, ')')
		sizePart = sizePart[:e]
		parts := strings.Split(sizePart, ",")
		t = t[:i]

		s, err := strconv.Atoi(parts[0])
		if err != nil {
			return schemaColumn{}, fmt.Errorf("invalid size for %s: %v", c.Name, err)
		}
		size = s

		if len(parts) > 1 {
			d, err := strconv.Atoi(parts[1])
			if err != nil {
				return schemaColumn{}, fmt.Errorf("invalid size for %s: %v", c.Name, err)
			}
			decimals = d
		}
	}

	col := schemaColumn{
		name:     c.Name,
		nullable: c.Nullable,
		typeName: t,
		size:     size,
		decimals: decimals,
	}

	return col, nil
}

func (s *libDB) queryValues(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	tbl, err := s.getTable(args, vm)
	if err != nil {
		return dune.NullValue, err
	}

	cols := tbl.Columns
	if len(cols) != 1 {
		return dune.NullValue, fmt.Errorf("the query must return 1 column, got %d", len(cols))
	}

	result := make([]dune.Value, len(tbl.Rows))

	for i, r := range tbl.Rows {
		result[i] = convertDBValue(r.Values[0])
	}

	return dune.NewArrayValues(result), nil
}

func (s *libDB) queryValuesRaw(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := s.onExecutingRaw(args, vm); err != nil {
		return dune.NullValue, err
	}

	l := len(args)
	if l == 0 {
		return dune.NullValue, errors.New("no query provided")
	}

	query, params, err := getRawQuery(s.db.Driver, args[0])
	if err != nil {
		return dune.NullValue, err
	}

	if l > 1 {
		params = append(params, getSqlParams(args[1:])...)
	}

	rows, err := s.db.QueryRaw(query, params...)
	if err != nil {
		return dune.NullValue, err
	}
	if rows == nil {
		return dune.NullValue, nil
	}

	defer rows.Close()

	tbl, err := dbx.ToTable(rows)
	if err != nil {
		return dune.NullValue, err
	}

	result := make([]dune.Value, len(tbl.Rows))

	for i, r := range tbl.Rows {
		result[i] = convertDBValue(r.Values[0])
	}

	return dune.NewArrayValues(result), nil
}

func (s *libDB) getTable(args []dune.Value, vm *dune.VM) (*dbx.Table, error) {
	var q sqx.Query
	var err error

	switch len(args) {
	case 0:
		return nil, errors.New("no query provided")
	default:
		if q, err = getQuery(args[0], args[1:]); err != nil {
			return nil, err
		}
	}

	if err := s.onExecuting(q, args[1:], vm); err != nil {
		return nil, err
	}

	var tbl *dbx.Table

	switch q.(type) {
	case *sqx.ShowQuery:
		sQuery, _, err := s.toSQLString(q, vm)
		if err != nil {
			return nil, err
		}
		rows, err := s.db.QueryRaw(sQuery)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		tbl, err = dbx.ToTable(rows)
		if err != nil {
			return nil, err
		}

	case *sqx.SelectQuery:
		sQuery, sParams, err := s.toSQLString(q, vm)
		if err != nil {
			return nil, err
		}
		rows, err := s.db.QueryRaw(sQuery, sParams...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		tbl, err = dbx.ToTable(rows)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("not a select query")
	}

	return tbl, nil
}

func (s *libDB) queryRaw(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !vm.HasPermission("trusted") {
		return dune.NullValue, ErrUnauthorized
	}

	if err := s.onExecutingRaw(args, vm); err != nil {
		return dune.NullValue, err
	}

	l := len(args)
	if l == 0 {
		return dune.NullValue, errors.New("no query provided")
	}

	q, params, err := getRawQuery(s.db.Driver, args[0])
	if err != nil {
		return dune.NullValue, err
	}

	if l > 1 {
		params = append(params, getSqlParams(args[1:])...)
	}

	rows, err := s.db.QueryRaw(q, params...)
	if err != nil {
		return dune.NullValue, err
	}
	if rows == nil {
		return dune.NullValue, nil
	}

	defer rows.Close()

	t, err := dbx.ToTable(rows)
	if err != nil {
		return dune.NullValue, err
	}

	result := make([]dune.Value, len(t.Rows))
	l = len(t.Columns)

	for i, r := range t.Rows {
		m := make(map[dune.Value]dune.Value, l)
		for k, col := range t.Columns {
			m[dune.NewString(col.Name)] = convertDBValue(r.Values[k])
		}
		result[i] = dune.NewMapValues(m)
	}

	return dune.NewArrayValues(result), nil
}

func (s *libDB) reader(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	var q sqx.Query
	var err error

	switch len(args) {
	case 0:
		return dune.NullValue, errors.New("no query provided")
	default:
		if q, err = getQuery(args[0], args[1:]); err != nil {
			return dune.NullValue, err
		}
	}

	if err := s.onExecuting(q, args[1:], vm); err != nil {
		return dune.NullValue, err
	}

	var dbxReader *dbx.Reader

	switch t := q.(type) {
	case *sqx.ShowQuery:
		sQuery, _, err := s.toSQLString(q, vm)
		if err != nil {
			return dune.NullValue, err
		}
		dbxReader, err = s.db.ReaderRaw(sQuery)
		if err != nil {
			return dune.NullValue, err
		}

	case *sqx.SelectQuery:
		sQuery, sParams, err := s.toSQLString(t, vm)
		if err != nil {
			return dune.NullValue, err
		}
		dbxReader, err = s.db.ReaderRaw(sQuery, sParams...)
		if err != nil {
			return dune.NullValue, err
		}

	default:
		return dune.NullValue, fmt.Errorf("not a select query")
	}

	r := newReader(dbxReader, vm)

	return dune.NewObject(r), nil
}

type query struct {
	query sqx.Query
}

func (s query) Type() string {
	return "sql.Query"
}

func (s query) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s query) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s query) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

func toSQL(query sqx.Query, args []dune.Value) (dune.Value, error) {
	var driver string
	var escapeIdents bool
	var format bool

	l := len(args)

	if l == 0 {
		driver = "mysql"
		format = true
	} else if l == 1 {
		driver = "mysql"
		if args[0].Type != dune.Bool {
			return dune.NullValue, fmt.Errorf("expected argument 1 to be a boolean, got %s", args[0].TypeName())
		}
		format = args[0].ToBool()
	} else {
		format = args[0].ToBool()
	}

	if l > 1 {
		if args[1].Type != dune.String {
			return dune.NullValue, fmt.Errorf("expected argument 2 to be a string, got %s", args[1].TypeName())
		}
		driver = args[1].String()
	}

	if l > 2 {
		if args[2].Type != dune.Bool {
			return dune.NullValue, fmt.Errorf("expected argument 3 to be a boolean, got %s", args[2].TypeName())
		}
		escapeIdents = args[2].ToBool()
	}

	if l > 3 {
		return dune.NullValue, fmt.Errorf("expected max 3 arguments, got %d", len(args))
	}

	w := sqx.NewWriter(query, "", "", "", driver)
	w.EscapeIdents = escapeIdents
	w.Format = format
	v, _, err := w.Write()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewString(v), nil
}

type showQuery struct {
	query *sqx.ShowQuery
}

func (s showQuery) Type() string {
	return "sql.ShowQuery"
}

func (s showQuery) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s showQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s showQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type createTableQuery struct {
	query *sqx.CreateTableQuery
}

func (s createTableQuery) Type() string {
	return "sql.CreateTableQuery"
}

func (s createTableQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s createTableQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type createDatabaseQuery struct {
	query *sqx.CreateDatabaseQuery
}

func (s createDatabaseQuery) Type() string {
	return "sql.CreateDatabaseQuery"
}

func (s createDatabaseQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s createDatabaseQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type modifyColumnQuery struct {
	query *sqx.ModifyColumnQuery
}

func (s modifyColumnQuery) Type() string {
	return "sql.ModifyColumnQuery"
}

func (s modifyColumnQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s modifyColumnQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type alterDropQuery struct {
	query *sqx.AlterDropQuery
}

func (s alterDropQuery) Type() string {
	return "sql.AlterDropQuery"
}

func (s alterDropQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s alterDropQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type dropTableQuery struct {
	query *sqx.DropTableQuery
}

func (s dropTableQuery) Type() string {
	return "sql.DropTableQuery"
}

func (s dropTableQuery) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s dropTableQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s dropTableQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type renameColumnQuery struct {
	query *sqx.RenameColumnQuery
}

func (s renameColumnQuery) Type() string {
	return "sql.RenameColumnQuery"
}

func (s renameColumnQuery) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s renameColumnQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s renameColumnQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type addColumnQuery struct {
	query *sqx.AddColumnQuery
}

func (s addColumnQuery) Type() string {
	return "sql.AddColumnQuery"
}

func (s addColumnQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s addColumnQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type dropDatabaseQuery struct {
	query *sqx.DropDatabaseQuery
}

func (s dropDatabaseQuery) Type() string {
	return "sql.DropDatabaseQuery"
}

func (s dropDatabaseQuery) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s dropDatabaseQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s dropDatabaseQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type addConstraintQuery struct {
	query *sqx.AddConstraintQuery
}

func (s addConstraintQuery) Type() string {
	return "sql.AddConstraintQuery"
}

func (s addConstraintQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s addConstraintQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type addFKQuery struct {
	query *sqx.AddFKQuery
}

func (s addFKQuery) Type() string {
	return "sql.AddFKQuery"
}

func (s addFKQuery) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s addFKQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s addFKQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type deleteQuery struct {
	query *sqx.DeleteQuery
}

func (s deleteQuery) Type() string {
	return "sql.DeleteQuery"
}

func (s deleteQuery) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s deleteQuery) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "hasWhere":
		return dune.NewBool(s.query.WherePart != nil), nil
	case "hasLimit":
		return dune.NewBool(s.query.LimitPart != nil), nil
	case "parameters":
		return dune.NewArrayValues(getParamers(s.query)), nil
	}
	return dune.UndefinedValue, nil
}

func (s deleteQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "where":
		return s.where
	case "and":
		return s.and
	case "or":
		return s.or
	case "limit":
		return s.limit
	case "toSQL":
		return s.toSQL
	case "join":
		return s.join
	}
	return nil
}

func (s deleteQuery) or(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}

	filter := args[0]

	// the filter can be a query object
	if filter.Type == dune.Object {
		f, ok := filter.ToObject().(selectQuery)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected argument to be a string or query, got %s", args[0].TypeName())
		}

		// when passing a query object the parameters are contained in the object
		if l > 1 {
			return dune.NullValue, fmt.Errorf("expected only 1 argument, got %d", l)
		}
		s.query.OrQuery(f.query)
		return dune.NewObject(s), nil
	}

	if filter.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", filter.Type)
	}

	v := filter.String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Or(v, params...); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s deleteQuery) and(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}

	filter := args[0]

	// the filter can be a query object
	if filter.Type == dune.Object {
		f, ok := filter.ToObject().(selectQuery)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected argument to be a string or query, got %s", args[0].TypeName())
		}

		// when passing a query object the parameters are contained in the object
		if l > 1 {
			return dune.NullValue, fmt.Errorf("expected only 1 argument, got %d", l)
		}
		s.query.AndQuery(f.query)
		return dune.NewObject(s), nil
	}

	// If its not an object the filter must be a string
	if filter.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string or query, got %s", args[0].TypeName())
	}

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	v := filter.String()

	if err := s.query.And(v, params...); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s deleteQuery) where(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}

	v := args[0].String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Where(v, params...); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s deleteQuery) limit(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 || l > 2 {
		return dune.NullValue, fmt.Errorf("expected 1 or 2 arguments, got %d", len(args))
	}
	if args[0].Type != dune.Int {
		// if the argument is null clear the limit
		if args[0].Type == dune.Null {
			s.query.LimitPart = nil
			return dune.NewObject(s), nil
		}
		return dune.NullValue, fmt.Errorf("expected argument 1 to be a int, got %d", args[0].Type)
	}

	if l == 2 {
		if args[1].Type != dune.Int {
			return dune.NullValue, fmt.Errorf("expected argument 2 to be a int, got %d", args[1].Type)
		}
	}

	switch l {
	case 1:
		s.query.Limit(int(args[0].ToInt()))
	case 2:
		s.query.LimitOffset(int(args[0].ToInt()), int(args[1].ToInt()))
	}

	return dune.NewObject(s), nil
}

func (s deleteQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

func (s deleteQuery) join(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}
	v := args[0].String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Join(v, params...); err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(s), nil
}

type insertQuery struct {
	query *sqx.InsertQuery
}

func (s insertQuery) Type() string {
	return "sql.InsertQuery"
}

func (s insertQuery) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s insertQuery) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "parameters":
		return dune.NewArrayValues(getParamers(s.query)), nil
	}
	return dune.UndefinedValue, nil
}

func (s insertQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "addColumn":
		return s.addColumn
	case "toSQL":
		return s.toSQL
	}
	return nil
}

func (s insertQuery) addColumn(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l != 2 {
		return dune.NullValue, fmt.Errorf("expected 2 arguments, got %d", l)
	}

	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument 1 to be a string, got %v", args[0].Type)
	}

	err := s.query.AddColumn(args[0].String(), args[1].Export(0))
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func (s insertQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

type updateQuery struct {
	query *sqx.UpdateQuery
}

func (s updateQuery) Type() string {
	return "sql.UpdateQuery"
}

func (s updateQuery) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s updateQuery) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "hasWhere":
		return dune.NewBool(s.query.WherePart != nil), nil
	case "hasLimit":
		return dune.NewBool(s.query.LimitPart != nil), nil
	case "parameters":
		return dune.NewArrayValues(getParamers(s.query)), nil
	}
	return dune.UndefinedValue, nil
}

func getParamers(q sqx.Query) []dune.Value {
	params := q.Parameters()
	result := make([]dune.Value, len(params))
	for i, v := range params {

		result[i] = getValue(v)
	}
	return result
}

func getValue(v interface{}) dune.Value {
	switch t := v.(type) {
	case dune.Value:
		return t
	case time.Time:
		return dune.NewObject(TimeObj(t))
	case nil:
		return dune.NullValue
	case int:
		return dune.NewInt(t)
	case int64:
		return dune.NewInt64(t)
	case float64:
		return dune.NewFloat(t)
	case rune:
		return dune.NewRune(t)
	case bool:
		return dune.NewBool(t)
	case []byte:
		return dune.NewBytes(t)
	case string:
		return dune.NewString(t)
	default:
		panic(fmt.Sprintf("Invalid type value %T: %v", t, t))
		//return Object(t)
	}
}

func (s updateQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "where":
		return s.where
	case "and":
		return s.and
	case "or":
		return s.or
	case "limit":
		return s.limit
	case "toSQL":
		return s.toSQL
	case "addColumns":
		return s.addColumns
	case "setColumns":
		return s.setColumns
	case "join":
		return s.join
	}
	return nil
}

func (s updateQuery) or(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}

	v := args[0].String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Or(v, params...); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s updateQuery) and(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}

	filter := args[0]

	// the filter can be a query object
	if filter.Type == dune.Object {
		f, ok := filter.ToObject().(selectQuery)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected argument to be a string or query, got %s", args[0].TypeName())
		}

		// when passing a query object the parameters are contained in the object
		if l > 1 {
			return dune.NullValue, fmt.Errorf("expected only 1 argument, got %d", l)
		}
		s.query.AndQuery(f.query)
		return dune.NewObject(s), nil
	}

	// If its not an object the filter must be a string
	if filter.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string or query, got %s", args[0].TypeName())
	}

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	v := filter.String()

	if err := s.query.And(v, params...); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s updateQuery) where(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}

	v := args[0].String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Where(v, params...); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s updateQuery) limit(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 || l > 2 {
		return dune.NullValue, fmt.Errorf("expected 1 or 2 arguments, got %d", len(args))
	}
	if args[0].Type != dune.Int {
		// if the argument is null clear the limit
		if args[0].Type == dune.Null {
			s.query.LimitPart = nil
			return dune.NewObject(s), nil
		}
		return dune.NullValue, fmt.Errorf("expected argument 1 to be a int, got %d", args[0].Type)
	}

	if l == 2 {
		if args[1].Type != dune.Int {
			return dune.NullValue, fmt.Errorf("expected argument 2 to be a int, got %d", args[1].Type)
		}
	}

	switch l {
	case 1:
		s.query.Limit(int(args[0].ToInt()))
	case 2:
		s.query.LimitOffset(int(args[0].ToInt()), int(args[1].ToInt()))
	}

	return dune.NewObject(s), nil
}

func (s updateQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

func (s updateQuery) addColumns(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.AddColumns(args[0].String(), params...); err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(s), nil
}

func (s updateQuery) setColumns(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}

	a := args[0]

	switch a.Type {
	case dune.String:
		if args[0].Type != dune.String {
			return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
		}

		var params []interface{}
		if l > 1 {
			params = make([]interface{}, l-1)
			for i, v := range args[1:] {
				params[i] = v.Export(0)
			}
		}
		if err := s.query.SetColumns(a.String(), params...); err != nil {
			return dune.NullValue, err
		}
	case dune.Null:
		s.query.Columns = nil
	default:
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", a.Type)
	}

	return dune.NewObject(s), nil
}

func (s updateQuery) join(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}
	v := args[0].String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Join(v, params...); err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(s), nil
}

type selectQuery struct {
	query *sqx.SelectQuery
}

func (s selectQuery) Type() string {
	return "sql.SelectQuery"
}

func (s selectQuery) String() string {
	v, err := toSQL(s.query, nil)
	if err != nil {
		return err.Error()
	}
	return v.String()
}

func (s selectQuery) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "columnsLength":
		return dune.NewInt(len(s.query.Columns)), nil
	case "hasLimit":
		return dune.NewBool(s.query.LimitPart != nil), nil
	case "hasFrom":
		return dune.NewBool(s.query.From != nil), nil
	case "hasWhere":
		return dune.NewBool(s.query.WherePart != nil), nil
	case "hasDistinct":
		return dune.NewBool(s.query.Distinct), nil
	case "hasOrderBy":
		return dune.NewBool(s.query.OrderByPart != nil), nil
	case "hasUnion":
		return dune.NewBool(s.query.UnionPart != nil), nil
	case "hasGroupBy":
		return dune.NewBool(s.query.GroupByPart != nil), nil
	case "hasHaving":
		return dune.NewBool(s.query.HavingPart != nil), nil
	case "parameters":
		return dune.NewArrayValues(getParamers(s.query)), nil
	}
	return dune.UndefinedValue, nil
}

func (s selectQuery) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "addColumns":
		return s.addColumns
	case "setColumns":
		return s.setColumns
	case "from":
		return s.from
	case "fromExpr":
		return s.fromExpr
	case "limit":
		return s.limit
	case "orderBy":
		return s.orderBy
	case "where":
		return s.where
	case "having":
		return s.having
	case "and":
		return s.and
	case "or":
		return s.or
	case "join":
		return s.join
	case "groupBy":
		return s.groupBy
	case "setFilter":
		return s.setFilter
	case "toSQL":
		return s.toSQL
	case "getFilterColumns":
		return s.getFilterColumns
	}
	return nil
}

func (s selectQuery) getFilterColumns(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 argument, got %d", len(args))
	}

	if s.query.WherePart == nil {
		return dune.NewArray(0), nil
	}

	cols := sqx.NameExprColumns(s.query.WherePart.Expr)

	vals := make([]dune.Value, len(cols))

	for i, col := range cols {
		vals[i] = dune.NewString(col.Name)
	}

	return dune.NewArrayValues(vals), nil
}

// setFilter copies all the elements of the query from the Where part
func (s selectQuery) setFilter(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	a := args[0]
	var o selectQuery

	switch a.Type {
	case dune.Null, dune.Undefined:
		// setting it null clears the filter
		s.query.WherePart = nil
		return dune.NewObject(s), nil

	case dune.Object:
		var ok bool
		o, ok = a.ToObject().(selectQuery)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected argument 1 to be a sql.SelectQuery, got %v", a.Type)
		}
	default:
		return dune.NullValue, fmt.Errorf("expected argument 1 to be a sql.SelectQuery, got %v", a.Type)
	}

	dst := s.query
	src := o.query

	dst.WherePart = src.WherePart

	if src.OrderByPart != nil {
		dst.OrderByPart = src.OrderByPart
	}

	if src.GroupByPart != nil {
		dst.GroupByPart = src.GroupByPart
	}

	if src.HavingPart != nil {
		dst.HavingPart = src.HavingPart
	}

	if src.LimitPart != nil {
		dst.LimitPart = src.LimitPart
	}

	if src.UnionPart != nil {
		dst.UnionPart = src.UnionPart
	}

	return dune.NewObject(s), nil
}

func (s selectQuery) toSQL(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return toSQL(s.query, args)
}

func (s selectQuery) groupBy(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	a := args[0]

	if a.IsNil() {
		s.query.GroupByPart = nil
		return dune.NewObject(s), nil
	}

	if a.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}

	v := a.String()

	if err := s.query.GroupBy(v); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s selectQuery) join(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}
	v := args[0].String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Join(v, params...); err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(s), nil
}

func (s selectQuery) or(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}

	filter := args[0]

	// the filter can be a query object
	if filter.Type == dune.Object {
		f, ok := filter.ToObject().(selectQuery)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected argument to be a string or query, got %s", args[0].TypeName())
		}

		// when passing a query object the parameters are contained in the object
		if l > 1 {
			return dune.NullValue, fmt.Errorf("expected only 1 argument, got %d", l)
		}
		s.query.OrQuery(f.query)
		return dune.NewObject(s), nil
	}

	if filter.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", filter.Type)
	}

	v := filter.String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Or(v, params...); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s selectQuery) and(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}

	filter := args[0]

	// the filter can be a query object
	if filter.Type == dune.Object {
		f, ok := filter.ToObject().(selectQuery)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected argument to be a string or query, got %s", args[0].TypeName())
		}

		// when passing a query object the parameters are contained in the object
		if l > 1 {
			return dune.NullValue, fmt.Errorf("expected only 1 argument, got %d", l)
		}
		s.query.AndQuery(f.query)
		return dune.NewObject(s), nil
	}

	// If its not an object the filter must be a string
	if filter.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string or query, got %s", args[0].TypeName())
	}

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	v := filter.String()

	if err := s.query.And(v, params...); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s selectQuery) where(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}

	v := args[0].String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Where(v, params...); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s selectQuery) orderBy(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	a := args[0]

	if a.IsNil() {
		s.query.OrderByPart = nil
		return dune.NewObject(s), nil
	}

	if a.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}
	v := a.String()

	if err := s.query.OrderBy(v); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s selectQuery) having(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}

	v := args[0].String()

	var params []interface{}
	if l > 1 {
		params = make([]interface{}, l-1)
		for i, v := range args[1:] {
			params[i] = v.Export(0)
		}
	}

	if err := s.query.Having(v, params...); err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(s), nil
}

func (s selectQuery) fromExpr(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object, dune.String); err != nil {
		return dune.NullValue, err
	}

	sel, ok := args[0].ToObjectOrNil().(selectQuery)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected select expression, got %T", args[0].ToObjectOrNil())
	}

	alias := args[1].String()

	parenExp := &sqx.ParenExpr{X: sel.query}

	exp := &sqx.FromAsExpr{From: parenExp, Alias: alias}

	s.query.From = append(s.query.From, exp)

	return dune.NewObject(s), nil
}

func (s selectQuery) from(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}

	v := "from " + args[0].String()

	if err := s.query.SetFrom(v); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s selectQuery) addColumns(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].Type)
	}
	v := args[0].String()

	if err := s.query.AddColumns(v); err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(s), nil
}

func (s selectQuery) setColumns(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	a := args[0]

	switch a.Type {
	case dune.String:
		if err := s.query.SetColumns(a.String()); err != nil {
			return dune.NullValue, err
		}
	case dune.Null:
		s.query.Columns = nil
	default:
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", a.Type)
	}

	return dune.NewObject(s), nil
}

func (s selectQuery) limit(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 || l > 2 {
		return dune.NullValue, fmt.Errorf("expected 1 or 2 arguments, got %d", len(args))
	}
	if args[0].Type != dune.Int {
		// if the argument is null clear the limit
		if args[0].Type == dune.Null {
			s.query.LimitPart = nil
			return dune.NewObject(s), nil
		}
		return dune.NullValue, fmt.Errorf("expected argument 1 to be a int, got %d", args[0].Type)
	}

	if l == 2 {
		if args[1].Type != dune.Int {
			return dune.NullValue, fmt.Errorf("expected argument 2 to be a int, got %d", args[1].Type)
		}
	}

	switch l {
	case 1:
		s.query.Limit(int(args[0].ToInt()))
	case 2:
		s.query.LimitOffset(int(args[0].ToInt()), int(args[1].ToInt()))
	}

	return dune.NewObject(s), nil
}

type table struct {
	dbxTable *dbx.Table
}

func (t *table) Type() string {
	return "sql.Table"
}

func (t *table) Export(recursionLevel int) interface{} {
	if t.dbxTable.Rows == nil {
		t.dbxTable.Rows = make([]*dbx.Row, 0)
	}
	return t.dbxTable
}

func (t *table) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "length":
		return dune.NewInt(len(t.dbxTable.Rows)), nil
	case "rows":
		return dune.NewObject(&rows{t.dbxTable}), nil
	case "columns":
		return dune.NewObject(columns{t.dbxTable.Columns}), nil
	}
	return dune.UndefinedValue, nil
}

type rows struct {
	table *dbx.Table
}

func (r *rows) Type() string {
	return "sql.Rows"
}

func (r *rows) Len() int {
	return len(r.table.Rows)
}

func (r *rows) Export(recursionLevel int) interface{} {
	return r.table.Rows
}

func (r *rows) Values() ([]dune.Value, error) {
	t := r.table
	rows := t.Rows
	values := make([]dune.Value, len(rows))
	for i, v := range rows {
		values[i] = dune.NewObject(newRow(v, t))
	}
	return values, nil
}

func (r *rows) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "length":
		return dune.NewInt(len(r.table.Rows)), nil
	}
	return dune.UndefinedValue, nil
}

func (r *rows) GetIndex(i int) (dune.Value, error) {
	t := r.table

	if i >= len(t.Rows) {
		return dune.NullValue, fmt.Errorf("index out of range")
	}

	return dune.NewObject(newRow(t.Rows[i], t)), nil
}

func newRow(r *dbx.Row, table *dbx.Table) *row {
	return &row{table: table, dbxRow: r}
}

type row struct {
	table  *dbx.Table
	dbxRow *dbx.Row
}

func (r *row) Type() string {
	return "sql.Row"
}

func (r *row) Export(recursionLevel int) interface{} {
	return r.dbxRow
}

func (r *row) Values() ([]dune.Value, error) {
	vs := r.dbxRow.Values
	values := make([]dune.Value, len(vs))
	for i, v := range vs {
		values[i] = convertDBValue(v)
	}
	return values, nil
}

func (r *row) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	// first look for values from the database
	if v, ok := r.dbxRow.Value(name); ok {
		return convertDBValue(v), nil
	}

	// custom properties
	switch name {
	case "length":
		return dune.NewInt(len(r.dbxRow.Values)), nil
	case "columns":
		return dune.NewObject(columns{r.table.Columns}), nil
	}

	return dune.UndefinedValue, nil
}

func (r *row) SetProperty(name string, v dune.Value, vm *dune.VM) error {
	i := r.dbxRow.ColumnIndex(name)
	if i != -1 {
		r.dbxRow.Values[i] = v.Export(0)
		return nil
	}
	return fmt.Errorf("column %s not exists", name)
}

func (r *row) SetIndex(i int, v dune.Value) error {
	if i >= len(r.dbxRow.Values) {
		return fmt.Errorf("index out of range")
	}
	r.dbxRow.Values[i] = v.Export(0)
	return nil
}

func (r *row) GetIndex(i int) (dune.Value, error) {
	if i >= len(r.dbxRow.Values) {
		return dune.NullValue, fmt.Errorf("index out of range")
	}

	return convertDBValue(r.dbxRow.Values[i]), nil
}
