package localization

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dunelang/dune"
)

var DefaultCulture *Culture

func init() {
	c := NewCulture("en-US")
	c.DecimalSeparator = '.'
	c.ThousandSeparator = ','
	c.NumberOfDecimals = 2
	c.CurrencySymbol = "$"
	c.CurrencyPattern = "-$0"
	c.ShortTimePattern = "HH:mm"
	c.ShortDatePattern = "MM-dd-yyyy"
	c.LongDatePattern = "ddd, MMM dd yyyy"
	c.DateMonthPattern = "ddd, MMM dd"
	c.DateMonthTimePattern = "ddd, MMM dd HH:mm"
	c.DateTimePattern = "MM-dd-yyyy HH:mm"
	c.LongDateTimePattern = "ddd, MMM dd yyyy HH:mm"
	c.FirstDayOfWeek = time.Sunday
	DefaultCulture = c
}

func NewCulture(name string) *Culture {
	return &Culture{
		Name: name,
	}
}

type Culture struct {
	Name string

	Language string

	DecimalSeparator  rune
	ThousandSeparator rune
	NumberOfDecimals  int
	CurrencySymbol    string
	CurrencyPattern   string

	DateMonthTimePattern string
	DateTimePattern      string
	LongDateTimePattern  string
	ShortDatePattern     string
	LongDatePattern      string
	DateMonthPattern     string
	ShortTimePattern     string
	FirstDayOfWeek       time.Weekday
}

func (c *Culture) Clone() *Culture {
	return &Culture{
		Name:                 c.Name,
		Language:             c.Language,
		DecimalSeparator:     c.DecimalSeparator,
		ThousandSeparator:    c.ThousandSeparator,
		NumberOfDecimals:     c.NumberOfDecimals,
		CurrencySymbol:       c.CurrencySymbol,
		CurrencyPattern:      c.CurrencyPattern,
		DateMonthTimePattern: c.DateMonthTimePattern,
		DateTimePattern:      c.DateTimePattern,
		LongDateTimePattern:  c.LongDateTimePattern,
		ShortDatePattern:     c.ShortDatePattern,
		LongDatePattern:      c.LongDatePattern,
		DateMonthPattern:     c.DateMonthPattern,
		ShortTimePattern:     c.ShortTimePattern,
		FirstDayOfWeek:       c.FirstDayOfWeek,
	}
}

func (c *Culture) Format(format string, v interface{}, translator *Translator, language string) string {
	if v == nil {
		return ""
	}

	switch t := v.(type) {
	case time.Time:
		return c.formatDate(t, format, translator, language)
	case int:
		return c.formatNum(format, float64(t))
	case int32:
		return c.formatNum(format, float64(t))
	case int64:
		return c.formatNum(format, float64(t))
	case float64:
		return c.formatNum(format, t)
	default:
		if format == "" {
			format = "%v"
		}
		return fmt.Sprintf(format, v)
	}
}

