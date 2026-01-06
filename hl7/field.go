// Package hl7 provides core types and utilities for HL7 v2.x message processing.
package hl7

import (
	"fmt"
	"strconv"
	"strings"
)

// Field represents a single HL7 field which may contain repetitions.
// Fields are separated by the field delimiter (|) and may contain multiple
// repetitions separated by the repetition delimiter (~).
type Field interface {
	// SeqNum returns the field sequence number (1-based per HL7 standard).
	SeqNum() int

	// Value returns the string value of the field.
	// If the field has repetitions, returns the value of the first repetition.
	Value() string

	// Component returns the component at the given 1-based index from the first repetition.
	// Returns false if the index is out of range.
	Component(index int) (Component, bool)

	// Components returns all components from the first repetition.
	Components() []Component

	// Repetition returns the repetition at the given 0-based index.
	// Returns false if the index is out of range.
	Repetition(index int) (Repetition, bool)

	// Repetitions returns all repetitions in this field.
	Repetitions() []Repetition

	// RepetitionCount returns the number of repetitions in this field.
	RepetitionCount() int

	// Get retrieves a value at the specified location within the field.
	// Location format: ".component.subcomponent" or "[rep].component.subcomponent"
	// Examples: ".1", ".1.2", "[0].1", "[1].2.3"
	Get(location string) (string, error)

	// Set updates a value at the specified location within the field.
	// Location format: ".component.subcomponent" or "[rep].component.subcomponent"
	Set(location string, value string) error

	// Bytes encodes the field to bytes using the provided delimiters.
	// Repetitions are joined with the Repetition delimiter (~).
	Bytes(delims *Delimiters) []byte

	// String returns the string representation of the field.
	String() string
}

// field is the concrete implementation of Field.
type field struct {
	seqNum      int
	repetitions []Repetition
	value       []rune // raw value when no repetitions parsed
}

// NewField creates a new Field with the given sequence number and string value.
// The value is stored as-is without parsing for repetitions.
func NewField(seq int, value string) Field {
	return &field{
		seqNum: seq,
		value:  []rune(value),
	}
}

// ParseField creates a Field from a rune slice, parsing repetitions.
// Repetitions are split on the Repetition delimiter (~).
func ParseField(seq int, data []rune, delims *Delimiters) (Field, error) {
	if delims == nil {
		delims = DefaultDelimiters()
	}

	f := &field{
		seqNum: seq,
	}

	// If data is empty, return empty field
	if len(data) == 0 {
		f.value = []rune{}
		return f, nil
	}

	// Split on repetition delimiter
	repDelim := delims.Repetition
	var reps [][]rune
	start := 0

	for i, r := range data {
		if r == repDelim {
			reps = append(reps, data[start:i])
			start = i + 1
		}
	}
	// Add the last segment
	reps = append(reps, data[start:])

	// If only one segment with no repetition delimiter found, store as raw value
	// but also parse as a single repetition to support component access
	if len(reps) == 1 {
		// Parse as a single repetition to enable component access
		rep, err := ParseRepetition(data, delims)
		if err != nil {
			return nil, err
		}
		f.repetitions = []Repetition{rep}
		return f, nil
	}

	// Create repetitions (parse each for components)
	f.repetitions = make([]Repetition, len(reps))
	for i, rep := range reps {
		r, err := ParseRepetition(rep, delims)
		if err != nil {
			return nil, err
		}
		f.repetitions[i] = r
	}

	return f, nil
}

// SeqNum returns the field sequence number.
func (f *field) SeqNum() int {
	return f.seqNum
}

// Value returns the string value of the field.
// Returns the full encoded value including all repetitions, components,
// and subcomponents with their delimiters.
func (f *field) Value() string {
	return f.String()
}

// Component returns the component at the given 1-based index from the first repetition.
// Returns false if the index is out of range or if there are no repetitions.
func (f *field) Component(index int) (Component, bool) {
	if len(f.repetitions) == 0 {
		return nil, false
	}
	return f.repetitions[0].Component(index)
}

