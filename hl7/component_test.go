package hl7

import (
	"bytes"
	"testing"
)

func TestNewComponent(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "simple value",
			value: "test",
			want:  "test",
		},
		{
			name:  "empty value",
			value: "",
			want:  "",
		},
		{
			name:  "value with special characters",
			value: "test&value",
			want:  "test&value",
		},
		{
			name:  "unicode value",
			value: "test\u00e9",
			want:  "test\u00e9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewComponent(tt.value)
			if got := c.Value(); got != tt.want {
				t.Errorf("NewComponent(%q).Value() = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseComponent_MultipleSubcomponents(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name              string
		data              string
		wantValue         string
		wantSubCompCount  int
		wantSubCompValues []string
	}{
		{
			name:              "three subcomponents",
			data:              "first&second&third",
			wantValue:         "first",
			wantSubCompCount:  3,
			wantSubCompValues: []string{"first", "second", "third"},
		},
		{
			name:              "two subcomponents",
			data:              "alpha&beta",
			wantValue:         "alpha",
			wantSubCompCount:  2,
			wantSubCompValues: []string{"alpha", "beta"},
		},
		{
			name:              "subcomponent with empty parts",
			data:              "first&&third",
			wantValue:         "first",
			wantSubCompCount:  3,
			wantSubCompValues: []string{"first", "", "third"},
		},
		{
			name:              "trailing empty subcomponent",
			data:              "first&second&",
			wantValue:         "first",
			wantSubCompCount:  3,
			wantSubCompValues: []string{"first", "second", ""},
		},
		{
			name:              "leading empty subcomponent",
			data:              "&second&third",
			wantValue:         "",
			wantSubCompCount:  3,
			wantSubCompValues: []string{"", "second", "third"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseComponent([]rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseComponent() error = %v", err)
			}

			if got := c.Value(); got != tt.wantValue {
				t.Errorf("Value() = %q, want %q", got, tt.wantValue)
			}

			subComps := c.SubComponents()
			if len(subComps) != tt.wantSubCompCount {
				t.Errorf("SubComponents() count = %d, want %d", len(subComps), tt.wantSubCompCount)
			}

			for i, wantValue := range tt.wantSubCompValues {
				if i >= len(subComps) {
					t.Errorf("missing subcomponent at index %d", i)
					continue
				}
				if got := subComps[i].Value(); got != wantValue {
					t.Errorf("SubComponents()[%d].Value() = %q, want %q", i, got, wantValue)
				}
			}
		})
	}
}

func TestParseComponent_NoSubcomponents(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name      string
		data      string
		wantValue string
	}{
		{
			name:      "simple value",
			data:      "simple",
			wantValue: "simple",
		},
		{
			name:      "empty value",
			data:      "",
			wantValue: "",
		},
		{
			name:      "value with other delimiters",
			data:      "test^value~repeat",
			wantValue: "test^value~repeat",
		},
		{
			name:      "unicode value",
			data:      "patient\u00e9name",
			wantValue: "patient\u00e9name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseComponent([]rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseComponent() error = %v", err)
			}

			if got := c.Value(); got != tt.wantValue {
				t.Errorf("Value() = %q, want %q", got, tt.wantValue)
			}

			// When there's no subcomponent delimiter, SubComponents should return empty
			subComps := c.SubComponents()
			if len(subComps) != 0 {
				t.Errorf("SubComponents() should be empty for non-parsed data, got %d", len(subComps))
			}
		})
	}
}

func TestParseComponent_NilDelimiters(t *testing.T) {
	// Should use default delimiters when nil is passed
	c, err := ParseComponent([]rune("first&second"), nil)
	if err != nil {
		t.Fatalf("ParseComponent() error = %v", err)
	}

	subComps := c.SubComponents()
	if len(subComps) != 2 {
		t.Errorf("SubComponents() count = %d, want 2", len(subComps))
	}
}

func TestComponent_SubComponent_1BasedIndex(t *testing.T) {
	delims := DefaultDelimiters()
	c, err := ParseComponent([]rune("first&second&third"), delims)
	if err != nil {
		t.Fatalf("ParseComponent() error = %v", err)
	}

	tests := []struct {
		name      string
		index     int
		wantValue string
		wantOK    bool
	}{
		{
			name:      "index 0 (invalid)",
			index:     0,
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "index 1 (first)",
			index:     1,
			wantValue: "first",
			wantOK:    true,
		},
		{
			name:      "index 2 (second)",
			index:     2,
			wantValue: "second",
			wantOK:    true,
		},
		{
			name:      "index 3 (third)",
			index:     3,
			wantValue: "third",
			wantOK:    true,
		},
		{
			name:      "index 4 (out of range)",
			index:     4,
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "negative index",
			index:     -1,
			wantValue: "",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc, ok := c.SubComponent(tt.index)
			if ok != tt.wantOK {
				t.Errorf("SubComponent(%d) ok = %v, want %v", tt.index, ok, tt.wantOK)
			}
			if ok && sc.Value() != tt.wantValue {
				t.Errorf("SubComponent(%d).Value() = %q, want %q", tt.index, sc.Value(), tt.wantValue)
			}
		})
	}
}

