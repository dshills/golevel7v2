// Package hl7 provides core types and errors for HL7 v2.x message parsing and encoding.
package hl7

import (
	"errors"
	"fmt"
)

// Severity indicates the severity level of a validation error.
type Severity int

const (
	// SeverityError indicates a critical error that prevents processing.
	SeverityError Severity = iota
	// SeverityWarning indicates a non-critical issue that should be addressed.
	SeverityWarning
	// SeverityInfo indicates an informational message.
	SeverityInfo
)

// String returns a human-readable representation of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "ERROR"
	case SeverityWarning:
		return "WARNING"
	case SeverityInfo:
		return "INFO"
	default:
		return "UNKNOWN"
	}
}

// Sentinel errors for common HL7 parsing and validation failures.
var (
	// ErrInvalidLocation indicates the location string format is invalid.
	ErrInvalidLocation = errors.New("invalid location")
	// ErrSegmentNotFound indicates the requested segment does not exist.
	ErrSegmentNotFound = errors.New("segment not found")
	// ErrFieldNotFound indicates the requested field does not exist.
	ErrFieldNotFound = errors.New("field not found")
	// ErrComponentNotFound indicates the requested component does not exist.
	ErrComponentNotFound = errors.New("component not found")
	// ErrSubComponentNotFound indicates the requested subcomponent does not exist.
	ErrSubComponentNotFound = errors.New("subcomponent not found")
	// ErrInvalidMessage indicates the message format is invalid.
	ErrInvalidMessage = errors.New("invalid message")
	// ErrEmptyMessage indicates an empty message was provided.
	ErrEmptyMessage = errors.New("empty message")
	// ErrMissingMSH indicates the required MSH segment is missing.
	ErrMissingMSH = errors.New("missing MSH segment")
	// ErrInvalidMSH indicates the MSH segment is malformed.
	ErrInvalidMSH = errors.New("invalid MSH segment")
	// ErrInvalidIndex indicates an invalid index was provided.
	ErrInvalidIndex = errors.New("invalid index")
)

// ParseError represents an error that occurred during message parsing.
type ParseError struct {
	// Message describes what went wrong.
	Message string
	// Location is the HL7 path where the error occurred (e.g., "PID-3-1").
	Location string
	// Line is the 1-based line number where the error occurred.
	Line int
	// Column is the 1-based column number where the error occurred.
	Column int
	// Cause is the underlying error that caused this parse error.
	Cause error
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	var msg string
	switch {
	case e.Location != "":
		msg = fmt.Sprintf("parse error at %s", e.Location)
	case e.Line > 0:
		msg = fmt.Sprintf("parse error at line %d, column %d", e.Line, e.Column)
	default:
		msg = "parse error"
	}

	if e.Message != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.Message)
	}

	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}

	return msg
}

// Unwrap returns the underlying cause of the parse error.
func (e *ParseError) Unwrap() error {
	return e.Cause
}

// LocationError represents an error related to an invalid HL7 location path.
type LocationError struct {
	// Location is the invalid location string.
	Location string
	// Reason describes why the location is invalid.
	Reason string
}

// Error implements the error interface.
func (e *LocationError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("invalid location %q: %s", e.Location, e.Reason)
	}
	return fmt.Sprintf("invalid location %q", e.Location)
}

// Unwrap returns the sentinel error for invalid locations.
func (e *LocationError) Unwrap() error {
	return ErrInvalidLocation
}

// ValidationError represents a validation failure for an HL7 element.
type ValidationError struct {
	// Location is the HL7 path of the element that failed validation.
	Location string
	// Rule is the name of the validation rule that failed.
	Rule string
	// Expected describes what was expected.
	Expected string
	// Actual describes what was actually found.
	Actual string
	// Severity indicates the severity level of this validation error.
	Severity Severity
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	msg := fmt.Sprintf("[%s] validation failed", e.Severity)

	if e.Location != "" {
		msg = fmt.Sprintf("%s at %s", msg, e.Location)
	}

	if e.Rule != "" {
		msg = fmt.Sprintf("%s: rule %q", msg, e.Rule)
	}

	switch {
	case e.Expected != "" && e.Actual != "":
		msg = fmt.Sprintf("%s, expected %s but got %s", msg, e.Expected, e.Actual)
	case e.Expected != "":
		msg = fmt.Sprintf("%s, expected %s", msg, e.Expected)
	case e.Actual != "":
		msg = fmt.Sprintf("%s, got %s", msg, e.Actual)
	}

	return msg
}

// Unwrap returns nil as ValidationError has no underlying cause.
func (e *ValidationError) Unwrap() error {
	return nil
}

// SegmentError represents an error related to a specific segment.
type SegmentError struct {
	// Segment is the segment name (e.g., "PID", "OBX").
	Segment string
	// Field is the field number within the segment (1-based).
	Field int
	// Reason describes what went wrong.
	Reason string
	// Cause is the underlying error.
	Cause error
}

// Error implements the error interface.
func (e *SegmentError) Error() string {
	var msg string
	if e.Field > 0 {
		msg = fmt.Sprintf("segment %s field %d", e.Segment, e.Field)
	} else {
		msg = fmt.Sprintf("segment %s", e.Segment)
	}

	if e.Reason != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.Reason)
	}

	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}

	return msg
}

// Unwrap returns the underlying cause of the segment error.
func (e *SegmentError) Unwrap() error {
	return e.Cause
}

// FieldError represents an error related to a specific field.
type FieldError struct {
	// Sequence is the field sequence number (1-based).
	Sequence int
	// Reason describes what went wrong.
	Reason string
	// Cause is the underlying error.
	Cause error
}

// Error implements the error interface.
func (e *FieldError) Error() string {
	msg := fmt.Sprintf("field %d", e.Sequence)

	if e.Reason != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.Reason)
	}

	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}

	return msg
}

// Unwrap returns the underlying cause of the field error.
func (e *FieldError) Unwrap() error {
	return e.Cause
}
