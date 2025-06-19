package models

import (
	"database/sql/driver"
	"errors"
)

// DayType custom enum type
type DayType string

const (
	// Monday represents the first day of the week
	Monday DayType = "monday"
	// Tuesday represents the second day of the week
	Tuesday DayType = "tuesday"
	// Wednesday represents the third day of the week
	Wednesday DayType = "wednesday"
	// Thursday represents the fourth day of the week
	Thursday DayType = "thursday"
	// Friday represents the fifth day of the week
	Friday DayType = "friday"
	// Saturday represents the sixth day of the week
	Saturday DayType = "saturday"
	// Sunday represents the seventh day of the week
	Sunday DayType = "sunday"
)

// Scan scans the value from the database into the DayType type
func (d *DayType) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan DayType: type assertion to []byte failed")
	}
	*d = DayType(string(b))
	return nil
}

// Value returns the value of the DayType type for storing in the database
func (d DayType) Value() (driver.Value, error) {
	return string(d), nil
}
