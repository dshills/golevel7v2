// Package encode provides HL7 v2.x message encoding functionality.
// It converts HL7 message structures back to the wire format with
// configurable options for line endings, MLLP framing, and delimiters.
package encode

// MLLP (Minimal Lower Layer Protocol) framing bytes.
const (
	// MLLPStartBlock is the vertical tab character that starts an MLLP frame.
	MLLPStartBlock = 0x0B
	// MLLPEndBlock is the file separator character that ends an MLLP frame.
	MLLPEndBlock = 0x1C
	// MLLPCarriageReturn is the carriage return after the end block.
	MLLPCarriageReturn = 0x0D
)

// Default encoder settings.
const (
	// DefaultLineEnding is the standard HL7 segment terminator (carriage return).
	DefaultLineEnding = "\r"
)

// encoderConfig holds the configuration options for encoding HL7 messages.
type encoderConfig struct {
	lineEnding         string // segment terminator, default "\r"
	includeMLLP        bool   // wrap in MLLP framing
	trailingDelimiters bool   // include trailing empty delimiters
}

// defaultConfig returns an encoderConfig with default settings.
func defaultConfig() encoderConfig {
	return encoderConfig{
		lineEnding:         DefaultLineEnding,
		includeMLLP:        false,
		trailingDelimiters: false,
	}
}

// EncoderOption is a functional option for configuring an encoder.
type EncoderOption func(*encoderConfig)

// WithLineEnding sets the segment terminator string.
// The default is "\r" (carriage return) per HL7 specification.
// Some systems may require "\r\n" (CRLF) for compatibility.
func WithLineEnding(ending string) EncoderOption {
	return func(c *encoderConfig) {
		c.lineEnding = ending
	}
}

// WithMLLP enables or disables MLLP (Minimal Lower Layer Protocol) framing.
// When enabled, messages are wrapped with:
//   - Start block: 0x0B (vertical tab)
//   - End block: 0x1C 0x0D (file separator + carriage return)
//
// MLLP is commonly used for TCP/IP transmission of HL7 messages.
func WithMLLP(enable bool) EncoderOption {
	return func(c *encoderConfig) {
		c.includeMLLP = enable
	}
}

// WithTrailingDelimiters controls whether trailing empty delimiters are included.
// When false (default), trailing empty fields, components, and subcomponents
// are omitted from the encoded output.
// When true, delimiters are preserved even for empty trailing elements.
func WithTrailingDelimiters(include bool) EncoderOption {
	return func(c *encoderConfig) {
		c.trailingDelimiters = include
	}
}
