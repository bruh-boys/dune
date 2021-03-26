package lib

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Time, `

declare namespace time {
    /**
     * The ISO time format.
     */
    export const RFC3339: string
    /**
     * The default date format.
     */
    export const DefaultDateFormat: string

    export const Nanosecond: number
    export const Microsecond: number
    export const Millisecond: number
    export const Second: number
    export const Minute: number
    export const Hour: number

    export const SecMillis: number
    export const MinMillis: number
    export const HourMillis: number
    export const DayMillis: number

    export function now(): Time
    export function nowUTC(): Time

    export const Monday: number
    export const Tuesday: number
    export const Wednesday: number
    export const Thursday: number
    export const Friday: number
    export const Saturday: number
    export const Sunday: number

    /**
     * The number of nanoseconds since the unix epoch.
     */
    export let unixNano: number

    export interface Location {
        name: string
    }

    export const utc: Location
    export const local: Location

    export function setLocation(name: string): void
    export function setDefaultLocation(name: string): void

    /**
     * Sets a fixed value for now() for testing.
     */
    export function setFixedNow(t: Time): void

    /**
     * Remove a fixed value for now().
     */
    export function unsetFixedNow(): void
    export function loadLocation(name: string): Location

	export function formatMinutes(v: number): string

    /**
     * 
     * @param seconds from unix epoch
     */
    export function unix(seconds: number): Time

    export function date(year?: number, month?: number, day?: number, hour?: number, min?: number, sec?: number, loc?: Location): Time
    export function localDate(year?: number, month?: number, day?: number, hour?: number, min?: number, sec?: number): Time


	export function parseDuration(s: string): Duration
	
    export function duration(nanoseconds: number | Duration): Duration
    export function toDuration(hour: number, minute?: number, second?: number): Duration
    export function toMilliseconds(hour: number, minute?: number, second?: number): number

    export function daysInMonth(year: number, month: number): number

    export interface Time {
        unix: number
        second: number
        nanosecond: number
        minute: number
        hour: number
        day: number
        /**
         * sunday = 0, monday = 1, ...
         */
        dayOfWeek: number
        month: number
        year: number
        yearDay: number
        location: Location
        /**
         * The time part in milliseconds
         */
        time(): number

        /**
         * Return the date discarding the time part in local time.
         */
        startOfDay(): Time
        /**
         * Returns the las moment of the day in local time
         */
        endOfDay(): Time
        utc(): Time
        local(): Time
        sub(t: Time): Duration
        add(t: Duration | number): Time
        addYears(t: number): Time
        addMonths(t: number): Time
        addDays(t: number): Time
        addHours(t: number): Time
        addMinutes(t: number): Time
        addSeconds(t: number): Time
        addMilliseconds(t: number): Time

        setDate(year?: number, month?: number, day?: number): Time
        addDate(year: number, month: number, day: number): Time
        setTime(hour?: number, minute?: number, second?: number, millisecond?: number): Time
        setTimeMillis(millis: number): Time

        format(f: string): string
		formatIn(f: string, loc: Location): string
		toISO(): string
		
		in(loc: Location): Time
        /**
         * setLocation returns the same time with the location. No conversions
         * are made. 9:00 UTC becomes 9:00 Europe/Madrid
         */
        setLocation(loc: Location): Time
        equal(t: Time): boolean
        after(t: Time): boolean
        afterOrEqual(t: Time): boolean
        before(t: Time): boolean
        beforeOrEqual(t: Time): boolean
        between(t1: Time, t2: Time): boolean
        sameDay(t: Time): boolean
    }

    export interface Duration {
        hours: number
        minutes: number
        seconds: number
        milliseconds: number
        nanoseconds: number
        equal(other: number | Duration): boolean
        greater(other: number | Duration): boolean
        lesser(other: number | Duration): boolean
        add(other: number | Duration): Duration
        sub(other: number | Duration): Duration
        multiply(other: number | Duration): Duration
    }

    export interface Period {
        start?: Time
        end?: Time
    }

    export function after(d: number | Duration, value?: any): sync.Channel
    export function sleep(millis: number): void
    export function sleep(d: Duration): void
    export function parse(value: any, format?: string): Time
    export function parseLocal(value: any, format?: string): Time
    export function parseInLocation(value: any, format: string, location: Location): Time
	
	
	export function newTicker(duration: number | Duration, func: Function): Ticker
    export function newTimer(duration: number, func: Function): Ticker

    export interface Ticker {
        stop(): void
    }

}



`)
}

const (
	secMillis  = 1000
	minMillis  = 60 * secMillis
	hourMillis = 60 * minMillis
	dayMillis  = 24 * hourMillis
)

