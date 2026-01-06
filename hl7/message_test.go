package hl7

import (
	"errors"
	"strings"
	"testing"
)

// mockSegment is a test double for Segment interface.
type mockSegment struct {
	name   string
	fields map[int]Field
	values map[string]string
}

func newMockSegment(name string) *mockSegment {
	return &mockSegment{
		name:   strings.ToUpper(name),
		fields: make(map[int]Field),
		values: make(map[string]string),
	}
}

func (s *mockSegment) Name() string { return s.name }

func (s *mockSegment) Field(seq int) (Field, bool) {
	f, ok := s.fields[seq]
	return f, ok
}

func (s *mockSegment) Fields(seq int) []Field {
	f, ok := s.fields[seq]
	if !ok {
		return []Field{}
	}
	return []Field{f}
}

func (s *mockSegment) AllFields() []Field {
	result := make([]Field, 0, len(s.fields))
	for _, f := range s.fields {
		result = append(result, f)
	}
	return result
}

func (s *mockSegment) FieldCount() int { return len(s.fields) }

func (s *mockSegment) Get(location string) (string, error) {
	v, ok := s.values[location]
	if !ok {
		return "", nil
	}
	return v, nil
}

func (s *mockSegment) GetAll(location string) ([]string, error) {
	v, ok := s.values[location]
	if !ok {
		return []string{}, nil
	}
	return []string{v}, nil
}

func (s *mockSegment) Set(location string, value string) error {
	s.values[location] = value
	return nil
}

func (s *mockSegment) SetField(seq int, field Field) error {
	s.fields[seq] = field
	return nil
}

func (s *mockSegment) AddField(field Field) error {
	seq := len(s.fields) + 1
	s.fields[seq] = field
	return nil
}

func (s *mockSegment) Bytes(delims *Delimiters) []byte {
	if delims == nil {
		delims = DefaultDelimiters()
	}
	// Simple encoding: NAME|field1|field2|...
	var parts []string
	parts = append(parts, s.name)
	for i := 1; i <= len(s.fields); i++ {
		if f, ok := s.fields[i]; ok {
			parts = append(parts, f.Value())
		} else {
			parts = append(parts, "")
		}
	}
	return []byte(strings.Join(parts, string(delims.Field)))
}

func (s *mockSegment) String() string {
	return string(s.Bytes(DefaultDelimiters()))
}

// mockField is a test double for Field interface.
type mockField struct {
	seqNum      int
	value       string
	repetitions []Repetition
}

func newMockField(seqNum int, value string) *mockField {
	return &mockField{
		seqNum:      seqNum,
		value:       value,
		repetitions: []Repetition{NewRepetition(value)},
	}
}

func (f *mockField) SeqNum() int   { return f.seqNum }
func (f *mockField) Value() string { return f.value }

func (f *mockField) Component(index int) (Component, bool) {
	if len(f.repetitions) == 0 {
		return nil, false
	}
	return f.repetitions[0].Component(index)
}

func (f *mockField) Components() []Component {
	if len(f.repetitions) == 0 {
		return []Component{}
	}
	return f.repetitions[0].Components()
}

func (f *mockField) Repetition(index int) (Repetition, bool) {
	if index < 0 || index >= len(f.repetitions) {
		return nil, false
	}
	return f.repetitions[index], true
}

func (f *mockField) Repetitions() []Repetition {
	return f.repetitions
}

func (f *mockField) RepetitionCount() int {
	return len(f.repetitions)
}

func (f *mockField) Get(_ string) (string, error) {
	return f.value, nil
}

func (f *mockField) Set(_ string, value string) error {
	f.value = value
	return nil
}

func (f *mockField) Bytes(_ *Delimiters) []byte {
	return []byte(f.value)
}

func (f *mockField) String() string {
	return f.value
}

