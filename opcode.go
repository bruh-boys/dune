package dune

import (
	"fmt"
	"io"
	"math"
)

type Opcode byte

const (
	op_loadConstant         Opcode = iota // load constant A := RK(B)
	op_move                               // A := B
	op_moveAndTest                        // A := B and  set C with true if B is not null or empty.
	op_add                                // A := B + C
	op_subtract                           // A := B - C
	op_multiply                           // A := B * C
	op_divide                             // A := B / C
	op_modulo                             // A := B % C
	op_exponentiate                       // A := B ** C  exponentiation ooperator
	op_binaryOr                           // A := B | C
	op_and                                // A := B & C
	op_xor                                // A := B ^ C
	op_leftShift                          // A := B << C
	op_rightShift                         // A := B >> C
	op_inc                                // A = A++. B up
	op_dec                                // A = A--. B up
	op_unm                                // A := -RK(B)
	op_not                                // A := !RK(B)
	op_bitwiseNot                         // A := ~B  bitwise not
	op_setRegister                        // Set register: A register number, B register value
	op_newClass                           // create a new instance of a class: A class type, B retAddress, C argsAddress
	op_newClassSingleArg                  // create a new instance of a class with a single arg: A class type, B retAddress, C argsAddress
	op_newArray                           // create a new array:  A array, B size
	op_newMap                             // create a new map:  A map, B size
	op_keys                               // get the keys of a map or the indexes of an array: A := keys(B)
	op_values                             // get the values of a map or array: A := values(B)
	op_length                             // get the length of an array: A := len(B)
	op_getEnumValue                       // get value from enum: A dest, B enum C key
	op_getIndexOrKey                      // get array index or map key: A dest, B source C index or key
	op_getOptChain                        // get optional chaining: A dest, B source C index or key. Reg0 stores the PC to jump if B is null.
	op_setIndexOrKey                      // set array index or map key: A array or Map, B index or key, C value
	op_spread                             // spread last index of array A
	op_jump                               // jump A positions
	op_jumpBack                           // jump back A positions
	op_jumpIfEqual                        // jump if A and B are equal C instructions.
	op_jumpIfNotEqual                     // jump if A and B are different C instructions.
	op_testJump                           // test if true or not null and jump: test A and jump B instructions. C=(0=jump if true, 1 jump if false)
	op_equal                              // equality test
	op_notEqual                           // inequality test
	op_strictEqual                        // strict equality test
	op_strictNotEqual                     // strict inequality test
	op_less                               // less than
	op_lessOrEqual                        // less or equal than
	op_call                               // call: A funcIndex, B retAddress, C argsAddress
	op_calOptChain                        // call optional chaining: A funcIndex, B retAddress, C argsAddress. Reg0 stores the PC to jump if B is null.
	op_callSingleArg                      // call with single argument: A funcIndex, B retAddress, C argsAddress
	op_calOptChainSingleArg               // call optional with single argument: A funcIndex, B retAddress, C argsAddress. Reg0 stores the PC to jump if B is null.
	op_readNativeProperty                 // Read native property: A := B
	op_return                             // return from a call: A dest
	op_createClosure                      // create closure: A dest R(B value) funcIndex
	op_throw                              // throw. A contains the error
	op_try                                // try-catch: jump to A absolute pc, set the error to B. C: the 'finally' absolute pc.
	op_tryEnd                             // try-end: set the last try body as ended.
	op_catchEnd                           // catch-end: set the last catch body as ended. It is only emmited if there is no finally
	op_finallyEnd                         // finally-end: set the last finally body as ended.
	op_tryExit                            // try exit: a continue inside try/catch inside a loop for example
	op_deleteProperty                     // delete object property
)

const (
	vm_next = iota
	vm_continue
	vm_exit
)

