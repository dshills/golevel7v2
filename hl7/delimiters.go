// Package hl7 provides core types and utilities for HL7 v2.x message processing.
package hl7

import (
	"errors"
	"fmt"
)

// Segment terminator constant
const (
	// SegmentTerminator is the carriage return character that terminates HL7 segments.
	SegmentTerminator = '\r' // 0x0D
)

// Standard HL7 delimiter defaults
const (
	DefaultFieldDelimiter        = '|'
	DefaultComponentDelimiter    = '^'
	DefaultRepetitionDelimiter   = '~'
	DefaultEscapeDelimiter       = '\\'
	DefaultSubComponentDelimiter = '&'
	DefaultTruncationDelimiter   = '#'
)

// Minimum MSH segment length requirements
const (
	// minMSHLength is the minimum length for a valid MSH segment header (MSH + field delimiter + 4 encoding chars)
	minMSHLength = 8
)

// Errors returned by delimiter parsing
var (
	ErrEmptyInput       = errors.New("empty input")
	ErrNotMSHSegment    = errors.New("segment does not start with MSH")
	ErrMSHTooShort      = errors.New("MSH segment too short to contain delimiters")
	ErrMissingDelimiter = errors.New("missing required delimiter in MSH-2")
)

// Delimiters holds the HL7 message delimiter characters.
// These are extracted from MSH-1 (field separator) and MSH-2 (encoding characters).
type Delimiters struct {
	Field        rune // MSH-1: Field separator (default: |)
	Component    rune // First char of MSH-2: Component separator (default: ^)
	Repetition   rune // Second char of MSH-2: Repetition separator (default: ~)
	Escape       rune // Third char of MSH-2: Escape character (default: \)
	SubComponent rune // Fourth char of MSH-2: Sub-component separator (default: &)
	Truncation   rune // Fifth char of MSH-2 (optional): Truncation character (default: #)
}

// DefaultDelimiters returns a Delimiters instance with standard HL7 v2.x values.
//
// Standard delimiters:
//   - Field: | (pipe)
//   - Component: ^ (caret)
//   - Repetition: ~ (tilde)
//   - Escape: \ (backslash)
//   - SubComponent: & (ampersand)
//   - Truncation: # (hash)
func DefaultDelimiters() *Delimiters {
	return &Delimiters{
		Field:        DefaultFieldDelimiter,
		Component:    DefaultComponentDelimiter,
		Repetition:   DefaultRepetitionDelimiter,
		Escape:       DefaultEscapeDelimiter,
		SubComponent: DefaultSubComponentDelimiter,
		Truncation:   DefaultTruncationDelimiter,
	}
}

// ParseDelimiters extracts delimiters from an MSH segment.
//
// The MSH segment format is:
//
//	MSH|^~\&|...
//	   │└─┴─┴─┴── MSH-2: Encoding characters (component, repetition, escape, subcomponent, [truncation])
//	   └──────── MSH-1: Field separator
//
// The function expects at least the first 8 bytes of the MSH segment.
// If the truncation character (5th encoding char) is not present, it defaults to '#'.
func ParseDelimiters(mshSegment []byte) (*Delimiters, error) {
	if len(mshSegment) == 0 {
		return nil, ErrEmptyInput
	}

	// Verify segment starts with "MSH"
	if len(mshSegment) < 3 || string(mshSegment[:3]) != "MSH" {
		return nil, ErrNotMSHSegment
	}

	// Need at least MSH + field delimiter + 4 encoding characters
	if len(mshSegment) < minMSHLength {
		return nil, ErrMSHTooShort
	}

	d := &Delimiters{
		Field:      rune(mshSegment[3]), // MSH-1: character immediately after "MSH"
		Component:  rune(mshSegment[4]), // First encoding character
		Repetition: rune(mshSegment[5]), // Second encoding character
		Escape:     rune(mshSegment[6]), // Third encoding character
		Truncation: DefaultTruncationDelimiter,
	}

	// MSH-2 ends at the next field separator or end of segment
	// Find the end of MSH-2 to get SubComponent and optional Truncation
	msh2Start := 4
	msh2End := msh2Start

	for i := msh2Start; i < len(mshSegment); i++ {
		if rune(mshSegment[i]) == d.Field || mshSegment[i] == byte(SegmentTerminator) {
			msh2End = i
			break
		}
		msh2End = i + 1
	}

	msh2Len := msh2End - msh2Start

	// MSH-2 must have at least 4 characters (component, repetition, escape, subcomponent)
	if msh2Len < 4 {
		return nil, fmt.Errorf("%w: expected at least 4 characters, got %d", ErrMissingDelimiter, msh2Len)
	}

	d.SubComponent = rune(mshSegment[7]) // Fourth encoding character

	// Truncation character is optional (HL7 v2.7+)
	if msh2Len >= 5 {
		d.Truncation = rune(mshSegment[8])
	}

	return d, nil
}

// String returns the encoding characters string (MSH-2 value).
// This includes component, repetition, escape, subcomponent, and truncation characters.
//
// Example: "^~\\&#" for default delimiters
func (d *Delimiters) String() string {
	return fmt.Sprintf("%c%c%c%c%c",
		d.Component,
		d.Repetition,
		d.Escape,
		d.SubComponent,
		d.Truncation,
	)
}

// EncodingCharacters returns the encoding characters string (MSH-2 value).
// This is an alias for String() for clarity when working with HL7 terminology.
func (d *Delimiters) EncodingCharacters() string {
	return d.String()
}

// MSH1 returns the field separator character as a string (MSH-1 value).
func (d *Delimiters) MSH1() string {
	return string(d.Field)
}

// MSH2 returns the encoding characters string (MSH-2 value).
// This is an alias for EncodingCharacters().
func (d *Delimiters) MSH2() string {
	return d.EncodingCharacters()
}

// Equal returns true if two Delimiters instances have the same values.
func (d *Delimiters) Equal(other *Delimiters) bool {
	if d == nil || other == nil {
		return d == other
	}
	return d.Field == other.Field &&
		d.Component == other.Component &&
		d.Repetition == other.Repetition &&
		d.Escape == other.Escape &&
		d.SubComponent == other.SubComponent &&
		d.Truncation == other.Truncation
}
