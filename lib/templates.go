package lib

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dunelang/dune"

	"github.com/dunelang/dune/filesystem"
	"github.com/dunelang/dune/lib/templates"
)

func init() {
	dune.RegisterLib(Templates, `


declare namespace templates {
    /**
     * Reads the file and processes includes
     */
    export function exec(code: string, model?: any): string
    export function preprocess(path: string, fs?: io.FileSystem): string
    export function render(text: string, model?: any): string
    export function renderHTML(text: string, model?: any): string
    /**
     * 
     * @param headerFunc By defauult is: function render(w: io.Writer, model: any)
     */
    export function compile(text: string): string
    /**
     * 
     * @param headerFunc By defauult is: function render(w: io.Writer, model: any)
     */
    export function compileHTML(text: string): string

    /**
     * 
     * @param headerFunc By defauult is: function render(w: io.Writer, model: any)
     */
    export function writeHTML(w: io.Writer, path: string, model?: any, fs?: io.FileSystem): void

    /**
     * 
     * @param headerFunc By defauult is: function render(w: io.Writer, model: any)
     */
    export function writeHTMLTemplate(w: io.Writer, template: string, model?: any): void
}

`)
}

var includesRegex = regexp.MustCompile(`<!-- include "(.*?)" -->`)

var Templates = []dune.NativeFunction{
	{
		Name:      "templates.exec",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			var buf []byte
			var model dune.Value

			l := len(args)
			if l == 0 || l > 2 {
				return dune.NullValue, fmt.Errorf("expected one or two arguments, got %d", l)
			}

			a := args[0]
			switch a.Type {
			case dune.String, dune.Bytes:
				buf = a.ToBytes()
			default:
				return dune.NullValue, ErrInvalidType
			}

			if l == 2 {
				model = args[1]
			}

			return execTemplate(buf, nil, model, vm)
		},
	},
	{
		Name:      "templates.render",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return render(false, this, args, vm)
		},
	},
	{
		Name:      "templates.renderHTML",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return render(true, this, args, vm)
		},
	},
	{
		Name:      "templates.writeHTML",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return writeHTML(true, this, args, vm)
		},
	},
	{
		Name:      "templates.writeHTMLTemplate",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return writeHTMLTemplate(true, this, args, vm)
		},
	},
	{
		Name:      "templates.compile",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return compileTemplate(false, this, args, vm)
		},
	},
	{
		Name:      "templates.compileHTML",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return compileTemplate(true, this, args, vm)
		},
	},
	{
		Name:      "templates.preprocess",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			var path string
			var fs filesystem.FS

			l := len(args)
			switch l {
			case 1:
				path = args[0].String()
				fs = vm.FileSystem
			case 2:
				path = args[0].String()
				fo, ok := args[1].ToObjectOrNil().(*FileSystemObj)
				if !ok {
					return dune.NullValue, fmt.Errorf("expected a fileSystem, got %s", args[0].TypeName())
				}
				fs = fo.FS
			default:
				return dune.NullValue, fmt.Errorf("expected one or two arguments, got %d", l)
			}

			buf, err := readFile(path, fs, vm)
			if err != nil {
				if os.IsNotExist(err) {
					return dune.NullValue, nil
				}
				return dune.NullValue, fmt.Errorf("error reading template '%s':_ %v", path, err)
			}

			includes := includesRegex.FindAllSubmatchIndex(buf, -1)
			for i := len(includes) - 1; i >= 0; i-- {
				loc := includes[i]
				start := loc[0]
				end := loc[1]
				include := string(buf[loc[2]:loc[3]])
				b, err := readFile(include, fs, vm)
				if err != nil {
					if os.IsNotExist(err) {
						// try the path relative to the template dir
						localPath := filepath.Join(filepath.Dir(path), include)
						b, err = readFile(localPath, fs, vm)
						if err != nil {
							return dune.NullValue, fmt.Errorf("error reading include '%s':_ %v", include, err)
						}
					} else {
						return dune.NullValue, fmt.Errorf("error reading include '%s':_ %v", include, err)
					}
				}
				buf = append(buf[:start], append(b, buf[end:]...)...)
			}

			return dune.NewString(string(buf)), nil
		},
	},
}

