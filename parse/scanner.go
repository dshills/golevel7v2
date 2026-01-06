// Package parse provides HL7 v2.x message parsing functionality.
package parse

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/dshills/golevel7/hl7"
)

// Scanner-specific errors.
var (
	// ErrMessageTooLarge is returned when a message exceeds the maximum size.
	ErrMessageTooLarge = errors.New("message exceeds maximum size")
)

// Default scanner configuration values.
const (
	defaultMaxMessageSize = 10 * 1024 * 1024 // 10 MB max message size
	defaultBufferSize     = 64 * 1024        // 64 KB buffer
)

// Scanner provides streaming HL7 message parsing from an io.Reader.
// It supports both MLLP-framed messages and plain CR-delimited messages.
type Scanner interface {
	// Scan advances to the next message. Returns true if a message was found.
	Scan() bool

	// Message returns the last parsed message.
	// Returns nil if Scan hasn't been called or returned false.
	Message() hl7.Message

	// Err returns any error encountered during scanning.
	// Returns nil if no error occurred.
	Err() error
}

// scanner is the concrete implementation of Scanner.
type scanner struct {
	reader         *bufio.Reader
	parser         Parser
	config         parserConfig
	message        hl7.Message
	err            error
	maxMessageSize int
	pending        []byte // bytes read ahead that belong to the next message
}

// ScannerOption is a functional option for configuring the scanner.
type ScannerOption func(*scanner)

// WithMaxMessageSize sets the maximum allowed message size in bytes.
// Default is 10 MB.
func WithMaxMessageSize(size int) ScannerOption {
	return func(s *scanner) {
		if size > 0 {
			s.maxMessageSize = size
		}
	}
}

// NewScanner creates a new Scanner that reads from the given io.Reader.
// The scanner will parse messages using the provided ParserOptions.
func NewScanner(r io.Reader, opts ...ParserOption) Scanner {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &scanner{
		reader:         bufio.NewReaderSize(r, defaultBufferSize),
		parser:         New(opts...),
		config:         cfg,
		maxMessageSize: defaultMaxMessageSize,
	}
}

// NewScannerWithOptions creates a new Scanner with additional scanner-specific options.
func NewScannerWithOptions(r io.Reader, parserOpts []ParserOption, scannerOpts ...ScannerOption) Scanner {
	s := NewScanner(r, parserOpts...).(*scanner)
	for _, opt := range scannerOpts {
		opt(s)
	}
	return s
}

// Scan advances to the next message.
func (s *scanner) Scan() bool {
	s.message = nil

	// Read message data
	data, err := s.readMessage()
	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return false
	}

	if len(data) == 0 {
		return false
	}

	// Parse the message
	msg, err := s.parser.Parse(data)
	if err != nil {
		s.err = err
		return false
	}

	s.message = msg
	return true
}

// Message returns the last parsed message.
func (s *scanner) Message() hl7.Message {
	return s.message
}

// Err returns any error encountered during scanning.
func (s *scanner) Err() error {
	return s.err
}

// readMessage reads a complete HL7 message from the reader.
// It handles both MLLP-framed and plain CR-delimited messages.
func (s *scanner) readMessage() ([]byte, error) {
	// Check if we have pending data from previous read
	var firstByte byte
	if len(s.pending) > 0 {
		firstByte = s.pending[0]
	} else {
		// Peek at the first byte to determine framing mode
		peeked, err := s.reader.Peek(1)
		if err != nil {
			return nil, err
		}
		firstByte = peeked[0]
	}

	// Check if this is MLLP framing
	if firstByte == mllpStartByte {
		return s.readMLLPMessage()
	}

	// Plain message (CR-delimited segments, double CR or EOF ends message)
	return s.readPlainMessage()
}

// readMLLPMessage reads an MLLP-framed message.
// MLLP format: <VT>message<FS><CR> where VT=0x0B, FS=0x1C, CR=0x0D
func (s *scanner) readMLLPMessage() ([]byte, error) {
	// Handle pending data
	if len(s.pending) > 0 && s.pending[0] == mllpStartByte {
		s.pending = s.pending[1:]
	} else {
		// Read and discard the start byte (0x0B)
		_, err := s.reader.ReadByte()
		if err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	size := 0

	// First, drain any pending data
	for len(s.pending) > 0 {
		b := s.pending[0]
		s.pending = s.pending[1:]

		if b == mllpEndByte1 {
			// Check for trailing CR in pending
			if len(s.pending) > 0 && s.pending[0] == mllpEndByte2 {
				s.pending = s.pending[1:]
			}
			return buf.Bytes(), nil
		}

		size++
		if size > s.maxMessageSize {
			return nil, ErrMessageTooLarge
		}
		buf.WriteByte(b)
	}

	for {
		b, err := s.reader.ReadByte()
		if err != nil {
			return nil, err
		}

		// Check for FS (end of message marker)
		if b == mllpEndByte1 {
			// Read the trailing CR if present
			nextByte, err := s.reader.Peek(1)
			if err == nil && len(nextByte) > 0 && nextByte[0] == mllpEndByte2 {
				_, _ = s.reader.ReadByte() // Consume the CR
			}
			break
		}

		size++
		if size > s.maxMessageSize {
			return nil, ErrMessageTooLarge
		}

		buf.WriteByte(b)
	}

	return buf.Bytes(), nil
}

// readPlainMessage reads a plain (non-MLLP) HL7 message.
// Messages are terminated by:
// - Double CR (\r\r)
// - Start of a new MSH segment after a CR
// - EOF
func (s *scanner) readPlainMessage() ([]byte, error) {
	var buf bytes.Buffer
	size := 0
	terminator := byte(s.config.segmentTerminator)

	// First, drain any pending data
	for i := 0; i < len(s.pending); i++ {
		b := s.pending[i]

		// Check for double CR
		if b == terminator && i+1 < len(s.pending) && s.pending[i+1] == terminator {
			buf.Write(s.pending[:i])
			s.pending = s.pending[i+2:]
			return buf.Bytes(), nil
		}

		// Check for new MSH
		if b == terminator && i+4 < len(s.pending) {
			if s.pending[i+1] == 'M' && s.pending[i+2] == 'S' && s.pending[i+3] == 'H' {
				buf.Write(s.pending[:i])
				s.pending = s.pending[i+1:]
				return buf.Bytes(), nil
			}
		}
	}

	// Add all pending data to buffer
	if len(s.pending) > 0 {
		buf.Write(s.pending)
		size = len(s.pending)
		s.pending = nil
	}

	for {
		b, err := s.reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				// Return what we have if there's any data
				if buf.Len() > 0 {
					return buf.Bytes(), nil
				}
				return nil, io.EOF
			}
			return nil, err
		}

		// Check for CR followed by another CR (double CR = message separator)
		if b == terminator {
			// Peek to see what's next
			peek, peekErr := s.reader.Peek(1)
			if peekErr == nil && len(peek) > 0 {
				if peek[0] == terminator {
					// Double CR - end of message
					_, _ = s.reader.ReadByte() // consume the second CR
					return buf.Bytes(), nil
				}
				if peek[0] == 'M' {
					// Peek further to see if it's MSH
					peek3, peekErr := s.reader.Peek(3)
					if peekErr == nil && len(peek3) >= 3 && peek3[0] == 'M' && peek3[1] == 'S' && peek3[2] == 'H' {
						// New message starting - current message ends here
						return buf.Bytes(), nil
					}
				}
			}
		}

		size++
		if size > s.maxMessageSize {
			return nil, ErrMessageTooLarge
		}

		buf.WriteByte(b)
	}
}
