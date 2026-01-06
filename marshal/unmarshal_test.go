package marshal

import (
	"errors"
	"testing"
	"time"

	"github.com/dshills/golevel7/hl7"
)

// Patient demonstrates a typical struct for unmarshaling patient data
type Patient struct {
	ID        string    `hl7:"PID.3"`
	LastName  string    `hl7:"PID.5.1"`
	FirstName string    `hl7:"PID.5.2"`
	DOB       time.Time `hl7:"PID.7,format=20060102"`
	Gender    string    `hl7:"PID.8"`
}

// mockMessage is a simple test message implementation
type mockMessage struct {
	data map[string][]string
}

func newMockMessage() *mockMessage {
	return &mockMessage{
		data: make(map[string][]string),
	}
}

func (m *mockMessage) SetVal(location, value string) {
	m.data[location] = []string{value}
}

func (m *mockMessage) Set(location, value string) error {
	m.data[location] = []string{value}
	return nil
}

func (m *mockMessage) SetAll(location string, values []string) {
	m.data[location] = values
}

func (m *mockMessage) Segment(_ string) (hl7.Segment, bool)     { return nil, false }
func (m *mockMessage) Segments(_ string) []hl7.Segment          { return nil }
func (m *mockMessage) AllSegments() []hl7.Segment               { return nil }
func (m *mockMessage) AddSegment(_ hl7.Segment) error           { return nil }
func (m *mockMessage) InsertSegment(_ int, _ hl7.Segment) error { return nil }
func (m *mockMessage) RemoveSegment(_ string) bool              { return false }
func (m *mockMessage) Bytes() []byte                            { return nil }
func (m *mockMessage) String() string                           { return "" }
func (m *mockMessage) Type() string                             { return "" }
func (m *mockMessage) ControlID() string                        { return "" }
func (m *mockMessage) Version() string                          { return "" }
func (m *mockMessage) Delimiters() *hl7.Delimiters              { return hl7.DefaultDelimiters() }

func (m *mockMessage) Get(location string) (string, error) {
	if vals, ok := m.data[location]; ok && len(vals) > 0 {
		return vals[0], nil
	}
	return "", nil
}

func (m *mockMessage) GetAll(location string) ([]string, error) {
	if vals, ok := m.data[location]; ok {
		return vals, nil
	}
	return nil, nil
}

func (m *mockMessage) GetAt(loc *hl7.Location) (string, error) {
	return m.Get(loc.String())
}

func (m *mockMessage) GetAllAt(loc *hl7.Location) ([]string, error) {
	return m.GetAll(loc.String())
}

func (m *mockMessage) SetAt(loc *hl7.Location, value string) error {
	m.data[loc.String()] = []string{value}
	return nil
}

func TestNewUnmarshaler(t *testing.T) {
	u := NewUnmarshaler()
	if u == nil {
		t.Fatal("NewUnmarshaler() returned nil")
	}
}

