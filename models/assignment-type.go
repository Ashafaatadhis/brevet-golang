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
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan AssignmentType: type assertion to []byte failed")
	}
	*a = AssignmentType(string(b))
	return nil
}

// Value returns the value of the AssignmentType
func (a AssignmentType) Value() (driver.Value, error) {
	return string(a), nil
}
