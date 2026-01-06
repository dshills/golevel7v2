package hl7

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Location represents a position within an HL7 message structure.
// It follows the HL7 path notation: SEG[idx].field[rep].component.subcomponent
//
// Field, Component, and SubComponent are 1-based per the HL7 standard.
// SegmentIndex and Repetition are 0-based.
// A value of -1 indicates "first" or "all" depending on context.
//
// Location string format examples:
//   - "PID" -> segment only
//   - "PID.5" -> segment + field
//   - "PID.5.1" -> segment + field + component
//   - "PID.5.1.2" -> full path
//   - "PID[1].5" -> second PID segment, field 5
//   - "PID.5[0].1" -> field 5, first repetition, component 1
type Location struct {
	Segment      string // Segment name (e.g., "PID", "MSH")
	SegmentIndex int    // 0-based segment index, -1 for first/all
	Field        int    // 1-based field number, -1 for all
	Repetition   int    // 0-based repetition index, -1 for first/all
	Component    int    // 1-based component number, -1 for all
	SubComponent int    // 1-based subcomponent number, -1 for all
}

// Common errors returned by location parsing and validation.
var (
	ErrEmptyLocation       = errors.New("location string is empty")
	ErrInvalidSegment      = errors.New("invalid segment name")
	ErrInvalidField        = errors.New("invalid field number")
	ErrInvalidComponent    = errors.New("invalid component number")
	ErrInvalidSubComponent = errors.New("invalid subcomponent number")
	ErrInvalidFormat       = errors.New("invalid location format")
)

// Note: ErrInvalidIndex is defined in errors.go

