function testTryLoopCatch() {
    let a = ""
    for (let i = 1; i <= 2; i++) {
        try {
            a += "0"
            let b = 1 / 0
        } catch (e) {
            a += "1"
        }
    }

    assert.equal("0101", a)
}

function testTryLoopCatchFinally() {
    let a = ""
    for (let i = 1; i <= 2; i++) {
        try {
            a += "0"
            let b = 1 / 0
        } catch (e) {
            a += "1"
        } finally {
            a += "2"
        }
    }

    assert.equal("012012", a)
}


function testTryLoopFinally() {
    let a = ""
    for (let i = 1; i <= 2; i++) {
        try {
            a += "0"
        } finally {
            a += "1"
        }
    }

    assert.equal("0101", a)
}

function testTryFail() {
    let fn = () => {
        try {
            let a = 1 / 0
            return "this can't happen"
        } catch (error) {
            throw "this is expected"
        }
    }

    assert.exception("this is expected", fn)
}

function testTry1() {
    let fn = () => {
        try {
            let a = 1 / 0
            return "this can't happen"
        } catch (error) {
            return "OK"
        }
    }

    assert.equal("OK", fn())
}

function testTry2() {
    let fun = () => {
        let a = "a"
        try {
            throw "lalala"
        } catch (error) {
            a = "b"
        } finally {
            a = "OK"
        }
        return a
    }

    assert.equal("OK", fun())
}


function testTry3() {
    let fun = () => {
        let a
        for (let i = 1; i < 2; i++) {
            try {
                a = "b"
            } finally {
                a = "OK"
            }
        }
        return a
    }

    assert.equal("OK", fun())
}

function testTryNestedLoop1() {
    let fun = () => {
        let a = 1
        try {
            for (let i = 1; i < 2; i++) {
                if (true) {
                    try {
                        a++
                    } catch (error) {
                        a = -1
                    }
                }
                continue
            }
        } catch (error) {
            fmt.println(error)
        }
        return a
    }

    assert.equal(2, fun())
}

function testTryNestedLoop2() {
    let fun = () => {
        let a = 1
        try {
            for (let i = 1; i < 2; i++) {
                if (true) {
                    try {
                        a++
                    } catch (error) {
                        a = -1
                    } finally {
                        a++
                    }
                }
                continue
            }
        } catch (error) {
            fmt.println(error)
        }
        return a
    }

    assert.equal(3, fun())
}

function testTryNestedLoop3() {
    let fun = () => {
        let a = 1
        try {
            for (let i = 1; i < 3; i++) {
                if (true) {
                    try {
                        a++
                    } catch (error) {
                        a = -1
                    } finally {
                        a++
                    }
                }
                continue
            }
        } catch (error) {
            fmt.println(error)
        }
        return a
    }

    assert.equal(5, fun())
}

function testTryNestedLoopWithLabel1() {
    let fun = () => {
        let a = 1
        OUTER:
        for (let j = 1; j < 2; j++) {
            try {
                for (let i = 1; i < 2; i++) {
                    if (true) {
                        try {
                            a++
                        } catch (error) {
                            fmt.println(error)
                        }
                    }
                    continue OUTER
                }

            } catch (error) {
                fmt.println(error)
            }
        }
        return a
    }

    assert.equal(2, fun())
}

function testTryNestedLoopWithLabel12() {
    let fun = () => {
        let a = 1
        OUTER:
        for (let j = 1; j < 2; j++) {
            try {
                for (let i = 1; i < 2; i++) {
                    if (true) {
                        try {
                            a++
                        } catch (error) {
                            fmt.println(error)
                        } finally {
                            a++
                        }
                    }
                    continue OUTER
                }

            } catch (error) {
                fmt.println(error)
            }
        }
        return a
    }

    assert.equal(3, fun())
}

function testTryNestedLoopWithLabel3() {
    let fun = () => {
        let a = 0
        try {
            OUTER:
            for (let j = 1; j < 3; j++) {
                try {
                    try {
                        for (let i = 1; i < 2; i++) {
                            if (true) {
                                try {
                                    a++
                                } catch (error) {
                                    panic(error)
                                } finally {
                                    a++
                                }
                            }
                            continue OUTER
                        }

                    } catch (error) {
                        panic(error)
                    } finally {
                        a++
                    }
                } catch (error) {
                    panic(error)
                } finally {
                    a++
                }
            }
        } catch (error) {
            panic(error)
        } finally {
            a++
        }
        return a
    }

    assert.equal(9, fun())
}

function testTry10() {
    let a = 0
    try {
        a++
    } catch{
        a++
    }

    assert.equal(1, a)
}

function testTry11() {
    let a = 0
    try {
        a++
        throw "test"
    } catch{
        a++
    }
    assert.equal(2, a)
}

function testTry12() {
    let a = 0
    try {
        a++
    } catch{
        a++
    } finally {
        a++
    }
    assert.equal(2, a)
}

function testTry13() {
    let a = 0
    try {
        a++
        throw "test"
    } catch{
        a++
    } finally {
        a++
    }
    assert.equal(3, a)
}

function testTry14() {
    let a = 0

    try {
        a++
        throw "test"
    } catch (error) {

        a++
    } finally {
        a++
    }
    assert.equal(3, a)
}

function testTry15() {
    let a = 0

    try {
        a++
        throw "test"
    } catch (error) {
        a++
    } finally {
        a++
    }
    assert.equal(3, a)
}

function testTry16() {
    let a = 0

    try {
        a++
    } finally {
        a++
    }
    assert.equal(2, a)
}

function testTry17() {
    let a = 0

    try {
        a++
        throw ['inside']
    } catch (error) {
        a++
    }

    assert.equal(2, a)
}


function testTry18() {
    let a = 0

    try {
        a++
        throw []
    } catch (error) {
        a++
    }
    finally {
        a++
    }

    assert.equal(3, a)
}

function testTry20() {
    let a = 0
    try {
        a++
        try {
            a++
            throw "Error"
        } catch {
            a++
            throw "otroError"
            a++
        }
    } catch (error) {
        a++
    }
    finally {
        a++
    }
    assert.equal(5, a)
}

function testTry21() {
    let a = 0
    try {
        a++
        try {
            a++
            throw "Error"
        } catch {
            a++
            try {
                throw "otroError"
            } catch{
                try {
                    throw "otroMas"
                } catch (error) {
                    a++
                } finally {
                    throw "desde el 5"
                }
            }
            a++
        }
    } catch (error) {
        a++
    } finally {
        a++
    }
    assert.equal(6, a)
}


function testTry22() {
    let a = 0
    try {
        try {
            a++
            throw "Desde el 1"
        } catch{
            try {
                a++
                throw "Desde el 2"
            } catch (error) {
                a++
                throw "desde el 3"
            } finally {
                a++
                throw "desde el 4"
            }
        }
        finally {
            a++
            throw "desde el 5"
        }
    } catch (error) {
        a++
    } finally {
        a++
    }
    assert.equal(7, a)
}






