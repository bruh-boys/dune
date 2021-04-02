package parser

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dunelang/dune/ast"
	"github.com/dunelang/dune/filesystem"
)

var Optimizations = true

func Parse(fs filesystem.FS, path string) (*ast.Module, error) {
	p := newParser(fs)
	return p.Parse(path)
}

func ParseStr(code string) (*ast.Module, error) {
	r := strings.NewReader(code)
	l := ast.New(r, "")
	if err := l.Run(); err != nil {
		return nil, err
	}

	p := newParser(nil)
	p.tokens = l.Tokens
	p.index = 0

	file, err := p.parse()
	if err != nil {
		return nil, err
	}

	a := newAST()
	a.File = file
	a.File.Global = p.global
	p.importedPaths = make(map[string]bool)

	// ast.Print(a.File)

	return a, nil
}

func ParseExpr(code string) (ast.Expr, error) {
	r := strings.NewReader(code)
	l := ast.New(r, "")
	if err := l.Run(); err != nil {
		return nil, fmt.Errorf("lex error: %w", err)
	}

	p := newParser(nil)
	p.tokens = l.Tokens
	p.index = 0

	expr, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return expr, nil
}

func newParser(fs filesystem.FS) *parser {
	return &parser{FS: fs}
}

type parser struct {
	Config        *Config
	tokens        []*ast.Token
	index         int
	FS            filesystem.FS
	global        []ast.Stmt
	importedPaths map[string]bool
}

func (p *parser) SetFS(fs filesystem.FS) {
	p.FS = fs
}

func (p *parser) Parse(path string) (*ast.Module, error) {
	if p.FS == nil {
		return nil, fmt.Errorf("there is no filesystem.FS")
	}

	abs, err := p.FS.Abs(path)
	if err != nil {
		return nil, err
	}

	if _, err := p.FS.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			// try vendor
			abs, err = p.FS.Abs(filepath.Join("vendor", path))
			if err != nil {
				return nil, err
			}

			if _, err := p.FS.Stat(abs); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	conf, err := ReadConfig(p.FS, abs)
	if err != nil {
		return nil, err
	}

	p.Config = conf

	file, err := p.parseFile(abs)
	if err != nil {
		return nil, err
	}

	a := newAST()
	a.BasePath = p.Config.BasePath
	a.File = file
	p.importedPaths = make(map[string]bool)

	if err := p.parseImports(a, file); err != nil {
		return nil, err
	}

	a.File.Global = p.global

	return a, nil
}

func (p *parser) isTypeDefinitionFile(path string) bool {
	if p.Config == nil {
		// this means that there is no filesystem.FS so there can't be a .d.ts
		return false
	}

	path = filepath.Join(p.Config.BasePath, path)
	_, err := p.FS.Stat(path + ".d.ts")
	return !os.IsNotExist(err)
}

// ParseStatements parses code directly. Filename is used to provide error lines.
func (p *parser) ParseStatements(code, fileName string) (*ast.Module, error) {
	r := strings.NewReader(code)

	l := ast.New(r, fileName)
	if err := l.Run(); err != nil {
		return nil, err
	}

	p.tokens = l.Tokens
	p.index = 0

	file, err := p.parse()
	if err != nil {
		return nil, err
	}

	file.Path = fileName
	file.Global = p.global

	a := newAST()
	a.File = file
	p.importedPaths = make(map[string]bool)

	if err := p.parseImports(a, file); err != nil {
		return nil, err
	}

	return a, nil
}

func newAST() *ast.Module {
	return &ast.Module{
		Modules: make(map[string]*ast.File),
	}
}

func unmarshalWithComments(b []byte) (map[string]interface{}, error) {
	var lines []string
	for _, line := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(strings.TrimLeft(line, " \t"), "//") {
			lines = append(lines, line)
		}
	}
	b = []byte(strings.Join(lines, "\n"))

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (p *parser) parseImports(ast *ast.Module, file *ast.File) error {
	if len(file.Imports) == 0 {
		return nil
	}

	if p.FS == nil {
		return NewError(file.Imports[0].Pos,
			"Can't parse imports because there is no filesystem.FS")
	}

	for _, imp := range file.Imports {
		path := imp.Path
		parentPath := filepath.Dir(file.Path)

		absPath, isTypeDef, err := p.findSource(path, parentPath)
		if err != nil {
			return NewError(imp.Pos, "Import error '%s': %v", path, err)
		}

		if isTypeDef {
			continue
		}

		code, err := p.readSource(absPath)
		if err != nil {
			return NewError(imp.Pos, "ErrNotFound. Import error '%s': %v", absPath, err)
		}

		absWithoutExt := strings.TrimSuffix(absPath, ".ts")

		// set the absolute path of the import because it can be outside the basepath
		// because of tsconfig dirs
		imp.AbsPath = absWithoutExt

		// an import without alias is an import of a
		// regular source file, not a module
		notModule := imp.Alias == ""

		if notModule {
			if _, ok := p.importedPaths[absPath]; ok {
				// already imported
				continue
			}
		} else {
			if _, ok := ast.Modules[absWithoutExt]; ok {
				// already imported
				continue
			}
		}

		f, err := p.parseCode(code, absPath)
		if err != nil {
			return err
		}

		ast.Modules[absWithoutExt] = f

		// parse the imports of the imported file
		if err := p.parseImports(ast, f); err != nil {
			return err
		}
	}

	return nil
}

func (p *parser) findSource(path, parentPath string) (string, bool, error) {
	// if the path is absolute then try it directly
	if filepath.IsAbs(path) {
		file, isTypeDef := p.findSourceFile(path)
		if isTypeDef {
			return "", true, nil
		}
		if file != "" {
			return file, false, nil
		}
	}

	// try the base paths defined in the config
	fs := p.FS
	basePath := p.Config.BasePath
	for _, confPath := range p.Config.Paths {
		testPath := filepath.Join(basePath, strings.TrimSuffix(confPath, "*"), path)
		file, err := fs.Abs(testPath)
		if err != nil {
			return "", false, err
		}

		absExt, isTypeDef := p.findSourceFile(file)
		if isTypeDef {
			return "", true, nil
		}
		if absExt != "" {
			return absExt, false, nil
		}
	}

	testPath := filepath.Join(parentPath, path)
	file, err := fs.Abs(testPath)
	if err != nil {
		return "", false, err
	}

	absExt, isTypeDef := p.findSourceFile(file)
	if isTypeDef {
		return "", true, nil
	}
	if absExt != "" {
		return absExt, false, nil
	}

	// try relative to the file where is defined

	// try vendor
	absExt, isTypeDef = p.findSourceFile(filepath.Join("vendor", file))
	if isTypeDef {
		return "", true, nil
	}
	if absExt != "" {
		return absExt, false, nil
	}

	return "", false, os.ErrNotExist
}

func (p *parser) findSourceFile(absPath string) (file string, isTypeDef bool) {
	file = absPath + ".d.ts"
	if filesystem.Exists(p.FS, file) {
		return file, true
	}

	file = absPath + ".ts"
	if filesystem.Exists(p.FS, file) {
		return file, false
	}

	return "", false
}