var Time = []dune.NativeFunction{
	{
		Name: "->time.Monday",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Monday)), nil
		},
	},
	{
		Name: "->time.Tuesday",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Tuesday)), nil
		},
	},
	{
		Name: "->time.Wednesday",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Wednesday)), nil
		},
	},
	{
		Name: "->time.Thursday",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Thursday)), nil
		},
	},
	{
		Name: "->time.Friday",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Friday)), nil
		},
	},
	{
		Name: "->time.Saturday",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Saturday)), nil
		},
	},
	{
		Name: "->time.Sunday",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Sunday)), nil
		},
	},

	{
		Name: "->time.SecMillis",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(secMillis)), nil
		},
	},
	{
		Name: "->time.MinMillis",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(minMillis)), nil
		},
	},
	{
		Name: "->time.HourMillis",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(hourMillis)), nil
		},
	},
	{
		Name: "->time.DayMillis",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(dayMillis)), nil
		},
	},
	{
		Name: "->time.Nanosecond",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Nanosecond)), nil
		},
	},
	{
		Name: "->time.Microsecond",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Microsecond)), nil
		},
	},
	{
		Name: "->time.Millisecond",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Millisecond)), nil
		},
	},
	{
		Name: "->time.Second",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(time.Second)), nil
		},
	},
	{
		Name: "->time.Minute",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt64(int64(time.Minute)), nil
		},
	},
	{
		Name: "->time.Hour",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt64(int64(time.Hour)), nil
		},
	},
	{
		Name: "->time.RFC3339",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewString(time.RFC3339), nil
		},
	},
	{
		Name: "->time.DefaultDateFormat",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewString("2006-1-2"), nil
		},
	},
	{
		Name:      "time.daysInMonth",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int, dune.Int); err != nil {
				return dune.NullValue, err
			}

			year := args[0].ToInt()
			month := args[1].ToInt()
			days := time.Date(int(year), time.Month(month), 0, 0, 0, 0, 0, time.UTC).Day()
			return dune.NewInt(days), nil
		},
	},
	{
		Name:      "time.date",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return getDate(args, vm, time.UTC)
		},
	},
	{
		Name:      "time.localDate",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			loc := GetLocation(vm)
			return getDate(args, vm, loc)
		},
	},
	{
		Name:      "time.parseDuration",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			s := args[0].String()
			ln := len(s)
			if ln < 2 {
				return dune.NullValue, vm.NewPublicError("invalid duration. Format is for example: 1s or 2d")
			}

			v, err := strconv.Atoi(s[:ln-1])
			if err != nil {
				return dune.NullValue, vm.NewPublicError("invalid duration for %s: must be an int", s)
			}

			var d time.Duration

			switch s[ln-1:] {
			case "s":
				d = time.Duration(v) * time.Second
			case "m":
				d = time.Duration(v) * time.Minute
			case "h":
				d = time.Duration(v) * time.Hour
			case "d":
				d = time.Duration(v) * time.Hour * 24
			default:
				return dune.NullValue, vm.NewPublicError("invalid duration: %s", s)
			}

			return dune.NewObject(Duration(d)), nil
		},
	},
	{
		Name:      "time.duration",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}

			d, err := ToDuration(args[0])
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(Duration(d)), nil
		},
	},
	{
		Name:      "time.toDuration",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Int, dune.Int, dune.Int); err != nil {
				return dune.NullValue, err
			}

			l := len(args)

			if l == 0 {
				return dune.NullValue, fmt.Errorf("expected at least one parameter")
			}

			d := args[0].ToInt() * int64(time.Hour)

			if l > 1 {
				d += args[1].ToInt() * int64(time.Minute)
			}

			if l > 2 {
				d += args[2].ToInt() * int64(time.Second)
			}

			return dune.NewObject(Duration(d)), nil
		},
	},
	{
		Name:      "time.toMilliseconds",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Int, dune.Int, dune.Int); err != nil {
				return dune.NullValue, err
			}

			l := len(args)

			if l == 0 {
				return dune.NullValue, fmt.Errorf("expected at least one parameter")
			}

			m := args[0].ToInt() * 60 * 60 * 1000

			if l > 1 {
				m += args[1].ToInt() * 60 * 1000
			}

			if l > 2 {
				m += args[2].ToInt() * 1000
			}

			return dune.NewInt64(m), nil
		},
	},
	{
		Name:      "time.unix",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}
			sec := args[0].ToInt()
			t := time.Unix(sec, 0)
			return dune.NewObject(TimeObj(t)), nil
		},
	},
	{
		Name:      "time.setDefaultLocation",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.HasPermission("trusted") {
				return dune.NullValue, ErrUnauthorized
			}

			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			name := args[0].String()

			l, err := time.LoadLocation(name)
			if err != nil {
				return dune.NullValue, fmt.Errorf("error loading timezone %s: %w", name, err)
			}

			time.Local = l

			return dune.NullValue, nil
		},
	},
	{
		Name:      "time.setLocation",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.HasPermission("trusted") {
				return dune.NullValue, ErrUnauthorized
			}

			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			name := args[0].String()

			l, err := time.LoadLocation(name)
			if err != nil {
				return dune.NullValue, fmt.Errorf("error loading timezone %s: %w", name, err)
			}

			vm.Location = l

			return dune.NullValue, nil
		},
	},
	{
		Name:      "time.setFixedNow",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.HasPermission("trusted") {
				return dune.NullValue, ErrUnauthorized
			}

			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			t, ok := args[0].ToObjectOrNil().(TimeObj)
			if !ok {
				return dune.NullValue, ErrInvalidType
			}

			vm.Now = time.Time(t)
			return dune.NullValue, nil
		},
	},
	{
		Name:      "time.unsetFixedNow",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.HasPermission("trusted") {
				return dune.NullValue, ErrUnauthorized
			}

			vm.Now = time.Time{}
			return dune.NullValue, nil
		},
	},
	{
		Name: "time.now",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.Now.IsZero() {
				return dune.NewObject(TimeObj(vm.Now)), nil
			}
			loc := GetLocation(vm)
			t := time.Now().In(loc)
			return dune.NewObject(TimeObj(t)), nil
		},
	},
	{
		Name: "time.nowUTC",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			n := vm.Now
			if !n.IsZero() {
				return dune.NewObject(TimeObj(n.UTC())), nil
			}
			return dune.NewObject(TimeObj(time.Now().UTC())), nil
		},
	},
	{
		Name: "time.unixNano",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			n := vm.Now
			if !n.IsZero() {
				return dune.NewInt64(n.UnixNano()), nil
			}
			return dune.NewInt64(time.Now().UnixNano()), nil
		},
	},
	{
		Name:      "time.sleep",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if len(args) != 1 {
				return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
			}

			d, err := ToDuration(args[0])
			if err != nil {
				return dune.NullValue, err
			}

			time.Sleep(d)
			return dune.NullValue, nil
		},
	},
	{
		Name: "->time.utc",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewObject(location{time.UTC}), nil
		},
	},
	{
		Name: "->time.local",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := GetLocation(vm)
			return dune.NewObject(location{l}), nil
		},
	},
	{
		Name:      "time.loadLocation",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			l, err := time.LoadLocation(args[0].String())
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewObject(location{l}), nil
		},
	},
	{
		Name:      "time.parse",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			var value string
			var format string

			switch len(args) {
			case 1:
				if err := ValidateArgs(args, dune.String); err != nil {
					return dune.NullValue, err
				}
				value = args[0].String()
			case 2:
				if err := ValidateArgs(args, dune.String, dune.String); err != nil {
					return dune.NullValue, err
				}
				value = args[0].String()
				format = args[1].String()
			default:
				return dune.NullValue, fmt.Errorf("expected 1 or 2 params, got %d", len(args))
			}

			t, err := parseDate(value, format, time.UTC)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(TimeObj(t)), nil
		},
	},
	{
		Name:      "time.parseLocal",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := GetLocation(vm)
			return parseInLocation(l, this, args, vm)
		},
	},
	{
		Name:      "time.parseInLocation",
		Arguments: 3,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			location, ok := args[2].ToObjectOrNil().(location)
			if !ok {
				return dune.NullValue, fmt.Errorf("invalid location, got %s", args[2].TypeName())
			}
			return parseInLocation(location.l, this, args, vm)
		},
	},
	{
		Name:      "time.formatMinutes",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			var min float64

			a := args[0]

			switch a.Type {
			case dune.Int, dune.Float:
				min = a.ToFloat()

			default:
				return dune.NullValue, fmt.Errorf("expected a number, got %v", a.TypeName())
			}

			negative := min < 0

			min = math.Abs(min)

			h := int(math.Floor(min / 60))
			m := int(min) % 60

			s := fmt.Sprintf("%02d:%02d", h, m)

			if negative {
				s = "-" + s
			}

			return dune.NewString(s), nil
		},
	},
	{
		Name:      "time.newTicker",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.HasPermission("async") {
				return dune.NullValue, ErrUnauthorized
			}

			d, err := ToDuration(args[0])
			if err != nil {
				return dune.NullValue, err
			}

			v := args[1]
			switch v.Type {
			case dune.Func:

			case dune.Object:
				if _, ok := v.ToObjectOrNil().(*dune.Closure); !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", v.TypeName())
				}

			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", v.TypeName())
			}

			ticker := time.NewTicker(d)

			go func() {
				for range ticker.C {
					if err := runAsyncFuncOrClosure(vm, v); err != nil {
						fmt.Fprintln(vm.GetStderr(), err)
					}
				}
			}()

			return dune.NewObject(&tickerObj{ticker}), nil
		},
	},
	{
		Name:      "time.newTimer",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if !vm.HasPermission("async") {
				return dune.NullValue, ErrUnauthorized
			}

			d, err := ToDuration(args[0])
			if err != nil {
				return dune.NullValue, fmt.Errorf("expected time.Duration, got: %s", args[0].TypeName())
			}

			v := args[1]
			switch v.Type {
			case dune.Func:
			case dune.Object:
			default:
				return dune.NullValue, fmt.Errorf("%v is not a function", v.TypeName())
			}

			timer := time.NewTimer(d)

			go func() {
				for range timer.C {
					if err := runAsyncFuncOrClosure(vm, v); err != nil {
						fmt.Fprintln(vm.GetStderr(), err)
					}
				}
			}()

			return dune.NewObject(&timerObj{timer}), nil
		},
	},
	{
		Name:      "time.after",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			l := len(args)
			if l == 0 || l > 2 {
				return dune.NullValue, fmt.Errorf("expected 1 or 2 args")
			}

			d, err := ToDuration(args[0])
			if err != nil {
				return dune.NullValue, fmt.Errorf("expected time.Duration, got: %s", args[0].TypeName())
			}

			ch := make(chan dune.Value)
			timer := time.NewTimer(d)

			if l == 1 {
				go func() {
					t := <-timer.C
					ch <- dune.NewObject(TimeObj(t))
				}()
			} else {
				go func() {
					<-timer.C
					ch <- args[1]
				}()
			}

			c := &channel{c: ch}
			return dune.NewObject(c), nil
		},
	},
}

