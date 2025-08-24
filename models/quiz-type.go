package models

import (
	"database/sql/driver"
	"errors"
)

// QuizType tipe enum
type QuizType string

const (
	// QuizTypeTF type
	QuizTypeTF QuizType = "tf"
	// QuizTypeMC type
	QuizTypeMC QuizType = "mc"
)

// Scan implements the Scanner interface
func (qt *QuizType) Scan(value any) error {

	switch v := value.(type) {
	case []byte:
		*qt = QuizType(string(v))
		return nil
	case string:
		*qt = QuizType(v)
		return nil
	}
	return errors.New("failed to scan QuizType: invalid type")

}

// Value implements the Valuer interface
func (qt QuizType) Value() (driver.Value, error) {
	return string(qt), nil
}
