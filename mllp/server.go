package mllp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dshills/golevel7/encode"
	"github.com/dshills/golevel7/parse"
)

// Server defines the interface for an MLLP server.
//
// A Server listens for incoming TCP connections, reads MLLP-framed HL7
// messages, passes them to a handler, and sends back the response.
//
// The server handles multiple concurrent connections and provides
// graceful shutdown capabilities.
type Server interface {
	// Serve accepts incoming connections on the listener and handles them.
	// This method blocks until the listener is closed or Shutdown is called.
	// Returns ErrServerClosed after graceful shutdown.
	Serve(listener net.Listener) error

	// Shutdown gracefully shuts down the server.
	// It stops accepting new connections and waits for existing connections
	// to complete or for the context to be canceled.
	Shutdown(ctx context.Context) error
}

// server is the concrete implementation of the Server interface.
type server struct {
	config       serverConfig
	encoder      encode.Encoder
	parser       parse.Parser
	listener     net.Listener
	connections  map[net.Conn]struct{}
	connMu       sync.Mutex
	activeConns  atomic.Int32
	shutdown     atomic.Bool
	shutdownChan chan struct{}
	shutdownOnce sync.Once
	wg           sync.WaitGroup
}

// NewServer creates a new MLLP server with the provided options.
//
// At minimum, a handler must be configured using WithHandler.
//
// Example:
//
//	handler := mllp.HandlerFunc(func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
//	    // Process message and return ACK
//	    return createACK(msg), nil
//	})
//
//	server := mllp.NewServer(
//	    mllp.WithHandler(handler),
//	    mllp.WithMaxConnections(100),
//	    mllp.WithReadTimeout(60*time.Second),
//	)
//
//	listener, _ := net.Listen("tcp", ":2575")
//	log.Fatal(server.Serve(listener))
func NewServer(opts ...ServerOption) Server {
	config := defaultServerConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return &server{
		config:       config,
		encoder:      encode.New(),
		parser:       parse.New(),
		connections:  make(map[net.Conn]struct{}),
		shutdownChan: make(chan struct{}),
	}
}

// Serve accepts incoming connections and handles them.
func (s *server) Serve(listener net.Listener) error {
	if s.config.handler == nil {
		return ErrNoHandler
	}

	// Wrap with TLS if configured
	if s.config.tlsConfig != nil {
		listener = tls.NewListener(listener, s.config.tlsConfig)
	}

	s.listener = listener

	for {
		// Check if we're shutting down
		if s.shutdown.Load() {
			return ErrServerClosed
		}

		conn, err := listener.Accept()
		if err != nil {
			// Check if this is due to shutdown
			if s.shutdown.Load() {
				return ErrServerClosed
			}

			// Check for temporary errors
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}

			return fmt.Errorf("mllp: accept error: %w", err)
		}

		// Check connection limit
		if s.activeConns.Load() >= int32(s.config.maxConnections) {
			_ = conn.Close()
			continue
		}

		s.wg.Add(1)
		s.activeConns.Add(1)

		s.connMu.Lock()
		s.connections[conn] = struct{}{}
		s.connMu.Unlock()

		go s.handleConnection(conn)
	}
}

// handleConnection processes messages from a single client connection.
func (s *server) handleConnection(conn net.Conn) {
	defer func() {
		s.connMu.Lock()
		delete(s.connections, conn)
		s.connMu.Unlock()

		s.activeConns.Add(-1)
		_ = conn.Close()
		s.wg.Done()
	}()

	reader := NewReader(conn, MaxMessageSize)
	writer := NewWriter(conn)

	for {
		// Check for shutdown
		if s.shutdown.Load() {
			return
		}

		// Set read deadline
		if s.config.readTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(s.config.readTimeout))
		}

		// Read next message
		data, err := reader.ReadMessage()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, ErrConnectionClosed) {
				return // Client disconnected
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				return // Read timeout, close connection
			}
			// Log error and continue or close depending on severity
			return
		}

		// Parse the message
		msg, err := s.parser.Parse(data)
		if err != nil {
			// Send NAK if possible, then continue
			// For now, just log and continue
			continue
		}

		// Create context for handler
		ctx, cancel := context.WithCancel(context.Background())

		// Handle message
		resp, err := s.config.handler.HandleMessage(ctx, msg)
		cancel()

		if err != nil {
			// Handler returned error - could send NAK here
			continue
		}

		if resp == nil {
			continue // No response to send
		}

		// Set write deadline
		if s.config.writeTimeout > 0 {
			_ = conn.SetWriteDeadline(time.Now().Add(s.config.writeTimeout))
		}

		// Encode response
		respData, err := s.encoder.Encode(resp)
		if err != nil {
			continue
		}

		// Send response with MLLP framing
		if err := writer.WriteMessage(respData); err != nil {
			return // Write error, close connection
		}
	}
}

// Shutdown gracefully shuts down the server.
func (s *server) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		s.shutdown.Store(true)
		close(s.shutdownChan)

		// Close the listener to stop accepting new connections
		if s.listener != nil {
			_ = s.listener.Close()
		}
	})

	// Wait for all connections to complete or context to be canceled
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// Force close all connections
		s.connMu.Lock()
		for conn := range s.connections {
			_ = conn.Close()
		}
		s.connMu.Unlock()

		// Wait a short time for goroutines to clean up
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}

		return ctx.Err()
	}
}

// ActiveConnections returns the number of active client connections.
func (s *server) ActiveConnections() int {
	return int(s.activeConns.Load())
}

// Ensure server implements Server at compile time.
var _ Server = (*server)(nil)