func GetLocation(vm *dune.VM) *time.Location {
	if vm.Location != nil {
		return vm.Location
	}
	return time.Local
}

func parseDate(value, format string, loc *time.Location) (time.Time, error) {
	var formats []string

	if format != "" {
		formats = []string{format}
	} else {
		formats = []string{
			"2006-01-02",
			"2006-01-02T15:04",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04:05Z07:00",
		}
	}

	for _, f := range formats {
		t, err := time.ParseInLocation(f, value, loc)
		if err == nil {
			// GO BUG?? Sometimes it creates a timezone without name!!
			if t.Location().String() == "" {
				t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
			}
			return t, nil
		}
	}

	return time.Time{}, dune.NewPublicError(fmt.Sprintf("Error parsing date: %s", value))
}

func getDate(args []dune.Value, vm *dune.VM, defaultLoc *time.Location) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.Int, dune.Int,
		dune.Int, dune.Int, dune.Int, dune.Int, dune.Object); err != nil {
		return dune.NullValue, err
	}

	var loc *time.Location
	var year, month, day, hour, min, sec int

	switch len(args) {
	case 0:
		year = 1
		month = 1
		day = 1
		loc = defaultLoc

	case 1:
		year = int(args[0].ToInt())
		month = 1
		day = 1
		loc = defaultLoc

	case 2:
		year = int(args[0].ToInt())
		month = int(args[1].ToInt())
		day = 1
		loc = defaultLoc

	case 3:
		year = int(args[0].ToInt())
		month = int(args[1].ToInt())
		day = int(args[2].ToInt())
		loc = defaultLoc

	case 4:
		year = int(args[0].ToInt())
		month = int(args[1].ToInt())
		day = int(args[2].ToInt())
		hour = int(args[3].ToInt())
		loc = defaultLoc

	case 5:
		year = int(args[0].ToInt())
		month = int(args[1].ToInt())
		day = int(args[2].ToInt())
		hour = int(args[3].ToInt())
		min = int(args[4].ToInt())
		loc = defaultLoc

	case 6:
		year = int(args[0].ToInt())
		month = int(args[1].ToInt())
		day = int(args[2].ToInt())
		hour = int(args[3].ToInt())
		min = int(args[4].ToInt())
		sec = int(args[5].ToInt())
		loc = defaultLoc

	case 7:
		year = int(args[0].ToInt())
		month = int(args[1].ToInt())
		day = int(args[2].ToInt())
		hour = int(args[3].ToInt())
		min = int(args[4].ToInt())
		sec = int(args[5].ToInt())
		location, ok := args[6].ToObjectOrNil().(location)
		if !ok {
			return dune.NullValue, fmt.Errorf("invalid location, got %s", args[6].TypeName())
		}
		loc = location.l
	}

	d := time.Date(year, time.Month(month), day, hour, min, sec, 0, loc)
	return dune.NewObject(TimeObj(d)), nil
}

