package dune

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dunelang/dune/filesystem"
)

// Tests: Expressions
func TestExpression1(t *testing.T) {
	data := []struct {
		expression string
		expected   interface{}
	}{
		{"return 3", int64(3)},
		{"return 3+3", int64(6)},
		{"return -1 + 2", int64(1)},
		{"return -1 - 2", int64(-3)},
		{"return -1 + -2", int64(-3)},
		{"return 1 + -2", int64(-1)},
		{"return 3 * 3", int64(9)},
		{"return 6 / 2", float64(3)},
		{"return 2 ** 3", float64(8)},
		{"return 11 % 3", int64(2)},
		{"return 3 + 10 / 2", float64(8)},
		{"return 3 + 10 % 2", int64(3)},
		{"return 3 + 2 * 2", int64(7)},
		{"return (3 + 2) * 2", int64(10)},
		{"return (3 + 2) * (3 * 2)", int64(30)},
		{"return ((3 + 2) * 2) + (3 + 2 * 2) - 2", int64(15)},

		{"return 1.000000000000001 == 1", false},
		{"return 1.2 == 1", false},
		{"return 1.1 === 1", false},
		{"return 1.2 != 1", true},
		{"return 1.1 !== 1", true},
		{"return true == 1", true},
		{"return true == 2", false},
		{"return true == 0", false},
		{"return false == 0", true},
		{"return false == 1", false},
		{"return true == 5", false},
		{"return true && false", false},
		{"return true || false", true},
		{"return true || false && true", true},
		{"return (true || false) && true", true},
		{`return 1 == "1"`, false},
		{`return 1 == 1`, true},
		{`return 1 == true`, true},
		{`return 2 == true`, false},
		{`return 1 === true`, false},
		{`return 1 == 1.1`, false},
		{"return false || 3", 3},
		{"return 0 || 3", 3},
		{"return false ?? 3", false},
		{"return 0 ?? 3", 0},
		{"return null ?? 3", 3},
		{"return undefined ?? 3", 3},
		{"return 3 > 2", true},
		{"return 3 >= 2", true},
		{"return 1.2 > 1", true},
		{"return 1.2 >= 1", true},
		{"return 3 < 4", true},
		{"return 3 <= 4", true},
		{"return 1.2 < 1", false},
		{"return 1.2 <= 1", false},
		{"return 3 != 2", true},
		{"return 3 == 3", true},
		{"return !false", true},
		{"return !true", false},

		{"return 1 + null", 1},
		{"return 1 + undefined", 1},
		{"return 1 - null", 1},
		{"return 1 - undefined", 1},
		{"return 1 * null", 1},
		{"return 1 * undefined", 1},
		{`return "1" + null`, "1"},
		{`return "1" + undefined`, "1"},
		{`return '1' + null`, '1'},
		{`return '1' + undefined`, '1'},

		{"return null + 1", 1},
		{"return undefined + 1", 1},
		{"return null - 1", 1},
		{"return undefined - 1", 1},
		{"return null * 1", 1},
		{"return undefined * 1", 1},
		{"return null / 1", 0.0},
		{"return undefined / 1", 0.0},
		{"return null % 1", 0},
		{"return undefined % 1", 0},

		{"return 0 == undefined", false},
		{"return 0 == null", false},
		{"return 0 === undefined", false},
		{"return 0 === null", false},
		{"return undefined == 0", false},
		{"return null == 0", false},
		{"return undefined === 0", false},
		{"return null === 0", false},

		{"return 1 > undefined", true},
		{"return 1 >= undefined", true},
		{"return 1 > null", true},
		{"return 1 >= null", true},
		{"return 1.1 > undefined", true},
		{"return 1.1 >= undefined", true},
		{"return 1.1 > null", true},
		{"return 1.1 >= null", true},

		{"return undefined > 1", false},
		{"return undefined >= 1", false},
		{"return null > 1", false},
		{"return null >= 1", false},
		{"return undefined > 1.1", false},
		{"return undefined >= 1.1", false},
		{"return null > 1.1", false},
		{"return null >= 1.1", false},

		{"return 1 ==1.000000000000001 ", false},
		{"return 1 == 1.2", false},
		{"return 1 === 1.1", false},
		{"return 1 != 1.2", true},
		{"return 1 !== 1.1", true},
		{"return 1 == true", true},
		{"return 2 == true", false},
		{"return 0 == true", false},
		{"return 0 == false", true},
		{"return 1 == false", false},
		{"return 5 == true", false},
		{"return false && true", false},
		{"return false || true", true},
		{`return "1" == 1`, false},
		{`return 1 == 1`, true},
		{`return true == 1`, true},
		{`return true == 2`, false},
		{`return true === 1`, false},
		{`return 1.1 == 1`, false},
		{"return 3 || false", 3},
		{"return 3 || 0", 3},
		{"return 3 ?? false", 3},
		{"return 3 ?? 0", 3},
		{"return 3 ?? null", 3},
		{"return 3 ?? undefined", 3},
		{"return 2 > 3", false},
		{"return 2 >= 3", false},
		{"return 1 > 1.2", false},
		{"return 1 >= 1.2", false},
		{"return 4 < 3", false},
		{"return 4 <= 3", false},
		{"return 1 < 1.2", true},
		{"return 1 <= 1.2", true},

		{"let a = 0; return !a", true},
		{"let a = 0.1 - 0.1; return !a", true},

		{"return true ? 1 : 2", int64(1)},
		{"return false ? 1 : 2", int64(2)},
		{"return 1 == null ? 1 : 2", int64(2)},
		{"return 0xA + 0xB", int64(21)},
		{"return 0xAA ^ 0xBB", int64(17)},
		{"return 0xFF", int64(255)},
		{"return 1 | 2", int64(3)},
		{"return 1 | 5", int64(5)},
		{"return 3 ^ 6", int64(5)},
		{"return 3 & 6", int64(2)},
		{"return 50 >> 2", int64(12)},
		{"return 2 << 5", int64(64)},
		{"return 0010 << 1", int64(16)},

		{"return (a => 2)()", int64(2)},
		{"return (a => a)(2)", int64(2)},
		{"return (a => a + 1)(2)", int64(3)},
		{"return (a => () => a + 1)(2)()", int64(3)},

		{`
			let a = 1;
			a++;
			return a;`, int64(2)},
		{`
			let a = 1;
			a--;
			return a;`, int64(0)},
		{`
			let a = 1;
			a += 2;
			return a;`, int64(3)},
		{`
			let a = 1;
			a -= 2;
			return a;`, int64(-1)},
		{`
			let a = 2;
			a *= 2;
			return a;`, int64(4)},
		{`
			let a = 6;
			a /= 2;
			return a;`, float64(3),
		},
	}

	for _, d := range data {
		assertValue(t, d.expected, d.expression)
	}
}

func TestBitShiftPrecendence(t *testing.T) {
	assertValue(t, 5, `
		return 1 | 1 << 2
`)
}

func TestNullCoalesceDontEvalRight(t *testing.T) {
	assertValue(t, 0, `
		return 0 ?? foo()

		function foo() {
			throw "fail"
		}
`)
}

func TestLORDontEvalRight(t *testing.T) {
	assertValue(t, 1, `
		return 1 || foo()

		function foo() {
			throw "fail"
		}
`)
}

func TestMain(t *testing.T) {
	assertValue(t, 5, `
		function main() {
			return 2 + 3
		}
	`)
}

func TestLoopBasic(t *testing.T) {
	assertValue(t, 2, `
		function main() {
			let b = 0
			for(let i = 0; i < 2; i++) {
				b += 1
			}
			return b
		}
	`)
}

//Tests: For
func TestLoop0(t *testing.T) {
	assertValue(t, 10, `
		let a = 0;
		for (let i = 0; i < 10; i++) {
			a++
		}		
		return a
	`)
}