func TestUnmarshaler_Errors(t *testing.T) {
	u := NewUnmarshaler()

	tests := []struct {
		name    string
		msg     hl7.Message
		target  interface{}
		wantErr error
	}{
		{
			name:    "nil message",
			msg:     nil,
			target:  &Patient{},
			wantErr: ErrNilMessage,
		},
		{
			name:    "not a pointer",
			msg:     newMockMessage(),
			target:  Patient{},
			wantErr: ErrNotPointer,
		},
		{
			name:    "nil pointer",
			msg:     newMockMessage(),
			target:  (*Patient)(nil),
			wantErr: ErrNilPointer,
		},
		{
			name:    "not a struct",
			msg:     newMockMessage(),
			target:  new(string),
			wantErr: ErrNotStruct,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.msg, tt.target)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnmarshaler_StringFields(t *testing.T) {
	type Simple struct {
		ID   string `hl7:"PID.3"`
		Name string `hl7:"PID.5"`
	}

	msg := newMockMessage()
	_ = msg.Set("PID.3", "12345")
	_ = msg.Set("PID.5", "Smith^John")

	u := NewUnmarshaler()
	var s Simple
	err := u.Unmarshal(msg, &s)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if s.ID != "12345" {
		t.Errorf("ID = %q, want %q", s.ID, "12345")
	}
	if s.Name != "Smith^John" {
		t.Errorf("Name = %q, want %q", s.Name, "Smith^John")
	}
}

func TestUnmarshaler_ComponentAccess(t *testing.T) {
	type Name struct {
		LastName  string `hl7:"PID.5.1"`
		FirstName string `hl7:"PID.5.2"`
		Middle    string `hl7:"PID.5.3"`
	}

	msg := newMockMessage()
	_ = msg.Set("PID.5.1", "Smith")
	_ = msg.Set("PID.5.2", "John")
	_ = msg.Set("PID.5.3", "Q")

	u := NewUnmarshaler()
	var n Name
	err := u.Unmarshal(msg, &n)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if n.LastName != "Smith" {
		t.Errorf("LastName = %q, want %q", n.LastName, "Smith")
	}
	if n.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", n.FirstName, "John")
	}
	if n.Middle != "Q" {
		t.Errorf("Middle = %q, want %q", n.Middle, "Q")
	}
}

func TestUnmarshaler_IntFields(t *testing.T) {
	type Numbers struct {
		Count  int   `hl7:"OBX.1"`
		Total  int64 `hl7:"OBX.2"`
		Amount int32 `hl7:"OBX.3"`
	}

	msg := newMockMessage()
	_ = msg.Set("OBX.1", "42")
	_ = msg.Set("OBX.2", "1234567890")
	_ = msg.Set("OBX.3", "100")

	u := NewUnmarshaler()
	var n Numbers
	err := u.Unmarshal(msg, &n)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if n.Count != 42 {
		t.Errorf("Count = %d, want %d", n.Count, 42)
	}
	if n.Total != 1234567890 {
		t.Errorf("Total = %d, want %d", n.Total, 1234567890)
	}
	if n.Amount != 100 {
		t.Errorf("Amount = %d, want %d", n.Amount, 100)
	}
}

func TestUnmarshaler_FloatFields(t *testing.T) {
	type Values struct {
		Price    float64 `hl7:"OBX.5"`
		Quantity float32 `hl7:"OBX.6"`
	}

	msg := newMockMessage()
	_ = msg.Set("OBX.5", "99.99")
	_ = msg.Set("OBX.6", "3.14")

	u := NewUnmarshaler()
	var v Values
	err := u.Unmarshal(msg, &v)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if v.Price != 99.99 {
		t.Errorf("Price = %f, want %f", v.Price, 99.99)
	}
	// Float32 comparison with tolerance
	if v.Quantity < 3.13 || v.Quantity > 3.15 {
		t.Errorf("Quantity = %f, want ~3.14", v.Quantity)
	}
}

func TestUnmarshaler_BoolFields(t *testing.T) {
	type Flags struct {
		Active    bool `hl7:"PID.30"`
		Verified  bool `hl7:"PID.31"`
		Confirmed bool `hl7:"PID.32"`
		Accepted  bool `hl7:"PID.33"`
	}

	msg := newMockMessage()
	_ = msg.Set("PID.30", "true")
	_ = msg.Set("PID.31", "Y")
	_ = msg.Set("PID.32", "1")
	_ = msg.Set("PID.33", "yes")

	u := NewUnmarshaler()
	var f Flags
	err := u.Unmarshal(msg, &f)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !f.Active {
		t.Error("Active should be true")
	}
	if !f.Verified {
		t.Error("Verified should be true")
	}
	if !f.Confirmed {
		t.Error("Confirmed should be true")
	}
	if !f.Accepted {
		t.Error("Accepted should be true")
	}
}

func TestUnmarshaler_BoolFalseValues(t *testing.T) {
	type Flags struct {
		A bool `hl7:"PID.30"`
		B bool `hl7:"PID.31"`
		C bool `hl7:"PID.32"`
		D bool `hl7:"PID.33"`
	}

	msg := newMockMessage()
	_ = msg.Set("PID.30", "false")
	_ = msg.Set("PID.31", "N")
	_ = msg.Set("PID.32", "0")
	_ = msg.Set("PID.33", "no")

	u := NewUnmarshaler()
	var f Flags
	err := u.Unmarshal(msg, &f)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if f.A {
		t.Error("A should be false")
	}
	if f.B {
		t.Error("B should be false")
	}
	if f.C {
		t.Error("C should be false")
	}
	if f.D {
		t.Error("D should be false")
	}
}

func TestUnmarshaler_TimeField(t *testing.T) {
	type Appointment struct {
		DateTime time.Time `hl7:"SCH.11,format=20060102150405"`
		DateOnly time.Time `hl7:"SCH.12,format=20060102"`
	}

	msg := newMockMessage()
	_ = msg.Set("SCH.11", "20231215143022")
	_ = msg.Set("SCH.12", "20231225")

	u := NewUnmarshaler()
	var a Appointment
	err := u.Unmarshal(msg, &a)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	expectedDT := time.Date(2023, 12, 15, 14, 30, 22, 0, time.UTC)
	if !a.DateTime.Equal(expectedDT) {
		t.Errorf("DateTime = %v, want %v", a.DateTime, expectedDT)
	}

	expectedDate := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
	if !a.DateOnly.Equal(expectedDate) {
		t.Errorf("DateOnly = %v, want %v", a.DateOnly, expectedDate)
	}
}

func TestUnmarshaler_SliceField(t *testing.T) {
	type Identifiers struct {
		IDs []string `hl7:"PID.3"`
	}

	msg := newMockMessage()
	msg.SetAll("PID.3", []string{"ID1", "ID2", "ID3"})

	u := NewUnmarshaler()
	var ids Identifiers
	err := u.Unmarshal(msg, &ids)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(ids.IDs) != 3 {
		t.Fatalf("len(IDs) = %d, want 3", len(ids.IDs))
	}
	if ids.IDs[0] != "ID1" {
		t.Errorf("IDs[0] = %q, want %q", ids.IDs[0], "ID1")
	}
	if ids.IDs[1] != "ID2" {
		t.Errorf("IDs[1] = %q, want %q", ids.IDs[1], "ID2")
	}
	if ids.IDs[2] != "ID3" {
		t.Errorf("IDs[2] = %q, want %q", ids.IDs[2], "ID3")
	}
}

func TestUnmarshaler_PointerField(t *testing.T) {
	type Optional struct {
		Name *string `hl7:"PID.5"`
		Age  *int    `hl7:"PID.6"`
	}

	msg := newMockMessage()
	_ = msg.Set("PID.5", "John")
	_ = msg.Set("PID.6", "30")

	u := NewUnmarshaler()
	var o Optional
	err := u.Unmarshal(msg, &o)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if o.Name == nil {
		t.Fatal("Name should not be nil")
	}
	if *o.Name != "John" {
		t.Errorf("*Name = %q, want %q", *o.Name, "John")
	}
	if o.Age == nil {
		t.Fatal("Age should not be nil")
	}
	if *o.Age != 30 {
		t.Errorf("*Age = %d, want %d", *o.Age, 30)
	}
}

func TestUnmarshaler_PointerFieldNil(t *testing.T) {
	type Optional struct {
		Name *string `hl7:"PID.5"`
	}

	msg := newMockMessage()
	// Don't set PID.5

	u := NewUnmarshaler()
	var o Optional
	err := u.Unmarshal(msg, &o)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if o.Name != nil {
		t.Errorf("Name should be nil, got %q", *o.Name)
	}
}

func TestUnmarshaler_IgnoreField(t *testing.T) {
	type WithIgnore struct {
		ID       string `hl7:"PID.3"`
		Ignored  string `hl7:"-"`
		Internal string
	}

	msg := newMockMessage()
	_ = msg.Set("PID.3", "12345")

	u := NewUnmarshaler()
	w := WithIgnore{
		Ignored:  "original",
		Internal: "internal",
	}
	err := u.Unmarshal(msg, &w)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if w.ID != "12345" {
		t.Errorf("ID = %q, want %q", w.ID, "12345")
	}
	if w.Ignored != "original" {
		t.Errorf("Ignored = %q, want %q (should be unchanged)", w.Ignored, "original")
	}
	if w.Internal != "internal" {
		t.Errorf("Internal = %q, want %q (should be unchanged)", w.Internal, "internal")
	}
}

func TestUnmarshaler_MissingFields(t *testing.T) {
	type Simple struct {
		ID   string `hl7:"PID.3"`
		Name string `hl7:"PID.5"`
	}

	msg := newMockMessage()
	// Don't set any values

	u := NewUnmarshaler()
	var s Simple
	err := u.Unmarshal(msg, &s)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Fields should be empty strings (zero values)
	if s.ID != "" {
		t.Errorf("ID = %q, want empty string", s.ID)
	}
	if s.Name != "" {
		t.Errorf("Name = %q, want empty string", s.Name)
	}
}

func TestUnmarshaler_CustomTagName(t *testing.T) {
	type Custom struct {
		ID string `custom:"PID.3"`
	}

	msg := newMockMessage()
	_ = msg.Set("PID.3", "12345")

	u := NewUnmarshaler(WithTagName("custom"))
	var c Custom
	err := u.Unmarshal(msg, &c)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if c.ID != "12345" {
		t.Errorf("ID = %q, want %q", c.ID, "12345")
	}
}

func TestUnmarshaler_PatientStruct(t *testing.T) {
	msg := newMockMessage()
	_ = msg.Set("PID.3", "PAT001")
	_ = msg.Set("PID.5.1", "Smith")
	_ = msg.Set("PID.5.2", "John")
	_ = msg.Set("PID.7", "19850315")
	_ = msg.Set("PID.8", "M")

	u := NewUnmarshaler()
	var p Patient
	err := u.Unmarshal(msg, &p)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if p.ID != "PAT001" {
		t.Errorf("ID = %q, want %q", p.ID, "PAT001")
	}
	if p.LastName != "Smith" {
		t.Errorf("LastName = %q, want %q", p.LastName, "Smith")
	}
	if p.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", p.FirstName, "John")
	}
	expectedDOB := time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC)
	if !p.DOB.Equal(expectedDOB) {
		t.Errorf("DOB = %v, want %v", p.DOB, expectedDOB)
	}
	if p.Gender != "M" {
		t.Errorf("Gender = %q, want %q", p.Gender, "M")
	}
}

