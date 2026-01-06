// Package hl7 provides core types and utilities for HL7 v2.x message processing.
package hl7

import (
	"errors"
	"testing"
)

func TestNewField(t *testing.T) {
	tests := []struct {
		name    string
		seq     int
		value   string
		wantSeq int
		wantVal string
	}{
		{
			name:    "simple value",
			seq:     1,
			value:   "TestValue",
			wantSeq: 1,
			wantVal: "TestValue",
		},
		{
			name:    "empty value",
			seq:     5,
			value:   "",
			wantSeq: 5,
			wantVal: "",
		},
		{
			name:    "value with special chars",
			seq:     3,
			value:   "Test^Value~With|Delims",
			wantSeq: 3,
			wantVal: "Test^Value~With|Delims", // Not parsed, stored as-is
		},
		{
			name:    "unicode value",
			seq:     2,
			value:   "Patient Name: Jose Garcia",
			wantSeq: 2,
			wantVal: "Patient Name: Jose Garcia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewField(tt.seq, tt.value)

			if f.SeqNum() != tt.wantSeq {
				t.Errorf("SeqNum() = %v, want %v", f.SeqNum(), tt.wantSeq)
			}

			if f.Value() != tt.wantVal {
				t.Errorf("Value() = %v, want %v", f.Value(), tt.wantVal)
			}
		})
	}
}

func TestParseField_SimpleValue(t *testing.T) {
	tests := []struct {
		name    string
		seq     int
		data    string
		want    string
		wantErr bool
	}{
		{
			name: "simple string",
			seq:  1,
			data: "SimpleValue",
			want: "SimpleValue",
		},
		{
			name: "empty string",
			seq:  2,
			data: "",
			want: "",
		},
		{
			name: "numeric value",
			seq:  3,
			data: "12345",
			want: "12345",
		},
		{
			name: "value with spaces",
			seq:  4,
			data: "Hello World",
			want: "Hello World",
		},
	}

	delims := DefaultDelimiters()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ParseField(tt.seq, []rune(tt.data), delims)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && f.Value() != tt.want {
				t.Errorf("Value() = %v, want %v", f.Value(), tt.want)
			}

			if err == nil && f.SeqNum() != tt.seq {
				t.Errorf("SeqNum() = %v, want %v", f.SeqNum(), tt.seq)
			}
		})
	}
}

func TestParseField_WithRepetitions(t *testing.T) {
	tests := []struct {
		name       string
		seq        int
		data       string
		wantRepCnt int
		wantFirst  string // Expected Value() of first repetition (full encoded)
		wantSecond string // Expected Value() of second repetition (full encoded)
		wantValue  string // Expected Value() of field (full encoded with all repetitions)
	}{
		{
			name:       "two repetitions",
			seq:        1,
			data:       "First~Second",
			wantRepCnt: 2,
			wantFirst:  "First",
			wantSecond: "Second",
			wantValue:  "First~Second",
		},
		{
			name:       "three repetitions",
			seq:        2,
			data:       "A~B~C",
			wantRepCnt: 3,
			wantFirst:  "A",
			wantSecond: "B",
			wantValue:  "A~B~C",
		},
		{
			name:       "repetitions with empty",
			seq:        3,
			data:       "First~~Third",
			wantRepCnt: 3,
			wantFirst:  "First",
			wantSecond: "",
			wantValue:  "First~~Third",
		},
		{
			name:       "repetitions with components",
			seq:        4,
			data:       "Smith^John~Doe^Jane",
			wantRepCnt: 2,
			wantFirst:  "Smith^John",
			wantSecond: "Doe^Jane",
			wantValue:  "Smith^John~Doe^Jane",
		},
	}

	delims := DefaultDelimiters()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ParseField(tt.seq, []rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseField() error = %v", err)
			}

			if f.RepetitionCount() != tt.wantRepCnt {
				t.Errorf("RepetitionCount() = %v, want %v", f.RepetitionCount(), tt.wantRepCnt)
			}

			// Check first repetition
			rep0, ok := f.Repetition(0)
			if !ok {
				t.Fatal("Repetition(0) returned false")
			}
			if rep0.Value() != tt.wantFirst {
				t.Errorf("Repetition(0).Value() = %v, want %v", rep0.Value(), tt.wantFirst)
			}

			// Check second repetition
			rep1, ok := f.Repetition(1)
			if !ok {
				t.Fatal("Repetition(1) returned false")
			}
			if rep1.Value() != tt.wantSecond {
				t.Errorf("Repetition(1).Value() = %v, want %v", rep1.Value(), tt.wantSecond)
			}

			// Check that Value() returns full encoded field value
			if f.Value() != tt.wantValue {
				t.Errorf("Value() = %v, want %v", f.Value(), tt.wantValue)
			}
		})
	}
}

