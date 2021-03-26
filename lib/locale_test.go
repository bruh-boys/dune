package lib

import "testing"

func TestTranslate1(t *testing.T) {
	v := runTest(t, `
		return T("Hello {{name}}!", "Bill")
		
	`)

	if v.String() != "Hello Bill!" {
		t.Fatal(v.String())
	}
}

func TestTranslate2(t *testing.T) {
	v := runTest(t, `
		locale.defaultLocalizer.translator.add("es", "Hello {{name}}!", "Hola {{name}}!")
		runtime.vm.language = "es"
		return T("Hello {{name}}!", "Bill")
		
	`)

	if v.String() != "Hola Bill!" {
		t.Fatal(v.String())
	}
}

func TestTranslate3(t *testing.T) {
	v := runTest(t, `
		let loc = locale.defaultLocalizer	
		loc.translator.add("es", "Hello", "Hola")	
		return loc.translate("es", "Hello")
		
	`)

	if v.String() != "Hola" {
		t.Fatal(v.String())
	}
}

func TestFormatInt1(t *testing.T) {
	runTest(t, `
		let tests = [
			[1, "1"],
			[10, "10"],
			[100, "100"],
			[1000, "1,000"],
			[10000, "10,000"],
			[100000, "100,000"],
			[1000000, "1,000,000"],
			[10000000, "10,000,000"],
			[-1, "-1"],
			[-10, "-10"],
			[-100, "-100"],
			[-1000, "-1,000"],
			[-10000, "-10,000"],
			[-100000, "-100,000"],
			[-1000000, "-1,000,000"],
			[-10000000, "-10,000,000"],
		]

		for(let t of tests) {
			let v = locale.format("i", t[0])
			if(v != t[1]) {
				throw fmt.sprintf("%d, expected %s, got %s", t[0], t[1], v)
			}
		}		
	`)
}

func TestFormatInt2(t *testing.T) {
	v := runTest(t, `
		let loc = locale.newLocalizer()
		loc.culture = locale.newCulture("es")
		loc.culture.thousandSeparator = "."
		loc.culture.decimalSeparator = ","
		return loc.format("i", 1000)
		
	`)

	if v.String() != "1.000" {
		t.Fatal(v.String())
	}
}

func TestFormatFloat1(t *testing.T) {
	runTest(t, `
		let tests = [
			[1, "1.00"],
			[10, "10.00"],
			[100, "100.00"],
			[1000, "1,000.00"],
			[10000, "10,000.00"],
			[100000, "100,000.00"],
			[1000000, "1,000,000.00"],
			[10000000, "10,000,000.00"],
			[-1, "-1.00"],
			[-10, "-10.00"],
			[-100, "-100.00"],
			[-1000, "-1,000.00"],
			[-10000, "-10,000.00"],
			[-100000, "-100,000.00"],
			[-1000000, "-1,000,000.00"],
			[-10000000, "-10,000,000.00"],
			[1.33, "1.33"],
			[10.33, "10.33"],
			[100.33, "100.33"],
			[1000.33, "1,000.33"],
		]

		for(let t of tests) {
			let v = locale.format("f", t[0])
			if(v != t[1]) {
				throw fmt.sprintf("%.2f, expected %s, got %s", t[0], t[1], v)
			}
		}		
	`)
}

func TestFormatFloat2(t *testing.T) {
	v := runTest(t, `
		let loc = locale.newLocalizer()
		loc.culture = locale.newCulture("es")
		loc.culture.thousandSeparator = "."
		loc.culture.decimalSeparator = ","
		loc.culture.numberOfDecimals = 5
		return loc.format("f", 1000.33)
		
	`)

	if v.String() != "1.000,33000" {
		t.Fatal(v.String())
	}
}

func TestFormatCurrency1(t *testing.T) {
	runTest(t, `
		let tests = [
			[1, "$1.00"],
			[10, "$10.00"],
			[100, "$100.00"],
			[1000, "$1,000.00"],
			[10000, "$10,000.00"],
			[100000, "$100,000.00"],
			[1000000, "$1,000,000.00"],
			[10000000, "$10,000,000.00"],
			[-1, "-$1.00"],
			[-10, "-$10.00"],
			[-100, "-$100.00"],
			[-1000, "-$1,000.00"],
			[-10000, "-$10,000.00"],
			[-100000, "-$100,000.00"],
			[-1000000, "-$1,000,000.00"],
			[-10000000, "-$10,000,000.00"],
			[1.33, "$1.33"],
			[10.33, "$10.33"],
			[100.33, "$100.33"],
			[1000.33, "$1,000.33"],
		]

		for(let t of tests) {
			let v = locale.format("c", t[0])
			if(v != t[1]) {
				throw fmt.sprintf("%.2f, expected %s, got %s", t[0], t[1], v)
			}
		}		
	`)
}

