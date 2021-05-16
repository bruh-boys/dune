package lib

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/dunelang/dune"
	"github.com/dunelang/dune/ast"
	"github.com/dunelang/dune/binary"
	"github.com/dunelang/dune/filesystem"
	"github.com/dunelang/dune/parser"
)

func init() {
	dune.RegisterLib(Bytecode, `

declare namespace ast {
	export interface Program {
		string(): string
	}
}
	
declare namespace bytecode {
    /**
     * 
     * @param path 
     * @param fileSystem 
     * @param scriptMode if statements outside of functions are allowed.
     */
	export function compile(path: string, fileSystem?: io.FileSystem): runtime.Program
	
	export function hash(path: string, fileSystem?: io.FileSystem): string

    export function compileStr(code: string): runtime.Program

    export function parseStr(code: string): ast.Program

    /**
     * Load a binary program from the file system
     * @param path the path to the main binary.
     * @param fs the file trusted. If empty it will use the current fs.
     */
    export function load(path: string, fs?: io.FileSystem): runtime.Program

    export function loadProgram(b: byte[]): runtime.Program
    export function readProgram(r: io.Reader): runtime.Program
    export function writeProgram(w: io.Writer, p: runtime.Program): void
}

`)
}

var Bytecode = []dune.NativeFunction{
	{
		Name:        "bytecode.compile",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return compile(args, vm)
		},
	},
	{
		Name:        "bytecode.hash",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String, dune.Bool, dune.Bool, dune.Object); err != nil {
				return dune.NullValue, err
			}

			path := args[0].String()

			var fs filesystem.FS

			l := len(args)

			if l > 1 {
				filesystem, ok := args[1].ToObjectOrNil().(*FileSystemObj)
				if !ok {
					return dune.NullValue, fmt.Errorf("expected a filesystem, got %v", args[1])
				}
				fs = filesystem.FS
			} else {
				fs = vm.FileSystem
			}

			hash, err := parser.Hash(fs, path)
			if err != nil {
				return dune.NullValue, fmt.Errorf("compiling %s: %w", path, err)
			}

			s := base64.StdEncoding.EncodeToString(hash)
			return dune.NewString(s), nil
		},
	},
	{
		Name:        "bytecode.compileStr",
		Arguments:   1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			code := args[0].String()

			p, err := dune.CompileStr(code)
			if err != nil {
				return dune.NullValue, errors.New(err.Error())
			}

			return dune.NewObject(&program{prog: p}), nil
		},
	},
	{
		Name:        "bytecode.parseStr",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			code := args[0].String()

			p, err := dune.ParseStr(code)
			if err != nil {
				return dune.NullValue, errors.New(err.Error())
			}

			return dune.NewObject(&astProgram{prog: p}), nil
		},
	},
	{
		Name:        "bytecode.loadProgram",
		Arguments:   1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			p, err := binary.Load(args[0].ToBytes())
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(&program{prog: p}), nil
		},
	},
	{
		Name:        "bytecode.load",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String, dune.Object); err != nil {
				return dune.NullValue, err
			}

			l := len(args)
			path := args[0].String()
			var fs filesystem.FS

			if l > 1 {
				filesystem, ok := args[1].ToObjectOrNil().(*FileSystemObj)
				if !ok {
					return dune.NullValue, fmt.Errorf("expected a filesystem, got %v", args[1])
				}
				fs = filesystem.FS
			} else {
				fs = vm.FileSystem
			}

			f, err := fs.Open(path)
			if err != nil {
				return dune.NullValue, err
			}
			defer f.Close()

			p, err := binary.Read(f)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(&program{prog: p}), nil
		},
	},
	{
		Name:        "bytecode.readProgram",
		Arguments:   1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			r, ok := args[0].ToObjectOrNil().(io.Reader)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be io.Reader, got %T", args[0].ToObjectOrNil())
			}

			p, err := binary.Read(r)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(&program{prog: p}), nil
		},
	},
	{
		Name:        "bytecode.writeProgram",
		Arguments:   2,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object, dune.Object); err != nil {
				return dune.NullValue, err
			}

			w, ok := args[0].ToObjectOrNil().(io.Writer)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected parameter 1 to be io.Reader, got %T", args[0].ToObjectOrNil())
			}

			p, ok := args[1].ToObjectOrNil().(*program)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected parameter 2 to be a program, got %T", args[0].ToObjectOrNil())
			}

			if err := binary.Write(w, p.prog); err != nil {
				return dune.NullValue, err
			}

			return dune.NullValue, nil
		},
	},
}

type astProgram struct {
	prog *ast.Module
}

func (p *astProgram) Type() string {
	return "ast.Program"
}

func (p *astProgram) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "string":
		return p.string
	}
	return nil
}

func (p *astProgram) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	s, err := ast.Sprint(p.prog)
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewString(s), nil
}

func compile(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.String, dune.Object); err != nil {
		return dune.NullValue, err
	}

	path := args[0].String()

	var fs filesystem.FS

	l := len(args)

	if l > 1 {
		filesystem, ok := args[1].ToObjectOrNil().(*FileSystemObj)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected a filesystem, got %v", args[1])
		}
		fs = filesystem.FS
	} else {
		fs = vm.FileSystem
	}

	p, err := dune.Compile(fs, path)
	if err != nil {
		return dune.NullValue, fmt.Errorf("compiling %s: %w", path, err)
	}

	return dune.NewObject(&program{prog: p}), nil
}