func TestParseField_WithComponents(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		compIdx    int
		wantComp   string
		wantCompOK bool
	}{
		{
			name:       "simple components",
			data:       "Smith^John^Q",
			compIdx:    1,
			wantComp:   "Smith",
			wantCompOK: true,
		},
		{
			name:       "get second component",
			data:       "Smith^John^Q",
			compIdx:    2,
			wantComp:   "John",
			wantCompOK: true,
		},
		{
			name:       "get third component",
			data:       "Smith^John^Q",
			compIdx:    3,
			wantComp:   "Q",
			wantCompOK: true,
		},
		{
			name:       "component out of range",
			data:       "Smith^John",
			compIdx:    5,
			wantComp:   "",
			wantCompOK: false,
		},
		{
			name:       "zero index invalid",
			data:       "Smith^John",
			compIdx:    0,
			wantComp:   "",
			wantCompOK: false,
		},
		{
			name:       "negative index invalid",
			data:       "Smith^John",
			compIdx:    -1,
			wantComp:   "",
			wantCompOK: false,
		},
	}

	delims := DefaultDelimiters()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ParseField(1, []rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseField() error = %v", err)
			}

			comp, ok := f.Component(tt.compIdx)
			if ok != tt.wantCompOK {
				t.Errorf("Component(%d) ok = %v, want %v", tt.compIdx, ok, tt.wantCompOK)
			}

			if ok && comp.Value() != tt.wantComp {
				t.Errorf("Component(%d).Value() = %v, want %v", tt.compIdx, comp.Value(), tt.wantComp)
			}
		})
	}
}

func TestField_Components(t *testing.T) {
	delims := DefaultDelimiters()

	t.Run("field with components", func(t *testing.T) {
		f, err := ParseField(1, []rune("A^B^C"), delims)
		if err != nil {
			t.Fatalf("ParseField() error = %v", err)
		}

		comps := f.Components()
		if len(comps) != 3 {
			t.Errorf("Components() len = %v, want 3", len(comps))
		}

		expected := []string{"A", "B", "C"}
		for i, comp := range comps {
			if comp.Value() != expected[i] {
				t.Errorf("Components()[%d].Value() = %v, want %v", i, comp.Value(), expected[i])
			}
		}
	})

	t.Run("field without components", func(t *testing.T) {
		f, err := ParseField(1, []rune("SimpleValue"), delims)
		if err != nil {
			t.Fatalf("ParseField() error = %v", err)
		}

		comps := f.Components()
		if len(comps) != 0 {
			t.Errorf("Components() len = %v, want 0 for non-component field", len(comps))
		}
	})

	t.Run("empty field", func(t *testing.T) {
		f, err := ParseField(1, []rune(""), delims)
		if err != nil {
			t.Fatalf("ParseField() error = %v", err)
		}

		comps := f.Components()
		if len(comps) != 0 {
			t.Errorf("Components() len = %v, want 0", len(comps))
		}
	})
}