func exec(i *Instruction, vm *VM) int {
	switch i.Opcode {
	case op_loadConstant:
		return exec_loadConstant(i, vm)

	case op_move:
		return exec_move(i, vm)

	case op_moveAndTest:
		return exec_moveAndTest(i, vm)

	case op_add:
		return exec_add(i, vm)

	case op_subtract:
		return exec_subtract(i, vm)

	case op_multiply:
		return exec_multiply(i, vm)

	case op_divide:
		return exec_divide(i, vm)

	case op_modulo:
		return exec_modulo(i, vm)

	case op_exponentiate:
		return exec_exponentiate(i, vm)

	case op_binaryOr:
		return exec_binaryOr(i, vm)

	case op_and:
		return exec_and(i, vm)

	case op_xor:
		return exec_xor(i, vm)

	case op_leftShift:
		return exec_leftShift(i, vm)

	case op_rightShift:
		return exec_rightShift(i, vm)

	case op_inc:
		return exec_inc(i, vm)

	case op_dec:
		return exec_dec(i, vm)

	case op_unm:
		return exec_unm(i, vm)

	case op_not:
		return exec_not(i, vm)

	case op_bitwiseNot:
		return exec_bitwiseNot(i, vm)

	case op_setRegister:
		return exec_setRegister(i, vm)

	case op_newClass:
		return exec_newClass(i, vm)

	case op_newClassSingleArg:
		return exec_newClassSingleArg(i, vm)

	case op_newArray:
		return exec_newArray(i, vm)

	case op_newMap:
		return exec_newMap(i, vm)

	case op_keys:
		return exec_keys(i, vm)

	case op_values:
		return exec_values(i, vm)

	case op_length:
		return exec_length(i, vm)

	case op_getEnumValue:
		return exec_getEnumValue(i, vm)

	case op_getIndexOrKey:
		return exec_getIndexOrKey(i, vm)

	case op_getOptChain:
		return exec_getOptChain(i, vm)

	case op_setIndexOrKey:
		return exec_setIndexOrKey(i, vm)

	case op_spread:
		return exec_spread(i, vm)

	case op_jump:
		return exec_jump(i, vm)

	case op_jumpBack:
		return exec_jumpBack(i, vm)

	case op_jumpIfEqual:
		return exec_jumpIfEqual(i, vm)

	case op_jumpIfNotEqual:
		return exec_jumpIfNotEqual(i, vm)

	case op_testJump:
		return exec_testJump(i, vm)

	case op_equal:
		return exec_equal(i, vm)

	case op_notEqual:
		return exec_notEqual(i, vm)

	case op_strictEqual:
		return exec_strictEqual(i, vm)

	case op_strictNotEqual:
		return exec_strictNotEqual(i, vm)

	case op_less:
		return exec_less(i, vm)

	case op_lessOrEqual:
		return exec_lessOrEqual(i, vm)

	case op_call:
		return exec_call(i, vm)

	case op_calOptChain:
		return exec_calOptChain(i, vm)

	case op_callSingleArg:
		return exec_callSingleArg(i, vm)

	case op_calOptChainSingleArg:
		return exec_calOptChainSingleArg(i, vm)

	case op_readNativeProperty:
		return exec_readNativeProperty(i, vm)

	case op_return:
		return exec_return(i, vm)

	case op_createClosure:
		return exec_createClosure(vm)

	case op_throw:
		return exec_throw(i, vm)

	case op_try:
		return exec_try(i, vm)

	case op_tryEnd:
		return exec_tryEnd(vm)

	case op_catchEnd:
		return exec_catchEnd(vm)

	case op_finallyEnd:
		return exec_finallyEnd(vm)

	case op_tryExit:
		return exec_tryExit(vm)

	case op_deleteProperty:
		return exec_deleteProperty(i, vm)

	default:
		panic(fmt.Sprintf("Invalid opcode: %v", i))
	}
}

func exec_move(instr *Instruction, vm *VM) int {
	vm.set(instr.A, vm.get(instr.B))
	return vm_next
}

func exec_moveAndTest(instr *Instruction, vm *VM) int {
	// set A with B if it has instr.A value and C with true is set.
	bv := vm.get(instr.B)
	vm.set(instr.A, bv)
	switch bv.Type {
	case Bool:
		vm.set(instr.C, bv)

	case Int:
		if bv.ToInt() == 0 {
			vm.set(instr.C, FalseValue)
		} else {
			vm.set(instr.C, TrueValue)
		}

	case Float:
		if bv.ToFloat() == 0 {
			vm.set(instr.C, FalseValue)
		} else {
			vm.set(instr.C, TrueValue)
		}

	default:
		if bv.IsNilOrEmpty() {
			vm.set(instr.C, FalseValue)
		} else {
			vm.set(instr.C, TrueValue)
		}
	}

	return vm_next
}

func exec_loadConstant(instr *Instruction, vm *VM) int {
	k := vm.Program.Constants[instr.B.Value]
	vm.set(instr.A, k)
	return vm_next
}

