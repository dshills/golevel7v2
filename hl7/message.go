// Package hl7 provides core types and utilities for HL7 v2.x message processing.
package hl7

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Message-specific errors.
var (
	// ErrNilSegment indicates a nil segment was provided.
	ErrNilSegment = errors.New("segment is nil")
	// ErrIndexOutOfRange indicates an index is out of valid range.
	ErrIndexOutOfRange = errors.New("index out of range")
)

// Message represents a complete HL7 v2.x message.
// A message consists of segments separated by carriage returns.
// The first segment must be MSH (Message Header).
type Message interface {
	// Segment returns the first segment with the given name.
	// Returns false if no segment with that name exists.
	Segment(name string) (Segment, bool)

	// Segments returns all segments with the given name.
	// Returns an empty slice if no segments match.
	Segments(name string) []Segment

	// AllSegments returns all segments in the message.
	AllSegments() []Segment

	// Get returns the value at the given location string.
	// Location format: SEG[idx].field[rep].component.subcomponent
	// Examples: "PID.5", "PID.5.1", "OBX[1].5"
	Get(location string) (string, error)

	// GetAll returns all values at the given location string.
	// Useful for retrieving all repetitions or all matching segments.
	GetAll(location string) ([]string, error)

	// Set sets the value at the given location string.
	Set(location string, value string) error

	// GetAt returns the value at the given Location struct.
	GetAt(loc *Location) (string, error)

	// GetAllAt returns all values at the given Location struct.
	GetAllAt(loc *Location) ([]string, error)

	// SetAt sets the value at the given Location struct.
	SetAt(loc *Location, value string) error

	// AddSegment appends a segment to the message.
	AddSegment(seg Segment) error

	// InsertSegment inserts a segment at the given index.
	// Returns an error if the index is out of range.
	InsertSegment(index int, seg Segment) error

	// RemoveSegment removes the first segment with the given name.
	// Returns true if a segment was removed.
	RemoveSegment(name string) bool

	// Bytes returns the encoded message as bytes.
	// Segments are separated by carriage returns.
	Bytes() []byte

	// String returns the string representation of the message.
	String() string

	// Type returns the message type from MSH.9 (e.g., "ADT^A01").
	Type() string

	// ControlID returns the message control ID from MSH.10.
	ControlID() string

	// Version returns the HL7 version from MSH.12.
	Version() string

	// Delimiters returns the message delimiters.
	Delimiters() *Delimiters
}

// message is the concrete implementation of Message.
type message struct {
	segments   []Segment
	delimiters *Delimiters
}

// NewMessage creates a new Message with optional segments and delimiters.
// If segments is nil or empty, creates an empty message.
// If delims is nil, uses default delimiters.
func NewMessage(segments []Segment, delims *Delimiters) Message {
	if delims == nil {
		delims = DefaultDelimiters()
	}
	if segments == nil {
		segments = []Segment{}
	}
	return &message{
		segments:   segments,
		delimiters: delims,
	}
}

// NewEmptyMessage creates a new empty Message with default delimiters.
func NewEmptyMessage() Message {
	return &message{
		segments:   []Segment{},
		delimiters: DefaultDelimiters(),
	}
}

// NewMessageWithDelimiters creates a new empty Message with the specified delimiters.
func NewMessageWithDelimiters(delims *Delimiters) Message {
	if delims == nil {
		delims = DefaultDelimiters()
	}
	return &message{
		segments:   []Segment{},
		delimiters: delims,
	}
}

// Segment returns the first segment with the given name.
func (m *message) Segment(name string) (Segment, bool) {
	name = strings.ToUpper(name)
	for _, seg := range m.segments {
		if seg.Name() == name {
			return seg, true
		}
	}
	return nil, false
}

// Segments returns all segments with the given name.
func (m *message) Segments(name string) []Segment {
	name = strings.ToUpper(name)
	var result []Segment
	for _, seg := range m.segments {
		if seg.Name() == name {
			result = append(result, seg)
		}
	}
	if result == nil {
		return []Segment{}
	}
	return result
}

// AllSegments returns all segments in the message.
func (m *message) AllSegments() []Segment {
	result := make([]Segment, len(m.segments))
	copy(result, m.segments)
	return result
}

