
function testSwitch1() {
    switch (1 + 1) {
        case 1:
            throw "should not be here"
        case 2:
            return
        default:
            throw "should not be here"
    }
}

function testSwitch2() {
    let a = "foo"
    switch (a) {
        case "foo":
            return
        case "bar":
        default:
            throw "should not be here"
    }
}

function testSwitch3() {
    let a = "foo"
    switch (a) {
        case "":
        case "foo":
            return
        default:
            throw "should not be here"
    }
}

function testSwitch4() {
    let a = "foo"
    switch (a) {
        case "":
        case "foo":
        case "bar":
            return
        default:
            throw "should not be here"
    }
}

function testSwitch41() {
    let a = "foo"
    switch (a) {
        case "foo":
            break
        case "bar":
            throw "should not be here"
        default:
            throw "should not be here"
    }
    return
}

function testSwitch5() {
    let a = "foo"
    switch (a) {
        case "foo":
            switch (a) {
                case "foo":
                    if (a == "foo") {
                        return
                    }
                    break
            }
            break
        default:
            throw "should not be here"
    }
}

function testSwitch6() {
    let a = "foo"
    switch (a) {
        case "foo":
            switch (a) {
                case "foo":
                    if (a == "foo") {
                        return
                    }
                    break
            }
            break
        default:
            throw "should not be here"
    }
}


function testSwitch10() {
    switch (null) {
        default:
            return
    }
    throw "should not be here"
}

function testSwitch11() {
    switch (true) {
        case true:
            return
        //@ts-ignore
        case false:
            throw "should not be here"
        default:
            throw "should not be here"
    }
    throw "should not be here"
}

function testSwitch12() {
    let a = 1
    switch (a) {
        case 1:
            a = 2
            break
        case 2:
            throw "should not be here"
        default:
            throw "should not be here"
    }
}

function testSwitch13() {
    let a = 1
    switch (a) {
        case 1:
        case 2:
            break
        default:
            throw "should not be here"
    }
}

function testSwitch14() {
    let a = 1
    switch (a) {
        case 1:
        case 2:
            a = 3
            break
        case 3:
            throw "should not be here"
        default:
            throw "should not be here"
    }
}

function testSwitch15() {
    let a = 1.00
    switch (a) {
        case 1:
            return
        case 1.00:
            throw "should not be here"
        default:
            throw "should not be here"
    }
}

function testSwitch16() {
    switch ("a") {
        default:
        case "a": return
        //@ts-ignore
        case "b": throw "should not be here"
    }
    throw "should not be here"
}

function testSwitch17() {
    switch ("a") {
        default:
            break;
        case "a": return
        //@ts-ignore
        case "b": throw "should not be here"
    }
    throw "should not be here"
}

function testSwitch18() {
    let a = 0;
    switch ("a") {
        case "a":
            a++;
            break
        //@ts-ignore
        case "b":
            a++;
            break
    }
    assert.equal(1, a)
}


function testSwitchFallThrough() {
    let a = 0;
    switch ("a") {
        // fallthrough if the block is empty
        case "a":
        //@ts-ignore 
        case "b":
            a++;
            break
    }
    assert.equal(1, a)
}


function testSwitchswitchwithLabel() {
    let a = "foo"
    outer:
    switch (a) {
        case "foo":
            break outer
        default:
            throw "should not be here"

    }
}

function testSwitchswitchwithLabelAndFor() {
    let a = "foo"
    let indice = 0;

    parent:
    switch (a) {
        case "foo":
            outer:
            for (var b = 0; b < 3; b++) {
                indice++
                inner:
                switch (a) {
                    case "foo":
                        for (var i = 0; i < 3; i++) {
                            break inner
                        }
                        //indice++
                        break
                }
            }
            //indice++
            break
    }
    assert.equal(3, indice)
}
