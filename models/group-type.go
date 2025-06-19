package models

import (
	"database/sql/driver"
	"errors"
)

// GroupType enum
type GroupType string

const (
	// MahasiswaGunadarma represents a GroupType of Gunadarma University students
	MahasiswaGunadarma GroupType = "mahasiswa_gunadarma"
	// MahasiswaNonGunadarma represents a GroupType of non-Gunadarma University students
	MahasiswaNonGunadarma GroupType = "mahasiswa_non_gunadarma"
	// Umum represents a general GroupType that is not specific to any university
	Umum GroupType = "umum"
)

// Scan implements the sql.Scanner interface for GroupType
func (gt *GroupType) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan GroupType: type assertion to []byte failed")
	}
	*gt = GroupType(string(b))
	return nil
}

// Value implements the driver.Valuer interface for GroupType
func (gt GroupType) Value() (driver.Value, error) {
	return string(gt), nil
}
