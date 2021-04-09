package lib

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	runTest(t, `        
		let values = [
			[ "2s", time.duration(2 * time.Second) ],
			[ "2m", time.duration(2 * time.Minute) ],
			[ "2h", time.duration(2 * time.Hour) ],
			[ "2d", time.duration(2 * time.Hour * 24) ],
		]

		for(let v of values) {
			let p = time.parseDuration(v[0])
			assert.equal(p, v[1])
		}
	`)
}
func TestEquatable(t *testing.T) {
	runTest(t, `        
		let d1 = time.date(2020, 10, 1, 22, 30)
		let d2 = time.date(2020, 10, 1, 22, 30)

		if(d1 != d2) {
			throw "Expected to be equal"
		}
	`)
}

func TestComparable(t *testing.T) {
	runTest(t, `        
		let d1 = time.date(2020, 10, 1, 22, 30)
		let d2 = time.date(2020, 10, 1, 22, 31)

		if(d1 > d2) {
			throw "Expected to be lesser"
		}

		if(d1 >= d2) {
			throw "Expected to be lesser"
		}

		if(d2 < d1) {
			throw "Expected to be lesser"
		}

		if(d2 <= d1) {
			throw "Expected to be lesser"
		}
	`)
}

func TestSameDay(t *testing.T) {
	runTest(t, `        
		let d1 = time.date(2020, 10, 1, 22, 30)
		let d2 = time.date(2020, 10, 1, 18, 30)
		let d3 = time.date(2020, 10, 2, 18, 30)

		if(!d1.sameDay(d2)) {
			throw "Expected same day"
		}

		if(d1.sameDay(d3)) {
			throw "Expected not the same day"
		}
	`)
}

func TestTimeSetDate(t *testing.T) {
	vm, err := runExpr(t, `
		let x = time.date(2017, 11, 1, 22)
		x = x.setDate(2018, 12, 2)
	`)
	if err != nil {
		t.Fatal(err)
	}

	v, _ := vm.RegisterValue("x")
	d, ok := v.Export(10).(time.Time)
	if !ok {
		t.Fatal("Expected time")
	}

	if d.Year() != 2018 {
		t.Fatal()
	}
	if d.Month() != time.December {
		t.Fatal()
	}
	if d.Day() != 2 {
		t.Fatal()
	}
	if d.Hour() != 22 {
		t.Fatal()
	}
}

func TestTimeSetTime(t *testing.T) {
	vm, err := runExpr(t, `
		let x = time.date(2017, 11, 1, 22, 30)
		x = x.setTime(10, 15, 1)
	`)
	if err != nil {
		t.Fatal(err)
	}

	v, _ := vm.RegisterValue("x")
	d, ok := v.Export(10).(time.Time)
	if !ok {
		t.Fatal("Expected time")
	}

	if d.Year() != 2017 {
		t.Fatal()
	}
	if d.Month() != time.November {
		t.Fatal()
	}
	if d.Day() != 1 {
		t.Fatal()
	}
	if d.Hour() != 10 {
		t.Fatal()
	}
	if d.Minute() != 15 {
		t.Fatal()
	}
	if d.Second() != 1 {
		t.Fatal()
	}
}
