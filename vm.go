package dune

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dunelang/dune/ast"
	"github.com/dunelang/dune/filesystem"
	"github.com/dunelang/dune/parser"
)

const VERSION = "0.96"

var ErrFunctionNotExist = errors.New("function not found")

func Run(fs filesystem.FS, path string) (Value, error) {
	p, err := Compile(fs, path)
	if err != nil {
		return NullValue, err
	}

	vm := NewVM(p)
	return vm.Run()
}

func RunStr(code string) (Value, error) {
	p, err := CompileStr(code)
	if err != nil {
		return NullValue, err
	}

	vm := NewVM(p)
	return vm.Run()
}

func Parse(fs filesystem.FS, path string) (*ast.Program, error) {
	return parser.Parse(fs, path)
}

func ParseStr(code string) (*ast.Program, error) {
	return parser.ParseStr(code)
}

func ParseExpr(code string) (ast.Expr, error) {
	return parser.ParseExpr(code)
}

func NewVM(p *Program) *VM {
	vm := &VM{Program: p}

	globalFrame := &stackFrame{
		funcIndex: 0,
		values:    make([]Value, p.Functions[0].MaxRegIndex),
	}

	vm.callStack = []*stackFrame{globalFrame}
	vm.allocations = int64(p.kSize)
	return vm
}

func NewInitializedVM(p *Program, globals []Value) *VM {
	vm := &VM{
		Program:     p,
		initialized: true,
	}

	if len(globals) != p.Functions[0].MaxRegIndex {
		panic("invalid globals size")
	}

	globalFrame := &stackFrame{
		funcIndex: 0,
		values:    globals,
	}

	vm.callStack = []*stackFrame{globalFrame}
	vm.allocations = int64(p.kSize)
	return vm
}

type VM struct {
	Program        *Program
	MaxSteps       int64
	MaxAllocations int64
	MaxFrames      int
	RetValue       Value
	Error          error
	Context        Value
	FileSystem     filesystem.FS
	Language       string
	Localizer      Localizer
	Location       *time.Location
	Now            time.Time
	Stdin          io.Reader
	Stdout         io.Writer
	Stderr         io.Writer

	fp           int
	steps        int64
	allocations  int64
	initialized  bool
	callStack    []*stackFrame
	tryCatchs    []*tryCatch
	optchainPC   *Address
	optchainDest *Address
	optchainSrc  *Address
	frameCache   []*stackFrame
}

func (vm *VM) GetStdin() io.Reader {
	if vm.Stdin != nil {
		return vm.Stdin
	}
	return os.Stdin
}

func (vm *VM) GetStdout() io.Writer {
	if vm.Stdout != nil {
		return vm.Stdout
	}
	return os.Stdout
}

func (vm *VM) GetStderr() io.Writer {
	if vm.Stderr != nil {
		return vm.Stderr
	}
	return os.Stderr
}

func (vm *VM) Allocations() int64 {
	return vm.allocations
}

func (vm *VM) Steps() int64 {
	return vm.steps
}

func (vm *VM) ResetSteps() {
	vm.steps = 0
}

func (vm *VM) AddSteps(n int64) error {
	if vm.MaxSteps > 0 {
		vm.steps += n

		// Go doesn't check overflows
		if vm.steps < 0 {
			return vm.NewError("Step limit overflow: %d", vm.steps)
		}

		if vm.steps > vm.MaxSteps {
			return vm.NewError("Step limit reached: %d", vm.MaxSteps)
		}
	}
	return nil
}

func (vm *VM) CurrentFunc() *Function {
	frame := vm.callStack[vm.fp]
	return vm.Program.Functions[frame.funcIndex]
}

func (vm *VM) HasPermission(name string) bool {
	if vm.Program.HasPermission(name) {
		return true
	}

	if vm.CurrentFunc().HasPermission(name) {
		return true
	}

	return false
}

func (vm *VM) Clone(p *Program) *VM {
	m := NewVM(p)
	m.MaxAllocations = vm.MaxAllocations
	m.MaxFrames = vm.MaxFrames
	m.MaxSteps = vm.MaxSteps
	m.FileSystem = vm.FileSystem
	m.Context = vm.Context
	m.Language = vm.Language
	m.Localizer = vm.Localizer
	m.Location = vm.Location
	m.Stdin = vm.Stdin
	m.Stdout = vm.Stdout
	m.Stderr = vm.Stderr
	m.Now = vm.Now
	return m
}

func (vm *VM) CloneInitialized(p *Program, globals []Value) *VM {
	m := NewInitializedVM(p, globals)
	m.MaxAllocations = vm.MaxAllocations
	m.MaxFrames = vm.MaxFrames
	m.MaxSteps = vm.MaxSteps
	m.FileSystem = vm.FileSystem
	m.Context = vm.Context
	m.Language = vm.Language
	m.Localizer = vm.Localizer
	m.Location = vm.Location
	m.Stdin = vm.Stdin
	m.Stdout = vm.Stdout
	m.Stderr = vm.Stderr
	m.Now = vm.Now
	return m
}

func (vm *VM) Initialized() bool {
	return vm.initialized
}

func (vm *VM) Initialize() error {
	vm.run(false)

	if vm.Error == io.EOF {
		vm.Error = nil
	}

	if vm.Error != nil {
		vm.cleanupFrame(0)
		return vm.Error
	}

	// at this point global values have been initialized and
	// RunFunc can be called
	vm.initialized = true

	// the global function ends with a return so restore the fp back
	vm.fp = 0

	return nil
}