// Get returns the value at the given location string.
func (m *message) Get(location string) (string, error) {
	loc, err := ParseLocation(location)
	if err != nil {
		return "", err
	}
	return m.GetAt(loc)
}

// GetAll returns all values at the given location string.
func (m *message) GetAll(location string) ([]string, error) {
	loc, err := ParseLocation(location)
	if err != nil {
		return nil, err
	}
	return m.GetAllAt(loc)
}

// Set sets the value at the given location string.
func (m *message) Set(location string, value string) error {
	loc, err := ParseLocation(location)
	if err != nil {
		return err
	}
	return m.SetAt(loc, value)
}

// GetAt returns the value at the given Location struct.
func (m *message) GetAt(loc *Location) (string, error) {
	if loc == nil {
		return "", fmt.Errorf("%w: nil location", ErrInvalidLocation)
	}
	if !loc.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidLocation, loc.String())
	}

	// Find the segment
	segs := m.Segments(loc.Segment)
	if len(segs) == 0 {
		return "", fmt.Errorf("%w: %s", ErrSegmentNotFound, loc.Segment)
	}

	// Determine which segment to use
	segIndex := 0
	if loc.HasSegmentIndex() {
		segIndex = loc.SegmentIndex
	}
	if segIndex >= len(segs) {
		return "", fmt.Errorf("%w: segment %s[%d]", ErrSegmentNotFound, loc.Segment, segIndex)
	}
	seg := segs[segIndex]

	// If only segment is specified, return empty (segment exists)
	if !loc.HasField() {
		return "", nil
	}

	// Get the field
	field, ok := seg.Field(loc.Field)
	if !ok {
		return "", fmt.Errorf("%w: %s.%d", ErrFieldNotFound, loc.Segment, loc.Field)
	}

	// Get the repetition
	repIndex := 0
	if loc.HasRepetition() {
		repIndex = loc.Repetition
	}
	rep, ok := field.Repetition(repIndex)
	if !ok {
		// No repetition at this index, return empty
		return "", nil
	}

	// If no component specified, return the repetition value
	if !loc.HasComponent() {
		return rep.Value(), nil
	}

	// Get the component
	comp, ok := rep.Component(loc.Component)
	if !ok {
		return "", nil
	}

	// If no subcomponent specified, return the component value
	if !loc.HasSubComponent() {
		return comp.Value(), nil
	}

	// Get the subcomponent
	subcomp, ok := comp.SubComponent(loc.SubComponent)
	if !ok {
		return "", nil
	}

	return subcomp.Value(), nil
}

// GetAllAt returns all values at the given Location struct.
func (m *message) GetAllAt(loc *Location) ([]string, error) {
	if loc == nil {
		return nil, fmt.Errorf("%w: nil location", ErrInvalidLocation)
	}
	if !loc.IsValid() {
		return nil, fmt.Errorf("%w: %s", ErrInvalidLocation, loc.String())
	}

	// Find matching segments
	segs := m.Segments(loc.Segment)
	if len(segs) == 0 {
		return []string{}, nil
	}

	var results []string

	// Determine which segments to process
	var targetSegs []Segment
	if loc.HasSegmentIndex() {
		if loc.SegmentIndex < len(segs) {
			targetSegs = []Segment{segs[loc.SegmentIndex]}
		}
	} else {
		targetSegs = segs
	}

	for _, seg := range targetSegs {
		// If no field specified, add empty string for each segment
		if !loc.HasField() {
			results = append(results, "")
			continue
		}

		field, ok := seg.Field(loc.Field)
		if !ok {
			continue
		}

		// Determine which repetitions to process
		var reps []Repetition
		if loc.HasRepetition() {
			rep, ok := field.Repetition(loc.Repetition)
			if ok {
				reps = []Repetition{rep}
			}
		} else {
			reps = field.Repetitions()
		}

		for _, rep := range reps {
			// If no component specified, add repetition value
			if !loc.HasComponent() {
				results = append(results, rep.Value())
				continue
			}

			comp, ok := rep.Component(loc.Component)
			if !ok {
				results = append(results, "")
				continue
			}

			// If no subcomponent specified, add component value
			if !loc.HasSubComponent() {
				results = append(results, comp.Value())
				continue
			}

			subcomp, ok := comp.SubComponent(loc.SubComponent)
			if !ok {
				results = append(results, "")
				continue
			}

			results = append(results, subcomp.Value())
		}
	}

	if results == nil {
		return []string{}, nil
	}
	return results, nil
}

