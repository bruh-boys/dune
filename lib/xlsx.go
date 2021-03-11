package lib

import (
	"fmt"
	"io"
	"time"

	"github.com/dunelang/dune"

	"github.com/tealeg/xlsx"
)

func init() {
	dune.RegisterLib(XLSX, `


declare namespace xlsx {
    export function newFile(): XLSXFile
    export function openFile(path: string): XLSXFile
    export function openFile(file: io.File): XLSXFile
    export function openReaderAt(r: io.ReaderAt, size: number): XLSXFile 
    export function openBinary(file: io.File): XLSXFile
    export function newStyle(): Style

    export interface XLSXFile {
        sheets: XLSXSheet[]
        addSheet(name: string): XLSXSheet
        save(path?: string): void
        write(w: io.Writer): void
    }

    export interface XLSXSheet {
        rows: XLSXRow[]
        col(i: number): Col
        addRow(): XLSXRow
    }

    export interface Col {
        width: number
    }

    export interface XLSXRow {
        cells: XLSXCell[]
        height: number
        addCell(v?: any): XLSXCell
    }

    export interface XLSXCell {
        value: any
        numberFormat: string
        style: Style
        getDate(): time.Time
        merge(hCells: number, vCells: number): void
    }

    export interface Style {
        alignment: Alignment
        applyAlignment: boolean
        font: Font
        applyFont: boolean
    }

    export interface Alignment {
        horizontal: string
        vertical: string
    }

    export interface Font {
        bold: boolean
        size: number
    }
}

`)
}

var XLSX = []dune.NativeFunction{
	{
		Name:      "xlsx.openReaderAt",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object, dune.Int); err != nil {
				return dune.NullValue, err
			}

			r, ok := args[0].ToObjectOrNil().(io.ReaderAt)
			if !ok {
				return dune.NullValue, fmt.Errorf("invalid argument type. Expected a io.ReaderAt, got %s", args[0].TypeName())
			}

			size := args[1].ToInt()

			reader, err := xlsx.OpenReaderAt(r, size)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(&xlsxFile{obj: reader}), nil
		},
	},
	{
		Name:      "xlsx.openBinary",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			b := args[0].ToBytes()

			reader, err := xlsx.OpenBinary(b)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(&xlsxFile{obj: reader}), nil
		},
	},
	{
		Name:      "xlsx.openFile",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			var r io.ReaderAt
			var size int64

			a := args[0]

			switch a.Type {
			case dune.Object:
				f, ok := a.ToObject().(*file)
				if !ok {
					return dune.NullValue, fmt.Errorf("invalid argument type. Expected a io.ReaderAt, got %s", a.TypeName())
				}
				r = f
				st, err := f.Stat()
				if err != nil {
					return dune.NullValue, err
				}
				size = st.Size()

			case dune.String:
				f, err := vm.FileSystem.Open(a.ToString())
				if err != nil {
					return dune.NullValue, err
				}
				r = f
				st, err := f.Stat()
				if err != nil {
					return dune.NullValue, err
				}
				size = st.Size()

			default:
				return dune.NullValue, fmt.Errorf("invalid argument type. Expected a io.ReaderAt, got %s", a.TypeName())
			}

			reader, err := xlsx.OpenReaderAt(r, size)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(&xlsxFile{obj: reader, path: a.ToString()}), nil
		},
	},
	{
		Name:      "xlsx.newFile",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args); err != nil {
				return dune.NullValue, err
			}

			file := xlsx.NewFile()
			return dune.NewObject(&xlsxFile{obj: file}), nil
		},
	},
	{
		Name:      "xlsx.newStyle",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args); err != nil {
				return dune.NullValue, err
			}

			s := xlsx.NewStyle()
			return dune.NewObject(&xlsxStyle{obj: s}), nil
		},
	},
}

type xlsxFile struct {
	path string
	obj  *xlsx.File
}

func (f *xlsxFile) Type() string {
	return "xlsx.Reader"
}

func (x *xlsxFile) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "sheets":
		xSheets := x.obj.Sheets
		sheets := make([]dune.Value, len(xSheets))
		for i, c := range xSheets {
			sheets[i] = dune.NewObject(&xlsxSheet{obj: c})
		}
		return dune.NewArrayValues(sheets), nil
	}
	return dune.UndefinedValue, nil
}

func (x *xlsxFile) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "addSheet":
		return x.addSheet
	case "save":
		return x.save
	case "write":
		return x.write
	}
	return nil
}

func (x *xlsxFile) addSheet(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].ToString()
	xSheet, err := x.obj.AddSheet(name)
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(&xlsxSheet{obj: xSheet}), nil
}

