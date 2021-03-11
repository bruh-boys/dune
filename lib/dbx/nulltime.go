package dbx

import (
	"fmt"
	"time"
)

// NullTime scans null time values
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
// The value type must be time.Time otherwise Scan fails.
func (t *NullTime) Scan(value interface{}) error {
	if value == nil {
		t.Time, t.Valid = time.Time{}, false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time, t.Valid = v, true
		return nil
	}

	t.Valid = false
	return fmt.Errorf("can't convert value to time.Time")
}
