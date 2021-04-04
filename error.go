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

func NewTypeError(errorType, msg string, args ...interface{}) *VMError {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &VMError{ErrorType: errorType, Message: msg}
}

func Wrap(msg string, err error) error {
	e, ok := err.(*VMError)
	if !ok {
		return fmt.Errorf("%s: %w", msg, err)
	}

	w := &VMError{
		Message: msg,
		Wrapped: e,
	}

	return w
}

type VMError struct {
	Message     string
	ErrorType   string
	TraceLines  []TraceLine
	Wrapped     *VMError
	IsRethrow   bool
	pc          int
	instruction *Instruction
	goError     error
}

func (e *VMError) Type() string {
	return "Error"
}

func (e *VMError) String() string {
	return e.Error()
}

func (e *VMError) ErrorMessage() string {
	return e.Message
}

func (e *VMError) Error() string {
	var b = &bytes.Buffer{}

	b.WriteString(e.Message)

	if len(e.TraceLines) > 0 {
		b.WriteRune('\n')
		for _, s := range e.TraceLines {
			if s.Function == "" || s.Line == 0 {
				continue // this is an empty position
			}
			fmt.Fprintf(b, " -> %s\n", s.String())
		}
	}

	return b.String()
}

func (e *VMError) Is(msg string) bool {
	if e.ErrorType == msg {
		return true
	}

	if goErrorIs(e.goError, msg) {
		return true
	}

	wrap := e.Wrapped
	for wrap != nil {
		if wrap.ErrorType == msg {
			return true
		}
		if goErrorIs(wrap.goError, msg) {
			return true
		}
		wrap = wrap.Wrapped
	}
	return false
}

func (e *VMError) Stack() string {
	var b = &bytes.Buffer{}

	for _, s := range e.TraceLines {
		if s.Function == "" && s.File == "" && s.Line == 0 {
			continue // this is an empty position
		}
		fmt.Fprintf(b, " -> %s\n", s.String())
	}

	return b.String()
}

func (e *VMError) stackLines() []string {
	lines := make([]string, len(e.TraceLines))

	for _, s := range e.TraceLines {
		if s.Function == "" && s.File == "" && s.Line == 0 {
			continue // this is an empty position
		}
		lines = append(lines, s.String())
	}

	return lines
}

func (e *VMError) GetProperty(name string, vm *VM) (Value, error) {
	switch name {
	case "type":
		return NewString(e.ErrorType), nil
	case "message":
		return NewString(e.Message), nil
	case "pc":
		return NewInt(e.pc), nil
	case "stackTrace":
		return NewString(e.Stack()), nil
	}
	return UndefinedValue, nil
}

func (e *VMError) GetMethod(name string) NativeMethod {
	switch name {
	case "is":
		return e.is
	case "string":
		return e.string
	}
	return nil
}

func (e *VMError) is(args []Value, vm *VM) (Value, error) {
	if len(args) != 1 {
		return NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	if args[0].Type != String {
		return NullValue, fmt.Errorf("expected a string, got %s", args[0].TypeName())
	}
	v := e.Is(args[0].String())
	return NewBool(v), nil
}

func (e *VMError) string(args []Value, vm *VM) (Value, error) {
	return NewString(e.Error()), nil
}

func (e *VMError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ErrorType  string
		Message    string
		TraceLines []TraceLine
	}{
		ErrorType:  e.ErrorType,
		Message:    e.Message,
		TraceLines: e.TraceLines,
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
