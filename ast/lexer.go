//go:generate stringer -type=Type

package ast

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var reservedWords = map[string]Type{
	"if":        IF,
	"else":      ELSE,
	"for":       FOR,
	"while":     WHILE,
	"break":     BREAK,
	"continue":  CONTINUE,
	"return":    RETURN,
	"true":      TRUE,
	"false":     FALSE,
	"import":    IMPORT,
	"export":    EXPORT,
	"function":  FUNCTION,
	"interface": INTERFACE,
	"var":       VAR,
	"let":       LET,
	"const":     CONST,
	"enum":      ENUM,
	"switch":    SWITCH,
	"case":      CASE,
	"default":   DEFAULT,
	"null":      NULL,
	"undefined": UNDEFINED,
	"try":       TRY,
	"catch":     CATCH,
	"throw":     THROW,
	"finally":   FINALLY,
	"new":       NEW,
	"class":     CLASS,
	"delete":    DELETE,
	"typeof":    TYPEOF,
}

type Token struct {
	Type Type
	Str  string
	Pos  Position
}

func (t Token) String() string {
	return fmt.Sprintf("{%v: %v}", t.Type, t.Str)
}

type Type byte

const (
	ERROR Type = iota
	EOF
	COMMENT
	MULTILINE_COMMENT
	ATTRIBUTE

	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)
	IDENT  // main
	INT    // 12345
	HEX    // 0xFF
	FLOAT  // 123.45
	RUNE   // 'a'
	STRING // "abc"

	// Operators and delimiters
	ADD // +
	SUB // -
	MUL // *
	DIV // /
	MOD // %

	AND // &
	BOR // | binary or
	XOR // ^
	LSH // << left shift
	RSH // >> right shift
	BNT // ~ bitwise not

	QUESTION // ?

	ADD_ASSIGN // +=
	SUB_ASSIGN // -=
	MUL_ASSIGN // *=
	DIV_ASSIGN // /=
	XOR_ASSIGN // ^=
	BOR_ASSIGN // |=
	MOD_ASSIGN // %=

	LAND // &&
	LOR  // ||
	NOR  // ??
	INC  // ++
	DEC  // --
	EXP  // **

	EQL    // ==
	SEQ    // ===
	NEQ    // !=
	SNE    // !==
	LSS    // <
	GTR    // >
	ASSIGN // =
	NOT    // !

	LEQ // <=
	GEQ // >=

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :

	DECL   // :=
	LAMBDA // =>

	// Keywords
	BREAK
	CONTINUE
	IF
	ELSE
	FOR
	WHILE
	RETURN
	IMPORT
	SWITCH
	CASE
	DEFAULT

	LET
	VAR
	CONST
	FUNCTION
	ENUM
	NULL
	UNDEFINED
	INTERFACE
	EXPORT

	NEW
	CLASS

	TRUE
	FALSE

	TRY
	CATCH
	FINALLY
	THROW

	TYPEOF
	DELETE
)

const (
	eof = byte(EOF)
)

type Position struct {
	FileName string
	Line     int
	Column   int
}

func (p Position) String() string {
	var buf bytes.Buffer

	switch p.FileName {
	case "", ".":
	default:
		buf.WriteString(p.FileName)
	}

	fmt.Fprintf(&buf, ":%d", p.Line) // show in base 1

	//	if p.Column != -1 {
	//		fmt.Fprintf(&buf, " column %d", p.Column)
	//	}

	return buf.String()
}

type NodeError interface {
	Message() string
	Position() Position
	Error() string
}

type Lexer struct {
	Pos    Position
	reader *bufio.Reader
	Tokens []*Token
}

func New(reader io.Reader, fileName string) *Lexer {
	return &Lexer{
		Pos:    Position{FileName: fileName, Column: 0, Line: 1},
		reader: bufio.NewReaderSize(reader, 4096),
	}
}

type lexError struct {
	Pos     Position
	Message string
	Token   string
}

func (e *lexError) Position() Position {
	return e.Pos
}

func (e *lexError) Error() string {
	return e.Message
}

