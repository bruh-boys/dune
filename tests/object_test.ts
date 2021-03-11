
function testGroupBy1() {
    let a = [
        { a: 1, b: 2 },
        { a: 1, b: 3 },
        { a: 2, b: 4 }
    ]

    let g = a.groupBy(t => t.a)
    assert.equal(2, Object.len(g))
    assert.equal(2, g[1].length)
    assert.equal(1, g[2].length)
}

function testGroupBy2() {
    let a = [
        { a: 1, b: 2 },
        { a: 1, b: 3 },
        { a: 2, b: 4 }
    ]

    let g = a.groupBy(t => t.a)
    let v = Object.keys(g).sum()
    assert.equal(3, v)
}

function testGroupBy3() {
    let a = [
        { a: "1", b: 2 },
        { a: 1, b: 3 },
        { a: 2, b: 4 }
    ]

    let g = a.groupBy(t => t.a)
    assert.equal(3, Object.keys(g).length)
}

function testGroupBy4() {
    let a = [
        { a: "1", b: 2 },
        { a: "1", b: 3 },
        { a: 2, b: 4 }
    ]

    let g = a.groupBy(t => t.a)
    assert.equal(2, Object.keys(g).length)
}

function testMap1() {
    let a = { foo: { bar: 3 } }
    assert.equal(3, a.foo.bar)
}

function testMapDelete() {
    let a = { foo: 1 }
    Object.deleteKey(a, "foo")
    assert.equal(undefined, a.foo)
}

function testMap3() {
    let a = { foo: 1 }
    let b = Object.clone(a)

    b.foo++

    assert.equal(2, b.foo)
    assert.equal(1, a.foo)
}


function testMapBasicMap() {
    let a = { "0": 0, "1": 1, "2": 2, "3": 3, "4": 4 }
    let sum = 0

    for (var i = 0; i < Object.len(a); i++) {
        //@ts-ignore
        sum = sum + a[convert.toString(i)]
    }
    assert.equal(10, sum)
}


function testMapGetMapValues() {
    let a = { "0": 0, "1": 1, "2": 2, "3": 3, "4": 4 }
    let val

    for (var i = 0; i < Object.len(a); i++) {
        val = Object.values(a)
    }
    assert.equal(5, val.length)
}


function testMapMapOverFlow() {
    let a = { "0": 0, "1": 1, "2": 2, "3": 3, "4": 4 }
    let e

    for (var i = 0; i <= Object.len(a); i++) {
        //@ts-ignore
        e = a[i]
    }

    assert.equal(undefined, e)
}

