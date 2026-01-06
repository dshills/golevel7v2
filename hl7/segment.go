// Package hl7 provides core types and utilities for HL7 v2.x message processing.
package hl7

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Segment represents an HL7 segment (e.g., MSH, PID, OBX).
// A segment is a logical grouping of fields that contains related information.
type Segment interface {
	// Name returns the 3-letter segment identifier (e.g., "MSH", "PID").
	Name() string

	// Field returns the field at the 1-based sequence number.
	// Returns false if the field doesn't exist.
	Field(seq int) (Field, bool)

	// Fields returns all fields at the same sequence number (for repeating fields).
	Fields(seq int) []Field

	// AllFields returns all fields in the segment.
	AllFields() []Field

	// FieldCount returns the number of fields in the segment.
	FieldCount() int

	// Get retrieves a value at the specified location.
	// Format: ".field.component.subcomponent" (e.g., ".5.1" for field 5, component 1)
	Get(location string) (string, error)

	// GetAll retrieves all values at the specified location (for repeating fields).
	GetAll(location string) ([]string, error)

	// Set sets a value at the specified location.
	Set(location string, value string) error

	// SetField sets the field at the 1-based sequence number.
	SetField(seq int, field Field) error

	// AddField adds a field to the segment.
	AddField(field Field) error

	// Bytes encodes the segment to bytes using the provided delimiters.
	Bytes(delims *Delimiters) []byte

	// String returns the string representation.
	String() string
}

// segment is the concrete implementation of Segment.
type segment struct {
	name   string
	fields []Field
	value  []rune
}

// NewSegment creates a new empty segment with the given name.
// The name should be a 3-letter segment identifier (e.g., "MSH", "PID").
func NewSegment(name string) Segment {
	return &segment{
		name:   strings.ToUpper(name),
		fields: make([]Field, 0),
	}
}

// ParseSegment parses a segment from a rune slice using the provided delimiters.
// The data should contain the entire segment without the segment terminator.
//
// MSH segments receive special handling:
//   - MSH-1 is the field separator character itself (|)
//   - MSH-2 contains the encoding characters (^~\&)
//   - Field numbering starts at 1, where MSH.1 = |, MSH.2 = ^~\&, MSH.3 = first data field
func ParseSegment(data []rune, delims *Delimiters) (Segment, error) {
	if len(data) == 0 {
		return nil, &ParseError{Message: "empty segment data"}
	}

	if delims == nil {
		delims = DefaultDelimiters()
	}

	// Find the segment name (first 3 characters or up to first delimiter)
	nameEnd := 0
	for i, r := range data {
		if r == delims.Field || r == SegmentTerminator {
			nameEnd = i
			break
		}
		nameEnd = i + 1
	}

	if nameEnd < 3 {
		return nil, &ParseError{
			Message: fmt.Sprintf("segment name too short: %q", string(data[:nameEnd])),
		}
	}

	name := strings.ToUpper(string(data[:3]))

	seg := &segment{
		name:   name,
		fields: make([]Field, 0),
		value:  make([]rune, len(data)),
	}
	copy(seg.value, data)

	// MSH segment has special handling
	if name == "MSH" {
		return parseMSHSegment(data, delims)
	}

	// Parse regular segment fields
	return parseRegularSegment(data, delims, name)
}

// parseMSHSegment handles the special parsing rules for MSH segments.
// MSH-1 is the field separator, MSH-2 is the encoding characters.
func parseMSHSegment(data []rune, delims *Delimiters) (Segment, error) {
	seg := &segment{
		name:   "MSH",
		fields: make([]Field, 0),
		value:  make([]rune, len(data)),
	}
	copy(seg.value, data)

	// MSH must have at least "MSH|" (4 characters)
	if len(data) < 4 {
		return nil, &ParseError{
			Message: "MSH segment too short",
		}
	}

	// MSH-1: Field separator (the character after "MSH")
	fieldSep := data[3]
	msh1Field := NewField(1, string(fieldSep))
	seg.fields = append(seg.fields, msh1Field)

	// If there's no more data after MSH|, we're done
	if len(data) < 5 {
		return seg, nil
	}

	// Find the end of MSH-2 (encoding characters)
	// MSH-2 ends at the next field separator
	msh2Start := 4
	msh2End := msh2Start
	for i := msh2Start; i < len(data); i++ {
		if data[i] == fieldSep {
			msh2End = i
			break
		}
		msh2End = i + 1
	}

	// MSH-2: Encoding characters
	msh2Value := string(data[msh2Start:msh2End])
	msh2Field := NewField(2, msh2Value)
	seg.fields = append(seg.fields, msh2Field)

	// Parse remaining fields (MSH-3 onwards)
	if msh2End < len(data) {
		remainingData := data[msh2End:]
		fields, err := parseSegmentFields(remainingData, delims, 3)
		if err != nil {
			return nil, err
		}
		// Skip the first empty field from the leading delimiter
		if len(fields) > 0 {
			seg.fields = append(seg.fields, fields[1:]...)
		}
	}

	return seg, nil
}

