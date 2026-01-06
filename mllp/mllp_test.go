package mllp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/dshills/golevel7/hl7"
)

// TestFrameUnframe tests MLLP framing and unframing functions.
func TestFrameUnframe(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []byte
		wantErr error
	}{
		{
			name:  "simple message",
			input: []byte("MSH|^~\\&|TEST"),
			want:  []byte{0x0B, 'M', 'S', 'H', '|', '^', '~', '\\', '&', '|', 'T', 'E', 'S', 'T', 0x1C, 0x0D},
		},
		{
			name:  "empty message",
			input: []byte{},
			want:  []byte{0x0B, 0x1C, 0x0D},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			framed := Frame(tt.input)
			if !bytes.Equal(framed, tt.want) {
				t.Errorf("Frame() = %v, want %v", framed, tt.want)
			}

			unframed, err := Unframe(framed)
			if err != nil {
				t.Errorf("Unframe() error = %v", err)
				return
			}
			if !bytes.Equal(unframed, tt.input) {
				t.Errorf("Unframe() = %v, want %v", unframed, tt.input)
			}
		})
	}
}

// TestUnframeErrors tests error conditions in Unframe.
func TestUnframeErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr error
	}{
		{
			name:    "missing start block",
			input:   []byte{'M', 'S', 'H', 0x1C, 0x0D},
			wantErr: ErrInvalidStartBlock,
		},
		{
			name:    "missing end block",
			input:   []byte{0x0B, 'M', 'S', 'H'},
			wantErr: ErrInvalidEndBlock,
		},
		{
			name:    "wrong end sequence",
			input:   []byte{0x0B, 'M', 'S', 'H', 0x0D, 0x0D},
			wantErr: ErrInvalidEndBlock,
		},
		{
			name:    "too short",
			input:   []byte{0x0B},
			wantErr: ErrInvalidStartBlock,
		},
		{
			name:    "empty",
			input:   []byte{},
			wantErr: ErrInvalidStartBlock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Unframe(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Unframe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// mockConn implements net.Conn for testing.
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	mu       sync.Mutex
	closed   bool
}

func newMockConn(readData []byte) *mockConn {
	return &mockConn{
		readBuf:  bytes.NewBuffer(readData),
		writeBuf: &bytes.Buffer{},
	}
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, io.EOF
	}
	return c.readBuf.Read(b)
}

func (c *mockConn) Write(b []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, errors.New("connection closed")
	}
	return c.writeBuf.Write(b)
}

func (c *mockConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (c *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (c *mockConn) SetDeadline(_ time.Time) error      { return nil }
func (c *mockConn) SetReadDeadline(_ time.Time) error  { return nil }
func (c *mockConn) SetWriteDeadline(_ time.Time) error { return nil }

func (c *mockConn) Written() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.writeBuf.Bytes()
}

// TestReaderReadMessage tests the MLLP reader.
func TestReaderReadMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []byte
		wantErr error
	}{
		{
			name:  "simple message",
			input: Frame([]byte("MSH|^~\\&|TEST")),
			want:  []byte("MSH|^~\\&|TEST"),
		},
		{
			name:  "message with garbage before",
			input: append([]byte("garbage"), Frame([]byte("MSH|^~\\&|TEST"))...),
			want:  []byte("MSH|^~\\&|TEST"),
		},
		{
			name:    "incomplete message",
			input:   []byte{0x0B, 'M', 'S', 'H'},
			wantErr: io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewReader(bytes.NewReader(tt.input), MaxMessageSize)
			got, err := reader.ReadMessage()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ReadMessage() expected error %v, got nil", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ReadMessage() error = %v", err)
				return
			}

			if !bytes.Equal(got, tt.want) {
				t.Errorf("ReadMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestReaderMaxSize tests message size limits.
func TestReaderMaxSize(t *testing.T) {
	// Create a message larger than the max size
	largeData := make([]byte, 100)
	for i := range largeData {
		largeData[i] = 'A'
	}
	framed := Frame(largeData)

	reader := NewReader(bytes.NewReader(framed), 50)
	_, err := reader.ReadMessage()

	if !errors.Is(err, ErrMessageTooLarge) {
		t.Errorf("ReadMessage() error = %v, want %v", err, ErrMessageTooLarge)
	}
}

// TestWriterWriteMessage tests the MLLP writer.
func TestWriterWriteMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewWriter(buf)

	data := []byte("MSH|^~\\&|TEST")
	err := writer.WriteMessage(data)
	if err != nil {
		t.Errorf("WriteMessage() error = %v", err)
		return
	}

	expected := Frame(data)
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("WriteMessage() wrote %v, want %v", buf.Bytes(), expected)
	}
}