func TestUnmarshaler_NestedStruct(t *testing.T) {
	type Name struct {
		Last   string `hl7:"1"`
		First  string `hl7:"2"`
		Middle string `hl7:"3"`
	}
	type PatientWithNested struct {
		ID   string `hl7:"PID.3"`
		Name Name   `hl7:"PID.5"`
	}

	msg := newMockMessage()
	_ = msg.Set("PID.3", "12345")
	_ = msg.Set("PID.5.1", "Smith")
	_ = msg.Set("PID.5.2", "John")
	_ = msg.Set("PID.5.3", "Q")

	u := NewUnmarshaler()
	var p PatientWithNested
	err := u.Unmarshal(msg, &p)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if p.ID != "12345" {
		t.Errorf("ID = %q, want %q", p.ID, "12345")
	}
	if p.Name.Last != "Smith" {
		t.Errorf("Name.Last = %q, want %q", p.Name.Last, "Smith")
	}
	if p.Name.First != "John" {
		t.Errorf("Name.First = %q, want %q", p.Name.First, "John")
	}
	if p.Name.Middle != "Q" {
		t.Errorf("Name.Middle = %q, want %q", p.Name.Middle, "Q")
	}
}

func TestStartsWithSegment(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"PID", true},
		{"PID.5", true},
		{"PID.5.1", true},
		{"PID[0].5", true},
		{"MSH", true},
		{"OBX", true},
		{"1", false},
		{"1.2", false},
		{".1.2", false},
		{"PI", false},
		{"pid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := startsWithSegment(tt.input); got != tt.want {
				t.Errorf("startsWithSegment(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