func TestLoop1(t *testing.T) {
	assertValue(t, 10, `
		let a = 0;
		for (let i = 0, l = 10; i < l; i++) {
			a++
		}		
		return a
	`)
}

func TestLoop2(t *testing.T) {
	assertValue(t, 3, `
		let a = 0;
		for (let i = 0; i < 10 && a < 3; i++) {
			a++
		}		
		return a
	`)
}

func TestLoop3(t *testing.T) {
	assertValue(t, 3, `
		let a = 0;
		let i = 0;
		
		function inc() { i++ }
			
		for (i = 0; i < 10 && a < 3; inc()) {
			a++
		}
		return a
	`)
}

func TestLoopLabel1(t *testing.T) {
	assertValue(t, 10, `
		let a = 0;
	
		foo:
		for(var k = 0; k < 5; k++) {
			a++;
			
			for(var i = 0; i < 5; i++) {	
				a++;
				 
				if(k > 3) {
					break foo
				}
				continue foo
				a++;
			}
		}
		return a
	`)
}

func TestLoopLabel2(t *testing.T) {
	assertValue(t, 8, `
		let a = 0
		let nums = [1,2,3,4,5]
	
		foo:
		for(var k of nums) {	
			a++;
			
			for(var i = 0; i < 5; i++) {	
				a++;
				
				if(k > 3) {
					break foo
				}
				continue foo
				a++;
			}
		}
		return a
	`)
}

func TestLoopLabel3(t *testing.T) {
	assertValue(t, 10, `
		let a = 0;
		let nums = [1,2,3,4,5]
	
		foo:
		for(var k in nums) {	
			a++;
			
			for(var i = 0; i < 5; i++) {	
				a++;
				
				if(k > 3) {
					break foo
				}
				continue foo
				a++;
			}
		}
		return a
	`)
}

func TestLoopLabel4(t *testing.T) {
	assertValue(t, 4, `
		let a = 0;
		foo:
       		for(;;) {	
			a++;

			for(var i = 0; i < 5; i++) {	
				a++;
				
				if(a > 3) {
					break foo
				}
				continue foo
				a++;
			}
		}
		return a
	`)
}

func TestLoopLabel5(t *testing.T) {
	assertValue(t, 4, `
		let a = 0;
		foo:
       		while(true) {	
			a++;

			for(var i = 0; i < 5; i++) {	
				a++;
				
				if(a > 3) {
					break foo
				}
				continue foo
				a++;
			}
		}
		return a
	`)
}

func TestLoopLabel6(t *testing.T) {
	assertValue(t, 3, `
			var i = 0;

			LABEL:
			for (var value of [1]) {
				try {
					i++
					throw "exception"
				}
				catch{
					i++
					break LABEL
				} finally {
					i++
				}
			}
			return i
	`)
}

func TestCall1(t *testing.T) {
	assertValue(t, 5, `
		function sum(a, b) {
			return a + b
		}
		function main() {
			return sum(2, 3)
		}
	`)
}

func TestCall2(t *testing.T) {
	assertValue(t, 2, `
		function foo(a) {
			return bar(a)
		}
		function bar(a) {
			return a + 1
		}
		function main() {
			return foo(1)
		}
	`)
}

func TestRet(t *testing.T) {
	assertValue(t, 3, `
		function foo() {
			let i = 0
			while(true) {
				if (i == 3) {
					return i
				}
				i++
			}
		}

		function main() {
			return foo()
		}
	`)
}

func TestFib(t *testing.T) {
	assertValue(t, 8, `
		function fib(n) {
			if (n < 2) {
				return n
			}
			return fib(n - 1) + fib(n - 2)
		}

		function main() {
			return fib(6)
		}
	`)
}

func TestStacktrace(t *testing.T) {
	p := compileTest(t, `
		function foo() {
			bar()
		}

		function bar() {
			throw "snap!"
		}

		function main() {
			foo()
		}
	`)

	vm := NewVM(p)
	_, err := vm.Run()

	se := normalize(`
		-> line 7
		-> line 3
		-> line 11
	`)

	if !strings.Contains(normalize(err.Error()), se) {
		t.Fatal(err)
	}
}

func TestStacktrace2(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "/main.ts", []byte(`
		import * as foo from "other/path/bar"

		function main() {
			return foo.bar()
		}
	`))

	filesystem.WritePath(fs, "/other/path/bar.ts", []byte(`
		export function bar() {
			return 1 / 0
		}
	`))

	p, err := Compile(fs, "main.ts")
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM(p)
	_, err = vm.Run()

	se := normalize(`
		-> /other/path/bar.ts:3
		-> /main.ts:5
	`)

	if !strings.Contains(normalize(err.Error()), se) {
		t.Fatal(err)
	}
}

func TestRepeatedSwitchValue(t *testing.T) {
	_, err := CompileStr(`
		let v = 0
		switch(v) {
		case 1:
			break
		case 1: 
			break
		}
	`)
	if err == nil || !strings.Contains(err.Error(), "Duplicate case") {
		t.Fatal(err)
	}
}
func TestRepeatedSwitchValue2(t *testing.T) {
	_, err := CompileStr(`
		let v = 0
		switch(v) {
		case "aa":
			break
		case "aa":
			break
		}
	`)
	if err == nil || !strings.Contains(err.Error(), "Duplicate case") {
		t.Fatal(err)
	}
}

func TestReturnFromScript(t *testing.T) {
	assertValue(t, 5, `
		return 5
	`)
}

func TestDivideByZero(t *testing.T) {
	data := []string{
		"let x = 1 / 0",
		"let x = 1 / null",
		"let x = 1 / undefined",
		"let x = 1 % 0",
		"let x = 1 % null",
		"let x = 1 % undefined",
	}

	for i, d := range data {
		p := compileTest(t, d)
		_, err := NewVM(p).Run()
		if err == nil || !strings.Contains(err.Error(), "divide by zero") {
			t.Fatalf("should throw excetion: %d", i)
		}
	}
}

func TestTry0(t *testing.T) {
	p := compileTest(t, `
		let x;		
		try {
			x = 1 / 0;
		} 
		finally {
			x = 5;
		}
	`)

	vm := NewVM(p)
	_, err := vm.Run()
	if err == nil || !strings.Contains(err.Error(), "divide by zero") {
		t.Fatal("should throw excetion")
	}

	v, _ := vm.RegisterValue("x")
	if v != NewValue(5) {
		t.Fatal(v)
	}
}

func TestTry1(t *testing.T) {
	assertValue(t, 5, `
		let x	
		try {	
			x = 1 / 0
		} catch {
			x = 5
		}
		return x
	`)
}

func TestTry2(t *testing.T) {
	assertValue(t, -2, `
		let x;		
		try {
			x = 1 / 0;
		} catch {
			x = -1;
		} finally {
			x -= 1;
		}
		return x
	`)
}

func TestTry3(t *testing.T) {
	assertValue(t, 0, `
		let x;		
		try {
			x = 1;
		} catch {
			x = -1;
		} finally {
			x -= 1;
		}
		return x
	`)
}

func TestTry4(t *testing.T) {
	p := compileTest(t, `
		let x;		
		try {
			x = 1 / 0;
		} catch {
			x = 1 / 0;
		} finally {
			x -= 1;
		}
		return x
	`)

	vm := NewVM(p)
	_, err := vm.Run()
	if err == nil || !strings.Contains(err.Error(), "divide by zero") {
		t.Fatal("should throw excetion")
	}
}

func TestTryFinally(t *testing.T) {
	assertValue(t, -3, `
		let x;		
		try {
			x = 0;
			try {
				try {
					// noop
				} catch {
					x = -1;
				} finally {
					x -= 1; // <---------
				}
			} catch {
				x = -1;
			} finally {
				x -= 1; // <---------
			}
		} catch {
			x = -1;
		} finally {
			x -= 1; // <---------
		}
		return x
	`)
}