func (l *Lexer) error(tok string, msg string) *lexError {
	return &lexError{l.Pos, msg, tok}
}

func (l *Lexer) Run() error {

	for {
		token := &Token{}

		l.skipWhiteSpace()
		c := l.next()
		if c == byte(EOF) {
			return nil
		}

		var buf bytes.Buffer

		switch {
		case isIdent(c, 0):
			token.Type = IDENT
			err := l.readIdent(c, &buf)
			token.Str = buf.String()
			if err != nil {
				return err
			}

			if typ, ok := reservedWords[token.Str]; ok {
				token.Type = typ
			}

		case isDecimal(c):
			if err := l.readNumber(c, &buf, token); err != nil {
				return err
			}

		default:
			switch c {
			case '\'':
				err := l.readString('\'', &buf)
				b := buf.String()
				token.Str = b
				if len(b) == 1 {
					token.Type = RUNE
				} else {
					token.Type = STRING
				}
				if err != nil {
					return err
				}
			case '\\':
				token.Type = RUNE
				token.Str = "\\"
				l.next()
			case '"':
				token.Type = STRING
				err := l.readString(c, &buf)
				token.Str = buf.String()
				if err != nil {
					return err
				}
			case '`':
				token.Type = STRING
				err := l.readMultilineString(c, &buf)
				token.Str = buf.String()
				if err != nil {
					return err
				}
			case '+':
				if l.peek() == '+' {
					token.Type = INC
					token.Str = "++"
					l.next()
				} else if l.peek() == '=' {
					token.Type = ADD_ASSIGN
					token.Str = "+="
					l.next()
				} else {
					token.Type = ADD
					token.Str = string(c)
				}
			case '-':
				if l.peek() == '-' {
					token.Type = DEC
					token.Str = "--"
					l.next()
				} else if l.peek() == '=' {
					token.Type = SUB_ASSIGN
					token.Str = "-="
					l.next()
				} else {
					token.Type = SUB
					token.Str = string(c)
				}
			case '*':
				switch l.peek() {
				case '=':
					token.Type = MUL_ASSIGN
					token.Str = "*="
					l.next()
				case '*':
					token.Type = EXP
					token.Str = "**"
					l.next()
				default:
					token.Type = MUL
					token.Str = string(c)
				}
			case '^':
				if l.peek() == '=' {
					token.Type = XOR_ASSIGN
					token.Str = "^="
					l.next()
				} else {
					token.Type = XOR
					token.Str = string(c)
				}
			case '%':
				if l.peek() == '=' {
					token.Type = MOD_ASSIGN
					token.Str = "%="
					l.next()
				} else {
					token.Type = MOD
					token.Str = string(c)
				}
			case '/':
				switch l.peek() {
				case '=':
					token.Type = DIV_ASSIGN
					token.Str = "/="
					l.next()
				case '/':
					err := l.readComment(&buf)
					str := buf.String()

					trimmed := strings.TrimLeft(str, " \t")
					if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
						token.Type = ATTRIBUTE
						str = trimmed[1 : len(trimmed)-1]
						str = strings.TrimLeft(str, " \t")
					} else {
						token.Type = COMMENT
					}
					token.Str = str
					if err != nil {
						return err
					}
				case '*':
					err := l.readMultilineComment(&buf)
					token.Type = MULTILINE_COMMENT
					token.Str = buf.String()
					if err != nil {
						return err
					}
				default:
					token.Type = DIV
					token.Str = string(c)
				}
			case '=':
				if l.peek() == '=' {
					l.next()
					if l.peek() == '=' {
						l.next()
						token.Type = SEQ
						token.Str = "==="
					} else {
						token.Type = EQL
						token.Str = "=="
					}
				} else if l.peek() == '>' {
					token.Type = LAMBDA
					token.Str = "=>"
					l.next()
				} else {
					token.Type = ASSIGN
					token.Str = string(c)
				}
			case '<':
				if l.peek() == '=' {
					token.Type = LEQ
					token.Str = "<="
					l.next()
				} else if l.peek() == '<' {
					token.Type = LSH
					token.Str = string(c)
					l.next()
				} else {
					token.Type = LSS
					token.Str = string(c)
				}
			case '>':
				if l.peek() == '=' {
					token.Type = GEQ
					token.Str = ">="
					l.next()
				} else if l.peek() == '>' {
					token.Type = RSH
					token.Str = string(c)
					l.next()
				} else {
					token.Type = GTR
					token.Str = string(c)
				}
			case '!':
				if l.peek() == '=' {
					l.next()
					if l.peek() == '=' {
						l.next()
						token.Type = SNE
						token.Str = "!=="
					} else {
						token.Type = NEQ
						token.Str = "!="
					}
				} else {
					token.Type = NOT
					token.Str = string(c)
				}
			case '&':
				if l.peek() == '&' {
					token.Type = LAND
					token.Str = "&&"
					l.next()
				} else {
					token.Type = AND
					token.Str = string(c)
				}
			case '|':
				if l.peek() == '|' {
					token.Type = LOR
					token.Str = "||"
					l.next()
				} else if l.peek() == '=' {
					token.Type = BOR_ASSIGN
					token.Str = string(c)
					l.next()
				} else {
					token.Type = BOR
					token.Str = string(c)
				}
			case '?':
				if l.peek() == '?' {
					token.Type = NOR
					token.Str = "??"
					l.next()
				} else {
					token.Type = QUESTION
					token.Str = string(c)
				}
			case '~':
				token.Type = BNT
				token.Str = string(c)
			case '(':
				token.Type = LPAREN
				token.Str = string(c)
			case ')':
				token.Type = RPAREN
				token.Str = string(c)
			case '{':
				token.Type = LBRACE
				token.Str = string(c)
			case '}':
				token.Type = RBRACE
				token.Str = string(c)
			case '[':
				token.Type = LBRACK
				token.Str = string(c)
			case ']':
				token.Type = RBRACK
				token.Str = string(c)
			case ',':
				token.Type = COMMA
				token.Str = string(c)
			case '.':
				token.Type = PERIOD
				token.Str = string(c)
			case ';':
				token.Type = SEMICOLON
				token.Str = string(c)
			case ':':
				if l.peek() == '=' {
					token.Type = DECL
					token.Str = ":="
					l.next()
				} else {
					token.Type = COLON
					token.Str = string(c)
				}
			}
		}

		l.addToken(token)
	}
}

