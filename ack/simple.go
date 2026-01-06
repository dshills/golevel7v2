package ack

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/dshills/golevel7/hl7"
)

// simpleMessage is a minimal Message implementation for ACK building.
// It provides just enough functionality to construct ACK messages.
type simpleMessage struct {
	segments []hl7.Segment
	delims   *hl7.Delimiters
}

// newSimpleMessage creates a new simpleMessage with the given delimiters.
func newSimpleMessage(delims *hl7.Delimiters) *simpleMessage {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}
	return &simpleMessage{
		segments: make([]hl7.Segment, 0),
		delims:   delims,
	}
}

// Segment returns the first segment with the given name.
func (m *simpleMessage) Segment(name string) (hl7.Segment, bool) {
	for _, seg := range m.segments {
		if seg.Name() == name {
			return seg, true
		}
	}
	return nil, false
}

// Segments returns all segments with the given name.
func (m *simpleMessage) Segments(name string) []hl7.Segment {
	var result []hl7.Segment
	for _, seg := range m.segments {
		if seg.Name() == name {
			result = append(result, seg)
		}
	}
	return result
}

// AllSegments returns all segments in the message.
func (m *simpleMessage) AllSegments() []hl7.Segment {
	result := make([]hl7.Segment, len(m.segments))
	copy(result, m.segments)
	return result
}

// Get retrieves a value at the specified location.
func (m *simpleMessage) Get(location string) (string, error) {
	loc, err := hl7.ParseLocation(location)
	if err != nil {
		return "", err
	}
	return m.GetAt(loc)
}

// GetAll retrieves all values at the specified location.
func (m *simpleMessage) GetAll(location string) ([]string, error) {
	loc, err := hl7.ParseLocation(location)
	if err != nil {
		return nil, err
	}
	return m.GetAllAt(loc)
}

// Set sets a value at the specified location.
func (m *simpleMessage) Set(location string, value string) error {
	loc, err := hl7.ParseLocation(location)
	if err != nil {
		return err
	}
	return m.SetAt(loc, value)
}

// GetAt retrieves a value using a pre-parsed Location.
func (m *simpleMessage) GetAt(loc *hl7.Location) (string, error) {
	seg, ok := m.Segment(loc.Segment)
	if !ok {
		return "", nil
	}

	if loc.Field < 0 {
		return "", nil
	}

	return seg.Get(fmt.Sprintf("%d", loc.Field))
}

// GetAllAt retrieves all values at the specified Location.
func (m *simpleMessage) GetAllAt(loc *hl7.Location) ([]string, error) {
	val, err := m.GetAt(loc)
	if err != nil {
		return nil, err
	}
	if val == "" {
		return nil, nil
	}
	return []string{val}, nil
}

// SetAt sets a value using a pre-parsed Location.
func (m *simpleMessage) SetAt(loc *hl7.Location, value string) error {
	seg, ok := m.Segment(loc.Segment)
	if !ok {
		return fmt.Errorf("segment %s not found", loc.Segment)
	}

	if loc.Field < 0 {
		return nil
	}

	return seg.Set(fmt.Sprintf("%d", loc.Field), value)
}

// AddSegment appends a segment to the message.
func (m *simpleMessage) AddSegment(seg hl7.Segment) error {
	if seg == nil {
		return fmt.Errorf("cannot add nil segment")
	}
	m.segments = append(m.segments, seg)
	return nil
}

// InsertSegment inserts a segment at the specified index.
func (m *simpleMessage) InsertSegment(index int, seg hl7.Segment) error {
	if seg == nil {
		return fmt.Errorf("cannot insert nil segment")
	}
	if index < 0 || index > len(m.segments) {
		return fmt.Errorf("index %d out of bounds", index)
	}
	m.segments = append(m.segments[:index], append([]hl7.Segment{seg}, m.segments[index:]...)...)
	return nil
}

// RemoveSegment removes the first segment with the given name.
func (m *simpleMessage) RemoveSegment(name string) bool {
	for i, seg := range m.segments {
		if seg.Name() == name {
			m.segments = append(m.segments[:i], m.segments[i+1:]...)
			return true
		}
	}
	return false
}