func TestTryFinally2(t *testing.T) {
	assertValue(t, -3, `
		let x;		
		try {
			try {
				try {
					x = 0
				} finally {
					x -= 1; // <---------
				}
			} finally {
				x -= 1; // <---------
			}
		} finally {
			x -= 1; // <---------
		}
		return x
	`)
}

func TestTryFinally3(t *testing.T) {
	assertRegister(t, "x", -3, `
		let x;		
		try {
			try {
				try {
					x = 0;
					return;  // <--------------- exit
				} finally {
					x -= 1; // <---------
				}
			} finally {
				x -= 1; // <---------
			}
		} finally {
			x -= 1; // <---------
		}
	`)
}

func TestTryFinally4(t *testing.T) {
	assertRegister(t, "x", -3, `
		let x;		
		try {
			try {
				try {
					x = 0;
				} finally {
					x -= 1; // <--------- Must execute
					return;  // <--------------- exit
				}
			} finally {
				x -= 1; // <--------- Must execute
			}
		} finally {
			x -= 1; // <--------- Must execute
		}
	`)
}

func TestTryFinally5(t *testing.T) {
	assertRegister(t, "x", -3, `
		let x = 0	
		function foo() {
			try {
				try {
					try {
						x = 0
					} finally {
						x -= 1
						return // <--------------- exit
					}
				} finally {
					x -= 1 // <--------- Must execute
				}
			} finally {
				x -= 1 // <--------- Must execute
			}
		}
		
		function bar() {
			foo()
		}
		
		bar()	
	`)
}

func TestTryFinally6(t *testing.T) {
	assertRegister(t, "x", -3, `
		let x = 5;	
		function foo() {
			try {
				try {
					try {
						x = 1 / 0;
					} 
					catch(e) {
						x = 0;
					}
					finally {
						x -= 1; // <---------
						return;  // <--------------- exit
					}
				} finally {
					x -= 1; // <---------
				}
			} finally {
				x -= 1; // <---------
			}
		}
		
		function bar() {
			foo()
		}
		
		bar()	
	`)
}

// test that try and finally are in the same scope
func TestTryFinally8(t *testing.T) {
	assertValue(t, 1, `
		let x;
		
		function foo() {
	        try {
				// a L0
				let a = 1
	            return a;
	        }
	        finally {
				// b L1 because is in the same 
				// scope as the try body
	            let b = 2
	        }
		}
				
		return foo()
	`)
}

// test that a return before a finally is not overwritten
func TestTryFinally9(t *testing.T) {
	assertValue(t, 8, `
		function foo() {
		    try {
		        if (true) {
		            return bar(8)
		        }
		       	bar(11)
		    }
		    finally {
		        bar(7)
		    }
		}
		
		function bar(x) { return x }
				
		return foo()
	`)
}

// test that a return before a finally is not overwritten
//
// this tests is like the previous but changes the registers slightly.
func TestTryFinally10(t *testing.T) {
	assertValue(t, 8, `
		function foo() {
		    try {
		        if (true) {
		            return bar(8);
		        }
		    }
		    finally {
		        bar(7)
		    }
		}
		
		function bar(x) { return x }
				
		return foo()
	`)
}

// test that returning inside a finally has precedence
func TestTryFinally11(t *testing.T) {
	assertValue(t, 2, `
		function foo() {
		    try {
		        if (true) {
		            return bar(8);
		        }
		    }
		    finally {
		        bar(7)
				return 2
		    }
		}
		
		function bar(x) { return x }
				
		return foo()
	`)
}

// test that returning inside a finally has precedence
func TestTryFinally12(t *testing.T) {
	assertValue(t, 2, `
		function foo() {
		    try {
		        return bar(8);
		    }
		    finally {
				return 2
		    }
		}
		
		function bar(x) { return x }
				
		return foo()
	`)
}

// test that returning inside a finally has precedence
func TestTryFinally13(t *testing.T) {
	assertValue(t, 3, `
		function foo() {
		    try {
		        try {
			        return bar(8);
			    }
			    finally {
					return 2
			    }
		    }
		    finally {
				return 3
		    }
		}
		
		function bar(x) { return x }
				
		return foo()
	`)
}

// test that returning inside a finally has precedence
func TestTryFinally14(t *testing.T) {
	assertValue(t, 5, `
		function foo() {
		    try {
			    try {
			        try {
				        return bar(8);
				    }
				    finally {
						return 2
				    }
			    }
			    finally {
					return 3
			    }
		    }
		    finally {
				return 5
		    }
		}
		
		function bar(x) { return x }
				
		return foo()
	`)
}

func TestTryFinally15(t *testing.T) {
	assertValue(t, 1, `
		function foo() {
			let a = 0

			try {
				a++
			} finally {
				let b = 33
			}

			return a
		}

		function main() {
			return foo()
		}
	`)
}

func TestTryFinally16(t *testing.T) {
	assertValue(t, 1, `
		function main() {
			let a = 0

			try {
				a++
				let b = 33
			} catch(e) {
				let c = 44
			} finally {
				let d = 55
			}

			return a
		}
	`)
}

func TestTryFinally17(t *testing.T) {
	assertValue(t, 1, `
		function main() {
			let a = 0

			try {
				a++
				let b = 33
			} catch(e) {
				let c = 44
			} finally {
				return a
			}

			return 10
		}
	`)
}

func TestTryFinally19(t *testing.T) {
	assertValue(t, 33, `
		function foo() {
			let a = 0

			try {
				a++
			} finally {
				return 33
			}

			return 1
		}

		function main() {
			return foo()
		}
	`)
}

func TestTryFinally20(t *testing.T) {
	assertValue(t, 33, `
		function foo() {
			let a = 0

			try {
				a++
				return 22
			} finally {
				return 33
			}

			return 1
		}

		function main() {
			return foo()
		}
	`)
}

func TestTryFinally21(t *testing.T) {
	assertValue(t, 33, `
		function foo() {
			let a = 0

			try {
				a++
				return 11
			} catch {
				return 22
			} finally {
				return 33
			}

			return 1
		}

		function main() {
			return foo()
		}
	`)
}

func TestTryThrow1(t *testing.T) {
	p := compileTest(t, `
		throw "foo"			
	`)

	vm := NewVM(p)
	_, err := vm.Run()
	if err == nil {
		t.Fatal("Should fail")
	}
}

func TestTryThrow2(t *testing.T) {
	assertValue(t, 3, `
		let x = 1
		
		try {
		 	throw "foo";
		} 
		catch(e) {
			x += 1
		}
		finally {
			x += 1
		}	

		return x
	`)
}

func TestTryThrow3(t *testing.T) {
	assertValue(t, 3, `
		let x = 1;
		
		try {
		 	throw "foo";
			x = 2	
		} 
		catch(e) {		
			try {
			 	throw "foo";
			} 
			catch(e) {
				x = 3
			}
		}

		return x
	`)
}

// Check that the stackframe is restored
func TestTryThrow4(t *testing.T) {
	assertValue(t, "3", `
		let x;
		
		function foo() {
			foo2()
		}
		
		function foo2() {
			foo3()
		}
		
		function foo3() {
			throw "3"
		}
		
		try {		 	
			foo()	
		} 
		catch(e) {		
			x = e.message
		}

		return x
	`)
}

// Check that the stackframe is restored
func TestTryThrow5(t *testing.T) {
	p := compileTest(t, `
		let x
		
		function b() {
			try {
				throw "xx"
			}	
			finally {
				x = 10
			}
		}
	
		function a() {
			try {
				b()
			}			
			finally {
				x++
			}
		}
		
		a()
		return x
	`)

	vm := NewVM(p)
	_, err := vm.Run()
	if err == nil {
		t.Fatal("Should fail")
	}

	v, ok := vm.RegisterValue("x")
	if !ok {
		t.Fatal("Reg x not found")
	}

	if v != NewValue(11) {
		t.Fatal(v)
	}
}

