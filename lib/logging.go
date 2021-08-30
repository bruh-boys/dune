package lib

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dunelang/dune/filesystem"
	"github.com/dunelang/dune/lib/logging"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Log, `

declare namespace logging {	
    export const defaultLogger: Logger
	export function setDefaultLogger(logger: Logger): void
	
    export function fatal(format: any, ...v: any[]): void
    export function system(format: any, ...v: any[]): void
    export function write(table: string, format: any, ...v: any[]): void

    export function newLogger(path: string, fs?: io.FileSystem): Logger

    export interface Logger {
        path: string
		debug: boolean
        save(table: string, data: string, ...v: any): void
        query(table: string, start: time.Time, end: time.Time, offset?: number, limit?: number): Scanner
    }

    export interface Scanner {
        scan(): boolean
        data(): DataPoint
        setFilter(v: string): void
    }

    export interface DataPoint {
        text: string
        time: time.Time
        string(): string
    }
}
`)
}

var defaultLogger *logger

var Log = []dune.NativeFunction{
	{
		Name:        "->logging.defaultLogger",
		Arguments:   0,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if defaultLogger == nil {
				return dune.NullValue, nil
			}
			return dune.NewObject(defaultLogger), nil
		},
	},
	{
		Name:        "logging.setDefaultLogger",
		Arguments:   1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			db, ok := args[0].ToObjectOrNil().(*logger)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected a logging, got %s", args[0].TypeName())
			}

			defaultLogger = db

			return dune.NullValue, nil
		},
	},
	{
		Name:        "logging.fatal",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			err := writeLog("system", args)
			if err != nil {
				return dune.NullValue, err
			}

			os.Exit(1)
			return dune.NullValue, nil
		},
	},
	{
		Name:        "logging.write",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l < 2 {
				return dune.NullValue, fmt.Errorf("expected at least 2 parameters, got %d", len(args))
			}

			name := args[0]
			if name.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be a string, got %s", name.Type)
			}

			err := writeLog(name.String(), args[1:])
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:        "logging.system",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if len(args) == 0 {
				return dune.NullValue, fmt.Errorf("expected at least 1 parameter, got 0")
			}

			err := writeLog("system", args)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:        "logging.newLogger",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String, dune.Object); err != nil {
				return dune.NullValue, err
			}

			ln := len(args)
			if ln == 0 || ln > 2 {
				return dune.NullValue, fmt.Errorf("expected 1 or 2 arguments, got %d", ln)
			}

			path := args[0].String()

			var fs filesystem.FS

			if ln == 2 {
				afs, ok := args[1].ToObjectOrNil().(*FileSystemObj)
				if !ok {
					return dune.NullValue, fmt.Errorf("invalid argument 2 type: %s", args[1].TypeName())
				}
				fs = afs.FS
			} else {
				fs = filesystem.OS
			}

			t := &logger{
				db: logging.New(path, fs),
			}

			return dune.NewObject(t), nil
		},
	},
}

func toStringLog(v dune.Value) string {
	switch v.Type {
	case dune.Null:
		return "null"
	case dune.String:
		// need to escape the % to prevent interfering with fmt
		return strings.Replace(v.String(), "%", "%%", -1)
	}

	return v.String()
}

func writeLog(table string, args []dune.Value) error {
	var line string

	ln := len(args)
	if ln == 1 {
		line = toStringLog(args[0])
	} else {
		format := args[0].String()
		values := make([]interface{}, ln-1)
		for i, v := range args[1:] {
			values[i] = toStringLog(v)
		}
		line = fmt.Sprintf(format, values...)
	}

	if defaultLogger == nil {
		fmt.Println(line)
		return nil
	}

	return defaultLogger.db.Save(table, line)
}

type logger struct {
	db    *logging.Logger
	debug bool
}

func (*logger) Type() string {
	return "logging.Logger"
}

func (t *logger) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "path":
		return dune.NewString(t.db.Path), nil
	}
	return dune.UndefinedValue, nil
}

func (t *logger) SetField(name string, v dune.Value, vm *dune.VM) error {
	if !vm.HasPermission("trusted") {
		return ErrUnauthorized
	}

	switch name {
	case "debug":
		if v.Type != dune.Bool {
			return ErrInvalidType
		}
		t.debug = v.ToBool()
		return nil
	}

	return ErrReadOnlyOrUndefined
}

func (t *logger) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "save":
		return t.save
	case "query":
		return t.query
	case "close":
		return t.close
	}
	return nil
}

func (t *logger) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}
	err := t.db.Close()
	return dune.NullValue, err
}

func (t *logger) save(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}
	err := t.db.Save(args[0].String(), args[1].String())
	return dune.NullValue, err
}

func (t *logger) query(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.String, dune.Object, dune.Object, dune.Int, dune.Int); err != nil {
		return dune.NullValue, err
	}

	var table string
	var start, end time.Time
	var offset, limit int

	l := len(args)

	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected the table")
	}

	table = args[0].String()

	if l == 1 {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		end = now
	} else if l == 2 {
		s, ok := args[1].ToObjectOrNil().(TimeObj)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected time, got %s", args[1].TypeName())
		}
		start = time.Time(s)
		end = time.Now()
	} else if l >= 3 {
		s, ok := args[1].ToObjectOrNil().(TimeObj)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected time, got %s", args[1].TypeName())
		}
		start = time.Time(s)
		s, ok = args[2].ToObjectOrNil().(TimeObj)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected time, got %s", args[2].TypeName())
		}
		end = time.Time(s)
	}

	if l >= 4 {
		offset = int(args[3].ToInt())
	}

	if l >= 5 {
		limit = int(args[4].ToInt())
	}

	scanner := t.db.Query(table, start, end, offset, limit)
	return dune.NewObject(&logScanner{scanner}), nil
}

type logScanner struct {
	s *logging.Scanner
}

func (*logScanner) Type() string {
	return "logging.Scanner"
}

func (s *logScanner) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "scan":
		return s.scan
	case "data":
		return s.data
	case "setFilter":
		return s.setFilter
	}
	return nil
}

func (s *logScanner) setFilter(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected a string arg")
	}

	a := args[0]
	if a.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected a string arg")
	}

	s.s.SetFilter(a.String())
	return dune.NullValue, nil
}

func (s *logScanner) scan(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	ok := s.s.Scan()
	return dune.NewBool(ok), nil
}

func (s *logScanner) data(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	dp, _ := s.s.Data()
	return dune.NewObject(&dataPoint{dp}), nil
}

type dataPoint struct {
	d logging.DataPoint
}

func (*dataPoint) Type() string {
	return "logging.DataPoint"
}

func (d *dataPoint) String() string {
	return d.d.String()
}

func (d *dataPoint) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "text":
		return dune.NewString(d.d.Text), nil
	case "time":
		return dune.NewObject(TimeObj(d.d.Time)), nil
	}
	return dune.UndefinedValue, nil
}

func (d *dataPoint) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "string":
		return d.string
	}
	return nil
}

func (d *dataPoint) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}
	return dune.NewString(d.d.String()), nil
}
