
function testFor1() {
    let a = 0
    for (; false;) {
        a++
    }
    assert.equal(0, a)
}

function testFor2() {
    let a = 0
    for (var i = 0; i < 10; i++) {
        a++;
    }
    assert.equal(10, a)
}

function testFor3() {
    let a = 0
    let found = false
    for (var i = 0; (i < 10 && !found); i++) {
        a++;
        if (a > 5) {
            found = true
        }
    }
    assert.equal(6, a)
}

function testFor4() {
    let a = 0
    for (var i = 0; i < 10; i++) {
        a++;
        i++;
    }
    assert.equal(5, a)
}

function testFor5() {
    let a = 0
    for (; ;) {
        a++;
        if (a > 5) {
            break;
        }
    }
    assert.equal(6, a)
}

function testFor6() {
    let a = 0
    let b = 0;
    for (var i = 0; i < 10; i++) {
        if (i % 2 == 0) {
            b = b + i
            continue
        }
        a++
    }
    assert.equal(5, a)
    assert.equal(20, b)
}

//testing others...
function testFor7() {
    let a = 0
    for (var i = 0; i < 10; i += 2) {
        a++
    }
    assert.equal(5, a)
}

function testFor8() {
    let a = 0
    for (var i = 10; i > 0; i--) {
        a++
    }
    assert.equal(10, a)
}

function testFor9() {
    let a = 0
    for (var i = 10; i > 0; i -= 2) {
        a++
    }
    assert.equal(5, a)
}

function testFor10() {
    let a = 0
    for (var i = 0; i >> i; i++) {
        a++
        break
    }
    assert.equal(0, a)
}

function testFor11() {
    let a = 0

    for (var i = -576460752303423486; true; i--) {
        a++
        if (a >= 5) {
            break
        }
    }
    assert.equal(5, a)
}


function testFor12() {
    let a = 0

    for (var i = 576460752303423486; true; i++) {
        a++
        if (a >= 5) {
            break
        }
    }
    assert.equal(5, a)
}

function testFor20() {
    var accessed = false;
    for (var i = 0; false;) {
        accessed = true;
        break;
    }
    assert.equal(false, accessed)
}

function testFor21() {
    var accessed = false;
    for (var i = 0; "1";) {
        accessed = true;
        break;
    }
    assert.equal(true, accessed)
}

function testFor22() {
    var count = 0;
    for (var i = 0; null;) {
        count++;
    }
    assert.equal(0, count)
}

function testFor23() {
    var count = 0;
    for (var i = 0; false;) {
        count++;
    }
    assert.equal(0, count)
}

function testFor24() {
    var count = 0;
    for (var i = 0; -0;) {
        count++;
    }
    assert.equal(0, count)
}

function testFor25() {
    var count = 0;
    for (var i = 0; 2;) {
        count++;
        break;
    }
    assert.equal(1, count)
}

function testFor26() {
    let s = 0
    for (var index = 0; index < 10; index += 1) {
        if (index < 5) {
            continue;
        }
        s += index;
    }
    assert.equal(5 + 6 + 7 + 8 + 9, s)
}

function testFor30() {
    for (var i = 0; i < 10; i++) {
        i *= 2;
        if (i === 3) {
            throw "should not be here"
        }
    }
}

function testFor31() {
    let s = 0
    for (var i = 0; i < 2; i++) {
        for (var j = 0; j < 2; j++) {
            s++
        }
    }
    assert.equal(2 * 2, s)
}

function testFor32() {
    let s = 0
    for (var i = 0; i < 2; i++) {
        for (var j = 0; j < 2; j++) {
            if (j == 1) {
                continue
            }
            s++
        }
    }
    assert.equal(2, s)
}

function testFor33() {
    let s = 0
    for (var i = 0; i < 2; i++) {
        for (var j = 0; j < 2; j++) {
            s++
            break
        }
    }
    assert.equal(2, s)
}

function testFor34() {
    let s = 0
    let r = 0
    for (var i = 0; i < 2; i++) {
        for (var j = 0; j < 2; j++) {
            switch (j) {
                case 1: {
                    s += 4
                    break
                }
                    break
                case 2: {
                    throw "should not be here"
                }
                    break
                default:
                    break
            }
            r++
        }
    }
    assert.equal(8, s)
    assert.equal(4, r)
}

