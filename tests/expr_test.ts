
function testExpressions() {
    let tests: any = {
        test1: { exp: 3, r: 3 },
        test2: { exp: 3 + 3, r: 6 },
        test3: { exp: -3 + 3, r: 0 },
        test4: { exp: -3 + - 3, r: -6 },
        test5: { exp: -3 - 3, r: -6 },
        test6: { exp: -3 - -3, r: 0 },
        test7: { exp: 3 - -3, r: 6 },
        test8: { exp: 1 + 1.1, r: 2.1 },
        test9: { exp: 1.2 + 1.1, r: 2.3 },

        test10: { exp: 1 / 2, r: 0.5 },
        test11: { exp: 5 % 3, r: 2 },
        test12: { exp: 5 * 3, r: 15 },
        test13: { exp: 11 % 3, r: 2 },

        test14: { exp: 3 + 10 / 2, r: 8 },
        test15: { exp: 3 + 10 % 2, r: 3 },
        test16: { exp: 3 + 2 * 2, r: 7 },
        test17: { exp: (3 + 2) * 2, r: 10 },
        test18: { exp: (3 + 2) * (3 * 2), r: 30 },
        test19: { exp: ((3 + 2) * 2) + (3 + 2 * 2) - 2, r: 15 },

        test20: { exp: 1.0 === 1, r: false },
        test21: { exp: 1 === 1.0, r: false },

        //@ts-ignore
        test22: { exp: 1.000000000000001 == 1, r: false },
        //@ts-ignore
        test23: { exp: 1.2 == 1, r: false },
        //@ts-ignore
        test24: { exp: 1.1 === 1, r: false },
        //@ts-ignore
        test25: { exp: 1.2 != 1, r: true },
        //@ts-ignore
        test26: { exp: 1.1 !== 1, r: true },

        testNum: { exp: 1_000_000, r: 1000000 },
        testNum2: { exp: 1_0_0_0, r: 1000 },

        //@ts-ignore
        test27: { exp: true == 1, r: true },
        //@ts-ignore
        test28: { exp: true == 2, r: false },
        //@ts-ignore
        test29: { exp: true == 0, r: false },
        //@ts-ignore
        test30: { exp: false == 0, r: true },
        //@ts-ignore
        test31: { exp: false == false, r: true },
        //@ts-ignore
        test32: { exp: false == 1, r: false },
        test33: { exp: true && false, r: false },
        test34: { exp: true || false, r: true },
        test35: { exp: true || false && true, r: true },
        test36: { exp: (true || false) && true, r: true },

        test36_A1: { exp: false || 2, r: 2 },
        test36_A2: { exp: false ?? 2, r: false },
        test36_B1: { exp: 0 || 2, r: 2 },
        test36_B2: { exp: 0 ?? 2, r: 0 },

        test37: { exp: 3 > 2, r: true },
        test38: { exp: 3 >= 2, r: true },
        test39: { exp: 3 < 4, r: true },
        test40: { exp: 3 <= 4, r: true },
        test41: { exp: 1 < 1.1, r: true },
        test42: { exp: 1 <= 1.1, r: true },
        test43: { exp: 1 > 1.1, r: false },
        test44: { exp: 1 >= 1.1, r: false },
        test45: { exp: 1.1 < 1, r: false },
        test46: { exp: 1.1 <= 1, r: false },
        test47: { exp: 1.1 > 1, r: true },
        test48: { exp: 1.1 >= 1, r: true },

        //@ts-ignore
        test49: { exp: 3 != 2, r: true },
        test50: { exp: 3 == 3, r: true },
        test51: { exp: !false, r: true },
        test52: { exp: !true, r: false },

        test53: { exp: true ? 1 : 2, r: 1 },
        test54: { exp: false ? 1 : 2, r: 2 },
        test55: { exp: 1 == null ? 1 : 2, r: 2 },

        test56: { exp: 1 + "b", r: "1b" },
        test57: { exp: "a" + 1, r: "a1" },
        test58: { exp: "a" + "b", r: "ab" },

        test59: { exp: 0xA + 0xB, r: 21 },
        test60: { exp: 0xAA ^ 0xBB, r: 17 },
        test61: { exp: 0xFF, r: 255 },
        test62: { exp: 1 | 2, r: 3 },
        test63: { exp: 1 | 5, r: 5 },
        test64: { exp: 3 ^ 6, r: 5 },
        test65: { exp: 3 & 6, r: 2 },
        test66: { exp: 50 >> 2, r: 12 },
        test67: { exp: 2 << 5, r: 64 },

        //NOT
        testNot1: { exp: !false, r: true },
        testNot2: { exp: !true, r: false },

        //OR binary
        testOr1: { exp: 0 | 0, r: 0 },
        testOr2: { exp: 0 | 1, r: 1 },
        testOr3: { exp: 1 | 0, r: 1 },
        testOr4: { exp: 1 | 1, r: 1 },
        testOr5: { exp: 1 | 2, r: 3 },
        testOr6: { exp: 1 | 5, r: 5 },
        testOr7: { exp: 6 | 9, r: 15 },
        testOr8: { exp: 7 | 7, r: 7 },

        //AND binary
        testAnd1: { exp: 0 & 0, r: 0 },
        testAnd2: { exp: 0 & 1, r: 0 },
        testAnd3: { exp: 1 & 0, r: 0 },
        testAnd4: { exp: 1 & 1, r: 1 },
        testAnd5: { exp: 2 & 2, r: 2 },
        testAnd6: { exp: 3 & 6, r: 2 },
        testAnd7: { exp: 3 & 4, r: 0 },

        //XOR binary
        testXor1: { exp: 0 ^ 0, r: 0 },
        testXor2: { exp: 0 ^ 1, r: 1 },
        testXor3: { exp: 1 ^ 0, r: 1 },
        testXor4: { exp: 1 ^ 1, r: 0 },
        testXor5: { exp: 2 ^ 2, r: 0 },
        testXor6: { exp: 3 ^ 2, r: 1 },
        testXor7: { exp: 2 ^ 4, r: 6 },

        //truth tables
        //OR 
        testOrC1: { exp: false || false, r: false },
        testOrC2: { exp: true || false, r: true },
        testOrC3: { exp: false || true, r: true },
        testOrC4: { exp: true || true, r: true },

        //AND
        testAndC1: { exp: false && false, r: false },
        testAndC2: { exp: true && false, r: false },
        testAndC3: { exp: false && true, r: false },
        testAndC4: { exp: true && true, r: true },

        //Boolean comp
        testCmpbool1: { exp: false == false, r: true },
        //@ts-ignore
        testCmpbool2: { exp: true == false, r: false },
        //@ts-ignore
        testCmpbool3: { exp: false == true, r: false },
        //@ts-ignore
        testCmpbool4: { exp: true == true, r: true },

        //Int
        testCmpInt1: { exp: 1 > 0, r: true },
        testCmpInt2: { exp: 1 > 1, r: false },
        testCmpInt3: { exp: 1 > 2, r: false },
        testCmpInt4: { exp: 1 >= 0, r: true },
        testCmpInt5: { exp: 1 >= 1, r: true },
        testCmpInt6: { exp: 1 >= 2, r: false },
        testCmpInt7: { exp: 1 < 0, r: false },
        testCmpInt8: { exp: 1 < 1, r: false },
        testCmpInt9: { exp: 1 < 2, r: true },
        testCmpInt10: { exp: 1 <= 0, r: false },
        testCmpInt11: { exp: 1 <= 1, r: true },
        testCmpInt12: { exp: 1 <= 2, r: true },

        //dec
        testCmpDec1: { exp: 1.1 > 0.1, r: true },
        testCmpDec2: { exp: 1.1 > 1.1, r: false },
        testCmpDec3: { exp: 1.1 > 1.01, r: true },
        testCmpDec4: { exp: 1.1 >= 1.09, r: true },
        testCmpDec5: { exp: 1.1 >= 1.101, r: false },
        testCmpDec6: { exp: 1.1 >= 1.2, r: false },
        testCmpDec7: { exp: 1.1 < 0.1, r: false },
        testCmpDec8: { exp: 1.1 < 1.1, r: false },
        testCmpDec9: { exp: 1.1 < 1.101, r: true },
        testCmpDec10: { exp: 1.1 <= 1.09, r: false },
        testCmpDec11: { exp: 1.1 <= 1.1, r: true },
        testCmpDec12: { exp: 1.1 <= 1.101, r: true },

        //dec vs int
        testCmpIntDec1: { exp: 1 > 0.999, r: true },
        testCmpIntDec2: { exp: 1 > 1.001, r: false },
        testCmpIntDec4: { exp: 1 >= 0.99, r: true },
        testCmpIntDec5: { exp: 1 >= 1.00, r: true },
        testCmpIntDec6: { exp: 1 >= 1.001, r: false },
        testCmpIntDec7: { exp: 1 < 0.1, r: false },
        testCmpIntDec8: { exp: 1 < 1.1, r: true },
        testCmpIntDec9: { exp: 1 < 1.101, r: true },
        testCmpIntDec10: { exp: 1 <= 0.99, r: false },
        testCmpIntDec11: { exp: 1 <= 1.00, r: true },
        testCmpIntDec12: { exp: 1 <= 1.001, r: true },

        testCmpIntDec13: { exp: 0.999 < 1, r: true },
        testCmpIntDec14: { exp: 1.001 < 1, r: false },
        testCmpIntDec15: { exp: 0.99 <= 1, r: true },
        testCmpIntDec16: { exp: 1.00 <= 1, r: true },
        testCmpIntDec17: { exp: 1.001 <= 1, r: false },
        testCmpIntDec18: { exp: 0.1 > 1, r: false },
        testCmpIntDec19: { exp: 1.1 > 1, r: true },
        testCmpIntDec20: { exp: 1.101 > 1, r: true },
        testCmpIntDec21: { exp: 0.99 >= 1, r: false },
        testCmpIntDec22: { exp: 1.00 >= 1, r: true },
        testCmpIntDec23: { exp: 1.001 >= 1, r: true },

        //varoius complex 
        testComplex1: { exp: !(false && true), r: true },
        testComplex2: { exp: !(false || true), r: false },

        testComplex3: { exp: ((1 > 0) && (1 > 2)), r: false },
        testComplex4: { exp: !((1 > 0) && (1 > 2)), r: true },
        testComplex5: { exp: ((1.01 > 1) && (1 > 0.99)), r: true },

        //Complex Or/And left and right
        //And Or
        testComplexAndOr1: { exp: ((true && true) || true), r: true },
        testComplexAndOr2: { exp: ((true && false) || true), r: true },
        testComplexAndOr3: { exp: ((false && true) || true), r: true },
        testComplexAndOr4: { exp: ((false && false) || true), r: true },
        testComplexAndOr5: { exp: ((true && true) || false), r: true },
        testComplexAndOr6: { exp: ((true && false) || false), r: false },
        testComplexAndOr7: { exp: ((false && true) || false), r: false },
        testComplexAndOr8: { exp: ((false && false) || false), r: false },

        //And Or
        testComplexAndOr25: { exp: (true && (true || true)), r: true },
        testComplexAndOr26: { exp: (true && (true || false)), r: true },
        testComplexAndOr27: { exp: (true && (false || true)), r: true },
        testComplexAndOr28: { exp: (true && (false || false)), r: false },
        testComplexAndOr29: { exp: (false && (true || true)), r: false },
        testComplexAndOr30: { exp: (false && (true || false)), r: false },
        testComplexAndOr31: { exp: (false && (false || true)), r: false },
        testComplexAndOr32: { exp: (false && (false || false)), r: false },

        //Or And
        testComplexAndOr9: { exp: ((true || true) && true), r: true },
        testComplexAndOr10: { exp: ((true || false) && true), r: true },
        testComplexAndOr11: { exp: ((false || true) && true), r: true },
        testComplexAndOr12: { exp: ((false || false) && true), r: false },
        testComplexAndOr13: { exp: ((true || true) && false), r: false },
        testComplexAndOr14: { exp: ((true || false) && false), r: false },
        testComplexAndOr15: { exp: ((false || true) && false), r: false },
        testComplexAndOr16: { exp: ((false || false) && false), r: false },

        //Or And
        testComplexAndOr17: { exp: (true || (true && true)), r: true },
        testComplexAndOr18: { exp: (true || (true && false)), r: true },
        testComplexAndOr19: { exp: (true || (false && true)), r: true },
        testComplexAndOr20: { exp: (true || (false && false)), r: true },
        testComplexAndOr21: { exp: (false || (true && true)), r: true },
        testComplexAndOr22: { exp: (false || (true && false)), r: false },
        testComplexAndOr23: { exp: (false || (false && true)), r: false },
        testComplexAndOr24: { exp: (false || (false && false)), r: false },

        //Some strings
        testStr1: { exp: "1" > "0", r: true },
        testStr2: { exp: "1" > "10", r: false },
        testStr3: { exp: "1" > "02", r: true },
        testStr4: { exp: "1" > "2", r: false },
        testStr5: { exp: "1" > "1", r: false },

        testStr6: { exp: "0" < "1", r: true },
        testStr7: { exp: "10" < "1", r: false },
        testStr8: { exp: "02" < "1", r: true },
        testStr9: { exp: "2" < "1", r: false },
        testStr10: { exp: "1" < "1", r: false },

        testStr11: { exp: "1" >= "0", r: true },
        testStr12: { exp: "1" >= "10", r: false },
        testStr13: { exp: "1" >= "02", r: true },
        testStr14: { exp: "1" >= "2", r: false },
        testStr15: { exp: "1" <= "1", r: true },

        testStr16: { exp: "10" <= "1", r: false },
        testStr17: { exp: "02" <= "1", r: true },
        testStr18: { exp: "2" <= "1", r: false },
        testStr19: { exp: "1" <= "1", r: true },
        testStr20: { exp: "0" <= "1", r: true },

        //Comparations ^_^ 
        //int
        testCmp1: { exp: 1 == 1, r: true },
        testCmp2: { exp: 1 === 1, r: true },
        testCmp3: { exp: 1 != 1, r: false },
        //@ts-ignore
        testCmp4: { exp: 1 != 2, r: true },
        //@ts-ignore
        testCmp5: { exp: 2 != 1, r: true },

        //dec
        testCmp6: { exp: 1.01 == 1.01, r: true },
        testCmp7: { exp: 1.01 === 1.01, r: true },
        //@ts-ignore
        testCmp8: { exp: 1.01 != 1.02, r: true },
        //@ts-ignore
        testCmp9: { exp: 1.01 != 2.01, r: true },
        //@ts-ignore
        testCmp10: { exp: 2.01 != 1.01, r: true },

        //int vs dev
        testCmp11: { exp: 1.00 == 1, r: true },
        testCmp12: { exp: 1 == 1.00, r: true },
        testCmp13: { exp: 1 === 1.00, r: false },
        testCmp14: { exp: 1.00 === 1, r: false },
        testCmp15: { exp: 1.00 !== 1, r: true },
        testCmp16: { exp: 1 !== 1.00, r: true },
        //@ts-ignore
        testCmp17: { exp: 1.00 != 2, r: true },
        //@ts-ignore
        testCmp18: { exp: 2 != 1.00, r: true },

        //strings
        testCmp19: { exp: "" == "", r: true },
        testCmp20: { exp: " " == " ", r: true },
        testCmp21: { exp: "a" == "a", r: true },
        testCmp22: { exp: "a" === "a", r: true },
        //@ts-ignore
        testCmp23: { exp: "a" != "A", r: true },
        //@ts-ignore
        testCmp24: { exp: "A" != "a", r: true },

        //some strings
        testCmpStr1: { exp: "1" > "1.00", r: false },
        testCmpStr2: { exp: "1" > "10", r: false },
        testCmpStr3: { exp: "2" > "10", r: true },
        //@ts-ignore
        testCmpStr4: { exp: "1" != "1.00", r: true },
        testCmpStr5: { exp: "1" == "1", r: true },
    }

    for (let key in tests) {
        let t = tests[key]
        if (t.exp != t.r) {
            throw key + ": Expected " + t.r + " but got " + t.exp
        }
    }
}