// TestHandlerFunc tests the HandlerFunc adapter.
func TestHandlerFunc(t *testing.T) {
	called := false
	var receivedMsg hl7.Message

	handler := HandlerFunc(func(_ context.Context, msg hl7.Message) (hl7.Message, error) {
		called = true
		receivedMsg = msg
		return msg, nil
	})

	// Create a mock message (nil for this test since we're just checking the call)
	ctx := context.Background()
	resp, err := handler.HandleMessage(ctx, nil)

	if !called {
		t.Error("handler was not called")
	}
	if err != nil {
		t.Errorf("HandleMessage() error = %v", err)
	}
	if receivedMsg != nil {
		t.Error("receivedMsg should be nil")
	}
	if resp != nil {
		t.Error("response should be nil")
	}
}

// TestChainMiddleware tests middleware chaining.
func TestChainMiddleware(t *testing.T) {
	order := []string{}

	middleware1 := func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
			order = append(order, "m1-before")
			resp, err := next.HandleMessage(ctx, msg)
			order = append(order, "m1-after")
			return resp, err
		})
	}

	middleware2 := func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
			order = append(order, "m2-before")
			resp, err := next.HandleMessage(ctx, msg)
			order = append(order, "m2-after")
			return resp, err
		})
	}

	baseHandler := HandlerFunc(func(_ context.Context, _ hl7.Message) (hl7.Message, error) {
		order = append(order, "handler")
		return nil, nil
	})

	chained := Chain(baseHandler, middleware1, middleware2)
	_, _ = chained.HandleMessage(context.Background(), nil)

	expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	if len(order) != len(expected) {
		t.Errorf("Chain() order length = %d, want %d", len(order), len(expected))
		return
	}

	for i, v := range order {
		if v != expected[i] {
			t.Errorf("Chain() order[%d] = %s, want %s", i, v, expected[i])
		}
	}
}

// TestClientOptions tests client configuration options.
func TestClientOptions(t *testing.T) {
	tests := []struct {
		name   string
		opts   []ClientOption
		check  func(*clientConfig) bool
		errMsg string
	}{
		{
			name: "default config",
			opts: nil,
			check: func(c *clientConfig) bool {
				return c.timeout == DefaultTimeout &&
					c.retryAttempts == DefaultRetryAttempts &&
					c.keepAlive == true
			},
			errMsg: "default config values not set correctly",
		},
		{
			name: "with timeout",
			opts: []ClientOption{WithTimeout(5 * time.Second)},
			check: func(c *clientConfig) bool {
				return c.timeout == 5*time.Second
			},
			errMsg: "timeout not set correctly",
		},
		{
			name: "with retry",
			opts: []ClientOption{WithRetry(3, 2*time.Second)},
			check: func(c *clientConfig) bool {
				return c.retryAttempts == 3 && c.retryBackoff == 2*time.Second
			},
			errMsg: "retry config not set correctly",
		},
		{
			name: "with keep alive disabled",
			opts: []ClientOption{WithKeepAlive(false)},
			check: func(c *clientConfig) bool {
				return c.keepAlive == false
			},
			errMsg: "keep alive not set correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := defaultClientConfig()
			for _, opt := range tt.opts {
				opt(&config)
			}
			if !tt.check(&config) {
				t.Error(tt.errMsg)
			}
		})
	}
}