// Bytes returns the message encoded as HL7 format bytes.
func (m *simpleMessage) Bytes() []byte {
	var buf bytes.Buffer
	for i, seg := range m.segments {
		if i > 0 {
			buf.WriteByte(byte(hl7.SegmentTerminator))
		}
		buf.Write(seg.Bytes(m.delims))
	}
	buf.WriteByte(byte(hl7.SegmentTerminator))
	return buf.Bytes()
}

// String returns the message as an HL7 format string.
func (m *simpleMessage) String() string {
	return string(m.Bytes())
}

// Type returns the message type from MSH-9.
func (m *simpleMessage) Type() string {
	msh, ok := m.Segment("MSH")
	if !ok {
		return ""
	}
	val, _ := msh.Get("9")
	return val
}

// ControlID returns the message control ID from MSH-10.
func (m *simpleMessage) ControlID() string {
	msh, ok := m.Segment("MSH")
	if !ok {
		return ""
	}
	val, _ := msh.Get("10")
	return val
}

// Version returns the HL7 version from MSH-12.
func (m *simpleMessage) Version() string {
	msh, ok := m.Segment("MSH")
	if !ok {
		return ""
	}
	val, _ := msh.Get("12")
	return val
}

// Delimiters returns the delimiter configuration for this message.
func (m *simpleMessage) Delimiters() *hl7.Delimiters {
	return m.delims
}

// simpleSegment is a minimal Segment implementation for ACK building.
type simpleSegment struct {
	name   string
	fields map[int]string
	delims *hl7.Delimiters
}

// newSimpleSegment creates a new simpleSegment with the given name.
func newSimpleSegment(name string, delims *hl7.Delimiters) *simpleSegment {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}
	return &simpleSegment{
		name:   name,
		fields: make(map[int]string),
		delims: delims,
	}
}

// Name returns the segment name.
func (s *simpleSegment) Name() string {
	return s.name
}

// Field returns the field at the specified 1-based index.
func (s *simpleSegment) Field(index int) (hl7.Field, bool) {
	val, ok := s.fields[index]
	if !ok {
		return nil, false
	}
	return hl7.NewField(index, val), true
}

// Fields returns all fields at the same sequence number.
func (s *simpleSegment) Fields(seq int) []hl7.Field {
	f, ok := s.Field(seq)
	if !ok {
		return nil
	}
	return []hl7.Field{f}
}

// AllFields returns all fields in the segment.
func (s *simpleSegment) AllFields() []hl7.Field {
	if len(s.fields) == 0 {
		return nil
	}

	// Find max field index
	maxIndex := 0
	for idx := range s.fields {
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	result := make([]hl7.Field, maxIndex)
	for i := 1; i <= maxIndex; i++ {
		if val, ok := s.fields[i]; ok {
			result[i-1] = hl7.NewField(i, val)
		} else {
			result[i-1] = hl7.NewField(i, "")
		}
	}
	return result
}

// FieldCount returns the number of fields.
func (s *simpleSegment) FieldCount() int {
	if len(s.fields) == 0 {
		return 0
	}
	maxIndex := 0
	for idx := range s.fields {
		if idx > maxIndex {
			maxIndex = idx
		}
	}
	return maxIndex
}

// Get retrieves a value at the specified field location.
func (s *simpleSegment) Get(location string) (string, error) {
	// Parse field index from location (e.g., "3" or "3.1" or "3.1.2")
	parts := strings.Split(location, ".")
	if len(parts) == 0 || parts[0] == "" {
		return "", nil
	}

	fieldIdx, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid field index: %s", parts[0])
	}

	val, ok := s.fields[fieldIdx]
	if !ok {
		return "", nil
	}

	// If requesting specific component/subcomponent, parse the value
	if len(parts) > 1 {
		// Parse the field value to extract component
		components := strings.Split(val, string(s.delims.Component))
		compIdx, err := strconv.Atoi(parts[1])
		if err != nil || compIdx < 1 || compIdx > len(components) {
			return "", nil
		}
		val = components[compIdx-1]

		// Check for subcomponent
		if len(parts) > 2 {
			subcomponents := strings.Split(val, string(s.delims.SubComponent))
			subIdx, err := strconv.Atoi(parts[2])
			if err != nil || subIdx < 1 || subIdx > len(subcomponents) {
				return "", nil
			}
			val = subcomponents[subIdx-1]
		}
	}

	return val, nil
}