func (x *xlsxFile) save(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	var path string
	if len(args) == 1 {
		path = args[0].ToString()
	} else {
		path = x.path
	}

	if path == "" {
		return dune.NullValue, fmt.Errorf("need a name to save the file")
	}

	f, err := vm.FileSystem.OpenForWrite(path)
	if err != nil {
		return dune.NullValue, err
	}

	if err := x.obj.Write(f); err != nil {
		return dune.NullValue, err
	}

	err = f.Close()
	return dune.NullValue, err
}

func (x *xlsxFile) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	w, ok := args[0].ToObjectOrNil().(io.Writer)
	if !ok {
		return dune.NullValue, fmt.Errorf("invalid argument type. Expected a io.Writer, got %s", args[0].TypeName())
	}

	err := x.obj.Write(w)

	return dune.NullValue, err
}

type xlsxSheet struct {
	obj  *xlsx.Sheet
	rows []dune.Value // important to cache this for large files
}

func (x *xlsxSheet) Type() string {
	return "xlsx.Sheet"
}

func (x *xlsxSheet) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "rows":
		if x.rows == nil {
			xRows := x.obj.Rows
			rows := make([]dune.Value, len(xRows))
			for i, c := range xRows {
				rows[i] = dune.NewObject(&xlsxRow{c})
			}
			x.rows = rows
		}
		return dune.NewArrayValues(x.rows), nil
	}
	return dune.UndefinedValue, nil
}

func (x *xlsxSheet) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "addRow":
		return x.addRow
	case "col":
		return x.col
	}
	return nil
}

func (x *xlsxSheet) col(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Int); err != nil {
		return dune.NullValue, err
	}
	i := args[0].ToInt()
	col := x.obj.Col(int(i))
	return dune.NewObject(&xlsxCol{col}), nil
}

func (x *xlsxSheet) addRow(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}
	xRow := x.obj.AddRow()
	return dune.NewObject(&xlsxRow{xRow}), nil
}

type xlsxRow struct {
	obj *xlsx.Row
}

func (x *xlsxRow) Type() string {
	return "xlsx.Row"
}

func (x *xlsxRow) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "cells":
		xCells := x.obj.Cells
		cells := make([]dune.Value, len(xCells))
		for i, c := range xCells {
			cells[i] = dune.NewObject(&xlsxCell{c})
		}
		return dune.NewArrayValues(cells), nil
	case "height":
		return dune.NewFloat(x.obj.Height), nil
	}
	return dune.UndefinedValue, nil
}

func (x *xlsxRow) SetProperty(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "height":
		switch v.Type {
		case dune.Float:
		case dune.Int:
			x.obj.SetHeight(v.ToFloat())
			return nil
		default:
			return ErrInvalidType
		}
	}

	return ErrReadOnlyOrUndefined
}

func (x *xlsxRow) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "addCell":
		return x.addCell
	}
	return nil
}

func (x *xlsxRow) addCell(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)

	if l > 1 {
		return dune.NullValue, fmt.Errorf("expected 0 or 1 arguments, got %d", l)
	}

	xCell := x.obj.AddCell()
	cell := &xlsxCell{xCell}

	if l == 1 {
		if err := cell.setValue(args[0], vm); err != nil {
			return dune.NullValue, err
		}
	}

	return dune.NewObject(cell), nil
}

type xlsxCell struct {
	obj *xlsx.Cell
}

func (x *xlsxCell) Type() string {
	return "xlsx.Cell"
}

func (x *xlsxCell) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "getDate":
		return x.getDate
	case "merge":
		return x.merge
	}
	return nil
}

func (x *xlsxCell) merge(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Int, dune.Int); err != nil {
		return dune.NullValue, err
	}
	x.obj.Merge(int(args[0].ToInt()), int(args[1].ToInt()))
	return dune.NullValue, nil
}

func (x *xlsxCell) getDate(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}
	cell := x.obj
	t, err := cell.GetTime(false)
	if err != nil {
		if cell.Value == "" {
			return dune.NullValue, nil
		}
		return dune.NullValue, err
	}

	// many times the value es like: 2019-09-28 08:17:59.999996814 +0000 UTC
	t = t.Round(time.Millisecond)

	loc := GetLocation(vm)

	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)

	return dune.NewObject(TimeObj(t)), nil
}

func (x *xlsxCell) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {

	case "numberFormat":
		return dune.NewString(x.obj.GetNumberFormat()), nil

	case "value":
		cell := x.obj
		if cell.Value == "" {
			return dune.NullValue, nil
		}
		switch cell.Type() {
		case xlsx.CellTypeNumeric:
			f, err := cell.Float()
			if err != nil {
				return dune.NullValue, err
			}
			i := int64(f)
			if f == float64(i) {
				return dune.NewInt64(i), nil
			}
			return dune.NewFloat(f), nil

		case xlsx.CellTypeBool:
			return dune.NewBool(cell.Bool()), nil

		case xlsx.CellTypeDate:
			t, err := cell.GetTime(false)
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewObject(TimeObj(t)), nil

		default:
			return dune.NewString(cell.String()), nil
		}

	case "style":
		return dune.NewObject(&xlsxStyle{x.obj.GetStyle()}), nil
	}
	return dune.UndefinedValue, nil
}