func TestScope(t *testing.T) {
	ast, err := ParseStr(`
		function foo(a) {
			let a = 1
		}
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewCompiler().Compile(ast)
	if err == nil || !strings.Contains(err.Error(), "Redeclared identifier") {
		t.Fatal(err)
	}
}

func TestClosure0(t *testing.T) {
	assertValue(t, 3, `
		function foo(a) {
			return function () {
				return a
			}
		}
		return foo(3)()
	`)
}

func TestClosure01(t *testing.T) {
	assertValue(t, 1, `
		function main() {
			let a = 0			
			let foo = () => { a++ }
			foo() 
			return a
		}
	`)
}

func TestClosure02(t *testing.T) {
	assertValue(t, 1, `
		function main() {
			let a = 0			
			let foo = () => { 
				a++
				for(let i = 0; i < 10; i++) {
					try {
						a += 1 / 0
					} catch {
						a++
					} finally {
						a--
					}
				}
			}			
			foo() 
			a--
			foo()
			return a
		}
	`)
}

func TestClosure1(t *testing.T) {
	assertValue(t, 15, `
	function foo() {
		let a = 15;			
		return function () {
			return function () {
				return function () {
					return a
				}
			}
		}
	}
	return foo()()()()
	`)
}

func TestClosure2(t *testing.T) {
	assertValue(t, 3, `	
		function counter() {
			let i = 0
			return function() {
				i++
				return i
			}
		}		
		
		let next = counter()
		let r = next()
		r += next()		
		return r
	`)
}

func TestClosure3(t *testing.T) {
	assertValue(t, 3, `	
		function counter() {
			let i = 0
			return {
				fn: function() {
					i++
					return i
				}
			}
		}		
		
		let next = counter().fn
		let r = next()
		r += next()
		return r
	`)
}

func TestClosure4(t *testing.T) {
	assertValue(t, 3, `	
		function counter() {
			let i = 0;			
			return () => { i++; return i }
		}		
		
		let next = counter();			
		let r = next();
		r += next();		
		return r;
	`)
}

func TestClosure5(t *testing.T) {
	assertValue(t, 3, `	
		function newDev(a, b) {
			return { a: a, b: b }
		} 
		
		function bar(dev) {
			return dev.a
		}
		
		function newReader(a, b) {
			let dev = newDev(a, b)			
			return { foo: () => bar(dev) }
		}
	
		return newReader(3,2).foo();
	`)
}

func TestClosure6(t *testing.T) {
	assertValue(t, 3, `
		function counterWrap() {
			let f = counter();
			return f;
		}
			
		function counter() {
			let i = 0;			
			return () => { i++; return i }
		}		
		
		let next = counterWrap();			
		let r = next();
		r += next();		
		return r;
	`)
}

func TestClosure7(t *testing.T) {
	assertValue(t, 21, `
		function foo() {
			let a = 8;	
			let b = 5;			
			return function () {
				let c = 2;
				return function () {		
				let j = 6;			
					return function () {
						return a + b + j + c;
					}
				}
			}
		}
		return foo()()()();
	`)
}

func TestClassClosure1(t *testing.T) {
	assertValue(t, 9, `
		class Foo {
			powFunc(a) {
				return () => a * a 
			}
			sumFunc(a, b) {
				return () => a + b 
			}
		}

		let foo = new Foo()
		let a = foo.powFunc(2)()
		let b = foo.sumFunc(2, 3)()
		return a + b
	`)
}

func TestClassClosure2(t *testing.T) {
	assertValue(t, 10, `
		class Foo {
			z;
			constructor(z) {
				this.z = z
			}
			powFunc(a) {
				return () => a * a 
			}
			sumFunc(a, b) {
				return () => a + b 
			}
		}

		let foo = new Foo(1)
		let a = foo.powFunc(2)()
		let b = foo.sumFunc(2, 3)()
		return a + b + foo.z // 1 + 4 + 5
	`)
}

func TestConstant1(t *testing.T) {
	assertValue(t, 33, `
		const a = 33
		return a
	`)
}

func TestConstant2(t *testing.T) {
	assertValue(t, 33, `
		const a = 33
		let b = a
		return a
	`)
}

func TestConstant3(t *testing.T) {
	assertValue(t, 33, `
		const a = 33
		function main() {
			return a
		}
	`)
}

func TestConstant4(t *testing.T) {
	_, err := CompileStr(`
		const a = 33
		a++
	`)

	if err == nil || !strings.Contains(err.Error(), "can't modify a constant") {
		t.Fatal("can't modify a constant", err)
	}
}

func TestConstant5(t *testing.T) {
	_, err := CompileStr(`
		const a = 33
		a--
	`)

	if err == nil || !strings.Contains(err.Error(), "can't modify a constant") {
		t.Fatal("can't modify a constant", err)
	}
}

func TestConstant6(t *testing.T) {
	_, err := CompileStr(`
		const a = 33
		a *= 8
	`)

	if err == nil || !strings.Contains(err.Error(), "can't modify a constant") {
		t.Fatal("can't modify a constant", err)
	}
}

func TestConstant7(t *testing.T) {
	_, err := CompileStr(`
		const a = 33
		a /= 8
	`)

	if err == nil || !strings.Contains(err.Error(), "can't modify a constant") {
		t.Fatal("can't modify a constant", err)
	}
}

func TestConstant8(t *testing.T) {
	_, err := CompileStr(`
		const a = 33
		a = 22
	`)

	if err == nil || !strings.Contains(err.Error(), "can't modify a constant") {
		t.Fatal("can't modify a constant", err)
	}
}

func TestOptionalChaining1(t *testing.T) {
	assertValue(t, 2, `
		let a
		return a?.b || 2
	`)
}

func TestOptionalChaining2(t *testing.T) {
	assertValue(t, nil, `
		let a
		return a?.()
	`)
}

func TestOptionalChaining3(t *testing.T) {
	assertValue(t, 2, `
		let a
		return a?.[0] || 2
	`)
}

func TestOptionalChaining4(t *testing.T) {
	assertValue(t, 2, `
		let a
		return a?.() || 2
	`)
}

func TestOptionalChaining5(t *testing.T) {
	assertValue(t, 2, `
		let a
		return a?.[0] || 2
	`)
}

func TestOptionalChaining6(t *testing.T) {
	libs := []NativeFunction{
		{
			Name: "String.prototype.toUpper",
			Function: func(this Value, args []Value, vm *VM) (Value, error) {
				s := strings.ToUpper(this.String())
				return NewString(s), nil
			},
		},
	}

	assertNativeValue(t, libs, "FOO", `
		let a = { b: { fn: v => v } }
		return a?.b?.fn?.("foo").toUpper()
	`)
}

func TestOptionalChaining7(t *testing.T) {
	libs := []NativeFunction{
		{
			Name: "String.prototype.toUpper",
			Function: func(this Value, args []Value, vm *VM) (Value, error) {
				s := strings.ToUpper(this.String())
				return NewString(s), nil
			},
		},
	}

	assertNativeValue(t, libs, nil, `
		let a = { b: null }
		return a?.b?.fn?.("foo").toUpper()
	`)
}

func TestOptionalChaining8(t *testing.T) {
	assertValue(t, nil, `
		let a 
		return a?.fn?.().foo()
	`)
}

func TestOptionalChaining9(t *testing.T) {
	assertValue(t, 2, `
		let a = { fn: v => v, b: 2 }
		return a?.fn?.(a?.x || a.b)
	`)
}

func TestOptionalChaining10(t *testing.T) {
	assertValue(t, nil, `
		let a = { b: v => v + 1 }
		return a.bb?.(3)
	`)
}

func TestOptionalChaining11(t *testing.T) {
	assertValue(t, nil, `
		let a = {}
		return a?.(33)
	`)
}

func TestOptionalChaining12(t *testing.T) {
	assertValue(t, nil, `
		let a = "foo"
		return a?.bar()
	`)
}

func TestOptionalChaining13(t *testing.T) {
	assertValue(t, nil, `
		let a = null
		return a?.bar()
	`)
}

func TestOptionalChaining14(t *testing.T) {
	assertValue(t, nil, `
		let a
		return a?.bar()
	`)
}

func TestOptionalChainLoop(t *testing.T) {
	// check that the old value is overwriten with null
	assertValue(t, nil, `
		let arr = [
			{ a: { b: 2 } },
			{ a: null },
		]

		let a
		for (let v of arr) {
			a = v.a?.b
		}
		return a
	`)
}

func TestOptionalChainLoop2(t *testing.T) {
	// check that the old value is overwriten with null
	assertValue(t, nil, `
	let arr = [
		{ a: { b: {} } },
		{ a: null },
	]

	let a
	for (let v of arr) {
		a = v.a?.b?.c?.d?.e
	}
	return a
	`)
}

func TestOptionalChainLoop3(t *testing.T) {
	// check that the old value is overwriten with null
	assertValue(t, nil, `
		let arr = [
			{ b: "c" },
			{ b: null },
		]
		let a = 3
		for (let v of arr) {
			a = v.b?.trim()
		}
		return a
	`)
}

func TestOptionalChainLoop4(t *testing.T) {
	// check that the old value is overwriten with null
	assertValue(t, nil, `
		let arr = [
			{ b: "c" },
			{ b: null },
		]
		let a = 3
		for (let v of arr) {
			a = v.b?.trim().trim()
		}
		return a
	`)
}

func TestOptionalChainLoop5(t *testing.T) {
	// check that the old value is overwriten with null
	assertValue(t, nil, `
		let arr = [
			{ b: [1] },
			{ b: null },
		]
		let a = 3
		for (let v of arr) {
			a = v.b?.[0]
		}
		return a
	`)
}

func TestOptionalChainLoop6(t *testing.T) {
	// check that the old value is overwriten with null
	assertValue(t, nil, `
		let arr = [
			{ b: [[1]] },
			{ b: null },
		]
		let a = 3
		for (let v of arr) {
			a = v.b?.[0][0]
		}
		return a
	`)
}

func TestOptionalChainLoop7(t *testing.T) {
	// check that the old value is overwriten with null
	assertValue(t, nil, `
		let arr = [
			{ b: [[1]] },
			{ b: [null] },
		]
		let a = 3
		for (let v of arr) {
			a = v.b?.[0]?.[0]
		}
		return a
	`)
}

// Tests: Enum
func TestEnum1(t *testing.T) {
	assertValue(t, 4, `
		return Direction.Right
		enum Direction {
		    Up = 1,
		    Down,
		    Left,
		    Right
		}
	`)
}

func TestEnum2(t *testing.T) {
	assertValue(t, 3, `
		enum Direction {
		    Up,
		    Down,
		    Left,
		    Right
		}
		return Direction.Right
	`)
}

func TestEnum3(t *testing.T) {
	assertValue(t, 5, `
		enum Direction {
		    Up = 5,
		    Down,
		    Left,
		    Right
		}
		return Direction.Up
	`)
}

func TestEnum4(t *testing.T) {
	assertValue(t, 8, `
		enum Direction {
		    Up = 5,
		    Down,
		    Left,
		    Right
		}
		return Direction.Right
	`)
}

func TestEnum5(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as foo from "foo"
		function main() {
			return foo.Direction.Right
		}
	`))

	filesystem.WritePath(fs, "foo.ts", []byte(`
		export enum Direction {
			Right = 5,
		}
	`))

	assertValueFS(t, fs, "main.ts", 5)
}

