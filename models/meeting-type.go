package models

import (
	"database/sql/driver"
	"errors"
)

// MeetingType enum
type MeetingType string

const (
	// BasicMeeting represents a regular session
	BasicMeeting MeetingType = "basic"
	// ExamMeeting represents an exam session
	ExamMeeting MeetingType = "exam"
)

// Scan implements the sql.Scanner interface for MeetingType
func (mt *MeetingType) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*mt = MeetingType(string(v))
		return nil
	case string:
		*mt = MeetingType(v)
		return nil
	}
	return errors.New("failed to scan MeetingType: invalid type")
}

// Value implements the driver.Valuer interface for MeetingType
func (mt MeetingType) Value() (driver.Value, error) {
	return string(mt), nil
}