func (vm *VM) Run(args ...Value) (Value, error) {
	// reset the error in case is reused
	vm.Error = nil

	if !vm.initialized {
		if err := vm.Initialize(); err != nil {
			return NullValue, err
		}
	}

	// if it has an entry point call it
	f, ok := vm.Program.Function("main")
	if !ok {
		// cleanup if there is nothing more to do
		vm.cleanupFrame(0)
		return vm.RetValue, nil
	}

	return vm.runFunc(f, false, UndefinedValue, true, nil, args...)
}

// RunFunc executes a function by name
func (vm *VM) RunFunc(name string, args ...Value) (Value, error) {
	f, ok := vm.Program.Function(name)
	if !ok {
		return NullValue, fmt.Errorf("%s: %w", name, ErrFunctionNotExist)
	}
	return vm.runFunc(f, false, UndefinedValue, false, nil, args...)
}

// RunFuncIndex executes a program function by index
func (vm *VM) RunFuncIndex(index int, args ...Value) (Value, error) {
	f := vm.Program.Functions[index]
	return vm.runFunc(f, false, UndefinedValue, false, nil, args...)
}

// RunClosure executes a program closure
func (vm *VM) RunClosure(c *Closure, args ...Value) (Value, error) {
	f := vm.Program.Functions[c.FuncIndex]
	return vm.runFunc(f, false, UndefinedValue, false, c.closures, args...)
}

// RunMethod executes a class method
func (vm *VM) RunMethod(c *Method, args ...Value) (Value, error) {
	f := vm.Program.Functions[c.FuncIndex]
	return vm.runFunc(f, true, c.ThisObject, false, nil, args...)
}

func (vm *VM) Globals() []Value {
	return vm.callStack[0].values
}

func (vm *VM) runFunc(f *Function, isMethod bool, this Value, finalizeGlobals bool, closures []*closureRegister, args ...Value) (Value, error) {
	if !isMethod && f.IsClass {
		return NullValue, fmt.Errorf("can't call a method directly")
	}

	if !f.Variadic && f.Arguments < len(args) {
		return NullValue, fmt.Errorf("function '%s' expects only %d parameters, got %d",
			f.Name, f.Arguments, len(args))
	}

	minArgs := f.Arguments - f.OptionalArguments
	if f.Variadic {
		minArgs--
	}

	if minArgs > len(args) {
		var errMsg string
		if f.OptionalArguments == 0 {
			errMsg = "function '%s' expects %d parameters, got %d"
		} else {
			errMsg = "function '%s' expects at least %d parameters, got %d"
		}
		return NullValue, fmt.Errorf(errMsg, f.Name, minArgs, len(args))
	}

	currentFp := vm.fp
	currentTryCatchs := vm.tryCatchs

	// reset for the call
	vm.tryCatchs = nil

	// store the last pc for the return
	currentFrame := vm.callStack[vm.fp]
	currentFrame.retAddress = Void

	// add a new frame
	frame := vm.addFrame(f)
	frame.funcIndex = f.Index
	frame.maxRegIndex = f.MaxRegIndex
	frame.exit = true
	frame.closures = closures

	lenArgs := len(args)
	locals := vm.callStack[vm.fp].values

	if f.Variadic {
		regularArgs := f.Arguments - 1
		// set the parameters that are not variadic
		if regularArgs > 0 {
			for i := 0; i < regularArgs; i++ {
				if i >= lenArgs {
					// skip if not enouth parameters have been provided
					break
				}
				v := args[i]
				locals[i] = v
				if err := vm.AddAllocations(v.Size()); err != nil {
					return NullValue, err
				}
			}
		}
		// set the variadic as an array with the rest of the parameters
		if lenArgs > regularArgs {
			v := NewArrayValues(args[regularArgs:])
			if err := vm.AddAllocations(v.Size()); err != nil {
				return NullValue, err
			}
			locals[regularArgs] = v
		} else {
			// if no arguments are passed set the variadic param as an empty array
			locals[regularArgs] = NewArray(0)
		}
	} else {
		for i := 0; i < f.Arguments; i++ {
			if i >= lenArgs {
				// skip if not enouth parameters have been provided
				break
			}
			v := args[i]
			if err := vm.AddAllocations(v.Size()); err != nil {
				return NullValue, err
			}
			locals[i] = v
		}
	}

	if isMethod {
		// this is always the next value after the arguments
		locals[f.Arguments] = this
	}

	vm.run(finalizeGlobals)

	// restore
	vm.tryCatchs = currentTryCatchs
	vm.fp = currentFp

	err := vm.Error
	vm.Error = nil

	if err != nil && err != io.EOF {
		return NullValue, err
	}

	return vm.RetValue, nil
}

func (vm *VM) addFrame(f *Function) *stackFrame {
	var frame *stackFrame

	if len(vm.frameCache) > 0 {
		frame = vm.frameCache[0]
		vm.frameCache[0] = nil
		vm.frameCache = vm.frameCache[1:]

		frame.retAddress = nil
		frame.exit = false
		frame.maxRegIndex = 0
		frame.pc = 0

		// expand if necesary
		ln := len(frame.values)
		if ln < f.MaxRegIndex {
			frame.values = append(frame.values, make([]Value, f.MaxRegIndex-ln)...)
		}
	} else {
		frame = &stackFrame{values: make([]Value, f.MaxRegIndex)}
	}

	vm.fp++
	vm.callStack = append(vm.callStack[:vm.fp], frame)
	return frame
}

