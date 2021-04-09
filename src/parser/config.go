package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dunelang/dune/filesystem"
)

type Config struct {
	BasePath string
	Paths    []string
}

func ReadConfig(fs filesystem.FS, file string) (*Config, error) {
	abs, err := fs.Abs(filepath.Dir(file))
	if err != nil {
		return nil, err
	}

	const name = "tsconfig.json"

	dir := abs
	for {
		path := filepath.Join(dir, name)

		conf, err := parseConfig(fs, path)
		if err == nil {
			return conf, nil
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error reading %s: %w", path, err)
		}
		base := filepath.Dir(dir)
		if base == dir {
			// if no tsconfig.json is found then the main directory is the base path
			return &Config{BasePath: abs, Paths: []string{"*"}}, nil
		}
		dir = base
	}
}

func parseConfig(fs filesystem.FS, path string) (*Config, error) {
	b, err := filesystem.ReadAll(fs, path)
	if err != nil {
		return nil, err
	}

	m, err := unmarshalWithComments(b)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling %s: %w", path, err)
	}

	c := &Config{
		BasePath: filepath.Dir(path),
	}

	opt, ok := m["compilerOptions"].(map[string]interface{})
	if ok {
		pathsRoot, ok := opt["paths"].(map[string]interface{})
		if ok {
			if len(pathsRoot) != 1 {
				return nil, fmt.Errorf("invalid paths. Can only contain one root element (*)")
			}

			paths, ok := pathsRoot["*"].([]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid paths *. Must be an array of dirs")
			}

			c.Paths = make([]string, len(paths))

			for i, v := range paths {
				pt, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("invalid path in tsconfig paths: %s", v)
				}
				c.Paths[i] = pt
			}
		}
	}

	if len(c.Paths) == 0 {
		c.Paths = []string{"*"}
	}

	return c, nil
}
