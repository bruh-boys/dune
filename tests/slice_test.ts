
function testSlice1() {
    let a = []
    a.push(2)
    assert.equal(1, a.length)
    assert.equal(2, a[0])
}

function testSlice2() {
    let a = []
    a.push(2, 3)
    assert.equal(2, a.length)
    assert.equal(2, a[0])
    assert.equal(3, a[1])
}

function testSlice3() {
    let a = [2, 3]
    assert.equal(2, a.length)
    assert.equal(2, a[0])
    assert.equal(3, a[1])
}

function testSlice4() {
    let a = [2, 3]
    a.insertAt(1, 6)
    assert.equal(3, a.length)
    assert.equal(2, a[0])
    assert.equal(6, a[1])
    assert.equal(3, a[2])
}

function testSlice5() {
    let a = [2, 3]
    a.removeAt(1)
    assert.equal(1, a.length)
    assert.equal(2, a[0])
}

function testSlice6() {
    let a = [1]
    a.pushRange([2, 3])
    assert.equal(3, a.length)
    assert.equal(1, a[0])
    assert.equal(2, a[1])
    assert.equal(3, a[2])
}