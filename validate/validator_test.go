package validate

import (
	"testing"

	"github.com/dshills/golevel7/hl7"
)

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New() returned nil")
	}

	// With rules
	v2 := New(
		At("MSH.9").Required().Build(),
		At("MSH.10").Required().Build(),
	)
	if v2 == nil {
		t.Fatal("New() with rules returned nil")
	}
}

func TestNewWithRuleSet(t *testing.T) {
	rs := MSHRules()
	v := NewWithRuleSet(rs)
	if v == nil {
		t.Fatal("NewWithRuleSet() returned nil")
	}
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		rules     []Rule
		setup     func(*mockMessage)
		wantValid bool
		wantCount int
	}{
		{
			name: "all rules pass",
			rules: []Rule{
				At("MSH.9").Required().Build(),
				At("MSH.10").Required().Build(),
				At("MSH.12").Required().Build(),
			},
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "ADT^A01")
				m.setField("MSH.10", "12345")
				m.setField("MSH.12", "2.5")
			},
			wantValid: true,
			wantCount: 0,
		},
		{
			name: "one rule fails",
			rules: []Rule{
				At("MSH.9").Required().Build(),
				At("MSH.10").Required().Build(),
			},
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "ADT^A01")
				// MSH.10 missing
			},
			wantValid: false,
			wantCount: 1,
		},
		{
			name: "multiple rules fail",
			rules: []Rule{
				At("MSH.9").Required().Build(),
				At("MSH.10").Required().Build(),
				At("MSH.12").Required().Build(),
			},
			setup:     func(_ *mockMessage) {},
			wantValid: false,
			wantCount: 3,
		},
		{
			name:      "no rules always valid",
			rules:     []Rule{},
			setup:     func(_ *mockMessage) {},
			wantValid: true,
			wantCount: 0,
		},
		{
			name: "nil message",
			rules: []Rule{
				At("MSH.9").Required().Build(),
			},
			setup:     nil, // will test with nil
			wantValid: false,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(tt.rules...)

			var msg hl7.Message
			if tt.setup != nil {
				m := newMockMessage()
				tt.setup(m)
				msg = m
			}

			result := v.Validate(msg)

			if result.Valid() != tt.wantValid {
				t.Errorf("Valid() = %v, want %v", result.Valid(), tt.wantValid)
			}
			if len(result.Errors()) != tt.wantCount {
				t.Errorf("Errors() count = %d, want %d", len(result.Errors()), tt.wantCount)
			}
		})
	}
}

func TestValidationResult_Errors(t *testing.T) {
	v := New(
		At("MSH.9").Required().Build(),
		At("MSH.10").Required().Build(),
	)

	m := newMockMessage()
	result := v.Validate(m)

	errors := result.Errors()
	if len(errors) != 2 {
		t.Errorf("Errors() = %d, want 2", len(errors))
	}

	// Verify the returned slice is a copy
	errors[0] = ValidationError{Message: "modified"}
	errors2 := result.Errors()
	if errors2[0].Message == "modified" {
		t.Error("Errors() should return a copy, not the original slice")
	}
}

func TestValidationResult_Warnings(t *testing.T) {
	result := &validationResult{
		warnings: []ValidationWarning{
			{Location: "PID.5", Message: "Consider adding last name"},
		},
	}

	warnings := result.Warnings()
	if len(warnings) != 1 {
		t.Errorf("Warnings() = %d, want 1", len(warnings))
	}

	// Verify the returned slice is a copy
	warnings[0] = ValidationWarning{Message: "modified"}
	warnings2 := result.Warnings()
	if warnings2[0].Message == "modified" {
		t.Error("Warnings() should return a copy, not the original slice")
	}
}

func TestValidationResult_EmptySlices(t *testing.T) {
	result := &validationResult{}

	// Nil slices should return empty slices
	errors := result.Errors()
	if errors == nil {
		t.Error("Errors() should return empty slice, not nil")
	}

	warnings := result.Warnings()
	if warnings == nil {
		t.Error("Warnings() should return empty slice, not nil")
	}
}

// mockSegment implements hl7.Segment for testing
type mockSegment struct {
	name   string
	fields map[string]string
}

func newMockSegment(name string) *mockSegment {
	return &mockSegment{
		name:   name,
		fields: make(map[string]string),
	}
}

func (s *mockSegment) setField(location, value string) {
	s.fields[location] = value
}

func (s *mockSegment) Name() string { return s.name }

func (s *mockSegment) Field(_ int) (hl7.Field, bool) { return nil, false }
func (s *mockSegment) Fields(_ int) []hl7.Field      { return nil }
func (s *mockSegment) AllFields() []hl7.Field        { return nil }
func (s *mockSegment) FieldCount() int               { return 0 }

func (s *mockSegment) Get(location string) (string, error) {
	if v, ok := s.fields[location]; ok {
		return v, nil
	}
	return "", hl7.ErrFieldNotFound
}