func TestComponent_SubComponent_NoSubcomponents(t *testing.T) {
	// When a component has no subcomponents, SubComponent should return false
	c := NewComponent("simple")

	sc, ok := c.SubComponent(1)
	if ok {
		t.Errorf("SubComponent(1) should return false for component with no subcomponents, got %v", sc)
	}

	sc, ok = c.SubComponent(2)
	if ok {
		t.Errorf("SubComponent(2) should return false for component with no subcomponents, got %v", sc)
	}
}

func TestComponent_Set(t *testing.T) {
	// Create a component with subcomponents
	delims := DefaultDelimiters()
	c, err := ParseComponent([]rune("first&second&third"), delims)
	if err != nil {
		t.Fatalf("ParseComponent() error = %v", err)
	}

	// Verify initial state
	if len(c.SubComponents()) != 3 {
		t.Errorf("initial SubComponents() count = %d, want 3", len(c.SubComponents()))
	}

	// Set a new value
	if err := c.Set("newvalue"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify subcomponents are cleared
	if len(c.SubComponents()) != 0 {
		t.Errorf("after Set(), SubComponents() count = %d, want 0", len(c.SubComponents()))
	}

	// Verify new value
	if got := c.Value(); got != "newvalue" {
		t.Errorf("after Set(), Value() = %q, want %q", got, "newvalue")
	}
}

func TestComponent_SetSubComponent(t *testing.T) {
	tests := []struct {
		name              string
		initial           string
		setIndex          int
		setValue          string
		wantErr           bool
		wantSubCompCount  int
		wantSubCompValues []string
	}{
		{
			name:              "set existing subcomponent",
			initial:           "first&second&third",
			setIndex:          2,
			setValue:          "modified",
			wantErr:           false,
			wantSubCompCount:  3,
			wantSubCompValues: []string{"first", "modified", "third"},
		},
		{
			name:              "expand subcomponents",
			initial:           "first&second",
			setIndex:          5,
			setValue:          "fifth",
			wantErr:           false,
			wantSubCompCount:  5,
			wantSubCompValues: []string{"first", "second", "", "", "fifth"},
		},
		{
			name:     "invalid index 0",
			initial:  "test",
			setIndex: 0,
			setValue: "invalid",
			wantErr:  true,
		},
		{
			name:     "negative index",
			initial:  "test",
			setIndex: -1,
			setValue: "invalid",
			wantErr:  true,
		},
	}

	delims := DefaultDelimiters()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseComponent([]rune(tt.initial), delims)
			if err != nil {
				t.Fatalf("ParseComponent() error = %v", err)
			}

			err = c.SetSubComponent(tt.setIndex, tt.setValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetSubComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			subComps := c.SubComponents()
			if len(subComps) != tt.wantSubCompCount {
				t.Errorf("SubComponents() count = %d, want %d", len(subComps), tt.wantSubCompCount)
				return
			}

			for i, wantValue := range tt.wantSubCompValues {
				if got := subComps[i].Value(); got != wantValue {
					t.Errorf("SubComponents()[%d].Value() = %q, want %q", i, got, wantValue)
				}
			}
		})
	}
}

func TestComponent_SetSubComponent_FromRawValue(t *testing.T) {
	// When setting a subcomponent on a component that has only a raw value,
	// it should convert to the subcomponent model
	c := NewComponent("original")

	if err := c.SetSubComponent(2, "second"); err != nil {
		t.Fatalf("SetSubComponent() error = %v", err)
	}

	subComps := c.SubComponents()
	if len(subComps) != 2 {
		t.Errorf("SubComponents() count = %d, want 2", len(subComps))
		return
	}

	if got := subComps[0].Value(); got != "original" {
		t.Errorf("SubComponents()[0].Value() = %q, want %q", got, "original")
	}

	if got := subComps[1].Value(); got != "second" {
		t.Errorf("SubComponents()[1].Value() = %q, want %q", got, "second")
	}
}