func (p *parser) readSource(path string) (string, error) {
	r, err := p.FS.Open(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	code, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(code), nil
}

func (p *parser) parseFile(file string) (*ast.File, error) {

	r, err := p.FS.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	code, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return p.parseCode(string(code), file)
}

func (p *parser) parseCode(code, absPath string) (*ast.File, error) {
	if i := strings.Index(code, "//ts:ignore"); i != -1 {
		// discard everything below the ignore attribute if present.
		code = code[:i]
	}

	// ignore windows return
	code = strings.Replace(code, "\r", "", -1)

	r := strings.NewReader(code)
	l := ast.New(r, absPath)
	if err := l.Run(); err != nil {
		return nil, err
	}

	p.tokens = l.Tokens
	p.index = 0

	file, err := p.parse()
	if err != nil {
		return nil, err
	}

	file.Path = absPath
	return file, nil
}

func (p *parser) parse() (*ast.File, error) {
	file := &ast.File{}

	var attributes []*ast.Token
	var lastAttribute *ast.Token

loop:
	for {
		t := p.peek()
		switch t.Type {

		case ast.ATTRIBUTE:
			p.next()
			// it is only a file attribute if it is the first thing
			if len(attributes) == 0 && len(file.Imports) == 0 && len(file.Stms) == 0 {
				// even if it is at the top check if there is white space after the attributes
				if lastAttribute == nil || lastAttribute.Pos.Line == t.Pos.Line-1 {
					file.Attributes = append(file.Attributes, t.Str)
					lastAttribute = t
					continue
				}
			}
			attributes = append(attributes, t)
			lastAttribute = t

		case ast.IMPORT:
			if len(attributes) > 0 {
				return nil, NewError(attributes[0].Pos, "invalid attribute")
			}

			if len(file.Stms) > 0 {
				return nil, NewError(t.Pos, "non-declaration statement outside function body")
			}
			imp, err := p.parseImport()
			if err != nil {
				return nil, err
			}
			if imp != nil {
				file.Imports = append(file.Imports, imp)
			}

		case ast.FUNCTION:
			p.next()
			fnDec, err := p.parseFuncDeclStmt(false, t)
			if err != nil {
				return nil, err
			}

			if len(attributes) > 0 {
				for _, d := range attributes {
					fnDec.Attributes = append(fnDec.Attributes, d.Str)
				}
				attributes = nil
			} else if len(file.Stms) == 0 {
				if lastAttribute != nil && lastAttribute.Pos.Line == t.Pos.Line-1 {
					// if the attributes are associated to the first function, move them
					// from the file to the function
					fnDec.Attributes = file.Attributes
					file.Attributes = nil
					lastAttribute = nil
				}
			}

			file.Stms = append(file.Stms, fnDec)

		case ast.CLASS:
			classStmt, err := p.parseClassDeclStmt()
			if err != nil {
				return nil, err
			}
			if len(attributes) > 0 {
				for _, d := range attributes {
					classStmt.Attributes = append(classStmt.Attributes, d.Str)
				}
				attributes = nil
			} else if len(file.Stms) == 0 {
				if lastAttribute != nil && lastAttribute.Pos.Line == t.Pos.Line-1 {
					// if the attributes are associated to the first class, move them
					// from the file to the class
					classStmt.Attributes = file.Attributes
					file.Attributes = nil
					lastAttribute = nil
				}
			}

			file.Stms = append(file.Stms, classStmt)

		case ast.INTERFACE:
			if len(attributes) > 0 {
				return nil, NewError(attributes[0].Pos, "invalid attribute")
			}

			if err := p.ignoreInterface(); err != nil {
				return nil, err
			}

		case ast.EXPORT:
			// parseExportStmtOrNIL can return nil because there is no
			// equivalent statement like "export interface"
			exp, err := p.parseExportStmtOrNIL()
			if err != nil {
				return nil, err
			}
			if exp != nil {
				switch t := exp.(type) {
				case *ast.FuncDeclStmt:
					if len(attributes) > 0 {
						for _, d := range attributes {
							t.Attributes = append(t.Attributes, d.Str)
						}
						attributes = nil
					} else if len(file.Stms) == 0 {
						if lastAttribute != nil && lastAttribute.Pos.Line == t.Pos.Line-1 {
							// if the attributes are associated to the first function, move them
							// from the file to the function
							t.Attributes = file.Attributes
							file.Attributes = nil
							lastAttribute = nil
						}
					}
				case *ast.ClassDeclStmt:
					if len(attributes) > 0 {
						for _, d := range attributes {
							t.Attributes = append(t.Attributes, d.Str)
						}
						attributes = nil
					} else if len(file.Stms) == 0 {
						if lastAttribute != nil && lastAttribute.Pos.Line == t.Pos.Line-1 {
							// if the attributes are associated to the first class, move them
							// from the file to the class
							t.Attributes = file.Attributes
							file.Attributes = nil
							lastAttribute = nil
						}
					}

				}

				file.Stms = append(file.Stms, exp)
			}

		case ast.IDENT:
			if len(attributes) > 0 {
				return nil, NewError(attributes[0].Pos, "invalid attribute")
			}

			switch t.Str {
			case "type":
				// type definitions like: type a = "foo" | "bar";
				if err := p.ignoreTypeDefinition(); err != nil {
					return nil, err
				}

			case "declare":
				stmts, err := p.parseDeclareGlobal()
				if err != nil {
					return nil, err
				}
				p.global = append(p.global, stmts...)

			case "namespace":
				return nil, NewError(t.Pos, "Namespaces are not supported. Use modules instead.")

			default:
				stmt, err := p.parseStmt()
				if err != nil {
					return nil, err
				}
				file.Stms = append(file.Stms, stmt)
			}

		case ast.EOF:
			if len(attributes) > 0 {
				return nil, NewError(attributes[0].Pos, "invalid attribute")
			}

			break loop

		default:
			if len(attributes) > 0 {
				return nil, NewError(attributes[0].Pos, "invalid attribute")
			}

			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}

			file.Stms = append(file.Stms, stmt)
		}
	}

	file.Comments = p.parseComments()

	return file, nil
}

func (p *parser) parseComments() []*ast.Comment {
	var cs []*ast.Comment

	for _, t := range p.tokens {
		switch t.Type {
		case ast.COMMENT, ast.MULTILINE_COMMENT:
			cs = append(cs, &ast.Comment{
				MultiLine: t.Type == ast.MULTILINE_COMMENT,
				Str:       t.Str,
				Pos:       t.Pos,
			})
		}
	}

	return cs
}

func (p *parser) parseImport() (*ast.ImportStmt, error) {
	t, err := p.accept(ast.IMPORT)
	if err != nil {
		return nil, err
	}

	s := p.peek()
	switch s.Type {
	case ast.STRING:
		// if is a source file import: import "foo"
		p.next()
		p.ignore(ast.SEMICOLON, 1)
		if p.isTypeDefinitionFile(s.Str) {
			// ignore imports to type definition files
			return nil, nil
		}
		return &ast.ImportStmt{Pos: t.Pos, Path: s.Str}, nil

	case ast.LBRACE:
		return nil, NewError(t.Pos, "Partial imports are not supported. Use import *")
	}

	// it is a module import.
	// Only full imports are allowed: import * as foo from "x"
	if _, err := p.accept(ast.MUL); err != nil {
		return nil, err
	}

	a, err := p.acceptIdent()
	if err != nil {
		return nil, err
	}
	if a.Str != "as" {
		return nil, NewError(a.Pos, "Expected 'as'")
	}

	alias, err := p.accept(ast.IDENT)
	if err != nil {
		return nil, err
	}

	if i, err := p.accept(ast.IDENT); err != nil || i.Str != "from" {
		return nil, err
	}

	path, err := p.accept(ast.STRING)
	if err != nil {
		return nil, err
	}

	p.ignore(ast.SEMICOLON, 1)

	if p.isTypeDefinitionFile(path.Str) {
		// ignore imports to type definition files
		return nil, nil
	}

	imp := &ast.ImportStmt{
		Pos:   t.Pos,
		Alias: alias.Str,
		Path:  path.Str,
	}

	return imp, nil
}

func (p *parser) parseClassDeclStmt() (*ast.ClassDeclStmt, error) {
	var err error
	var t *ast.Token

	if t, err = p.accept(ast.CLASS); err != nil {
		return nil, err
	}

	c := &ast.ClassDeclStmt{Pos: t.Pos}

	// class name
	if t, err = p.accept(ast.IDENT); err != nil {
		return nil, err
	}
	c.Name = t.Str

	if _, err := p.accept(ast.LBRACE); err != nil {
		return nil, err
	}

	for {
		t := p.peek()
		switch t.Type {
		case ast.IDENT:
			var private bool
			switch t.Str {
			case "private":
				private = true
				t = p.next()

			case "exported":
				return nil, NewError(t.Pos, "Unexpected 'exported'. Members are exported by default")
			}

			if p.peekTwo().Type == ast.LPAREN {
				f, err := p.parseFuncDeclStmt(!private, t)
				if err != nil {
					return nil, err
				}
				c.Functions = append(c.Functions, f)
			} else {
				f, err := p.parseVarDeclStmt(false)
				if err != nil {
					return nil, err
				}
				f.Exported = !private
				c.Fields = append(c.Fields, f)
			}

		case ast.RBRACE:
			p.next()
			return c, nil

		case ast.EOF:
			return nil, NewError(t.Pos, "Unclosed class")

		default:
			return nil, NewError(t.Pos, "Unexepected %s", t.Str)
		}
	}
}