func (s *mockSegment) GetAll(location string) ([]string, error) {
	if v, ok := s.fields[location]; ok {
		return []string{v}, nil
	}
	return nil, hl7.ErrFieldNotFound
}

func (s *mockSegment) Set(location, value string) error {
	s.fields[location] = value
	return nil
}

func (s *mockSegment) SetField(_ int, _ hl7.Field) error { return nil }
func (s *mockSegment) AddField(_ hl7.Field) error        { return nil }
func (s *mockSegment) Bytes(_ *hl7.Delimiters) []byte    { return nil }
func (s *mockSegment) String() string                    { return "" }

var _ hl7.Segment = (*mockSegment)(nil)

func TestValidator_ValidateSegment(t *testing.T) {
	tests := []struct {
		name      string
		rules     []Rule
		segment   *mockSegment
		wantValid bool
		wantCount int
	}{
		{
			name: "applicable rules pass",
			rules: []Rule{
				At("MSH.9").Required().Build(),
				At("MSH.10").Required().Build(),
				At("PID.3").Required().Build(), // Should not apply
			},
			segment: func() *mockSegment {
				s := newMockSegment("MSH")
				s.setField("MSH.9", "ADT^A01")
				s.setField("MSH.10", "12345")
				return s
			}(),
			wantValid: true,
			wantCount: 0,
		},
		{
			name: "applicable rules fail",
			rules: []Rule{
				At("MSH.9").Required().Build(),
				At("MSH.10").Required().Build(),
			},
			segment: func() *mockSegment {
				s := newMockSegment("MSH")
				s.setField("MSH.9", "ADT^A01")
				// MSH.10 missing
				return s
			}(),
			wantValid: false,
			wantCount: 1,
		},
		{
			name: "nil segment",
			rules: []Rule{
				At("MSH.9").Required().Build(),
			},
			segment:   nil,
			wantValid: false,
			wantCount: 1,
		},
		{
			name: "no applicable rules",
			rules: []Rule{
				At("PID.3").Required().Build(),
				At("PV1.2").Required().Build(),
			},
			segment: func() *mockSegment {
				return newMockSegment("MSH")
			}(),
			wantValid: true,
			wantCount: 0,
		},
		{
			name: "rules with segment index",
			rules: []Rule{
				At("OBX[0].2").Required().Build(),
				At("OBX.3").Required().Build(),
			},
			segment: func() *mockSegment {
				s := newMockSegment("OBX")
				s.setField("OBX[0].2", "NM")
				s.setField("OBX.3", "TEST")
				return s
			}(),
			wantValid: true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(tt.rules...)

			var seg hl7.Segment
			if tt.segment != nil {
				seg = tt.segment
			}

			result := v.ValidateSegment(seg)

			if result.Valid() != tt.wantValid {
				t.Errorf("Valid() = %v, want %v", result.Valid(), tt.wantValid)
			}
			if len(result.Errors()) != tt.wantCount {
				t.Errorf("Errors() count = %d, want %d, errors: %v", len(result.Errors()), tt.wantCount, result.Errors())
			}
		})
	}
}

func TestSegmentWrapper(t *testing.T) {
	seg := newMockSegment("PID")
	seg.setField("PID.3", "12345")
	seg.setField("PID.5", "DOE^JOHN")

	wrapper := &segmentWrapper{seg: seg}

	// Test Get
	v, err := wrapper.Get("PID.3")
	if err != nil || v != "12345" {
		t.Errorf("Get() = %q, %v, want %q, nil", v, err, "12345")
	}

	// Test GetAll
	vals, err := wrapper.GetAll("PID.5")
	if err != nil || len(vals) != 1 || vals[0] != "DOE^JOHN" {
		t.Errorf("GetAll() = %v, %v, want [DOE^JOHN], nil", vals, err)
	}

	// Test Segment
	s, ok := wrapper.Segment("PID")
	if !ok || s == nil {
		t.Error("Segment(PID) should return the wrapped segment")
	}

	s, ok = wrapper.Segment("MSH")
	if ok || s != nil {
		t.Error("Segment(MSH) should return nil, false")
	}

	// Test Segments
	segs := wrapper.Segments("PID")
	if len(segs) != 1 {
		t.Errorf("Segments(PID) = %d segments, want 1", len(segs))
	}

	segs = wrapper.Segments("MSH")
	if len(segs) != 0 {
		t.Errorf("Segments(MSH) = %d segments, want 0", len(segs))
	}

	// Test AllSegments
	allSegs := wrapper.AllSegments()
	if len(allSegs) != 1 {
		t.Errorf("AllSegments() = %d segments, want 1", len(allSegs))
	}

	// Test stub methods
	if wrapper.Type() != "" {
		t.Error("Type() should return empty string")
	}
	if wrapper.ControlID() != "" {
		t.Error("ControlID() should return empty string")
	}
	if wrapper.Version() != "" {
		t.Error("Version() should return empty string")
	}
	if wrapper.Delimiters() != nil {
		t.Error("Delimiters() should return nil")
	}
}