// return a value from the current scope
func (vm *VM) RegisterValue(name string) (Value, bool) {
	// try the current frame
	if vm.fp > 0 {
		frame := vm.callStack[vm.fp]
		fn := vm.Program.Functions[frame.funcIndex]
		locals := frame.values
		for _, r := range fn.Registers {
			if r.Name == name {
				return locals[r.Index], true
			}
		}
	}

	// try globals
	fn := vm.Program.Functions[0]
	globals := vm.callStack[0].values
	for _, r := range fn.Registers {
		if r.Name == name {
			return globals[r.Index], true
		}
	}

	return NullValue, false
}

func (vm *VM) SetFinalizer(v Finalizable) {
	frame := vm.callStack[vm.fp]
	frame.finalizables = append(frame.finalizables, v)
}

func (vm *VM) SetGlobalFinalizer(v Finalizable) {
	frame := vm.callStack[0]
	frame.finalizables = append(frame.finalizables, v)
}

func (vm *VM) get(a *Address) Value {
	switch a.Kind {
	case AddrLocal:
		return vm.callStack[vm.fp].values[a.Value]
	case AddrFunc:
		return NewFunction(int(a.Value))
	case AddrConstant:
		return vm.Program.Constants[a.Value]
	case AddrGlobal:
		return vm.callStack[0].values[a.Value]
	case AddrNativeFunc:
		return NewNativeFunction(int(a.Value))
	case AddrEnum:
		return NewEnum(int(a.Value))
	case AddrData:
		return NewInt(int(a.Value))
	case AddrClosure:
		return vm.callStack[vm.fp].closures[a.Value].get()
	case AddrVoid:
		return NullValue
	case AddrUnresolved:
		panic(fmt.Sprintf("Unresolved address: %v", a))
	default:
		panic(fmt.Sprintf("Invalid address kind: %v", a))
	}
}

func (vm *VM) set(a *Address, v Value) {
	if err := vm.AddAllocations(v.Size()); err != nil {
		vm.Error = err
		return
	}

	switch a.Kind {
	case AddrLocal:
		vm.callStack[vm.fp].values[a.Value] = v
	case AddrClosure:
		vm.callStack[vm.fp].closures[a.Value].set(v)
	case AddrGlobal:
		vm.callStack[0].values[a.Value] = v
	case AddrConstant:
		panic(fmt.Sprintf("can't modify a constant: %v", a))
	default:
		panic(fmt.Sprintf("Invalid register address: %v", a))
	}
}

func (vm *VM) AddAllocations(size int) error {
	if vm.MaxAllocations == 0 {
		return nil
	}

	vm.allocations += int64(size)
	if vm.allocations > vm.MaxAllocations {
		return vm.NewError("Max allocations reached: %d", vm.MaxAllocations)
	}
	return nil
}

func (vm *VM) setPrototype(name string, this Value, dst *Address) bool {
	if m, ok := vm.getNativePrototype(name, this); ok {
		vm.set(dst, NewObject(m))
		return true
	}

	if m, ok := vm.getProgramPrototype(name, this); ok {
		vm.set(dst, NewObject(m))
		return true
	}
	return false
}

func (vm *VM) getNativePrototype(name string, this Value) (nativePrototype, bool) {
	f, ok := allNativeMap[name]
	if ok {
		return nativePrototype{this: this, fn: f.Index}, true
	}
	return nativePrototype{}, false
}

func (vm *VM) getProgramPrototype(name string, this Value) (*Method, bool) {
	p := vm.Program
	f, ok := p.Function(name)
	if !ok {
		return &Method{}, false
	}
	return &Method{FuncIndex: f.Index, ThisObject: this}, true
}

func (vm *VM) Stacktrace() []string {
	st := vm.stackTrace()
	s := make([]string, len(st))
	for i, l := range st {
		s[i] = l.String()
	}
	return s
}

func (vm *VM) stackTrace() []TraceLine {
	var trace []TraceLine

	p := vm.Program
	var lastLine TraceLine

	for i := vm.fp; i >= 0; i-- {
		frame := vm.callStack[i]
		f := p.Functions[frame.funcIndex]

		if f.IsGlobal && vm.initialized {
			// the global function has ended
			continue
		}

		var pc int
		if vm.fp == i || frame.pc == 0 {
			pc = frame.pc
		} else {
			pc = frame.pc - 1
		}

		line := p.ToTraceLine(f, pc)
		if line.SameLine(lastLine) {
			continue
		}

		trace = append(trace, line)
		lastLine = line
	}

	return trace
}

type ErrorMessenger interface {
	ErrorMessage() string
}

func (vm *VM) WrapError(err error) *VMError {
	var msg string
	var goError error

	switch t := err.(type) {
	case *VMError:
		if t.IsRethrow {
			t.IsRethrow = false
			return t
		}
		t.TraceLines = append(t.TraceLines, vm.stackTrace()...)
		return t
	case ErrorMessenger:
		msg = t.ErrorMessage()
	default:
		msg = t.Error()
		goError = t
	}

	return &VMError{
		Message:     msg,
		TraceLines:  vm.stackTrace(),
		instruction: vm.instruction(),
		goError:     goError,
	}
}

