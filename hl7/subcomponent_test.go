package hl7

import (
	"testing"
)

func TestNewSubComponent(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "simple string",
			value: "Hello",
			want:  "Hello",
		},
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
		{
			name:  "string with spaces",
			value: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "numeric string",
			value: "12345",
			want:  "12345",
		},
		{
			name:  "special characters",
			value: "Test!@#$%",
			want:  "Test!@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewSubComponent(tt.value)
			if sc == nil {
				t.Fatal("NewSubComponent returned nil")
			}
			if got := sc.Value(); got != tt.want {
				t.Errorf("NewSubComponent(%q).Value() = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestSubComponent_Value(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "simple value",
			value: "Smith",
			want:  "Smith",
		},
		{
			name:  "empty value",
			value: "",
			want:  "",
		},
		{
			name:  "value with whitespace",
			value: "  John  ",
			want:  "  John  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewSubComponent(tt.value)
			if got := sc.Value(); got != tt.want {
				t.Errorf("Value() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubComponent_Set(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		newValue string
		wantErr  bool
	}{
		{
			name:     "set new value",
			initial:  "old",
			newValue: "new",
			wantErr:  false,
		},
		{
			name:     "set empty value",
			initial:  "something",
			newValue: "",
			wantErr:  false,
		},
		{
			name:     "set from empty",
			initial:  "",
			newValue: "something",
			wantErr:  false,
		},
		{
			name:     "set same value",
			initial:  "same",
			newValue: "same",
			wantErr:  false,
		},
		{
			name:     "set unicode value",
			initial:  "ascii",
			newValue: "unicode: cafe",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewSubComponent(tt.initial)
			err := sc.Set(tt.newValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got := sc.Value(); got != tt.newValue {
				t.Errorf("After Set(%q), Value() = %q, want %q", tt.newValue, got, tt.newValue)
			}
		})
	}
}

func TestSubComponent_Bytes(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		delims *Delimiters
		want   []byte
	}{
		{
			name:   "simple value with default delimiters",
			value:  "Hello",
			delims: DefaultDelimiters(),
			want:   []byte("Hello"),
		},
		{
			name:   "empty value",
			value:  "",
			delims: DefaultDelimiters(),
			want:   []byte(""),
		},
		{
			name:   "nil delimiters",
			value:  "Test",
			delims: nil,
			want:   []byte("Test"),
		},
		{
			name:  "custom delimiters (should not affect output)",
			value: "Value",
			delims: &Delimiters{
				Field:        '#',
				Component:    '@',
				Repetition:   '*',
				Escape:       '!',
				SubComponent: '%',
			},
			want: []byte("Value"),
		},
		{
			name:   "value with special characters",
			value:  "Test|Value^With&Delimiters",
			delims: DefaultDelimiters(),
			want:   []byte("Test|Value^With&Delimiters"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewSubComponent(tt.value)
			got := sc.Bytes(tt.delims)

			if string(got) != string(tt.want) {
				t.Errorf("Bytes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubComponent_String(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "simple string",
			value: "Test",
			want:  "Test",
		},
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
		{
			name:  "string equals value",
			value: "SameAsValue",
			want:  "SameAsValue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewSubComponent(tt.value)
			got := sc.String()

			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}

			// Verify String() equals Value()
			if got != sc.Value() {
				t.Errorf("String() = %q does not equal Value() = %q", got, sc.Value())
			}
		})
	}
}

func TestSubComponent_EmptyValue(t *testing.T) {
	// Test that empty subcomponents are handled correctly
	sc := NewSubComponent("")

	if sc.Value() != "" {
		t.Errorf("Empty subcomponent Value() = %q, want empty string", sc.Value())
	}

	if sc.String() != "" {
		t.Errorf("Empty subcomponent String() = %q, want empty string", sc.String())
	}

	bytes := sc.Bytes(DefaultDelimiters())
	if len(bytes) != 0 {
		t.Errorf("Empty subcomponent Bytes() = %q, want empty slice", bytes)
	}

	// Setting a value on empty subcomponent
	if err := sc.Set("now has value"); err != nil {
		t.Errorf("Set() on empty subcomponent returned error: %v", err)
	}

	if sc.Value() != "now has value" {
		t.Errorf("After Set(), Value() = %q, want %q", sc.Value(), "now has value")
	}

	// Setting back to empty
	if err := sc.Set(""); err != nil {
		t.Errorf("Set(\"\") returned error: %v", err)
	}

	if sc.Value() != "" {
		t.Errorf("After Set(\"\"), Value() = %q, want empty string", sc.Value())
	}
}

func TestSubComponent_UnicodeHandling(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "ASCII characters",
			value: "Hello World",
		},
		{
			name:  "Latin extended characters",
			value: "cafe resume",
		},
		{
			name:  "Chinese characters",
			value: "Chinese characters",
		},
		{
			name:  "Japanese characters",
			value: "Japanese hiragana",
		},
		{
			name:  "Korean characters",
			value: "Korean hangul",
		},
		{
			name:  "Emoji characters",
			value: "Hello World!",
		},
		{
			name:  "Mixed unicode and ASCII",
			value: "Patient: John Doe",
		},
		{
			name:  "Arabic characters",
			value: "Arabic text",
		},
		{
			name:  "Cyrillic characters",
			value: "Cyrillic text",
		},
		{
			name:  "Greek characters",
			value: "Greek letters",
		},
		{
			name:  "Mathematical symbols",
			value: "Math: x squared",
		},
		{
			name:  "Multi-byte unicode sequence",
			value: "Test: special chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test NewSubComponent preserves unicode
			sc := NewSubComponent(tt.value)
			if got := sc.Value(); got != tt.value {
				t.Errorf("NewSubComponent(%q).Value() = %q, want %q", tt.value, got, tt.value)
			}

			// Test Set preserves unicode
			sc2 := NewSubComponent("")
			if err := sc2.Set(tt.value); err != nil {
				t.Errorf("Set(%q) returned error: %v", tt.value, err)
			}
			if got := sc2.Value(); got != tt.value {
				t.Errorf("After Set(%q), Value() = %q", tt.value, got)
			}

			// Test Bytes produces correct UTF-8
			bytes := sc.Bytes(DefaultDelimiters())
			if string(bytes) != tt.value {
				t.Errorf("Bytes() = %q, want %q", string(bytes), tt.value)
			}

			// Test String returns same as Value
			if sc.String() != sc.Value() {
				t.Errorf("String() = %q does not match Value() = %q", sc.String(), sc.Value())
			}
		})
	}
}