func parseInLocation(l *time.Location, this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 1, 2); err != nil {
		return dune.NullValue, err
	}
	if err := ValidateOptionalArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}

	value := args[0].String()

	var format string

	if len(args) == 2 {
		farg := args[1]
		switch farg.Type {
		case dune.String:
			format = farg.String()
		case dune.Null:
		default:
			return dune.NullValue, ErrInvalidType
		}
	}

	t, err := parseDate(value, format, l)
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewObject(TimeObj(t)), nil
}

type TimeObj time.Time

func (t TimeObj) Type() string {
	return "time"
}

func (t TimeObj) Size() int {
	return 1
}

func (t TimeObj) String() string {
	return time.Time(t).Format(time.RFC3339)
}

func (t TimeObj) Export(recursionLevel int) interface{} {
	return time.Time(t)
}

func (t TimeObj) Equals(v interface{}) bool {
	other, ok := v.(TimeObj)
	if !ok {
		return false
	}

	return time.Time(t).Equal(time.Time(other))
}

func (t TimeObj) Compare(v dune.Value) int {
	other, ok := v.ToObjectOrNil().(TimeObj)
	if !ok {
		return -2
	}

	t1 := time.Time(t)

	t2 := time.Time(other)

	if t1.Before(t2) {
		return -1
	}

	if t1.Equal(t2) {
		return 0
	}

	return 1
}

