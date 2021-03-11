
function testWhile1() {
    let a = 0
    while (false) {
        a++
    }
    assert.equal(0, a)
}

function testWhile2() {
    let a = 0
    while (true) {
        a++;
        break;
    }
    assert.equal(1, a)
}

function testWhile3() {
    let a = 0
    while (true) {
        a++;
        if (a == 2) {
            break;
        }
        else {
            continue;
        }
        a++;
    }

    assert.equal(2, a)
}

function testWhile4() {
    let a = 0
    while (a < 3) {
        a++;
        while (true) {
            break;
        }
    }
    assert.equal(3, a)
}

function testWhile6() {
    var i = 0;
    outer:
    while (true) {
        i++;
        if (i == 10) {
            break outer
            throw "should not be here 1"
        }
    }
    assert.equal(10, i)
}

function testWhile7() {
    var i = 0;
    var e = 0;
    var f = 0;
    while (true) {
        try {
            i++
            throw "Test bucle"
        }
        catch{
            e++
        } finally {
            f++
        }
        if (i >= 3) {
            break;
        }
    }
    assert.equal(3, i)
    assert.equal(3, e)
    assert.equal(3, f)
}

function testWhile8() {
    var i = 0;
    var e = 0;
    var f = 0;
    while (true) {
        try {
            i++
            throw "Test bucle"
        }
        catch{
            e++
            break;
        } finally {
            f++
        }

    }
    assert.equal(1, i)
    assert.equal(1, e)
    assert.equal(1, f)
}

function testWhile9() {
    var i = 0;
    var e = 0;
    var f = 0;
    while (true) {
        try {
            i++
        }
        catch{
            e++
        } finally {
            f++
        }
        break;
    }
    assert.equal(1, i)
    assert.equal(0, e)
    assert.equal(1, f)
}

function testWhile10() {
    var i = 0
    var e = 0
    var f = 0
    while (true) {
        for (var j = 0; j < 3; j++) {
            f++
            break
        }
        i++
        break
    }
    assert.equal(1, i)
    assert.equal(1, f)
}

function testWhile11() {
    var i = 0
    var e = 0
    var f = 0
    while (true) {
        for (var j = 0; j < 3; j++) {
            f++
            continue
        }
        i++
        break
    }
    assert.equal(1, i)
    assert.equal(3, f)
}

function testWhile12() {
    var i = 0
    while (true) {
        switch (true) {
            case true:
                break;
        }
        i++
        break
    }
    assert.equal(1, i)
}

function testWhile13() {
    var i = 0
    var j = 0
    var e = 0
    try {
        while (true) {
            try {
                i++
                throw "test"
            } finally {
                j++
            }
        }
    } catch{
        e++
    } finally {
        j++
    }
    assert.equal(1, i)
    assert.equal(2, j)
    assert.equal(1, e)
}


function testWhile14() {
    var i = 0
    while (true) {
        i++
        while (true) {
            i++
            while (true) {
                i++
                break
            }
            break
        }
        break
    }

    assert.equal(3, i)
}