func (c *Culture) formatDate(d time.Time, format string, translator *Translator, language string) string {
	switch format {
	case "t":
		format = c.ShortTimePattern
	case "d":
		format = c.ShortDatePattern
	case "D":
		format = c.LongDatePattern
	case "H":
		format = c.DateMonthTimePattern
	case "M":
		format = c.DateMonthPattern
	case "g":
		format = c.DateTimePattern
	case "G":
		format = c.LongDateTimePattern
	}

	var parts []datePart

	for i, l := 0, len(format)-1; i <= l; i++ {
		c := format[i]
		switch c {
		case 'd':
			if i < l && format[i+1] == 'd' {
				if i < l-1 && format[i+2] == 'd' {
					parts = append(parts, datePart{token: "ddd"})
					i += 2
					continue
				}
				parts = append(parts, datePart{token: "dd"})
				i++
				continue
			}
			parts = append(parts, datePart{token: "d"})
			continue

		case 'M':
			if i < l && format[i+1] == 'M' {
				if i < l-1 && format[i+2] == 'M' {
					parts = append(parts, datePart{token: "MMM"})
					i += 2
					continue
				}
				parts = append(parts, datePart{token: "MM"})
				i++
				continue
			}
			parts = append(parts, datePart{token: "M"})
			continue

		case 'y':
			if i < l && format[i+1] == 'y' {
				if i < l-2 && format[i+2] == 'y' && format[i+3] == 'y' {
					parts = append(parts, datePart{token: "yyyy"})
					i += 3
					continue
				}
				parts = append(parts, datePart{token: "yy"})
				i++
				continue
			}
			parts = append(parts, datePart{value: string(c)})

		case 'w':
			parts = append(parts, datePart{token: "w"})
			continue

		case 'h':
			if i < l && format[i+1] == 'h' {
				parts = append(parts, datePart{token: "hh"})
				i++
				continue
			}
			parts = append(parts, datePart{token: "h"})
			continue

		case 'H':
			if i < l && format[i+1] == 'H' {
				parts = append(parts, datePart{token: "HH"})
				i++
				continue
			}
			parts = append(parts, datePart{token: "H"})
			continue

		case 'm':
			if i < l && format[i+1] == 'm' {
				parts = append(parts, datePart{token: "mm"})
				i++
				continue
			}
			parts = append(parts, datePart{token: "m"})
			continue

		case 's':
			if i < l && format[i+1] == 's' {
				parts = append(parts, datePart{token: "ss"})
				i++
				continue
			}
			parts = append(parts, datePart{token: "s"})
			continue

		case 'z':
			parts = append(parts, datePart{token: "z"})
			continue

		case 'Z':
			parts = append(parts, datePart{token: "Z"})
			continue

		case 'a':
			parts = append(parts, datePart{token: "a"})
			continue

		case '\\':
			if i < l {
				parts = append(parts, datePart{value: string(format[i+1])})
				i++
				continue
			}

		default:
			parts = append(parts, datePart{value: string(c)})
		}
	}

	var result []string

	for _, part := range parts {
		switch part.token {
		case "d":
			result = append(result, strconv.Itoa(d.Day()))

		case "dd":
			result = append(result, padLeft(d.Day()))

		case "ddd":
			day, _ := translator.Translate(language, WeekDays[d.Weekday()])
			result = append(result, day)

		case "M":
			result = append(result, strconv.Itoa(int(d.Month())))

		case "MM":
			result = append(result, padLeft(int(d.Month())))

		case "MMM":
			month, _ := translator.Translate(language, MonthNames[d.Month()-1])
			result = append(result, month)

		case "yy":
			s := strconv.Itoa(d.Year())
			if len(s) > 2 {
				s = s[len(s)-2:]
			}
			result = append(result, s)

		case "yyyy":
			result = append(result, strconv.Itoa(d.Year()))

		case "w":
			result = append(result, strconv.Itoa(int(d.Weekday())))

		case "h":
			hours := d.Hour()
			if hours > 12 {
				hours -= 12
			}
			result = append(result, strconv.Itoa(hours))

		case "hh":
			hours := d.Hour()
			if hours > 12 {
				hours -= 12
			}
			result = append(result, padLeft(hours))

		case "HH":
			result = append(result, padLeft(d.Hour()))

		case "m":
			result = append(result, strconv.Itoa(d.Minute()))

		case "mm":
			result = append(result, padLeft(d.Minute()))

		case "s":
			result = append(result, strconv.Itoa(d.Second()))

		case "ss":
			result = append(result, padLeft(d.Second()))

		case "z":
			result = append(result, d.Format("MST"))

		case "Z":
			result = append(result, d.Format("-0700"))

		case "a":
			if d.Hour() < 12 {
				result = append(result, "AM")
			} else {
				result = append(result, "PM")
			}

		case "":
			result = append(result, part.value)

		default:
			result = append(result, part.token)
		}
	}

	return strings.Join(result, "")
}

func padLeft(v int) string {
	s := strconv.Itoa(v)
	if len(s) == 1 {
		s = "0" + s
	}
	return s
}

func (c *Culture) formatNum(format string, v float64) string {
	switch format {
	case "i":
		s := strconv.FormatInt(int64(v), 10)
		b := new(bytes.Buffer)
		c.formatIntPart(s, b)
		return b.String()

	case "f":
		if c.NumberOfDecimals == 0 {
			return c.formatNum("i", v)
		}
		s := strconv.FormatFloat(v, 'f', c.NumberOfDecimals, 64)
		b := new(bytes.Buffer)
		i := strings.IndexRune(s, '.')
		c.formatIntPart(s[:i], b)
		b.WriteRune(c.DecimalSeparator)
		b.WriteString(s[i+1:])
		return b.String()

	case "c":
		s := c.formatNum("f", math.Abs(v))
		p := c.CurrencyPattern
		if v >= 0 {
			p = strings.Replace(p, "-", "", 1)
		}
		return strings.Replace(p, "0", s, 1)

	default:
		return fmt.Sprintf("[invalid '%s']", format)
	}
}