// parseRegularSegment parses a non-MSH segment.
func parseRegularSegment(data []rune, delims *Delimiters, name string) (Segment, error) {
	seg := &segment{
		name:   name,
		fields: make([]Field, 0),
		value:  make([]rune, len(data)),
	}
	copy(seg.value, data)

	// Find the first field delimiter
	firstDelim := -1
	for i, r := range data {
		if r == delims.Field {
			firstDelim = i
			break
		}
	}

	// No fields after segment name
	if firstDelim == -1 {
		return seg, nil
	}

	// Parse fields starting after segment name
	fieldData := data[firstDelim:]
	fields, err := parseSegmentFields(fieldData, delims, 1)
	if err != nil {
		return nil, err
	}

	// Skip the first empty field from the leading delimiter
	if len(fields) > 0 {
		seg.fields = append(seg.fields, fields[1:]...)
	}

	return seg, nil
}

// parseSegmentFields splits field data by the field delimiter and creates Field objects.
// startSeq is the starting sequence number for the first field.
func parseSegmentFields(data []rune, delims *Delimiters, startSeq int) ([]Field, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var fields []Field
	var current []rune
	seqNum := startSeq

	for _, r := range data {
		if r == delims.Field {
			// Create field from accumulated data
			f, err := ParseField(seqNum, current, delims)
			if err != nil {
				return nil, err
			}
			fields = append(fields, f)
			current = nil
			seqNum++
		} else {
			current = append(current, r)
		}
	}

	// Don't forget the last field
	f, err := ParseField(seqNum, current, delims)
	if err != nil {
		return nil, err
	}
	fields = append(fields, f)

	return fields, nil
}

// Name returns the 3-letter segment identifier.
func (s *segment) Name() string {
	return s.name
}

// Field returns the field at the 1-based sequence number.
// For MSH: seq 1 = field separator, seq 2 = encoding characters, seq 3+ = data fields.
func (s *segment) Field(seq int) (Field, bool) {
	if seq < 1 || seq > len(s.fields) {
		return nil, false
	}
	return s.fields[seq-1], true
}

// Fields returns all fields at the same sequence number.
// This handles repeating fields (fields with the same sequence but separated by ~).
func (s *segment) Fields(seq int) []Field {
	f, ok := s.Field(seq)
	if !ok {
		return nil
	}
	// For now, return single field; repetition handling is within Field
	return []Field{f}
}

// AllFields returns all fields in the segment.
func (s *segment) AllFields() []Field {
	result := make([]Field, len(s.fields))
	copy(result, s.fields)
	return result
}

// FieldCount returns the number of fields in the segment.
func (s *segment) FieldCount() int {
	return len(s.fields)
}

// Get retrieves a value at the specified location within the segment.
// Format: ".field" or ".field.component" or ".field.component.subcomponent"
// Field numbers are 1-based.
func (s *segment) Get(location string) (string, error) {
	loc, err := parseSegmentLocation(location)
	if err != nil {
		return "", err
	}

	// Get the field
	field, ok := s.Field(loc.field)
	if !ok {
		return "", nil // Return empty string for missing fields (common in HL7)
	}

	// If only field specified, return full field value
	if loc.component == 0 && loc.subcomponent == 0 {
		return field.Value(), nil
	}

	// Build field-level location string
	fieldLoc := ""
	if loc.component > 0 {
		fieldLoc = fmt.Sprintf(".%d", loc.component)
		if loc.subcomponent > 0 {
			fieldLoc = fmt.Sprintf(".%d.%d", loc.component, loc.subcomponent)
		}
	}

	return field.Get(fieldLoc)
}

// GetAll retrieves all values at the specified location (for repeating fields).
func (s *segment) GetAll(location string) ([]string, error) {
	loc, err := parseSegmentLocation(location)
	if err != nil {
		return nil, err
	}

	field, ok := s.Field(loc.field)
	if !ok {
		return nil, nil
	}

	// If only field specified, return values from all repetitions
	if loc.component == 0 && loc.subcomponent == 0 {
		var results []string
		for i := 0; i < field.RepetitionCount(); i++ {
			rep, _ := field.Repetition(i)
			results = append(results, rep.Value())
		}
		if len(results) == 0 {
			results = append(results, field.Value())
		}
		return results, nil
	}

	// Build field-level location string for each repetition
	var results []string
	for i := 0; i < field.RepetitionCount(); i++ {
		fieldLoc := fmt.Sprintf("[%d].%d", i, loc.component)
		if loc.subcomponent > 0 {
			fieldLoc = fmt.Sprintf("[%d].%d.%d", i, loc.component, loc.subcomponent)
		}
		val, err := field.Get(fieldLoc)
		if err != nil {
			return nil, err
		}
		results = append(results, val)
	}

	if len(results) == 0 {
		// No repetitions parsed, try direct access
		fieldLoc := fmt.Sprintf(".%d", loc.component)
		if loc.subcomponent > 0 {
			fieldLoc = fmt.Sprintf(".%d.%d", loc.component, loc.subcomponent)
		}
		val, err := field.Get(fieldLoc)
		if err != nil {
			return nil, err
		}
		results = append(results, val)
	}

	return results, nil
}

