package models

import (
	"database/sql/driver"
	"errors"
)

// AttendanceStatus is a custom type for attendance status
type AttendanceStatus string

const (
	// Present is present
	Present AttendanceStatus = "present"
	// Absent is absent
	Absent AttendanceStatus = "absent"
	// Late is late
	Late AttendanceStatus = "late"
)

// Scan implements the sql.Scanner interface
func (s *AttendanceStatus) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan AttendanceStatus: type assertion to []byte failed")
	}
	*s = AttendanceStatus(string(b))
	return nil
}

// Value implements the driver.Valuer interface
func (s AttendanceStatus) Value() (driver.Value, error) {
	return string(s), nil
}
