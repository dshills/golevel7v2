package validate

import (
	"errors"
	"testing"

	"github.com/dshills/golevel7/hl7"
)

// mockMessage implements hl7.Message for testing purposes.
type mockMessage struct {
	fields   map[string]string
	segments map[string]hl7.Segment
}

func newMockMessage() *mockMessage {
	return &mockMessage{
		fields:   make(map[string]string),
		segments: make(map[string]hl7.Segment),
	}
}

func (m *mockMessage) setField(location, value string) {
	m.fields[location] = value
}

func (m *mockMessage) Get(location string) (string, error) {
	if v, ok := m.fields[location]; ok {
		return v, nil
	}
	return "", hl7.ErrFieldNotFound
}

func (m *mockMessage) GetAll(location string) ([]string, error) {
	if v, ok := m.fields[location]; ok {
		return []string{v}, nil
	}
	return nil, hl7.ErrFieldNotFound
}

func (m *mockMessage) Set(location, value string) error {
	m.fields[location] = value
	return nil
}

func (m *mockMessage) GetAt(loc *hl7.Location) (string, error) {
	return m.Get(loc.String())
}

func (m *mockMessage) GetAllAt(loc *hl7.Location) ([]string, error) {
	return m.GetAll(loc.String())
}

func (m *mockMessage) SetAt(loc *hl7.Location, value string) error {
	return m.Set(loc.String(), value)
}

func (m *mockMessage) Segment(name string) (hl7.Segment, bool) {
	seg, ok := m.segments[name]
	return seg, ok
}

func (m *mockMessage) Segments(name string) []hl7.Segment {
	if seg, ok := m.segments[name]; ok {
		return []hl7.Segment{seg}
	}
	return nil
}

func (m *mockMessage) AllSegments() []hl7.Segment {
	result := make([]hl7.Segment, 0, len(m.segments))
	for _, seg := range m.segments {
		result = append(result, seg)
	}
	return result
}

func (m *mockMessage) AddSegment(_ hl7.Segment) error           { return nil }
func (m *mockMessage) InsertSegment(_ int, _ hl7.Segment) error { return nil }
func (m *mockMessage) RemoveSegment(_ string) bool              { return false }
func (m *mockMessage) Bytes() []byte                            { return nil }
func (m *mockMessage) String() string                           { return "" }
func (m *mockMessage) Type() string                             { return "ADT^A01" }
func (m *mockMessage) ControlID() string                        { return "12345" }
func (m *mockMessage) Version() string                          { return "2.5" }
func (m *mockMessage) Delimiters() *hl7.Delimiters              { return hl7.DefaultDelimiters() }

// Ensure mockMessage implements hl7.Message
var _ hl7.Message = (*mockMessage)(nil)

func TestRequiredRule(t *testing.T) {
	tests := []struct {
		name      string
		location  string
		setup     func(*mockMessage)
		wantValid bool
		wantCount int
	}{
		{
			name:     "field present and non-empty",
			location: "MSH.9",
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "ADT^A01")
			},
			wantValid: true,
			wantCount: 0,
		},
		{
			name:     "field present but empty",
			location: "MSH.9",
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "")
			},
			wantValid: false,
			wantCount: 1,
		},
		{
			name:     "field present but whitespace only",
			location: "MSH.9",
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "   ")
			},
			wantValid: false,
			wantCount: 1,
		},
		{
			name:      "field not present",
			location:  "MSH.9",
			setup:     func(_ *mockMessage) {},
			wantValid: false,
			wantCount: 1,
		},
		{
			name:      "nil message",
			location:  "MSH.9",
			setup:     nil, // will pass nil message
			wantValid: false,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &requiredRule{location: tt.location}

			var msg hl7.Message
			if tt.setup != nil {
				m := newMockMessage()
				tt.setup(m)
				msg = m
			}

			errs := rule.Validate(msg)
			got := len(errs) == 0

			if got != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", got, tt.wantValid)
			}
			if len(errs) != tt.wantCount {
				t.Errorf("Validate() error count = %d, want %d", len(errs), tt.wantCount)
			}
		})
	}
}

