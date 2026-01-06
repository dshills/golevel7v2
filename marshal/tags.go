package marshal

import (
	"errors"
	"strings"
)

// Tag parsing errors.
var (
	// ErrEmptyTag indicates an empty tag string was provided.
	ErrEmptyTag = errors.New("empty tag")
	// ErrInvalidTagFormat indicates the tag format is invalid.
	ErrInvalidTagFormat = errors.New("invalid tag format")
)

// tagInfo holds parsed struct tag information.
type tagInfo struct {
	location   string // HL7 location path (e.g., "PID.5.1")
	omitEmpty  bool   // skip if field is zero value
	timeFormat string // custom time format for this field
	ignore     bool   // ignore this field (tag is "-")
}

// parseTag parses an HL7 struct tag into tagInfo.
// Tag format: "location[,option[,option...]]"
//
// Supported options:
//   - omitempty: skip field if zero value when marshaling
//   - format=<layout>: custom time format for time.Time fields
//   - -: ignore this field
//
// Examples:
//
//	`hl7:"PID.5.1"`                    - simple location
//	`hl7:"PID.5.1,omitempty"`          - with omitempty
//	`hl7:"PID.7,format=20060102"`      - with custom time format
//	`hl7:"PID.5.1,omitempty,format=20060102"` - multiple options
//	`hl7:"-"`                          - ignore field
func parseTag(tag string) (*tagInfo, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil, ErrEmptyTag
	}

	// Check for ignore marker
	if tag == "-" {
		return &tagInfo{ignore: true}, nil
	}

	info := &tagInfo{}

	// Split by comma to get location and options
	parts := strings.Split(tag, ",")

	// First part is always the location
	location := strings.TrimSpace(parts[0])
	if location == "" {
		return nil, ErrInvalidTagFormat
	}
	info.location = location

	// Parse remaining options
	for i := 1; i < len(parts); i++ {
		opt := strings.TrimSpace(parts[i])
		if opt == "" {
			continue
		}

		switch {
		case opt == "omitempty":
			info.omitEmpty = true
		case strings.HasPrefix(opt, "format="):
			info.timeFormat = strings.TrimPrefix(opt, "format=")
		default:
			// Unknown options are ignored for forward compatibility
		}
	}

	return info, nil
}

// hasLocation returns true if the tag specifies a location.
func (t *tagInfo) hasLocation() bool {
	return t != nil && t.location != "" && !t.ignore
}

// shouldOmit returns true if the field should be omitted when marshaling.
func (t *tagInfo) shouldOmit(globalOmitEmpty bool) bool {
	if t == nil {
		return false
	}
	return t.omitEmpty || globalOmitEmpty
}

// getTimeFormat returns the time format to use, with the given default.
func (t *tagInfo) getTimeFormat(defaultFormat string) string {
	if t != nil && t.timeFormat != "" {
		return t.timeFormat
	}
	return defaultFormat
}
