package lib

import (
	"fmt"
	"io"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(FileUtil, `

declare namespace fileutil { 
    export function copy(src: string, dst: string): byte[]

	export function isDirEmpty(path: string, fs?: io.FileSystem): boolean
}

`)
}

var FileUtil = []dune.NativeFunction{
	{
		Name:      "fileutil.isDirEmpty",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String, dune.Object); err != nil {
				return dune.NullValue, err
			}
			if err := ValidateArgRange(args, 1, 2); err != nil {
				return dune.NullValue, err
			}

			fs, ok := args[1].ToObject().(*FileSystemObj)
			if !ok {
				return dune.NullValue, fmt.Errorf("invalid filesystem argument, got %v", args[1])
			}

			name := args[0].String()

			f, err := fs.FS.Open(name)
			if err != nil {
				return dune.NullValue, err
			}
			defer f.Close()

			_, err = f.Readdir(1)
			if err != nil {
				if err == io.EOF {
					return dune.NewBool(true), nil
				}
				return dune.NullValue, err
			}

			return dune.NewBool(false), nil
		},
	},
	{
		Name:      "fileutil.copy",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			src := args[0].String()
			dst := args[1].String()

			fs := vm.FileSystem
			if fs == nil {
				return dune.NullValue, fmt.Errorf("no filesystem")
			}

			r, err := fs.Open(src)
			if err != nil {
				return dune.NullValue, err
			}

			defer r.Close()

			w, err := fs.OpenForWrite(dst)
			if err != nil {
				return dune.NullValue, err
			}

			defer w.Close()

			if _, err := io.Copy(w, r); err != nil {
				return dune.NullValue, err
			}

			return dune.NullValue, nil
		},
	},
}