func readFile(path string, fs filesystem.FS, vm *dune.VM) ([]byte, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}

	b, err := ReadAll(f, vm)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func compileTemplate(html bool, this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
	var code string

	l := len(args)
	switch l {
	case 1:
		code = args[0].String()
	default:
		return dune.NullValue, fmt.Errorf("expected one or two arguments, got %d", l)
	}

	var b []byte
	var sourceMap []int
	var err error

	if html {
		b, sourceMap, err = templates.CompileHtml(code)
	} else {
		b, sourceMap, err = templates.Compile(code)
	}

	if err != nil {
		return dune.NullValue, mapError(err, sourceMap)
	}

	return dune.NewString(string(b)), nil
}

func writeHTML(html bool, this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
	ln := len(args)

	if ln < 2 || ln > 4 {
		return dune.NullValue, fmt.Errorf("expected at 2, 3 or 4 arguments, got %d", ln)
	}

	if _, ok := args[0].ToObjectOrNil().(io.Writer); !ok {
		return dune.NullValue, fmt.Errorf("expected arg 1 to be a io.Writer, got %s", args[0].TypeName())
	}

	vPath := args[1]
	if vPath.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected arg 2 to be a string, got %s", vPath.TypeName())
	}

	var model dune.Value
	if ln > 2 {
		model = args[2]
	}

	var fs filesystem.FS
	if ln == 4 {
		vFS, ok := args[3].ToObjectOrNil().(*FileSystemObj)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected arg 4 to be a io.FileSystem, got %s", args[3].TypeName())
		}
		fs = vFS.FS
	} else {
		fs = vm.FileSystem
	}

	if fs == nil {
		return dune.NullValue, fmt.Errorf("there is no filesystem")
	}

	src, err := filesystem.ReadAll(fs, vPath.String())
	if err != nil {
		return dune.NullValue, err
	}

	var sourceMap []int
	var b []byte
	if html {
		b, sourceMap, err = templates.CompileHtml(string(src))
	} else {
		b, sourceMap, err = templates.Compile(string(src))
	}

	if err != nil {
		return dune.NullValue, mapError(err, sourceMap)
	}

	return write(args[0], b, sourceMap, model, vm)
}

func writeHTMLTemplate(html bool, this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
	ln := len(args)

	if ln < 2 || ln > 4 {
		return dune.NullValue, fmt.Errorf("expected at 2, 3 or 4 arguments, got %d", ln)
	}

	if _, ok := args[0].ToObjectOrNil().(io.Writer); !ok {
		return dune.NullValue, fmt.Errorf("expected arg 1 to be a io.Writer, got %s", args[0].TypeName())
	}

	template := args[1]
	if template.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected arg 2 to be a string, got %s", template.TypeName())
	}

	var model dune.Value
	if ln > 2 {
		model = args[2]
	}

	var sourceMap []int
	var b []byte
	var err error
	if html {
		b, sourceMap, err = templates.CompileHtml(template.String())
	} else {
		b, sourceMap, err = templates.Compile(template.String())
	}

	if err != nil {
		return dune.NullValue, mapError(err, sourceMap)
	}

	return write(args[0], b, sourceMap, model, vm)
}

func render(html bool, this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
	var template dune.Value
	var model dune.Value

	l := len(args)
	switch l {
	case 0:
		return dune.NullValue, fmt.Errorf("expected at least one argument, got %d", l)
	case 1:
		template = args[0]
		switch template.Type {
		case dune.String, dune.Bytes:
		default:
			return dune.NullValue, ErrInvalidType
		}
	case 2:
		template = args[0]
		switch template.Type {
		case dune.String, dune.Bytes:
		default:
			return dune.NullValue, ErrInvalidType
		}
		model = args[1]
	default:
		return dune.NullValue, fmt.Errorf("expected one or two arguments, got %d", l)
	}

	var b []byte
	var sourceMap []int
	var err error

	if html {
		b, sourceMap, err = templates.CompileHtml(template.String())
	} else {
		b, sourceMap, err = templates.Compile(template.String())
	}

	if err != nil {
		return dune.NullValue, errors.New(err.Error())
	}

	return execTemplate(b, sourceMap, model, vm)
}

