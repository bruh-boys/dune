
function testIfIfs() {
    //@ts-ignore
    if (1 == 2) {
        throw "fail"
    } else {
        return
    }
}

function testIf2() {
    //@ts-ignore
    if (1 != 1) {
        throw "fail"
    } else {
        return
    }
}

function testIf3() {
    if (1 != 1.0) {
        throw "fail"
    } else {
        return
    }
}

function testIf4() {
    if (1 !== 1.0) {
        return
    } else {
        throw "fail"
    }
}

function testIf5() {
    let a = 2
    if (a == 1) {
        throw "fail"
    } else if (a == 2) {
        return
    } else {
        throw "fail"
    }
}

function testIf6() {
    let a = 2
    if (a == 1 || a > 2) {
        throw "fail"
    } else if (a == 2 && a < 2) {
        throw "fail"
    } else {
        return
    }
}

function testIf7() {
    let a = 2
    if (a == 2) {
        if (a < 3 && a > 1) {
            return
        }
        throw "fail"
    } else {
        throw "fail"
    }
}


function testIf8() {
    if (1 == 1) {
        return
    }
    throw "fail"
}

function testIf9() {
    if (1.0 == 1) {
        return
    }
    throw "fail"
}

function testIf10() {
    if (1 == 1.0) {
        return
    }
    throw "fail"
}

function testIf11() {
    if ("a" == "a") {
        return
    }
    throw "fail"
}

function testIf12() {
    if (true) {
        return
    }
    throw "fail"
}

function testIf13() {
    if (false) {
        throw "fail"
    } else {
        if (true) {
            return
        }
    }
    throw "fail"
}

function testIf14() {
    if (false) {
        throw "fail"
    } else {
        if (false) {
            throw "fail"
        } else {
            return
        }
    }
    throw "fail"
}

function testIf15() {
    if (false) {
        throw "fail"
    } else if (true) {
        return
    }
    throw "fail"
}

function testIf16() {
    if (false) {
        throw "fail"
    } else if (false) {
        throw "fail"
    }
    else {
        return
    }
    throw "fail"
}