func (t TimeObj) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "unix":
		return dune.NewInt64(time.Time(t).Unix()), nil
	case "second":
		return dune.NewInt(time.Time(t).Second()), nil
	case "nanosecond":
		return dune.NewInt(time.Time(t).Nanosecond()), nil
	case "minute":
		return dune.NewInt(time.Time(t).Minute()), nil
	case "hour":
		return dune.NewInt(time.Time(t).Hour()), nil
	case "day":
		return dune.NewInt(time.Time(t).Day()), nil
	case "dayOfWeek":
		return dune.NewInt(int(time.Time(t).Weekday())), nil
	case "month":
		return dune.NewInt(int(time.Time(t).Month())), nil
	case "year":
		return dune.NewInt(time.Time(t).Year()), nil
	case "yearDay":
		return dune.NewInt(time.Time(t).YearDay()), nil
	case "location":
		l := time.Time(t).Location()
		if l == nil {
			return dune.NullValue, nil
		}
		return dune.NewObject(location{l}), nil
	}

	return dune.UndefinedValue, nil
}

func (t TimeObj) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "sub":
		return t.sub
	case "add":
		return t.add
	case "setDate":
		return t.setDate
	case "addDate":
		return t.addDateValues
	case "setTime":
		return t.setTime
	case "setTimeMillis":
		return t.setTimeMillis
	case "addMilliseconds":
		return t.addMilliseconds
	case "addSeconds":
		return t.addSeconds
	case "addMinutes":
		return t.addMinutes
	case "addHours":
		return t.addHours
	case "addYears":
		return t.addYears
	case "addMonths":
		return t.addMonths
	case "addDays":
		return t.addDays
	case "format":
		return t.format
	case "utc":
		return t.utc
	case "local":
		return t.local
	case "time":
		return t.time
	case "startOfDay":
		return t.startOfDay
	case "endOfDay":
		return t.endOfDay
	case "setLocation":
		return t.setLocation
	case "in":
		return t.in
	case "after":
		return t.after
	case "afterOrEqual":
		return t.afterOrEqual
	case "before":
		return t.before
	case "beforeOrEqual":
		return t.beforeOrEqual
	case "between":
		return t.between
	case "equal":
		return t.equal
	case "sameDay":
		return t.sameDay
	case "toISO":
		return t.toISO
	}
	return nil
}

func (t TimeObj) setDate(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	ln := len(args)
	if ln > 3 {
		return dune.NullValue, fmt.Errorf("expected max 3 arguments, got %d", ln)
	}

	tt := time.Time(t)

	var year int
	var month time.Month
	var day int

	if ln >= 1 {
		a := args[0]
		switch a.Type {
		case dune.Null, dune.Undefined:
			year = tt.Year()
		case dune.Int:
			year = int(a.ToInt())
		default:
			return dune.NullValue, ErrInvalidType
		}
	}

	if ln >= 2 {
		a := args[1]
		switch a.Type {
		case dune.Null, dune.Undefined:
			month = tt.Month()
		case dune.Int:
			month = time.Month(a.ToInt())
		default:
			return dune.NullValue, ErrInvalidType
		}
	}

	if ln >= 3 {
		a := args[2]
		switch a.Type {
		case dune.Null, dune.Undefined:
			day = tt.Day()
		case dune.Int:
			day = int(a.ToInt())
		default:
			return dune.NullValue, ErrInvalidType
		}
	}

	date := time.Date(year, month, day, tt.Hour(), tt.Minute(), tt.Second(), tt.Nanosecond(), tt.Location())
	return dune.NewObject(TimeObj(date)), nil
}

