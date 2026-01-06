// Package escape provides HL7 escape sequence encoding and decoding.
//
// HL7 v2.x uses escape sequences to encode special characters within field values.
// This package handles the standard escape sequences defined in the HL7 specification:
//
//   - \F\ - Field separator (|)
//   - \S\ - Component separator (^)
//   - \T\ - Subcomponent separator (&)
//   - \R\ - Repetition separator (~)
//   - \E\ - Escape character (\)
//   - \Xdd...\ - Hexadecimal encoded data
//   - \.br\ - Line break
package escape

import (
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"github.com/dshills/golevel7/hl7"
)

// Escaper handles HL7 escape sequence encoding and decoding.
type Escaper struct {
	delims *hl7.Delimiters
}

// New creates a new Escaper with the given delimiters.
// If delims is nil, default delimiters are used.
func New(delims *hl7.Delimiters) *Escaper {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}
	return &Escaper{delims: delims}
}

// Escape encodes special characters in the value using HL7 escape sequences.
// Characters that require escaping:
//   - Field separator -> \F\
//   - Component separator -> \S\
//   - Subcomponent separator -> \T\
//   - Repetition separator -> \R\
//   - Escape character -> \E\
func (e *Escaper) Escape(value string) string {
	if value == "" {
		return value
	}

	esc := e.delims.Escape

	// Pre-calculate if we need to escape anything
	needsEscape := false
	for _, r := range value {
		if r == e.delims.Field || r == e.delims.Component ||
			r == e.delims.SubComponent || r == e.delims.Repetition ||
			r == esc {
			needsEscape = true
			break
		}
	}

	if !needsEscape {
		return value
	}

	var sb strings.Builder
	sb.Grow(len(value) * 2) // Rough estimate for escaped output

	for _, r := range value {
		switch r {
		case esc:
			// Escape character -> \E\
			sb.WriteRune(esc)
			sb.WriteRune('E')
			sb.WriteRune(esc)
		case e.delims.Field:
			// Field separator -> \F\
			sb.WriteRune(esc)
			sb.WriteRune('F')
			sb.WriteRune(esc)
		case e.delims.Component:
			// Component separator -> \S\
			sb.WriteRune(esc)
			sb.WriteRune('S')
			sb.WriteRune(esc)
		case e.delims.SubComponent:
			// Subcomponent separator -> \T\
			sb.WriteRune(esc)
			sb.WriteRune('T')
			sb.WriteRune(esc)
		case e.delims.Repetition:
			// Repetition separator -> \R\
			sb.WriteRune(esc)
			sb.WriteRune('R')
			sb.WriteRune(esc)
		default:
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

// Unescape decodes HL7 escape sequences in the value.
// Supported escape sequences:
//   - \F\ -> Field separator
//   - \S\ -> Component separator
//   - \T\ -> Subcomponent separator
//   - \R\ -> Repetition separator
//   - \E\ -> Escape character
//   - \Xdd...\ -> Hexadecimal data (dd are hex digits)
//   - \.br\ -> Line break (\n)
//
// Malformed escape sequences (unclosed or unrecognized) are passed through unchanged.
func (e *Escaper) Unescape(value string) string {
	if value == "" {
		return value
	}

	esc := e.delims.Escape

	// Quick check if there's anything to unescape
	if !strings.ContainsRune(value, esc) {
		return value
	}

	var sb strings.Builder
	sb.Grow(len(value))

	runes := []rune(value)
	i := 0

	for i < len(runes) {
		if runes[i] != esc {
			sb.WriteRune(runes[i])
			i++
			continue
		}

		// Found escape character, try to parse escape sequence
		seq, length := e.parseEscapeSequence(runes, i)
		if length > 0 {
			sb.WriteString(seq)
			i += length
		} else {
			// Not a valid escape sequence, output the escape character as-is
			sb.WriteRune(runes[i])
			i++
		}
	}

	return sb.String()
}

// parseEscapeSequence attempts to parse an escape sequence starting at position i.
// Returns the decoded string and the number of runes consumed.
// Returns ("", 0) if the sequence is malformed or unrecognized.
func (e *Escaper) parseEscapeSequence(runes []rune, i int) (string, int) {
	esc := e.delims.Escape

	// Minimum escape sequence is 3 characters: \X\
	if i+2 >= len(runes) {
		return "", 0
	}

	// Find the closing escape character
	closeIdx := -1
	for j := i + 1; j < len(runes); j++ {
		if runes[j] == esc {
			closeIdx = j
			break
		}
	}

	if closeIdx == -1 {
		// No closing escape character found
		return "", 0
	}

	// Extract the content between escape characters
	content := string(runes[i+1 : closeIdx])
	length := closeIdx - i + 1

	// Handle standard single-character escape codes
	if len(content) == 1 {
		switch content[0] {
		case 'F':
			return string(e.delims.Field), length
		case 'S':
			return string(e.delims.Component), length
		case 'T':
			return string(e.delims.SubComponent), length
		case 'R':
			return string(e.delims.Repetition), length
		case 'E':
			return string(esc), length
		}
	}

	// Handle hex encoding: \Xdd...\
	if len(content) >= 2 && (content[0] == 'X' || content[0] == 'x') {
		hexStr := content[1:]
		decoded, err := e.decodeHex(hexStr)
		if err == nil {
			return decoded, length
		}
		// Invalid hex, fall through to return unchanged
	}

	// Handle line break: \.br\
	if content == ".br" {
		return "\n", length
	}

	// Handle other formatting escape sequences
	// These are less common but defined in the spec
	switch content {
	case ".sp":
		// Spacing - typically ignored or treated as space
		return " ", length
	case ".fi":
		// Start word wrap - typically ignored
		return "", length
	case ".nf":
		// End word wrap - typically ignored
		return "", length
	case ".in":
		// Indent - typically ignored
		return "", length
	case ".ti":
		// Temporary indent - typically ignored
		return "", length
	case ".sk":
		// Skip line - treat as newline
		return "\n", length
	case ".ce":
		// Center - typically ignored
		return "", length
	}

	// Unrecognized escape sequence - return as-is (the entire sequence)
	return string(runes[i : closeIdx+1]), length
}

// decodeHex decodes a hexadecimal string into its byte representation.
// The hex string should contain pairs of hex digits representing bytes.
func (e *Escaper) decodeHex(hexStr string) (string, error) {
	// Hex string must have even length
	if len(hexStr)%2 != 0 {
		return "", hex.ErrLength
	}

	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}

	// Validate that the decoded bytes form valid UTF-8
	if utf8.Valid(decoded) {
		return string(decoded), nil
	}

	// For invalid UTF-8, return the raw bytes as a string
	// This preserves binary data that might be intentionally non-UTF-8
	return string(decoded), nil
}

// EncodeHex encodes a string as a hexadecimal escape sequence.
// Returns the hex-encoded string in the format \Xdd...\
func (e *Escaper) EncodeHex(value string) string {
	if value == "" {
		return value
	}

	esc := e.delims.Escape
	hexStr := hex.EncodeToString([]byte(value))

	var sb strings.Builder
	sb.WriteRune(esc)
	sb.WriteRune('X')
	sb.WriteString(strings.ToUpper(hexStr))
	sb.WriteRune(esc)

	return sb.String()
}