func TestComponent_Bytes(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name string
		data string
		want []byte
	}{
		{
			name: "with subcomponents",
			data: "first&second&third",
			want: []byte("first&second&third"),
		},
		{
			name: "without subcomponents",
			data: "simple",
			want: []byte("simple"),
		},
		{
			name: "empty",
			data: "",
			want: []byte(""),
		},
		{
			name: "empty subcomponents",
			data: "first&&third",
			want: []byte("first&&third"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseComponent([]rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseComponent() error = %v", err)
			}

			got := c.Bytes(delims)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Bytes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComponent_Bytes_NilDelimiters(t *testing.T) {
	c, err := ParseComponent([]rune("first&second"), DefaultDelimiters())
	if err != nil {
		t.Fatalf("ParseComponent() error = %v", err)
	}

	// Should use default delimiters when nil is passed
	got := c.Bytes(nil)
	want := []byte("first&second")
	if !bytes.Equal(got, want) {
		t.Errorf("Bytes(nil) = %q, want %q", got, want)
	}
}

func TestComponent_Bytes_CustomDelimiters(t *testing.T) {
	// Parse with default delimiters
	c, err := ParseComponent([]rune("first&second"), DefaultDelimiters())
	if err != nil {
		t.Fatalf("ParseComponent() error = %v", err)
	}

	// Encode with custom delimiters
	customDelims := &Delimiters{
		Field:        '|',
		Component:    '^',
		Repetition:   '~',
		Escape:       '\\',
		SubComponent: '#', // Custom subcomponent delimiter
		Truncation:   '@',
	}

	got := c.Bytes(customDelims)
	want := []byte("first#second")
	if !bytes.Equal(got, want) {
		t.Errorf("Bytes(customDelims) = %q, want %q", got, want)
	}
}

func TestComponent_String(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "with subcomponents",
			data: "first&second&third",
			want: "first&second&third",
		},
		{
			name: "without subcomponents",
			data: "simple",
			want: "simple",
		},
		{
			name: "empty",
			data: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseComponent([]rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseComponent() error = %v", err)
			}

			if got := c.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComponent_RoundTrip(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name string
		data string
	}{
		{
			name: "with subcomponents",
			data: "first&second&third",
		},
		{
			name: "simple value",
			data: "simple",
		},
		{
			name: "empty subcomponents",
			data: "first&&third",
		},
		{
			name: "single empty",
			data: "",
		},
		{
			name: "unicode content",
			data: "test\u00e9&value\u00f1",
		},
		{
			name: "only empty subcomponents",
			data: "&&",
		},
		{
			name: "complex medical data",
			data: "Smith&John&Q&Jr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the data
			c1, err := ParseComponent([]rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseComponent() error = %v", err)
			}

			// Encode back to bytes
			encoded := c1.Bytes(delims)

			// Parse again
			c2, err := ParseComponent([]rune(string(encoded)), delims)
			if err != nil {
				t.Fatalf("second ParseComponent() error = %v", err)
			}

			// Compare
			if c1.String() != c2.String() {
				t.Errorf("round-trip failed: original %q, after round-trip %q", c1.String(), c2.String())
			}

			// Also compare to original data
			if string(encoded) != tt.data {
				t.Errorf("encoded data = %q, want %q", string(encoded), tt.data)
			}
		})
	}
}

func TestComponent_SubComponents_ReturnsCopy(t *testing.T) {
	delims := DefaultDelimiters()
	c, err := ParseComponent([]rune("first&second&third"), delims)
	if err != nil {
		t.Fatalf("ParseComponent() error = %v", err)
	}

	// Get subcomponents
	subComps1 := c.SubComponents()
	subComps2 := c.SubComponents()

	// Verify they are different slices
	if &subComps1[0] == &subComps2[0] {
		t.Error("SubComponents() should return a copy, not the same slice")
	}

	// Verify original is not affected by modifying the returned slice
	original := c.Value()
	subComps1[0] = NewSubComponent("modified")

	if c.Value() != original {
		t.Errorf("modifying returned slice affected original: got %q, want %q", c.Value(), original)
	}
}

func TestComponent_Value_EmptySubcomponents(t *testing.T) {
	delims := DefaultDelimiters()

	// Component with empty first subcomponent
	c, err := ParseComponent([]rune("&second&third"), delims)
	if err != nil {
		t.Fatalf("ParseComponent() error = %v", err)
	}

	// Value should return first subcomponent (empty)
	if got := c.Value(); got != "" {
		t.Errorf("Value() = %q, want empty string", got)
	}

	// But second subcomponent should have value
	sc, ok := c.SubComponent(2)
	if !ok {
		t.Fatal("SubComponent(2) should return true")
	}
	if got := sc.Value(); got != "second" {
		t.Errorf("SubComponent(2).Value() = %q, want %q", got, "second")
	}
}
