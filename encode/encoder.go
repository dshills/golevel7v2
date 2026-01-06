package encode

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/dshills/golevel7/hl7"
)

// Encoder encodes HL7 messages to their wire format representation.
type Encoder interface {
	// Encode converts an HL7 message to bytes.
	// Returns the encoded message with appropriate line endings and optional MLLP framing.
	Encode(msg hl7.Message) ([]byte, error)

	// EncodeToWriter writes an encoded HL7 message to the provided writer.
	// The context can be used for cancellation during long-running writes.
	EncodeToWriter(ctx context.Context, w io.Writer, msg hl7.Message) error
}

// encoder is the concrete implementation of Encoder.
type encoder struct {
	config encoderConfig
}

// New creates a new Encoder with the given options.
// If no options are provided, default settings are used:
//   - Line ending: "\r" (carriage return)
//   - MLLP framing: disabled
//   - Trailing delimiters: excluded
func New(opts ...EncoderOption) Encoder {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &encoder{config: cfg}
}

// Encode converts an HL7 message to bytes.
// The message segments are encoded using their Bytes() method and joined
// with the configured line ending. If MLLP is enabled, the result is
// wrapped with MLLP framing bytes.
func (e *encoder) Encode(msg hl7.Message) ([]byte, error) {
	if msg == nil {
		return nil, &Error{Message: "cannot encode nil message"}
	}

	segments := msg.AllSegments()
	if len(segments) == 0 {
		return nil, &Error{Message: "message has no segments"}
	}

	// Get delimiters from the message
	delims := msg.Delimiters()
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	// Calculate approximate capacity for buffer
	// Average segment is roughly 80 bytes, plus line endings
	estimatedSize := len(segments) * 100
	if e.config.includeMLLP {
		estimatedSize += 3 // MLLP framing bytes
	}

	var buf bytes.Buffer
	buf.Grow(estimatedSize)

	// Add MLLP start block if enabled
	if e.config.includeMLLP {
		buf.WriteByte(MLLPStartBlock)
	}

	// Encode each segment
	for i, seg := range segments {
		if i > 0 {
			buf.WriteString(e.config.lineEnding)
		}
		segBytes := seg.Bytes(delims)
		buf.Write(segBytes)
	}

	// Add final line ending after last segment
	buf.WriteString(e.config.lineEnding)

	// Add MLLP end block if enabled
	if e.config.includeMLLP {
		buf.WriteByte(MLLPEndBlock)
		buf.WriteByte(MLLPCarriageReturn)
	}

	return buf.Bytes(), nil
}

// EncodeToWriter writes an encoded HL7 message to the provided writer.
// It checks for context cancellation before each segment write operation.
// This method is suitable for streaming large messages or writing to
// network connections where cancellation support is important.
func (e *encoder) EncodeToWriter(ctx context.Context, w io.Writer, msg hl7.Message) error {
	if msg == nil {
		return &Error{Message: "cannot encode nil message"}
	}

	segments := msg.AllSegments()
	if len(segments) == 0 {
		return &Error{Message: "message has no segments"}
	}

	// Check context before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get delimiters from the message
	delims := msg.Delimiters()
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	// Add MLLP start block if enabled
	if e.config.includeMLLP {
		if _, err := w.Write([]byte{MLLPStartBlock}); err != nil {
			return &Error{Message: "failed to write MLLP start block", Cause: err}
		}
	}

	lineEndingBytes := []byte(e.config.lineEnding)

	// Encode each segment
	for i, seg := range segments {
		// Check for context cancellation between segments
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if i > 0 {
			if _, err := w.Write(lineEndingBytes); err != nil {
				return &Error{
					Message:  "failed to write line ending",
					Segment:  seg.Name(),
					Position: i,
					Cause:    err,
				}
			}
		}

		segBytes := seg.Bytes(delims)
		if _, err := w.Write(segBytes); err != nil {
			return &Error{
				Message:  "failed to write segment",
				Segment:  seg.Name(),
				Position: i,
				Cause:    err,
			}
		}
	}

	// Add final line ending after last segment
	if _, err := w.Write(lineEndingBytes); err != nil {
		return &Error{Message: "failed to write final line ending", Cause: err}
	}

	// Add MLLP end block if enabled
	if e.config.includeMLLP {
		if _, err := w.Write([]byte{MLLPEndBlock, MLLPCarriageReturn}); err != nil {
			return &Error{Message: "failed to write MLLP end block", Cause: err}
		}
	}

	return nil
}

// Error represents an error that occurred during message encoding.
type Error struct {
	// Message describes what went wrong.
	Message string
	// Segment is the segment name where the error occurred (if applicable).
	Segment string
	// Position is the segment index where the error occurred (if applicable).
	Position int
	// Cause is the underlying error that caused this encode error.
	Cause error
}

// Error implements the error interface.
func (e *Error) Error() string {
	msg := "encode error"
	if e.Segment != "" {
		msg = fmt.Sprintf("%s at segment %s", msg, e.Segment)
		if e.Position > 0 {
			msg = fmt.Sprintf("%s (position %d)", msg, e.Position)
		}
	}
	if e.Message != "" {
		msg = msg + ": " + e.Message
	}
	if e.Cause != nil {
		msg = msg + ": " + e.Cause.Error()
	}
	return msg
}

// Unwrap returns the underlying cause of the encode error.
func (e *Error) Unwrap() error {
	return e.Cause
}
