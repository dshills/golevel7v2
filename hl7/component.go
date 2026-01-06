// Package hl7 provides core types and utilities for HL7 v2.x message processing.
package hl7

// Component represents a component within an HL7 field or repetition.
// Components are separated by the component delimiter (^) and may contain
// subcomponents separated by the subcomponent delimiter (&).
type Component interface {
	// Value returns the string value of the component.
	// If the component has subcomponents, this returns the value of the first subcomponent.
	Value() string

	// SubComponent returns the subcomponent at the given 1-based index.
	// Returns false if the index is out of range.
	SubComponent(index int) (SubComponent, bool)

	// SubComponents returns all subcomponents in this component.
	SubComponents() []SubComponent

	// Set updates the component value, replacing all subcomponents.
	Set(value string) error

	// SetSubComponent sets the value of a subcomponent at the given 1-based index.
	// If the index is beyond the current subcomponent count, intermediate
	// empty subcomponents are created.
	SetSubComponent(index int, value string) error

	// Bytes encodes the component to bytes using the provided delimiters.
	Bytes(delims *Delimiters) []byte

	// String returns the string representation of the component.
	String() string
}

// component is the concrete implementation of Component.
type component struct {
	subComponents []SubComponent
	value         []rune // raw value when no subcomponents parsed
}

// NewComponent creates a new Component with the given string value.
// The value is stored as-is without parsing for subcomponents.
func NewComponent(value string) Component {
	return &component{
		value: []rune(value),
	}
}

// ParseComponent creates a Component from a rune slice, parsing subcomponents.
// Subcomponents are split on the SubComponent delimiter (&).
func ParseComponent(data []rune, delims *Delimiters) (Component, error) {
	if delims == nil {
		delims = DefaultDelimiters()
	}

	c := &component{}

	// If data is empty, return empty component
	if len(data) == 0 {
		c.value = []rune{}
		return c, nil
	}

	// Split on subcomponent delimiter
	subCompDelim := delims.SubComponent
	var subComps [][]rune
	start := 0

	for i, r := range data {
		if r == subCompDelim {
			subComps = append(subComps, data[start:i])
			start = i + 1
		}
	}
	// Add the last segment
	subComps = append(subComps, data[start:])

	// If only one segment with no subcomponent delimiter found, store as raw value
	if len(subComps) == 1 {
		valueCopy := make([]rune, len(data))
		copy(valueCopy, data)
		c.value = valueCopy
		return c, nil
	}

	// Create subcomponents
	c.subComponents = make([]SubComponent, len(subComps))
	for i, sc := range subComps {
		c.subComponents[i] = ParseSubComponent(sc)
	}

	return c, nil
}

// Value returns the string value of the component.
// If the component has subcomponents, returns the value of the first subcomponent.
// Otherwise, returns the raw value.
func (c *component) Value() string {
	if len(c.subComponents) > 0 {
		return c.subComponents[0].Value()
	}
	return string(c.value)
}

// SubComponent returns the subcomponent at the given 1-based index.
// Returns false if the index is out of range or if there are no subcomponents.
func (c *component) SubComponent(index int) (SubComponent, bool) {
	// HL7 uses 1-based indexing
	if index < 1 || index > len(c.subComponents) {
		return nil, false
	}
	return c.subComponents[index-1], true
}

// SubComponents returns all subcomponents in this component.
// Returns an empty slice if no subcomponents have been parsed.
func (c *component) SubComponents() []SubComponent {
	if c.subComponents == nil {
		return []SubComponent{}
	}
	// Return a copy to prevent external modification
	result := make([]SubComponent, len(c.subComponents))
	copy(result, c.subComponents)
	return result
}

// Set updates the component value, replacing all subcomponents.
func (c *component) Set(value string) error {
	c.value = []rune(value)
	c.subComponents = nil
	return nil
}

// SetSubComponent sets the value of a subcomponent at the given 1-based index.
// If the index is beyond the current subcomponent count, intermediate
// empty subcomponents are created.
func (c *component) SetSubComponent(index int, value string) error {
	if index < 1 {
		return ErrInvalidIndex
	}

	// If we have a raw value but no subcomponents, convert to subcomponent model
	if len(c.subComponents) == 0 && len(c.value) > 0 {
		c.subComponents = []SubComponent{NewSubComponent(string(c.value))}
		c.value = nil
	}

	// Expand subcomponents slice if needed
	for len(c.subComponents) < index {
		c.subComponents = append(c.subComponents, NewSubComponent(""))
	}

	c.subComponents[index-1] = NewSubComponent(value)
	return nil
}

// Bytes encodes the component to bytes using the provided delimiters.
// If the component has subcomponents, they are joined with the SubComponent delimiter.
func (c *component) Bytes(delims *Delimiters) []byte {
	if delims == nil {
		delims = DefaultDelimiters()
	}

	// If no subcomponents, return raw value
	if len(c.subComponents) == 0 {
		return []byte(string(c.value))
	}

	// Build output with subcomponent delimiter
	var result []byte
	for i, sc := range c.subComponents {
		if i > 0 {
			result = append(result, byte(delims.SubComponent))
		}
		result = append(result, sc.Bytes(delims)...)
	}

	return result
}

// String returns the string representation of the component.
// This encodes the component using default delimiters.
func (c *component) String() string {
	return string(c.Bytes(DefaultDelimiters()))
}