func (l *Lexer) readNumber(c byte, buf *bytes.Buffer, token *Token) error {
	if c == '0' && l.peek() == 'x' {
		if err := l.readHexadecimal(buf); err != nil {
			return l.error(buf.String(), "Invalid hex number")
		}
		token.Type = HEX
		token.Str = buf.String()
		return nil
	}

	token.Type = INT
	err := l.readDecimal(c, buf)
	token.Str = buf.String()
	if err != nil {
		return err
	}
	c = l.peek()
	if c == '.' {
		buf.WriteByte(c)
		l.next()
		c = l.next()
		if !isDecimal(c) {
			return l.error(buf.String(), "Invalid number")
		}
		token.Type = FLOAT
		err = l.readDecimal(c, buf)
		token.Str = buf.String()
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Lexer) addToken(t *Token) {
	t.Pos = l.Pos
	t.Pos.Column-- // base 0 for consistency with line nums
	l.Tokens = append(l.Tokens, t)
}

func (l *Lexer) readMultilineString(quote byte, b *bytes.Buffer) error {
	c := l.next()
	for c != quote {
		if c == byte(EOF) {
			return l.error(b.String(), "unterminated multiline string")
		}
		b.WriteByte(c)
		c = l.next()
	}
	return nil
}

func (l *Lexer) readString(quote byte, b *bytes.Buffer) error {
	c := l.next()
	for c != quote {
		if c == '\n' || c == '\r' || c == byte(EOF) {
			return l.error(b.String(), "unterminated string")
		}

		if c == '\\' {
			c = l.next()
			switch c {
			case 't':
				b.WriteByte('\t')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 'x':
				// lex as hex (\xnn)
				v := string(l.next()) + string(l.next())
				i, err := strconv.ParseInt(v, 16, 64)
				if err != nil {
					return l.error(b.String(), "Invalid hex value")
				}
				b.WriteByte(byte(i))
				c = l.next()
				continue
			case '\'':
				b.WriteByte('\'')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				if isDecimal(c) {
					// lex as three-digit octal (\nnn)
					v := string(c) + string(l.next()) + string(l.next())
					i, err := strconv.ParseInt(v, 8, 64)
					if err != nil {
						return l.error(b.String(), "Invalid octal value")
					}
					b.WriteByte(byte(i))
					c = l.next()
					continue
				}

				b.WriteByte('\\')
				b.WriteByte(c)
			}

			c = l.next()
			continue
		}

		b.WriteByte(c)
		c = l.next()
	}
	return nil
}

func (l *Lexer) readComment(b *bytes.Buffer) error {
	l.next()

loop:
	for {
		c := l.peek()
		switch c {
		case '\n', eof:
			break loop
		}
		b.WriteByte(c)
		l.next()
	}

	return nil
}

func (l *Lexer) readMultilineComment(b *bytes.Buffer) error {
	l.next()

loop:
	for {
		c := l.next()
		switch c {
		case eof:
			break loop
		case '*':
			if l.peek() == '/' {
				l.next()
				break loop
			}
		}
		b.WriteByte(c)
	}

	return nil
}

func (l *Lexer) readIdent(c byte, b *bytes.Buffer) error {
	b.WriteByte(c)
	for isIdent(l.peek(), 1) {
		b.WriteByte(l.next())
	}
	return nil
}

func (l *Lexer) readHexadecimal(b *bytes.Buffer) error {
	b.WriteByte('0')
	l.next()
	b.WriteByte('x')

	var c byte

	for {
		c = l.peek()
		if !isHex(c) {
			break
		}
		b.WriteByte(l.next())
	}

	if 'G' <= c && c <= 'Z' || 'g' <= c && c <= 'z' {
		return l.error(b.String(), "Invalid hex number")
	}

	return nil
}

func (l *Lexer) readDecimal(c byte, b *bytes.Buffer) error {
	b.WriteByte(c)
	for {
		p := l.peek()

		if p == '_' {
			// ignore as separator
			l.next()
			continue
		}

		if isDecimal(p) {
			b.WriteByte(l.next())
			continue
		}

		break
	}
	return nil
}

func (l *Lexer) readNext() byte {
	ch, err := l.reader.ReadByte()
	if err == io.EOF {
		return byte(EOF)
	}
	return ch
}

func (l *Lexer) peek() byte {
	ch := l.readNext()
	if ch != byte(EOF) {
		l.reader.UnreadByte()
	}
	return ch
}

func (l *Lexer) next() byte {
	ch := l.readNext()
	switch ch {
	case '\n', '\r':
		l.newline(ch)
		ch = '\n'
	case byte(EOF):
	default:
		l.Pos.Column++
	}
	return ch
}

func (l *Lexer) newline(ch byte) {
	l.Pos.Line++
	l.Pos.Column = 0
	next := l.peek()
	if ch == '\n' && next == '\r' || ch == '\r' && next == '\n' {
		l.reader.ReadByte()
	}
}

func (l *Lexer) skipWhiteSpace() {
loop:
	if isWhitespace(l.peek()) {
		l.next()
		goto loop
	}
}

func isWhitespace(ch byte) bool {
	switch ch {
	case '\t', '\r', ' ', '\n':
		return true
	}
	return false
}

func isDecimal(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHex(ch byte) bool {
	return 'A' <= ch && ch <= 'F' || 'a' <= ch && ch <= 'f' || isDecimal(ch)
}

func isIdent(ch byte, pos int) bool {
	return ch == '_' ||
		'A' <= ch && ch <= 'Z' ||
		'a' <= ch && ch <= 'z' ||
		isDecimal(ch) && pos > 0
}
