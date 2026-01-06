// Package parse provides HL7 v2.x message parsing functionality.
package parse

import "github.com/dshills/golevel7/hl7"

// Default parser configuration values.
const (
	defaultMaxSegments    = 1000  // DoS protection: maximum segments per message
	defaultMaxFieldLength = 65536 // DoS protection: maximum field length in bytes
)

// parserConfig holds the parser configuration.
type parserConfig struct {
	strictMode         bool            // Enable strict parsing mode
	allowEmptySegments bool            // Allow empty segments in messages
	customDelimiters   *hl7.Delimiters // Use custom delimiters instead of extracting from MSH
	maxSegments        int             // Maximum segments allowed (DoS protection)
	maxFieldLength     int             // Maximum field length allowed (DoS protection)
	segmentTerminator  rune            // Segment terminator character (default CR)
}

// defaultConfig returns a parser configuration with default values.
func defaultConfig() parserConfig {
	return parserConfig{
		strictMode:         false,
		allowEmptySegments: false,
		customDelimiters:   nil,
		maxSegments:        defaultMaxSegments,
		maxFieldLength:     defaultMaxFieldLength,
		segmentTerminator:  hl7.SegmentTerminator,
	}
}

// ParserOption is a functional option for configuring the parser.
type ParserOption func(*parserConfig)

// WithStrictMode enables or disables strict parsing mode.
// In strict mode, the parser is more rigorous about HL7 compliance
// and will reject messages with minor formatting issues.
func WithStrictMode(strict bool) ParserOption {
	return func(c *parserConfig) {
		c.strictMode = strict
	}
}

// WithAllowEmptySegments configures whether empty segments are allowed.
// When enabled, segments with no fields (just the segment name) are permitted.
func WithAllowEmptySegments(allow bool) ParserOption {
	return func(c *parserConfig) {
		c.allowEmptySegments = allow
	}
}

// WithCustomDelimiters sets custom delimiters for parsing.
// When set, the parser will use these delimiters instead of extracting
// them from the MSH segment. This is useful for parsing non-standard
// messages or message fragments.
func WithCustomDelimiters(d *hl7.Delimiters) ParserOption {
	return func(c *parserConfig) {
		c.customDelimiters = d
	}
}

// WithMaxSegments sets the maximum number of segments allowed in a message.
// This is a DoS protection mechanism to prevent processing of maliciously
// large messages. Default is 1000.
func WithMaxSegments(limit int) ParserOption {
	return func(c *parserConfig) {
		if limit > 0 {
			c.maxSegments = limit
		}
	}
}

// WithMaxFieldLength sets the maximum field length allowed.
// This is a DoS protection mechanism to prevent processing of messages
// with excessively large fields. Default is 65536 bytes.
func WithMaxFieldLength(limit int) ParserOption {
	return func(c *parserConfig) {
		if limit > 0 {
			c.maxFieldLength = limit
		}
	}
}

// WithSegmentTerminator sets the segment terminator character.
// The default is carriage return (CR, 0x0D) as per HL7 standard.
// Some implementations use line feed (LF, 0x0A) or other characters.
func WithSegmentTerminator(term rune) ParserOption {
	return func(c *parserConfig) {
		c.segmentTerminator = term
	}
}