//nested with label
function testFor35() {
    let s = 0
    outer:
    for (var i = 0; i < 2; i++) {
        for (var j = 0; j < 2; j++) {
            s++
            continue outer
        }
    }
    assert.equal(2, s)
}

function testFor36() {
    let s = 0
    outer:
    for (var i = 0; i < 2; i++) {
        for (var j = 0; j < 2; j++) {
            s++
            break outer
        }
    }
    assert.equal(1, s)
}

function testFor37() {
    let s = ""
    outer:
    for (var index = 0; index < 4; index += 1) {
        nested:
        for (var index_n = 0; index_n <= index; index_n++) {
            if (index * index_n == 6) {
                continue outer;
            }
            s += "" + index + index_n;
        }
    }
    assert.equal("0010112021223031", s)
}

function testFor38() {
    let s = ""
    outer: for (var index = 0; index < 4; index += 1) {
        nested: for (var index_n = 0; index_n <= index; index_n++) {
            if (index * index_n == 6) {
                continue;
            }
            s += "" + index + index_n;
        }
    }
    assert.equal("001011202122303133", s)
}

function testFor39() {
    let s = ""
    outer: for (var index = 0; index < 4; index += 1) {
        nested: for (var index_n = 0; index_n <= index; index_n++) {
            if (index * index_n == 6) {
                continue nested;
            }
            s += "" + index + index_n;
        }
    }
    assert.equal("001011202122303133", s)
}