func TestEnum6(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as foo from "foo"
		function main() {
			return foo.bar()
		}
	`))

	filesystem.WritePath(fs, "foo.ts", []byte(`
		export function bar() {
			return Direction.Right
		}
		enum Direction {
			Right = 5,
		}
	`))

	assertValueFS(t, fs, "main.ts", 5)
}

func TestEnum7(t *testing.T) {
	assertValue(t, 5, `
		enum Direction {
			Up = 5,
			Down
		}
		let a = Direction
		return a.Up
	`)
}

func TestEnum8(t *testing.T) {
	assertValue(t, 3, `
		enum Direction {
			Up = 1,
			Down
		}

		let b = 0
		for(let key in Direction) {
			b += Direction[key]
		}

		return b
	`)
}

func TestEnum9(t *testing.T) {
	assertValue(t, 3, `
		enum Direction {
			Up = 1,
			Down
		}

		let b = 0
		let a = Direction
		for(let key in a) {
			b += a[key]
		}

		return b
	`)
}

func TestEnum10(t *testing.T) {
	assertValue(t, 5, `
		enum Direction {
			Up = 5,
			Down
		}
		return Direction["Up"]
	`)
}

func TestEnumParser(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as lib1 from "lib1"

		function main() {
			return lib1.foo()
		}
	`))

	filesystem.WritePath(fs, "lib1.ts", []byte(`
		import * as lib2 from "lib2"
		
		export function foo() {
			return lib2.FileAccess.tmp
		}
	`))

	filesystem.WritePath(fs, "lib2.ts", []byte(`
		export enum FileAccess {
			private,
			tmp,
			public
		}
	`))

	assertValueFS(t, fs, "main.ts", 1)
}

func TestEnumString(t *testing.T) {
	assertValue(t, "up", `
		enum Direction {
		    Up = "up",
		    Down = "down"
		}
		return Direction.Up
	`)
}

// Tests: Error
func TestError(t *testing.T) {
	assertValue(t, "Attempt to divide by zero", `
		let x;		
		try {
			x = 1 / 0;
		} catch(e) {
			x = e.message
		}
		return x
	`)
}

func TestMap(t *testing.T) {
	assertValue(t, 5, `
		let a = {}
		a[null] = 5
		return a[null]
	`)
}

func TestDelete(t *testing.T) {
	assertValue(t, true, `
		let a = { foo: 1 }
		delete a.foo
		return a.foo == undefined
	`)
}

// func TestTypeof1(t *testing.T) {
// 	assertValue(t, "number", `
// 		let a = 1
// 		return typeof a
// 	`)
// }

// func TestTypeof2(t *testing.T) {
// 	assertValue(t, "number", `
// 		let a = 1.1
// 		return typeof a
// 	`)
// }

// func TestTypeof3(t *testing.T) {
// 	assertValue(t, "string", `
// 		let a = ""
// 		return typeof a
// 	`)
// }
// func TestTypeof4(t *testing.T) {
// 	assertValue(t, true, `
// 		let a = ""
// 		return typeof a == "string" || 33
// 	`)
// }

func TestClass0(t *testing.T) {
	assertValue(t, "John", `
		return new Person("John").getName()

		class Person {
			name: string
		
			constructor(name: string) {
				this.name = name
			}
		
			getName() {
				return this.name
			}
		}
	`)
}

func TestClass1(t *testing.T) {
	assertValue(t, 2, `
		class Foo {
			getNum() {
				return 2
			}
		}
		return new Foo().getNum()
	`)
}

func TestClass2(t *testing.T) {
	assertValue(t, 2, `
		class Foo {
			bar: number
			constructor(x: number) {
				this.bar = x
			}
			getNum() {
				return this.bar
			}
		}

		return new Foo(2).getNum()
	`)
}

func TestClass3(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as foo from "foo"
		function main() {
			return new foo.Foo().getNum()
		}
	`))

	filesystem.WritePath(fs, "foo.ts", []byte(`
		export class Foo {
			getNum() {
				return 5
			}
		}
	`))

	assertValueFS(t, fs, "main.ts", 5)
}
func TestClass4(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as foo from "foo"
		function main() {
			return new foo.Foo().getNum()
		}
	`))

	filesystem.WritePath(fs, "foo.ts", []byte(`
		class Foo {
			getNum() {
				return 5
			}
		}
	`))

	_, err := Compile(fs, "main.ts")
	if err == nil || !strings.Contains(err.Error(), "Undeclared identifier") {
		t.Fatal(err)
	}
}

