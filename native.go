package dune

import (
	"strings"
)

var allNativeFuncs []NativeFunction
var allNativeMap map[string]NativeFunction = make(map[string]NativeFunction)
var typeDefs = []string{header}

type NativeObject interface {
	GetMethod(name string) NativeMethod
	GetField(name string, vm *VM) (Value, error)
	SetField(name string, v Value, vm *VM) error
}

// NativeFunction is a function written in Go as opposed to an interpreted function
type NativeFunction struct {
	Name        string
	Arguments   int
	Index       int
	Permissions []string
	Function    func(this Value, args []Value, vm *VM) (Value, error)
}

type NativeMethod func(args []Value, vm *VM) (Value, error)

func (NativeMethod) Type() string {
	return "[native method]"
}

type nativePrototype struct {
	this Value
	fn   int
}

func (nativePrototype) Type() string {
	return "[native prototype]"
}

func AddNativeFunc(f NativeFunction) {
	// replace if it already exists
	if existingFunc, ok := allNativeMap[f.Name]; ok {
		f.Index = existingFunc.Index
		allNativeMap[f.Name] = f
		return
	}

	f.Index = len(allNativeFuncs)
	allNativeFuncs = append(allNativeFuncs, f)
	allNativeMap[f.Name] = f
}

func RegisterLib(funcs []NativeFunction, dts string) {
	for _, f := range funcs {
		AddNativeFunc(f)
	}

	if dts != "" {
		typeDefs = append(typeDefs, dts)
	}
}

func NativeFuncFromIndex(i int) NativeFunction {
	return allNativeFuncs[i]
}

func NativeFuncFromName(name string) (NativeFunction, bool) {
	f, ok := allNativeMap[name]
	return f, ok
}

func AllNativeFuncs() []NativeFunction {
	return allNativeFuncs
}

func TypeDefs() string {
	return strings.Join(typeDefs, "\n\n")
}

const header = `/**
 * ------------------------------------------------------------------
 * Native definitions.
 * ------------------------------------------------------------------
 */

// for the ts compiler
interface Boolean { }
interface Function { }
interface IArguments { }
interface Number { }
interface Object { }
interface RegExp { }
interface byte { }

declare const Symbol: symbol

interface Symbol {
    iterator: symbol
}

interface IteratorReturnResult<TReturn> {
    done: true;
    value: TReturn;
}

type IteratorResult<T, TReturn = any> = IteratorReturnResult<TReturn>;

interface Iterator<T, TReturn = any, TNext = undefined> {
    next(...args: [] | [TNext]): IteratorResult<T, TReturn>;
    return?(value?: TReturn): IteratorResult<T, TReturn>;
    throw?(e?: any): IteratorResult<T, TReturn>;
}

interface Iterable<T> {
    [Symbol.iterator](): Iterator<T>;
}

interface IterableIterator<T> extends Iterator<T> {
    [Symbol.iterator](): IterableIterator<T>;
}

declare const Array: any

interface Array<T> {
    [n: number]: T
    [Symbol.iterator](): IterableIterator<T>
    slice(start?: number, count?: number): Array<T>
    range(start?: number, end?: number): Array<T>
    append(v: T[]): T[]
    push(...v: T[]): void
    pushRange(v: T[]): void
    length: number
    insertAt(i: number, v: T): void
    removeAt(i: number): void
    removeAt(from: number, to: number): void
    indexOf(v: T): number
    join(sep: string): T
    sort(comprarer: (a: T, b: T) => boolean): void
}



`
