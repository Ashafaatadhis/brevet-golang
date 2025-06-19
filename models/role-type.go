package models

import (
	"database/sql/driver"
	"errors"
)

// RoleType defines valid RoleTypes in the system
type RoleType string

const (
	// RoleTypeSiswa represents a student RoleType
	RoleTypeSiswa RoleType = "siswa"
	// RoleTypeGuru represents a teacher RoleType
	RoleTypeGuru RoleType = "guru"
	// RoleTypeAdmin represents an administrator RoleType
	RoleTypeAdmin RoleType = "admin"
)

// Scan implements the sql.Scanner interface
func (r *RoleType) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*r = RoleType(string(v))
		return nil
	case string:
		*r = RoleType(v)
		return nil
	}
	return errors.New("failed to scan RoleType: unsupported type")
}

// Value implements the driver.Valuer interface
func (r RoleType) Value() (driver.Value, error) {
	return string(r), nil
}