func (t TimeObj) addDateValues(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Int, dune.Int, dune.Int); err != nil {
		return dune.NullValue, err
	}

	tt := time.Time(t).AddDate(int(args[0].ToInt()), int(args[1].ToInt()), int(args[2].ToInt()))
	return dune.NewObject(TimeObj(tt)), nil
}

func (t TimeObj) setTimeMillis(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Int); err != nil {
		return dune.NullValue, err
	}

	millis := int(args[0].ToInt())

	hour := millis / (60 * 60 * 1000)

	mod := millis % (60 * 60 * 1000)

	min := mod / (60 * 1000)

	secMod := mod % (60 * 1000)

	sec := secMod / 1000

	ms := secMod % 1000

	loc := GetLocation(vm)

	// always operate with time in local time
	tt := time.Time(t).In(loc)

	date := time.Date(tt.Year(), tt.Month(), tt.Day(), hour, min, sec, ms, tt.Location())

	return dune.NewObject(TimeObj(date)), nil
}

func (t TimeObj) setTime(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.Int, dune.Int, dune.Int, dune.Int); err != nil {
		return dune.NullValue, err
	}

	var hour, min, sec, milli int

	l := len(args)

	if l >= 1 {
		hour = int(args[0].ToInt())
	}

	if l >= 2 {
		min = int(args[1].ToInt())
	}

	if l >= 3 {
		sec = int(args[2].ToInt())
	}

	if l >= 4 {
		milli = int(args[3].ToInt())
	}

	loc := GetLocation(vm)

	// always operate with time in local time
	tt := time.Time(t).In(loc)

	date := time.Date(tt.Year(), tt.Month(), tt.Day(), hour, min, sec, milli, tt.Location())

	return dune.NewObject(TimeObj(date)), nil
}

func (t TimeObj) sameDay(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t2, ok := args[0].ToObject().(TimeObj)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected time.Time, got %s", args[0].TypeName())
	}

	tt1 := time.Time(t)
	tt2 := time.Time(t2)
	eq := tt1.Year() == tt2.Year() && tt1.Month() == tt2.Month() && tt1.Day() == tt2.Day()
	return dune.NewBool(eq), nil
}

func (t TimeObj) toISO(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	str := time.Time(t).Format("2006-01-02T15:04:05")
	return dune.NewString(str), nil
}

func (t TimeObj) equal(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t2, ok := args[0].ToObject().(TimeObj)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected time.Time, got %s", args[0].TypeName())
	}

	eq := time.Time(t).Equal(time.Time(t2))
	return dune.NewBool(eq), nil
}

func (t TimeObj) after(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t2, ok := args[0].ToObject().(TimeObj)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected time.Time, got %s", args[0].TypeName())
	}

	after := time.Time(t).After(time.Time(t2))
	return dune.NewBool(after), nil
}

func (t TimeObj) afterOrEqual(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t2, ok := args[0].ToObject().(TimeObj)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected time.Time, got %s", args[0].TypeName())
	}

	after := !time.Time(t).Before(time.Time(t2))
	return dune.NewBool(after), nil
}

func (t TimeObj) before(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t2, ok := args[0].ToObject().(TimeObj)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected time.Time, got %s", args[0].TypeName())
	}

	after := time.Time(t).Before(time.Time(t2))
	return dune.NewBool(after), nil
}

func (t TimeObj) beforeOrEqual(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t2, ok := args[0].ToObject().(TimeObj)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected time.Time, got %s", args[0].TypeName())
	}

	after := !time.Time(t).After(time.Time(t2))
	return dune.NewBool(after), nil
}

func (t TimeObj) between(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t1, ok := args[0].ToObject().(TimeObj)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected time.Time, got %s", args[0].TypeName())
	}

	t2, ok := args[1].ToObject().(TimeObj)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected time.Time, got %s", args[1].TypeName())
	}

	t0 := time.Time(t)

	between := !t0.Before(time.Time(t1)) && !t0.After(time.Time(t2))

	return dune.NewBool(between), nil
}

// SetLocation returns exactly the same time but with a different location
func (t TimeObj) setLocation(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	loc, ok := args[0].ToObject().(location)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected location, got %s", args[0].TypeName())
	}

	a := time.Time(t)
	b := time.Date(a.Year(), a.Month(), a.Day(),
		a.Hour(), a.Minute(), a.Second(), a.Nanosecond(), loc.l)

	return dune.NewObject(TimeObj(b)), nil
}

