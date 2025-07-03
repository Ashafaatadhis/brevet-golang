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
	switch v := value.(type) {
	case string:
		*c = CourseType(v)
		return nil
	case []byte:
		*c = CourseType(string(v))
		return nil
	default:
		return errors.New("failed to scan CourseType: unsupported type")
	}
}

// Value implements the driver.Valuer interface
func (c CourseType) Value() (driver.Value, error) {
	return string(c), nil
}
