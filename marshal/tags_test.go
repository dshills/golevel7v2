package marshal

import (
	"errors"
	"testing"
)

func TestParseTag(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		wantLoc    string
		wantOmit   bool
		wantFormat string
		wantIgnore bool
		wantErr    error
	}{
		{
			name:       "simple location",
			tag:        "PID.5.1",
			wantLoc:    "PID.5.1",
			wantOmit:   false,
			wantFormat: "",
			wantIgnore: false,
			wantErr:    nil,
		},
		{
			name:       "location with omitempty",
			tag:        "PID.5.1,omitempty",
			wantLoc:    "PID.5.1",
			wantOmit:   true,
			wantFormat: "",
			wantIgnore: false,
			wantErr:    nil,
		},
		{
			name:       "location with format",
			tag:        "PID.7,format=20060102",
			wantLoc:    "PID.7",
			wantOmit:   false,
			wantFormat: "20060102",
			wantIgnore: false,
			wantErr:    nil,
		},
		{
			name:       "location with all options",
			tag:        "PID.5.1,omitempty,format=20060102",
			wantLoc:    "PID.5.1",
			wantOmit:   true,
			wantFormat: "20060102",
			wantIgnore: false,
			wantErr:    nil,
		},
		{
			name:       "ignore marker",
			tag:        "-",
			wantLoc:    "",
			wantOmit:   false,
			wantFormat: "",
			wantIgnore: true,
			wantErr:    nil,
		},
		{
			name:       "empty tag",
			tag:        "",
			wantLoc:    "",
			wantOmit:   false,
			wantFormat: "",
			wantIgnore: false,
			wantErr:    ErrEmptyTag,
		},
		{
			name:       "whitespace only tag",
			tag:        "   ",
			wantLoc:    "",
			wantOmit:   false,
			wantFormat: "",
			wantIgnore: false,
			wantErr:    ErrEmptyTag,
		},
		{
			name:       "location with spaces",
			tag:        "  PID.5.1 , omitempty ",
			wantLoc:    "PID.5.1",
			wantOmit:   true,
			wantFormat: "",
			wantIgnore: false,
			wantErr:    nil,
		},
		{
			name:       "empty location with options",
			tag:        ",omitempty",
			wantLoc:    "",
			wantOmit:   false,
			wantFormat: "",
			wantIgnore: false,
			wantErr:    ErrInvalidTagFormat,
		},
		{
			name:       "unknown option ignored",
			tag:        "PID.5.1,unknown,omitempty",
			wantLoc:    "PID.5.1",
			wantOmit:   true,
			wantFormat: "",
			wantIgnore: false,
			wantErr:    nil,
		},
		{
			name:       "field only",
			tag:        "PID.5",
			wantLoc:    "PID.5",
			wantOmit:   false,
			wantFormat: "",
			wantIgnore: false,
			wantErr:    nil,
		},
		{
			name:       "segment only",
			tag:        "PID",
			wantLoc:    "PID",
			wantOmit:   false,
			wantFormat: "",
			wantIgnore: false,
			wantErr:    nil,
		},
		{
			name:       "complex format",
			tag:        "MSH.7,format=20060102150405.0000",
			wantLoc:    "MSH.7",
			wantOmit:   false,
			wantFormat: "20060102150405.0000",
			wantIgnore: false,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTag(tt.tag)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("parseTag(%q) expected error %v, got nil", tt.tag, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("parseTag(%q) error = %v, want %v", tt.tag, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("parseTag(%q) unexpected error: %v", tt.tag, err)
				return
			}

			if got.location != tt.wantLoc {
				t.Errorf("parseTag(%q).location = %q, want %q", tt.tag, got.location, tt.wantLoc)
			}
			if got.omitEmpty != tt.wantOmit {
				t.Errorf("parseTag(%q).omitEmpty = %v, want %v", tt.tag, got.omitEmpty, tt.wantOmit)
			}
			if got.timeFormat != tt.wantFormat {
				t.Errorf("parseTag(%q).timeFormat = %q, want %q", tt.tag, got.timeFormat, tt.wantFormat)
			}
			if got.ignore != tt.wantIgnore {
				t.Errorf("parseTag(%q).ignore = %v, want %v", tt.tag, got.ignore, tt.wantIgnore)
			}
		})
	}
}

func TestTagInfo_HasLocation(t *testing.T) {
	tests := []struct {
		name string
		tag  *tagInfo
		want bool
	}{
		{
			name: "nil tagInfo",
			tag:  nil,
			want: false,
		},
		{
			name: "empty location",
			tag:  &tagInfo{location: ""},
			want: false,
		},
		{
			name: "ignored field",
			tag:  &tagInfo{location: "PID.5", ignore: true},
			want: false,
		},
		{
			name: "valid location",
			tag:  &tagInfo{location: "PID.5"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tag.hasLocation(); got != tt.want {
				t.Errorf("hasLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagInfo_ShouldOmit(t *testing.T) {
	tests := []struct {
		name            string
		tag             *tagInfo
		globalOmitEmpty bool
		want            bool
	}{
		{
			name:            "nil tagInfo",
			tag:             nil,
			globalOmitEmpty: false,
			want:            false,
		},
		{
			name:            "tag omitEmpty true",
			tag:             &tagInfo{omitEmpty: true},
			globalOmitEmpty: false,
			want:            true,
		},
		{
			name:            "global omitEmpty true",
			tag:             &tagInfo{omitEmpty: false},
			globalOmitEmpty: true,
			want:            true,
		},
		{
			name:            "both false",
			tag:             &tagInfo{omitEmpty: false},
			globalOmitEmpty: false,
			want:            false,
		},
		{
			name:            "both true",
			tag:             &tagInfo{omitEmpty: true},
			globalOmitEmpty: true,
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tag.shouldOmit(tt.globalOmitEmpty); got != tt.want {
				t.Errorf("shouldOmit(%v) = %v, want %v", tt.globalOmitEmpty, got, tt.want)
			}
		})
	}
}

func TestTagInfo_GetTimeFormat(t *testing.T) {
	tests := []struct {
		name          string
		tag           *tagInfo
		defaultFormat string
		want          string
	}{
		{
			name:          "nil tagInfo",
			tag:           nil,
			defaultFormat: "20060102",
			want:          "20060102",
		},
		{
			name:          "empty tag format uses default",
			tag:           &tagInfo{timeFormat: ""},
			defaultFormat: "20060102",
			want:          "20060102",
		},
		{
			name:          "tag format overrides default",
			tag:           &tagInfo{timeFormat: "20060102150405"},
			defaultFormat: "20060102",
			want:          "20060102150405",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tag.getTimeFormat(tt.defaultFormat); got != tt.want {
				t.Errorf("getTimeFormat(%q) = %q, want %q", tt.defaultFormat, got, tt.want)
			}
		})
	}
}
