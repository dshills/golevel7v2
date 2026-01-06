// Package mllp provides MLLP (Minimal Lower Layer Protocol) client and server
// implementations for HL7 v2.x message transport over TCP/IP.
//
// MLLP is the standard framing protocol for HL7 v2.x messages transmitted over
// TCP/IP connections. It wraps each message with a start block (0x0B) and end
// block (0x1C 0x0D) to delimit message boundaries in the TCP stream.
//
// Example client usage:
//
//	client, err := mllp.NewClient("hospital.local:2575",
//	    mllp.WithTimeout(30*time.Second),
//	    mllp.WithRetry(3, time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	response, err := client.Send(ctx, message)
//
// Example server usage:
//
//	server := mllp.NewServer(
//	    mllp.WithHandler(myHandler),
//	    mllp.WithMaxConnections(100),
//	)
//
//	listener, _ := net.Listen("tcp", ":2575")
//	if err := server.Serve(listener); err != nil {
//	    log.Fatal(err)
//	}
package mllp

import (
	"crypto/tls"
	"time"
)

// Default configuration values for MLLP clients and servers.
const (
	// DefaultTimeout is the default timeout for client operations.
	DefaultTimeout = 30 * time.Second

	// DefaultReadTimeout is the default timeout for reading messages.
	DefaultReadTimeout = 60 * time.Second

	// DefaultWriteTimeout is the default timeout for writing messages.
	DefaultWriteTimeout = 30 * time.Second

	// DefaultMaxConnections is the default maximum number of concurrent connections.
	DefaultMaxConnections = 100

	// DefaultRetryAttempts is the default number of retry attempts.
	DefaultRetryAttempts = 0

	// DefaultRetryBackoff is the default backoff duration between retries.
	DefaultRetryBackoff = time.Second
)

// clientConfig holds the configuration for an MLLP client.
type clientConfig struct {
	timeout       time.Duration
	retryAttempts int
	retryBackoff  time.Duration
	tlsConfig     *tls.Config
	keepAlive     bool
}

// defaultClientConfig returns a clientConfig with default values.
func defaultClientConfig() clientConfig {
	return clientConfig{
		timeout:       DefaultTimeout,
		retryAttempts: DefaultRetryAttempts,
		retryBackoff:  DefaultRetryBackoff,
		tlsConfig:     nil,
		keepAlive:     true,
	}
}

// ClientOption is a functional option for configuring an MLLP client.
type ClientOption func(*clientConfig)

// WithTimeout sets the timeout for client operations including connect,
// send, and receive. Default is 30 seconds.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// WithRetry configures retry behavior for failed send operations.
// The client will retry up to 'attempts' times with 'backoff' duration
// between each attempt. Default is no retries (attempts=0).
func WithRetry(attempts int, backoff time.Duration) ClientOption {
	return func(c *clientConfig) {
		if attempts >= 0 {
			c.retryAttempts = attempts
		}
		if backoff > 0 {
			c.retryBackoff = backoff
		}
	}
}

// WithTLS configures TLS for secure connections.
// Pass nil to disable TLS (default). When enabled, the client will use
// TLS to encrypt the connection to the server.
func WithTLS(config *tls.Config) ClientOption {
	return func(c *clientConfig) {
		c.tlsConfig = config
	}
}

// WithKeepAlive enables or disables TCP keep-alive for the connection.
// When enabled (default), the operating system will periodically send
// keep-alive probes to detect broken connections.
func WithKeepAlive(enable bool) ClientOption {
	return func(c *clientConfig) {
		c.keepAlive = enable
	}
}

// serverConfig holds the configuration for an MLLP server.
type serverConfig struct {
	handler        Handler
	maxConnections int
	readTimeout    time.Duration
	writeTimeout   time.Duration
	tlsConfig      *tls.Config
}

// defaultServerConfig returns a serverConfig with default values.
func defaultServerConfig() serverConfig {
	return serverConfig{
		handler:        nil,
		maxConnections: DefaultMaxConnections,
		readTimeout:    DefaultReadTimeout,
		writeTimeout:   DefaultWriteTimeout,
		tlsConfig:      nil,
	}
}

// ServerOption is a functional option for configuring an MLLP server.
type ServerOption func(*serverConfig)

// WithHandler sets the message handler for the server.
// The handler's HandleMessage method will be called for each incoming message.
// This option is required - the server will reject messages without a handler.
func WithHandler(h Handler) ServerOption {
	return func(c *serverConfig) {
		c.handler = h
	}
}

// WithMaxConnections sets the maximum number of concurrent client connections.
// New connections will be rejected once this limit is reached.
// Default is 100 connections.
func WithMaxConnections(limit int) ServerOption {
	return func(c *serverConfig) {
		if limit > 0 {
			c.maxConnections = limit
		}
	}
}

// WithReadTimeout sets the timeout for reading messages from clients.
// The timeout is reset each time a complete message is received.
// Default is 60 seconds.
func WithReadTimeout(d time.Duration) ServerOption {
	return func(c *serverConfig) {
		if d > 0 {
			c.readTimeout = d
		}
	}
}

// WithWriteTimeout sets the timeout for writing responses to clients.
// Default is 30 seconds.
func WithWriteTimeout(d time.Duration) ServerOption {
	return func(c *serverConfig) {
		if d > 0 {
			c.writeTimeout = d
		}
	}
}

// WithTLSConfig configures TLS for secure server connections.
// Pass nil to disable TLS (default). When enabled, the server will
// require TLS connections from all clients.
func WithTLSConfig(config *tls.Config) ServerOption {
	return func(c *serverConfig) {
		c.tlsConfig = config
	}
}
