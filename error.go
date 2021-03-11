package dune

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

func NewPublicError(msg string) *Error {
	return &Error{message: msg, public: true}
}

func Wrap(msg string, err error) error {
	e, ok := err.(*Error)
	if !ok {
		return fmt.Errorf("%s: %w", msg, err)
	}

	w := &Error{message: msg, public: e.public}
	w.Wrap(e)
	return w
}

type Error struct {
	pc          int
	message     string
	public      bool
	instruction *Instruction
	stacktrace  []TraceLine
	goError     error
	Wrapped     []*Error
	IsRethrow   bool
}

func (e *Error) Type() string {
	return "Error"
}

func (e *Error) Public() bool {
	return e.public
}

func (e *Error) String() string {
	return e.Error()
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) Wrap(inner *Error) {
	if e.public && inner.public {
		e.message += ": " + inner.message
	}

	e.Wrapped = append(e.Wrapped, inner)
}

func (e *Error) Error() string {
	var b = &bytes.Buffer{}

	b.WriteString(e.message)

	if len(e.stacktrace) > 0 {
		b.WriteRune('\n')
		for _, s := range e.stacktrace {
			if s.Function == "" || s.Line == 0 {
				continue // this is an empty position
			}
			fmt.Fprintf(b, " -> %s\n", s.String())
		}
	}

	for _, inner := range e.Wrapped {
		fmt.Fprintf(b, "\n%s\n", inner.Error())
	}

	return b.String()
}

func (e *Error) Is(msg string) bool {
	if e.message == msg {
		return true
	}

	if goErrorIs(e.goError, msg) {
		return true
	}

	for _, wrap := range e.Wrapped {
		if wrap.message == msg {
			return true
		}
		if goErrorIs(wrap.goError, msg) {
			return true
		}
	}
	return false
}

func (e *Error) Stack() string {
	var b = &bytes.Buffer{}

	for _, s := range e.stacktrace {
		if s.Function == "" && s.File == "" && s.Line == 0 {
			continue // this is an empty position
		}
		fmt.Fprintf(b, " -> %s\n", s.String())
	}

	return b.String()
}

func (e *Error) stackLines() []string {
	lines := make([]string, len(e.stacktrace))

	for _, s := range e.stacktrace {
		if s.Function == "" && s.File == "" && s.Line == 0 {
			continue // this is an empty position
		}
		lines = append(lines, s.String())
	}

	return lines
}

func (e *Error) GetProperty(name string, vm *VM) (Value, error) {
	switch name {
	case "public":
		return NewBool(e.public), nil
	case "message":
		return NewString(e.message), nil
	case "pc":
		return NewInt(e.pc), nil
	case "stackTrace":
		return NewString(e.Stack()), nil
	}

	return UndefinedValue, nil
}

func (e *Error) GetMethod(name string) NativeMethod {
	switch name {
	case "is":
		return e.is
	case "toString":
		return e.toString
	}
	return nil
}

func (e *Error) is(args []Value, vm *VM) (Value, error) {
	if len(args) != 1 {
		return NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	if args[0].Type != String {
		return NullValue, fmt.Errorf("expected a string, got %s", args[0].TypeName())
	}
	v := e.Is(args[0].ToString())
	return NewBool(v), nil
}

func (e *Error) toString(args []Value, vm *VM) (Value, error) {
	return NewString(e.Error()), nil
}

func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Message    string
		StackTrace []string
	}{
		Message:    e.message,
		StackTrace: e.stackLines(),
	})
}

func goErrorIs(err error, msg string) bool {
	if err == nil {
		return false
	}
	for {
		if err.Error() == msg {
			return true
		}
		err = errors.Unwrap(err)
		if err == nil {
			return false
		}
	}
}

func Stacktrace() string {
	c := callers()
	return stacktrace(c)
}

func stacktrace(stack *stack) string {
	var buf bytes.Buffer

	for _, f := range stack.StackTrace() {
		pc := f.pc()
		fn := runtime.FuncForPC(pc)

		if strings.HasPrefix(fn.Name(), "dune.") {
			// ignore Go src
			continue
		}

		file, _ := fn.FileLine(pc)

		buf.WriteString(" -> ")
		buf.WriteString(file)
		buf.WriteRune(':')
		buf.WriteString(strconv.Itoa(f.line()))
		buf.WriteRune('\n')
	}

	return buf.String()
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(4, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

// Frame represents a program counter inside a stack frame.
type Frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f Frame) pc() uintptr { return uintptr(f) - 1 }

// StackTrace is stack of Frames from innermost (newest) to outermost (oldest).
type StackTrace []Frame

// stack represents a stack of program counters.
type stack []uintptr

func (s *stack) StackTrace() StackTrace {
	f := make([]Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = Frame((*s)[i])
	}
	return f
}

// line returns the line number of source code of the
// function for this Frame's pc.
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}
