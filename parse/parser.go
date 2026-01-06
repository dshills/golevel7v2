// Package parse provides HL7 v2.x message parsing functionality.
package parse

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/dshills/golevel7/hl7"
)

// MLLP (Minimal Lower Layer Protocol) framing bytes.
const (
	mllpStartByte = 0x0B // Vertical Tab (VT)
	mllpEndByte1  = 0x1C // File Separator (FS)
	mllpEndByte2  = 0x0D // Carriage Return (CR)
)

// Parser-specific errors.
var (
	// ErrTooManySegments is returned when the message exceeds maxSegments.
	ErrTooManySegments = errors.New("message exceeds maximum segment count")
	// ErrFieldTooLong is returned when a field exceeds maxFieldLength.
	ErrFieldTooLong = errors.New("field exceeds maximum length")
	// ErrContextCanceled is returned when the parsing context is canceled.
	ErrContextCanceled = errors.New("parsing canceled")
	// ErrEmptySegment is returned when an empty segment is found and not allowed.
	ErrEmptySegment = errors.New("empty segment not allowed")
)

// Parser defines the interface for HL7 message parsing.
type Parser interface {
	// Parse parses raw HL7 message data into a Message.
	// The input data may include MLLP framing which will be stripped.
	Parse(data []byte) (hl7.Message, error)

	// ParseContext parses raw HL7 message data with context support.
	// Allows for cancellation during parsing of large messages.
	ParseContext(ctx context.Context, data []byte) (hl7.Message, error)
}

// parser is the concrete implementation of Parser.
type parser struct {
	config parserConfig
}

// New creates a new Parser with the given options.
func New(opts ...ParserOption) Parser {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &parser{config: cfg}
}

// Parse parses raw HL7 message data into a Message.
func (p *parser) Parse(data []byte) (hl7.Message, error) {
	return p.ParseContext(context.Background(), data)
}

// ParseContext parses raw HL7 message data with context support.
func (p *parser) ParseContext(ctx context.Context, data []byte) (hl7.Message, error) {
	// Check for cancellation at start
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
	default:
	}

	// Strip MLLP framing if present
	data = stripMLLP(data)

	// Validate non-empty (including whitespace-only input)
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, hl7.ErrEmptyMessage
	}

	// Get delimiters - either custom or from MSH
	delims, err := p.getDelimiters(data)
	if err != nil {
		return nil, err
	}

	// Split message into segment data
	segmentData := p.splitSegments(data)

	// Validate segment count
	if len(segmentData) > p.config.maxSegments {
		return nil, fmt.Errorf("%w: got %d, max %d", ErrTooManySegments, len(segmentData), p.config.maxSegments)
	}

	// Create message with delimiters
	msg := hl7.NewMessageWithDelimiters(delims)

	// Parse segments
	for i, sd := range segmentData {
		// Check for cancellation periodically
		if i%100 == 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
			default:
			}
		}

		// Handle empty segments
		if len(bytes.TrimSpace(sd)) == 0 {
			if p.config.allowEmptySegments {
				continue
			}
			if p.config.strictMode {
				return nil, &hl7.ParseError{
					Message: ErrEmptySegment.Error(),
					Line:    i + 1,
				}
			}
			continue
		}

		// Check field length limit
		if err := p.checkFieldLengths(sd, delims); err != nil {
			return nil, &hl7.ParseError{
				Message: err.Error(),
				Line:    i + 1,
				Cause:   err,
			}
		}

		// Parse segment - ParseSegment expects []rune
		seg, err := hl7.ParseSegment([]rune(string(sd)), delims)
		if err != nil {
			return nil, &hl7.ParseError{
				Message: "failed to parse segment",
				Line:    i + 1,
				Cause:   err,
			}
		}

		if err := msg.AddSegment(seg); err != nil {
			return nil, &hl7.ParseError{
				Message: "failed to add segment",
				Line:    i + 1,
				Cause:   err,
			}
		}
	}

	// Validate MSH is first segment
	allSegs := msg.AllSegments()
	if len(allSegs) == 0 {
		return nil, hl7.ErrMissingMSH
	}
	if allSegs[0].Name() != "MSH" {
		return nil, hl7.ErrMissingMSH
	}

	return msg, nil
}

// stripMLLP removes MLLP framing from the data if present.
// MLLP format: <VT>message<FS><CR> where VT=0x0B, FS=0x1C, CR=0x0D
func stripMLLP(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	// Check for start byte
	if data[0] == mllpStartByte {
		data = data[1:]
	}

	// Check for end bytes (FS CR)
	if len(data) >= 2 {
		if data[len(data)-2] == mllpEndByte1 && data[len(data)-1] == mllpEndByte2 {
			data = data[:len(data)-2]
		} else if data[len(data)-1] == mllpEndByte1 {
			// Some implementations only use FS without CR
			data = data[:len(data)-1]
		}
	}

	return data
}

// getDelimiters returns the delimiters to use for parsing.
func (p *parser) getDelimiters(data []byte) (*hl7.Delimiters, error) {
	if p.config.customDelimiters != nil {
		return p.config.customDelimiters, nil
	}
	return hl7.ParseDelimiters(data)
}

// splitSegments splits message data into individual segment byte slices.
// Empty segments are included (as empty slices) so they can be detected during parsing.
func (p *parser) splitSegments(data []byte) [][]byte {
	terminator := byte(p.config.segmentTerminator)
	var segments [][]byte
	start := 0

	for i := 0; i < len(data); i++ {
		if data[i] == terminator {
			// Include segment (may be empty)
			segments = append(segments, data[start:i])
			start = i + 1
		}
	}

	// Handle last segment without terminator
	if start < len(data) {
		remaining := bytes.TrimSpace(data[start:])
		if len(remaining) > 0 {
			segments = append(segments, remaining)
		}
	}

	return segments
}

// checkFieldLengths validates that no field exceeds the maximum length.
func (p *parser) checkFieldLengths(segmentData []byte, delims *hl7.Delimiters) error {
	fieldDelim := byte(delims.Field)
	start := 0
	fieldNum := 0

	for i := 0; i <= len(segmentData); i++ {
		if i == len(segmentData) || segmentData[i] == fieldDelim {
			fieldLen := i - start
			if fieldLen > p.config.maxFieldLength {
				return fmt.Errorf("%w: field %d is %d bytes, max %d",
					ErrFieldTooLong, fieldNum, fieldLen, p.config.maxFieldLength)
			}
			start = i + 1
			fieldNum++
		}
	}

	return nil
}