// Set sets a value at the specified location within the segment.
func (s *segment) Set(location string, value string) error {
	loc, err := parseSegmentLocation(location)
	if err != nil {
		return err
	}

	// Ensure field exists, create if necessary
	for len(s.fields) < loc.field {
		s.fields = append(s.fields, NewField(len(s.fields)+1, ""))
	}

	// If only setting field value
	if loc.component == 0 && loc.subcomponent == 0 {
		return s.fields[loc.field-1].Set("", value)
	}

	// Build field-level location string
	fieldLoc := fmt.Sprintf(".%d", loc.component)
	if loc.subcomponent > 0 {
		fieldLoc = fmt.Sprintf(".%d.%d", loc.component, loc.subcomponent)
	}

	return s.fields[loc.field-1].Set(fieldLoc, value)
}

// SetField sets the field at the 1-based sequence number.
func (s *segment) SetField(seq int, field Field) error {
	if seq < 1 {
		return &FieldError{
			Sequence: seq,
			Reason:   "sequence number must be >= 1",
		}
	}

	// Expand fields slice if necessary
	for len(s.fields) < seq {
		s.fields = append(s.fields, NewField(len(s.fields)+1, ""))
	}

	s.fields[seq-1] = field
	return nil
}

// AddField adds a field to the end of the segment.
func (s *segment) AddField(field Field) error {
	s.fields = append(s.fields, field)
	return nil
}

// Bytes encodes the segment to bytes using the provided delimiters.
// MSH segments are encoded with special handling for MSH-1 and MSH-2.
func (s *segment) Bytes(delims *Delimiters) []byte {
	if delims == nil {
		delims = DefaultDelimiters()
	}

	var buf bytes.Buffer

	// Write segment name
	buf.WriteString(s.name)

	if len(s.fields) == 0 {
		return buf.Bytes()
	}

	// MSH special handling
	if s.name == "MSH" {
		return s.encodeMSH(delims)
	}

	// Regular segment encoding
	for _, field := range s.fields {
		buf.WriteRune(delims.Field)
		if field != nil {
			buf.Write(field.Bytes(delims))
		}
	}

	return buf.Bytes()
}

// encodeMSH handles special encoding for MSH segments.
// MSH-1 is written as the field separator (not preceded by separator).
// MSH-2 contains the encoding characters.
func (s *segment) encodeMSH(delims *Delimiters) []byte {
	var buf bytes.Buffer

	buf.WriteString("MSH")

	// MSH-1: Field separator (no preceding separator)
	if len(s.fields) > 0 && s.fields[0] != nil {
		buf.WriteString(s.fields[0].Value())
	} else {
		buf.WriteRune(delims.Field)
	}

	// MSH-2: Encoding characters (no preceding separator, value written directly)
	if len(s.fields) > 1 && s.fields[1] != nil {
		buf.WriteString(s.fields[1].Value())
	}

	// MSH-3 onwards: Regular field encoding (preceded by field separator)
	for i := 2; i < len(s.fields); i++ {
		buf.WriteRune(delims.Field)
		if s.fields[i] != nil {
			buf.Write(s.fields[i].Bytes(delims))
		}
	}

	return buf.Bytes()
}

// String returns the string representation of the segment using default delimiters.
func (s *segment) String() string {
	return string(s.Bytes(DefaultDelimiters()))
}

// segmentLocation holds parsed location components for internal use.
type segmentLocation struct {
	field        int
	component    int
	subcomponent int
}

// parseSegmentLocation parses a location string like ".5.1.2" into components.
// Returns field, component, subcomponent (all 1-based, 0 means not specified).
func parseSegmentLocation(location string) (*segmentLocation, error) {
	location = strings.TrimSpace(location)
	if location == "" {
		return nil, &LocationError{
			Location: location,
			Reason:   "empty location",
		}
	}

	// Remove leading dot if present
	location = strings.TrimPrefix(location, ".")

	if location == "" {
		return nil, &LocationError{
			Location: ".",
			Reason:   "no field specified",
		}
	}

	parts := strings.Split(location, ".")
	loc := &segmentLocation{}

	// Parse field number
	field, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, &LocationError{
			Location: parts[0],
			Reason:   "invalid field number",
		}
	}
	if field < 1 {
		return nil, &LocationError{
			Location: parts[0],
			Reason:   "field number must be >= 1",
		}
	}
	loc.field = field

	// Parse component number if present
	if len(parts) > 1 && parts[1] != "" {
		comp, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, &LocationError{
				Location: parts[1],
				Reason:   "invalid component number",
			}
		}
		if comp < 1 {
			return nil, &LocationError{
				Location: parts[1],
				Reason:   "component number must be >= 1",
			}
		}
		loc.component = comp
	}

	// Parse subcomponent number if present
	if len(parts) > 2 && parts[2] != "" {
		subcomp, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, &LocationError{
				Location: parts[2],
				Reason:   "invalid subcomponent number",
			}
		}
		if subcomp < 1 {
			return nil, &LocationError{
				Location: parts[2],
				Reason:   "subcomponent number must be >= 1",
			}
		}
		loc.subcomponent = subcomp
	}

	return loc, nil
}