func TestField_Repetitions(t *testing.T) {
	delims := DefaultDelimiters()

	t.Run("field with repetitions", func(t *testing.T) {
		f, err := ParseField(1, []rune("A~B~C"), delims)
		if err != nil {
			t.Fatalf("ParseField() error = %v", err)
		}

		reps := f.Repetitions()
		if len(reps) != 3 {
			t.Errorf("Repetitions() len = %v, want 3", len(reps))
		}

		expected := []string{"A", "B", "C"}
		for i, rep := range reps {
			if rep.Value() != expected[i] {
				t.Errorf("Repetitions()[%d].Value() = %v, want %v", i, rep.Value(), expected[i])
			}
		}
	})

	t.Run("field without explicit repetitions", func(t *testing.T) {
		f, err := ParseField(1, []rune("SingleValue"), delims)
		if err != nil {
			t.Fatalf("ParseField() error = %v", err)
		}

		reps := f.Repetitions()
		if len(reps) != 1 {
			t.Errorf("Repetitions() len = %v, want 1", len(reps))
		}

		if reps[0].Value() != "SingleValue" {
			t.Errorf("Repetitions()[0].Value() = %v, want SingleValue", reps[0].Value())
		}
	})

	t.Run("repetitions slice is a copy", func(t *testing.T) {
		f, err := ParseField(1, []rune("A~B"), delims)
		if err != nil {
			t.Fatalf("ParseField() error = %v", err)
		}

		reps1 := f.Repetitions()
		reps2 := f.Repetitions()

		// Modifying one slice shouldn't affect the other
		if &reps1[0] == &reps2[0] {
			t.Error("Repetitions() should return a copy, not the original slice")
		}
	})
}

func TestField_Repetition(t *testing.T) {
	delims := DefaultDelimiters()

	f, err := ParseField(1, []rune("First~Second~Third"), delims)
	if err != nil {
		t.Fatalf("ParseField() error = %v", err)
	}

	tests := []struct {
		name   string
		index  int
		want   string
		wantOK bool
	}{
		{"index 0", 0, "First", true},
		{"index 1", 1, "Second", true},
		{"index 2", 2, "Third", true},
		{"index out of range", 3, "", false},
		{"negative index", -1, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, ok := f.Repetition(tt.index)
			if ok != tt.wantOK {
				t.Errorf("Repetition(%d) ok = %v, want %v", tt.index, ok, tt.wantOK)
			}

			if ok && rep.Value() != tt.want {
				t.Errorf("Repetition(%d).Value() = %v, want %v", tt.index, rep.Value(), tt.want)
			}
		})
	}
}

func TestField_Get(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name     string
		data     string
		location string
		want     string
		wantErr  bool
	}{
		{
			name:     "empty location returns value",
			data:     "TestValue",
			location: "",
			want:     "TestValue",
		},
		{
			name:     "get component 1",
			data:     "Smith^John^Q",
			location: ".1",
			want:     "Smith",
		},
		{
			name:     "get component 2",
			data:     "Smith^John^Q",
			location: ".2",
			want:     "John",
		},
		{
			name:     "get component 3",
			data:     "Smith^John^Q",
			location: ".3",
			want:     "Q",
		},
		{
			name:     "get subcomponent",
			data:     "Smith&Jr^John",
			location: ".1.2",
			want:     "Jr",
		},
		{
			name:     "get repetition by index",
			data:     "First~Second~Third",
			location: "[1]",
			want:     "Second",
		},
		{
			name:     "get repetition 0",
			data:     "First~Second",
			location: "[0]",
			want:     "First",
		},
		{
			name:     "get repetition component",
			data:     "Smith^John~Doe^Jane",
			location: "[1].1",
			want:     "Doe",
		},
		{
			name:     "get repetition component 2",
			data:     "Smith^John~Doe^Jane",
			location: "[1].2",
			want:     "Jane",
		},
		{
			name:     "component not found returns empty",
			data:     "Smith^John",
			location: ".5",
			want:     "",
		},
		{
			name:     "repetition not found returns empty",
			data:     "First~Second",
			location: "[5]",
			want:     "",
		},
		{
			name:     "invalid location format",
			data:     "TestValue",
			location: "[abc]",
			wantErr:  true,
		},
		{
			name:     "invalid component format",
			data:     "TestValue",
			location: ".abc",
			wantErr:  true,
		},
		{
			name:     "unclosed bracket",
			data:     "TestValue",
			location: "[1",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ParseField(1, []rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseField() error = %v", err)
			}

			got, err := f.Get(tt.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get(%q) error = %v, wantErr %v", tt.location, err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("Get(%q) = %v, want %v", tt.location, got, tt.want)
			}
		})
	}
}