// segmentPattern validates segment names (3 uppercase letters or alphanumeric starting with letter).
var segmentPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{2}$`)

// locationPattern parses the full location string.
// Format: SEG[idx].field[rep].component.subcomponent
var locationPattern = regexp.MustCompile(
	`^([A-Z][A-Z0-9]{2})` + // Segment name (required)
		`(?:\[(\d+)\])?` + // Optional segment index [idx]
		`(?:\.(\d+)` + // Optional field .field
		`(?:\[(\d+)\])?` + // Optional repetition [rep]
		`(?:\.(\d+)` + // Optional component .component
		`(?:\.(\d+))?)?)?$`, // Optional subcomponent .subcomponent
)

// NewLocation creates a new Location with the given parameters.
// SegmentIndex and Repetition default to -1 (first/all).
// Pass -1 for field, component, or subcomponent to indicate "all" or "not specified".
func NewLocation(segment string, field, component, subcomponent int) *Location {
	return &Location{
		Segment:      strings.ToUpper(segment),
		SegmentIndex: -1,
		Field:        field,
		Repetition:   -1,
		Component:    component,
		SubComponent: subcomponent,
	}
}

// NewLocationFull creates a new Location with all parameters explicitly set.
func NewLocationFull(segment string, segmentIndex, field, repetition, component, subcomponent int) *Location {
	return &Location{
		Segment:      strings.ToUpper(segment),
		SegmentIndex: segmentIndex,
		Field:        field,
		Repetition:   repetition,
		Component:    component,
		SubComponent: subcomponent,
	}
}

// ParseLocation parses a location string into a Location struct.
// Format: SEG[idx].field[rep].component.subcomponent
//
// Examples:
//   - "PID" -> segment only
//   - "PID.5" -> segment + field
//   - "PID.5.1" -> segment + field + component
//   - "PID.5.1.2" -> full path
//   - "PID[1].5" -> second PID segment, field 5
//   - "PID.5[0].1" -> field 5, first repetition, component 1
func ParseLocation(s string) (*Location, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrEmptyLocation
	}

	matches := locationPattern.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidFormat, s)
	}

	loc := &Location{
		Segment:      matches[1],
		SegmentIndex: -1,
		Field:        -1,
		Repetition:   -1,
		Component:    -1,
		SubComponent: -1,
	}

	// Parse segment index [idx]
	if matches[2] != "" {
		idx, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, fmt.Errorf("%w: segment index %q", ErrInvalidIndex, matches[2])
		}
		if idx < 0 {
			return nil, fmt.Errorf("%w: segment index must be non-negative", ErrInvalidIndex)
		}
		loc.SegmentIndex = idx
	}

	// Parse field number
	if matches[3] != "" {
		field, err := strconv.Atoi(matches[3])
		if err != nil {
			return nil, fmt.Errorf("%w: %q", ErrInvalidField, matches[3])
		}
		if field < 0 {
			return nil, fmt.Errorf("%w: field must be non-negative", ErrInvalidField)
		}
		loc.Field = field
	}

	// Parse repetition [rep]
	if matches[4] != "" {
		rep, err := strconv.Atoi(matches[4])
		if err != nil {
			return nil, fmt.Errorf("%w: repetition %q", ErrInvalidIndex, matches[4])
		}
		if rep < 0 {
			return nil, fmt.Errorf("%w: repetition must be non-negative", ErrInvalidIndex)
		}
		loc.Repetition = rep
	}

	// Parse component number
	if matches[5] != "" {
		comp, err := strconv.Atoi(matches[5])
		if err != nil {
			return nil, fmt.Errorf("%w: %q", ErrInvalidComponent, matches[5])
		}
		if comp < 0 {
			return nil, fmt.Errorf("%w: component must be non-negative", ErrInvalidComponent)
		}
		loc.Component = comp
	}

	// Parse subcomponent number
	if matches[6] != "" {
		subcomp, err := strconv.Atoi(matches[6])
		if err != nil {
			return nil, fmt.Errorf("%w: %q", ErrInvalidSubComponent, matches[6])
		}
		if subcomp < 0 {
			return nil, fmt.Errorf("%w: subcomponent must be non-negative", ErrInvalidSubComponent)
		}
		loc.SubComponent = subcomp
	}

	return loc, nil
}

// String converts the Location back to a location string.
// The output format matches the input format for ParseLocation.
func (l *Location) String() string {
	if l == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(l.Segment)

	if l.SegmentIndex >= 0 {
		sb.WriteString(fmt.Sprintf("[%d]", l.SegmentIndex))
	}

	if l.Field >= 0 {
		sb.WriteString(fmt.Sprintf(".%d", l.Field))

		if l.Repetition >= 0 {
			sb.WriteString(fmt.Sprintf("[%d]", l.Repetition))
		}

		if l.Component >= 0 {
			sb.WriteString(fmt.Sprintf(".%d", l.Component))

			if l.SubComponent >= 0 {
				sb.WriteString(fmt.Sprintf(".%d", l.SubComponent))
			}
		}
	}

	return sb.String()
}

// IsValid returns true if the location has a valid structure.
// A valid location must have:
//   - A valid segment name (3 uppercase alphanumeric characters starting with a letter)
//   - If specified, field >= 0
//   - If specified, component >= 0 (and field must be specified)
//   - If specified, subcomponent >= 0 (and component must be specified)
func (l *Location) IsValid() bool {
	if l == nil {
		return false
	}

	// Segment must be valid
	if !segmentPattern.MatchString(l.Segment) {
		return false
	}

	// SegmentIndex must be -1 or non-negative
	if l.SegmentIndex < -1 {
		return false
	}

	// Field must be -1 or non-negative
	if l.Field < -1 {
		return false
	}

	// Repetition must be -1 or non-negative
	if l.Repetition < -1 {
		return false
	}

	// Component requires field
	if l.Component >= 0 && l.Field < 0 {
		return false
	}

	// Component must be -1 or non-negative
	if l.Component < -1 {
		return false
	}

	// SubComponent requires component
	if l.SubComponent >= 0 && l.Component < 0 {
		return false
	}

	// SubComponent must be -1 or non-negative
	if l.SubComponent < -1 {
		return false
	}

	// Repetition without field doesn't make sense
	if l.Repetition >= 0 && l.Field < 0 {
		return false
	}

	return true
}

// HasSegment returns true if the location specifies a segment.
func (l *Location) HasSegment() bool {
	if l == nil {
		return false
	}
	return l.Segment != ""
}

// HasField returns true if the location specifies a field (field >= 0).
func (l *Location) HasField() bool {
	if l == nil {
		return false
	}
	return l.Field >= 0
}

// HasComponent returns true if the location specifies a component (component >= 0).
func (l *Location) HasComponent() bool {
	if l == nil {
		return false
	}
	return l.Component >= 0
}

// HasSubComponent returns true if the location specifies a subcomponent (subcomponent >= 0).
func (l *Location) HasSubComponent() bool {
	if l == nil {
		return false
	}
	return l.SubComponent >= 0
}

// HasSegmentIndex returns true if the location specifies a segment index (segmentIndex >= 0).
func (l *Location) HasSegmentIndex() bool {
	if l == nil {
		return false
	}
	return l.SegmentIndex >= 0
}

// HasRepetition returns true if the location specifies a repetition (repetition >= 0).
func (l *Location) HasRepetition() bool {
	if l == nil {
		return false
	}
	return l.Repetition >= 0
}

// Equal returns true if two locations are equal.
func (l *Location) Equal(other *Location) bool {
	if l == nil && other == nil {
		return true
	}
	if l == nil || other == nil {
		return false
	}
	return l.Segment == other.Segment &&
		l.SegmentIndex == other.SegmentIndex &&
		l.Field == other.Field &&
		l.Repetition == other.Repetition &&
		l.Component == other.Component &&
		l.SubComponent == other.SubComponent
}

// Clone creates a deep copy of the location.
func (l *Location) Clone() *Location {
	if l == nil {
		return nil
	}
	return &Location{
		Segment:      l.Segment,
		SegmentIndex: l.SegmentIndex,
		Field:        l.Field,
		Repetition:   l.Repetition,
		Component:    l.Component,
		SubComponent: l.SubComponent,
	}
}

// Depth returns the depth of specificity:
// 0 = segment only, 1 = field, 2 = component, 3 = subcomponent
func (l *Location) Depth() int {
	if l == nil || !l.HasSegment() {
		return -1
	}
	if l.HasSubComponent() {
		return 3
	}
	if l.HasComponent() {
		return 2
	}
	if l.HasField() {
		return 1
	}
	return 0
}