// TestNewMessage tests the NewMessage constructor.
func TestNewMessage(t *testing.T) {
	msg := NewMessage(nil, nil)

	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}

	if segs := msg.AllSegments(); len(segs) != 0 {
		t.Errorf("NewMessage should have 0 segments, got %d", len(segs))
	}

	if delims := msg.Delimiters(); delims == nil {
		t.Error("NewMessage should have non-nil delimiters")
	} else {
		defaultDelims := DefaultDelimiters()
		if !delims.Equal(defaultDelims) {
			t.Error("NewMessage should have default delimiters")
		}
	}
}

// TestNewEmptyMessage tests the NewEmptyMessage constructor.
func TestNewEmptyMessage(t *testing.T) {
	msg := NewEmptyMessage()

	if msg == nil {
		t.Fatal("NewEmptyMessage returned nil")
	}

	if segs := msg.AllSegments(); len(segs) != 0 {
		t.Errorf("NewEmptyMessage should have 0 segments, got %d", len(segs))
	}

	if delims := msg.Delimiters(); delims == nil {
		t.Error("NewEmptyMessage should have non-nil delimiters")
	} else {
		defaultDelims := DefaultDelimiters()
		if !delims.Equal(defaultDelims) {
			t.Error("NewEmptyMessage should have default delimiters")
		}
	}
}

// TestNewMessage_WithSegments tests creating a message with initial segments.
func TestNewMessage_WithSegments(t *testing.T) {
	msh := newMockSegment("MSH")
	pid := newMockSegment("PID")

	msg := NewMessage([]Segment{msh, pid}, nil)

	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}

	segs := msg.AllSegments()
	if len(segs) != 2 {
		t.Errorf("NewMessage with segments should have 2 segments, got %d", len(segs))
	}

	if segs[0].Name() != "MSH" {
		t.Errorf("First segment should be MSH, got %s", segs[0].Name())
	}
	if segs[1].Name() != "PID" {
		t.Errorf("Second segment should be PID, got %s", segs[1].Name())
	}
}

