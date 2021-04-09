
function testVariadicLen() {
    assert.equal(1, len(1))
    assert.equal(3, len(1, 2, 3))
}

function testVariadicLen2() {
    let a = [1, 2, 3]
    assert.equal(3, len(...a))
}

function testVariadicLen3() {
    let a = [1, 2, 3]
    assert.equal(5, len(1, 2, ...a))
}

function testVariadicSum1() {
    let a = [1, 2, 3]
    assert.equal(6, sum(...a))
}

function testVariadicSum2() {
    let a = [1, 2, 3]
    assert.equal(9, sum(1, 2, ...a))
}

function len(...a: number[]) {
    return a.length
}

function sum(...a: number[]) {
    let v = 0
    for (let i = 0, l = a.length; i < l; i++) {
        v += a[i]
    }
    return v
}