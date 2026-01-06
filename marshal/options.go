// Package marshal provides struct marshaling and unmarshaling for HL7 v2.x messages.
// It enables bidirectional conversion between Go structs and HL7 messages using
// struct tags to specify field locations.
package marshal

import "time"

// Option configures the marshaler/unmarshaler behavior.
type Option func(*marshalConfig)

// marshalConfig holds configuration for marshaling/unmarshaling operations.
type marshalConfig struct {
	tagName      string         // struct tag name, default "hl7"
	omitEmpty    bool           // skip zero-value fields when marshaling
	timeFormat   string         // for time.Time fields, default "20060102150405"
	timeLocation *time.Location // timezone for time parsing, default UTC
}

// defaultConfig returns the default marshal configuration.
func defaultConfig() *marshalConfig {
	return &marshalConfig{
		tagName:      "hl7",
		omitEmpty:    false,
		timeFormat:   "20060102150405",
		timeLocation: time.UTC,
	}
}

// applyOptions applies the given options to the configuration.
func (c *marshalConfig) applyOptions(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

// WithTagName sets the struct tag name to use for HL7 field mapping.
// Default is "hl7".
//
// Example:
//
//	type Patient struct {
//	    Name string `custom:"PID.5"`
//	}
//	m := NewMarshaler(WithTagName("custom"))
func WithTagName(name string) Option {
	return func(c *marshalConfig) {
		if name != "" {
			c.tagName = name
		}
	}
}

// WithOmitEmpty controls whether zero-value fields are omitted when marshaling.
// When true, fields with zero values will not be written to the message.
// Default is false.
//
// Example:
//
//	m := NewMarshaler(WithOmitEmpty(true))
func WithOmitEmpty(omit bool) Option {
	return func(c *marshalConfig) {
		c.omitEmpty = omit
	}
}

// WithTimeFormat sets the time format string for parsing and formatting time.Time fields.
// Default is "20060102150405" (HL7 DTM format: YYYYMMDDHHMMSS).
//
// Common HL7 time formats:
//   - "20060102" - Date only (YYYYMMDD)
//   - "20060102150405" - Date and time (YYYYMMDDHHMMSS)
//   - "20060102150405.000" - With milliseconds
//   - "20060102150405-0700" - With timezone offset
//
// Example:
//
//	m := NewMarshaler(WithTimeFormat("20060102"))
func WithTimeFormat(format string) Option {
	return func(c *marshalConfig) {
		if format != "" {
			c.timeFormat = format
		}
	}
}

// WithTimeLocation sets the timezone for parsing time values that don't include
// timezone information. Default is UTC.
//
// Example:
//
//	loc, _ := time.LoadLocation("America/New_York")
//	m := NewMarshaler(WithTimeLocation(loc))
func WithTimeLocation(loc *time.Location) Option {
	return func(c *marshalConfig) {
		if loc != nil {
			c.timeLocation = loc
		}
	}
}