// SetAt sets the value at the given Location struct.
func (m *message) SetAt(loc *Location, value string) error {
	if loc == nil {
		return fmt.Errorf("%w: nil location", ErrInvalidLocation)
	}
	if !loc.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidLocation, loc.String())
	}
	if !loc.HasField() {
		return fmt.Errorf("%w: field is required for Set operation", ErrInvalidLocation)
	}

	// Find or create the segment
	segs := m.Segments(loc.Segment)
	segIndex := 0
	if loc.HasSegmentIndex() {
		segIndex = loc.SegmentIndex
	}

	var seg Segment
	if segIndex < len(segs) {
		seg = segs[segIndex]
	} else {
		// Need to create segments up to the requested index
		// For now, we only support setting on existing segments
		return fmt.Errorf("%w: %s[%d]", ErrSegmentNotFound, loc.Segment, segIndex)
	}

	// Build location string for the segment
	var segLoc strings.Builder
	segLoc.WriteString(fmt.Sprintf("%d", loc.Field))
	if loc.HasRepetition() {
		segLoc.WriteString(fmt.Sprintf("[%d]", loc.Repetition))
	}
	if loc.HasComponent() {
		segLoc.WriteString(fmt.Sprintf(".%d", loc.Component))
		if loc.HasSubComponent() {
			segLoc.WriteString(fmt.Sprintf(".%d", loc.SubComponent))
		}
	}

	return seg.Set(segLoc.String(), value)
}

// AddSegment appends a segment to the message.
func (m *message) AddSegment(seg Segment) error {
	if seg == nil {
		return ErrNilSegment
	}
	m.segments = append(m.segments, seg)
	return nil
}

// InsertSegment inserts a segment at the given index.
func (m *message) InsertSegment(index int, seg Segment) error {
	if seg == nil {
		return ErrNilSegment
	}
	if index < 0 || index > len(m.segments) {
		return fmt.Errorf("%w: %d", ErrIndexOutOfRange, index)
	}

	// Insert at index
	m.segments = append(m.segments, nil)
	copy(m.segments[index+1:], m.segments[index:])
	m.segments[index] = seg
	return nil
}

// RemoveSegment removes the first segment with the given name.
func (m *message) RemoveSegment(name string) bool {
	name = strings.ToUpper(name)
	for i, seg := range m.segments {
		if seg.Name() == name {
			m.segments = append(m.segments[:i], m.segments[i+1:]...)
			return true
		}
	}
	return false
}

// Bytes returns the encoded message as bytes.
func (m *message) Bytes() []byte {
	if len(m.segments) == 0 {
		return []byte{}
	}

	var buf bytes.Buffer
	for i, seg := range m.segments {
		if i > 0 {
			buf.WriteByte(byte(SegmentTerminator))
		}
		buf.Write(seg.Bytes(m.delimiters))
	}
	buf.WriteByte(byte(SegmentTerminator))
	return buf.Bytes()
}

// String returns the string representation of the message.
func (m *message) String() string {
	return string(m.Bytes())
}

// Type returns the message type from MSH.9.
func (m *message) Type() string {
	msh, ok := m.Segment("MSH")
	if !ok {
		return ""
	}
	val, err := msh.Get("9")
	if err != nil {
		return ""
	}
	return val
}

// ControlID returns the message control ID from MSH.10.
func (m *message) ControlID() string {
	msh, ok := m.Segment("MSH")
	if !ok {
		return ""
	}
	val, err := msh.Get("10")
	if err != nil {
		return ""
	}
	return val
}

// Version returns the HL7 version from MSH.12.
func (m *message) Version() string {
	msh, ok := m.Segment("MSH")
	if !ok {
		return ""
	}
	val, err := msh.Get("12")
	if err != nil {
		return ""
	}
	return val
}

// Delimiters returns the message delimiters.
func (m *message) Delimiters() *Delimiters {
	return m.delimiters
}