func (t TimeObj) in(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	loc, ok := args[0].ToObject().(location)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected location, got %s", args[0].TypeName())
	}

	tt := time.Time(t).In(loc.l)

	return dune.NewObject(TimeObj(tt)), nil
}

func (t TimeObj) startOfDay(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	loc := GetLocation(vm)

	tt := time.Time(t).In(loc)

	// construct a new date ignoring the time part
	u := TimeObj(time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, loc))

	return dune.NewObject(u), nil
}

func (t TimeObj) endOfDay(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	loc := GetLocation(vm)

	tt := time.Time(t).In(loc)

	u := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, loc)
	u = u.AddDate(0, 0, 1).Add(-1 * time.Second)
	return dune.NewObject(TimeObj(u)), nil
}

// return the milliseconds of the time part
func (t TimeObj) time(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	loc := GetLocation(vm)

	// always operate with time in local time
	tt := time.Time(t).In(loc)

	millis := (tt.Hour() * 60 * 60 * 1000) + (tt.Minute() * 60 * 1000) + (tt.Second() * 1000)

	return dune.NewInt(millis), nil
}

func (t TimeObj) utc(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	u := TimeObj(time.Time(t).UTC())

	return dune.NewObject(u), nil
}

func (t TimeObj) local(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 0 {
		return dune.NullValue, fmt.Errorf("expected 0 arguments, got %d", len(args))
	}

	loc := GetLocation(vm)

	l := TimeObj(time.Time(t).In(loc))
	return dune.NewObject(l), nil
}

func (t TimeObj) sub(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	var a = args[0]
	if a.Type != dune.Object {
		return dune.NullValue, fmt.Errorf("expected time, got %s", a.TypeName())
	}

	at, ok := a.ToObject().(TimeObj)
	if !ok {
		return dune.NullValue, fmt.Errorf("expected time, got %s", a.TypeName())
	}

	d := time.Time(t).Sub(time.Time(at))

	return dune.NewObject(Duration(d)), nil
}

func (t TimeObj) add(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	var a = args[0]
	var ad time.Duration

	switch a.Type {
	case dune.Int:
		dd, err := ToDuration(a)
		if err != nil {
			return dune.NullValue, err
		}
		ad = dd
	case dune.Object:
		dur, ok := a.ToObject().(Duration)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected duration, got %s", a.TypeName())
		}
		ad = time.Duration(dur)
	}

	d := time.Time(t).Add(ad)

	return dune.NewObject(TimeObj(d)), nil
}

func parseAddArg(a dune.Value) (int, error) {
	switch a.Type {
	case dune.Int:
		return int(a.ToInt()), nil
	case dune.Float:
		f := a.ToFloat()
		if f != float64(int(f)) {
			return 0, fmt.Errorf("expected int, got %s", a.TypeName())
		}
		return int(f), nil
	default:
		return 0, fmt.Errorf("expected int, got %s", a.TypeName())
	}

}