// TestServerOptions tests server configuration options.
func TestServerOptions(t *testing.T) {
	dummyHandler := HandlerFunc(func(_ context.Context, _ hl7.Message) (hl7.Message, error) {
		return nil, nil
	})

	tests := []struct {
		name   string
		opts   []ServerOption
		check  func(*serverConfig) bool
		errMsg string
	}{
		{
			name: "default config",
			opts: nil,
			check: func(c *serverConfig) bool {
				return c.maxConnections == DefaultMaxConnections &&
					c.readTimeout == DefaultReadTimeout &&
					c.writeTimeout == DefaultWriteTimeout &&
					c.handler == nil
			},
			errMsg: "default config values not set correctly",
		},
		{
			name: "with handler",
			opts: []ServerOption{WithHandler(dummyHandler)},
			check: func(c *serverConfig) bool {
				return c.handler != nil
			},
			errMsg: "handler not set correctly",
		},
		{
			name: "with max connections",
			opts: []ServerOption{WithMaxConnections(50)},
			check: func(c *serverConfig) bool {
				return c.maxConnections == 50
			},
			errMsg: "max connections not set correctly",
		},
		{
			name: "with read timeout",
			opts: []ServerOption{WithReadTimeout(120 * time.Second)},
			check: func(c *serverConfig) bool {
				return c.readTimeout == 120*time.Second
			},
			errMsg: "read timeout not set correctly",
		},
		{
			name: "with write timeout",
			opts: []ServerOption{WithWriteTimeout(45 * time.Second)},
			check: func(c *serverConfig) bool {
				return c.writeTimeout == 45*time.Second
			},
			errMsg: "write timeout not set correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := defaultServerConfig()
			for _, opt := range tt.opts {
				opt(&config)
			}
			if !tt.check(&config) {
				t.Error(tt.errMsg)
			}
		})
	}
}

// TestServerNoHandler tests server error when no handler is configured.
func TestServerNoHandler(t *testing.T) {
	server := NewServer()

	// Create a mock listener that returns one connection then closes
	listener := &mockListener{
		conns: []*mockConn{newMockConn([]byte{})},
	}

	err := server.Serve(listener)
	if !errors.Is(err, ErrNoHandler) {
		t.Errorf("Serve() error = %v, want %v", err, ErrNoHandler)
	}
}

// mockListener implements net.Listener for testing.
type mockListener struct {
	conns    []*mockConn
	index    int
	mu       sync.Mutex
	closed   bool
	acceptCh chan struct{}
}

func (l *mockListener) Accept() (net.Conn, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil, errors.New("listener closed")
	}

	if l.index >= len(l.conns) {
		// Block until closed
		l.mu.Unlock()
		if l.acceptCh == nil {
			l.acceptCh = make(chan struct{})
		}
		<-l.acceptCh
		l.mu.Lock()
		return nil, errors.New("listener closed")
	}

	conn := l.conns[l.index]
	l.index++
	return conn, nil
}

func (l *mockListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true
	if l.acceptCh != nil {
		close(l.acceptCh)
	}
	return nil
}

func (l *mockListener) Addr() net.Addr {
	return &net.TCPAddr{}
}

// TestMLLPConstants verifies MLLP framing constants.
func TestMLLPConstants(t *testing.T) {
	if StartBlock != 0x0B {
		t.Errorf("StartBlock = %#x, want %#x", StartBlock, 0x0B)
	}
	if EndBlock != 0x1C {
		t.Errorf("EndBlock = %#x, want %#x", EndBlock, 0x1C)
	}
	if CarriageReturn != 0x0D {
		t.Errorf("CarriageReturn = %#x, want %#x", CarriageReturn, 0x0D)
	}
}

