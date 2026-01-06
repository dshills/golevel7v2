package segments

import (
	"github.com/dshills/golevel7/hl7"
)

// getFieldValue extracts a string value from a segment field at the given position.
// Returns an empty string if the field does not exist.
// Uses Field.String() to get the full field value including all components and repetitions.
func getFieldValue(seg hl7.Segment, fieldNum int) string {
	if f, ok := seg.Field(fieldNum); ok {
		return f.String()
	}
	return ""
}

// buildSegmentData constructs a segment string from a name and slice of field values.
// Empty trailing fields are omitted to avoid unnecessary trailing delimiters.
func buildSegmentData(name string, fields []string, delims *hl7.Delimiters) string {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	fieldSep := string(delims.Field)

	// Find the last non-empty field to avoid trailing delimiters
	lastNonEmpty := -1
	for i := len(fields) - 1; i >= 0; i-- {
		if fields[i] != "" {
			lastNonEmpty = i
			break
		}
	}

	data := name

	// Append fields up to and including the last non-empty field
	for i := 0; i <= lastNonEmpty; i++ {
		data += fieldSep + fields[i]
	}

	return data
}