/*
The syntax of a enum is:

	enum Direction {
	    Up = 1,
	    Down,
	    Left,
	    Right
	}

	enum Direction {
		Up = "up",
		Down = "down"
	}

we can parse it as a map expression:

	let Direction = {
	    Up: 1,
	    Down: 2,
	    Left: 3,
	    Right: 4
	}

*/
func (p *parser) parseEnumDeclStmt(exported bool) (*ast.EnumDeclStmt, error) {
	p.next()

	t, err := p.accept(ast.IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(ast.LBRACE); err != nil {
		return nil, err
	}

	var values []ast.EnumValue
	i := 0

	for p.peek().Type != ast.RBRACE {
		tt := p.next()
		switch t.Type {
		case ast.IDENT, ast.DELETE:
		default:
			return nil, NewError(t.Pos, "Expecting enum value got %v", tt.Type)
		}

		v := &ast.EnumValue{
			Pos:  tt.Pos,
			Name: tt.Str,
		}

		tt = p.peek()

		// allow to specify the value
		if tt.Type == ast.ASSIGN {
			p.next()

			tt = p.peek()

			switch tt.Type {
			case ast.INT:
				if tt, err = p.accept(ast.INT); err != nil {
					return nil, err
				}
				v.Kind = ast.INT
				v.Value = &ast.ConstantExpr{Pos: tt.Pos, Kind: ast.INT, Value: tt.Str}

				// if this is the first element start the counter from the first value
				if i == 0 {
					if i, err = strconv.Atoi(tt.Str); err != nil {
						return nil, err
					}
				}

			case ast.STRING:
				p.next()
				v.Kind = ast.STRING
				v.Value = &ast.ConstantExpr{Pos: tt.Pos, Kind: ast.STRING, Value: tt.Str}

			default:
				return nil, NewError(tt.Pos, "Expecting ast.INT or STRING, got %v", tt.Type)
			}
		} else {
			v.Kind = ast.INT
			v.Value = &ast.ConstantExpr{Pos: tt.Pos, Kind: ast.INT, Value: strconv.Itoa(i)}
		}

		i++
		p.ignore(ast.COMMA, 1)
		values = append(values, *v)
	}

	enum := &ast.EnumDeclStmt{
		Pos:      t.Pos,
		Name:     t.Str,
		Values:   values,
		Exported: exported,
	}

	// consume the rbrace
	p.next()

	// optional semicolon
	p.ignore(ast.SEMICOLON, 1)

	return enum, nil
}

func (p *parser) parseFuncDeclStmt(exported bool, t *ast.Token) (*ast.FuncDeclStmt, error) {
	var err error

	f := &ast.FuncDeclStmt{Pos: t.Pos}

	// func name
	if t, err = p.accept(ast.IDENT); err != nil {
		return nil, err
	}
	f.Name = t.Str

	if err := p.ignoreGenericDecl(); err != nil {
		return nil, err
	}

	args, variadic, err := p.parseArguments()
	if err != nil {
		return nil, err
	}
	f.Args = args
	f.Variadic = variadic
	f.Exported = exported

	if err := p.ignoreUnionTypeDecl(); err != nil {
		return nil, err
	}

	if err := p.ignoreGenericDecl(); err != nil {
		return nil, err
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	f.Body = body

	p.ignore(ast.SEMICOLON, 1)

	if Optimizations {
		p.setTailCall(f)
	}

	return f, nil
}

// if the function ends in a tail call, replace it
func (p *parser) setTailCall(f *ast.FuncDeclStmt) {
	list := f.Body.List
	ln := len(list)
	if ln == 0 {
		return
	}

	lastStmt := list[ln-1]

	var call *ast.CallExpr

	switch t := lastStmt.(type) {
	case *ast.ReturnStmt:
		var ok bool
		call, ok = t.Value.(*ast.CallExpr)
		if !ok {
			return
		}

	case *ast.CallStmt:
		call = t.CallExpr

	default:
		return
	}

	ident, ok := call.Ident.(*ast.IdentExpr)
	if !ok {
		return
	}

	if ident.Name != f.Name {
		return
	}

	list[ln-1] = &ast.TailCallStmt{CallExpr: call}
}

func (p *parser) parseLambda() (*ast.FuncDeclExpr, error) {
	t := p.peek()
	f := &ast.FuncDeclExpr{Pos: t.Pos}

	switch t.Type {
	case ast.LPAREN:
		args, variadic, err := p.parseArguments()
		if err != nil {
			return nil, err
		}
		f.Args = args
		f.Variadic = variadic
	case ast.IDENT:
		// only one ident is supported without parenthesis
		t := p.next()
		list := []*ast.Field{{Pos: t.Pos, Name: t.Str}}
		f.Args = &ast.Arguments{Opening: t.Pos, List: list}
	}

	_, err := p.accept(ast.LAMBDA)
	if err != nil {
		return nil, err
	}

	// if it's a lambda with body: "(t) => { return t }"
	if p.peek().Type == ast.LBRACE {
		block, err := p.parseBlockStmt()
		if err != nil {
			return nil, err
		}
		f.Body = block
		return f, nil
	}

	// The body is an expression: "(t) => t * 2"
	expr, err := p.parseValueExpression()
	if err != nil {
		return nil, err
	}

	// We need to create a implicit return statment with the expression.
	ret := &ast.ReturnStmt{Pos: expr.Position(), Value: expr}
	f.Body = &ast.BlockStmt{Lbrace: expr.Position(), List: []ast.Stmt{ret}}
	return f, nil
}

func (p *parser) parseFuncDeclExpr() (*ast.FuncDeclExpr, error) {
	t, err := p.accept(ast.FUNCTION)
	if err != nil {
		return nil, err
	}

	f := &ast.FuncDeclExpr{Pos: t.Pos}

	args, variadic, err := p.parseArguments()
	if err != nil {
		return nil, err
	}
	f.Args = args
	f.Variadic = variadic

	if err := p.ignoreUnionTypeDecl(); err != nil {
		return nil, err
	}

	if err := p.ignoreGenericDecl(); err != nil {
		return nil, err
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	f.Body = body

	p.ignore(ast.SEMICOLON, 1)
	return f, nil
}

// returns a boolean indicating if it is variadic
func (p *parser) parseArguments() (*ast.Arguments, bool, error) {
	openning, err := p.accept(ast.LPAREN)
	if err != nil {
		return nil, false, err
	}

	var variadic bool

	var fields []*ast.Field
	for {
		if p.peek().Type == ast.PERIOD {
			for i := 0; i < 3; i++ {
				if t, err := p.accept(ast.PERIOD); err != nil {
					return nil, false, NewError(t.Pos, "Invalid period. ¿wrong variadic attempt?")
				}
			}
			variadic = true
		}

		if p.peek().Type != ast.IDENT {
			break
		}

		t := p.next()

		f := &ast.Field{Pos: t.Pos, Name: t.Str}

		if p.peek().Type == ast.QUESTION {
			f.Optional = true
			p.next()
		}

		fields = append(fields, f)

		p.ignore(ast.QUESTION, 1)

		if err := p.ignoreUnionTypeDecl(); err != nil {
			return nil, false, err
		}

		if err := p.ignoreGenericDecl(); err != nil {
			return nil, false, err
		}

		if variadic {
			if p.peek().Type == ast.COMMA {
				return nil, false, NewError(t.Pos, "No more parameters allowed after a variadic one")
			}
			break // variadic parameter must be the last one
		}

		p.ignore(ast.COMMA, 1)
	}

	if _, err := p.accept(ast.RPAREN); err != nil {
		return nil, false, err
	}

	return &ast.Arguments{Opening: openning.Pos, List: fields}, variadic, nil
}

func (p *parser) parseBlockStmt() (*ast.BlockStmt, error) {
	t, err := p.accept(ast.LBRACE)
	if err != nil {
		return nil, err
	}

	b := &ast.BlockStmt{Lbrace: t.Pos}

	for {
		j := p.peek()
		switch j.Type {

		case ast.RBRACE:
			p.next()
			b.Rbrace = j.Pos
			return b, nil

		case ast.EOF:
			return nil, NewError(t.Pos, "Unclosed block")

		case ast.IDENT:
			if t.Str == "type" {
				// type definitions like: type a = "foo" | "bar";
				if err := p.ignoreTypeDefinition(); err != nil {
					return nil, err
				}
			} else {
				stmt, err := p.parseStmt()
				if err != nil {
					return nil, err
				}
				b.List = append(b.List, stmt)
			}

		default:
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			b.List = append(b.List, stmt)
		}
	}
}

func (p *parser) parseStmt() (ast.Stmt, error) {
	t := p.peek()
	switch t.Type {
	case ast.EXPORT:
		// parseExportStmtOrNIL can return nil because there is no
		// equivalent statement like "export interface"
		stmt, err := p.parseExportStmtOrNIL()
		if err != nil {
			return nil, err
		}
		if stmt == nil {
			return p.parseStmt()
		}
		return stmt, nil
	case ast.ENUM:
		return p.parseEnumDeclStmt(false)
	case ast.LET, ast.VAR:
		p.next()
		return p.parseVarDeclStmt(false)
	case ast.CONST:
		p.next()
		return p.parseVarDeclStmt(true)
	case ast.NEW:
		return p.parseNewInstanceStmt()
	case ast.FOR:
		return p.parseForStmt()
	case ast.WHILE:
		return p.parseWhileStmt()
	case ast.IF:
		return p.parseIfStmt()
	case ast.SWITCH:
		return p.parseSwitchStmt()
	case ast.LPAREN:
		return p.parseParenStmt()
	case ast.LBRACE:
		return p.parseBlockStmt()
	case ast.IDENT:
		return p.parseIdentStmt()
	case ast.DELETE:
		return p.parseDeleteStmt()
	case ast.RETURN:
		return p.parseReturnStmt()
	case ast.THROW:
		return p.parseThrow()
	case ast.TRY:
		return p.parseTryStmt()
	case ast.BREAK:
		p.next()
		var label string
		t2 := p.peek()
		if t2.Type == ast.IDENT {
			label = t2.Str
			p.next()
		}
		p.ignore(ast.SEMICOLON, 1)
		return &ast.BreakStmt{Pos: t.Pos, Label: label}, nil
	case ast.CONTINUE:
		p.next()
		var label string
		t2 := p.peek()
		if t2.Type == ast.IDENT {
			label = t2.Str
			p.next()
		}
		p.ignore(ast.SEMICOLON, 1)
		return &ast.ContinueStmt{Pos: t.Pos, Label: label}, nil
	case ast.EOF:
		return nil, NewError(t.Pos, "Expecting statement but got ast.EOF")
	default:
		return nil, NewError(t.Pos, "Expecting statement but got %v", t.Str)
	}
}

func (p *parser) parseThrow() (*ast.ThrowStmt, error) {
	t, err := p.accept(ast.THROW)
	if err != nil {
		return nil, err
	}

	exp, err := p.parseValueExpression()
	if err != nil {
		return nil, err
	}

	p.ignore(ast.SEMICOLON, 1)

	return &ast.ThrowStmt{Pos: t.Pos, Value: exp}, nil
}

func (p *parser) parseReturnStmt() (*ast.ReturnStmt, error) {
	t, err := p.accept(ast.RETURN)
	if err != nil {
		return nil, err
	}

	// empty return;
	if p.peek().Type == ast.SEMICOLON {
		p.next()
		return &ast.ReturnStmt{Pos: t.Pos}, nil
	}

	// empty return if the expresion is in the next line
	n := p.peek()
	if n.Type == ast.RBRACE || n.Pos.Line != t.Pos.Line {
		return &ast.ReturnStmt{Pos: t.Pos}, nil
	}

	exp, err := p.parseValueExpression()
	if err != nil {
		return nil, err
	}

	p.ignore(ast.SEMICOLON, 1)

	return &ast.ReturnStmt{Pos: t.Pos, Value: exp}, nil
}

func (p *parser) parseSwitchStmt() (*ast.SwitchStmt, error) {
	t, err := p.accept(ast.SWITCH)
	if err != nil {
		return nil, err
	}
	sw := &ast.SwitchStmt{Pos: t.Pos}

	if _, err := p.accept(ast.LPAREN); err != nil {
		return nil, err
	}
	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	sw.Expression = exp

	if _, err := p.accept(ast.RPAREN); err != nil {
		return nil, err
	}

	if _, err := p.accept(ast.LBRACE); err != nil {
		return nil, err
	}

	var blocks []*ast.CaseBlock

LOOP:
	for {
		switch p.peek().Type {
		case ast.CASE:
			block, err := p.parseCaseBlock()
			if err != nil {
				return nil, err
			}
			if err := assertDuplicates(sw.Blocks, block); err != nil {
				return nil, err
			}
			blocks = append(blocks, block)
			sw.Blocks = append(sw.Blocks, block)

		case ast.DEFAULT:
			pos := p.peek().Pos
			d, err := p.parseDefaultBlock()
			if err != nil {
				return nil, err
			}
			block := &ast.CaseBlock{Pos: pos, Stmts: d}
			blocks = append(blocks, block)
			sw.Default = block

		default:
			break LOOP
		}
	}

	if _, err := p.accept(ast.RBRACE); err != nil {
		return nil, err
	}

	// validate fallthroughs.
	// Only empty cases or the last one are allowed without a break or exit stmt
	for i, l := 0, len(blocks)-1; i < l; i++ {
		block := blocks[i]
		ln := len(block.Stmts)
		if ln == 0 {
			continue
		}
		last := block.Stmts[ln-1]
		switch last.(type) {
		case *ast.BreakStmt, *ast.ContinueStmt, *ast.ReturnStmt, *ast.ThrowStmt:
			continue
		}
		return nil, NewError(block.Pos, "Fallthrough is only allowed in empty case")
	}

	return sw, nil
}

func assertDuplicates(blocks []*ast.CaseBlock, block *ast.CaseBlock) error {
	for _, b := range blocks {
		switch t1 := b.Expression.(type) {
		case *ast.ConstantExpr:
			switch t2 := block.Expression.(type) {
			case *ast.ConstantExpr:
				if t1.Kind == t2.Kind && t1.Value == t2.Value {
					return NewError(block.Pos, "Duplicate case: %s", t2.Value)
				}
			}
		}
	}
	return nil
}

func (p *parser) parseDefaultBlock() ([]ast.Stmt, error) {
	t, err := p.accept(ast.DEFAULT)
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(ast.COLON); err != nil {
		return nil, err
	}

	var statements []ast.Stmt

	for {
		j := p.peek()
		switch j.Type {

		case ast.CASE, ast.RBRACE:
			return statements, nil

		case ast.EOF:
			return nil, NewError(t.Pos, "Unclosed case")

		case ast.IDENT:
			if t.Str == "type" {
				// type definitions like: type a = "foo" | "bar";
				if err := p.ignoreTypeDefinition(); err != nil {
					return nil, err
				}
			} else {
				stmt, err := p.parseStmt()
				if err != nil {
					return nil, err
				}
				statements = append(statements, stmt)
			}

		default:
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		}
	}
}

func (p *parser) parseCaseBlock() (*ast.CaseBlock, error) {
	t, err := p.accept(ast.CASE)
	if err != nil {
		return nil, err
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(ast.COLON); err != nil {
		return nil, err
	}

	block := &ast.CaseBlock{Pos: t.Pos, Expression: expr}

	for {
		j := p.peek()
		switch j.Type {

		case ast.CASE, ast.DEFAULT, ast.RBRACE:
			return block, nil

		case ast.EOF:
			return nil, NewError(t.Pos, "Unclosed case")

		case ast.IDENT:
			if t.Str == "type" {
				// type definitions like: type a = "foo" | "bar";
				if err := p.ignoreTypeDefinition(); err != nil {
					return nil, err
				}
			} else {
				stmt, err := p.parseStmt()
				if err != nil {
					return nil, err
				}
				block.Stmts = append(block.Stmts, stmt)
			}

		default:
			stmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			block.Stmts = append(block.Stmts, stmt)
		}
	}
}

func (p *parser) parseTryStmt() (*ast.TryStmt, error) {
	t, err := p.accept(ast.TRY)
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}

	try := &ast.TryStmt{Pos: t.Pos, Body: body}

	if p.peek().Type == ast.CATCH {
		p.next()
		if p.peek().Type == ast.LPAREN {
			p.next()
			i, err := p.parseSimpleIdentExpr()
			if err != nil {
				return nil, err
			}
			try.CatchIdent = &ast.VarDeclStmt{Pos: i.Pos, Name: i.Name}
			if _, err := p.accept(ast.RPAREN); err != nil {
				return nil, err
			}
		}

		catch, err := p.parseBlockStmt()
		if err != nil {
			return nil, err
		}
		try.Catch = catch
	}

	if p.peek().Type == ast.FINALLY {
		p.next()
		finally, err := p.parseBlockStmt()
		if err != nil {
			return nil, err
		}
		try.Finally = finally
	}

	return try, nil
}

func (p *parser) parseIfStmt() (*ast.IfStmt, error) {
	t, err := p.accept(ast.IF)
	if err != nil {
		return nil, err
	}
	ifStmt := &ast.IfStmt{Pos: t.Pos}

	if _, err := p.accept(ast.LPAREN); err != nil {
		return nil, err
	}

	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(ast.RPAREN); err != nil {
		return nil, err
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	ifStmt.IfBlocks = append(ifStmt.IfBlocks, &ast.IfBlock{exp, body})

	for {
		if p.peek().Type != ast.ELSE {
			break
		}

		p.next()

		switch p.peek().Type {
		case ast.IF:
			p.next()
			exp, err = p.parseExpression()
			if err != nil {
				return nil, err
			}

			body, err = p.parseBlockStmt()
			if err != nil {
				return nil, err
			}
			ifStmt.IfBlocks = append(ifStmt.IfBlocks, &ast.IfBlock{exp, body})

		case ast.LBRACE:
			body, err = p.parseBlockStmt()
			if err != nil {
				return nil, err
			}
			ifStmt.Else = body

		default:
			return nil, NewError(t.Pos, "Expecting ast.ELSE")
		}
	}

	return ifStmt, nil
}

func (p *parser) parseWhileStmt() (*ast.WhileStmt, error) {
	t, err := p.accept(ast.WHILE)
	if err != nil {
		return nil, err
	}
	w := &ast.WhileStmt{Pos: t.Pos}

	if _, err := p.accept(ast.LPAREN); err != nil {
		return nil, err
	}

	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	w.Expression = exp

	if _, err := p.accept(ast.RPAREN); err != nil {
		return nil, err
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	w.Body = body
	return w, nil
}

func (p *parser) parseForStmt() (*ast.ForStmt, error) {
	t, err := p.accept(ast.FOR)
	if err != nil {
		return nil, err
	}
	f := &ast.ForStmt{Pos: t.Pos}

	if err := p.parseForDeclarationPart(f); err != nil {
		return nil, err
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	f.Body = body
	return f, nil
}

func (p *parser) parseForDeclarationPart(f *ast.ForStmt) error {
	if _, err := p.accept(ast.LPAREN); err != nil {
		return err
	}

	// parse the declaration part
	t := p.peek()
	switch t.Type {
	case ast.LET, ast.VAR:
		switch p.peekThree().Str {
		case "of", "in":
			return p.parseForInOfDeclarationPart(f)

		default:
			p.next()
			dec, err := p.parseVarDeclStmt(false)
			if err != nil {
				return err
			}
			f.Declaration = append(f.Declaration, dec)

			// allow multiple declarations
			for p.peek().Type == ast.COMMA {
				p.next()
				dec, err := p.parseVarDeclStmt(false)
				if err != nil {
					return err
				}
				f.Declaration = append(f.Declaration, dec)
			}
		}

	case ast.IDENT:
		dec, err := p.parseIdentStmt()
		if err != nil {
			return err
		}
		f.Declaration = append(f.Declaration, dec)
	case ast.SEMICOLON:
		p.next()
	default:
		return NewError(t.Pos, "Expecting declaration")
	}

	// the comparisson part
	switch p.peek().Type {
	case ast.SEMICOLON:
		p.next()
	default:
		exp, err := p.parseExpression()
		if err != nil {
			return err
		}
		f.Expression = exp
	}

	// the step part
	switch p.peek().Type {
	case ast.SEMICOLON:
		p.next()
	case ast.RPAREN:
	default:
		stm, err := p.parseIdentStmt()
		if err != nil {
			return err
		}
		f.Step = stm
	}

	if _, err := p.accept(ast.RPAREN); err != nil {
		return err
	}
	return nil
}

func (p *parser) parseForInOfDeclarationPart(f *ast.ForStmt) error {
	t := p.peek()

	switch t.Type {
	case ast.LET, ast.VAR:
		switch p.peekThree().Str {
		case "of", "in":
			dec, err := p.parseForInOfVarDeclStmt()
			if err != nil {
				return err
			}
			f.Declaration = []ast.Stmt{dec}
		default:
			NewError(t.Pos, "expected on or in")
		}
	default:
		return NewError(t.Pos, "Expecting declaration")
	}

	t = p.next()
	switch t.Str {
	case "of":
		exp, err := p.parseExpression()
		if err != nil {
			return err
		}
		f.OfExpression = exp
	case "in":
		exp, err := p.parseExpression()
		if err != nil {
			return err
		}
		f.InExpression = exp
	default:
		return NewError(t.Pos, "Expecting ON or IN, got: '%s'", t.Str)
	}

	if _, err := p.accept(ast.RPAREN); err != nil {
		return err
	}
	return nil
}

func (p *parser) parseForInOfVarDeclStmt() (*ast.VarDeclStmt, error) {
	p.next()

	t, err := p.accept(ast.IDENT)
	if err != nil {
		return nil, err
	}

	if err := p.ignoreUnionTypeDecl(); err != nil {
		return nil, err
	}

	return &ast.VarDeclStmt{Pos: t.Pos, Name: t.Str}, nil
}

func (p *parser) isPrototype() bool {
	if p.peek().Type != ast.IDENT {
		return false
	}
	if p.peekTwo().Type != ast.PERIOD {
		return false
	}
	return p.peekThree().Str == "prototype"
}

func (p *parser) parseMethod() (*ast.FuncDeclStmt, error) {
	// the format is Map.prototype.Foo = function() {...}
	t, err := p.accept(ast.IDENT)
	if err != nil {
		return nil, err
	}
	rt := t.Str

	// consume the ".prototype." part
	if _, err := p.accept(ast.PERIOD); err != nil {
		return nil, err
	}
	t, err = p.accept(ast.IDENT)
	if err != nil {
		return nil, err
	}
	if t.Str != "prototype" {
		return nil, NewError(t.Pos, "Expecting 'prototype', got: '%s'", t.Str)
	}

	if _, err := p.accept(ast.PERIOD); err != nil {
		return nil, err
	}

	f := &ast.FuncDeclStmt{Pos: t.Pos, ReceiverType: rt}

	// func name
	if t, err = p.accept(ast.IDENT); err != nil {
		return nil, err
	}
	f.Name = t.Str

	// consume the "= function" part
	if _, err := p.accept(ast.ASSIGN); err != nil {
		return nil, err
	}
	if _, err := p.accept(ast.FUNCTION); err != nil {
		return nil, err
	}

	if err := p.ignoreGenericDecl(); err != nil {
		return nil, err
	}

	args, variadic, err := p.parseArguments()
	if err != nil {
		return nil, err
	}
	f.Args = args
	f.Variadic = variadic

	if err := p.ignoreUnionTypeDecl(); err != nil {
		return nil, err
	}

	if err := p.ignoreGenericDecl(); err != nil {
		return nil, err
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	f.Body = body

	p.ignore(ast.SEMICOLON, 1)

	return f, nil
}

func (p *parser) parseNewInstanceStmt() (ast.Stmt, error) {
	peek := p.peek()

	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	switch t := exp.(type) {
	case *ast.CallExpr:
		return &ast.CallStmt{t}, nil
	default:
		return nil, NewError(peek.Pos, "Unexpected %v", peek.Type)
	}
}

func (p *parser) parseParenStmt() (ast.Stmt, error) {
	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	switch t := exp.(type) {
	case *ast.CallExpr:
		return &ast.CallStmt{t}, nil
	}

	t := p.peek()

	switch t.Type {
	case ast.ASSIGN:
		return p.parseAssignStmt(exp)
	case ast.ADD_ASSIGN, ast.SUB_ASSIGN, ast.MUL_ASSIGN,
		ast.DIV_ASSIGN, ast.BOR_ASSIGN, ast.XOR_ASSIGN:
		return p.parseAddOrSubAssignStmt(exp)
	case ast.INC:
		return p.parseIncStmt(exp)
	case ast.DEC:
		return p.parseDecStmt(exp)
	default:
		return nil, NewError(t.Pos, "Unexpected %v", t.Type)
	}
}

func (p *parser) parseIdentStmt() (ast.Stmt, error) {
	if p.isPrototype() {
		return p.parseMethod()
	}

	ident, err := p.parseIdentExpr()
	if err != nil {
		return nil, err
	}

	switch t := ident.(type) {
	case *ast.CallExpr:
		return &ast.CallStmt{t}, nil
	case *ast.IdentExpr:
		if p.peek().Type == ast.COLON {
			p.next()
			return p.parseLabelStmt(t.Name)
		}
	}

	t := p.peek()
	switch t.Type {
	case ast.ASSIGN:
		return p.parseAssignStmt(ident)
	case ast.ADD_ASSIGN, ast.SUB_ASSIGN, ast.MUL_ASSIGN,
		ast.DIV_ASSIGN, ast.BOR_ASSIGN, ast.XOR_ASSIGN, ast.MOD_ASSIGN:
		return p.parseAddOrSubAssignStmt(ident)
	case ast.INC:
		return p.parseIncStmt(ident)
	case ast.DEC:
		return p.parseDecStmt(ident)
	default:
		return nil, NewError(t.Pos, "Unexpected %v", t.Type)
	}
}

func (p *parser) parseDeleteStmt() (*ast.DeleteStmt, error) {
	td := p.next()

	t1, err := p.acceptIdent()
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(ast.PERIOD); err != nil {
		return nil, err
	}

	t2, err := p.acceptIdent()
	if err != nil {
		return nil, err
	}

	exp := &ast.DeleteStmt{
		Object:   t1.Str,
		Property: t2.Str,
		Pos:      td.Pos,
	}

	return exp, nil
}

func (p *parser) parseTypeof() (*ast.TypeofExpr, error) {
	return nil, NewError(p.peek().Pos, "The typeof operator is not valid. Use reflect.typeOf")

	// p.next()

	// exp, err := p.parseFactor()
	// if err != nil {
	// 	return nil, err
	// }

	// return &ast.TypeofExpr{Expr: exp}, nil
}

func (p *parser) parseLabelStmt(name string) (ast.Stmt, error) {
	t := p.peek()
	switch t.Type {
	case ast.FOR:
		stmt, err := p.parseForStmt()
		if err != nil {
			return nil, err
		}
		stmt.SetLabel(name)
		return stmt, nil
	case ast.WHILE:
		stmt, err := p.parseWhileStmt()
		if err != nil {
			return nil, err
		}
		stmt.SetLabel(name)
		return stmt, nil
	case ast.SWITCH:
		stmt, err := p.parseSwitchStmt()
		if err != nil {
			return nil, err
		}
		stmt.SetLabel(name)
		return stmt, nil
	default:
		return nil, NewError(t.Pos, "Unexpected %v", t.Type)
	}
}

func (p *parser) parseIncStmt(exp ast.Expr) (*ast.IncStmt, error) {
	_, err := p.accept(ast.INC)
	if err != nil {
		return nil, err
	}

	p.ignore(ast.SEMICOLON, 1)
	return &ast.IncStmt{Operator: ast.INC, Left: exp}, nil
}

func (p *parser) parseDecStmt(exp ast.Expr) (*ast.IncStmt, error) {
	_, err := p.accept(ast.DEC)
	if err != nil {
		return nil, err
	}

	p.ignore(ast.SEMICOLON, 1)
	return &ast.IncStmt{Operator: ast.DEC, Left: exp}, nil
}

func (p *parser) parseAddOrSubAssignStmt(ident ast.Expr) (*ast.AsignStmt, error) {
	var operator ast.Type
	switch p.next().Type {
	case ast.ADD_ASSIGN:
		operator = ast.ADD
	case ast.SUB_ASSIGN:
		operator = ast.SUB
	case ast.MUL_ASSIGN:
		operator = ast.MUL
	case ast.DIV_ASSIGN:
		operator = ast.DIV
	case ast.BOR_ASSIGN:
		operator = ast.BOR
	case ast.XOR_ASSIGN:
		operator = ast.XOR
	case ast.MOD_ASSIGN:
		operator = ast.MOD
	}

	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.ignore(ast.SEMICOLON, 1)

	// convert ast.IDENT += EXP to ast.IDENT = ast.IDENT + EXP
	b := &ast.BinaryExpr{
		Operator: operator,
		Left:     ident,
		Right:    exp,
	}

	d := &ast.AsignStmt{Left: ident, Value: b}
	return d, nil
}

func (p *parser) parseAssignStmt(exp ast.Expr) (*ast.AsignStmt, error) {
	_, err := p.accept(ast.ASSIGN)
	if err != nil {
		return nil, err
	}

	right, err := p.parseValueExpression()
	if err != nil {
		return nil, err
	}

	p.ignore(ast.SEMICOLON, 1)
	return &ast.AsignStmt{exp, right}, nil
}

func (p *parser) parseDeclareGlobal() ([]ast.Stmt, error) {
	if t := p.next(); t.Str != "declare" {
		return nil, NewError(t.Pos, "Expected declare")
	}

	if t := p.next(); t.Str != "global" {
		return nil, NewError(t.Pos, "Expected global")
	}

	if _, err := p.accept(ast.LBRACE); err != nil {
		return nil, err
	}

	var stmts []ast.Stmt

loop:
	for {
		t := p.peek()
		switch t.Type {

		case ast.INTERFACE:
			if err := p.ignoreInterface(); err != nil {
				return nil, err
			}

		case ast.RBRACE:
			p.next()
			break loop

		case ast.ENUM:
			enum, err := p.parseEnumDeclStmt(false)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, enum)

		case ast.CONST:
			p.next()
			c, err := p.parseVarDeclStmt(true)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, c)

		default:
			return nil, NewError(t.Pos, "Unexpected token: %s", t.Str)
		}
	}

	return stmts, nil
}

// parseExportStmtOrNIL can return nil because there is no
// equivalent statement like "export interface"
func (p *parser) parseExportStmtOrNIL() (ast.Stmt, error) {
	if _, err := p.accept(ast.EXPORT); err != nil {
		return nil, err
	}

	t := p.peek()
	switch t.Type {
	case ast.ENUM:
		return p.parseEnumDeclStmt(true)

	case ast.FUNCTION:
		p.next()
		return p.parseFuncDeclStmt(true, t)

	case ast.CLASS:
		cl, err := p.parseClassDeclStmt()
		if err != nil {
			return nil, err
		}
		cl.Exported = true
		return cl, nil

	case ast.INTERFACE:
		if err := p.ignoreInterface(); err != nil {
			return nil, err
		}
		return nil, nil

	case ast.IDENT:
		if t.Str == "type" {
			err := p.ignoreTypeDefinition()
			return nil, err
		}
		return nil, NewError(t.Pos, "Unexpected %v after export", t.Type)

	case ast.LET, ast.VAR:
		p.next()
		s, err := p.parseVarDeclStmt(false)
		if err != nil {
			return nil, err
		}
		s.Exported = true
		return s, nil
	case ast.CONST:
		p.next()
		s, err := p.parseVarDeclStmt(true)
		if err != nil {
			return nil, err
		}
		s.Exported = true
		return s, nil

	default:
		return nil, NewError(t.Pos, "Unexpected %v after export", t.Type)
	}
}

func (p *parser) parseVarDeclStmt(isConst bool) (*ast.VarDeclStmt, error) {
	t, err := p.accept(ast.IDENT)
	if err != nil {
		return nil, err
	}

	if err := p.ignoreUnionTypeDecl(); err != nil {
		return nil, err
	}

	if p.peek().Type != ast.ASSIGN {
		p.ignore(ast.SEMICOLON, 1)
		v := &ast.ConstantExpr{t.Pos, ast.UNDEFINED, "undefined"}
		return &ast.VarDeclStmt{Pos: t.Pos, Name: t.Str, Value: v}, nil
	}

	if _, err := p.accept(ast.ASSIGN); err != nil {
		return nil, err
	}

	expr, err := p.parseValueExpression()
	if err != nil {
		return nil, err
	}

	v := &ast.VarDeclStmt{
		Pos:   t.Pos,
		Name:  t.Str,
		Value: expr,
		Const: isConst,
	}

	p.ignore(ast.SEMICOLON, 1)
	return v, nil
}

func (p *parser) ignoreUnionTypeDecl() error {
	if p.peek().Type != ast.COLON {
		return nil
	}
	p.next()

	if err := p.ignoreTypeDecl(); err != nil {
		return err
	}

	// allow union types
	for p.peek().Type == ast.BOR {
		p.next()
		if err := p.ignoreTypeDecl(); err != nil {
			return err
		}
	}

	if err := p.ignoreGenericDecl(); err != nil {
		return err
	}

	return nil
}

func (p *parser) ignoreLambda() error {
	if err := p.ignoreFuncArgs(); err != nil {
		return err
	}

	if _, err := p.accept(ast.LAMBDA); err != nil {
		return err
	}

	if err := p.ignoreTypeDecl(); err != nil {
		return err
	}

	if err := p.ignoreGenericDecl(); err != nil {
		return err
	}

	p.ignore(ast.SEMICOLON, 1)
	return nil
}

func (p *parser) ignoreAsExpression() error {
	t := p.peek()
	if t.Type != ast.IDENT || t.Str != "as" {
		return nil
	}

	p.next()

	if err := p.ignoreTypeDecl(); err != nil {
		return err
	}

	return nil
}

func (p *parser) ignoreTypeDecl() error {
loop:
	for {
		t := p.peek()
		switch t.Type {
		case ast.LPAREN:
			if p.peekTwo().Type == ast.IDENT && p.peekThree().Type != ast.COLON {
				return p.ignoreCompositeType()
			}
			if err := p.ignoreLambda(); err != nil {
				return err
			}
		case ast.LBRACE:
			if err := p.ignoreInterfaceBody(); err != nil {
				return err
			}
		case ast.LBRACK:
			p.next()
			if err := p.ignoreInterfaceBody(); err != nil {
				return err
			}
			if _, err := p.accept(ast.RBRACK); err != nil {
				return err
			}
		case ast.IDENT, ast.NULL:
			p.next()
			// the type may have a package prefix
			if p.peek().Type == ast.PERIOD {
				p.next()
				if _, err := p.accept(ast.IDENT); err != nil {
					return err
				}
			}
		case ast.STRING:
			p.next()
		default:
			return NewError(t.Pos, "Expecting ast.IDENT or NULL, got %s", t.Str)
		}

		switch p.peek().Type {
		case ast.BOR, ast.AND:
			p.next()
		default:
			break loop
		}
	}

	// allow union types
	for p.peek().Type == ast.LBRACK {
		p.next()
		if _, err := p.accept(ast.RBRACK); err != nil {
			return err
		}
	}

	return nil
}

// for example: (SaleLine & BankAccount)[]
func (p *parser) ignoreCompositeType() error {
	if _, err := p.accept(ast.LPAREN); err != nil {
		return err
	}

loop:
	for {
		if _, err := p.accept(ast.IDENT); err != nil {
			return err
		}

		switch p.peek().Type {
		case ast.BOR, ast.AND:
			p.next()
		default:
			break loop
		}
	}

	if _, err := p.accept(ast.RPAREN); err != nil {
		return err
	}

	for p.peek().Type == ast.LBRACK {
		p.next()
		if _, err := p.accept(ast.RBRACK); err != nil {
			return err
		}
	}

	return nil
}

// ignore type declarations like:
//
//   type foo = "bar" | "foo";
//
// or:
//
//   type foo = () => void;
//
func (p *parser) ignoreTypeDefinition() error {
	p.next()

	if _, err := p.accept(ast.IDENT); err != nil {
		return err
	}

	if _, err := p.accept(ast.ASSIGN); err != nil {
		return err
	}

	if err := p.ignoreGenericDecl(); err != nil {
		return err
	}

	if p.peek().Type == ast.LPAREN {
		return p.ignoreLambda()
	}

loop:
	for {
		t := p.peek()
		switch t.Type {
		case ast.STRING, ast.IDENT:
			p.next()
		default:
			return NewError(t.Pos, "Expecting string or ident, got %v", t.Type)
		}

		t = p.peek()
		switch t.Type {
		case ast.BOR, ast.AND, ast.PERIOD:
			p.next()
		default:
			break loop
		}
	}

	p.ignore(ast.SEMICOLON, 1)
	return nil
}

func (p *parser) ignoreTypeAssert() error {
	if p.peek().Type != ast.LSS {
		return nil
	}

	p.next()
	if err := p.ignoreType(); err != nil {
		return err
	}

	p.ignore(ast.QUESTION, 1)
	if _, err := p.accept(ast.GTR); err != nil {
		return err
	}
	return nil
}

// with generics we cant know in advance if ast.IDENT< is
// an expression or a generic declaration:
// For example  foo<T>() or foo < T
// So we peek until we know it and consume it or go back.
func (p *parser) tryIgnoreGenericDecl() {
	i := 0
	if t, _ := p.peekToken(i, false); t.Type != ast.LSS {
		return
	}
	i++

	if t, _ := p.peekToken(i, false); t.Type != ast.IDENT {
		return
	}
	i++

	// if it is a selector, advance all its elements
	for {
		if t, _ := p.peekToken(i, false); t.Type == ast.PERIOD {
			i++
		} else {
			break
		}
		if t, _ := p.peekToken(i, false); t.Type != ast.IDENT {
			return
		}
		i++
	}

	for {
		if t, _ := p.peekToken(i, false); t.Type != ast.COMMA {
			break
		}
		i++

		if t, _ := p.peekToken(i, false); t.Type != ast.IDENT {
			return
		}
		i++

		// if it is a selector, advance all its elements
		for {
			if t, _ := p.peekToken(i, false); t.Type == ast.PERIOD {
				i++
			} else {
				break
			}
			if t, _ := p.peekToken(i, false); t.Type != ast.IDENT {
				return
			}
			i++
		}
	}

	if t, _ := p.peekToken(i, false); t.Type != ast.GTR {
		return
	}
	i++

	// now we know it was a generic declaration
	p.index += i
}

func (p *parser) ignoreGenericDecl() error {
	if p.peek().Type != ast.LSS {
		return nil
	}

	p.next()
	if err := p.ignoreType(); err != nil {
		return err
	}

	n := p.peek()
	if n.Type == ast.IDENT && n.Str == "extends" {
		p.next()
		if _, err := p.acceptIdent(); err != nil {
			return err
		}
	}

	for p.peek().Type == ast.COMMA {
		p.next()
		if err := p.ignoreType(); err != nil {
			return err
		}

		n := p.peek()
		if n.Type == ast.IDENT && n.Str == "extends" {
			p.next()
			if _, err := p.acceptIdent(); err != nil {
				return err
			}
		}
	}

	if _, err := p.accept(ast.GTR); err != nil {
		return err
	}

	return nil
}

func (p *parser) acceptIdent() (*ast.Token, error) {
	t := p.next()

	switch t.Type {
	case ast.IDENT, ast.FUNCTION, ast.CLASS:
	// "Function" is  a valid interface name in compiler

	default:
		return nil, NewError(t.Pos, "Expecting ast.IDENT got %v", t.Type)
	}

	return t, nil
}

// Ignore the [] after a type: string[]
func (p *parser) ignoreArrayDecl() error {
	if p.peek().Type != ast.LBRACK {
		return nil
	}

	p.next()
	_, err := p.accept(ast.RBRACK)
	return err
}

func (p *parser) ignoreType() error {
	if _, err := p.acceptIdent(); err != nil {
		return err
	}

	if err := p.ignoreArrayDecl(); err != nil {
		return err
	}

	for p.peek().Type == ast.PERIOD {
		p.next()
		if _, err := p.acceptIdent(); err != nil {
			return err
		}
		if err := p.ignoreArrayDecl(); err != nil {
			return err
		}
	}

	return nil
}

func (p *parser) ignoreInterface() error {
	if _, err := p.accept(ast.INTERFACE); err != nil {
		return err
	}

	if _, err := p.acceptIdent(); err != nil {
		return err
	}

	if err := p.ignoreGenericDecl(); err != nil {
		return err
	}

	switch p.peek().Str {
	case "extends", "implements":
		p.next()
		for {
			if _, err := p.acceptIdent(); err != nil {
				return err
			}

			for p.peek().Type == ast.PERIOD {
				p.next()
				if _, err := p.acceptIdent(); err != nil {
					return err
				}
			}

			if p.peek().Type == ast.COMMA {
				p.next()
				continue
			}
			break
		}
	}

	return p.ignoreInterfaceBody()
}

func (p *parser) ignoreInterfaceBody() error {
	if _, err := p.accept(ast.LBRACE); err != nil {
		return err
	}

	for p.peek().Type != ast.RBRACE {

		switch p.peek().Type {
		case ast.LBRACK:
			// ignore an indexed interface
			// For example: [n: number]: T;
			p.next()
			if _, err := p.accept(ast.IDENT); err != nil {
				return err
			}
			if err := p.ignoreUnionTypeDecl(); err != nil {
				return err
			}
			if _, err := p.accept(ast.RBRACK); err != nil {
				return err
			}

		default:
			t := p.peek()
			switch t.Type {
			case ast.IDENT, ast.DEFAULT, ast.FUNCTION, ast.CLASS, ast.ENUM:
				p.next()
			default:
				return NewError(t.Pos, "Expected IDENT, got %v", t.Type)
			}

			switch p.peek().Type {
			case ast.LSS:
				if err := p.ignoreGenericDecl(); err != nil {
					return err
				}
				if err := p.ignoreFuncArgs(); err != nil {
					return err
				}
			case ast.LPAREN:
				if err := p.ignoreFuncArgs(); err != nil {
					return err
				}
			case ast.QUESTION:
				p.next()
			}
		}

		if err := p.ignoreUnionTypeDecl(); err != nil {
			return err
		}

		t := p.peek()
		switch t.Type {
		case ast.COMMA, ast.SEMICOLON:
			p.next()
		}
	}

	if _, err := p.accept(ast.RBRACE); err != nil {
		return err
	}

	p.ignore(ast.SEMICOLON, 1)
	return nil
}

func (p *parser) ignoreFuncArgs() error {
	if _, err := p.accept(ast.LPAREN); err != nil {
		return err
	}

	for p.peek().Type == ast.IDENT {
		p.next()

		p.ignore(ast.QUESTION, 1)

		if err := p.ignoreUnionTypeDecl(); err != nil {
			return err
		}

		p.ignore(ast.COMMA, 1)
	}

	if _, err := p.accept(ast.RPAREN); err != nil {
		return err
	}

	return nil
}

func (p *parser) parseCallExpr(exp ast.Expr, optional bool) (*ast.CallExpr, error) {
	l, err := p.accept(ast.LPAREN)
	if err != nil {
		return nil, err
	}

	args, spread, err := p.parseCallArgs()
	if err != nil {
		return nil, err
	}

	r, err := p.accept(ast.RPAREN)
	if err != nil {
		return nil, err
	}

	call := &ast.CallExpr{
		Ident:    exp,
		Lparen:   l.Pos,
		Args:     args,
		Rparen:   r.Pos,
		Spread:   spread,
		Optional: optional,
	}

	return call, nil
}

func (p *parser) parseCallArgs() ([]ast.Expr, bool, error) {
	var args []ast.Expr

	var spread bool

loop:
	for {
		t := p.peek()
		switch t.Type {

		case ast.RPAREN:
			break loop

		case ast.COMMA:
			p.next()

		case ast.PERIOD:
			for i := 0; i < 3; i++ {
				if _, err := p.accept(ast.PERIOD); err != nil {
					return nil, false, NewError(t.Pos, "Expecting spread operator")
				}
			}
			spread = true
			exp, err := p.parseValueExpression()
			if err != nil {
				return nil, false, err
			}
			args = append(args, exp)
			if p.peek().Type != ast.RPAREN {
				return nil, false, NewError(t.Pos, "Spread argument must be the last")
			}
			break loop

		default:
			exp, err := p.parseValueExpression()
			if err != nil {
				return nil, false, err
			}
			args = append(args, exp)
		}
	}

	return args, spread, nil
}

// A value can be a function or a boolean expression.
// A function does not have any arithmetic combination possible so
// its parsed at the top level.
func (p *parser) parseValueExpression() (ast.Expr, error) {
	switch p.peek().Type {
	case ast.FUNCTION:
		return p.parseFuncDeclExpr()
	case ast.LPAREN:
		// if a expression starts with a paren we need to guess if it is
		// a lambda. Since the parser is not backtracking we try some basic
		// heuristics:
		switch p.peekTwo().Type {
		case ast.RPAREN:
			// its a lambda with format: "() => ..."
			if p.peekThree().Type == ast.LAMBDA {
				return p.parseLambda()
			}
		case ast.IDENT:
			switch p.peekThree().Type {
			case ast.COMMA:
				// its a lambda with format: "(t,...) => ..."
				return p.parseLambda()
			case ast.COLON:
				// its a lambda with format: "(t:type...) => ..."
				return p.parseLambda()
			case ast.RPAREN:
				// its a lambda with format: "(t) => ..."
				return p.parseLambda()
			}
		}
	case ast.IDENT:
		// its a lambda with format: "t => ..."
		if p.peekTwo().Type == ast.LAMBDA {
			return p.parseLambda()
		}
	}

	return p.parseExpression()
}

func (p *parser) parseExpression() (ast.Expr, error) {
	lh, err := p.parseRelation()
	if err != nil {
		return nil, err
	}
	var e ast.Expr = lh

	if p.peek().Type == ast.QUESTION {
		p.next()
		return p.parseTernaryExpression(e)
	}

loop:
	for {
		t := p.peek()
		switch t.Type {

		case ast.LOR, ast.NOR, ast.LAND:
			p.next()
			rh, err := p.parseRelation()
			if err != nil {
				return nil, err
			}
			e = &ast.BinaryExpr{Left: e, Right: rh, Operator: t.Type}

		default:
			break loop
		}
	}

	return e, nil
}

func (p *parser) parseTernaryExpression(expr ast.Expr) (*ast.TernaryExpr, error) {
	lh, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(ast.COLON); err != nil {
		return nil, err
	}

	rh, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.TernaryExpr{Condition: expr, Left: lh, Right: rh}, nil
}

func (p *parser) parseRelation() (ast.Expr, error) {
	lh, err := p.parseAdditiveExpr()
	if err != nil {
		return nil, err
	}

	var e ast.Expr = lh
loop:
	for {
		t := p.peek()
		switch t.Type {
		case ast.AND, ast.XOR, ast.BOR:
			p.next()
			rh, err := p.parseAdditiveExpr()
			if err != nil {
				return nil, err
			}

			e = &ast.BinaryExpr{
				Left:     e,
				Right:    rh,
				Operator: t.Type,
			}
		case ast.EQL, ast.NEQ, ast.SEQ, ast.SNE, ast.LSS, ast.LEQ, ast.GTR, ast.GEQ:
			p.next()
			rh, err := p.parseAdditiveExpr()
			if err != nil {
				return nil, err
			}

			e = &ast.BinaryExpr{
				Left:     e,
				Right:    rh,
				Operator: t.Type,
			}
		default:
			break loop
		}
	}

	return e, nil
}

func (p *parser) parseAdditiveExpr() (ast.Expr, error) {
	lh, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	var e ast.Expr = lh

loop:
	for {
		t := p.peek()
		switch t.Type {
		case ast.ADD, ast.SUB, ast.LSH, ast.RSH:
			p.next()
			rh, err := p.parseTerm()
			if err != nil {
				return nil, err
			}

			e = &ast.BinaryExpr{
				Left:     e,
				Right:    rh,
				Operator: t.Type,
			}
		case ast.SEMICOLON:
			p.next()
			break loop
		default:
			break loop
		}
	}

	return e, nil
}

func (p *parser) parseTerm() (ast.Expr, error) {
	lh, err := p.parseSignedFactor()
	if err != nil {
		return nil, err
	}

	var e ast.Expr = lh

loop:
	for {
		t := p.peek()
		switch t.Type {
		case ast.MUL, ast.DIV, ast.MOD, ast.EXP:
			p.next()
			rh, err := p.parseSignedFactor()
			if err != nil {
				return nil, err
			}
			e = &ast.BinaryExpr{Left: e, Right: rh, Operator: t.Type}
		default:
			break loop
		}
	}

	return e, nil
}

func (p *parser) parseSignedFactor() (ast.Expr, error) {
	t := p.peek()
	switch t.Type {
	case ast.ADD, ast.SUB, ast.BNT, ast.NOT:
		p.next()
		expr, err := p.parseFactor()
		if err != nil {
			return nil, err
		}

		expr = &ast.UnaryExpr{Pos: t.Pos, Operator: t.Type, Operand: expr}
		if err := p.ignoreAsExpression(); err != nil {
			return nil, err
		}
		return expr, nil
	}

	expr, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	if err := p.ignoreAsExpression(); err != nil {
		return nil, err
	}

	return expr, nil
}

func (p *parser) parseMapExpr() (*ast.MapDeclExpr, error) {
	t, err := p.accept(ast.LBRACE)
	if err != nil {
		return nil, err
	}

	items, err := p.parseMapElementList()
	if err != nil {
		return nil, err
	}

	exp := &ast.MapDeclExpr{t.Pos, items}

	_, err = p.accept(ast.RBRACE)
	if err != nil {
		return nil, err
	}

	return exp, nil
}

func (p *parser) parseIndexDeclExpr() (*ast.ArrayDeclExpr, error) {
	t, err := p.accept(ast.LBRACK)
	if err != nil {
		return nil, err
	}

	items, err := p.parseIndexDeclElementList()
	if err != nil {
		return nil, err
	}

	exp := &ast.ArrayDeclExpr{t.Pos, items}

	_, err = p.accept(ast.RBRACK)
	if err != nil {
		return nil, err
	}

	return exp, nil
}
func (p *parser) parseMapElementList() ([]ast.KeyValue, error) {
	var args []ast.KeyValue

loop:
	for {
		t := p.peek()
		switch t.Type {
		case ast.RBRACE:
			break loop
		case ast.COMMA:
			p.next()
		default:
			key := p.next()
			switch key.Type {
			case ast.STRING, ast.INT, ast.IDENT, ast.FUNCTION, ast.DEFAULT:
			default:
				return nil, NewError(t.Pos, "Expecting string or ident as key")
			}
			_, err := p.accept(ast.COLON)
			if err != nil {
				return nil, err
			}
			exp, err := p.parseValueExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, ast.KeyValue{Key: key.Str, KeyType: key.Type, Value: exp})
		}
	}

	return args, nil
}

func (p *parser) parseIndexDeclElementList() ([]ast.Expr, error) {
	var args []ast.Expr

loop:
	for {
		t := p.peek()
		switch t.Type {
		case ast.RBRACK:
			break loop
		case ast.COMMA:
			p.next()
		default:
			exp, err := p.parseValueExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, exp)
		}
	}

	return args, nil
}

func (p *parser) parseIdentExpr() (ast.Expr, error) {
	exp, err := p.parseSimpleIdentExpr()
	if err != nil {
		return nil, err
	}

	v, err := p.parseValueExpr(exp)
	if err != nil {
		return nil, err
	}

	switch t := v.(type) {
	case *ast.SelectorExpr:
		t.First = true
	case *ast.CallExpr:
		t.First = true
	case *ast.IndexExpr:
		t.First = true
	}

	if err := p.ignoreAsExpression(); err != nil {
		return nil, err
	}

	return v, nil
}

// parse the right part after a value, for example:
//    foo.bar or (foo).bar or (foo)[]
func (p *parser) parseValueExpr(exp ast.Expr) (ast.Expr, error) {
	var err error

LOOP:
	for {
		switch p.peek().Type {
		// optional chaining
		case ast.QUESTION:
			if p.peekTwo().Type == ast.PERIOD {
				p.next()
				switch p.peekTwo().Type {
				case ast.LPAREN:
					p.next()
					// func?.(args)
					exp, err = p.parseCallExpr(exp, true)
					if err != nil {
						return nil, err
					}
				case ast.LBRACK:
					p.next()
					// arr?.[index] or obj?.[expr]
					exp, err = p.parseIndexExpr(exp, true)
					if err != nil {
						return nil, err
					}
				default:
					// obj?.prop
					exp, err = p.parseSelectorExpr(exp, true)
					if err != nil {
						return nil, err
					}
				}
			} else {
				break LOOP
			}
		case ast.PERIOD:
			exp, err = p.parseSelectorExpr(exp, false)
			if err != nil {
				return nil, err
			}
		case ast.LBRACK:
			exp, err = p.parseIndexExpr(exp, false)
			if err != nil {
				return nil, err
			}
		case ast.LPAREN:
			exp, err = p.parseCallExpr(exp, false)
			if err != nil {
				return nil, err
			}
		case ast.SEMICOLON:
			p.next()
			break LOOP
		default:
			break LOOP
		}
	}

	return exp, nil
}

func (p *parser) parseSimpleIdentExpr() (*ast.IdentExpr, error) {
	t := p.peek()

	switch t.Type {
	case ast.IDENT,
		ast.NEW,
		ast.BREAK,
		ast.CONTINUE,
		ast.IF,
		ast.ELSE,
		ast.FOR,
		ast.WHILE,
		ast.RETURN,
		ast.IMPORT,
		ast.SWITCH,
		ast.CASE,
		ast.DEFAULT,
		ast.LET,
		ast.VAR,
		ast.CLASS,
		ast.CONST,
		ast.FUNCTION,
		ast.ENUM,
		ast.NULL,
		ast.UNDEFINED,
		ast.INTERFACE,
		ast.EXPORT,
		ast.TRUE,
		ast.FALSE,
		ast.TRY,
		ast.CATCH,
		ast.FINALLY,
		ast.THROW,
		ast.TYPEOF,
		ast.DELETE:
		p.next()
	default:
		return nil, NewError(t.Pos, "Expecting IDENT, got %v", t.Type)
	}

	p.tryIgnoreGenericDecl()
	return &ast.IdentExpr{t.Pos, t.Str}, nil
}

func (p *parser) parseSelectorExpr(exp ast.Expr, optional bool) (*ast.SelectorExpr, error) {
	_, err := p.accept(ast.PERIOD)
	if err != nil {
		return nil, err
	}

	sel, err := p.parseSimpleIdentExpr()
	if err != nil {
		return nil, err
	}

	return &ast.SelectorExpr{X: exp, Sel: sel, Optional: optional}, nil
}

func (p *parser) parseIndexExpr(exp ast.Expr, optional bool) (ast.Expr, error) {
	l, err := p.accept(ast.LBRACK)
	if err != nil {
		return nil, err
	}

	i, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	r, err := p.accept(ast.RBRACK)
	if err != nil {
		return nil, err
	}

	arr := &ast.IndexExpr{
		Left:     exp,
		Lbrack:   l.Pos,
		Index:    i,
		Rbrack:   r.Pos,
		Optional: optional,
	}

	return arr, nil
}

func (p *parser) parseNewInstanceExpr() (ast.Expr, error) {
	var exp ast.Expr
	var err error

	exp, err = p.parseSimpleIdentExpr()
	if err != nil {
		return nil, err
	}

	if p.peek().Type == ast.PERIOD {
		exp, err = p.parseSelectorExpr(exp, false)
		if err != nil {
			return nil, err
		}
	}

	l, err := p.accept(ast.LPAREN)
	if err != nil {
		return nil, err
	}

	args, spread, err := p.parseCallArgs()
	if err != nil {
		return nil, err
	}

	r, err := p.accept(ast.RPAREN)
	if err != nil {
		return nil, err
	}

	n := &ast.NewInstanceExpr{
		Name:   exp,
		Lparen: l.Pos,
		Args:   args,
		Rparen: r.Pos,
		Spread: spread,
	}

	return n, nil
}

func (p *parser) parseFactor() (ast.Expr, error) {
	if err := p.ignoreTypeAssert(); err != nil {
		return nil, err
	}

	t := p.peek()
	switch t.Type {
	case ast.HEX:
		p.next()
		i, err := strconv.ParseInt(t.Str, 0, 64)
		if err != nil {
			return nil, NewError(t.Pos, "Error parsing Hex: %v", err)
		}
		return &ast.ConstantExpr{t.Pos, ast.INT, strconv.Itoa(int(i))}, nil

	case ast.INT, ast.FLOAT, ast.STRING, ast.RUNE:
		p.next()
		return &ast.ConstantExpr{t.Pos, t.Type, t.Str}, nil

	case ast.NULL:
		p.next()
		// the compiler internally uses nil instead of null.z
		return &ast.ConstantExpr{t.Pos, ast.NULL, t.Str}, nil

	case ast.UNDEFINED:
		p.next()
		// the compiler internally uses nil instead of null.
		return &ast.ConstantExpr{t.Pos, ast.UNDEFINED, t.Str}, nil

	case ast.NEW:
		p.next()
		exp, err := p.parseNewInstanceExpr()
		if err != nil {
			return nil, err
		}

		// check if there is an expression after the parenthesis like: (foo).bar
		if exp, err = p.parseValueExpr(exp); err != nil {
			return nil, err
		}

		return exp, nil

	case ast.IDENT, ast.DEFAULT:
		exp, err := p.parseIdentExpr()
		if err != nil {
			return nil, err
		}
		return exp, nil

	case ast.LPAREN:
		p.next()
		exp, err := p.parseValueExpression()
		if err != nil {
			return nil, err
		}

		if t := p.next(); t.Type != ast.RPAREN {
			return nil, NewError(t.Pos, "Expecting ), got %s", t.Str)
		}

		// check if there is an expression after the parenthesis like: (foo).bar
		if exp, err = p.parseValueExpr(exp); err != nil {
			return nil, err
		}

		return exp, nil

	case ast.TRUE, ast.FALSE:
		p.next()
		return &ast.ConstantExpr{t.Pos, t.Type, t.Str}, nil

	case ast.LBRACK:
		return p.parseIndexDeclExpr()

	case ast.LBRACE:
		return p.parseMapExpr()

	case ast.TYPEOF:
		return p.parseTypeof()

	default:
		return nil, NewError(t.Pos, "Expecting expression, got %s", t.Str)
	}
}

func (p *parser) peek() *ast.Token {
	t, _ := p.peekToken(0, false)
	return t
}

// peek two positions forward
func (p *parser) peekTwo() *ast.Token {
	t, _ := p.peekToken(1, false)
	return t
}

// peek three positions forward (needed for the "for x := range ...")
func (p *parser) peekThree() *ast.Token {
	t, _ := p.peekToken(2, false)
	return t
}

func (p *parser) next() *ast.Token {
	t, i := p.peekToken(0, false)
	p.index += i
	return t
}

func (p *parser) ignore(k ast.Type, count int) {
	for i := 0; i < count; i++ {
		if p.peek().Type == k {
			p.next()
		}
	}
}

func (p *parser) accept(k ast.Type) (*ast.Token, error) {
	t := p.next()
	if t.Type != k {
		return nil, NewError(t.Pos, "Expecting %v got %v", k, t.Type)
	}
	return t, nil
}

// peekToken returns the nth token skipping comments if requested.
func (p *parser) peekToken(count int, comments bool) (*ast.Token, int) {
	var i, n int
	l := len(p.tokens)
	for {
		pi := p.index + i
		if pi >= l {
			return &ast.Token{Type: ast.EOF}, -1
		}
		t := p.tokens[pi]
		if !comments && isComment(t) {
			i++
			continue
		}
		i++
		if n >= count {
			return t, i
		}
		n++
	}
}

func isComment(t *ast.Token) bool {
	switch t.Type {
	case ast.COMMENT, ast.MULTILINE_COMMENT:
		return true
	}
	return false
}

type Error struct {
	Pos     ast.Position
	message string
}

func (e Error) Message() string {
	return e.message
}

func (e Error) Position() ast.Position {
	return e.Pos
}

func (e Error) Error() string {
	return fmt.Sprintf("%s\n -> %v", e.message, e.Position())
}

func NewError(p ast.Position, format string, args ...interface{}) Error {
	return Error{p, fmt.Sprintf(format, args...)}
}

// Make a hash of all the sources.
func Hash(fs filesystem.FS, path string) ([]byte, error) {
	p := newParser(fs)
	m, err := p.Parse(path)
	if err != nil {
		return nil, err
	}

	hash := md5.New()

	var files []string
	files = append(files, m.File.Path)
	for _, f := range m.Modules {
		files = append(files, f.Path)
	}

	sort.Strings(files)

	for _, f := range files {
		addModuleHash(fs, f, hash)
	}

	return hash.Sum(nil), nil
}

func addModuleHash(fs filesystem.FS, path string, h hash.Hash) error {
	data, err := filesystem.ReadAll(fs, path)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", path, err)
	}
	h.Write(data)
	return nil
}