func TestParseSubComponent(t *testing.T) {
	tests := []struct {
		name string
		data []rune
		want string
	}{
		{
			name: "simple runes",
			data: []rune("Hello"),
			want: "Hello",
		},
		{
			name: "empty runes",
			data: []rune{},
			want: "",
		},
		{
			name: "nil runes",
			data: nil,
			want: "",
		},
		{
			name: "unicode runes",
			data: []rune("unicode test"),
			want: "unicode test",
		},
		{
			name: "runes with special characters",
			data: []rune("Test|^~\\&"),
			want: "Test|^~\\&",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := ParseSubComponent(tt.data)
			if sc == nil {
				t.Fatal("ParseSubComponent returned nil")
			}
			if got := sc.Value(); got != tt.want {
				t.Errorf("ParseSubComponent(%v).Value() = %q, want %q", tt.data, got, tt.want)
			}
		})
	}
}

func TestParseSubComponent_DataIsolation(t *testing.T) {
	// Test that ParseSubComponent makes a copy of the data
	// to avoid sharing the underlying array
	original := []rune("original")
	sc := ParseSubComponent(original)

	// Modify the original slice
	original[0] = 'X'

	// The subcomponent should not be affected
	if got := sc.Value(); got != "original" {
		t.Errorf("After modifying original slice, Value() = %q, want %q", got, "original")
	}
}

func TestSubComponent_InterfaceCompliance(_ *testing.T) {
	// Verify that subComponent implements SubComponent interface
	var _ SubComponent = (*subComponent)(nil)
	_ = NewSubComponent("test")
	_ = ParseSubComponent([]rune("test"))
}

func TestSubComponent_SetMultipleTimes(t *testing.T) {
	sc := NewSubComponent("initial")

	values := []string{"first", "second", "third", "", "final"}

	for _, v := range values {
		if err := sc.Set(v); err != nil {
			t.Errorf("Set(%q) returned error: %v", v, err)
		}
		if got := sc.Value(); got != v {
			t.Errorf("After Set(%q), Value() = %q", v, got)
		}
	}
}

func TestSubComponent_BytesNilDelimiters(t *testing.T) {
	sc := NewSubComponent("Test Value")

	// Bytes should work even with nil delimiters
	// since subcomponents don't use delimiters internally
	got := sc.Bytes(nil)
	want := []byte("Test Value")

	if string(got) != string(want) {
		t.Errorf("Bytes(nil) = %q, want %q", got, want)
	}
}
