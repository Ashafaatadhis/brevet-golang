package models

import (
	"database/sql/driver"
	"errors"
)

// MaterialType enum
type MaterialType string

const (
	// MaterialPDF represents a MaterialType PDF
	MaterialPDF MaterialType = "pdf"
	// MaterialWord represents a MaterialType Word
	MaterialWord MaterialType = "word"
	// MaterialPPT represents a MaterialType Powerpoint
	MaterialPPT MaterialType = "ppt"
	// MaterialLink represents a MaterialType Link
	MaterialLink MaterialType = "link"
)

// Scan implements the sql.Scanner interface for GroupType
func (mt *MaterialType) Scan(value any) error {

	switch v := value.(type) {
	case []byte:
		*mt = MaterialType(string(v))
		return nil
	case string:
		*mt = MaterialType(v)
		return nil
	}
	return errors.New("failed to scan MaterialType: type assertion to []byte failed")
}

// Value implements the driver.Valuer interface for GroupType
func (mt MaterialType) Value() (driver.Value, error) {
	return string(mt), nil
}
