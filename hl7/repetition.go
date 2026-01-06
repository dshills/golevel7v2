// Package hl7 provides core types and utilities for HL7 v2.x message processing.
package hl7

// Repetition represents a single instance within a repeating HL7 field.
// Fields can repeat using the repetition delimiter (~), and each repetition
// may contain multiple components separated by the component delimiter (^).
type Repetition interface {
	// Value returns the string value of the repetition.
	// If the repetition has components, returns the value of the first component.
	Value() string

	// Component returns the component at the given 1-based index.
	// Returns false if the index is out of range.
	Component(index int) (Component, bool)

	// Components returns all components in this repetition.
	Components() []Component

	// Bytes encodes the repetition to bytes using the provided delimiters.
	Bytes(delims *Delimiters) []byte

	// String returns the string representation of the repetition.
	String() string
}

// repetition is the concrete implementation of Repetition.
type repetition struct {
	components []Component
	value      []rune // raw value when no components parsed
}

// NewRepetition creates a new Repetition with the given string value.
// The value is stored as-is without parsing for components.
func NewRepetition(value string) Repetition {
	return &repetition{
		value: []rune(value),
	}
}

// ParseRepetition creates a Repetition from a rune slice, parsing components.
// Components are split on the Component delimiter (^).
func ParseRepetition(data []rune, delims *Delimiters) (Repetition, error) {
	if delims == nil {
		delims = DefaultDelimiters()
	}

	r := &repetition{}

	// If data is empty, return empty repetition
	if len(data) == 0 {
		r.value = []rune{}
		return r, nil
	}

	// Split on component delimiter
	compDelim := delims.Component
	var comps [][]rune
	start := 0

	for i, char := range data {
		if char == compDelim {
			comps = append(comps, data[start:i])
			start = i + 1
		}
	}
	// Add the last segment
	comps = append(comps, data[start:])

	// If only one segment with no component delimiter found, store as raw value
	if len(comps) == 1 {
		valueCopy := make([]rune, len(data))
		copy(valueCopy, data)
		r.value = valueCopy
		return r, nil
	}

	// Create components (parse each for subcomponents)
	r.components = make([]Component, len(comps))
	for i, comp := range comps {
		c, err := ParseComponent(comp, delims)
		if err != nil {
			return nil, err
		}
		r.components[i] = c
	}

	return r, nil
}

// Value returns the string value of the repetition.
// Returns the full encoded value including all components with their delimiters.
func (r *repetition) Value() string {
	if len(r.components) > 0 {
		return r.String()
	}
	return string(r.value)
}

// Component returns the component at the given 1-based index.
// Returns false if the index is out of range.
// If no components have been parsed but a raw value exists, component index 1
// returns the raw value parsed as a component (supporting subcomponent access).
func (r *repetition) Component(index int) (Component, bool) {
	// HL7 uses 1-based indexing
	if index < 1 {
		return nil, false
	}

	// If we have parsed components, use them
	if len(r.components) > 0 {
		if index > len(r.components) {
			return nil, false
		}
		return r.components[index-1], true
	}

	// If no components but we have a raw value, treat it as component 1
	// This allows subcomponent access (e.g., "ID&SubID" -> component 1 has subcomponents)
	if index == 1 && len(r.value) > 0 {
		// Parse the raw value as a component to enable subcomponent access
		comp, err := ParseComponent(r.value, nil)
		if err != nil {
			return nil, false
		}
		return comp, true
	}

	return nil, false
}

// Components returns all components in this repetition.
// Returns an empty slice if no components have been parsed.
func (r *repetition) Components() []Component {
	if r.components == nil {
		return []Component{}
	}
	// Return a copy to prevent external modification
	result := make([]Component, len(r.components))
	copy(result, r.components)
	return result
}

// Bytes encodes the repetition to bytes using the provided delimiters.
// If the repetition has components, they are joined with the Component delimiter.
func (r *repetition) Bytes(delims *Delimiters) []byte {
	if delims == nil {
		delims = DefaultDelimiters()
	}

	// If no components, return raw value
	if len(r.components) == 0 {
		return []byte(string(r.value))
	}

	// Build output with component delimiter
	var result []byte
	for i, comp := range r.components {
		if i > 0 {
			result = append(result, byte(delims.Component))
		}
		result = append(result, comp.Bytes(delims)...)
	}

	return result
}

// String returns the string representation of the repetition.
// This encodes the repetition using default delimiters.
func (r *repetition) String() string {
	return string(r.Bytes(DefaultDelimiters()))
}
