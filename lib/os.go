package lib

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(OS, `

declare namespace os {
	export const ErrNotExist: string

    export const stdin: io.File
    export const stdout: io.File
    export const stderr: io.File
    export const fileSystem: io.FileSystem

    export function readLine(): string

    export function exec(name: string, ...params: string[]): string

    /**
     * Reads an environment variable.
     */
    export function getEnv(key: string): string
    /**
     * Sets an environment variable.
     */
    export function setEnv(key: string, value: string): void

    export function exit(code?: number): void

    export const userHomeDir: string
	export const pathSeparator: string
	
    export function hostName(): string
	 
    export function mapPath(path: string): string

    export function newCommand(name: string, ...params: any[]): Command

    export interface Command {
        dir: string
        env: string[]
        stdin: io.File
        stdout: io.File
        stderr: io.File

        run(): void
        start(): void
        output(): string
        combinedOutput(): string
	}
	
	export function getWd(): string
	export function open(path: string): io.File
	export function openIfExists(path: string): io.File
	export function openForWrite(path: string): io.File
	export function openForAppend(path: string): io.File
	export function chdir(dir: string): void
	export function exists(path: string): boolean
	export function rename(source: string, dest: string): void
	export function removeAll(path: string): void
	export function readAll(path: string): byte[]
	export function readAllIfExists(path: string): byte[]
	export function readString(path: string): string
	export function readStringIfExists(path: string): string
	export function write(path: string, data: string | io.Reader | byte[]): void
	export function append(path: string, data: string | byte[]): void
	export function mkdir(path: string): void
	export function stat(path: string): io.FileInfo
	export function readDir(path: string): io.FileInfo[]
	export function readNames(path: string, recursive?: boolean): string[]


}


`)
}

var OS = []dune.NativeFunction{
	{
		Name: "->os.ErrNotExist",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewString(os.ErrNotExist.Error()), nil
		},
	},
	{
		Name: "os.hostName",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			name, err := os.Hostname()
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewString(name), nil
		},
	},
	{
		Name: "->os.pathSeparator",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewString(string(os.PathSeparator)), nil
		},
	},
	{
		Name:        "->os.userHomeDir",
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			d, err := os.UserHomeDir()
			if err != nil {
				return dune.NullValue, ErrUnauthorized
			}
			return dune.NewString(d), nil
		},
	},
	{
		Name:        "->os.stdout",
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			f := &file{f: os.Stdout}
			return dune.NewObject(f), nil
		},
	},
	{
		Name:        "->os.stdin",
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			f := &file{f: os.Stdin}
			return dune.NewObject(f), nil
		},
	},
	{
		Name:        "->os.stderr",
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			f := &file{f: os.Stderr}
			return dune.NewObject(f), nil
		},
	},
	{
		Name:        "os.mapPath",
		Arguments:   1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			p := args[0].String()

			if len(p) > 0 && p[0] == '~' {
				usr, err := user.Current()
				if err != nil {
					return dune.NullValue, err
				}
				p = filepath.Join(usr.HomeDir, p[1:])
			}

			return dune.NewString(p), nil
		},
	},
	{
		Name:        "os.exit",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}

			var exitCode int
			if len(args) > 0 {
				exitCode = int(args[0].ToInt())
			}

			os.Exit(exitCode)
			return dune.NullValue, nil
		},
	},
	{
		Name:        "os.exec",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l == 0 {
				return dune.NullValue, fmt.Errorf("expected at least 1 argument")
			}

			values := make([]string, l)
			for i, v := range args {
				values[i] = v.String()
			}

			cmd := exec.Command(values[0], values[1:]...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			if err := cmd.Run(); err != nil {
				return dune.NullValue, err
			}

			return dune.NullValue, nil
		},
	},
	{
		Name:        "os.newCommand",
		Arguments:   -1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l == 0 {
				return dune.NullValue, fmt.Errorf("expected at least 1 argument")
			}

			values := make([]string, l)
			for i, v := range args {
				values[i] = v.String()
			}

			cmd := newCommand(values[0], values[1:]...)

			return dune.NewObject(cmd), nil
		},
	},

	{
		Name:      "os.getWd",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}

			f := &FileSystemObj{fs}
			return f.getWd(args, vm)
		},
	},
	{
		Name:      "os.open",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.open(args, vm)
		},
	},
	{
		Name:      "os.openIfExists",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.openIfExists(args, vm)
		},
	},
	{
		Name:      "os.openForWrite",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.openForWrite(args, vm)
		},
	},
	{
		Name:      "os.openForAppend",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.openForAppend(args, vm)
		},
	},
	{
		Name:      "os.chdir",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.chdir(args, vm)
		},
	},
	{
		Name:      "os.exists",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.exists(args, vm)
		},
	},
	{
		Name:      "os.rename",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.rename(args, vm)
		},
	},
	{
		Name:      "os.removeAll",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.removeAll(args, vm)
		},
	},
	{
		Name:      "os.readAll",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.readAll(args, vm)
		},
	},
	{
		Name:      "os.readAllIfExists",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.readAllIfExists(args, vm)
		},
	},
	{
		Name:      "os.readString",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.readString(args, vm)
		},
	},
	{
		Name:      "os.readStringIfExists",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.readStringIfExists(args, vm)
		},
	},
	{
		Name:      "os.write",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.write(args, vm)
		},
	},
	{
		Name:      "os.append",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.append(args, vm)
		},
	},
	{
		Name:      "os.mkdir",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.mkdir(args, vm)
		},
	},
	{
		Name:      "os.stat",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.stat(args, vm)
		},
	},
	{
		Name:      "os.readDir",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.readDir(args, vm)
		},
	},
	{
		Name:      "os.readNames",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}
			f := &FileSystemObj{fs}
			return f.readNames(args, vm)
		},
	},

	{
		Name:        "os.readLine",
		Arguments:   0,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {

			r := bufio.NewReader(os.Stdin)
			s, err := r.ReadString('\n')
			if err != nil {
				return dune.NullValue, err
			}

			// trim the \n
			s = s[:len(s)-1]

			return dune.NewString(s), nil
		},
	},
	{
		Name: "->os.fileSystem",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if vm.FileSystem == nil {
				return dune.NullValue, nil
			}
			return dune.NewObject(NewFileSystem(vm.FileSystem)), nil
		},
	},
	{
		Name:        "os.getEnv",
		Arguments:   1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			s := os.Getenv(args[0].String())
			return dune.NewString(s), nil
		},
	},
	{
		Name:        "os.setEnv",
		Arguments:   2,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			if err := os.Setenv(args[0].String(), args[1].String()); err != nil {
				return dune.NullValue, err
			}
			return dune.NullValue, nil
		},
	},
}