func (vm *VM) NewTypeError(errorType, format string, a ...interface{}) *VMError {
	err := vm.NewError(format, a...)
	err.ErrorType = errorType
	return err
}

func (vm *VM) NewError(format string, a ...interface{}) *VMError {
	st := vm.stackTrace()

	if len(a) > 0 {
		format = fmt.Sprintf(format, a...)
	}

	return &VMError{
		Message:     format,
		instruction: vm.instruction(),
		TraceLines:  st,
	}
}

func (vm *VM) returnFromFinally() bool {
	l := len(vm.tryCatchs)
	if l == 0 {
		return false
	}
	fp := vm.fp

	// loop trough all the nested finally's and execute them
	for i := l - 1; i >= 0; i-- {
		try := vm.tryCatchs[i]

		// only execute the finallys of its own frame. Other frames
		// will execute their own.
		if try.fp != fp {
			return false
		}

		finallyPC := try.finallyPC

		// if there is no finally
		if finallyPC == -1 {
			vm.tryCatchs = vm.tryCatchs[:l-1]
			continue
		}

		frame := vm.callStack[fp]

		// if we are already inside the finally and returning from it
		fi := frame.funcIndex
		f := vm.Program.Functions[fi]
		fnEndPC := len(f.Instructions)

		if frame.pc >= finallyPC && (frame.pc <= fnEndPC) {
			vm.tryCatchs = vm.tryCatchs[:l-1]
			continue
		}

		try.retPC = frame.pc
		vm.setPC(finallyPC)
		return true
	}

	return false
}

// returns true if the error is handled
func (vm *VM) handle(err error) bool {
	ln := len(vm.tryCatchs)
	if ln == 0 {
		vm.cleanupNotGlobalFrame(vm.fp)
		vm.Error = err
		return false
	}

	try := vm.tryCatchs[ln-1]

	// if its handled is an exception inside the catch
	if try.catchExecuted {
		// execute the finally even if the catch has thrown an exception
		if try.finallyPC != -1 && !try.finallyExecuted {
			try.err = err
			try.finallyExecuted = true
			vm.restoreStackframe(try)
			vm.setPC(try.finallyPC)
			return true
		}

		// jump to the parent catch if exists
		vm.tryCatchs = vm.tryCatchs[:ln-1]
		if ln > 1 {
			for i := ln - 2; i >= 0; i-- {
				try = vm.tryCatchs[i]
				if try.catchExecuted {
					// consume try-catchs that have thrown inside the catch
					continue
				}
				break
			}
			if try.catchExecuted {
				// if no catch found then it is unhandled
				vm.Error = err
				return false
			}
		} else {
			// or return an unhandled error.
			// An error thrown in the catch has precedence
			if try.err != nil {
				vm.Error = try.err
			} else {
				vm.Error = err
			}
			return false
		}
	}

	// mark this try as handled in case a exception is throw inside the
	// catch block to discard it.
	try.catchExecuted = true

	vm.Error = nil // handled

	jumpTo := try.catchPC
	if jumpTo == -1 {
		// if there is no catch block the err is unhandled
		try.err = err

		// if there is no catch block go directly to the finally block.
		jumpTo = try.finallyPC
		if jumpTo == -1 {
			// TODO: this could be catched by the compiler
			vm.Error = vm.NewError("try without catch or finally")
			return false
		}
	}

	// If jumps to finally directly the error is unhandled
	// check also that it doesn't have an empty catch
	if jumpTo == try.finallyPC {
		try.finallyExecuted = true
	}

	// restore the framepointer and local memory
	// where the try-catch is declared
	vm.restoreStackframe(try)

	// advance to the catch part
	vm.setPC(jumpTo)

	if try.errorReg != Void {
		e, ok := err.(*VMError)
		if !ok {
			e = &VMError{Message: err.Error()}
		}
		vm.set(try.errorReg, NewObject(e))
	}

	return true
}

// restore the framepointer and local memory
// where the try-catch is declared
func (vm *VM) restoreStackframe(try *tryCatch) {
	if vm.fp == try.fp {
		return
	}

	// clean up frames until the catch
	for i := vm.fp; i > try.fp; i-- {
		vm.cleanupNotGlobalFrame(i)
	}

	vm.fp = try.fp
}

func (vm *VM) runFinalizables(frame *stackFrame) {
	if frame.finalizables != nil {
		fzs := frame.finalizables
		frame.finalizables = nil

		for _, v := range fzs {
			if err := v.Close(); err != nil {
				// don't overwrite the main error
				if vm.Error == nil {
					// set the error but continue running all the finalizers
					vm.Error = err
				}
			}
		}
	}
}

func (vm *VM) cleanupNotGlobalFrame(index int) {
	if index == 0 {
		return
	}
	frame := vm.callStack[index]
	vm.runFinalizables(frame)
}

func (vm *VM) cleanupFrame(index int) {
	frame := vm.callStack[index]
	vm.runFinalizables(frame)
}

func (vm *VM) FinalizeGlobals() {
	vm.cleanupFrame(0)
}

