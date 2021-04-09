/*
Package templates generates source code for templates.

 - Code must go into <% and %>
 - Expressions can appear into <%= and %>
 - Unescaped output can appear whith <%== and %>
 - Functions can be included with a header tag "<%@" at the beginning of the template
*/
package templates

import (
	"bytes"
	"strings"
)

// Compile returns the generated renderer code and a sourcemap slice
// in which each index contains the original line.
func Compile(source string) ([]byte, []int, error) {
	return compile(source, false)
}

// CompileHtml returns the generated renderer code and a sourcemap slice
// in which each index contains the original line.
// htmlEncode makes <%= %> blocks html encoded and <%== %> not encoded.
func CompileHtml(source string) ([]byte, []int, error) {
	return compile(source, true)
}

// Compile returns the generated renderer code and a sourcemap slice
// in which each index contains the original line.
// htmlEncode makes <%= %> blocks html encoded and <%== %> not encoded.
func compile(source string, htmlEncode bool) ([]byte, []int, error) {
	l := newLexer(strings.NewReader(source))
	if err := l.Run(); err != nil {
		return nil, nil, err
	}

	const WRITE = "w.write("
	const CLOSE = ")"

	var escapedOpen string
	var escapedClose string

	if htmlEncode {
		escapedOpen = "w.write(html.encode("
		escapedClose = "))"
	} else {
		escapedOpen = WRITE
		escapedClose = CLOSE
	}

	var buf bytes.Buffer

	sourceMap := make([]int, 0, 500)

	var line int

	var headers bool

	// write headers
	for i, k := 0, len(l.tokens); i < k; i++ {
		t := l.tokens[i]
		switch t.kind {
		case text:
			if strings.Trim(t.str, " \n\r\t") == "" {
				continue
			}

		case code:
			if strings.HasPrefix(t.str, "@") {
				buf.WriteString(strings.Trim(t.str[1:], " \n\r\t"))
				buf.WriteRune('\n')
				l.tokens = append(l.tokens[:i], l.tokens[i+1:]...)
				k--
				headers = true
				continue
			}
		}

		break
	}

	if headers {
		buf.WriteString("\n")
	}

	buf.WriteString("function render(w: io.Writer, model: any) {\n")
	line++

	var lastKind kind

	// generate the code after imports
	for i, k := 0, len(l.tokens); i < k; i++ {
		t := l.tokens[i]
		breaks := strings.Count(t.str, "\n")

		switch t.kind {
		case text:
			str := t.str
			if lastKind == code && strings.HasPrefix(str, "\n") {
				str = str[1:]
			}

			if t.str == "`" {
				buf.WriteString(WRITE)
				buf.WriteString("\"`\"")
				buf.WriteString(CLOSE)
				buf.WriteString("\n")
			} else {
				buf.WriteString(WRITE)
				buf.WriteString("`")
				buf.WriteString(escape(str))
				buf.WriteString("`")
				buf.WriteString(CLOSE)
				buf.WriteString("\n")
			}

			last := line + breaks
			for i := line; i <= last; i++ {
				sourceMap = append(sourceMap, t.pos.line)
			}
			line = last + 1

		case expression:
			buf.WriteString(escapedOpen)
			buf.WriteString(escapeExpr(t.str))
			buf.WriteString(escapedClose)
			buf.WriteString("\n")

			last := line + breaks
			for i := line; i <= last; i++ {
				sourceMap = append(sourceMap, t.pos.line)
			}
			line = last + 1

		case unescapedExp:
			buf.WriteString(WRITE)
			buf.WriteString(escapeExpr(t.str))
			buf.WriteString(CLOSE)
			buf.WriteString("\n")

			last := line + breaks
			for i := line; i <= last; i++ {
				sourceMap = append(sourceMap, t.pos.line)
			}
			line = last + 1

		case code:
			str := escapeCode(t.str)
			if str != "" {
				buf.WriteString(str)
				buf.WriteString("\n")

				last := line + breaks
				for i := line; i <= last; i++ {
					sourceMap = append(sourceMap, t.pos.line+i-line)
				}
				line = last + 1
			}
		}

		lastKind = t.kind
	}

	buf.WriteString("\n}\n")

	return buf.Bytes(), sourceMap, nil
}

func escape(s string) string {
	return s
}

func escapeExpr(s string) string {
	return strings.Trim(s, " \t")
}

func escapeCode(s string) string {
	return strings.Trim(s, " \t")
}
