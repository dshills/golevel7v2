package hl7

import (
	"errors"
	"testing"
)

func TestParseLocation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Location
		wantErr error
	}{
		// Basic segment only
		{
			name:  "segment only - PID",
			input: "PID",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment only - MSH",
			input: "MSH",
			want: &Location{
				Segment:      "MSH",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment only - OBX",
			input: "OBX",
			want: &Location{
				Segment:      "OBX",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment only - ZZ1 (custom segment)",
			input: "ZZ1",
			want: &Location{
				Segment:      "ZZ1",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},

		// Segment with index
		{
			name:  "segment with index 0",
			input: "PID[0]",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: 0,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment with index 1",
			input: "OBX[1]",
			want: &Location{
				Segment:      "OBX",
				SegmentIndex: 1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment with large index",
			input: "OBX[99]",
			want: &Location{
				Segment:      "OBX",
				SegmentIndex: 99,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},

		// Segment + field
		{
			name:  "segment and field",
			input: "PID.5",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment and field 0",
			input: "MSH.0",
			want: &Location{
				Segment:      "MSH",
				SegmentIndex: -1,
				Field:        0,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment index and field",
			input: "PID[1].5",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: 1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},

		// Segment + field + repetition
		{
			name:  "segment field repetition",
			input: "PID.5[0]",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   0,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment field repetition 2",
			input: "PID.11[2]",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        11,
				Repetition:   2,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment index field repetition",
			input: "OBX[3].5[1]",
			want: &Location{
				Segment:      "OBX",
				SegmentIndex: 3,
				Field:        5,
				Repetition:   1,
				Component:    -1,
				SubComponent: -1,
			},
		},

		// Segment + field + component
		{
			name:  "segment field component",
			input: "PID.5.1",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: -1,
			},
		},
		{
			name:  "segment field repetition component",
			input: "PID.5[0].1",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   0,
				Component:    1,
				SubComponent: -1,
			},
		},
		{
			name:  "full path without subcomponent",
			input: "PID[0].5[1].3",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: 0,
				Field:        5,
				Repetition:   1,
				Component:    3,
				SubComponent: -1,
			},
		},

		// Full path with subcomponent
		{
			name:  "full path",
			input: "PID.5.1.2",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: 2,
			},
		},
		{
			name:  "full path with all indices",
			input: "PID[2].5[3].1.2",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: 2,
				Field:        5,
				Repetition:   3,
				Component:    1,
				SubComponent: 2,
			},
		},
		{
			name:  "MSH full path",
			input: "MSH.9.1.1",
			want: &Location{
				Segment:      "MSH",
				SegmentIndex: -1,
				Field:        9,
				Repetition:   -1,
				Component:    1,
				SubComponent: 1,
			},
		},

		// Whitespace handling
		{
			name:  "leading whitespace",
			input: "  PID.5",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:  "trailing whitespace",
			input: "PID.5  ",
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},

		// Error cases
		{
			name:    "empty string",
			input:   "",
			wantErr: ErrEmptyLocation,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: ErrEmptyLocation,
		},
		{
			name:    "lowercase segment",
			input:   "pid",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "mixed case segment",
			input:   "Pid",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "two letter segment",
			input:   "PI",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "four letter segment",
			input:   "PIDD",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "segment starting with number",
			input:   "1ID",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "invalid characters in segment",
			input:   "PI_",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "field without dot",
			input:   "PID5",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "negative field",
			input:   "PID.-1",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "non-numeric field",
			input:   "PID.abc",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "empty brackets",
			input:   "PID[]",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "unclosed bracket",
			input:   "PID[0",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "double dot",
			input:   "PID..5",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "trailing dot",
			input:   "PID.5.",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "component without field",
			input:   "PID..1",
			wantErr: ErrInvalidFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLocation(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseLocation(%q) expected error containing %v, got nil", tt.input, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseLocation(%q) error = %v, want error containing %v", tt.input, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseLocation(%q) unexpected error: %v", tt.input, err)
				return
			}

			if !got.Equal(tt.want) {
				t.Errorf("ParseLocation(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLocation_String(t *testing.T) {
	tests := []struct {
		name string
		loc  *Location
		want string
	}{
		{
			name: "segment only",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: "PID",
		},
		{
			name: "segment with index",
			loc: &Location{
				Segment:      "OBX",
				SegmentIndex: 2,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: "OBX[2]",
		},
		{
			name: "segment and field",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: "PID.5",
		},
		{
			name: "segment index and field",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: 1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: "PID[1].5",
		},
		{
			name: "segment field repetition",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   0,
				Component:    -1,
				SubComponent: -1,
			},
			want: "PID.5[0]",
		},
		{
			name: "segment field component",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: -1,
			},
			want: "PID.5.1",
		},
		{
			name: "segment field repetition component",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   0,
				Component:    1,
				SubComponent: -1,
			},
			want: "PID.5[0].1",
		},
		{
			name: "full path",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: 2,
			},
			want: "PID.5.1.2",
		},
		{
			name: "full path with all indices",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: 2,
				Field:        5,
				Repetition:   3,
				Component:    1,
				SubComponent: 2,
			},
			want: "PID[2].5[3].1.2",
		},
		{
			name: "nil location",
			loc:  nil,
			want: "",
		},
		{
			name: "field 0",
			loc: &Location{
				Segment:      "MSH",
				SegmentIndex: -1,
				Field:        0,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: "MSH.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.loc.String()
			if got != tt.want {
				t.Errorf("Location.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLocation_RoundTrip(t *testing.T) {
	// Test that parsing and stringifying produces the same result
	inputs := []string{
		"PID",
		"MSH",
		"OBX[0]",
		"OBX[5]",
		"PID.5",
		"PID[1].5",
		"PID.5[0]",
		"PID.5.1",
		"PID.5[0].1",
		"PID.5.1.2",
		"PID[2].5[3].1.2",
		"MSH.9.1.1",
		"ZZ1.1.1.1",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			loc, err := ParseLocation(input)
			if err != nil {
				t.Fatalf("ParseLocation(%q) failed: %v", input, err)
			}

			output := loc.String()
			if output != input {
				t.Errorf("Round trip failed: input=%q, output=%q", input, output)
			}

			// Parse again to verify
			loc2, err := ParseLocation(output)
			if err != nil {
				t.Fatalf("ParseLocation(%q) second pass failed: %v", output, err)
			}

			if !loc.Equal(loc2) {
				t.Errorf("Double round trip locations not equal: %+v vs %+v", loc, loc2)
			}
		})
	}
}

func TestNewLocation(t *testing.T) {
	tests := []struct {
		name         string
		segment      string
		field        int
		component    int
		subcomponent int
		want         *Location
	}{
		{
			name:         "segment only",
			segment:      "PID",
			field:        -1,
			component:    -1,
			subcomponent: -1,
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:         "segment and field",
			segment:      "PID",
			field:        5,
			component:    -1,
			subcomponent: -1,
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name:         "full path",
			segment:      "PID",
			field:        5,
			component:    1,
			subcomponent: 2,
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: 2,
			},
		},
		{
			name:         "lowercase segment converted",
			segment:      "pid",
			field:        5,
			component:    -1,
			subcomponent: -1,
			want: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLocation(tt.segment, tt.field, tt.component, tt.subcomponent)
			if !got.Equal(tt.want) {
				t.Errorf("NewLocation() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestNewLocationFull(t *testing.T) {
	got := NewLocationFull("PID", 2, 5, 3, 1, 2)
	want := &Location{
		Segment:      "PID",
		SegmentIndex: 2,
		Field:        5,
		Repetition:   3,
		Component:    1,
		SubComponent: 2,
	}

	if !got.Equal(want) {
		t.Errorf("NewLocationFull() = %+v, want %+v", got, want)
	}
}

func TestLocation_IsValid(t *testing.T) {
	tests := []struct {
		name string
		loc  *Location
		want bool
	}{
		{
			name: "valid segment only",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: true,
		},
		{
			name: "valid full path",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: 0,
				Field:        5,
				Repetition:   0,
				Component:    1,
				SubComponent: 2,
			},
			want: true,
		},
		{
			name: "valid with alphanumeric segment",
			loc: &Location{
				Segment:      "ZZ1",
				SegmentIndex: -1,
				Field:        1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: true,
		},
		{
			name: "nil location",
			loc:  nil,
			want: false,
		},
		{
			name: "empty segment",
			loc: &Location{
				Segment:      "",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "lowercase segment",
			loc: &Location{
				Segment:      "pid",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "two letter segment",
			loc: &Location{
				Segment:      "PI",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "four letter segment",
			loc: &Location{
				Segment:      "PIDD",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "segment starting with number",
			loc: &Location{
				Segment:      "1ID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "invalid segment index",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -2,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "invalid field",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -2,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "invalid repetition",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -2,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "invalid component",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -2,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "invalid subcomponent",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: -2,
			},
			want: false,
		},
		{
			name: "component without field",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "subcomponent without component",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: 1,
			},
			want: false,
		},
		{
			name: "repetition without field",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   0,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.loc.IsValid()
			if got != tt.want {
				t.Errorf("Location.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocation_HasMethods(t *testing.T) {
	tests := []struct {
		name            string
		loc             *Location
		hasSegment      bool
		hasField        bool
		hasComponent    bool
		hasSubComponent bool
		hasSegmentIndex bool
		hasRepetition   bool
	}{
		{
			name: "segment only",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			hasSegment:      true,
			hasField:        false,
			hasComponent:    false,
			hasSubComponent: false,
			hasSegmentIndex: false,
			hasRepetition:   false,
		},
		{
			name: "full path",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: 0,
				Field:        5,
				Repetition:   0,
				Component:    1,
				SubComponent: 2,
			},
			hasSegment:      true,
			hasField:        true,
			hasComponent:    true,
			hasSubComponent: true,
			hasSegmentIndex: true,
			hasRepetition:   true,
		},
		{
			name: "segment and field only",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			hasSegment:      true,
			hasField:        true,
			hasComponent:    false,
			hasSubComponent: false,
			hasSegmentIndex: false,
			hasRepetition:   false,
		},
		{
			name:            "nil location",
			loc:             nil,
			hasSegment:      false,
			hasField:        false,
			hasComponent:    false,
			hasSubComponent: false,
			hasSegmentIndex: false,
			hasRepetition:   false,
		},
		{
			name: "empty segment",
			loc: &Location{
				Segment:      "",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			hasSegment:      false,
			hasField:        false,
			hasComponent:    false,
			hasSubComponent: false,
			hasSegmentIndex: false,
			hasRepetition:   false,
		},
		{
			name: "field 0 is valid",
			loc: &Location{
				Segment:      "MSH",
				SegmentIndex: -1,
				Field:        0,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			hasSegment:      true,
			hasField:        true,
			hasComponent:    false,
			hasSubComponent: false,
			hasSegmentIndex: false,
			hasRepetition:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.loc.HasSegment(); got != tt.hasSegment {
				t.Errorf("HasSegment() = %v, want %v", got, tt.hasSegment)
			}
			if got := tt.loc.HasField(); got != tt.hasField {
				t.Errorf("HasField() = %v, want %v", got, tt.hasField)
			}
			if got := tt.loc.HasComponent(); got != tt.hasComponent {
				t.Errorf("HasComponent() = %v, want %v", got, tt.hasComponent)
			}
			if got := tt.loc.HasSubComponent(); got != tt.hasSubComponent {
				t.Errorf("HasSubComponent() = %v, want %v", got, tt.hasSubComponent)
			}
			if got := tt.loc.HasSegmentIndex(); got != tt.hasSegmentIndex {
				t.Errorf("HasSegmentIndex() = %v, want %v", got, tt.hasSegmentIndex)
			}
			if got := tt.loc.HasRepetition(); got != tt.hasRepetition {
				t.Errorf("HasRepetition() = %v, want %v", got, tt.hasRepetition)
			}
		})
	}
}

func TestLocation_Equal(t *testing.T) {
	tests := []struct {
		name string
		loc1 *Location
		loc2 *Location
		want bool
	}{
		{
			name: "both nil",
			loc1: nil,
			loc2: nil,
			want: true,
		},
		{
			name: "first nil",
			loc1: nil,
			loc2: &Location{Segment: "PID"},
			want: false,
		},
		{
			name: "second nil",
			loc1: &Location{Segment: "PID"},
			loc2: nil,
			want: false,
		},
		{
			name: "equal simple",
			loc1: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			loc2: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: true,
		},
		{
			name: "different segment",
			loc1: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			loc2: &Location{
				Segment:      "OBX",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "different segment index",
			loc1: &Location{
				Segment:      "PID",
				SegmentIndex: 0,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			loc2: &Location{
				Segment:      "PID",
				SegmentIndex: 1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "different field",
			loc1: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			loc2: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        6,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "different repetition",
			loc1: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   0,
				Component:    -1,
				SubComponent: -1,
			},
			loc2: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   1,
				Component:    -1,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "different component",
			loc1: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: -1,
			},
			loc2: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    2,
				SubComponent: -1,
			},
			want: false,
		},
		{
			name: "different subcomponent",
			loc1: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: 1,
			},
			loc2: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: 2,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.loc1.Equal(tt.loc2)
			if got != tt.want {
				t.Errorf("Location.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocation_Clone(t *testing.T) {
	tests := []struct {
		name string
		loc  *Location
	}{
		{
			name: "nil location",
			loc:  nil,
		},
		{
			name: "simple location",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
		},
		{
			name: "full location",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: 2,
				Field:        5,
				Repetition:   3,
				Component:    1,
				SubComponent: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.loc.Clone()

			if tt.loc == nil {
				if clone != nil {
					t.Error("Clone of nil should be nil")
				}
				return
			}

			if clone == tt.loc {
				t.Error("Clone should return a new pointer")
			}

			if !clone.Equal(tt.loc) {
				t.Errorf("Clone() = %+v, want %+v", clone, tt.loc)
			}

			// Verify it's a deep copy by modifying clone
			clone.Field = 999
			if tt.loc.Field == 999 {
				t.Error("Modifying clone should not affect original")
			}
		})
	}
}

func TestLocation_Depth(t *testing.T) {
	tests := []struct {
		name string
		loc  *Location
		want int
	}{
		{
			name: "nil location",
			loc:  nil,
			want: -1,
		},
		{
			name: "empty segment",
			loc: &Location{
				Segment:      "",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: -1,
		},
		{
			name: "segment only",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        -1,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: 0,
		},
		{
			name: "field level",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    -1,
				SubComponent: -1,
			},
			want: 1,
		},
		{
			name: "component level",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: -1,
			},
			want: 2,
		},
		{
			name: "subcomponent level",
			loc: &Location{
				Segment:      "PID",
				SegmentIndex: -1,
				Field:        5,
				Repetition:   -1,
				Component:    1,
				SubComponent: 2,
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.loc.Depth()
			if got != tt.want {
				t.Errorf("Location.Depth() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseLocation(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"segment_only", "PID"},
		{"segment_field", "PID.5"},
		{"segment_field_component", "PID.5.1"},
		{"full_path", "PID.5.1.2"},
		{"full_path_with_indices", "PID[2].5[3].1.2"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ParseLocation(bm.input)
			}
		})
	}
}

func BenchmarkLocation_String(b *testing.B) {
	loc := &Location{
		Segment:      "PID",
		SegmentIndex: 2,
		Field:        5,
		Repetition:   3,
		Component:    1,
		SubComponent: 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loc.String()
	}
}

func BenchmarkLocation_IsValid(b *testing.B) {
	loc := &Location{
		Segment:      "PID",
		SegmentIndex: 2,
		Field:        5,
		Repetition:   3,
		Component:    1,
		SubComponent: 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loc.IsValid()
	}
}