func (x *xlsxCell) SetProperty(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "error":
		return x.setValue(v, vm)
	case "style":
		s, ok := v.ToObjectOrNil().(*xlsxStyle)
		if !ok {
			return ErrInvalidType
		}
		x.obj.SetStyle(s.obj)
		return nil
	}

	return ErrReadOnlyOrUndefined
}

func (x *xlsxCell) setValue(v dune.Value, vm *dune.VM) error {
	switch v.Type {
	case dune.Int:
		x.obj.SetInt64(v.ToInt())
	case dune.Float:
		x.obj.SetFloat(v.ToFloat())
	case dune.Bool:
		x.obj.SetInt64(v.ToInt())
	case dune.String, dune.Rune, dune.Bytes:
		x.obj.SetString(v.ToString())
	case dune.Object:
		switch t := v.ToObject().(type) {
		case TimeObj:
			loc := GetLocation(vm)
			x.obj.SetDateWithOptions(time.Time(t), xlsx.DateTimeOptions{
				Location:        loc,
				ExcelTimeFormat: xlsx.DefaultDateTimeFormat,
			})
		}
	case dune.Null, dune.Undefined:
	default:
		return fmt.Errorf("invalid cell value: %s", v.TypeName())
	}

	return nil
}

type xlsxCol struct {
	obj *xlsx.Col
}

func (x *xlsxCol) Type() string {
	return "xlsx.Col"
}

func (x *xlsxCol) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "width":
		return dune.NewFloat(x.obj.Width), nil
	}
	return dune.UndefinedValue, nil
}

func (x *xlsxCol) SetProperty(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "width":
		switch v.Type {
		case dune.Float:
		case dune.Int:
			x.obj.Width = v.ToFloat()
			return nil
		default:
			return ErrInvalidType
		}
	}

	return ErrReadOnlyOrUndefined
}

type xlsxStyle struct {
	obj *xlsx.Style
}

func (x *xlsxStyle) Type() string {
	return "xlsx.Style"
}

func (x *xlsxStyle) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "alignment":
		return dune.NewObject(&xlsxAlignment{&x.obj.Alignment}), nil
	case "applyAlignment":
		return dune.NewBool(x.obj.ApplyAlignment), nil
	case "applyFont":
		return dune.NewBool(x.obj.ApplyFont), nil
	case "font":
		return dune.NewObject(&xlsxFont{&x.obj.Font}), nil
	}
	return dune.UndefinedValue, nil
}

func (x *xlsxStyle) SetProperty(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "alignment":
		a, ok := v.ToObjectOrNil().(*xlsxAlignment)
		if !ok {
			return ErrInvalidType
		}
		x.obj.Alignment = *a.obj
		return nil
	case "applyAlignment":
		if v.Type != dune.Bool {
			return ErrInvalidType
		}
		x.obj.ApplyAlignment = v.ToBool()
		return nil
	case "font":
		a, ok := v.ToObjectOrNil().(*xlsxFont)
		if !ok {
			return ErrInvalidType
		}
		x.obj.Font = *a.obj
		return nil
	case "applyFont":
		if v.Type != dune.Bool {
			return ErrInvalidType
		}
		x.obj.ApplyFont = v.ToBool()
		return nil
	}

	return ErrReadOnlyOrUndefined
}

type xlsxAlignment struct {
	obj *xlsx.Alignment
}

func (x *xlsxAlignment) Type() string {
	return "xlsx.Alignment"
}

func (x *xlsxAlignment) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "horizontal":
		return dune.NewString(x.obj.Horizontal), nil
	case "vertical":
		return dune.NewString(x.obj.Vertical), nil
	}
	return dune.UndefinedValue, nil
}

func (x *xlsxAlignment) SetProperty(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "horizontal":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		x.obj.Horizontal = v.ToString()
		return nil
	case "vertical":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		x.obj.Vertical = v.ToString()
		return nil
	}

	return ErrReadOnlyOrUndefined
}

type xlsxFont struct {
	obj *xlsx.Font
}

func (x *xlsxFont) Type() string {
	return "xlsx.Font"
}

func (x *xlsxFont) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "bold":
		return dune.NewBool(x.obj.Bold), nil
	case "size":
		return dune.NewInt(x.obj.Size), nil
	}
	return dune.UndefinedValue, nil
}

func (x *xlsxFont) SetProperty(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "bold":
		if v.Type != dune.Bool {
			return ErrInvalidType
		}
		x.obj.Bold = v.ToBool()
		return nil
	case "size":
		if v.Type != dune.Int {
			return ErrInvalidType
		}
		x.obj.Size = int(v.ToInt())
		return nil
	}

	return ErrReadOnlyOrUndefined
}