// Components returns all components from the first repetition.
// Returns an empty slice if no repetitions have been parsed.
func (f *field) Components() []Component {
	if len(f.repetitions) == 0 {
		return []Component{}
	}
	return f.repetitions[0].Components()
}

// Repetition returns the repetition at the given 0-based index.
// Returns false if the index is out of range.
func (f *field) Repetition(index int) (Repetition, bool) {
	// Repetitions use 0-based indexing
	if index < 0 || index >= len(f.repetitions) {
		return nil, false
	}
	return f.repetitions[index], true
}

// Repetitions returns all repetitions in this field.
// Returns an empty slice if no repetitions have been parsed.
func (f *field) Repetitions() []Repetition {
	if f.repetitions == nil {
		return []Repetition{}
	}
	// Return a copy to prevent external modification
	result := make([]Repetition, len(f.repetitions))
	copy(result, f.repetitions)
	return result
}

// RepetitionCount returns the number of repetitions in this field.
func (f *field) RepetitionCount() int {
	return len(f.repetitions)
}

// Get retrieves a value at the specified location within the field.
// Location format: ".component" or ".component.subcomponent" or "[rep].component.subcomponent"
// Examples: ".1", ".1.2", "[0].1", "[1].2.3"
// Returns an empty string if the location is not found.
func (f *field) Get(location string) (string, error) {
	if location == "" {
		return f.Value(), nil
	}

	loc, err := parseFieldLocation(location)
	if err != nil {
		return "", err
	}

	// Determine which repetition to access
	repIdx := 0
	if loc.Repetition >= 0 {
		repIdx = loc.Repetition
	}

	// Get the repetition
	rep, ok := f.Repetition(repIdx)
	if !ok {
		return "", nil // Repetition not found returns empty string
	}

	// Get the component if specified
	if loc.Component > 0 {
		comp, ok := rep.Component(loc.Component)
		if !ok {
			return "", nil // Component not found returns empty string
		}

		// Get the subcomponent if specified
		if loc.SubComponent > 0 {
			subComp, ok := comp.SubComponent(loc.SubComponent)
			if !ok {
				return "", nil // SubComponent not found returns empty string
			}
			return subComp.Value(), nil
		}

		return comp.Value(), nil
	}

	return rep.Value(), nil
}

// Set updates a value at the specified location within the field.
// Location format: ".component" or ".component.subcomponent" or "[rep].component.subcomponent"
// If the repetition doesn't exist, it will be created along with intermediate repetitions.
func (f *field) Set(location string, value string) error {
	if location == "" {
		// Set the entire field value - create a single repetition to support
		// both Value() and Repetition(0) access patterns
		f.value = nil
		f.repetitions = []Repetition{NewRepetition(value)}
		return nil
	}

	loc, err := parseFieldLocation(location)
	if err != nil {
		return err
	}

	// Determine which repetition to access
	repIdx := 0
	if loc.Repetition >= 0 {
		repIdx = loc.Repetition
	}

	// Ensure we have enough repetitions
	f.ensureRepetitions(repIdx + 1)

	// If no component is specified, set the repetition value
	if loc.Component <= 0 {
		// Replace the entire repetition with a new one containing the value
		f.repetitions[repIdx] = NewRepetition(value)
		return nil
	}

	// Get the repetition as a mutable concrete type
	rep := f.repetitions[repIdx].(*repetition)

	// If we have a raw value but no components, convert to component model
	if len(rep.components) == 0 && len(rep.value) > 0 {
		rep.components = []Component{NewComponent(string(rep.value))}
		rep.value = nil
	}

	// Ensure we have enough components
	for len(rep.components) < loc.Component {
		rep.components = append(rep.components, NewComponent(""))
	}

	// If no subcomponent is specified, set the component value
	if loc.SubComponent <= 0 {
		rep.components[loc.Component-1] = NewComponent(value)
		return nil
	}

	// Set the subcomponent value
	return rep.components[loc.Component-1].SetSubComponent(loc.SubComponent, value)
}