// TestNewMessageWithDelimiters tests the NewMessageWithDelimiters constructor.
func TestNewMessageWithDelimiters(t *testing.T) {
	tests := []struct {
		name   string
		delims *Delimiters
		want   *Delimiters
	}{
		{
			name:   "nil delimiters defaults to standard",
			delims: nil,
			want:   DefaultDelimiters(),
		},
		{
			name: "custom delimiters",
			delims: &Delimiters{
				Field:        '#',
				Component:    '@',
				Repetition:   '*',
				Escape:       '!',
				SubComponent: '%',
				Truncation:   '+',
			},
			want: &Delimiters{
				Field:        '#',
				Component:    '@',
				Repetition:   '*',
				Escape:       '!',
				SubComponent: '%',
				Truncation:   '+',
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessageWithDelimiters(tt.delims)

			if msg == nil {
				t.Fatal("NewMessageWithDelimiters returned nil")
			}

			got := msg.Delimiters()
			if !got.Equal(tt.want) {
				t.Errorf("Delimiters = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestMessage_AddSegment tests adding segments to a message.
func TestMessage_AddSegment(t *testing.T) {
	msg := NewEmptyMessage()

	// Add MSH segment
	msh := newMockSegment("MSH")
	if err := msg.AddSegment(msh); err != nil {
		t.Fatalf("AddSegment(MSH) error = %v", err)
	}

	// Add PID segment
	pid := newMockSegment("PID")
	if err := msg.AddSegment(pid); err != nil {
		t.Fatalf("AddSegment(PID) error = %v", err)
	}

	// Verify segments
	segs := msg.AllSegments()
	if len(segs) != 2 {
		t.Errorf("AllSegments() length = %d, want 2", len(segs))
	}

	if segs[0].Name() != "MSH" {
		t.Errorf("First segment name = %q, want %q", segs[0].Name(), "MSH")
	}
	if segs[1].Name() != "PID" {
		t.Errorf("Second segment name = %q, want %q", segs[1].Name(), "PID")
	}
}

// TestMessage_AddSegment_Nil tests adding nil segment.
func TestMessage_AddSegment_Nil(t *testing.T) {
	msg := NewEmptyMessage()

	err := msg.AddSegment(nil)
	if err == nil {
		t.Error("AddSegment(nil) should return error")
	}
	if !errors.Is(err, ErrNilSegment) {
		t.Errorf("AddSegment(nil) error = %v, want %v", err, ErrNilSegment)
	}
}

// TestMessage_InsertSegment tests inserting segments at specific positions.
func TestMessage_InsertSegment(t *testing.T) {
	tests := []struct {
		name      string
		initial   []string
		insertAt  int
		insertSeg string
		wantOrder []string
		wantErr   bool
	}{
		{
			name:      "insert at beginning",
			initial:   []string{"MSH", "PID"},
			insertAt:  0,
			insertSeg: "EVN",
			wantOrder: []string{"EVN", "MSH", "PID"},
			wantErr:   false,
		},
		{
			name:      "insert in middle",
			initial:   []string{"MSH", "PID"},
			insertAt:  1,
			insertSeg: "EVN",
			wantOrder: []string{"MSH", "EVN", "PID"},
			wantErr:   false,
		},
		{
			name:      "insert at end",
			initial:   []string{"MSH", "PID"},
			insertAt:  2,
			insertSeg: "EVN",
			wantOrder: []string{"MSH", "PID", "EVN"},
			wantErr:   false,
		},
		{
			name:      "negative index",
			initial:   []string{"MSH"},
			insertAt:  -1,
			insertSeg: "PID",
			wantErr:   true,
		},
		{
			name:      "index too large",
			initial:   []string{"MSH"},
			insertAt:  5,
			insertSeg: "PID",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewEmptyMessage()

			// Add initial segments
			for _, name := range tt.initial {
				_ = msg.AddSegment(newMockSegment(name))
			}

			// Insert new segment
			err := msg.InsertSegment(tt.insertAt, newMockSegment(tt.insertSeg))

			if tt.wantErr {
				if err == nil {
					t.Error("InsertSegment should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("InsertSegment error = %v", err)
			}

			// Verify order
			segs := msg.AllSegments()
			if len(segs) != len(tt.wantOrder) {
				t.Fatalf("AllSegments() length = %d, want %d", len(segs), len(tt.wantOrder))
			}

			for i, wantName := range tt.wantOrder {
				if segs[i].Name() != wantName {
					t.Errorf("Segment[%d].Name() = %q, want %q", i, segs[i].Name(), wantName)
				}
			}
		})
	}
}

// TestMessage_InsertSegment_Nil tests inserting nil segment.
func TestMessage_InsertSegment_Nil(t *testing.T) {
	msg := NewEmptyMessage()

	err := msg.InsertSegment(0, nil)
	if err == nil {
		t.Error("InsertSegment(nil) should return error")
	}
	if !errors.Is(err, ErrNilSegment) {
		t.Errorf("InsertSegment(nil) error = %v, want %v", err, ErrNilSegment)
	}
}

// TestMessage_RemoveSegment tests removing segments by name.
func TestMessage_RemoveSegment(t *testing.T) {
	tests := []struct {
		name       string
		initial    []string
		removeName string
		wantRemove bool
		wantOrder  []string
	}{
		{
			name:       "remove existing segment",
			initial:    []string{"MSH", "EVN", "PID"},
			removeName: "EVN",
			wantRemove: true,
			wantOrder:  []string{"MSH", "PID"},
		},
		{
			name:       "remove first of duplicates",
			initial:    []string{"MSH", "OBX", "OBX", "OBX"},
			removeName: "OBX",
			wantRemove: true,
			wantOrder:  []string{"MSH", "OBX", "OBX"},
		},
		{
			name:       "remove non-existent segment",
			initial:    []string{"MSH", "PID"},
			removeName: "ZZZ",
			wantRemove: false,
			wantOrder:  []string{"MSH", "PID"},
		},
		{
			name:       "case insensitive remove",
			initial:    []string{"MSH", "PID"},
			removeName: "pid",
			wantRemove: true,
			wantOrder:  []string{"MSH"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewEmptyMessage()

			for _, name := range tt.initial {
				_ = msg.AddSegment(newMockSegment(name))
			}

			removed := msg.RemoveSegment(tt.removeName)

			if removed != tt.wantRemove {
				t.Errorf("RemoveSegment(%q) = %v, want %v", tt.removeName, removed, tt.wantRemove)
			}

			segs := msg.AllSegments()
			if len(segs) != len(tt.wantOrder) {
				t.Fatalf("AllSegments() length = %d, want %d", len(segs), len(tt.wantOrder))
			}

			for i, wantName := range tt.wantOrder {
				if segs[i].Name() != wantName {
					t.Errorf("Segment[%d].Name() = %q, want %q", i, segs[i].Name(), wantName)
				}
			}
		})
	}
}

// TestMessage_Segment tests getting a single segment by name.
func TestMessage_Segment(t *testing.T) {
	msg := NewEmptyMessage()
	_ = msg.AddSegment(newMockSegment("MSH"))
	_ = msg.AddSegment(newMockSegment("PID"))
	_ = msg.AddSegment(newMockSegment("OBX"))
	_ = msg.AddSegment(newMockSegment("OBX"))

	tests := []struct {
		name     string
		segName  string
		wantOK   bool
		wantName string
	}{
		{
			name:     "existing segment",
			segName:  "PID",
			wantOK:   true,
			wantName: "PID",
		},
		{
			name:     "first of multiple",
			segName:  "OBX",
			wantOK:   true,
			wantName: "OBX",
		},
		{
			name:    "non-existent segment",
			segName: "ZZZ",
			wantOK:  false,
		},
		{
			name:     "case insensitive",
			segName:  "pid",
			wantOK:   true,
			wantName: "PID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, ok := msg.Segment(tt.segName)

			if ok != tt.wantOK {
				t.Errorf("Segment(%q) ok = %v, want %v", tt.segName, ok, tt.wantOK)
			}

			if tt.wantOK {
				if seg == nil {
					t.Fatal("Segment should not be nil when ok is true")
				}
				if seg.Name() != tt.wantName {
					t.Errorf("Segment(%q).Name() = %q, want %q", tt.segName, seg.Name(), tt.wantName)
				}
			}
		})
	}
}

// TestMessage_Segments tests getting all segments by name.
func TestMessage_Segments(t *testing.T) {
	msg := NewEmptyMessage()
	_ = msg.AddSegment(newMockSegment("MSH"))
	_ = msg.AddSegment(newMockSegment("OBX"))
	_ = msg.AddSegment(newMockSegment("PID"))
	_ = msg.AddSegment(newMockSegment("OBX"))
	_ = msg.AddSegment(newMockSegment("OBX"))

	tests := []struct {
		name    string
		segName string
		want    int
	}{
		{
			name:    "single segment",
			segName: "MSH",
			want:    1,
		},
		{
			name:    "multiple segments",
			segName: "OBX",
			want:    3,
		},
		{
			name:    "non-existent segment",
			segName: "ZZZ",
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segs := msg.Segments(tt.segName)

			if len(segs) != tt.want {
				t.Errorf("Segments(%q) length = %d, want %d", tt.segName, len(segs), tt.want)
			}
		})
	}
}

// TestMessage_Bytes tests message encoding.
func TestMessage_Bytes(t *testing.T) {
	msg := NewEmptyMessage()

	// Empty message
	if bytes := msg.Bytes(); len(bytes) != 0 {
		t.Errorf("Empty message Bytes() = %q, want empty", bytes)
	}

	// Add segments
	msh := newMockSegment("MSH")
	_ = msh.SetField(9, newMockField(9, "ADT^A01"))
	_ = msg.AddSegment(msh)

	pid := newMockSegment("PID")
	_ = pid.SetField(5, newMockField(5, "DOE^JOHN"))
	_ = msg.AddSegment(pid)

	bytes := msg.Bytes()

	// Should contain segment terminator between segments and at end
	bytesStr := string(bytes)
	if !strings.Contains(bytesStr, "MSH") {
		t.Error("Bytes() should contain MSH")
	}
	if !strings.Contains(bytesStr, "PID") {
		t.Error("Bytes() should contain PID")
	}
	if !strings.HasSuffix(bytesStr, "\r") {
		t.Error("Bytes() should end with segment terminator")
	}
}

// TestMessage_String tests string representation.
func TestMessage_String(t *testing.T) {
	msg := NewEmptyMessage()
	_ = msg.AddSegment(newMockSegment("MSH"))

	str := msg.String()
	if str == "" {
		t.Error("String() should not be empty for message with segments")
	}
	if !strings.Contains(str, "MSH") {
		t.Error("String() should contain MSH")
	}
}

// TestMessage_Type tests Type() extraction from MSH.9.
func TestMessage_Type(t *testing.T) {
	tests := []struct {
		name     string
		mshValue string
		want     string
	}{
		{
			name:     "ADT message",
			mshValue: "ADT^A01",
			want:     "ADT^A01",
		},
		{
			name:     "ORU message",
			mshValue: "ORU^R01",
			want:     "ORU^R01",
		},
		{
			name:     "empty type",
			mshValue: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewEmptyMessage()
			msh := newMockSegment("MSH")
			msh.values["9"] = tt.mshValue
			_ = msg.AddSegment(msh)

			got := msg.Type()
			if got != tt.want {
				t.Errorf("Type() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestMessage_Type_NoMSH tests Type() when MSH is missing.
func TestMessage_Type_NoMSH(t *testing.T) {
	msg := NewEmptyMessage()
	_ = msg.AddSegment(newMockSegment("PID"))

	if got := msg.Type(); got != "" {
		t.Errorf("Type() without MSH = %q, want empty", got)
	}
}

// TestMessage_ControlID tests ControlID() extraction from MSH.10.
func TestMessage_ControlID(t *testing.T) {
	msg := NewEmptyMessage()
	msh := newMockSegment("MSH")
	msh.values["10"] = "MSG00001"
	_ = msg.AddSegment(msh)

	got := msg.ControlID()
	if got != "MSG00001" {
		t.Errorf("ControlID() = %q, want %q", got, "MSG00001")
	}
}

// TestMessage_Version tests Version() extraction from MSH.12.
func TestMessage_Version(t *testing.T) {
	msg := NewEmptyMessage()
	msh := newMockSegment("MSH")
	msh.values["12"] = "2.5.1"
	_ = msg.AddSegment(msh)

	got := msg.Version()
	if got != "2.5.1" {
		t.Errorf("Version() = %q, want %q", got, "2.5.1")
	}
}

// TestMessage_GetAt tests GetAt with Location struct.
func TestMessage_GetAt(t *testing.T) {
	msg := NewEmptyMessage()

	// Create MSH with fields
	msh := newMockSegment("MSH")
	msh.fields[9] = newMockField(9, "ADT^A01")
	msh.fields[10] = newMockField(10, "MSG00001")
	_ = msg.AddSegment(msh)

	// Create PID with fields
	pid := newMockSegment("PID")
	pid.fields[5] = newMockField(5, "DOE^JOHN")
	_ = msg.AddSegment(pid)

	tests := []struct {
		name    string
		loc     *Location
		want    string
		wantErr bool
	}{
		{
			name:    "nil location",
			loc:     nil,
			wantErr: true,
		},
		{
			name: "segment only",
			loc: &Location{
				Segment:      "MSH",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "field value",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want:    "DOE^JOHN",
			wantErr: false,
		},
		{
			name: "non-existent segment",
			loc: &Location{
				Segment:      "ZZZ",
				SegmentIndex: -1,
				Field:        1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			wantErr: true,
		},
		{
			name: "non-existent field",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        99,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := msg.GetAt(tt.loc)

			if tt.wantErr {
				if err == nil {
					t.Error("GetAt should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetAt error = %v", err)
			}

			if got != tt.want {
				t.Errorf("GetAt() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestMessage_Get tests Get with location string.
func TestMessage_Get(t *testing.T) {
	msg := NewEmptyMessage()

	msh := newMockSegment("MSH")
	msh.fields[9] = newMockField(9, "ADT^A01")
	_ = msg.AddSegment(msh)

	pid := newMockSegment("PID")
	pid.fields[5] = newMockField(5, "DOE^JOHN")
	_ = msg.AddSegment(pid)

	tests := []struct {
		name     string
		location string
		want     string
		wantErr  bool
	}{
		{
			name:     "valid location",
			location: "PID.5",
			want:     "DOE^JOHN",
			wantErr:  false,
		},
		{
			name:     "MSH field",
			location: "MSH.9",
			want:     "ADT^A01",
			wantErr:  false,
		},
		{
			name:     "invalid location format",
			location: "invalid",
			wantErr:  true,
		},
		{
			name:     "empty location",
			location: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := msg.Get(tt.location)

			if tt.wantErr {
				if err == nil {
					t.Error("Get should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("Get error = %v", err)
			}

			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.location, got, tt.want)
			}
		})
	}
}

// TestMessage_GetAllAt tests GetAllAt with Location struct.
func TestMessage_GetAllAt(t *testing.T) {
	msg := NewEmptyMessage()

	// Add multiple OBX segments
	obx1 := newMockSegment("OBX")
	obx1.fields[5] = newMockField(5, "VALUE1")
	_ = msg.AddSegment(obx1)

	obx2 := newMockSegment("OBX")
	obx2.fields[5] = newMockField(5, "VALUE2")
	_ = msg.AddSegment(obx2)

	tests := []struct {
		name    string
		loc     *Location
		want    []string
		wantErr bool
	}{
		{
			name:    "nil location",
			loc:     nil,
			wantErr: true,
		},
		{
			name: "all OBX.5 values",
			loc: &Location{
				Segment:      "OBX",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want:    []string{"VALUE1", "VALUE2"},
			wantErr: false,
		},
		{
			name: "specific segment index",
			loc: &Location{
				Segment:      "OBX",
				SegmentIndex: 1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want:    []string{"VALUE2"},
			wantErr: false,
		},
		{
			name: "non-existent segment",
			loc: &Location{
				Segment:      "ZZZ",
				SegmentIndex: -1,
				Field:        1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := msg.GetAllAt(tt.loc)

			if tt.wantErr {
				if err == nil {
					t.Error("GetAllAt should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetAllAt error = %v", err)
			}

			if len(got) != len(tt.want) {
				t.Errorf("GetAllAt() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("GetAllAt()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

// TestMessage_GetAll tests GetAll with location string.
func TestMessage_GetAll(t *testing.T) {
	msg := NewEmptyMessage()

	obx1 := newMockSegment("OBX")
	obx1.fields[5] = newMockField(5, "VALUE1")
	_ = msg.AddSegment(obx1)

	obx2 := newMockSegment("OBX")
	obx2.fields[5] = newMockField(5, "VALUE2")
	_ = msg.AddSegment(obx2)

	got, err := msg.GetAll("OBX.5")
	if err != nil {
		t.Fatalf("GetAll error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("GetAll(OBX.5) length = %d, want 2", len(got))
	}
}

// TestMessage_SetAt tests SetAt with Location struct.
func TestMessage_SetAt(t *testing.T) {
	msg := NewEmptyMessage()
	pid := newMockSegment("PID")
	_ = msg.AddSegment(pid)

	tests := []struct {
		name    string
		loc     *Location
		value   string
		wantErr bool
	}{
		{
			name:    "nil location",
			loc:     nil,
			value:   "test",
			wantErr: true,
		},
		{
			name: "location without field",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			value:   "test",
			wantErr: true,
		},
		{
			name: "valid location",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			value:   "DOE^JOHN",
			wantErr: false,
		},
		{
			name: "non-existent segment",
			loc: &Location{
				Segment:      "ZZZ",
				SegmentIndex: 0,
				Field:        1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			value:   "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := msg.SetAt(tt.loc, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("SetAt should return error")
				}
				return
			}

			if err != nil {
				t.Errorf("SetAt error = %v", err)
			}
		})
	}
}

// TestMessage_Set tests Set with location string.
func TestMessage_Set(t *testing.T) {
	msg := NewEmptyMessage()
	pid := newMockSegment("PID")
	_ = msg.AddSegment(pid)

	tests := []struct {
		name     string
		location string
		value    string
		wantErr  bool
	}{
		{
			name:     "valid location",
			location: "PID.5",
			value:    "DOE^JOHN",
			wantErr:  false,
		},
		{
			name:     "invalid location",
			location: "invalid",
			value:    "test",
			wantErr:  true,
		},
		{
			name:     "empty location",
			location: "",
			value:    "test",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := msg.Set(tt.location, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("Set should return error")
				}
				return
			}

			if err != nil {
				t.Errorf("Set error = %v", err)
			}
		})
	}
}

// TestMessage_InterfaceCompliance verifies message implements Message interface.
func TestMessage_InterfaceCompliance(_ *testing.T) {
	var _ Message = (*message)(nil)
	_ = NewEmptyMessage()
	_ = NewMessageWithDelimiters(nil)
}

// TestMessage_Delimiters tests Delimiters() accessor.
func TestMessage_Delimiters(t *testing.T) {
	customDelims := &Delimiters{
		Field:        '#',
		Component:    '@',
		Repetition:   '*',
		Escape:       '!',
		SubComponent: '%',
		Truncation:   '+',
	}

	msg := NewMessageWithDelimiters(customDelims)

	got := msg.Delimiters()
	if got == nil {
		t.Fatal("Delimiters() returned nil")
		return
	}

	if got.Field != '#' {
		t.Errorf("Delimiters().Field = %c, want #", got.Field)
	}
	if got.Component != '@' {
		t.Errorf("Delimiters().Component = %c, want @", got.Component)
	}
}

// TestMessage_EmptySegmentLookup tests segment lookup on empty message.
func TestMessage_EmptySegmentLookup(t *testing.T) {
	msg := NewEmptyMessage()

	// Segment lookup
	seg, ok := msg.Segment("PID")
	if ok {
		t.Error("Segment should return false for empty message")
	}
	if seg != nil {
		t.Error("Segment should return nil for empty message")
	}

	// Segments lookup
	segs := msg.Segments("PID")
	if len(segs) != 0 {
		t.Error("Segments should return empty slice for empty message")
	}
}

// TestMessage_AllSegments_Copy tests that AllSegments returns a copy.
func TestMessage_AllSegments_Copy(t *testing.T) {
	msg := NewEmptyMessage()
	_ = msg.AddSegment(newMockSegment("MSH"))
	_ = msg.AddSegment(newMockSegment("PID"))

	// Get segments
	segs1 := msg.AllSegments()
	segs2 := msg.AllSegments()

	// Modify first slice
	segs1[0] = newMockSegment("XXX")

	// Second slice should be unaffected
	if segs2[0].Name() != "MSH" {
		t.Error("AllSegments should return a copy, not the original slice")
	}
}