func (t TimeObj) addYears(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	i, err := parseAddArg(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	return t.addDate(int(i), 0, 0)
}

func (t TimeObj) addMonths(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	i, err := parseAddArg(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	return t.addDate(0, int(i), 0)
}

func (t TimeObj) addDays(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	i, err := parseAddArg(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	return t.addDate(0, 0, int(i))
}

func (t TimeObj) addHours(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	i, err := parseAddArg(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	d := time.Time(t).Add(time.Duration(i) * time.Hour)

	return dune.NewObject(TimeObj(d)), nil
}

func (t TimeObj) addMinutes(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	i, err := parseAddArg(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	d := time.Time(t).Add(time.Duration(i) * time.Minute)
	return dune.NewObject(TimeObj(d)), nil
}

func (t TimeObj) addMilliseconds(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	i, err := parseAddArg(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	d := time.Time(t).Add(time.Duration(i) * time.Millisecond)
	return dune.NewObject(TimeObj(d)), nil
}

func (t TimeObj) addSeconds(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	i, err := parseAddArg(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	d := time.Time(t).Add(time.Duration(i) * time.Second)
	return dune.NewObject(TimeObj(d)), nil
}

func (t TimeObj) addDate(years, months, days int) (dune.Value, error) {
	d := time.Time(t).AddDate(years, months, days)
	return dune.NewObject(TimeObj(d)), nil
}

func (t TimeObj) format(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	format := args[0].String()
	s := time.Time(t)

	loc := vm.Localizer
	if loc == nil {
		loc = defaultLocalizer
	}

	formatted := loc.Format(vm.Language, format, s)

	return dune.NewString(formatted), nil
}

type Duration time.Duration

func (t Duration) Type() string {
	return "duration"
}

func (t Duration) Size() int {
	return 1
}

func (t Duration) Export(recursionLevel int) interface{} {
	return time.Duration(t)
}

func (t Duration) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "hours":
		return dune.NewFloat(time.Duration(t).Hours()), nil
	case "minutes":
		return dune.NewFloat(time.Duration(t).Minutes()), nil
	case "seconds":
		return dune.NewFloat(time.Duration(t).Seconds()), nil
	case "milliseconds":
		return dune.NewInt64(time.Duration(t).Nanoseconds() / 1000000), nil
	case "nanoseconds":
		return dune.NewInt64(time.Duration(t).Nanoseconds()), nil
	}

	return dune.UndefinedValue, nil
}

func (t Duration) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "equal":
		return t.equal
	case "greater":
		return t.greater
	case "lesser":
		return t.lesser
	case "add":
		return t.add
	case "sub":
		return t.sub
	case "multiply":
		return t.multiply
	}
	return nil
}

func (t Duration) add(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	d, err := ToDuration(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(Duration(time.Duration(t) + d)), nil
}

func (t Duration) sub(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	d, err := ToDuration(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(Duration(time.Duration(t) - d)), nil
}

func (t Duration) multiply(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	d, err := ToDuration(args[0])
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(Duration(time.Duration(t) * d)), nil
}

func (t Duration) equal(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	a := args[0]
	var t2 int

	switch a.Type {
	case dune.Object:
		d, ok := a.ToObject().(Duration)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected time.Duration, got %s", a.TypeName())
		}
		t2 = int(d)

	case dune.Int:
		t2 = int(a.ToInt())

	case dune.Float:
		t2 = int(a.ToInt())

	default:
		return dune.NullValue, fmt.Errorf("expected time.Duration, got %s", a.TypeName())
	}

	eq := int(t) == t2
	return dune.NewBool(eq), nil
}

func (t Duration) greater(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	a := args[0]
	var t2 int

	switch a.Type {
	case dune.Object:
		d, ok := a.ToObject().(Duration)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected time.Duration, got %s", a.TypeName())
		}
		t2 = int(d)

	case dune.Int:
		t2 = int(a.ToInt())

	case dune.Float:
		t2 = int(a.ToInt())

	default:
		return dune.NullValue, fmt.Errorf("expected time.Duration, got %s", a.TypeName())
	}

	eq := int(t) > t2
	return dune.NewBool(eq), nil
}

func (t Duration) lesser(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	a := args[0]
	var t2 int

	switch a.Type {
	case dune.Object:
		d, ok := a.ToObject().(Duration)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected time.Duration, got %s", a.TypeName())
		}
		t2 = int(d)

	case dune.Int:
		t2 = int(a.ToInt())

	case dune.Float:
		t2 = int(a.ToInt())

	default:
		return dune.NullValue, fmt.Errorf("expected time.Duration, got %s", a.TypeName())
	}

	eq := int(t) < t2
	return dune.NewBool(eq), nil
}

type location struct {
	l *time.Location
}

func (l location) Type() string {
	return "time.Location"
}

func (l location) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "name":
		name := l.l.String()
		if name == "Local" {
			name = ""
		}
		return dune.NewString(name), nil
	}

	return dune.UndefinedValue, nil
}

func ToDuration(v dune.Value) (time.Duration, error) {
	switch v.Type {
	case dune.Object:
		d, ok := v.ToObject().(Duration)
		if !ok {
			return 0, fmt.Errorf("expected time.Duration, got %s", v.TypeName())
		}
		return time.Duration(d), nil

	case dune.Int:
		return time.Duration(v.ToInt()), nil

	default:
		return 0, fmt.Errorf("expected time.Duration, got %s", v.TypeName())
	}
}

type timerObj struct {
	timer *time.Timer
}

func (t *timerObj) Type() string {
	return "sync.Timer"
}

func (t *timerObj) Size() int {
	return 1
}

func (t *timerObj) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "reset":
		return t.reset
	case "stop":
		return t.stop
	}
	return nil
}

func (t *timerObj) reset(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 arg, got 0")
	}

	d, err := ToDuration(args[0])
	if err != nil {
		return dune.NullValue, fmt.Errorf("expected time.Duration, got: %s", args[0].TypeName())
	}

	t.timer.Reset(d)
	return dune.NullValue, nil
}

func (t *timerObj) stop(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	t.timer.Stop()
	return dune.NullValue, nil
}

type tickerObj struct {
	ticker *time.Ticker
}

func (t *tickerObj) Type() string {
	return "sync.Ticker"
}

func (t *tickerObj) Size() int {
	return 1
}

func (t *tickerObj) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "stop":
		return t.stop
	}
	return nil
}

func (t *tickerObj) stop(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	t.ticker.Stop()
	return dune.NullValue, nil
}
