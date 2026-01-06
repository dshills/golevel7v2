package encode

import (
	"bufio"
	"io"
	"sync"

	"github.com/dshills/golevel7/hl7"
)

// Writer provides a streaming interface for writing HL7 messages.
// It buffers writes for efficiency and supports configurable encoding options.
type Writer interface {
	// Write encodes and writes an HL7 message to the underlying writer.
	// The message is encoded according to the configured options.
	Write(msg hl7.Message) error

	// Flush flushes any buffered data to the underlying writer.
	Flush() error

	// Close flushes any remaining data and releases resources.
	// After Close is called, the Writer should not be used.
	Close() error
}

// writer is the concrete implementation of Writer.
type writer struct {
	w      *bufio.Writer
	config encoderConfig
	mu     sync.Mutex
	closed bool
}

// NewWriter creates a new Writer that writes encoded HL7 messages to w.
// The Writer uses buffered I/O for efficiency.
// Options control encoding behavior such as line endings and MLLP framing.
func NewWriter(w io.Writer, opts ...EncoderOption) Writer {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &writer{
		w:      bufio.NewWriter(w),
		config: cfg,
	}
}

// Write encodes and writes an HL7 message to the underlying writer.
// This method is safe for concurrent use.
func (wr *writer) Write(msg hl7.Message) error {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	if wr.closed {
		return &Error{Message: "writer is closed"}
	}

	if msg == nil {
		return &Error{Message: "cannot write nil message"}
	}

	segments := msg.AllSegments()
	if len(segments) == 0 {
		return &Error{Message: "message has no segments"}
	}

	// Get delimiters from the message
	delims := msg.Delimiters()
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	// Add MLLP start block if enabled
	if wr.config.includeMLLP {
		if err := wr.w.WriteByte(MLLPStartBlock); err != nil {
			return &Error{Message: "failed to write MLLP start block", Cause: err}
		}
	}

	lineEndingBytes := []byte(wr.config.lineEnding)

	// Encode each segment
	for i, seg := range segments {
		if i > 0 {
			if _, err := wr.w.Write(lineEndingBytes); err != nil {
				return &Error{
					Message:  "failed to write line ending",
					Segment:  seg.Name(),
					Position: i,
					Cause:    err,
				}
			}
		}

		segBytes := seg.Bytes(delims)
		if _, err := wr.w.Write(segBytes); err != nil {
			return &Error{
				Message:  "failed to write segment",
				Segment:  seg.Name(),
				Position: i,
				Cause:    err,
			}
		}
	}

	// Add final line ending after last segment
	if _, err := wr.w.Write(lineEndingBytes); err != nil {
		return &Error{Message: "failed to write final line ending", Cause: err}
	}

	// Add MLLP end block if enabled
	if wr.config.includeMLLP {
		if _, err := wr.w.Write([]byte{MLLPEndBlock, MLLPCarriageReturn}); err != nil {
			return &Error{Message: "failed to write MLLP end block", Cause: err}
		}
	}

	return nil
}

// Flush flushes any buffered data to the underlying writer.
// This method is safe for concurrent use.
func (wr *writer) Flush() error {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	if wr.closed {
		return &Error{Message: "writer is closed"}
	}

	if err := wr.w.Flush(); err != nil {
		return &Error{Message: "failed to flush buffer", Cause: err}
	}

	return nil
}

// Close flushes any remaining data and marks the writer as closed.
// After Close is called, subsequent Write or Flush calls will return an error.
// This method is safe for concurrent use.
func (wr *writer) Close() error {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	if wr.closed {
		return nil // Already closed, no error
	}

	// Flush any remaining data
	err := wr.w.Flush()
	wr.closed = true

	if err != nil {
		return &Error{Message: "failed to flush on close", Cause: err}
	}

	return nil
}
