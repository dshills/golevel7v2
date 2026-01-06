package mllp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/dshills/golevel7/encode"
	"github.com/dshills/golevel7/hl7"
	"github.com/dshills/golevel7/parse"
)

// Client defines the interface for an MLLP client.
//
// A Client sends HL7 messages to a remote server and receives responses.
// It handles MLLP framing, connection management, and optionally retries.
//
// Clients are safe for concurrent use. Each Send call will use the
// underlying connection, with appropriate synchronization.
type Client interface {
	// Send sends an HL7 message and waits for a response.
	// The context can be used to cancel the operation or set deadlines.
	// Returns the response message (typically an ACK) or an error.
	Send(ctx context.Context, msg hl7.Message) (hl7.Message, error)

	// SendAsync sends an HL7 message without waiting for a response.
	// This is useful for fire-and-forget scenarios where ACKs are not required.
	// The context can be used to cancel the send operation.
	SendAsync(ctx context.Context, msg hl7.Message) error

	// Close closes the connection to the server.
	// After Close is called, subsequent calls to Send will return an error.
	Close() error
}

// client is the concrete implementation of the Client interface.
type client struct {
	addr    string
	config  clientConfig
	conn    net.Conn
	reader  *Reader
	writer  *Writer
	encoder encode.Encoder
	parser  parse.Parser
	mu      sync.Mutex
	closed  bool
}

// NewClient creates a new MLLP client connected to the specified address.
// The address should be in the format "host:port".
//
// Options can be provided to configure timeout, TLS, retry behavior, etc.
// The connection is established lazily on the first Send call.
//
// Example:
//
//	client, err := mllp.NewClient("hospital.local:2575",
//	    mllp.WithTimeout(30*time.Second),
//	    mllp.WithTLS(tlsConfig),
//	)
func NewClient(addr string, opts ...ClientOption) (Client, error) {
	config := defaultClientConfig()
	for _, opt := range opts {
		opt(&config)
	}

	c := &client{
		addr:    addr,
		config:  config,
		encoder: encode.New(),
		parser:  parse.New(),
	}

	return c, nil
}

// Dial creates a new MLLP client and immediately establishes a connection.
// This is a convenience function that combines NewClient with an immediate connect.
//
// Unlike NewClient which connects lazily, Dial will return an error if the
// connection cannot be established.
func Dial(addr string, opts ...ClientOption) (Client, error) {
	c, err := NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}

	// Force immediate connection
	client := c.(*client)
	if err := client.connect(); err != nil {
		return nil, err
	}

	return c, nil
}

// connect establishes the connection to the server.
func (c *client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrConnectionClosed
	}

	if c.conn != nil {
		return nil // Already connected
	}

	var conn net.Conn
	var err error

	dialer := &net.Dialer{
		Timeout:   c.config.timeout,
		KeepAlive: 30 * time.Second,
	}

	if !c.config.keepAlive {
		dialer.KeepAlive = -1 // Disable keep-alive
	}

	if c.config.tlsConfig != nil {
		conn, err = tls.DialWithDialer(dialer, "tcp", c.addr, c.config.tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", c.addr)
	}

	if err != nil {
		return fmt.Errorf("mllp: failed to connect to %s: %w", c.addr, err)
	}

	c.conn = conn
	c.reader = NewReader(conn, MaxMessageSize)
	c.writer = NewWriter(conn)

	return nil
}

// Send sends an HL7 message and waits for a response.
func (c *client) Send(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
	if msg == nil {
		return nil, errors.New("mllp: cannot send nil message")
	}

	// Ensure we're connected
	if err := c.connect(); err != nil {
		return nil, err
	}

	var lastErr error
	attempts := c.config.retryAttempts + 1

	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.config.retryBackoff):
			}

			// Reconnect on retry
			c.mu.Lock()
			if c.conn != nil {
				_ = c.conn.Close()
				c.conn = nil
			}
			c.mu.Unlock()

			if err := c.connect(); err != nil {
				lastErr = err
				continue
			}
		}

		resp, err := c.doSend(ctx, msg)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("mllp: send failed after %d attempts: %w", attempts, lastErr)
}

// doSend performs the actual send/receive operation.
func (c *client) doSend(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, ErrConnectionClosed
	}

	if c.conn == nil {
		return nil, ErrConnectionClosed
	}

	// Set deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetDeadline(deadline)
	} else {
		_ = c.conn.SetDeadline(time.Now().Add(c.config.timeout))
	}
	defer func() { _ = c.conn.SetDeadline(time.Time{}) }() // Clear deadline

	// Encode the message
	data, err := c.encoder.Encode(msg)
	if err != nil {
		return nil, fmt.Errorf("mllp: failed to encode message: %w", err)
	}

	// Send with MLLP framing
	if err := c.writer.WriteMessage(data); err != nil {
		return nil, fmt.Errorf("mllp: failed to send message: %w", err)
	}

	// Read response
	respData, err := c.reader.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("mllp: failed to read response: %w", err)
	}

	// Parse response
	resp, err := c.parser.Parse(respData)
	if err != nil {
		return nil, fmt.Errorf("mllp: failed to parse response: %w", err)
	}

	return resp, nil
}

// SendAsync sends an HL7 message without waiting for a response.
func (c *client) SendAsync(ctx context.Context, msg hl7.Message) error {
	if msg == nil {
		return errors.New("mllp: cannot send nil message")
	}

	// Ensure we're connected
	if err := c.connect(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrConnectionClosed
	}

	if c.conn == nil {
		return ErrConnectionClosed
	}

	// Set deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetWriteDeadline(deadline)
	} else {
		_ = c.conn.SetWriteDeadline(time.Now().Add(c.config.timeout))
	}
	defer func() { _ = c.conn.SetWriteDeadline(time.Time{}) }()

	// Encode the message
	data, err := c.encoder.Encode(msg)
	if err != nil {
		return fmt.Errorf("mllp: failed to encode message: %w", err)
	}

	// Send with MLLP framing
	if err := c.writer.WriteMessage(data); err != nil {
		return fmt.Errorf("mllp: failed to send message: %w", err)
	}

	return nil
}

// Close closes the connection to the server.
func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.reader = nil
		c.writer = nil
		return err
	}

	return nil
}

// Ensure client implements Client at compile time.
var _ Client = (*client)(nil)