func TestValueRule(t *testing.T) {
	tests := []struct {
		name      string
		location  string
		expected  string
		setup     func(*mockMessage)
		wantValid bool
	}{
		{
			name:     "value matches exactly",
			location: "MSH.12",
			expected: "2.5",
			setup: func(m *mockMessage) {
				m.setField("MSH.12", "2.5")
			},
			wantValid: true,
		},
		{
			name:     "value does not match",
			location: "MSH.12",
			expected: "2.5",
			setup: func(m *mockMessage) {
				m.setField("MSH.12", "2.4")
			},
			wantValid: false,
		},
		{
			name:      "field not present",
			location:  "MSH.12",
			expected:  "2.5",
			setup:     func(_ *mockMessage) {},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &valueRule{location: tt.location, expected: tt.expected}
			m := newMockMessage()
			tt.setup(m)

			errs := rule.Validate(m)
			got := len(errs) == 0

			if got != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestPatternRule(t *testing.T) {
	tests := []struct {
		name      string
		location  string
		pattern   string
		setup     func(*mockMessage)
		wantValid bool
	}{
		{
			name:     "value matches pattern",
			location: "PID.3",
			pattern:  `^\d{6}$`,
			setup: func(m *mockMessage) {
				m.setField("PID.3", "123456")
			},
			wantValid: true,
		},
		{
			name:     "value does not match pattern",
			location: "PID.3",
			pattern:  `^\d{6}$`,
			setup: func(m *mockMessage) {
				m.setField("PID.3", "12345")
			},
			wantValid: false,
		},
		{
			name:     "empty value passes pattern validation",
			location: "PID.3",
			pattern:  `^\d{6}$`,
			setup: func(m *mockMessage) {
				m.setField("PID.3", "")
			},
			wantValid: true, // empty passes pattern (use required for presence)
		},
		{
			name:      "missing field passes pattern validation",
			location:  "PID.3",
			pattern:   `^\d{6}$`,
			setup:     func(_ *mockMessage) {},
			wantValid: true, // missing passes pattern (use required for presence)
		},
		{
			name:     "email pattern",
			location: "PID.13",
			pattern:  `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			setup: func(m *mockMessage) {
				m.setField("PID.13", "test@example.com")
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := At(tt.location).Pattern(tt.pattern).Build()
			m := newMockMessage()
			tt.setup(m)

			errs := rule.Validate(m)
			got := len(errs) == 0

			if got != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v, errs = %v", got, tt.wantValid, errs)
			}
		})
	}
}

func TestLengthRule(t *testing.T) {
	tests := []struct {
		name      string
		location  string
		min       int
		max       int
		setup     func(*mockMessage)
		wantValid bool
	}{
		{
			name:     "length within bounds",
			location: "MSH.10",
			min:      1,
			max:      20,
			setup: func(m *mockMessage) {
				m.setField("MSH.10", "12345")
			},
			wantValid: true,
		},
		{
			name:     "length below minimum",
			location: "MSH.10",
			min:      5,
			max:      20,
			setup: func(m *mockMessage) {
				m.setField("MSH.10", "123")
			},
			wantValid: false,
		},
		{
			name:     "length above maximum",
			location: "MSH.10",
			min:      1,
			max:      5,
			setup: func(m *mockMessage) {
				m.setField("MSH.10", "1234567890")
			},
			wantValid: false,
		},
		{
			name:     "exact minimum length",
			location: "MSH.10",
			min:      5,
			max:      10,
			setup: func(m *mockMessage) {
				m.setField("MSH.10", "12345")
			},
			wantValid: true,
		},
		{
			name:     "exact maximum length",
			location: "MSH.10",
			min:      5,
			max:      10,
			setup: func(m *mockMessage) {
				m.setField("MSH.10", "1234567890")
			},
			wantValid: true,
		},
		{
			name:     "no minimum (only max)",
			location: "MSH.10",
			min:      0,
			max:      5,
			setup: func(m *mockMessage) {
				m.setField("MSH.10", "12")
			},
			wantValid: true,
		},
		{
			name:     "no maximum (only min)",
			location: "MSH.10",
			min:      3,
			max:      0,
			setup: func(m *mockMessage) {
				m.setField("MSH.10", "1234567890123")
			},
			wantValid: true,
		},
		{
			name:      "missing field passes length validation",
			location:  "MSH.10",
			min:       5,
			max:       20,
			setup:     func(_ *mockMessage) {},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &lengthRule{location: tt.location, min: tt.min, max: tt.max}
			m := newMockMessage()
			tt.setup(m)

			errs := rule.Validate(m)
			got := len(errs) == 0

			if got != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v, errs = %v", got, tt.wantValid, errs)
			}
		})
	}
}

func TestOneOfRule(t *testing.T) {
	tests := []struct {
		name      string
		location  string
		allowed   []string
		setup     func(*mockMessage)
		wantValid bool
	}{
		{
			name:     "value in allowed list",
			location: "PV1.2",
			allowed:  []string{"I", "O", "E", "P", "R"},
			setup: func(m *mockMessage) {
				m.setField("PV1.2", "I")
			},
			wantValid: true,
		},
		{
			name:     "value not in allowed list",
			location: "PV1.2",
			allowed:  []string{"I", "O", "E", "P", "R"},
			setup: func(m *mockMessage) {
				m.setField("PV1.2", "X")
			},
			wantValid: false,
		},
		{
			name:     "empty value passes oneOf validation",
			location: "PV1.2",
			allowed:  []string{"I", "O", "E"},
			setup: func(m *mockMessage) {
				m.setField("PV1.2", "")
			},
			wantValid: true,
		},
		{
			name:      "missing field passes oneOf validation",
			location:  "PV1.2",
			allowed:   []string{"I", "O", "E"},
			setup:     func(_ *mockMessage) {},
			wantValid: true,
		},
		{
			name:     "case sensitive match",
			location: "PV1.2",
			allowed:  []string{"I", "O", "E"},
			setup: func(m *mockMessage) {
				m.setField("PV1.2", "i")
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &oneOfRule{location: tt.location, allowed: tt.allowed}
			m := newMockMessage()
			tt.setup(m)

			errs := rule.Validate(m)
			got := len(errs) == 0

			if got != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestCustomRule(t *testing.T) {
	tests := []struct {
		name      string
		location  string
		fn        func(string) error
		setup     func(*mockMessage)
		wantValid bool
	}{
		{
			name:     "custom validation passes",
			location: "PID.7",
			fn: func(v string) error {
				if len(v) != 8 {
					return errors.New("date must be 8 characters (YYYYMMDD)")
				}
				return nil
			},
			setup: func(m *mockMessage) {
				m.setField("PID.7", "19850101")
			},
			wantValid: true,
		},
		{
			name:     "custom validation fails",
			location: "PID.7",
			fn: func(v string) error {
				if len(v) != 8 {
					return errors.New("date must be 8 characters (YYYYMMDD)")
				}
				return nil
			},
			setup: func(m *mockMessage) {
				m.setField("PID.7", "1985-01-01")
			},
			wantValid: false,
		},
		{
			name:     "missing field passes custom validation",
			location: "PID.7",
			fn: func(_ string) error {
				return errors.New("should not be called")
			},
			setup:     func(_ *mockMessage) {},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &customRule{location: tt.location, fn: tt.fn}
			m := newMockMessage()
			tt.setup(m)

			errs := rule.Validate(m)
			got := len(errs) == 0

			if got != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestCompositeRule(t *testing.T) {
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
				&requiredRule{location: "MSH.9"},
				&lengthRule{location: "MSH.9", min: 1, max: 20},
			},
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "ADT^A01")
			},
			wantValid: true,
			wantCount: 0,
		},
		{
			name: "one rule fails",
			rules: []Rule{
				&requiredRule{location: "MSH.9"},
				&lengthRule{location: "MSH.9", min: 10, max: 20},
			},
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "ADT")
			},
			wantValid: false,
			wantCount: 1,
		},
		{
			name: "multiple rules fail",
			rules: []Rule{
				&requiredRule{location: "MSH.9"},
				&requiredRule{location: "MSH.10"},
			},
			setup: func(_ *mockMessage) {
				// both fields missing
			},
			wantValid: false,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &compositeRule{location: "MSH", rules: tt.rules}
			m := newMockMessage()
			tt.setup(m)

			errs := rule.Validate(m)
			got := len(errs) == 0

			if got != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", got, tt.wantValid)
			}
			if len(errs) != tt.wantCount {
				t.Errorf("Validate() error count = %d, want %d", len(errs), tt.wantCount)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		contains []string
	}{
		{
			name: "full error",
			err: ValidationError{
				Location: "MSH.9",
				Rule:     "required",
				Message:  "field is required",
				Expected: "non-empty value",
				Actual:   "empty",
			},
			contains: []string{"MSH.9", "required", "field is required", "non-empty value", "empty"},
		},
		{
			name: "minimal error",
			err: ValidationError{
				Message: "something went wrong",
			},
			contains: []string{"validation error", "something went wrong"},
		},
		{
			name: "expected only",
			err: ValidationError{
				Location: "PID.3",
				Expected: "numeric value",
			},
			contains: []string{"PID.3", "numeric value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, s := range tt.contains {
				if !containsString(errStr, s) {
					t.Errorf("Error() = %q, should contain %q", errStr, s)
				}
			}
		})
	}
}

func TestValidationWarning_String(t *testing.T) {
	warn := ValidationWarning{
		Location: "PID.5",
		Rule:     "recommended",
		Message:  "patient name should have a last name component",
	}

	str := warn.String()
	if !containsString(str, "PID.5") {
		t.Errorf("String() = %q, should contain location", str)
	}
	if !containsString(str, "recommended") {
		t.Errorf("String() = %q, should contain rule", str)
	}
	if !containsString(str, "patient name") {
		t.Errorf("String() = %q, should contain message", str)
	}
}

func TestRuleDescriptions(t *testing.T) {
	tests := []struct {
		name            string
		rule            Rule
		wantDescription string
	}{
		{
			name:            "required rule default description",
			rule:            &requiredRule{location: "MSH.9"},
			wantDescription: "MSH.9 is required",
		},
		{
			name:            "required rule custom description",
			rule:            &requiredRule{location: "MSH.9", description: "Message Type must be present"},
			wantDescription: "Message Type must be present",
		},
		{
			name:            "value rule default description",
			rule:            &valueRule{location: "MSH.12", expected: "2.5"},
			wantDescription: `MSH.12 must equal "2.5"`,
		},
		{
			name:            "length rule both bounds",
			rule:            &lengthRule{location: "MSH.10", min: 5, max: 20},
			wantDescription: "MSH.10 length must be between 5 and 20",
		},
		{
			name:            "length rule min only",
			rule:            &lengthRule{location: "MSH.10", min: 5},
			wantDescription: "MSH.10 length must be at least 5",
		},
		{
			name:            "length rule max only",
			rule:            &lengthRule{location: "MSH.10", max: 20},
			wantDescription: "MSH.10 length must be at most 20",
		},
		{
			name:            "oneOf rule",
			rule:            &oneOfRule{location: "PV1.2", allowed: []string{"I", "O", "E"}},
			wantDescription: "PV1.2 must be one of [I, O, E]",
		},
		{
			name:            "custom rule default description",
			rule:            &customRule{location: "PID.7", fn: func(string) error { return nil }},
			wantDescription: "PID.7 custom validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rule.Description()
			if got != tt.wantDescription {
				t.Errorf("Description() = %q, want %q", got, tt.wantDescription)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