func newCommand(name string, arg ...string) *command {
	return &command{
		command: exec.Command(name, arg...),
	}
}

type command struct {
	command *exec.Cmd
}

func (*command) Type() string {
	return "os.command"
}

func (c *command) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "run":
		return c.run
	case "start":
		return c.start
	case "output":
		return c.output
	case "combinedOutput":
		return c.combinedOutput
	}
	return nil
}

func (c *command) GetProperty(key string, vm *dune.VM) (dune.Value, error) {
	switch key {
	case "stdin":
		return dune.NewObject(c.command.Stdin), nil
	case "stdout":
		return dune.NewObject(c.command.Stdout), nil
	case "stderr":
		return dune.NewObject(c.command.Stderr), nil
	case "dir":
		return dune.NewString(c.command.Dir), nil
	case "env":
		items := c.command.Env
		a := make([]dune.Value, len(items))
		for i, v := range items {
			a[i] = dune.NewString(v)
		}
		return dune.NewArrayValues(a), nil
	}
	return dune.UndefinedValue, nil
}

func (c *command) SetProperty(key string, v dune.Value, vm *dune.VM) error {
	switch key {
	case "stdin":
		if v.Type != dune.Object {
			return ErrInvalidType
		}
		o, ok := v.ToObject().(io.Reader)
		if !ok {
			return fmt.Errorf("expected a Reader")
		}
		c.command.Stdin = o
		return nil

	case "stdout":
		if v.Type != dune.Object {
			return ErrInvalidType
		}
		o, ok := v.ToObject().(io.Writer)
		if !ok {
			return fmt.Errorf("expected a Writer")
		}
		c.command.Stdout = o
		return nil

	case "stderr":
		if v.Type != dune.Object {
			return ErrInvalidType
		}
		o, ok := v.ToObject().(io.Writer)
		if !ok {
			return fmt.Errorf("expected a Writer")
		}
		c.command.Stderr = o
		return nil

	case "dir":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		c.command.Dir = v.String()
		return nil

	case "env":
		if v.Type != dune.Array {
			return ErrInvalidType
		}
		a := v.ToArray()
		b := make([]string, len(a))
		for i, v := range a {
			b[i] = v.String()
		}
		c.command.Env = b
		return nil
	}
	return ErrReadOnlyOrUndefined
}

func (c *command) run(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	err := c.command.Run()
	return dune.NullValue, err
}

func (c *command) start(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	err := c.command.Start()
	return dune.NullValue, err
}

func (c *command) output(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	b, err := c.command.Output()
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewString(string(b)), nil
}

func (c *command) combinedOutput(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	b, err := c.command.CombinedOutput()
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewString(string(b)), nil
}