func TestField_Set(t *testing.T) {
	delims := DefaultDelimiters()

	t.Run("set empty location replaces entire field", func(t *testing.T) {
		f, _ := ParseField(1, []rune("OldValue"), delims)

		err := f.Set("", "NewValue")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if f.Value() != "NewValue" {
			t.Errorf("Value() = %v, want NewValue", f.Value())
		}
	})

	t.Run("set component", func(t *testing.T) {
		f, _ := ParseField(1, []rune("Smith^John"), delims)

		err := f.Set(".2", "Jane")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		got, _ := f.Get(".2")
		if got != "Jane" {
			t.Errorf("Get(.2) = %v, want Jane", got)
		}
	})

	t.Run("set new component extends field", func(t *testing.T) {
		f, _ := ParseField(1, []rune("A^B"), delims)

		err := f.Set(".4", "D")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		got, _ := f.Get(".4")
		if got != "D" {
			t.Errorf("Get(.4) = %v, want D", got)
		}

		// Intermediate component should be empty
		got, _ = f.Get(".3")
		if got != "" {
			t.Errorf("Get(.3) = %v, want empty", got)
		}
	})

	t.Run("set subcomponent", func(t *testing.T) {
		f, _ := ParseField(1, []rune("Smith&Jr^John"), delims)

		err := f.Set(".1.2", "III")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		got, _ := f.Get(".1.2")
		if got != "III" {
			t.Errorf("Get(.1.2) = %v, want III", got)
		}
	})

	t.Run("set repetition value", func(t *testing.T) {
		f, _ := ParseField(1, []rune("First~Second"), delims)

		err := f.Set("[1]", "NewSecond")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		got, _ := f.Get("[1]")
		if got != "NewSecond" {
			t.Errorf("Get([1]) = %v, want NewSecond", got)
		}
	})

	t.Run("set new repetition extends field", func(t *testing.T) {
		f, _ := ParseField(1, []rune("First"), delims)

		err := f.Set("[2]", "Third")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if f.RepetitionCount() < 3 {
			t.Errorf("RepetitionCount() = %v, want >= 3", f.RepetitionCount())
		}

		got, _ := f.Get("[2]")
		if got != "Third" {
			t.Errorf("Get([2]) = %v, want Third", got)
		}
	})

	t.Run("set repetition component", func(t *testing.T) {
		f, _ := ParseField(1, []rune("Smith^John~Doe^Jane"), delims)

		err := f.Set("[1].2", "Mary")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		got, _ := f.Get("[1].2")
		if got != "Mary" {
			t.Errorf("Get([1].2) = %v, want Mary", got)
		}
	})

	t.Run("invalid location returns error", func(t *testing.T) {
		f, _ := ParseField(1, []rune("Test"), delims)

		err := f.Set("[abc]", "value")
		if err == nil {
			t.Error("Set() with invalid location should return error")
		}

		var locErr *LocationError
		if !errors.As(err, &locErr) {
			t.Errorf("error should be LocationError, got %T", err)
		}
	})
}