// TestReaderMultipleMessages tests reading multiple messages.
func TestReaderMultipleMessages(t *testing.T) {
	msg1 := Frame([]byte("MSG1"))
	msg2 := Frame([]byte("MSG2"))
	combined := make([]byte, 0, len(msg1)+len(msg2))
	combined = append(combined, msg1...)
	combined = append(combined, msg2...)

	reader := NewReader(bytes.NewReader(combined), MaxMessageSize)

	// Read first message
	got1, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() 1 error = %v", err)
	}
	if string(got1) != "MSG1" {
		t.Errorf("ReadMessage() 1 = %q, want %q", got1, "MSG1")
	}

	// Read second message
	got2, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() 2 error = %v", err)
	}
	if string(got2) != "MSG2" {
		t.Errorf("ReadMessage() 2 = %q, want %q", got2, "MSG2")
	}

	// Third read should return EOF
	_, err = reader.ReadMessage()
	if err != io.EOF {
		t.Errorf("ReadMessage() 3 error = %v, want EOF", err)
	}
}

// TestNewReaderDefaultMaxSize tests that NewReader uses default max size when 0 is passed.
func TestNewReaderDefaultMaxSize(t *testing.T) {
	reader := NewReader(bytes.NewReader([]byte{}), 0)
	if reader.maxSize != MaxMessageSize {
		t.Errorf("NewReader(0) maxSize = %d, want %d", reader.maxSize, MaxMessageSize)
	}
}

// TestNewReaderNegativeMaxSize tests that NewReader uses default max size when negative is passed.
func TestNewReaderNegativeMaxSize(t *testing.T) {
	reader := NewReader(bytes.NewReader([]byte{}), -1)
	if reader.maxSize != MaxMessageSize {
		t.Errorf("NewReader(-1) maxSize = %d, want %d", reader.maxSize, MaxMessageSize)
	}
}

// TestReaderFalseEndBlock tests a message with end block character in the data.
func TestReaderFalseEndBlock(t *testing.T) {
	// Message contains EndBlock followed by non-CR character (not a valid end sequence)
	data := []byte{StartBlock, 'A', EndBlock, 'B', EndBlock, CarriageReturn}
	reader := NewReader(bytes.NewReader(data), MaxMessageSize)

	got, err := reader.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	// The message should include "A", EndBlock, "B" since EndBlock-B is not a valid terminator
	expected := []byte{'A', EndBlock, 'B'}
	if !bytes.Equal(got, expected) {
		t.Errorf("ReadMessage() = %v, want %v", got, expected)
	}
}

// TestWriterWriteErrors tests error handling in Writer.
type failingWriter struct {
	failAfter int
	written   int
}

func (w *failingWriter) Write(b []byte) (int, error) {
	if w.written >= w.failAfter {
		return 0, errors.New("write failed")
	}
	w.written += len(b)
	return len(b), nil
}

func TestWriterWriteStartBlockError(t *testing.T) {
	w := &failingWriter{failAfter: 0}
	writer := NewWriter(w)

	err := writer.WriteMessage([]byte("test"))
	if err == nil {
		t.Error("WriteMessage() expected error, got nil")
	}
}

func TestWriterWriteDataError(t *testing.T) {
	w := &failingWriter{failAfter: 1}
	writer := NewWriter(w)

	err := writer.WriteMessage([]byte("test"))
	if err == nil {
		t.Error("WriteMessage() expected error, got nil")
	}
}

func TestWriterWriteEndBlockError(t *testing.T) {
	w := &failingWriter{failAfter: 5}
	writer := NewWriter(w)

	err := writer.WriteMessage([]byte("test"))
	if err == nil {
		t.Error("WriteMessage() expected error, got nil")
	}
}

// TestClientClose tests client close behavior.
func TestClientClose(t *testing.T) {
	client, err := NewClient("localhost:2575")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Close without connecting
	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Close again should be no-op
	err = client.Close()
	if err != nil {
		t.Errorf("Close() second call error = %v", err)
	}
}

