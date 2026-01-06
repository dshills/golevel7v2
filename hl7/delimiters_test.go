package hl7

import (
	"errors"
	"testing"
)

func TestDefaultDelimiters(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		want     rune
		wantDesc string
	}{
		{
			name:     "Field delimiter",
			field:    "Field",
			want:     '|',
			wantDesc: "pipe character",
		},
		{
			name:     "Component delimiter",
			field:    "Component",
			want:     '^',
			wantDesc: "caret character",
		},
		{
			name:     "Repetition delimiter",
			field:    "Repetition",
			want:     '~',
			wantDesc: "tilde character",
		},
		{
			name:     "Escape delimiter",
			field:    "Escape",
			want:     '\\',
			wantDesc: "backslash character",
		},
		{
			name:     "SubComponent delimiter",
			field:    "SubComponent",
			want:     '&',
			wantDesc: "ampersand character",
		},
		{
			name:     "Truncation delimiter",
			field:    "Truncation",
			want:     '#',
			wantDesc: "hash character",
		},
	}

	d := DefaultDelimiters()
	if d == nil {
		t.Fatal("DefaultDelimiters() returned nil")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got rune
			switch tt.field {
			case "Field":
				got = d.Field
			case "Component":
				got = d.Component
			case "Repetition":
				got = d.Repetition
			case "Escape":
				got = d.Escape
			case "SubComponent":
				got = d.SubComponent
			case "Truncation":
				got = d.Truncation
			}

			if got != tt.want {
				t.Errorf("DefaultDelimiters().%s = %q, want %q (%s)",
					tt.field, got, tt.want, tt.wantDesc)
			}
		})
	}
}

