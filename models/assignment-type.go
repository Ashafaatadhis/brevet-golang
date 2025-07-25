package models

import (
	"database/sql/driver"
	"errors"
)

// AssignmentType custom enum type
type AssignmentType string

const (
	// Essay assignment type
	Essay AssignmentType = "essay"
	// File assignment type
	File AssignmentType = "file"
)

// Scan scans the value from the database into the AssignmentType
func (a *AssignmentType) Scan(value any) error {

	switch v := value.(type) {
	case []byte:
		*a = AssignmentType(string(v))
		return nil
	case string:
		*a = AssignmentType(v)
		return nil
	}
	return errors.New("failed to scan AssignmentType: unsupported type")

}

// Value returns the value of the AssignmentType
func (a AssignmentType) Value() (driver.Value, error) {
	return string(a), nil
}
