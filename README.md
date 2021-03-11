# Dune 

## Install 

Download a [release](https://github.com/dunelang/dune/releases) or get it from source

```
go get github.com/dunelang/dune/cmd/dune
```


REPL:
```
Dune v0.95

> 1 + 2
3
```

Create a hello.ts:
```typescript
fmt.println("Hello world")
```

Run it:
```
$ dune hello.ts 
Hello world
```

Examples
---

Arrays have multiple built in functions:

```typescript
let items = [1, 2, 3, 4, 5]
let r = items.where(t => t > 2).select(t => t * t).sum()
console.log(r)
```

A web server:
```typescript
let s = http.newServer()
s.address = ":8080"
s.handler = (w, r) => w.write("Hello world")
s.start() 
```

Working with databases:

```typescript
let db = sql.open("mysql", "test:123@unix(/var/run/mysqld/mysqld.sock)/")

db.exec("CREATE TABLE people (id KEY, name TEXT)")

db.exec("INSERT INTO people (name) VALUES (?)", "Bob")

for (let r of db.query("SELECT id, name FROM people")) {
	console.log(r.id, r.name)
}
```

Permissions
---

By default a virtual machine has no access to the external world. The file system is virtual.

This code throws: unauthorized
```typescript
let p = bytecode.compileStr(`http.get("http://google.com")`)
let vm = runtime.newVM(p)
vm.run()
```

Trusted machines are unrestricted:
```typescript
vm.trusted = true
```

Programs can request permissions with directives:
```typescript
// [permissions networking]
http.get("http://google.com") 
```

Execution limits:
---

This code throws: Step limit reached: 100

```typescript
let p = bytecode.compileStr(`while(true) { }`)
let vm = runtime.newVM(p)
vm.maxSteps = 100
vm.run()
```

This code throws: Max stack frames reached: 5

```typescript
let p = bytecode.compileStr(`function main() { main() }`)
let vm = runtime.newVM(p)
vm.maxFrames = 5
vm.run()
```

MaxAllocations counts the number of variables set and in the case of strings their size so it is not very useful yet. This code throws: Max allocations reached: 10

```typescript
let p = bytecode.compileStr(`
	let v = "*"
	while(true) {
		v += v
	}
`)
let vm = runtime.newVM(p)
vm.maxAllocations = 10
vm.run()
```

Checkout more [examples](https://github.com/dunelang/examples).

Testing
---
Since everything is virtual, including the file system, programs are easily testable. There is an assert package in the standard library.

Writing programs
---

The syntax is a subset of Typescript to get type checking, autocompletion and refactoring tools from editors like VS Code. 

```
$ dune --init
$ code rand.ts
```

Write a basic program:
```typescript
export function main(len?: string) {
    let n = convert.toInt(len || "15")
    let v = crypto.randomAlphanumeric(n)
    fmt.println(v)
}   
```

```
$ dune rand
1RyXuMFKwmxPbTa6bpXk
```

To decompile a program:
```
$ dune -d /tmp/rand.ts 

=============================
Functions
=============================

0F @global
-----------------------------
  0     return              --     --     --   ;   @global


1F main
-----------------------------
  0     moveAndTest         2L     0L     3L   ;   /tmp/rand.ts:8
  1     testJump            3L     1D     0D   ;   /tmp/rand.ts:8
  2     move                2L     0K     --   ;   /tmp/rand.ts:8
  3     callSingleArg      81N     1L     2L   ;   /tmp/rand.ts:8
  4     callSingleArg     107N     4L     1L   ;   /tmp/rand.ts:9
  5     callSingleArg     131N     --     4L   ;   /tmp/rand.ts:10
  6     return              --     --     --   ;   /tmp/rand.ts:10

  0L len 0-6
  1L n 0-6
  2L @ 0-6
  3L @ 0-6
  4L v 4-6

=============================
Constants
=============================
0K string 15
```


Embedding
---

```Go
package main

import (
	"fmt"

	"github.com/dunelang/dune"
)

func main() {
	v, err := dune.RunStr("return 3 * 2")
	fmt.Println(v, err)
}
```

To call Go functions:
```Go
package main

import (
	"fmt"
	"log"

	"github.com/dunelang/dune"
)

func main() {
	dune.AddNativeFunc(dune.NativeFunction{
		Name:      "math.square",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0].ToInt()
			return dune.NewInt64(v * v), nil
		},
	})

	p, err := dune.CompileStr("return math.square(5)")
	if err != nil {
		log.Fatal(err)
	}

	v, err := dune.NewVM(p).Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(v)
}
```

To do anything useful besides basic calculations, it needs to import the standard library:

```Go
package main

import (
	"log"

	"github.com/dunelang/dune"
	_ "github.com/dunelang/dune/lib"
)

func main() {
	_, err := dune.RunStr(`
		let v = { foo: 33 }
		console.log(v)
	`)

	if err != nil {
		log.Fatal(err)
	}
}

```
