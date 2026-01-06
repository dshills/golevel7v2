package hl7

import (
	"testing"
)

func TestNewRepetition(t *testing.T) {
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
			name:  "value with special characters",
			value: "O'Brien",
			want:  "O'Brien",
		},
		{
			name:  "unicode value",
			value: "Muller",
			want:  "Muller",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := NewRepetition(tt.value)
			if got := rep.Value(); got != tt.want {
				t.Errorf("NewRepetition(%q).Value() = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseRepetition_SimpleValue(t *testing.T) {
	tests := []struct {
		name    string
		data    []rune
		delims  *Delimiters
		want    string
		wantErr bool
	}{
		{
			name:    "simple value no components",
			data:    []rune("Smith"),
			delims:  DefaultDelimiters(),
			want:    "Smith",
			wantErr: false,
		},
		{
			name:    "empty value",
			data:    []rune(""),
			delims:  DefaultDelimiters(),
			want:    "",
			wantErr: false,
		},
		{
			name:    "nil delimiters uses defaults",
			data:    []rune("Value"),
			delims:  nil,
			want:    "Value",
			wantErr: false,
		},
		{
			name:    "unicode characters",
			data:    []rune("Cafe"),
			delims:  DefaultDelimiters(),
			want:    "Cafe",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := ParseRepetition(tt.data, tt.delims)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepetition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got := rep.Value(); got != tt.want {
				t.Errorf("ParseRepetition().Value() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseRepetition_MultipleComponents(t *testing.T) {
	tests := []struct {
		name           string
		data           []rune
		delims         *Delimiters
		wantValue      string // Value() returns full encoded value with all components
		wantCompCount  int
		wantComponents []string
	}{
		{
			name:           "two components",
			data:           []rune("Smith^John"),
			delims:         DefaultDelimiters(),
			wantValue:      "Smith^John",
			wantCompCount:  2,
			wantComponents: []string{"Smith", "John"},
		},
		{
			name:           "three components",
			data:           []rune("Smith^John^Q"),
			delims:         DefaultDelimiters(),
			wantValue:      "Smith^John^Q",
			wantCompCount:  3,
			wantComponents: []string{"Smith", "John", "Q"},
		},
		{
			name:           "empty components",
			data:           []rune("Smith^^Q"),
			delims:         DefaultDelimiters(),
			wantValue:      "Smith^^Q",
			wantCompCount:  3,
			wantComponents: []string{"Smith", "", "Q"},
		},
		{
			name:           "all empty components",
			data:           []rune("^^"),
			delims:         DefaultDelimiters(),
			wantValue:      "^^",
			wantCompCount:  3,
			wantComponents: []string{"", "", ""},
		},
		{
			name:           "trailing delimiter",
			data:           []rune("Smith^"),
			delims:         DefaultDelimiters(),
			wantValue:      "Smith^",
			wantCompCount:  2,
			wantComponents: []string{"Smith", ""},
		},
		{
			name:           "leading delimiter",
			data:           []rune("^John"),
			delims:         DefaultDelimiters(),
			wantValue:      "^John",
			wantCompCount:  2,
			wantComponents: []string{"", "John"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := ParseRepetition(tt.data, tt.delims)
			if err != nil {
				t.Fatalf("ParseRepetition() error = %v", err)
			}

			if got := rep.Value(); got != tt.wantValue {
				t.Errorf("Value() = %q, want %q", got, tt.wantValue)
			}

			comps := rep.Components()
			if len(comps) != tt.wantCompCount {
				t.Errorf("Components() count = %d, want %d", len(comps), tt.wantCompCount)
			}

			for i, wantVal := range tt.wantComponents {
				if i >= len(comps) {
					t.Errorf("Missing component at index %d", i)
					continue
				}
				if got := comps[i].Value(); got != wantVal {
					t.Errorf("Component[%d].Value() = %q, want %q", i, got, wantVal)
				}
			}
		})
	}
}

func TestRepetition_ComponentAccess(t *testing.T) {
	// Parse a repetition with multiple components
	rep, err := ParseRepetition([]rune("Smith^John^Q^Jr"), DefaultDelimiters())
	if err != nil {
		t.Fatalf("ParseRepetition() error = %v", err)
	}

	tests := []struct {
		name      string
		index     int
		wantValue string
		wantOK    bool
	}{
		{
			name:      "first component (1-based)",
			index:     1,
			wantValue: "Smith",
			wantOK:    true,
		},
		{
			name:      "second component",
			index:     2,
			wantValue: "John",
			wantOK:    true,
		},
		{
			name:      "third component",
			index:     3,
			wantValue: "Q",
			wantOK:    true,
		},
		{
			name:      "fourth component",
			index:     4,
			wantValue: "Jr",
			wantOK:    true,
		},
		{
			name:      "zero index (invalid)",
			index:     0,
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "negative index (invalid)",
			index:     -1,
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "out of range index",
			index:     5,
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "way out of range",
			index:     100,
			wantValue: "",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, ok := rep.Component(tt.index)
			if ok != tt.wantOK {
				t.Errorf("Component(%d) ok = %v, want %v", tt.index, ok, tt.wantOK)
			}
			if ok && comp.Value() != tt.wantValue {
				t.Errorf("Component(%d).Value() = %q, want %q", tt.index, comp.Value(), tt.wantValue)
			}
		})
	}
}

func TestRepetition_ComponentAccess_NoComponents(t *testing.T) {
	// Create a repetition without explicit components
	rep := NewRepetition("SimpleValue")

	// In HL7 semantics, even a simple value can be accessed as component 1
	// This allows subcomponent access for values like "ID&SubID"
	comp, ok := rep.Component(1)
	if !ok {
		t.Errorf("Component(1) on simple repetition should return true (implicit component)")
	}
	if comp == nil {
		t.Errorf("Component(1) on simple repetition should return a component")
	} else if comp.Value() != "SimpleValue" {
		t.Errorf("Component(1).Value() = %q, want %q", comp.Value(), "SimpleValue")
	}

	// Component(2) and beyond should return false
	_, ok = rep.Component(2)
	if ok {
		t.Errorf("Component(2) on simple repetition should return false")
	}

	// Components() should return empty slice (no explicit components parsed)
	comps := rep.Components()
	if len(comps) != 0 {
		t.Errorf("Components() on simple repetition should return empty slice, got %d", len(comps))
	}
}

func TestRepetition_SubcomponentAccess_Simple(t *testing.T) {
	// Test that a simple value with subcomponents can be accessed
	rep := NewRepetition("ID&SubID")

	// Component 1 should give us access to subcomponents
	comp, ok := rep.Component(1)
	if !ok {
		t.Fatal("Component(1) should return true")
	}

	// Subcomponent 1 should be "ID"
	sub1, ok := comp.SubComponent(1)
	if !ok {
		t.Error("SubComponent(1) should return true")
	} else if sub1.Value() != "ID" {
		t.Errorf("SubComponent(1).Value() = %q, want %q", sub1.Value(), "ID")
	}

	// Subcomponent 2 should be "SubID"
	sub2, ok := comp.SubComponent(2)
	if !ok {
		t.Error("SubComponent(2) should return true")
	} else if sub2.Value() != "SubID" {
		t.Errorf("SubComponent(2).Value() = %q, want %q", sub2.Value(), "SubID")
	}
}

func TestRepetition_Bytes(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name string
		data []rune
		want string
	}{
		{
			name: "simple value",
			data: []rune("Smith"),
			want: "Smith",
		},
		{
			name: "two components",
			data: []rune("Smith^John"),
			want: "Smith^John",
		},
		{
			name: "multiple components",
			data: []rune("Smith^John^Q^Jr"),
			want: "Smith^John^Q^Jr",
		},
		{
			name: "empty components",
			data: []rune("Smith^^Q"),
			want: "Smith^^Q",
		},
		{
			name: "empty value",
			data: []rune(""),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := ParseRepetition(tt.data, delims)
			if err != nil {
				t.Fatalf("ParseRepetition() error = %v", err)
			}

			got := string(rep.Bytes(delims))
			if got != tt.want {
				t.Errorf("Bytes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRepetition_Bytes_CustomDelimiters(t *testing.T) {
	customDelims := &Delimiters{
		Field:        '|',
		Component:    '#', // Custom component delimiter
		Repetition:   '~',
		Escape:       '\\',
		SubComponent: '@', // Custom subcomponent delimiter
		Truncation:   '%',
	}

	// Parse with standard delimiters, encode with custom
	rep, err := ParseRepetition([]rune("A^B^C"), DefaultDelimiters())
	if err != nil {
		t.Fatalf("ParseRepetition() error = %v", err)
	}

	got := string(rep.Bytes(customDelims))
	want := "A#B#C"
	if got != want {
		t.Errorf("Bytes(customDelims) = %q, want %q", got, want)
	}
}

func TestRepetition_Bytes_NilDelimiters(t *testing.T) {
	rep, err := ParseRepetition([]rune("A^B"), DefaultDelimiters())
	if err != nil {
		t.Fatalf("ParseRepetition() error = %v", err)
	}

	// Bytes with nil should use defaults
	got := string(rep.Bytes(nil))
	want := "A^B"
	if got != want {
		t.Errorf("Bytes(nil) = %q, want %q", got, want)
	}
}

func TestRepetition_String(t *testing.T) {
	tests := []struct {
		name string
		data []rune
		want string
	}{
		{
			name: "simple value",
			data: []rune("Smith"),
			want: "Smith",
		},
		{
			name: "with components",
			data: []rune("Smith^John^Q"),
			want: "Smith^John^Q",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := ParseRepetition(tt.data, DefaultDelimiters())
			if err != nil {
				t.Fatalf("ParseRepetition() error = %v", err)
			}

			if got := rep.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRepetition_NestedStructure(t *testing.T) {
	// Test component -> subcomponent nesting
	// Format: comp1&subcomp1.1&subcomp1.2^comp2&subcomp2.1^comp3
	data := []rune("A&B&C^D&E^F")

	rep, err := ParseRepetition(data, DefaultDelimiters())
	if err != nil {
		t.Fatalf("ParseRepetition() error = %v", err)
	}

	// Should have 3 components
	comps := rep.Components()
	if len(comps) != 3 {
		t.Fatalf("Expected 3 components, got %d", len(comps))
	}

	// First component should have 3 subcomponents
	comp1 := comps[0]
	subComps1 := comp1.SubComponents()
	if len(subComps1) != 3 {
		t.Errorf("Component 1 should have 3 subcomponents, got %d", len(subComps1))
	}

	// Verify first component's subcomponent values
	expectedSubComps1 := []string{"A", "B", "C"}
	for i, want := range expectedSubComps1 {
		sc, ok := comp1.SubComponent(i + 1)
		if !ok {
			t.Errorf("SubComponent(%d) not found", i+1)
			continue
		}
		if got := sc.Value(); got != want {
			t.Errorf("SubComponent(%d).Value() = %q, want %q", i+1, got, want)
		}
	}

	// Second component should have 2 subcomponents
	comp2 := comps[1]
	subComps2 := comp2.SubComponents()
	if len(subComps2) != 2 {
		t.Errorf("Component 2 should have 2 subcomponents, got %d", len(subComps2))
	}

	// Third component should have no subcomponents (single value)
	comp3 := comps[2]
	subComps3 := comp3.SubComponents()
	if len(subComps3) != 0 {
		t.Errorf("Component 3 should have 0 subcomponents, got %d", len(subComps3))
	}
	if got := comp3.Value(); got != "F" {
		t.Errorf("Component 3 value = %q, want %q", got, "F")
	}

	// Test full roundtrip encoding
	got := string(rep.Bytes(DefaultDelimiters()))
	want := "A&B&C^D&E^F"
	if got != want {
		t.Errorf("Bytes() = %q, want %q", got, want)
	}
}

func TestRepetition_ComponentsReturnsCopy(t *testing.T) {
	rep, err := ParseRepetition([]rune("A^B^C"), DefaultDelimiters())
	if err != nil {
		t.Fatalf("ParseRepetition() error = %v", err)
	}

	// Get components twice
	comps1 := rep.Components()
	comps2 := rep.Components()

	// Modify the first slice
	if len(comps1) > 0 {
		comps1[0] = NewComponent("Modified")
	}

	// Second slice should be unchanged
	if comps2[0].Value() == "Modified" {
		t.Errorf("Components() should return a copy, not the internal slice")
	}
}

func TestRepetition_RoundTrip(t *testing.T) {
	// Test that parse -> encode produces the same result
	testCases := []string{
		"SimpleValue",
		"A^B",
		"A^B^C",
		"A&X&Y^B&Z^C",
		"^^",
		"A^^C",
		"^B^",
		"",
	}

	delims := DefaultDelimiters()

	for _, original := range testCases {
		t.Run(original, func(t *testing.T) {
			rep, err := ParseRepetition([]rune(original), delims)
			if err != nil {
				t.Fatalf("ParseRepetition(%q) error = %v", original, err)
			}

			encoded := string(rep.Bytes(delims))
			if encoded != original {
				t.Errorf("RoundTrip: original=%q, encoded=%q", original, encoded)
			}
		})
	}
}

func TestRepetition_ComplexHL7Name(t *testing.T) {
	// Test a realistic HL7 name field format
	// Format: LastName^FirstName^MiddleName^Suffix^Prefix^Degree
	data := []rune("Smith^John^Robert^III^Dr^MD")

	rep, err := ParseRepetition(data, DefaultDelimiters())
	if err != nil {
		t.Fatalf("ParseRepetition() error = %v", err)
	}

	expected := map[int]string{
		1: "Smith",  // Last name
		2: "John",   // First name
		3: "Robert", // Middle name
		4: "III",    // Suffix
		5: "Dr",     // Prefix
		6: "MD",     // Degree
	}

	for idx, want := range expected {
		comp, ok := rep.Component(idx)
		if !ok {
			t.Errorf("Component(%d) not found", idx)
			continue
		}
		if got := comp.Value(); got != want {
			t.Errorf("Component(%d) = %q, want %q", idx, got, want)
		}
	}
}
