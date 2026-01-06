package marshal

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.tagName != "hl7" {
		t.Errorf("tagName = %q, want %q", cfg.tagName, "hl7")
	}
	if cfg.omitEmpty != false {
		t.Errorf("omitEmpty = %v, want false", cfg.omitEmpty)
	}
	if cfg.timeFormat != "20060102150405" {
		t.Errorf("timeFormat = %q, want %q", cfg.timeFormat, "20060102150405")
	}
	if cfg.timeLocation != time.UTC {
		t.Errorf("timeLocation = %v, want UTC", cfg.timeLocation)
	}
}

func TestWithTagName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "custom tag name",
			input:    "custom",
			expected: "custom",
		},
		{
			name:     "empty tag name keeps default",
			input:    "",
			expected: "hl7", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			WithTagName(tt.input)(cfg)
			if cfg.tagName != tt.expected {
				t.Errorf("tagName = %q, want %q", cfg.tagName, tt.expected)
			}
		})
	}
}

func TestWithOmitEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected bool
	}{
		{
			name:     "set omitEmpty true",
			input:    true,
			expected: true,
		},
		{
			name:     "set omitEmpty false",
			input:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			WithOmitEmpty(tt.input)(cfg)
			if cfg.omitEmpty != tt.expected {
				t.Errorf("omitEmpty = %v, want %v", cfg.omitEmpty, tt.expected)
			}
		})
	}
}

func TestWithTimeFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "custom time format",
			input:    "20060102",
			expected: "20060102",
		},
		{
			name:     "empty time format keeps default",
			input:    "",
			expected: "20060102150405", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			WithTimeFormat(tt.input)(cfg)
			if cfg.timeFormat != tt.expected {
				t.Errorf("timeFormat = %q, want %q", cfg.timeFormat, tt.expected)
			}
		})
	}
}

func TestWithTimeLocation(t *testing.T) {
	tests := []struct {
		name     string
		input    *time.Location
		expected *time.Location
	}{
		{
			name:     "custom location",
			input:    time.Local,
			expected: time.Local,
		},
		{
			name:     "nil location keeps default",
			input:    nil,
			expected: time.UTC, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			WithTimeLocation(tt.input)(cfg)
			if cfg.timeLocation != tt.expected {
				t.Errorf("timeLocation = %v, want %v", cfg.timeLocation, tt.expected)
			}
		})
	}
}

func TestApplyOptions(t *testing.T) {
	cfg := defaultConfig()
	loc, _ := time.LoadLocation("America/New_York")

	cfg.applyOptions(
		WithTagName("custom"),
		WithOmitEmpty(true),
		WithTimeFormat("20060102"),
		WithTimeLocation(loc),
	)

	if cfg.tagName != "custom" {
		t.Errorf("tagName = %q, want %q", cfg.tagName, "custom")
	}
	if cfg.omitEmpty != true {
		t.Errorf("omitEmpty = %v, want true", cfg.omitEmpty)
	}
	if cfg.timeFormat != "20060102" {
		t.Errorf("timeFormat = %q, want %q", cfg.timeFormat, "20060102")
	}
	if cfg.timeLocation != loc {
		t.Errorf("timeLocation = %v, want %v", cfg.timeLocation, loc)
	}
}