func exec_add(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Float:
			vm.set(instr.A, NewFloat(lh.ToFloat()+rh.ToFloat()))
		case Int:
			vm.set(instr.A, NewInt64(lh.ToInt()+rh.ToInt()))
		case Rune:
			vm.set(instr.A, NewRune(lh.ToRune()+rh.ToRune()))
		case String:
			err := vm.AddAllocations(lh.Size())
			if err == nil {
				err = vm.AddAllocations(rh.Size())
			}
			if err != nil {
				if vm.handle(err) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			vm.set(instr.A, NewString(lh.ToString()+rh.ToString()))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle(vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type)) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Float:
		switch rh.Type {
		case Int, Float:
			vm.set(instr.A, NewFloat(lh.ToFloat()+rh.ToFloat()))
		case Rune:
			vm.set(instr.A, NewRune(lh.ToRune()+rh.ToRune()))
		case String:
			err := vm.AddAllocations(lh.Size())
			if err == nil {
				err = vm.AddAllocations(rh.Size())
			}
			if err != nil {
				if vm.handle(err) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			vm.set(instr.A, NewString(lh.ToString()+rh.ToString()))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle(vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type)) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Rune:
		switch rh.Type {
		case Rune, Int:
			vm.set(instr.A, NewRune(lh.ToRune()+rh.ToRune()))
		case String:
			err := vm.AddAllocations(lh.Size())
			if err == nil {
				err = vm.AddAllocations(rh.Size())
			}
			if err != nil {
				if vm.handle(err) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			vm.set(instr.A, NewString(lh.ToString()+rh.ToString()))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Bool:
		switch rh.Type {
		case String:
			vm.set(instr.A, NewString(lh.ToString()+rh.ToString()))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case String:
		switch rh.Type {
		case String, Int, Float, Bool, Rune:
			err := vm.AddAllocations(lh.Size())
			if err == nil {
				err = vm.AddAllocations(rh.Size())
			}
			if err != nil {
				if vm.handle((err)) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			vm.set(instr.A, NewString(lh.ToString()+rh.ToString()))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Null, Undefined:
		switch rh.Type {
		case String, Int, Float, Rune:
			vm.set(instr.A, rh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}

	return vm_next
}

func exec_subtract(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Float:
			vm.set(instr.A, NewFloat(lh.ToFloat()-rh.ToFloat()))
		case Int:
			vm.set(instr.A, NewInt64(lh.ToInt()-rh.ToInt()))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Float:
		switch rh.Type {
		case Int, Float:
			vm.set(instr.A, NewFloat(lh.ToFloat()-rh.ToFloat()))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Rune:
		switch rh.Type {
		case Rune, Int:
			vm.set(instr.A, NewRune(lh.ToRune()-rh.ToRune()))
		case String:
			rs := rh.ToString()
			if len(rs) != 1 {
				if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			vm.set(instr.A, NewRune(lh.ToRune()-rune(rs[0])))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case String:
		ls := lh.ToString()
		if len(ls) != 1 {
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
		lr := rune(ls[0])
		switch rh.Type {
		case Rune, Int:
			vm.set(instr.A, NewRune(lr-rh.ToRune()))
		case String:
			rs := rh.ToString()
			if len(rs) != 1 {
				if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			vm.set(instr.A, NewRune(lr-rune(rs[0])))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Null, Undefined:
		switch rh.Type {
		case String, Int, Float, Rune:
			vm.set(instr.A, rh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_multiply(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Float:
			vm.set(instr.A, NewFloat(lh.ToFloat()*rh.ToFloat()))
		case Int:
			vm.set(instr.A, NewInt64(lh.ToInt()*rh.ToInt()))
		case Rune:
			vm.set(instr.A, NewRune(lh.ToRune()*rh.ToRune()))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Float:
		switch rh.Type {
		case Int, Float:
			vm.set(instr.A, NewFloat(lh.ToFloat()*rh.ToFloat()))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Rune:
		switch rh.Type {
		case Rune, Int:
			vm.set(instr.A, NewRune(lh.ToRune()*rh.ToRune()))
		case Null, Undefined:
			vm.set(instr.A, lh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Null, Undefined:
		switch rh.Type {
		case Int, Float, Rune:
			vm.set(instr.A, rh)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}

	return vm_next
}

func exec_binaryOr(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Int:
			vm.set(instr.A, NewInt64(lh.ToInt()|rh.ToInt()))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_leftShift(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Int:
			vm.set(instr.A, NewInt64(int64(lh.ToInt()<<uint64(rh.ToInt()))))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_rightShift(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Int:
			vm.set(instr.A, NewInt64(int64(lh.ToInt()>>uint64(rh.ToInt()))))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_xor(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Int:
			vm.set(instr.A, NewInt(int(lh.ToInt())^int(rh.ToInt())))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_and(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Int:
			vm.set(instr.A, NewInt(int(lh.ToInt())&int(rh.ToInt())))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_divide(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)

	if lh.Type == Rune || rh.Type == Rune {
		switch lh.Type {
		case Int, Rune, Null, Undefined:
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
		switch rh.Type {
		case Int, Rune:
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
		vm.set(instr.A, NewRune(lh.ToRune()/rh.ToRune()))
	} else {
		switch lh.Type {
		case Int, Float, Null, Undefined:
			// handled below
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}

		switch rh.Type {
		case Int, Float, Null, Undefined:
			// handled below
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}

		rf := rh.ToFloat()
		if rf == 0 {
			if vm.handle((vm.NewError("Attempt to divide by zero"))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
		vm.set(instr.A, NewFloat(lh.ToFloat()/rf))
	}
	return vm_next
}

func exec_modulo(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Float:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	case Int:
		switch rh.Type {
		case Int, Null, Undefined:
			ri := rh.ToInt()
			if ri == 0 {
				if vm.handle((vm.NewError("Attempt to divide by zero"))) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			vm.set(instr.A, NewInt64(lh.ToInt()%ri))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Rune:
		switch rh.Type {
		case Rune, Int, Null, Undefined:
			ri := rh.ToRune()
			if ri == 0 {
				if vm.handle((vm.NewError("Attempt to divide by zero"))) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			vm.set(instr.A, NewRune(lh.ToRune()%ri))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Null, Undefined:
		ri := rh.ToFloat()
		if ri == 0 {
			if vm.handle((vm.NewError("Attempt to divide by zero"))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
		vm.set(instr.A, NewInt(0))
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_exponentiate(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)

	switch lh.Type {
	case Int, Float:
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}

	switch rh.Type {
	case Int, Float:
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}

	v := math.Pow(lh.ToFloat(), rh.ToFloat())
	vm.set(instr.A, NewFloat(v))
	return vm_next
}

func exec_setRegister(instr *Instruction, vm *VM) int {
	if instr.A.Kind != AddrData {
		panic(fmt.Sprintf("Compiler error: invalid register number kind: %v", instr.A))
	}

	if instr.B.Kind != AddrData {
		panic(fmt.Sprintf("Compiler error: invalid register value kind: %v", instr.B))
	}

	switch instr.A.Value {
	case 0:
		vm.reg0 = instr.B.Value
	default:
		panic(fmt.Sprintf("Compiler error: invalid register number: %v", instr.A))
	}

	return vm_next
}

func exec_bitwiseNot(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	switch lh.Type {
	case Int:
		vm.set(instr.A, NewInt64(^lh.ToInt()))
	default:
		if vm.handle((vm.NewError("Invalid operation on %v", lh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_unm(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	switch lh.Type {
	case Int:
		vm.set(instr.A, NewInt64(lh.ToInt()*-1))
	case Float:
		vm.set(instr.A, NewFloat(lh.ToFloat()*-1))
	default:
		if vm.handle((vm.NewError("Invalid operation on %v", lh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_not(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	switch lh.Type {
	case Bool:
		vm.set(instr.A, NewBool(!lh.ToBool()))
	case Int:
		vm.set(instr.A, NewBool(lh.ToInt() == 0))
	case Float:
		vm.set(instr.A, NewBool(lh.ToFloat() == 0))
	default:
		vm.set(instr.A, NewBool(lh.IsNilOrEmpty()))
	}
	return vm_next
}

func exec_inc(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.A)
	switch lh.Type {
	case Int:
		vm.set(instr.A, NewInt64(lh.ToInt()+1))
	case Float:
		vm.set(instr.A, NewFloat(lh.ToFloat()+1))
	default:
		if vm.handle((vm.NewError("Invalid operation on %v", lh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_dec(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.A)
	switch lh.Type {
	case Int:
		vm.set(instr.A, NewInt64(lh.ToInt()-1))
	case Float:
		vm.set(instr.A, NewFloat(lh.ToFloat()-1))
	default:
		if vm.handle((vm.NewError("Invalid operation on %v", lh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_less(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)

	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Float:
			vm.set(instr.A, NewBool(lh.ToFloat() < rh.ToFloat()))
		case Int:
			vm.set(instr.A, NewBool(lh.ToInt() < rh.ToInt()))
		case Null:
			vm.set(instr.A, NewBool(0 < rh.ToInt()))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Float:
		switch rh.Type {
		case Int, Float:
			vm.set(instr.A, NewBool(lh.ToFloat() < rh.ToFloat()))
		case Null:
			vm.set(instr.A, NewBool(0 < rh.ToInt()))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Rune:
		switch rh.Type {
		case Rune, Int:
			vm.set(instr.A, NewBool(lh.ToRune() < rh.ToRune()))
		case String:
			rs := rh.ToString()
			if len(rs) != 1 {
				if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			rr := rune(rs[0])
			vm.set(instr.A, NewBool(lh.ToRune() < rr))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case String:
		switch rh.Type {
		case String:
			vm.set(instr.A, NewBool(lh.ToString() < rh.ToString()))
		case Rune:
			ls := lh.ToString()
			if len(ls) != 1 {
				if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			lr := rune(ls[0])
			vm.set(instr.A, NewBool(lr < rh.ToRune()))
		case Null:
			vm.set(instr.A, FalseValue)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Null:
		switch rh.Type {
		case Int, Float:
			vm.set(instr.A, NewBool(0 < rh.ToFloat()))
		case String, Rune:
			vm.set(instr.A, FalseValue)
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Object:
		var set bool
		if rh.Type == Object {
			lComp, ok := lh.ToObjectOrNil().(Comparable)
			if ok {
				vComp := lComp.Compare(rh)
				// -2 means that both values are not comparable between each other
				if vComp != -2 {
					vm.set(instr.A, NewBool(vComp < 0))
					set = true
				}
			}
		}
		if !set {
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_lessOrEqual(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	switch lh.Type {
	case Int:
		switch rh.Type {
		case Float:
			vm.set(instr.A, NewBool(lh.ToFloat() <= rh.ToFloat()))
		case Int:
			vm.set(instr.A, NewBool(lh.ToInt() <= rh.ToInt()))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Float:
		switch rh.Type {
		case Int, Float:
			vm.set(instr.A, NewBool(lh.ToFloat() <= rh.ToFloat()))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case Rune:
		switch rh.Type {
		case Rune, Int:
			vm.set(instr.A, NewBool(lh.ToRune() <= rh.ToRune()))
		case String:
			rs := rh.ToString()
			if len(rs) != 1 {
				if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			rr := rune(rs[0])
			vm.set(instr.A, NewBool(lh.ToRune() <= rr))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	case String:
		switch rh.Type {
		case String:
			vm.set(instr.A, NewBool(lh.ToString() <= rh.ToString()))
		case Rune:
			ls := lh.ToString()
			if len(ls) != 1 {
				if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
					return vm_continue
				} else {
					return vm_exit
				}
			}
			lr := rune(ls[0])
			vm.set(instr.A, NewBool(lr <= rh.ToRune()))
		default:
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}

	case Object:
		var set bool
		if rh.Type == Object {
			lComp, ok := lh.ToObjectOrNil().(Comparable)
			if ok {
				vComp := lComp.Compare(rh)
				// -2 means that both values are not comparable between each other
				if vComp != -2 {
					vm.set(instr.A, NewBool(vComp <= 0))
					set = true
				}
			}
		}
		if !set {
			if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("Invalid operation on %v and %v", lh.Type, rh.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_equal(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	vm.set(instr.A, NewBool(lh.Equals(rh)))
	return vm_next
}

func exec_notEqual(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	vm.set(instr.A, NewBool(!lh.Equals(rh)))
	return vm_next
}

func exec_strictEqual(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	vm.set(instr.A, NewBool(lh.StrictEquals(rh)))
	return vm_next
}

func exec_strictNotEqual(instr *Instruction, vm *VM) int {
	lh := vm.get(instr.B)
	rh := vm.get(instr.C)
	vm.set(instr.A, NewBool(!lh.StrictEquals(rh)))
	return vm_next
}

func exec_newArray(instr *Instruction, vm *VM) int {
	vm.set(instr.A, NewArray(int(instr.B.Value)))
	return vm_next
}

func exec_newMap(instr *Instruction, vm *VM) int {
	vm.set(instr.A, NewMap(int(instr.B.Value)))
	return vm_next
}

func exec_getEnumValue(instr *Instruction, vm *VM) int {
	enum := vm.Program.Enums[int(instr.B.Value)]
	value := enum.Values[int(instr.C.Value)]
	k := vm.Program.Constants[value.KIndex]
	vm.set(instr.A, k)
	return vm_next
}

func exec_getIndexOrKey(instr *Instruction, vm *VM) int {
	if _, err := vm.getFromObject(instr, true); err != nil {
		if vm.handle((err)) {
			return vm_continue
		} else {
			vm.Error = err
			return vm_exit
		}
	}

	// set value in an array or map: A array, B index, C value
	return vm_next
}

func exec_getOptChain(instr *Instruction, vm *VM) int {
	ok, err := vm.getFromObject(instr, false)
	if err != nil {
		if vm.handle((err)) {
			return vm_continue
		} else {
			vm.Error = err
			return vm_exit
		}
	}

	if !ok {
		vm.incPC(int(vm.reg0))
		return vm_continue
	}

	// set value in an array or map: A array, B index, C value
	return vm_next
}

func exec_setIndexOrKey(instr *Instruction, vm *VM) int {
	if err := vm.setToObject(instr); err != nil {
		if vm.handle((err)) {
			return vm_continue
		} else {
			vm.Error = err
			return vm_exit
		}
	}
	return vm_next
}

func exec_spread(instr *Instruction, vm *VM) int {
	v := vm.get(instr.A)
	if v.Type != Array {
		if vm.handle((vm.NewError("Expected array, got %v", v.TypeName()))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}

	va := v.ToArrayObject().Array

	ln := len(va)
	if ln == 0 {
		return vm_next
	}

	last := va[ln-1]
	switch last.Type {
	case Null, Undefined:
		vm.set(instr.A, NewArrayValues(va[:ln-1]))
	case Array:
		n := append(va[:ln-1], last.ToArrayObject().Array...)
		vm.set(instr.A, NewArrayValues(n))
	default:
		if vm.handle((vm.NewError("Expected array, got %v", last.TypeName()))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}

	return vm_next
}

func exec_keys(instr *Instruction, vm *VM) int {
	// gets the keys of a map or the indexes of an array: A := keys(B)
	bv := vm.get(instr.B)
	switch bv.Type {

	case Null:
		// allow to iterate if not initialize (set an empty array)
		vm.set(instr.A, NewArray(0))

	case Array:
		s := bv.ToArray()
		ln := len(s)
		values := make([]Value, ln)
		for i := 0; i < ln; i++ {
			values[i] = NewInt(i)
		}
		vm.set(instr.A, NewArrayValues(values))

	case Map:
		m := bv.ToMap()
		m.RLock()
		s := m.Map
		values := make([]Value, len(s))
		i := 0
		for k := range s {
			values[i] = k
			i++
		}
		m.RUnlock()
		vm.set(instr.A, NewArrayValues(values))

	case Enum:
		i := bv.ToEnum()
		enum := vm.Program.Enums[i]
		ln := len(enum.Values)
		values := make([]Value, ln)
		for i := 0; i < ln; i++ {
			values[i] = NewInt(i)
		}
		vm.set(instr.A, NewArrayValues(values))

	case Object:
		obj := bv.ToObject()
		if n, ok := obj.(IndexIterator); ok {
			ln := n.Len()
			values := make([]Value, ln)
			for i := 0; i < ln; i++ {
				values[i] = NewInt(i)
			}
			vm.set(instr.A, NewArrayValues(values))
		} else {
			if vm.handle((vm.NewError("Expected a key or index enumerable, got %v", bv.TypeName()))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}

	default:
		if vm.handle((vm.NewError("Expected a enumerable, got %v", bv.TypeName()))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_values(instr *Instruction, vm *VM) int {
	// gets the values of a map or array: A := values(B)
	bv := vm.get(instr.B)
	switch bv.Type {

	case Null, Undefined:
		// allow to iterate if not initialize (set an empty array)
		vm.set(instr.A, NewArray(0))

	case Array:
		// copiar los valores para que si se modifican dentro de un loop no afecten a la iteración
		s := bv.ToArray()
		values := make([]Value, len(s))
		copy(values, s)
		vm.set(instr.A, NewArrayValues(values))
	case Bytes:
		s := bv.ToBytes()
		values := make([]Value, len(s))
		for i, v := range s {
			values[i] = NewInt(int(v))
		}
		vm.set(instr.A, NewArrayValues(values))
	case Map:
		m := bv.ToMap()
		m.RLock()
		s := m.Map
		values := make([]Value, len(s))
		i := 0
		for _, v := range s {
			values[i] = v
			i++
		}
		m.RUnlock()
		vm.set(instr.A, NewArrayValues(values))
	case Object:
		obj := bv.ToObject()
		if enum, ok := obj.(Enumerable); ok {
			vals, err := enum.Values()
			if err != nil {
				if vm.handle((vm.NewError("Enumerable error: %v", err))) {
					return vm_continue
				} else {
					return vm_exit
				}
			} else {
				vm.set(instr.A, NewArrayValues(vals))
			}
		} else if vm.handle((vm.NewError("Expected a enumerable, got %v", bv.String()))) {
			return vm_continue
		} else {
			return vm_exit
		}

	default:
		if vm.handle((vm.NewError("Expected a enumerable, got %v", bv.String()))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_length(instr *Instruction, vm *VM) int {
	bv := vm.get(instr.B)
	switch bv.Type {
	case Array:
		vm.set(instr.A, NewInt(len(bv.ToArray())))
	case Map:
		m := bv.ToMap()
		m.RLock()
		vm.set(instr.A, NewInt(len(m.Map)))
		m.RUnlock()
	case Object:
		if col, ok := bv.ToObject().(IndexIterator); ok {
			vm.set(instr.A, NewInt(col.Len()))
		} else {
			if vm.handle((vm.NewError("The value is not a collection %v", bv.Type))) {
				return vm_continue
			} else {
				return vm_exit
			}
		}
	default:
		if vm.handle((vm.NewError("The value is not a collection %v", bv.Type))) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

// Read native property A := B
func exec_readNativeProperty(instr *Instruction, vm *VM) int {
	n := vm.get(instr.B)

	i := n.ToNativeFunction()

	if err := vm.callNativeFunc(i, nil, instr.A, NullValue); err != nil {
		if vm.handle(vm.WrapError(err)) {
			return vm_continue
		} else {
			return vm_exit
		}
	}
	return vm_next
}

func exec_throw(instr *Instruction, vm *VM) int {
	v := vm.get(instr.A)

	var err error
	if v.Type == Object {
		if e, ok := v.ToObject().(*Error); ok {
			// don't alter the stack trace if it is a re throw
			if e.IsRethrow {
				err = e
			} else {
				err = vm.WrapError(e)
			}
		}
	}

	if err == nil {
		err = vm.NewError(v.String())
	}

	// check if is inside a catch to discard it.
	l := len(vm.tryCatchs) - 1
	if l < 0 {
		// run finalizers before exiting
		vm.cleanupNotGlobalFrame(vm.fp)

		// an unhandled error
		vm.Error = err
		return vm_exit
	}

	// check if we are in the finally block.
	// remove the current try if we are and let the next one handle it.
	try := vm.tryCatchs[l]
	if try.finallyExecuted {
		vm.tryCatchs = vm.tryCatchs[:l]
	}

	if vm.handle((err)) {
		return vm_continue
	} else {
		return vm_exit
	}
}

func exec_try(instr *Instruction, vm *VM) int {
	//  jump to A absolute pc, set the error to B. C: the 'finally' absolute pc.
	var catchPC int
	if instr.A.Kind == AddrVoid {
		catchPC = -1
	} else {
		catchPC = int(instr.A.Value)
	}

	try := &tryCatch{
		catchPC:  catchPC,
		retPC:    -1,
		errorReg: instr.B,
		fp:       vm.fp,
	}

	// set the finally pc if provided
	if instr.C.Kind == AddrData {
		try.finallyPC = int(instr.C.Value)
	} else {
		try.finallyPC = -1
	}

	vm.tryCatchs = append(vm.tryCatchs, try)
	return vm_next
}

func exec_tryExit(vm *VM) int {
	i := len(vm.tryCatchs) - 1

	if i >= 0 {
		try := vm.tryCatchs[i]

		// if there is no finally just remove it
		if try.finallyPC == -1 {
			vm.tryCatchs = vm.tryCatchs[:i]
			return vm_next
		}

		// make it continue to the next instruction after the finally ends
		try.retPC = vm.callStack[vm.fp].pc + 1

		// advance to the finally part
		vm.setPC(try.finallyPC)

		return vm_continue
	}

	return vm_continue
}

func exec_tryEnd(vm *VM) int {
	l := len(vm.tryCatchs) - 1

	try := vm.tryCatchs[l]
	if try.finallyPC == -1 {
		// if there is no finally, discard it
		vm.tryCatchs = vm.tryCatchs[:l]
	}
	return vm_next
}

func exec_catchEnd(vm *VM) int {
	l := len(vm.tryCatchs) - 1

	// don't need to check finally because cen is only emmited if there is no finally
	vm.tryCatchs = vm.tryCatchs[:l]
	return vm_next
}

func exec_finallyEnd(vm *VM) int {
	l := len(vm.tryCatchs) - 1
	try := vm.tryCatchs[l]
	vm.tryCatchs = vm.tryCatchs[:l]

	// if the error was unhandled because there was no catch block
	// the handle it now that the finally has been processed.
	if try.err != nil {
		vm.incPC(1)
		if vm.handle(try.err) {
			return vm_continue
		} else {
			return vm_exit
		}
	}

	if try.retPC != -1 {
		vm.setPC(try.retPC)
		return vm_continue
	}

	return vm_next
}

func exec_jumpIfEqual(instr *Instruction, vm *VM) int {
	// jump if A and B are equal C instructions.

	lh := vm.get(instr.A)
	rh := vm.get(instr.B)

	if lh.Equals(rh) {
		vm.incPC(int(instr.C.Value))
	}
	return vm_next
}

func exec_jumpIfNotEqual(instr *Instruction, vm *VM) int {
	// jump if A and B are different C instructions.

	lh := vm.get(instr.A)
	rh := vm.get(instr.B)

	if !lh.Equals(rh) {
		vm.incPC(int(instr.C.Value))
	}
	return vm_next
}

func exec_testJump(instr *Instruction, vm *VM) int {
	// test if true or not null and jump: test A and jump B instructions. C=(0=jump if true, 1 jump if false)

	av := vm.get(instr.A)
	cv := instr.C.Value

	switch jumpType(cv) {
	case jumpIfFalse:
		var expr bool
		switch av.Type {
		case Bool:
			expr = av.ToBool()
		case Int:
			// if the value is 0 treat it as null or empty like in javascript.
			expr = av.ToInt() != 0
		case Float:
			// if the value is 0 treat it as null or empty like in javascript.
			expr = av.ToFloat() != 0
		default:
			expr = !av.IsNilOrEmpty() // true if it has a value like in javascript
		}
		if expr {
			vm.incPC(int(instr.B.Value))
		}
	case jumpIfTrue:
		var expr bool
		switch av.Type {
		case Bool:
			expr = av.ToBool()
		case Int:
			// if the value is 0 treat it as null or empty like in javascript.
			expr = av.ToInt() != 0
		case Float:
			// if the value is 0 treat it as null or empty like in javascript.
			expr = av.ToFloat() != 0
		default:
			expr = !av.IsNilOrEmpty() // true if it has a value like in javascript
		}
		if !expr {
			vm.incPC(int(instr.B.Value))
		}
	case jumpIfNotNull:
		if !av.IsNil() {
			vm.incPC(int(instr.B.Value))
		}
	}

	return vm_next
}

func exec_jump(instr *Instruction, vm *VM) int {
	vm.incPC(int(instr.A.Value))
	return vm_next
}

func exec_jumpBack(instr *Instruction, vm *VM) int {
	vm.incPC(int(instr.A.Value) * -1)
	return vm_continue
}

func exec_createClosure(vm *VM) int {
	// R(A) dest R(B value) funcIndex
	instr := vm.instruction()
	funcIndex := instr.B.Value

	// copy  closures carried from parent functions
	frame := vm.callStack[vm.fp]
	f := vm.Program.Functions[frame.funcIndex]
	fLen := len(f.Closures)
	frLen := len(frame.closures)

	// mark it so it is not reused
	frame.inClosure = true

	c := &Closure{
		FuncIndex: int(funcIndex),
		closures:  make([]*closureRegister, fLen+frLen),
	}

	copy(c.closures, frame.closures)

	// copy closures defined in this function.
	for i, r := range f.Closures {
		c.closures[frLen+i] = &closureRegister{register: r, values: frame.values}
	}

	vm.set(instr.A, NewObject(c))
	return vm_next
}

func exec_newClass(instr *Instruction, vm *VM) int {
	// A class index, B retAddress, C argsAddress

	var args []Value
	if instr.C != Void {
		args = vm.get(instr.C).ToArrayObject().Array
	}

	i := newInstance(instr.A, vm)

	v := NewObject(i)
	vm.set(instr.B, v)

	f, ok := i.Function("constructor", vm.Program)
	if ok {
		return vm.callProgramFunc(f, Void, args, true, v, nil)
	}
	return vm_next
}

func exec_newClassSingleArg(instr *Instruction, vm *VM) int {
	// A class index, B retAddress, C argsAddress

	args := []Value{vm.get(instr.C)}

	i := newInstance(instr.A, vm)

	v := NewObject(i)
	vm.set(instr.B, v)

	f, ok := i.Function("constructor", vm.Program)
	if ok {
		return vm.callProgramFunc(f, Void, args, true, v, nil)
	}

	return vm_next
}

func exec_call(instr *Instruction, vm *VM) int {
	// A funcIndex, B retAddress, C argsAddress

	var args []Value
	if instr.C != Void {
		args = vm.get(instr.C).ToArrayObject().Array
	}

	return vm.call(instr.A, instr.B, args, false)
}

func exec_calOptChain(instr *Instruction, vm *VM) int {
	// A funcIndex, B retAddress, C argsAddress

	var args []Value
	if instr.C != Void {
		args = vm.get(instr.C).ToArrayObject().Array
	}

	return vm.call(instr.A, instr.B, args, true)
}

func exec_callSingleArg(instr *Instruction, vm *VM) int {
	// A funcIndex, B retAddress, C argsAddress
	args := []Value{vm.get(instr.C)}
	return vm.call(instr.A, instr.B, args, false)
}

func exec_calOptChainSingleArg(instr *Instruction, vm *VM) int {
	// A funcIndex, B retAddress, C argsAddress
	args := []Value{vm.get(instr.C)}
	return vm.call(instr.A, instr.B, args, true)
}

func exec_return(instr *Instruction, vm *VM) int {
	currentFrame := vm.callStack[vm.fp]

	// check if we are inside a try-finally
	if vm.returnFromFinally() {
		return vm_continue
	}

	// run finalizers for all functions except the global func
	// which is handled in the main run loop
	if vm.fp > 0 {
		vm.runFinalizables(currentFrame)
	}

	var retValue Value
	if instr.A != Void {
		retValue = vm.get(instr.A)
	}

	if vm.fp == 0 {
		// returning from main: exit
		vm.Error = io.EOF
		vm.RetValue = retValue
		return vm_exit
	}

	// pop one frame
	vm.callStack = vm.callStack[:vm.fp]
	vm.fp--

	// if the frame can be reused, clear its memory and add it to the cache.
	// This makes a huge impact in memory hungry programs.
	if !currentFrame.inClosure {
		currentFrame.finalizables = nil
		currentFrame.closures = nil
		for i := range currentFrame.values {
			currentFrame.values[i] = UndefinedValue
		}
		vm.frameCache = append(vm.frameCache, currentFrame)
	}

	prevFrame := vm.callStack[vm.fp]

	// set the return value
	if prevFrame.retAddress != Void {
		vm.set(prevFrame.retAddress, retValue)
	}

	if currentFrame.exit {
		vm.RetValue = retValue
		return vm_exit
	}

	return vm_continue
}

func exec_deleteProperty(instr *Instruction, vm *VM) int {
	obj := vm.get(instr.A)
	if obj.Type != Map {
		return vm_next
	}

	property := vm.get(instr.B)

	m := obj.ToMap()
	m.Lock()
	delete(m.Map, property)
	m.Unlock()
	return vm_next
}