func TestClass5(t *testing.T) {
	_, err := CompileStr(`
		let a = Foo

		class Foo {
			getNum() {
				return 5
			}
		}
	`)
	if err == nil || !strings.Contains(err.Error(), "invalid value") {
		t.Fatal(err)
	}
}

func TestClassPrivateField(t *testing.T) {
	p := compileTest(t, `
			class Foo {
				private a = 3
			}
			return new Foo().a
		`)
	vm := NewVM(p)

	// Print(p)

	_, err := vm.Run()
	if err == nil || !strings.Contains(err.Error(), "private field") {
		t.Fatal(err)
	}
}

func TestClassSetPrivateField(t *testing.T) {
	p := compileTest(t, `
			class Foo {
				private a
			}
			let foo = new Foo()
			foo.a = 3
		`)
	vm := NewVM(p)

	// Print(p)

	_, err := vm.Run()
	if err == nil || !strings.Contains(err.Error(), "private field") {
		t.Fatal(err)
	}
}

func TestClassPrivateMethod(t *testing.T) {
	p := compileTest(t, `
			class Foo {
				private bar() {
				}
			}
			return new Foo().bar()
		`)
	vm := NewVM(p)

	// Print(p)

	_, err := vm.Run()
	if err == nil || !strings.Contains(err.Error(), "private method") {
		t.Fatal(err)
	}
}

func TestClassInitializeFields(t *testing.T) {
	assertValue(t, 2, `
		class Foo {
			bar: number = 2
		}
		return new Foo(2).bar
	`)
}

func TestModuleImports1(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as foo from "foo"

		function main() {
			return foo.bar
		}
	`))

	filesystem.WritePath(fs, "foo.ts", []byte(`
		export const bar = 3
	`))

	assertValueFS(t, fs, "main.ts", 3)
}

func TestModuleImports2(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "/dir1/main.ts", []byte(`
		import * as foo from "/libs/foo"

		function main() {
			return foo.bar
		}
	`))

	filesystem.WritePath(fs, "/libs/foo.ts", []byte(`
		export const bar = 3
	`))

	assertValueFS(t, fs, "/dir1/main.ts", 3)
}

func TestModuleImports3(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "/dir1/main.ts", []byte(`
		import * as foo from "../libs/foo"

		function main() {
			return foo.bar
		}
	`))

	filesystem.WritePath(fs, "/libs/foo.ts", []byte(`
		export const bar = 3
	`))

	assertValueFS(t, fs, "/dir1/main.ts", 3)
}

func TestModuleImports4(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "/dir1/main.ts", []byte(`
		import * as foo from "../dir1/dir2/dir3/foo"

		function main() {
			return foo.bar
		}
	`))

	filesystem.WritePath(fs, "/dir1/dir2/dir3/foo.ts", []byte(`
		import * as xxx from "../../../other/path/bar"

		export const bar = xxx.foo()
	`))

	filesystem.WritePath(fs, "/other/path/bar.ts", []byte(`
		export function foo() {
			return 3
		}
	`))

	assertValueFS(t, fs, "/dir1/main.ts", 3)
}

func TestModuleImportsRelativeToModule(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "/dir1/main.ts", []byte(`
		import * as foo from "../dir1/foo"

		function main() {
			return foo.bar
		}
	`))

	filesystem.WritePath(fs, "/dir1/foo.ts", []byte(`
		import * as xxx from "../dir2/bar"

		export const bar = xxx.foo()
	`))

	filesystem.WritePath(fs, "/dir2/bar.ts", []byte(`
		export function foo() {
			return 3
		}
	`))

	assertValueFS(t, fs, "/dir1/main.ts", 3)
}

func TestModuleImports5(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "/dir1/main.ts", []byte(`
		import * as foo from "../other/path/bar"

		function main() {
			return new foo.Foo(3).bar
		}
	`))

	filesystem.WritePath(fs, "/other/path/bar.ts", []byte(`
		export class Foo {
			bar: number
			constructor(x: number) {
				this.bar = x
			}
			getNum() {
				return this.bar
			}
		}
	`))

	assertValueFS(t, fs, "/dir1/main.ts", 3)
}

func TestModuleImports6(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "/dir1/main.ts", []byte(`
		import * as foo from "bar"

		function main() {
			return new foo.Foo(3).bar
		}
	`))

	filesystem.WritePath(fs, "/dir1/bar.ts", []byte(`
		export class Foo {
			bar: number
			constructor(x: number) {
				this.bar = x
			}
			getNum() {
				return this.bar
			}
		}
	`))

	assertValueFS(t, fs, "/dir1/main.ts", 3)
}

func TestModuleImports7(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "/dir1/main.ts", []byte(`
		import * as foo from "./bar"

		function main() {
			return new foo.Foo(3).bar
		}
	`))

	filesystem.WritePath(fs, "/dir1/bar.ts", []byte(`
		export class Foo {
			bar: number
			constructor(x: number) {
				this.bar = x
			}
			getNum() {
				return this.bar
			}
		}
	`))

	assertValueFS(t, fs, "/dir1/main.ts", 3)
}

func TestModuleImportsFromOtherDir(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "/dir1/main.ts", []byte(`
		import * as foo from "./bar"

		function main() {
			return foo.bar()
		}
	`))

	filesystem.WritePath(fs, "/dir1/bar.ts", []byte(`
		export function bar() {
			return 3
		}
	`))

	fs.MkdirAll("/foo/bar")

	if err := fs.Chdir("/foo/bar"); err != nil {
		t.Fatal(err)
	}

	assertValueFS(t, fs, "../../dir1/main.ts", 3)
}

func TestModuleForSideEffects(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import "foo"

		function main() {
			return bar
		}
	`))

	filesystem.WritePath(fs, "foo.ts", []byte(`
		export const bar = 3
	`))

	_, err := Compile(fs, "main.ts")
	if err == nil || !strings.Contains(err.Error(), "Undeclared identifier") {
		t.Fatal(err)
	}
}

func TestModuleForSideEffects2(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import "foo"

		function main() {
			bar()
		}
	`))

	filesystem.WritePath(fs, "foo.ts", []byte(`
		export function bar() {}
	`))

	_, err := Compile(fs, "main.ts")
	if err == nil || !strings.Contains(err.Error(), "Undeclared identifier") {
		t.Fatal(err)
	}
}

func TestVisibility(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as foo from "foo"

		function main() {
			foo.bar()
		}
	`))

	filesystem.WritePath(fs, "foo.ts", []byte(`
		function bar() {}
	`))

	_, err := Compile(fs, "main.ts")
	if err == nil || !strings.Contains(err.Error(), "not exported") {
		t.Fatal(err)
	}
}

func TestModuleSameNames(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as foo from "bar"

		export function sum() {
			return 88
		}

		function main() {
			return foo.bar()
		}
	`))

	filesystem.WritePath(fs, "bar.ts", []byte(`
		export function bar() {
			return sum()
		}
		export function sum() {
			return 3
		}
	`))

	assertValueFS(t, fs, "main.ts", 3)
}

func TestVendor(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as foo from "foo"
		function main() {
			return foo.bar()
		}
	`))

	filesystem.WritePath(fs, "vendor/foo.ts", []byte(`
		import * as xxx from "zzz/xxx"
		export function bar() {
			return xxx.Z
		}
	`))

	filesystem.WritePath(fs, "vendor/zzz/xxx.ts", []byte(`
		export const Z = 3
	`))

	assertValueFS(t, fs, "main.ts", 3)
}