func TestField_Bytes(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "simple value",
			data: "SimpleValue",
			want: "SimpleValue",
		},
		{
			name: "value with components",
			data: "Smith^John^Q",
			want: "Smith^John^Q",
		},
		{
			name: "value with repetitions",
			data: "First~Second~Third",
			want: "First~Second~Third",
		},
		{
			name: "complex field",
			data: "Smith^John&Jr~Doe^Jane",
			want: "Smith^John&Jr~Doe^Jane",
		},
		{
			name: "empty value",
			data: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ParseField(1, []rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseField() error = %v", err)
			}

			got := string(f.Bytes(delims))
			if got != tt.want {
				t.Errorf("Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestField_Bytes_CustomDelimiters(t *testing.T) {
	customDelims := &Delimiters{
		Field:        '|',
		Component:    '#',
		Repetition:   '*',
		Escape:       '\\',
		SubComponent: '@',
	}

	// Parse with custom delimiters
	f, err := ParseField(1, []rune("A#B*C#D"), customDelims)
	if err != nil {
		t.Fatalf("ParseField() error = %v", err)
	}

	got := string(f.Bytes(customDelims))
	want := "A#B*C#D"
	if got != want {
		t.Errorf("Bytes() = %v, want %v", got, want)
	}
}

func TestField_String(t *testing.T) {
	delims := DefaultDelimiters()

	f, _ := ParseField(1, []rune("Smith^John~Doe^Jane"), delims)

	got := f.String()
	want := "Smith^John~Doe^Jane"

	if got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestField_SeqNum(t *testing.T) {
	tests := []struct {
		name string
		seq  int
	}{
		{"seq 1", 1},
		{"seq 5", 5},
		{"seq 0", 0},
		{"seq 100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewField(tt.seq, "test")
			if f.SeqNum() != tt.seq {
				t.Errorf("SeqNum() = %v, want %v", f.SeqNum(), tt.seq)
			}
		})
	}
}

func TestParseField_NilDelimiters(t *testing.T) {
	// Should use default delimiters when nil is passed
	f, err := ParseField(1, []rune("A^B~C"), nil)
	if err != nil {
		t.Fatalf("ParseField() error = %v", err)
	}

	if f.RepetitionCount() != 2 {
		t.Errorf("RepetitionCount() = %v, want 2", f.RepetitionCount())
	}

	comp, ok := f.Component(2)
	if !ok {
		t.Fatal("Component(2) returned false")
	}
	if comp.Value() != "B" {
		t.Errorf("Component(2).Value() = %v, want B", comp.Value())
	}
}

func TestField_Bytes_NilDelimiters(t *testing.T) {
	f, _ := ParseField(1, []rune("A^B~C"), nil)

	// Should use default delimiters when nil is passed
	got := string(f.Bytes(nil))
	want := "A^B~C"
	if got != want {
		t.Errorf("Bytes(nil) = %v, want %v", got, want)
	}
}

func TestParseFieldLocation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantRep     int
		wantComp    int
		wantSubComp int
		wantErr     bool
	}{
		{
			name:        "empty string",
			input:       "",
			wantRep:     -1,
			wantComp:    0,
			wantSubComp: 0,
		},
		{
			name:        "component only",
			input:       ".1",
			wantRep:     -1,
			wantComp:    1,
			wantSubComp: 0,
		},
		{
			name:        "component and subcomponent",
			input:       ".1.2",
			wantRep:     -1,
			wantComp:    1,
			wantSubComp: 2,
		},
		{
			name:        "repetition only",
			input:       "[0]",
			wantRep:     0,
			wantComp:    0,
			wantSubComp: 0,
		},
		{
			name:        "repetition and component",
			input:       "[1].2",
			wantRep:     1,
			wantComp:    2,
			wantSubComp: 0,
		},
		{
			name:        "full location",
			input:       "[2].3.4",
			wantRep:     2,
			wantComp:    3,
			wantSubComp: 4,
		},
		{
			name:    "invalid repetition",
			input:   "[abc]",
			wantErr: true,
		},
		{
			name:    "negative repetition",
			input:   "[-1]",
			wantErr: true,
		},
		{
			name:    "missing closing bracket",
			input:   "[1.2",
			wantErr: true,
		},
		{
			name:    "invalid component",
			input:   ".abc",
			wantErr: true,
		},
		{
			name:    "negative component",
			input:   ".-1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := parseFieldLocation(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseFieldLocation(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if loc.Repetition != tt.wantRep {
				t.Errorf("Repetition = %v, want %v", loc.Repetition, tt.wantRep)
			}
			if loc.Component != tt.wantComp {
				t.Errorf("Component = %v, want %v", loc.Component, tt.wantComp)
			}
			if loc.SubComponent != tt.wantSubComp {
				t.Errorf("SubComponent = %v, want %v", loc.SubComponent, tt.wantSubComp)
			}
		})
	}
}