func (vm *VM) run(finalizeGlobals bool) {
	defer func() {
		if r := recover(); r != nil {
			vm.recover(r)
		}
	}()

	if finalizeGlobals {
		defer func() {
			vm.runFinalizables(vm.callStack[0])
		}()
	}

	p := vm.Program
	// Print(p)

	for {
		if vm.MaxSteps > 0 {
			vm.steps++
			if vm.steps > vm.MaxSteps {
				vm.Error = vm.NewError("Step limit reached: %d", vm.MaxSteps)
				return
			}
		}

		frame := vm.callStack[vm.fp]
		f := p.Functions[frame.funcIndex]
		i := f.Instructions[frame.pc]

		// Print step
		// i := frame.funcIndex
		// fmt.Println("->", fmt.Sprintf("FN %-2d", i), fmt.Sprintf("PC %-6d", frame.pc), instr, "  "+f.Name)

		r := exec(i, vm)

		switch r {
		case vm_next:
			if vm.Error != nil {
				return
			}
			frame.pc++
			continue

		case vm_continue:
			continue

		case vm_exit:
			if vm.Error != nil {
				// if it is an unhandled execption execute all finalizables.
				// the global frame is called in the defer
				for i := vm.fp; i > 0; i-- {
					vm.cleanupFrame(i)
				}
			}
			return
		}
	}
}

func (vm *VM) recover(recoverValue interface{}) {
	frame := vm.callStack[vm.fp]

	msg := fmt.Sprintf("PANIC: [%d] %s", frame.pc, recoverValue)

	pt := "\n" + strings.Join(vm.Stacktrace(), "\n")
	pt = strings.Replace(pt, "\n", "\n [TS] -> ", -1)
	msg += pt

	st := strings.Replace("\n"+Stacktrace(), "\n ", "\n [Go] ", -1)
	msg += st

	var instr *Instruction
	i := frame.funcIndex
	f := vm.Program.Functions[i]
	if frame.pc < len(f.Instructions) {
		instr = f.Instructions[frame.pc]
	}

	// don't call newError because it's handling itself the stack trace.
	vm.Error = &VMError{
		Message:     msg,
		instruction: instr,
	}
}

func (vm *VM) setPC(pc int) {
	vm.callStack[vm.fp].pc = pc
}

func (vm *VM) incPC(steps int) {
	vm.callStack[vm.fp].pc += steps
}