func TestVendorRoot(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "dir/tsconfig.json", []byte(`
		{
			"compilerOptions": {
				"baseUrl": ".",
				"paths": {
					"*": [
						"*",
						"vendor/*"
					]
				}
			}
		}
	`))

	filesystem.WritePath(fs, "dir/main.ts", []byte(`
		import * as foo from "foo"
		function main() {
			return foo.bar()
		}
	`))

	filesystem.WritePath(fs, "dir/vendor/foo.ts", []byte(`
		import * as xxx from "zzz/xxx"
		export function bar() {
			return xxx.Z
		}
	`))

	filesystem.WritePath(fs, "dir/vendor/zzz/xxx.ts", []byte(`
		export const Z = 3
	`))

	assertValueFS(t, fs, "dir/main.ts", 3)
}

func TestInit0(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as lib1 from "libs/lib1"

		export let v

		function init() {
			v = 1
		}

		function main() {
			return v + lib1.v
		}
	`))

	filesystem.WritePath(fs, "libs/lib1.ts", []byte(`		
		export let v

		function init() {
			v = 2
		}
	`))

	assertValueFS(t, fs, "main.ts", 3)
}

func TestInit1(t *testing.T) {
	fs := filesystem.NewMemFS()
	filesystem.WritePath(fs, "main.ts", []byte(`
		import * as lib1 from "libs/lib1"
		import * as lib2 from "libs/lib2"

		export let v

		function init() {
			v = 0
		}

		function main() {
			return v + lib1.foo() + lib2.v
		}
	`))

	filesystem.WritePath(fs, "libs/lib1.ts", []byte(`
		import * as lib2 from "libs/lib2"
		
		export let v

		function init() {
			v = 1
		}

		export function foo() {
			return v + lib2.v
		}
	`))

	filesystem.WritePath(fs, "libs/lib2.ts", []byte(`		
		export let v

		function init() {
			v = 2
		}
	`))

	assertValueFS(t, fs, "main.ts", 5)
}

func TestParseInterface(t *testing.T) {
	_, err := ParseStr(`
			interface Foo {
				bar: string,
				test: [
					{ a: number, b: number }
				]
			}
		`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGlobalScope(t *testing.T) {
	fs := filesystem.NewMemFS()

	filesystem.WritePath(fs, "global.ts", []byte(`
		declare global {
			const PI = 3
		
			enum Foo {
				bar = 1
			}
		}	
	`))

	filesystem.WritePath(fs, "main.ts", []byte(`
		import "global"

		function main() {
			return PI + Foo.bar
		}
	`))

	assertValueFS(t, fs, "main.ts", 4)
}

func TestProgramPermissions1(t *testing.T) {
	p, err := CompileStr(`
		// [permissions foo]
	`)

	if err != nil {
		t.Fatal(err)
	}

	perm := p.Permissions()
	if len(perm) != 1 {
		t.Fatal(perm)
	}

	if perm[0] != "foo" {
		t.Fatal(perm)
	}

	if !p.HasPermission("foo") {
		t.Fatal(perm)
	}

	if p.HasPermission("bar") {
		t.Fatal("Invalid perm")
	}
}

func TestAttributes(t *testing.T) {
	p, err := CompileStr(`
		// [anonymous]
		export function index_GET(c: WebContext) {
			c.response.write("Hello world!! \n")
		}

	`)

	if err != nil {
		t.Fatal(err)
	}

	if len(p.Functions) != 2 {
		t.Fatal()
	}

	//fmt.Println(p.Functions[1].Attributes)

}

// func TestInline1(t *testing.T) {
// 	assertInlineValue(t, 6000, `
// 		function main() {
// 			let v = 0
// 			for(let i = 0; i < 1000; i++) {
// 				v += bar() + bar()
// 			}
// 			return v
// 		}

// 		function bar() {
// 			return 2 + 1
// 		}
// 	`)
// }

// func TestInline2(t *testing.T) {
// 	assertInlineValue(t, 3, `
// 		function main() {
// 			return bar(1, 2)
// 		}

// 		function bar(a, b) {
// 			return a + b
// 		}
// 	`)
// }

// func TestInline3(t *testing.T) {
// 	assertInlineValue(t, 3000, `
// 		function main() {
// 			let v = 0
// 			for(let i = 0; i < 1000; i++) {
// 				v += bar(1, 2)
// 			}
// 			return v
// 		}

// 		function bar(a, b) {
// 			return a + b
// 		}
// 	`)
// }

// func TestInline4(t *testing.T) {
// 	assertInlineValue(t, 3, `
// 		return bar(3)

// 		function bar(a) {
// 			return a
// 		}
// 	`)
// }

// func TestInline5(t *testing.T) {
// 	assertInlineValue(t, 3, `
// 		return bar(1, 2)

// 		function bar(a, b) {
// 			return a + b
// 		}
// 	`)
// }

// func TestInline6(t *testing.T) {
// 	assertInlineValue(t, 103, `
// 		let x = 100
// 		let dummy = 33
// 		let dummy2 = 33
// 		let dummy3 = 33
// 		let result = x + bar(1, 2)
// 		let dummy4 = 33
// 		let dummy5 = 33
// 		return result

// 		function bar(a, b) {
// 			return a + b
// 		}
// 	`)
// }

// func TestInline7(t *testing.T) {
// 	assertInlineValue(t, 103, `
// 		function main() {
// 			let x = 100
// 			let dummy = 33
// 			let dummy2 = 33
// 			let result = x + bar(1, 2)
// 			let dummy3 = 33
// 			let dummy4 = 33
// 			return result
// 		}

// 		function bar(a, b) {
// 			let dummy = -70
// 			let z = a + b
// 			let dummy2 = -70
// 			return z
// 		}
// 	`)
// }

// func TestInline8(t *testing.T) {
// 	fs := filesystem.NewMemFS()
// 	filesystem.WritePath(fs, "main.ts", []byte(`
// 		import * as foo from "foo"

// 		function main() {
// 			return foo.bar()
// 		}
// 	`))

// 	filesystem.WritePath(fs, "foo.ts", []byte(`
// 		export function bar() {
// 			return foo()
// 		}

// 		export function foo() {
// 			return 5
// 		}
// 	`))

// 	assertInlineFSValue(t, 5, fs)
// }

// func TestInlineOptionalParams1(t *testing.T) {
// 	assertInlineValue(t, 3, `
// 		return bar()

// 		function bar(a?) {
// 			return 3
// 		}
// 	`)
// }

func TestTailCall1(t *testing.T) {
	assertInlineValue(t, 6, `
		return fact(3)

		function fact(n, a?) {
			if(a == null) {
				a = 1
			}

			if(n == 0) {
				return a
			}

        	return fact(n - 1, n * a);
		}
	`)
}

func TestTailCall2(t *testing.T) {
	assertInlineValue(t, 6, `
		let a = 0
		foo()
		return a

		function foo() {
			a++
			if(a > 5) {
				return a
			}			
        	return foo()
		}
	`)
}

func assertNoCalls(t *testing.T, p *Program) {
	// ignore the global func
	for _, f := range p.Functions[1:] {
		for i, instr := range f.Instructions {
			switch instr.Opcode {
			case op_call, op_callSingleArg:
				Print(p)
				t.Fatalf("There is a call: %s, %d, %v", f.Name, i, instr)
			}
		}
	}
}

func TestNativeFunc(t *testing.T) {
	libs := []NativeFunction{
		{
			Name:      "math.square",
			Arguments: 1,
			Function: func(this Value, args []Value, vm *VM) (Value, error) {
				v := args[0].ToInt()
				return NewInt64(v * v), nil
			},
		},
	}

	assertNativeValue(t, libs, 4, `
		function main() {
			return math.square(2)
		}
	`)
}

func TestNativeProperty(t *testing.T) {
	libs := []NativeFunction{
		{
			Name: "->math.pi",
			Function: func(this Value, args []Value, vm *VM) (Value, error) {
				return NewFloat(3.1416), nil
			},
		},
	}

	assertNativeValue(t, libs, 3.1416, `
		function main() {
			return math.pi
		}
	`)
}

func TestNativeObject(t *testing.T) {
	libs := []NativeFunction{
		{
			Name: "tests.newObject",
			Function: func(this Value, args []Value, vm *VM) (Value, error) {
				return NewObject(obj{}), nil
			},
		},
	}

	assertNativeValue(t, libs, "Hi foo", `
		function main() {
			let obj = tests.newObject()
			return obj.sayHi(obj.name)
		}
	`)
}

func TestNativeFuncError(t *testing.T) {
	libs := []NativeFunction{
		{
			Name:      "math.square",
			Arguments: 1,
			Function: func(this Value, args []Value, vm *VM) (Value, error) {
				return NullValue, fmt.Errorf("snap!")
			},
		},
		{
			Name:      "log.error",
			Arguments: -1,
			Function: func(this Value, args []Value, vm *VM) (Value, error) {
				return NullValue, nil
			},
		},
	}

	assertNativeValue(t, libs, nil, `
		function main() {
			try {
				math.square(2)				
			} catch (error) {
				log.error("asdfas", error)
			}
		}
	`)
}

type obj struct{}

func (d obj) GetProperty(key string, vm *VM) (Value, error) {
	switch key {
	case "name":
		return NewString("foo"), nil
	}
	return UndefinedValue, nil
}

func (d obj) GetMethod(name string) NativeMethod {
	switch name {
	case "sayHi":
		return d.sayHI
	}
	return nil
}

func (d obj) sayHI(args []Value, vm *VM) (Value, error) {
	return NewString("Hi " + args[0].String()), nil
}

// func addPrintFunc() {
// 	AddBuiltinFunc("print")
// 	AddNativeFunc(NativeFunction{
// 		Name:      "print",
// 		Arguments: -1,
// 		Function: func(this Value, args []Value, vm *VM) (Value, error) {
// 			for _, v := range args {
// 				var text string
// 				switch v.Type {
// 				case String, Int, Float, Bool:
// 					text = v.ToString()
// 				default:
// 					b, err := json.MarshalIndent(v.Export(0), "", "    ")
// 					if err != nil {
// 						return NullValue, err
// 					}
// 					text = string(b)
// 				}

// 				fmt.Print(text)
// 				fmt.Print(" ")
// 			}
// 			fmt.Print("\n")
// 			return NullValue, nil
// 		},
// 	})
// }

func assertNativeValue(t *testing.T, funcs []NativeFunction, expected interface{}, code string) {
	a, err := ParseStr(code)
	if err != nil {
		t.Fatal(err)
	}

	c := NewCompiler()

	for _, f := range funcs {
		AddNativeFunc(f)
	}

	p, err := c.Compile(a)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM(p)

	// Print(p)
	// vm.MaxSteps = 10

	ret, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}

	v := NewValue(expected)

	if ret != v {
		t.Fatalf("Expected %v %T, got %v %T", expected, expected, ret, ret)
	}
}