func TestParseDelimiters(t *testing.T) {
	tests := []struct {
		name       string
		mshSegment []byte
		want       *Delimiters
		wantErr    error
	}{
		{
			name:       "standard delimiters",
			mshSegment: []byte("MSH|^~\\&|SendingApp|SendingFac|"),
			want: &Delimiters{
				Field:        '|',
				Component:    '^',
				Repetition:   '~',
				Escape:       '\\',
				SubComponent: '&',
				Truncation:   '#', // defaults when not present
			},
			wantErr: nil,
		},
		{
			name:       "standard delimiters with truncation",
			mshSegment: []byte("MSH|^~\\&#|SendingApp|SendingFac|"),
			want: &Delimiters{
				Field:        '|',
				Component:    '^',
				Repetition:   '~',
				Escape:       '\\',
				SubComponent: '&',
				Truncation:   '#',
			},
			wantErr: nil,
		},
		{
			name:       "custom truncation character",
			mshSegment: []byte("MSH|^~\\&@|SendingApp|SendingFac|"),
			want: &Delimiters{
				Field:        '|',
				Component:    '^',
				Repetition:   '~',
				Escape:       '\\',
				SubComponent: '&',
				Truncation:   '@',
			},
			wantErr: nil,
		},
		{
			name:       "custom delimiters",
			mshSegment: []byte("MSH!@#$%^|SendingApp"),
			want: &Delimiters{
				Field:        '!',
				Component:    '@',
				Repetition:   '#',
				Escape:       '$',
				SubComponent: '%',
				Truncation:   '^',
			},
			wantErr: nil,
		},
		{
			name:       "minimum valid MSH",
			mshSegment: []byte("MSH|^~\\&"),
			want: &Delimiters{
				Field:        '|',
				Component:    '^',
				Repetition:   '~',
				Escape:       '\\',
				SubComponent: '&',
				Truncation:   '#',
			},
			wantErr: nil,
		},
		{
			name:       "MSH with carriage return",
			mshSegment: []byte("MSH|^~\\&|\rPID|"),
			want: &Delimiters{
				Field:        '|',
				Component:    '^',
				Repetition:   '~',
				Escape:       '\\',
				SubComponent: '&',
				Truncation:   '#',
			},
			wantErr: nil,
		},
		{
			name:       "empty input",
			mshSegment: []byte{},
			want:       nil,
			wantErr:    ErrEmptyInput,
		},
		{
			name:       "nil input",
			mshSegment: nil,
			want:       nil,
			wantErr:    ErrEmptyInput,
		},
		{
			name:       "not MSH segment",
			mshSegment: []byte("PID|1|12345"),
			want:       nil,
			wantErr:    ErrNotMSHSegment,
		},
		{
			name:       "lowercase msh",
			mshSegment: []byte("msh|^~\\&|"),
			want:       nil,
			wantErr:    ErrNotMSHSegment,
		},
		{
			name:       "too short - only MSH",
			mshSegment: []byte("MSH"),
			want:       nil,
			wantErr:    ErrMSHTooShort,
		},
		{
			name:       "too short - MSH plus field only",
			mshSegment: []byte("MSH|"),
			want:       nil,
			wantErr:    ErrMSHTooShort,
		},
		{
			name:       "too short - missing encoding chars",
			mshSegment: []byte("MSH|^~"),
			want:       nil,
			wantErr:    ErrMSHTooShort,
		},
		{
			name:       "MSH-2 too short - only 3 chars before field sep",
			mshSegment: []byte("MSH|^~\\|"),
			want:       nil,
			wantErr:    ErrMissingDelimiter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDelimiters(tt.mshSegment)

			// Check error
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseDelimiters() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseDelimiters() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDelimiters() unexpected error = %v", err)
				return
			}

			// Check result
			if got == nil {
				t.Error("ParseDelimiters() returned nil without error")
				return
			}

			if !got.Equal(tt.want) {
				t.Errorf("ParseDelimiters() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDelimiters_String(t *testing.T) {
	tests := []struct {
		name       string
		delimiters *Delimiters
		want       string
	}{
		{
			name:       "default delimiters",
			delimiters: DefaultDelimiters(),
			want:       "^~\\&#",
		},
		{
			name: "custom delimiters",
			delimiters: &Delimiters{
				Field:        '!',
				Component:    '@',
				Repetition:   '#',
				Escape:       '$',
				SubComponent: '%',
				Truncation:   '^',
			},
			want: "@#$%^",
		},
		{
			name: "no truncation explicit",
			delimiters: &Delimiters{
				Field:        '|',
				Component:    '^',
				Repetition:   '~',
				Escape:       '\\',
				SubComponent: '&',
				Truncation:   0, // null character
			},
			want: "^~\\&\x00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.delimiters.String()
			if got != tt.want {
				t.Errorf("Delimiters.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDelimiters_EncodingCharacters(t *testing.T) {
	d := DefaultDelimiters()

	// EncodingCharacters should return the same as String
	if d.EncodingCharacters() != d.String() {
		t.Errorf("EncodingCharacters() = %q, want %q (same as String())",
			d.EncodingCharacters(), d.String())
	}
}

func TestDelimiters_MSH1(t *testing.T) {
	tests := []struct {
		name       string
		delimiters *Delimiters
		want       string
	}{
		{
			name:       "default field separator",
			delimiters: DefaultDelimiters(),
			want:       "|",
		},
		{
			name: "custom field separator",
			delimiters: &Delimiters{
				Field: '!',
			},
			want: "!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.delimiters.MSH1()
			if got != tt.want {
				t.Errorf("Delimiters.MSH1() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDelimiters_MSH2(t *testing.T) {
	d := DefaultDelimiters()

	// MSH2 should return the same as EncodingCharacters
	if d.MSH2() != d.EncodingCharacters() {
		t.Errorf("MSH2() = %q, want %q (same as EncodingCharacters())",
			d.MSH2(), d.EncodingCharacters())
	}
}

func TestDelimiters_Equal(t *testing.T) {
	tests := []struct {
		name  string
		d1    *Delimiters
		d2    *Delimiters
		equal bool
	}{
		{
			name:  "both nil",
			d1:    nil,
			d2:    nil,
			equal: true,
		},
		{
			name:  "first nil",
			d1:    nil,
			d2:    DefaultDelimiters(),
			equal: false,
		},
		{
			name:  "second nil",
			d1:    DefaultDelimiters(),
			d2:    nil,
			equal: false,
		},
		{
			name:  "equal default delimiters",
			d1:    DefaultDelimiters(),
			d2:    DefaultDelimiters(),
			equal: true,
		},
		{
			name: "different field",
			d1:   DefaultDelimiters(),
			d2: &Delimiters{
				Field:        '!',
				Component:    '^',
				Repetition:   '~',
				Escape:       '\\',
				SubComponent: '&',
				Truncation:   '#',
			},
			equal: false,
		},
		{
			name: "different component",
			d1:   DefaultDelimiters(),
			d2: &Delimiters{
				Field:        '|',
				Component:    '@',
				Repetition:   '~',
				Escape:       '\\',
				SubComponent: '&',
				Truncation:   '#',
			},
			equal: false,
		},
		{
			name: "different truncation",
			d1:   DefaultDelimiters(),
			d2: &Delimiters{
				Field:        '|',
				Component:    '^',
				Repetition:   '~',
				Escape:       '\\',
				SubComponent: '&',
				Truncation:   '@',
			},
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d1.Equal(tt.d2)
			if got != tt.equal {
				t.Errorf("Delimiters.Equal() = %v, want %v", got, tt.equal)
			}
		})
	}
}

func TestSegmentTerminator(t *testing.T) {
	// Verify the segment terminator is carriage return (0x0D)
	if SegmentTerminator != '\r' {
		t.Errorf("SegmentTerminator = %#x, want %#x (carriage return)", SegmentTerminator, '\r')
	}

	if SegmentTerminator != 0x0D {
		t.Errorf("SegmentTerminator = %#x, want 0x0D", SegmentTerminator)
	}
}

func TestParseDelimiters_RoundTrip(t *testing.T) {
	// Test that parsing delimiters and encoding them produces consistent results
	tests := []struct {
		name       string
		mshSegment []byte
		wantMSH2   string
	}{
		{
			name:       "standard delimiters",
			mshSegment: []byte("MSH|^~\\&#|App|Fac|"),
			wantMSH2:   "^~\\&#",
		},
		{
			name:       "custom delimiters",
			mshSegment: []byte("MSH!@#$%^|App|Fac|"),
			wantMSH2:   "@#$%^",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := ParseDelimiters(tt.mshSegment)
			if err != nil {
				t.Fatalf("ParseDelimiters() error = %v", err)
			}

			got := d.EncodingCharacters()
			if got != tt.wantMSH2 {
				t.Errorf("Round trip MSH-2 = %q, want %q", got, tt.wantMSH2)
			}
		})
	}
}
