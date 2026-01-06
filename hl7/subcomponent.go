// Package hl7 provides core types and utilities for HL7 v2.x message processing.
package hl7

// SubComponent is the atomic data unit in HL7 v2.x messages.
// It represents the smallest addressable element in the message hierarchy.
type SubComponent interface {
	// Value returns the string value of the subcomponent.
	Value() string

	// Set updates the subcomponent value.
	Set(value string) error

	// Bytes encodes the subcomponent to bytes using the provided delimiters.
	Bytes(delims *Delimiters) []byte

	// String returns the string representation (same as Value).
	String() string
}

// subComponent is the concrete implementation of SubComponent.
// It stores the value as runes to properly handle Unicode characters.
type subComponent struct {
	value []rune
}

// NewSubComponent creates a new SubComponent with the given string value.
func NewSubComponent(value string) SubComponent {
	return &subComponent{
		value: []rune(value),
	}
}

// ParseSubComponent creates a SubComponent from a rune slice.
// This is used during parsing to efficiently create subcomponents
// without intermediate string conversions.
func ParseSubComponent(data []rune) SubComponent {
	// Make a copy to avoid sharing the underlying array
	valueCopy := make([]rune, len(data))
	copy(valueCopy, data)
	return &subComponent{
		value: valueCopy,
	}
}

// Value returns the string value of the subcomponent.
func (sc *subComponent) Value() string {
	return string(sc.value)
}

// Set updates the subcomponent value.
func (sc *subComponent) Set(value string) error {
	sc.value = []rune(value)
	return nil
}

// Bytes encodes the subcomponent to bytes using the provided delimiters.
// The delimiters parameter is accepted for interface consistency with other
// HL7 element types, but subcomponents do not contain nested delimiters.
func (sc *subComponent) Bytes(_ *Delimiters) []byte {
	return []byte(string(sc.value))
}

// String returns the string representation of the subcomponent.
// This is equivalent to Value().
func (sc *subComponent) String() string {
	return sc.Value()
}