// TestClientSendNilMessage tests sending nil message.
func TestClientSendNilMessage(t *testing.T) {
	client, err := NewClient("localhost:2575")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	_, err = client.Send(context.Background(), nil)
	if err == nil {
		t.Error("Send(nil) expected error, got nil")
	}
}

// TestClientSendAsyncNilMessage tests sending nil message async.
func TestClientSendAsyncNilMessage(t *testing.T) {
	client, err := NewClient("localhost:2575")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	err = client.SendAsync(context.Background(), nil)
	if err == nil {
		t.Error("SendAsync(nil) expected error, got nil")
	}
}

// TestOptionValidation tests option validation.
func TestOptionValidation(t *testing.T) {
	// Test WithTimeout with negative duration doesn't change default
	config := defaultClientConfig()
	WithTimeout(-1 * time.Second)(&config)
	if config.timeout != DefaultTimeout {
		t.Errorf("WithTimeout(-1) changed timeout to %v", config.timeout)
	}

	// Test WithRetry with negative attempts doesn't change default
	config = defaultClientConfig()
	WithRetry(-1, time.Second)(&config)
	if config.retryAttempts != DefaultRetryAttempts {
		t.Errorf("WithRetry(-1) changed retryAttempts to %v", config.retryAttempts)
	}

	// Test WithRetry with negative backoff doesn't change default
	config = defaultClientConfig()
	WithRetry(3, -1*time.Second)(&config)
	if config.retryBackoff != DefaultRetryBackoff {
		t.Errorf("WithRetry with negative backoff changed retryBackoff to %v", config.retryBackoff)
	}

	// Test server options validation
	serverCfg := defaultServerConfig()
	WithMaxConnections(-1)(&serverCfg)
	if serverCfg.maxConnections != DefaultMaxConnections {
		t.Errorf("WithMaxConnections(-1) changed maxConnections to %v", serverCfg.maxConnections)
	}

	WithReadTimeout(-1 * time.Second)(&serverCfg)
	if serverCfg.readTimeout != DefaultReadTimeout {
		t.Errorf("WithReadTimeout(-1) changed readTimeout to %v", serverCfg.readTimeout)
	}

	WithWriteTimeout(-1 * time.Second)(&serverCfg)
	if serverCfg.writeTimeout != DefaultWriteTimeout {
		t.Errorf("WithWriteTimeout(-1) changed writeTimeout to %v", serverCfg.writeTimeout)
	}
}

// TestWithTLS tests the TLS configuration options.
func TestWithTLS(t *testing.T) {
	config := defaultClientConfig()
	WithTLS(nil)(&config)
	if config.tlsConfig != nil {
		t.Error("WithTLS(nil) should set tlsConfig to nil")
	}
}

// TestWithTLSConfig tests the TLS configuration for server.
func TestWithTLSConfig(t *testing.T) {
	serverCfg := defaultServerConfig()
	WithTLSConfig(nil)(&serverCfg)
	if serverCfg.tlsConfig != nil {
		t.Error("WithTLSConfig(nil) should set tlsConfig to nil")
	}
}

// TestServerShutdown tests server shutdown behavior.
func TestServerShutdown(t *testing.T) {
	handler := HandlerFunc(func(_ context.Context, _ hl7.Message) (hl7.Message, error) {
		return nil, nil
	})
	server := NewServer(WithHandler(handler))

	// Shutdown before serving should work
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() before Serve() error = %v", err)
	}

	// Calling shutdown multiple times should be safe
	err = server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() second call error = %v", err)
	}
}

// TestChainEmptyMiddleware tests chaining with no middleware.
func TestChainEmptyMiddleware(t *testing.T) {
	called := false
	handler := HandlerFunc(func(_ context.Context, _ hl7.Message) (hl7.Message, error) {
		called = true
		return nil, nil
	})

	chained := Chain(handler)
	_, _ = chained.HandleMessage(context.Background(), nil)

	if !called {
		t.Error("Chain with no middleware should call handler")
	}
}