func (vm *VM) call(a, b *Address, args []Value, optional bool) int {
	// TODO Handle variadic and spread with closures.
	// get the function
	var f *Function
	var closures []*closureRegister

	var isMethod bool
	var this Value

	switch a.Kind {
	case AddrFunc:
		f = vm.Program.Functions[a.Value]
	case AddrNativeFunc:
		if err := vm.callNativeFunc(int(a.Value), args, b, this); err != nil {
			if vm.handle(vm.WrapError(err)) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
		return vm_next
	default:
		value := vm.get(a)
		switch value.Type {
		case Func:
			f = vm.Program.Functions[value.ToFunction()]
		case Object:
			switch t := value.ToObject().(type) {
			case *Closure:
				f = vm.Program.Functions[t.FuncIndex]
				closures = t.closures
			case *Method:
				f = vm.Program.Functions[t.FuncIndex]
				isMethod = true
				this = t.ThisObject
			case nativePrototype:
				if err := vm.callNativeFunc(t.fn, args, b, t.this); err != nil {
					if vm.handle(vm.WrapError(err)) {
						return vm_continue
					} else {
						return vm_exit
					}
				}
				return vm_next
			case NativeMethod:
				if err := vm.callNativeMethod(t, args, b); err != nil {
					if vm.handle(vm.WrapError(err)) {
						return vm_continue
					} else {
						return vm_exit
					}
				}
				return vm_next
			default:
				if vm.handle((vm.NewError(fmt.Sprintf("Invalid value. Expected a function, got %v", value)))) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
		default:
			if optional {
				vm.closeOptChain()
				return vm_continue
			}
			if vm.handle((vm.NewError(fmt.Sprintf("Invalid value. Expected a function, got %v", value)))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	}

	return vm.callProgramFunc(f, b, args, isMethod, this, closures)
}

func (vm *VM) closeOptChain() {
	vm.incPC(int(vm.optchainPC.Value))

	if vm.optchainDest != nil && vm.optchainDest.Kind != AddrVoid {
		var v Value
		if vm.optchainSrc != nil && vm.optchainSrc.Kind != AddrVoid {
			v = vm.get(vm.optchainSrc)
		}
		vm.set(vm.optchainDest, v)
	}

	vm.optchainPC = nil
	vm.optchainDest = nil
	vm.optchainSrc = nil
}

func (vm *VM) callProgramFunc(f *Function, retAddr *Address, args []Value, isMethod bool, this Value, closures []*closureRegister) int {
	frame := vm.callStack[vm.fp]

	// set where to store the return value after the call in the current frame
	frame.retAddress = retAddr

	// add a new frame
	newFrame := vm.addFrame(f)
	newFrame.funcIndex = f.Index
	newFrame.maxRegIndex = f.MaxRegIndex
	newFrame.closures = closures

	if vm.MaxFrames > 0 && vm.fp > vm.MaxFrames {
		vm.Error = vm.NewError("Max stack frames reached: %d", vm.MaxFrames)
		return vm_exit
	}

	locals := newFrame.values

	// copy arguments
	if f.Arguments > 0 {
		count := len(args)

		if f.Variadic {
			regularArgs := f.Arguments - 1
			if count < regularArgs {
				copy(locals, args)
				// zero the rest of the args because memory can be reused
				for i := count; i < f.Arguments; i++ {
					locals[i] = NullValue
				}
			} else {
				for i := 0; i < regularArgs; i++ {
					locals[i] = args[i]
				}
				locals[regularArgs] = NewArrayValues(args[regularArgs:])
			}
		} else {
			if count > f.Arguments {
				// ignore if too many parameters are passed
				copy(locals, args[:f.Arguments])
			} else {
				copy(locals, args)
			}
		}
	}

	if isMethod {
		// this is always the next value after the arguments
		locals[f.Arguments] = this
	}

	return vm_next
}

func (vm *VM) callNativeFunc(i int, args []Value, retAddress *Address, this Value) error {
	f := allNativeFuncs[i]

	l := f.Arguments
	if l != -1 && l != len(args) {
		return fmt.Errorf("function '%s' expects %d parameters, got %d", f.Name, l, len(args))
	}

	for _, perm := range f.Permissions {
		if !vm.HasPermission(perm) {
			return errors.New("unauthorized")
		}
	}

	ret, err := f.Function(this, args, vm)
	if err != nil {
		return err
	}

	if retAddress != Void {
		vm.set(retAddress, ret)
	}

	return nil
}

func (vm *VM) callNativeMethod(m NativeMethod, args []Value, retAddress *Address) error {
	ret, err := m(args, vm)
	if err != nil {
		return err
	}

	if retAddress != Void {
		vm.set(retAddress, ret)
	}

	return nil
}

func (vm *VM) setToObject(instr *Instruction) error {
	av := vm.get(instr.A) // array or map
	bv := vm.get(instr.B) // index
	cv := vm.get(instr.C) // value

	if av.Type == Object {
		if cr, ok := av.ToObject().(*closureRegister); ok {
			// if it is a closure get the underlying value
			av = cr.get()
		}
	}

	if err := vm.AddAllocations(cv.Size()); err != nil {
		return err
	}

	switch bv.Type {
	case Int:
		switch av.Type {
		case Array:
			a := av.ToArray()
			i := int(bv.ToInt())
			if len(a) <= i {
				return vm.NewError("Index %d is out of range. Length is %d", i, len(a))
			}
			a[i] = cv
		case Bytes:
			if cv.Type != Int {
				return vm.NewError("Can't convert %v to byte", cv.TypeName())
			}
			av.ToBytes()[bv.ToInt()] = byte(cv.ToInt())
		case Object:
			i, ok := av.ToObject().(IndexerSetter)
			if !ok {
				return vm.NewError("Can't set by index %v", av.TypeName())
			}
			if err := i.SetIndex(int(bv.ToInt()), cv); err != nil {
				return vm.WrapError(err)
			}
		case Map:
			m := av.ToMap()
			m.Lock()
			m.Map[bv] = cv
			m.Unlock()
		default:
			return vm.NewError("Can't set %v by index", av.Type)
		}

	case Float:
		switch av.Type {
		case Map:
			m := av.ToMap()
			m.Lock()
			m.Map[bv] = cv
			m.Unlock()
		default:
			return vm.NewError("Invalid index %s for %s", bv.TypeName(), av.TypeName())
		}

	case Null:
		switch av.Type {
		case Map:
			m := av.ToMap()
			m.Lock()
			m.Map[bv] = cv
			m.Unlock()
		default:
			return vm.NewError("Invalid index %s for %s", bv.TypeName(), av.TypeName())
		}

	case String:
		switch av.Type {
		case Map:
			m := av.ToMap()
			m.Lock()
			m.Map[bv] = cv
			m.Unlock()
		case Object:
			obj := av.ToObject()
			key := bv.String()

			// try if it is a class instance with a property setter
			if instance, ok := obj.(*instance); ok {
				if set, ok := instance.PropertySetter(key, vm.Program); ok {
					args := []Value{cv}
					vm.callProgramFunc(set, Void, args, true, av, nil)
					return nil
				}
			}

			i, ok := obj.(FieldSetter)
			if !ok {
				return vm.NewError("Readonly or nonexistent field: %T", av.TypeName())
			}
			if err := i.SetField(key, cv, vm); err != nil {
				return vm.WrapError(err)
			}

		case Null:
			return vm.NewError("Can't set %s of null", bv.String())
		default:
			return vm.NewError("Readonly field or nonexistent field: %v", av.TypeName())
		}
	default:
		return vm.NewError("Invalid index %s", bv.TypeName())
	}

	return nil
}

// returns true if the source is not null
func (vm *VM) getFromObject(instr *Instruction, errIfNullOrUndefined bool) (bool, error) {
	bv := vm.get(instr.B) // source
	cv := vm.get(instr.C) // index or key

	if bv.IsNil() {
		if errIfNullOrUndefined {
			switch cv.Type {
			case Int:
				return false, vm.NewError("Cant read index %s of %s", cv.String(), bv.String())
			default:
				return false, vm.NewError("Cant read property %s of %s", cv.String(), bv.String())
			}
		} else {
			// set the address of the null or undefined value to
			// set the optchain result
			vm.optchainSrc = instr.B
			return false, nil
		}
	}

	if bv.Type == Object {
		if cr, ok := bv.ToObject().(*closureRegister); ok {
			// if it is a closure get the underlying value
			bv = cr.get()
		}
	}

	switch cv.Type {
	case Int:
		switch bv.Type {
		case Array:
			i := cv.ToInt()
			if i < 0 {
				return false, vm.NewError("Index out of range in string")
			}
			s := bv.ToArray()
			if len(s) <= int(i) {
				return false, vm.NewError("Index out of range")
			}
			vm.set(instr.A, s[i])

		case Enum:
			i := int(cv.ToInt())
			enum := vm.Program.Enums[int(bv.ToEnum())]
			if i < 0 || i >= len(enum.Values) {
				return false, vm.NewError("invalid enum index: %s %d [%d]", enum.Name, i, len(enum.Values))
			}
			k := vm.Program.Constants[enum.Values[i].KIndex]
			vm.set(instr.A, k)

		case Object:
			o := bv.ToObject()

			i, ok := o.(IndexerGetter)
			if !ok {
				return false, vm.NewError("%T can't be accessed by index", o)
			}

			v, err := i.GetIndex(int(cv.ToInt()))
			if err != nil {
				return false, vm.WrapError(err)
			}
			vm.set(instr.A, v)

		case String:
			i := cv.ToInt()
			if i < 0 {
				return false, vm.NewError("Index out of range in string")
			}
			vm.set(instr.A, NewRune(rune(bv.String()[i])))

		case Bytes:
			i := cv.ToInt()
			if i < 0 {
				return false, vm.NewError("Index out of range in string")
			}
			v := bv.ToBytes()
			b := v[i]
			vm.set(instr.A, NewInt(int(b)))

		case Map:
			m := bv.ToMap()
			m.RLock()
			v, ok := m.Map[cv]
			if !ok {
				v = UndefinedValue
			}
			m.RUnlock()
			vm.set(instr.A, v)

		default:
			return false, vm.NewError("The value must be Array or Indexer: %v", bv.Type)
		}

	case Float:
		switch bv.Type {
		case Map:
			m := bv.ToMap()
			m.RLock()
			v, ok := m.Map[cv]
			if !ok {
				v = UndefinedValue
			}
			m.RUnlock()
			vm.set(instr.A, v)
		default:
			return false, vm.NewError("Invalid index %s for %s", cv.TypeName(), bv.TypeName())
		}

	case Null:
		switch bv.Type {
		case Map:
			m := bv.ToMap()
			m.RLock()
			v, ok := m.Map[cv]
			if !ok {
				v = UndefinedValue
			}
			m.RUnlock()
			vm.set(instr.A, v)
		default:
			return false, vm.NewError("Invalid index %s for %s", cv.TypeName(), bv.TypeName())
		}

	// If is string is a property or method.
	case String:
		key := cv.String()

		switch bv.Type {

		case Enum:
			enum := vm.Program.Enums[int(bv.ToEnum())]
			v, i := enum.ValueByName(key)
			if i == -1 {
				return false, vm.NewError("invalid enum key: %s.%s", enum.Name, key)
			}
			k := vm.Program.Constants[v.KIndex]
			vm.set(instr.A, k)

		case Map:
			m := bv.ToMap()
			m.RLock()
			v, ok := m.Map[cv]
			if !ok {
				v = UndefinedValue
			}
			m.RUnlock()
			vm.set(instr.A, v)

		case Object:
			obj := bv.ToObject()

			if n, ok := obj.(Callable); ok {
				if m := n.GetMethod(key); m != nil {
					vm.set(instr.A, NewObject(m))
					return true, nil
				}
			}

			// try if it is a class instance with a property getter
			if instance, ok := obj.(*instance); ok {
				if get, ok := instance.PropertyGetter(key, vm.Program); ok {
					vm.callProgramFunc(get, instr.A, nil, true, bv, nil)
					return true, nil
				}
			}

			// try a native property
			if i, ok := obj.(FieldGetter); ok {
				v, err := i.GetField(key, vm)
				if err != nil {
					return false, vm.WrapError(err)
				}
				if v.Type != Undefined {
					vm.set(instr.A, v)
					return true, nil
				}
			}

			// try if it's an enunmerable method
			if _, ok := obj.(Enumerable); ok {
				if m, ok := vm.getNativePrototype("Array.prototype."+key, bv); ok {
					vm.set(instr.A, NewObject(m))
					return true, nil
				}
				if m, ok := vm.getProgramPrototype("Array.prototype."+key, bv); ok {
					vm.set(instr.A, NewObject(m))
					return true, nil
				}
			}

			// allow to call anything on an object.
			// If it doesn't exist set it to undefined
			vm.set(instr.A, UndefinedValue)

		case Array:
			switch key {
			case "length":
				vm.set(instr.A, NewInt(len(bv.ToArray())))
				return true, nil
			default:
				if !vm.setPrototype("Array.prototype."+key, bv, instr.A) {
					vm.set(instr.A, UndefinedValue)
					return false, nil
				}
				return true, nil
			}

		case String:
			switch key {
			case "length":
				vm.set(instr.A, NewInt(len(bv.String())))
				return true, nil
			case "runeCount":
				vm.set(instr.A, NewInt(utf8.RuneCountInString(bv.String())))
				return true, nil
			default:
				if !vm.setPrototype("String.prototype."+key, bv, instr.A) {
					vm.set(instr.A, UndefinedValue)
					return false, nil
				}
				return true, nil
			}

		case Undefined, Null:
			if errIfNullOrUndefined {
				return false, vm.NewError("Cant read property of %s", bv.String())
			}
			return false, nil

		case Bytes:
			switch key {
			case "length":
				vm.set(instr.A, NewInt(len(bv.ToBytes())))
				return true, nil
			default:
				if !vm.setPrototype("Bytes.prototype."+key, bv, instr.A) {
					if !vm.setPrototype("Array.prototype."+key, bv, instr.A) {
						vm.set(instr.A, UndefinedValue)
						return false, nil
					}
				}
				return true, nil
			}

		case Int, Float, Bool:
			return false, vm.NewError("Can't read '%s' from %s (%s)", cv.String(), bv.String(), bv.TypeName())

		default:
			vm.set(instr.A, UndefinedValue)
			return false, nil
		}

	default:
		return false, vm.NewError("Invalid index %s", cv.TypeName())
	}

	return true, nil
}

func (vm *VM) instruction() *Instruction {
	frame := vm.callStack[vm.fp]
	i := frame.funcIndex
	f := vm.Program.Functions[i]
	return f.Instructions[frame.pc]
}

type stackFrame struct {
	pc           int
	funcIndex    int
	maxRegIndex  int
	retAddress   *Address
	values       []Value
	closures     []*closureRegister
	finalizables []Finalizable
	exit         bool // if it should exit the program when returns
	inClosure    bool
}

type Method struct {
	ThisObject Value
	FuncIndex  int
}

func (*Method) Type() string {
	return "Method"
}

func (*Method) Export(recursionLevel int) interface{} {
	return "[method]"
}

type tryCatch struct {
	catchPC         int
	errorReg        *Address
	finallyPC       int
	fp              int
	retPC           int
	err             error
	catchExecuted   bool
	finallyExecuted bool
}

type Closure struct {
	FuncIndex int
	closures  []*closureRegister
}

func (c *Closure) Type() string {
	return "Closure"
}

func (c *Closure) Export(recursionLevel int) interface{} {
	return "[closure]"
}

type closureRegister struct {
	register *Register
	values   []Value
}

func (c *closureRegister) get() Value {
	return c.values[c.register.Index]
}

func (c *closureRegister) set(v Value) {
	c.values[c.register.Index] = v
}

func (c *closureRegister) Type() string {
	return c.get().Type.String()
}

func (c *closureRegister) Size() int {
	return c.get().Size()
}

func (c *closureRegister) Export(recursionLevel int) interface{} {
	return c.get().Export(recursionLevel)
}

func Eval(vm *VM, code string) error {
	// try to evaluate it as an expression to print the result
	ok, err := evalExpression(vm, code)
	if err == nil {
		return nil
	}

	// if the syntax was valid then return the error. If not, it was
	// not an expression so proceed to execute it as a statement
	if ok {
		return err
	}

	// run it as non expression code
	return eval(vm, code)
}

// returns if the expression syntax is valid
func evalExpression(vm *VM, code string) (bool, error) {
	code = fmt.Sprintf("print(%s)", code)

	// try if it is a valid expression
	if _, err := parser.ParseExpr(code); err != nil {
		return false, err
	}

	return true, eval(vm, code)
}

func eval(vm *VM, code string) error {
	frame := vm.callStack[0]

	p, err := appendCompile(vm.Program, code)
	if err != nil {
		return err
	}

	pc := frame.pc
	oProgram := vm.Program
	oValues := make([]Value, len(frame.values))
	for i, v := range frame.values {
		oValues[i] = v
	}

	// replace with the new program
	vm.Program = p

	// increase the frame memory for the new registers
	f := p.Functions[0]
	if f.MaxRegIndex > len(frame.values) {
		values := make([]Value, f.MaxRegIndex)
		copy(values, frame.values)
		frame.values = values
	}

	// run from the last frame pc
	vm.run(false)
	if vm.Error == io.EOF {
		vm.Error = nil
	}

	// on error restore everything and return the error
	if vm.Error != nil {
		frame.pc = pc
		vm.Program = oProgram
		frame.values = oValues
		err := vm.Error
		vm.Error = nil
		return err
	}

	return nil
}

func appendCompile(p *Program, code string) (*Program, error) {
	a, err := parser.ParseStr(code)
	if err != nil {
		return nil, err
	}

	// copy it so if it fails the original program is left intact
	p = p.Copy()

	f := p.Functions[0]
	ln := len(f.Instructions) - 1
	ret := f.Instructions[ln]
	f.Instructions = f.Instructions[:ln]

	fi := &functionInfo{
		function:    f,
		registerTop: f.MaxRegIndex,
	}

	c := &compiler{
		program:      p,
		functions:    make(map[string]*functionInfo),
		builtinFuncs: builtinFuncs,
		globalFunc:   fi,
		currentFunc:  fi,
	}

	c.functions[f.Name] = fi

	// add existing functions to prevent redeclarations
	for _, fn := range p.Functions[1:] {
		c.functions[fn.Name] = &functionInfo{function: fn}
	}

	if err := c.compileStmts(a.File.Stms); err != nil {
		return nil, err
	}

	if err := c.fixUnresolved(); err != nil {
		return nil, err
	}

	f.Instructions = append(f.Instructions, ret)

	return p, nil
}
