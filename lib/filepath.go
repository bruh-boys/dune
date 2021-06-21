package lib

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(FilePath, `

declare namespace filepath {
    /**
     * Clean returns the shortest path name equivalent to path
     */
	export function clean(path: string): string 
    /**
     * Returns the extension of a path
     */
	export function ext(path: string): string 
	
	export function abs(path: string): string

    /**
     *  Base returns the last element of path.
     *  Trailing path separators are removed before extracting the last element.
     *  If the path is empty, Base returns ".".
     *  If the path consists entirely of separators, Base returns a single separator.
     */
    export function base(path: string): string

    /**
     * Returns name of the file without the directory and extension.
     */
    export function nameWithoutExt(path: string): string

    /**
     * Returns directory part of the path.
     */
    export function dir(path: string): string

    export function join(...parts: string[]): string

    /**
     * joins the elemeents but respecting absolute paths.
     */
    export function joinAbs(...parts: string[]): string
}

`)
}

var FilePath = []dune.NativeFunction{
	{
		Name:      "filepath.clean",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			v := filepath.Clean(args[0].String())
			return dune.NewString(v), nil
		},
	},
	{
		Name:      "filepath.join",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			parts := make([]string, len(args))
			for i, v := range args {
				if v.Type != dune.String {
					return dune.NullValue, fmt.Errorf("argument %d is not a string (%s)", i, v.TypeName())
				}
				parts[i] = v.String()
			}

			path := filepath.Join(parts...)
			return dune.NewString(path), nil
		},
	},
	{
		Name:      "filepath.joinAbs",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			parts := make([]string, 0, len(args))
			for i, v := range args {
				if v.Type != dune.String {
					return dune.NullValue, fmt.Errorf("argument %d is not a string (%s)", i, v.TypeName())
				}
				s := v.String()
				if strings.HasPrefix(s, "/") {
					parts = nil
				}
				parts = append(parts, s)
			}

			path := filepath.Join(parts...)
			return dune.NewString(path), nil
		},
	},
	{
		Name:      "filepath.abs",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, ErrNoFileSystem
			}

			path, err := fs.Abs(args[0].String())
			if err != nil {
				return dune.NullValue, fmt.Errorf("abs %s: %w", args[0].String(), err)
			}

			return dune.NewString(path), nil
		},
	},
	{
		Name:      "filepath.ext",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			path := args[0].String()
			ext := filepath.Ext(path)
			return dune.NewString(ext), nil
		},
	},
	{
		Name:      "filepath.base",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			path := args[0].String()
			name := filepath.Base(path)
			return dune.NewString(name), nil
		},
	},
	{
		Name:      "filepath.nameWithoutExt",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			path := args[0].String()
			name := filepath.Base(path)
			if i := strings.LastIndexByte(name, '.'); i != -1 {
				name = name[:i]
			}
			return dune.NewString(name), nil
		},
	},
	{
		Name:      "filepath.dir",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			path := args[0].String()
			name := filepath.Dir(path)
			return dune.NewString(name), nil
		},
	},
}