func assertValueFS(t *testing.T, fs filesystem.FS, path string, expected interface{}) {
	p, err := Compile(fs, path)
	if err != nil {
		t.Fatal(err)
	}

	// Print(p)
	// PrintNames(p, true)
	// vm.MaxSteps = 50

	vm := NewVM(p)

	ret, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}

	v := NewValue(expected)

	if ret != v {
		t.Fatalf("Expected %v %T, got %v %T", expected, expected, ret, ret)
	}
}

func assertInlineValue(t *testing.T, expected interface{}, code string) {
	a, err := ParseStr(code)
	if err != nil {
		t.Fatal(err)
	}

	c := NewCompiler()

	p, err := c.Compile(a)
	if err != nil {
		t.Fatal(err)
	}

	assertNoCalls(t, p)

	// Print(p)

	vm := NewVM(p)
	vm.MaxSteps = 10000000

	ret, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}

	if ret != NewValue(expected) {
		t.Fatalf("Expected %v %T, got %v", expected, expected, ret.String())
	}
}

// func assertInlineFSValue(t *testing.T, expected interface{}, fs filesystem.FS) {
// 	a, err := Parse(fs, "main.ts")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	c := NewCompiler()

// 	p, err := c.Compile(a)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	assertNoCalls(t, p)

// 	// Print(p)

// 	vm := NewVM(p)
// 	vm.MaxSteps = 10000000

// 	ret, err := vm.Run()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if ret != NewValue(expected) {
// 		t.Fatalf("Expected %v %T, got %v", expected, expected, ret.ToString())
// 	}
// }

func assertValue(t *testing.T, expected interface{}, code string) {
	p := compileTest(t, code)
	vm := NewVM(p)

	// Print(p)
	// vm.MaxSteps = 50

	ret, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}

	if ret != NewValue(expected) {
		t.Fatalf("Expected %v %T, got %v", expected, expected, ret.String())
	}
}

func assertRegister(t *testing.T, register string, expected interface{}, code string) {
	p := compileTest(t, code)
	vm := NewVM(p)

	// Print(p)
	// vm.MaxSteps = 50

	_, err := vm.Run()
	if err != nil {
		t.Fatal(err)
	}

	v, _ := vm.RegisterValue(register)

	if v != NewValue(expected) {
		t.Fatalf("Expected %v, got %v", expected, v)
	}
}

func compileTest(t *testing.T, code string) *Program {
	p, err := CompileStr(code)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func normalize(s string) string {
	var reg = regexp.MustCompile(`\s+`)
	s = reg.ReplaceAllString(s, ` `)
	return s
}

func TestCopy(t *testing.T) {
	p1, err := CompileStr(`
	let a = 10

	function init() {
		a++
	}

	function foo() {
		a++
		return a
	}`)
	if err != nil {
		t.Fatal(err)
	}

	p1 = p1.Copy()

	vm := NewVM(p1)
	if err := vm.Initialize(); err != nil {
		t.Fatal(err)
	}

	v, err := vm.RunFunc("foo")
	if err != nil {
		t.Fatal(err)
	}

	if v.ToInt() != 12 {
		t.Fatal(v)
	}
}

func TestEval0(t *testing.T) {
	p, err := CompileStr(`
		let a = 1
	`)
	assertError(t, "", err)

	vm := NewVM(p)
	if _, err = vm.Run(); err != nil {
		t.Fatal(err)
	}

	g := vm.Globals()
	if len(g) != 1 {
		t.Fatal(g)
	}

	if g[0].ToInt() != 1 {
		t.Fatal(g)
	}

	err = Eval(vm, `a++
						foo()			
						function foo() {
							a++
						}`)

	assertError(t, "", err)
	err = Eval(vm, `
	    let b = 20
		b += 2
		`)

	assertError(t, "", err)

	g = vm.Globals()
	if len(g) != 2 {
		t.Fatal(g)
	}

	if g[0].ToInt() != 3 {
		t.Fatal(g)
	}

	if g[1].ToInt() != 22 {
		t.Fatal(g)
	}
}

func TestEval2(t *testing.T) {
	p, err := CompileStr(`
		let a = 1
	`)
	assertError(t, "", err)

	vm := NewVM(p)
	if _, err = vm.Run(); err != nil {
		t.Fatal(err)
	}

	err = Eval(vm, `let b = a`)
	assertError(t, "", err)

	err = Eval(vm, `b = a / 0`)
	assertError(t, "divide by zero", err)

	err = Eval(vm, `b = a  + 1`)
	assertError(t, "", err)

	v, ok := vm.RegisterValue("b")
	if !ok {
		t.Fatal("b not found")
	}

	if v.ToInt() != 2 {
		t.Fatal(v)
	}
}

func TestEval3(t *testing.T) {
	p, err := CompileStr(`
		let a = 1
	`)
	assertError(t, "", err)

	vm := NewVM(p)
	if _, err = vm.Run(); err != nil {
		t.Fatal(err)
	}

	err = Eval(vm, `let a = 1`)
	assertError(t, "Redeclared identifier", err)
}

func assertError(t *testing.T, msg string, err error) {
	if msg == "" {
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		return
	}

	if !strings.Contains(err.Error(), msg) {
		t.Fatalf("Expected %s, got %v", msg, err)
	}
}