// ensureRepetitions ensures the field has at least n repetitions.
func (f *field) ensureRepetitions(n int) {
	// If we have a raw value but no repetitions, convert to repetition model
	if len(f.repetitions) == 0 && len(f.value) > 0 {
		f.repetitions = []Repetition{NewRepetition(string(f.value))}
		f.value = nil
	}

	// Expand repetitions slice if needed
	for len(f.repetitions) < n {
		f.repetitions = append(f.repetitions, NewRepetition(""))
	}
}

// Bytes encodes the field to bytes using the provided delimiters.
// Repetitions are joined with the Repetition delimiter (~).
func (f *field) Bytes(delims *Delimiters) []byte {
	if delims == nil {
		delims = DefaultDelimiters()
	}

	// If no repetitions, return raw value
	if len(f.repetitions) == 0 {
		return []byte(string(f.value))
	}

	// Build output with repetition delimiter
	var result []byte
	for i, rep := range f.repetitions {
		if i > 0 {
			result = append(result, byte(delims.Repetition))
		}
		result = append(result, rep.Bytes(delims)...)
	}

	return result
}

// String returns the string representation of the field.
// This encodes the field using default delimiters.
func (f *field) String() string {
	return string(f.Bytes(DefaultDelimiters()))
}

// fieldLocation holds parsed location components for field-level queries.
type fieldLocation struct {
	Repetition   int // 0-based repetition index, -1 for first/default
	Component    int // 1-based component index, 0 for not specified
	SubComponent int // 1-based subcomponent index, 0 for not specified
}

// parseFieldLocation parses a field-level location string.
// Format: "[rep].component.subcomponent" or ".component.subcomponent"
// Examples: ".1", ".1.2", "[0].1", "[1].2.3"
func parseFieldLocation(s string) (*fieldLocation, error) {
	loc := &fieldLocation{
		Repetition:   -1, // default to first repetition
		Component:    0,
		SubComponent: 0,
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return loc, nil
	}

	// Check for repetition index [n]
	if strings.HasPrefix(s, "[") {
		endBracket := strings.Index(s, "]")
		if endBracket == -1 {
			return nil, &LocationError{Location: s, Reason: "missing closing bracket"}
		}
		repStr := s[1:endBracket]
		rep, err := strconv.Atoi(repStr)
		if err != nil {
			return nil, &LocationError{Location: s, Reason: fmt.Sprintf("invalid repetition index: %s", repStr)}
		}
		if rep < 0 {
			return nil, &LocationError{Location: s, Reason: "repetition index must be non-negative"}
		}
		loc.Repetition = rep
		s = s[endBracket+1:]
	}

	// Remove leading dot if present
	s = strings.TrimPrefix(s, ".")

	if s == "" {
		return loc, nil
	}

	// Split remaining parts by dot
	parts := strings.Split(s, ".")

	// Parse component (first part)
	if len(parts) >= 1 && parts[0] != "" {
		comp, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, &LocationError{Location: s, Reason: fmt.Sprintf("invalid component: %s", parts[0])}
		}
		if comp < 0 {
			return nil, &LocationError{Location: s, Reason: "component must be non-negative"}
		}
		loc.Component = comp
	}

	// Parse subcomponent (second part)
	if len(parts) >= 2 && parts[1] != "" {
		subComp, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, &LocationError{Location: s, Reason: fmt.Sprintf("invalid subcomponent: %s", parts[1])}
		}
		if subComp < 0 {
			return nil, &LocationError{Location: s, Reason: "subcomponent must be non-negative"}
		}
		loc.SubComponent = subComp
	}

	return loc, nil
}