func TestField_RoundTrip(t *testing.T) {
	delims := DefaultDelimiters()

	testCases := []string{
		"Simple",
		"Smith^John^Q",
		"First~Second~Third",
		"Smith^John&Jr~Doe^Jane&Sr",
		"A^B^C~D^E^F~G^H^I",
		"",
	}

	for _, original := range testCases {
		t.Run(original, func(t *testing.T) {
			// Parse
			f, err := ParseField(1, []rune(original), delims)
			if err != nil {
				t.Fatalf("ParseField() error = %v", err)
			}

			// Encode
			encoded := string(f.Bytes(delims))

			// Should match original
			if encoded != original {
				t.Errorf("Round trip failed: got %q, want %q", encoded, original)
			}
		})
	}
}

func TestField_ComplexHierarchy(t *testing.T) {
	delims := DefaultDelimiters()

	// Complex field with repetitions, components, and subcomponents
	data := "Smith&Jr^John^Q~Doe&Sr^Jane^M"
	f, err := ParseField(5, []rune(data), delims)
	if err != nil {
		t.Fatalf("ParseField() error = %v", err)
	}

	// Verify sequence number
	if f.SeqNum() != 5 {
		t.Errorf("SeqNum() = %v, want 5", f.SeqNum())
	}

	// Verify repetition count
	if f.RepetitionCount() != 2 {
		t.Errorf("RepetitionCount() = %v, want 2", f.RepetitionCount())
	}

	// Verify first repetition
	val, _ := f.Get("[0].1.1")
	if val != "Smith" {
		t.Errorf("Get([0].1.1) = %v, want Smith", val)
	}

	val, _ = f.Get("[0].1.2")
	if val != "Jr" {
		t.Errorf("Get([0].1.2) = %v, want Jr", val)
	}

	val, _ = f.Get("[0].2")
	if val != "John" {
		t.Errorf("Get([0].2) = %v, want John", val)
	}

	// Verify second repetition
	val, _ = f.Get("[1].1.1")
	if val != "Doe" {
		t.Errorf("Get([1].1.1) = %v, want Doe", val)
	}

	val, _ = f.Get("[1].2")
	if val != "Jane" {
		t.Errorf("Get([1].2) = %v, want Jane", val)
	}
}

func TestField_EmptyRepetitions(t *testing.T) {
	delims := DefaultDelimiters()

	// Field with empty repetitions in the middle
	data := "A~~C"
	f, err := ParseField(1, []rune(data), delims)
	if err != nil {
		t.Fatalf("ParseField() error = %v", err)
	}

	if f.RepetitionCount() != 3 {
		t.Errorf("RepetitionCount() = %v, want 3", f.RepetitionCount())
	}

	rep0, _ := f.Repetition(0)
	if rep0.Value() != "A" {
		t.Errorf("Repetition(0).Value() = %v, want A", rep0.Value())
	}

	rep1, _ := f.Repetition(1)
	if rep1.Value() != "" {
		t.Errorf("Repetition(1).Value() = %v, want empty", rep1.Value())
	}

	rep2, _ := f.Repetition(2)
	if rep2.Value() != "C" {
		t.Errorf("Repetition(2).Value() = %v, want C", rep2.Value())
	}
}
