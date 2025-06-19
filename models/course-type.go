package models

import (
	"database/sql/driver"
	"errors"
)

// CourseType enum type
type CourseType string

const (
	// CourseTypeOnline represents an online course
	CourseTypeOnline CourseType = "online"
	// CourseTypeOffline represents an offline course
	CourseTypeOffline CourseType = "offline"
)

// Scan implements the sql.Scanner interface
func (c *CourseType) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan CourseType: type assertion to []byte failed")
	}
	*c = CourseType(string(b))
	return nil
}

// Value implements the driver.Valuer interface
func (c CourseType) Value() (driver.Value, error) {
	return string(c), nil
}