func TestFormatCurrency2(t *testing.T) {
	v := runTest(t, `
		let loc = locale.newLocalizer()
		loc.culture = locale.newCulture("es")
		loc.culture.thousandSeparator = "."
		loc.culture.decimalSeparator = ","
		loc.culture.numberOfDecimals = 2
		loc.culture.currencyPattern = "-0€"
		return loc.format("c", 1000.33)
		
	`)

	if v.String() != "1.000,33€" {
		t.Fatal(v.String())
	}
}
func TestFormatCurrency3(t *testing.T) {
	v := runTest(t, `
		let loc = locale.newLocalizer()
		loc.culture = locale.newCulture("es")
		loc.culture.thousandSeparator = "."
		loc.culture.decimalSeparator = ","
		loc.culture.numberOfDecimals = 2
		loc.culture.currencyPattern = "-0€"
		return loc.format("c", -1000.33)
		
	`)

	if v.String() != "-1.000,33€" {
		t.Fatal(v.String())
	}
}

func TestFormatDate(t *testing.T) {
	runTest(t, `
		let tests = [
			[time.localDate(2020, 12, 10), "yy-MM-dd", "20-12-10"],
			[time.localDate(2020, 12, 10), "yyyy-MM-dd", "2020-12-10"],
			[time.localDate(2021, 1, 1), "yyyy-M-d", "2021-1-1"],
			[time.localDate(2021, 1, 1), "ddd", "Friday"],
			[time.localDate(2021, 1, 1), "MMM", "January"],
			[time.localDate(2021, 1, 1, 7, 8, 9), "h:m:s a", "7:8:9 AM"],
			[time.localDate(2021, 1, 1, 7, 8, 9), "hh:mm:ss a", "07:08:09 AM"],
			[time.localDate(2021, 1, 1, 18), "hh a", "06 PM"],
			[time.localDate(2021, 1, 1, 18), "HH a", "18 PM"],
		]

		for(let t of tests) {
			let v = locale.format(t[1], t[0])
			if(v != t[2]) {
				throw fmt.sprintf("%v, expected '%s', got '%s'", t[0], t[2], v)
			}
		}		
	`)
}

func TestParseNumber1(t *testing.T) {
	runTest(t, `
		let tests = [
			[1, "1"],
			[10, "10"],
			[100, "100"],
			[1000, "1,000"],
			[10000, "10,000"],
			[100000, "100,000"],
			[1000000, "1,000,000"],
			[10000000, "10,000,000"],
			[-1, "-1"],
			[-10, "-10"],
			[-100, "-100"],
			[-1000, "-1,000"],
			[-10000, "-10,000"],
			[-100000, "-100,000"],
			[-1000000, "-1,000,000"],
			[-10000000, "-10,000,000"],
			[1.22, "1.22"],
			[-1.22, "-1.22"],
		]

		for(let t of tests) {
			let v = locale.parseNumber(t[1])
			if(v != t[0]) {
				throw fmt.sprintf("expected %s, got %s", t[0], v)
			}
		}		
	`)
}
func TestParseNumber2(t *testing.T) {
	runTest(t, `
		let tests = [
			[1, "1"],
			[10, "10"],
			[100, "100"],
			[1000, "1.000"],
			[10000, "10.000"],
			[100000, "100.000"],
			[1000000, "1.000.000"],
			[10000000, "10.000.000"],
			[-1, "-1"],
			[-10, "-10"],
			[-100, "-100"],
			[-1000, "-1.000"],
			[-10000, "-10.000"],
			[-100000, "-100.000"],
			[-1000000, "-1.000.000"],
			[-10000000, "-10.000.000"],
			[1.22, "1,22"],
			[-1.22, "-1,22"],
			[-1000.22, "-1.000,22"],
		]

		let loc = locale.newLocalizer()
		loc.culture = locale.newCulture("es")
		loc.culture.thousandSeparator = "."
		loc.culture.decimalSeparator = ","

		for(let t of tests) {
			let v = loc.parseNumber(t[1])
			if(v != t[0]) {
				throw fmt.sprintf("expected %s, got %s", t[0], v)
			}
		}		
	`)
}
func TestParseDate(t *testing.T) {
	runTest(t, `
	let tests = [
		[time.localDate(2020, 12, 10), "yy-MM-dd", "20-12-10"],
		[time.localDate(2020, 12, 10), "yyyy-MM-dd", "2020-12-10"],
		[time.localDate(2021, 1, 1), "yyyy-M-d", "2021-1-1"],
	]

	for(let t of tests) {
		let v = locale.parseDate(t[2], t[1])
		if(!v.equal(t[0])) {
			throw fmt.sprintf("%v, expected '%s', got '%s'", t[0], t[2], v)
		}
	}	
	`)
}
