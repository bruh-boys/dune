

function testScope1() {
    let a = 3
    {
        let a = 7
        assert.equal(7, a)
        a++
        {
            let a = 99
            assert.equal(99, a)
            a++
            assert.equal(100, a)
        }
        assert.equal(8, a)
    }
    assert.equal(3, a)
}

function testScope2() {
    let a = 3
    {
        let a = 7
        a++
        assert.equal(8, a)
    }
    assert.equal(3, a)
}

function testScope3() {
    let a = 3
    {
        let a = 7
        {
            a++
        }
    }

    assert.equal(3, a)
}

function testScope4() {
    let a = 3
    if (a == 3) {
        let a = 7
        assert.equal(7, a)
        a++
        assert.equal(8, a)
    }
    assert.equal(3, a)
}