func (c *Culture) formatIntPart(v string, buf *bytes.Buffer) {
	if v[0] == '-' {
		buf.WriteRune('-')
		v = v[1:]
	}

	ln := len(v)

	for i := 0; i < ln; i++ {
		buf.WriteRune(rune(v[i]))

		if ln <= 3 || i >= ln-1 {
			continue
		}

		ix := ln - i - 1
		if ix%3 == 0 {
			buf.WriteRune(c.ThousandSeparator)
		}
	}

}

func (c *Culture) ParseNumber(v string) (float64, error) {
	if strings.ContainsRune(v, c.ThousandSeparator) {
		v = strings.Replace(v, string(c.ThousandSeparator), "", -1)
	}

	if c.DecimalSeparator != '.' {
		v = strings.Replace(v, string(c.DecimalSeparator), ".", -1)
	}

	return strconv.ParseFloat(v, 64)
}

func (c *Culture) ParseDate(value, format string, loc *time.Location) (time.Time, error) {
	var formats []string

	if format != "" {
		formats = []string{format}
	} else {
		formats = []string{
			c.ShortTimePattern,
			c.ShortDatePattern,
			c.DateTimePattern,
		}
	}

	if loc == nil {
		loc = time.Local
	}

	for _, f := range formats {
		goFormat := toGoDateFormat(f)
		t, err := time.ParseInLocation(goFormat, value, loc)
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

func toGoDateFormat(format string) string {
	var parts []string

	for i, l := 0, len(format)-1; i <= l; i++ {
		c := format[i]
		switch c {
		case 'd':
			if i < l && format[i+1] == 'd' {
				if i < l-1 && format[i+2] == 'd' {
					parts = append(parts, "Mon")
					i += 2
					continue
				}
				parts = append(parts, "02")
				i++
				continue
			}
			parts = append(parts, "2")
			continue

		case 'M':
			if i < l && format[i+1] == 'M' {
				if i < l-2 && format[i+2] == 'M' {
					parts = append(parts, "Jan")
					i += 2
					continue
				}
				parts = append(parts, "01")
				i++
				continue
			}
			parts = append(parts, "1")
			continue

		case 'y':
			if i < l && format[i+1] == 'y' {
				if i < l-2 && format[i+2] == 'y' && format[i+3] == 'y' {
					parts = append(parts, "2006")
					i += 3
					continue
				}
				parts = append(parts, "06")
				i++
				continue
			}
			parts = append(parts, string(c))

		case 'h':
			if i < l && format[i+1] == 'h' {
				parts = append(parts, "03")
				i++
				continue
			}
			parts = append(parts, "3")
			continue

		case 'H':
			if i < l && format[i+1] == 'H' {
				parts = append(parts, "15")
				i++
				continue
			}
			parts = append(parts, "3")
			continue

		case 'm':
			if i < l && format[i+1] == 'm' {
				parts = append(parts, "04")
				i++
				continue
			}
			parts = append(parts, "4")
			continue

		case 's':
			if i < l && format[i+1] == 's' {
				parts = append(parts, "05")
				i++
				continue
			}
			parts = append(parts, "5")
			continue

		case 'a':
			parts = append(parts, "PM")
			continue

		case 'z':
			parts = append(parts, "MST")
			continue

		case 'Z':
			parts = append(parts, "-0700")
			continue

		case '\\':
			if i < l {
				parts = append(parts, string(format[i+1]))
				i++
				continue
			}

		default:
			parts = append(parts, string(c))
		}
	}

	return strings.Join(parts, "")
}

const (
	ShortDatePattern     = "d"
	LongDatePattern      = "D"
	DateMonthPattern     = "M"
	DateTimePattern      = "g"
	DateMonthTimePattern = "H"
	LongDateTimePattern  = "G"
	ShortTimePattern     = "t" // HH:mm
	LongTimePattern      = "T" // HH:mm:ss
	CurrencyPattern      = "c"
	FloatPattern         = "f"
	IntPattern           = "i"
)

type datePart struct {
	token string
	value string
}

var WeekDays = []string{
	"@@Sunday",
	"@@Monday",
	"@@Tuesday",
	"@@Wednesday",
	"@@Thursday",
	"@@Friday",
	"@@Saturday",
}

var MonthNames = []string{
	"@@January",
	"@@February",
	"@@March",
	"@@April",
	"@@May",
	"@@June",
	"@@July",
	"@@August",
	"@@September",
	"@@October",
	"@@November",
	"@@December",
}
