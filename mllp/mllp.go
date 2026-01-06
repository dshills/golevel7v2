package mllp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

// MLLP (Minimal Lower Layer Protocol) framing bytes as defined in the
// HL7 v2.x standard for message transmission over TCP/IP.
const (
	// StartBlock is the start-of-message byte (0x0B, vertical tab).
	// Every MLLP message begins with this byte.
	StartBlock = 0x0B

	// EndBlock is the end-of-message byte (0x1C, file separator).
	// This byte signals the end of the HL7 message content.
	EndBlock = 0x1C

	// CarriageReturn follows the EndBlock (0x0D, carriage return).
	// The complete message terminator is EndBlock + CarriageReturn.
	CarriageReturn = 0x0D
)

// Common errors returned by MLLP operations.
var (
	// ErrInvalidStartBlock is returned when a message does not begin with StartBlock.
	ErrInvalidStartBlock = errors.New("mllp: message does not start with start block (0x0B)")

	// ErrInvalidEndBlock is returned when a message does not end with the proper trailer.
	ErrInvalidEndBlock = errors.New("mllp: message does not end with end block sequence (0x1C 0x0D)")

	// ErrMessageTooLarge is returned when a message exceeds the maximum allowed size.
	ErrMessageTooLarge = errors.New("mllp: message exceeds maximum allowed size")

	// ErrNoHandler is returned when a server receives a message but has no handler configured.
	ErrNoHandler = errors.New("mllp: no message handler configured")

	// ErrServerClosed is returned when operations are attempted on a closed server.
	ErrServerClosed = errors.New("mllp: server closed")

	// ErrConnectionClosed is returned when the connection is closed unexpectedly.
	ErrConnectionClosed = errors.New("mllp: connection closed")

	// ErrMaxConnectionsReached is returned when the server has reached its connection limit.
	ErrMaxConnectionsReached = errors.New("mllp: maximum connections reached")
)

// MaxMessageSize is the default maximum message size (16 MB).
// This can be overridden using configuration options.
const MaxMessageSize = 16 * 1024 * 1024

// Frame wraps raw HL7 message data with MLLP framing.
// The returned slice contains StartBlock + data + EndBlock + CarriageReturn.
func Frame(data []byte) []byte {
	result := make([]byte, len(data)+3)
	result[0] = StartBlock
	copy(result[1:], data)
	result[len(data)+1] = EndBlock
	result[len(data)+2] = CarriageReturn
	return result
}

// Unframe removes MLLP framing from a message and returns the raw HL7 data.
// Returns an error if the framing is invalid.
func Unframe(data []byte) ([]byte, error) {
	if len(data) < 3 {
		return nil, ErrInvalidStartBlock
	}

	if data[0] != StartBlock {
		return nil, ErrInvalidStartBlock
	}

	if len(data) < 3 || data[len(data)-2] != EndBlock || data[len(data)-1] != CarriageReturn {
		return nil, ErrInvalidEndBlock
	}

	return data[1 : len(data)-2], nil
}

// Reader wraps an io.Reader to read MLLP-framed messages.
type Reader struct {
	reader    *bufio.Reader
	maxSize   int
	buf       bytes.Buffer
	inMessage bool
}

// NewReader creates a new MLLP reader that reads from r.
// The maxSize parameter limits the maximum message size to prevent DoS attacks.
// If maxSize is 0, MaxMessageSize is used.
func NewReader(r io.Reader, maxSize int) *Reader {
	if maxSize <= 0 {
		maxSize = MaxMessageSize
	}
	return &Reader{
		reader:  bufio.NewReader(r),
		maxSize: maxSize,
	}
}

// ReadMessage reads the next MLLP-framed message from the underlying reader.
// It returns the raw HL7 message data without MLLP framing.
// Returns io.EOF when the connection is closed.
func (r *Reader) ReadMessage() ([]byte, error) {
	r.buf.Reset()
	r.inMessage = false

	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			if err == io.EOF && r.buf.Len() > 0 {
				return nil, ErrConnectionClosed
			}
			return nil, err
		}

		if !r.inMessage {
			// Looking for start block
			if b == StartBlock {
				r.inMessage = true
				continue
			}
			// Ignore bytes before start block (common with keep-alive)
			continue
		}

		// Check for end block
		if b == EndBlock {
			// Read the expected carriage return
			next, err := r.reader.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("mllp: error reading after end block: %w", err)
			}
			if next != CarriageReturn {
				// Not a valid end sequence, include both bytes in message
				r.buf.WriteByte(b)
				r.buf.WriteByte(next)
				continue
			}
			// Valid message complete
			return r.buf.Bytes(), nil
		}

		// Regular message byte
		if r.buf.Len() >= r.maxSize {
			return nil, ErrMessageTooLarge
		}
		r.buf.WriteByte(b)
	}
}

// Writer wraps an io.Writer to write MLLP-framed messages.
type Writer struct {
	writer io.Writer
}

// NewWriter creates a new MLLP writer that writes to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

// WriteMessage writes an HL7 message with MLLP framing to the underlying writer.
// It adds the start block before the message and end block + CR after.
func (w *Writer) WriteMessage(data []byte) error {
	// Write start block
	if _, err := w.writer.Write([]byte{StartBlock}); err != nil {
		return fmt.Errorf("mllp: error writing start block: %w", err)
	}

	// Write message data
	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("mllp: error writing message data: %w", err)
	}

	// Write end block and carriage return
	if _, err := w.writer.Write([]byte{EndBlock, CarriageReturn}); err != nil {
		return fmt.Errorf("mllp: error writing end block: %w", err)
	}

	return nil
}