// GetAll retrieves all values at the specified location.
func (s *simpleSegment) GetAll(location string) ([]string, error) {
	val, err := s.Get(location)
	if err != nil {
		return nil, err
	}
	if val == "" {
		return nil, nil
	}
	return []string{val}, nil
}

// Set sets a value at the specified field location.
func (s *simpleSegment) Set(location string, value string) error {
	// Parse field index from location
	parts := strings.Split(location, ".")
	if len(parts) == 0 || parts[0] == "" {
		return fmt.Errorf("invalid location: %s", location)
	}

	fieldIdx, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid field index: %s", parts[0])
	}

	// For simple case, just set the field value
	// Complex component paths would require more sophisticated handling
	if len(parts) == 1 {
		s.fields[fieldIdx] = value
		return nil
	}

	// Handle component setting
	existing := s.fields[fieldIdx]
	components := strings.Split(existing, string(s.delims.Component))

	compIdx, err := strconv.Atoi(parts[1])
	if err != nil || compIdx < 1 {
		return fmt.Errorf("invalid component index: %s", parts[1])
	}

	// Expand components slice if needed
	for len(components) < compIdx {
		components = append(components, "")
	}

	if len(parts) == 2 {
		components[compIdx-1] = value
	} else {
		// Handle subcomponent setting
		subcomponents := strings.Split(components[compIdx-1], string(s.delims.SubComponent))

		subIdx, err := strconv.Atoi(parts[2])
		if err != nil || subIdx < 1 {
			return fmt.Errorf("invalid subcomponent index: %s", parts[2])
		}

		// Expand subcomponents slice if needed
		for len(subcomponents) < subIdx {
			subcomponents = append(subcomponents, "")
		}
		subcomponents[subIdx-1] = value
		components[compIdx-1] = strings.Join(subcomponents, string(s.delims.SubComponent))
	}

	s.fields[fieldIdx] = strings.Join(components, string(s.delims.Component))
	return nil
}

// SetField sets or replaces the field at the specified index.
func (s *simpleSegment) SetField(index int, field hl7.Field) error {
	if field == nil {
		delete(s.fields, index)
		return nil
	}
	s.fields[index] = field.Value()
	return nil
}

// AddField adds a field to the segment.
func (s *simpleSegment) AddField(field hl7.Field) error {
	if field == nil {
		return nil
	}
	// Find the next available index
	nextIndex := s.FieldCount() + 1
	s.fields[nextIndex] = field.Value()
	return nil
}

// Bytes returns the segment encoded as HL7 format bytes.
func (s *simpleSegment) Bytes(delims *hl7.Delimiters) []byte {
	if delims == nil {
		delims = s.delims
	}

	var buf bytes.Buffer
	buf.WriteString(s.name)

	// Find max field index
	maxIndex := 0
	for idx := range s.fields {
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	// Special handling for MSH segment
	if s.name == "MSH" {
		// MSH-1 is the field separator itself
		buf.WriteByte(byte(delims.Field))
		// MSH-2 is the encoding characters
		buf.WriteString(delims.EncodingCharacters())

		// Fields 3+ follow normally
		for i := 3; i <= maxIndex; i++ {
			buf.WriteByte(byte(delims.Field))
			if val, ok := s.fields[i]; ok {
				buf.WriteString(val)
			}
		}
	} else {
		// Normal segment handling
		for i := 1; i <= maxIndex; i++ {
			buf.WriteByte(byte(delims.Field))
			if val, ok := s.fields[i]; ok {
				buf.WriteString(val)
			}
		}
	}

	return buf.Bytes()
}

// String returns the segment as an HL7 format string.
func (s *simpleSegment) String() string {
	return string(s.Bytes(s.delims))
}

// Delimiters returns the delimiter configuration for this segment.
func (s *simpleSegment) Delimiters() *hl7.Delimiters {
	return s.delims
}