//Mega nested for
function testFor40() {
    let s = ""
    for (var index0 = 0; index0 <= 1; index0++) {
        for (var index1 = 0; index1 <= index0; index1++) {
            for (var index2 = 0; index2 <= index1; index2++) {
                for (var index3 = 0; index3 <= index2; index3++) {
                    for (var index4 = 0; index4 <= index3; index4++) {
                        for (var index5 = 0; index5 <= index4; index5++) {
                            for (var index6 = 0; index6 <= index5; index6++) {
                                for (var index7 = 0; index7 <= index6; index7++) {
                                    for (var index8 = 0; index8 <= index1; index8++) {
                                        s += "" + index0 + index1 + index2 + index3 + index4 + index5 + index6 + index7 + index8 + '-';
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    assert.equal("000000000-100000000-110000000-110000001-111000000-111000001-111100000-111100001-111110000-111110001-111111000-111111001-111111100-111111101-111111110-111111111-", s)
}


function testFor41() {
    let s = 0
    let e = 0
    try {
        for (var i = 0; i < 2; i++) {
            for (var j = 0; j < 2; j++) {
                s++
                throw "Exception"
            }
        }
    } catch{
        e++
    }

    assert.equal(1, s)
    assert.equal(1, e)
}

function testFor42() {
    let s = 0
    let e = 0
    try {
        for (var i = 0; i < 2; i++) {
            try {
                for (var j = 0; j < 2; j++) {
                    s++
                    throw "Exception"
                }
            } catch{
                e++
            }
        }
    } catch{
        throw "should not be here"
    }

    assert.equal(2, s)
    assert.equal(2, e)
}

function testFor43() {
    let s = 0
    let e = 0
    let e1 = 0
    try {
        for (var i = 0; i < 2; i++) {
            try {
                for (var j = 0; j < 2; j++) {
                    s++
                    throw "Exception"
                }
            } catch{
                e++
                throw "Exception"
            }
        }
    } catch{
        e1++
    }

    assert.equal(1, s)
    assert.equal(1, e)
    assert.equal(1, e1)
}

function testFor44() {
    let s = 0
    let e = 0
    try {
        for (var i = 0; i < 2; i++) {
            try {
                for (var j = 0; j < 2; j++) {
                    try {
                        s++
                        switch (s) {
                            case 1: break;
                        }
                        throw "Exception"
                    } catch{
                        e++
                        break
                    }
                }
            } catch{
                throw "should not be here"
            }
        }
    } catch{
        throw "should not be here"
    }

    assert.equal(2, s)
    assert.equal(2, e)
}

function testFor45() {
    let s = 0
    let e = 0
    let e1 = 0
    try {
        outer:
        for (var i = 0; i < 2; i++) {
            try {
                for (var j = 0; j < 2; j++) {
                    try {
                        s++
                        switch (i) {
                            case 1: break;
                            default:
                                throw "Exception"
                        }
                    } catch{
                        e++
                        continue outer
                    }
                    if (i == 1) {
                        throw "Exception"
                    }
                }
            } catch{
                e1++
            }
        }
    } catch{
        throw "should not be here"
    }

    assert.equal(2, s)
    assert.equal(1, e)
    assert.equal(1, e1)
}


function testFor46() {
    let s = 0
    outer:
    for (var i = 0; i < 2; i++) {
        s++
        if (s == 1) {
            break outer
            throw "should not be here"
        }
    }
}


function testForIn1() {
    let sum = 0;
    var someArray = [1, 2, 3];
    for (var item in someArray) {
        sum = sum + someArray[item]
    }
    assert.equal(6, sum)
}

function testForIn2() {
    let i = 0
    for (var x in [1, null, 3, , 4]) {
        i++
    }
    assert.equal(4, i)
}

function testForIn3() {
    var obj = { a: 1, b: 2, c: 3 };
    var sum = 0

    for (var prop in obj) {
        sum++
    }
    assert.equal(3, sum)
}


function testForIn4() {
    var sum = 0
    for (var x in [1, 2, , , , 3]) {
        sum++
    }
    assert.equal(3, sum)
}


function testForOfForOfArrayEmpty() {
    //@ts-ignore
    var array = [];
    var i = 0;

    //@ts-ignore
    for (var value of array) {
        i += value
    }
    assert.equal(0, i)
}

function testForOfForOfArrayNull() {
    var array = null;
    var i = 0;

    //@ts-ignore
    for (var value of array) {
        i += value
    }
    assert.equal(0, i)
}

function testForOfForOfNull() {
    var i = 0;

    //@ts-ignore
    for (var value of null) {
        i += value
    }
    assert.equal(0, i)
}

function testForOfForOfUndefined() {
    var i = 0;

    //@ts-ignore
    for (var value of undefined) {
        i += value
    }
    assert.equal(0, i)
}

function testForOfNormalForOf() {
    var array = [0, 1, 2, 3];
    var i = 0;

    for (var value of array) {
        i += value
    }
    assert.equal(6, i)
}

function testForOfForOfWithArraytypes() {
    var array = [0, 'a', true, false, null, undefined, ,];
    var i = 0;

    for (var value of array) {
        assert.equal(value, array[i]);
        i++;
    }
    assert.equal(6, i)
}

function testForOfForOfWithException() {
    var i = 0;

    for (var value of [1]) {
        try {
            i++
            throw "exception"
        }
        catch{
            i++
        } finally {
            i++
        }
    }
    assert.equal(3, i)
}

function testForOfForOfWithLabel() {
    var i = 0;

    label:
    for (var value of [1, 2, 3]) {
        try {
            i++
            throw "exception"
        }
        catch{
            i++
            break label
        } finally {
            i++
        }
    }
    assert.equal(3, i)
}

function testForOfForOfUpdateArray() {
    var array = [0]
    var i = 0

    for (var item of array) {
        if (i > 3) {
            break
        }
        i++
        array.push(1)
        array.push(2)
    }
    assert.equal(1, i)
}

function testForOfForOfUpdateArray2() {
    var array = [0]
    var i = 0

    for (var x of array) {
        array.removeAt(0)
        i++
    }
    assert.equal(1, i)
}


function testForOfForOfNestedAndLabels() {
    var iterator = [1, 2, 3, 4]
    var loop = true;
    var i = 0;

    outer:
    while (loop) {
        loop = false;
        for (var x of iterator) {
            try {
                i++;
                continue outer;
            } catch (err) { }
        }
        i++;
    }
    assert.equal(1, i)
}

function testForOfForOfNestedAndLabels2() {
    var iterator = [1, 2, 3, 4]
    var loop = true;
    var i = 0;

    outer:
    while (loop) {
        loop = false;
        for (var x of iterator) {
            try {
                i++;
                throw "Ex"
            } catch (err) {
                break outer;
            }
        }
        i++;
    }
    assert.equal(1, i)
}

function testForOfForOfWithContinue() {
    var iterator = [1, 2, 3, 4]
    var i = 0

    for (var x of iterator) {
        try {
            i++
            continue
        } catch (err) { }

    }
    assert.equal(4, i)
}