func getVM(b []byte, vm *dune.VM) (*dune.VM, error) {
	p, err := dune.CompileStr(string(b))
	if err != nil {
		return nil, err
	}

	m := dune.NewVM(p)

	m.MaxAllocations = 10000000000 // vm.MaxAllocations
	m.MaxFrames = 10               // vm.MaxFrames
	m.MaxSteps = 10000000          // vm.MaxSteps

	return m, nil
}

func write(w dune.Value, b []byte, sourceMap []int, model dune.Value, vm *dune.VM) (dune.Value, error) {
	m, err := getVM(b, vm)
	if err != nil {
		return dune.NullValue, mapError(err, sourceMap)
	}

	if _, err := m.RunFunc("render", w, model); err != nil {
		// return the error with the stacktrace included in the message
		// because the caller in the program will have it's own stacktrace.
		return dune.NullValue, mapError(err, sourceMap)
	}

	if err := vm.AddSteps(m.Steps()); err != nil {
		return dune.NullValue, mapError(err, sourceMap)
	}

	return dune.NullValue, nil
}

func execTemplate(b []byte, sourceMap []int, model dune.Value, vm *dune.VM) (dune.Value, error) {
	m, err := getVM(b, vm)
	if err != nil {
		return dune.NullValue, mapError(err, sourceMap)
	}

	buf := &buffer{buf: &bytes.Buffer{}}

	if _, err := m.RunFunc("render", dune.NewObject(buf), model); err != nil {
		// return the error with the stacktrace included in the message
		// because the caller in the program will have it's own stacktrace.
		return dune.NullValue, mapError(err, sourceMap)
	}

	if err := vm.AddSteps(m.Steps()); err != nil {
		return dune.NullValue, mapError(err, sourceMap)
	}

	return dune.NewString(buf.buf.String()), nil
}

func mapError(e error, sourceMap []int) error {
	ln := len(sourceMap)

	if ln == 0 {
		return e
	}

	var lines []string
	r := strings.NewReader(e.Error())
	s := bufio.NewScanner(r)

	for s.Scan() {
		line := s.Text()
		if !strings.HasPrefix(line, " -> ") {
			lines = append(lines, line)
			continue
		}

		i := strings.LastIndexByte(line, ':')
		if i == -1 {
			i = strings.LastIndex(line, " line ")
			if i == -1 {
				lines = append(lines, line)
				continue
			}
			i += 5
		}

		n, err := strconv.Atoi(line[i+1:])
		if err != nil {
			return fmt.Errorf("error mapping error: %v. Original Errror: %w", err, e)
		}

		// error lines are reported in base 1 and the map is also in base 0
		n -= 2

		if n > ln {
			return fmt.Errorf("error mapping error: %v. Original Errror: %w", err, e)
		}

		// now show the mapped line in base 1 again
		m := sourceMap[n] + 1

		line = line[:i] + ":" + strconv.Itoa(m)
		lines = append(lines, line)
	}

	if err := s.Err(); err != nil {
		return fmt.Errorf("error mapping error: %v. Original Errror: %w", err, e)
	}

	sErr := strings.Join(lines, "\n")

	return fmt.Errorf("template error: %s", sErr)
}

type buffer struct {
	buf *bytes.Buffer
}

func (b buffer) Type() string {
	return "templates.Buffer"
}

func (b buffer) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "write":
		return b.write
	}
	return nil
}

func (b buffer) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	v := args[0]

	switch v.Type {
	case dune.Null, dune.Undefined:
		return dune.NullValue, nil
	case dune.String:
		b.buf.WriteString(v.String())
	case dune.Bytes:
		b.buf.Write(v.ToBytes())
	case dune.Int:
		fmt.Fprintf(b.buf, "%d", v.ToInt())
	case dune.Float:
		fmt.Fprintf(b.buf, "%f", v.ToFloat())
	case dune.Array:
		b.buf.WriteString("[array]")
	case dune.Object:
		b.buf.WriteString("[object]")
	case dune.Map:
		b.buf.WriteString("[object]")
	default:
		return dune.NullValue, ErrInvalidType
	}

	return dune.NullValue, nil
}
