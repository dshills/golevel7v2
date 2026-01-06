// Package mllp provides MLLP (Minimal Lower Layer Protocol) support for HL7 v2.x.
//
// MLLP is the standard transport protocol for HL7 messages over TCP/IP. It defines
// a simple framing mechanism using control characters to delimit message boundaries.
//
// # MLLP Frame Format
//
// An MLLP frame consists of:
//   - Start Block: 0x0B (vertical tab, VT)
//   - HL7 Message Data
//   - End Block: 0x1C (file separator, FS)
//   - Carriage Return: 0x0D (CR)
//
// Frame structure:
//
//	<VT>...HL7 Message Data...<FS><CR>
//	 |                        |   |
//	 0x0B                   0x1C 0x0D
//
// # Server Usage
//
// Create an MLLP server to receive HL7 messages:
//
//	// Define message handler
//	handler := func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
//	    // Process the message
//	    log.Printf("Received: %s", msg.Type())
//
//	    // Generate acknowledgment
//	    return ack.Generate(msg, ack.Accept())
//	}
//
//	// Create and start server
//	server := mllp.NewServer(":2575", handler)
//	if err := server.ListenAndServe(); err != nil {
//	    log.Fatal(err)
//	}
//
// Server with options:
//
//	server := mllp.NewServer(":2575", handler,
//	    mllp.WithReadTimeout(30*time.Second),
//	    mllp.WithWriteTimeout(30*time.Second),
//	    mllp.WithMaxMessageSize(1024*1024),  // 1MB max
//	    mllp.WithTLS(tlsConfig),
//	)
//
// # Client Usage
//
// Create an MLLP client to send HL7 messages:
//
//	// Connect to server
//	client, err := mllp.Dial("localhost:2575")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Send message and receive ACK
//	ackMsg, err := client.Send(ctx, msg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Check acknowledgment
//	ackCode, _ := ackMsg.Get("MSA.1")
//	if ackCode != "AA" {
//	    log.Printf("Message not accepted: %s", ackCode)
//	}
//
// Client with options:
//
//	client, err := mllp.Dial("localhost:2575",
//	    mllp.WithDialTimeout(10*time.Second),
//	    mllp.WithClientTLS(tlsConfig),
//	    mllp.WithRetry(3, time.Second),
//	)
//
// # Connection Pool
//
// For high-throughput scenarios, use a connection pool:
//
//	pool := mllp.NewPool("localhost:2575",
//	    mllp.WithPoolSize(10),
//	    mllp.WithPoolTimeout(5*time.Second),
//	)
//	defer pool.Close()
//
//	// Send message using pooled connection
//	ackMsg, err := pool.Send(ctx, msg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Reading and Writing Frames
//
// For low-level control, use the Reader and Writer types:
//
// Reading MLLP frames:
//
//	reader := mllp.NewReader(conn)
//	for {
//	    data, err := reader.ReadFrame()
//	    if err != nil {
//	        if err == io.EOF {
//	            break
//	        }
//	        log.Fatal(err)
//	    }
//	    // data contains the unwrapped HL7 message
//	    msg, _ := parser.Parse(data)
//	}
//
// Writing MLLP frames:
//
//	writer := mllp.NewWriter(conn)
//	if err := writer.WriteFrame(hl7Data); err != nil {
//	    log.Fatal(err)
//	}
//
// # Frame Validation
//
// Validate MLLP frame boundaries:
//
//	if mllp.IsValidFrame(data) {
//	    // Data has proper MLLP framing
//	}
//
//	// Extract message from frame
//	msg, err := mllp.Unwrap(framedData)
//	if err != nil {
//	    log.Fatal("invalid MLLP frame:", err)
//	}
//
//	// Wrap message in MLLP frame
//	framedData := mllp.Wrap(hl7Data)
//
// # Error Handling
//
// MLLP operations return specific error types:
//
//	ackMsg, err := client.Send(ctx, msg)
//	if err != nil {
//	    var connErr *mllp.ConnectionError
//	    var frameErr *mllp.FrameError
//	    var timeoutErr *mllp.TimeoutError
//
//	    switch {
//	    case errors.As(err, &connErr):
//	        log.Printf("Connection failed: %v", connErr)
//	        // Retry with new connection
//	    case errors.As(err, &frameErr):
//	        log.Printf("Invalid frame: %v", frameErr)
//	        // Protocol error
//	    case errors.As(err, &timeoutErr):
//	        log.Printf("Timeout: %v", timeoutErr)
//	        // Consider retry
//	    default:
//	        log.Printf("Unknown error: %v", err)
//	    }
//	}
//
// # TLS Support
//
// Enable TLS for secure connections:
//
// Server:
//
//	cert, _ := tls.LoadX509KeyPair("server.crt", "server.key")
//	tlsConfig := &tls.Config{
//	    Certificates: []tls.Certificate{cert},
//	    MinVersion:   tls.VersionTLS12,
//	}
//
//	server := mllp.NewServer(":2575", handler,
//	    mllp.WithTLS(tlsConfig),
//	)
//
// Client:
//
//	tlsConfig := &tls.Config{
//	    InsecureSkipVerify: false,
//	    MinVersion:         tls.VersionTLS12,
//	}
//
//	client, _ := mllp.Dial("localhost:2575",
//	    mllp.WithClientTLS(tlsConfig),
//	)
//
// # Graceful Shutdown
//
// Properly shutdown server connections:
//
//	// Create server with context
//	ctx, cancel := context.WithCancel(context.Background())
//	server := mllp.NewServer(":2575", handler)
//
//	// Start server in goroutine
//	go func() {
//	    if err := server.ListenAndServe(); err != nil {
//	        log.Printf("Server stopped: %v", err)
//	    }
//	}()
//
//	// Handle shutdown signal
//	sigCh := make(chan os.Signal, 1)
//	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
//	<-sigCh
//
//	// Graceful shutdown with timeout
//	shutdownCtx, _ := context.WithTimeout(ctx, 30*time.Second)
//	if err := server.Shutdown(shutdownCtx); err != nil {
//	    log.Printf("Shutdown error: %v", err)
//	}
//
// # Example: Complete MLLP Service
//
//	func main() {
//	    parser := parse.New()
//	    encoder := encode.New()
//
//	    handler := func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
//	        // Log message receipt
//	        log.Printf("Received %s from %s",
//	            msg.Type(), msg.Get("MSH.3"))
//
//	        // Validate message
//	        validator := validate.NewValidator(
//	            validate.Required("MSH.9"),
//	            validate.Required("PID.3"),
//	        )
//	        if errs := validator.Validate(msg); len(errs) > 0 {
//	            return ack.Generate(msg,
//	                ack.Error("Validation failed"),
//	                ack.WithText(errs[0].Error()),
//	            )
//	        }
//
//	        // Process based on message type
//	        switch msg.Type() {
//	        case "ADT^A01":
//	            if err := handleAdmit(msg); err != nil {
//	                return ack.Generate(msg, ack.Error(err.Error()))
//	            }
//	        case "ORU^R01":
//	            if err := handleLabResult(msg); err != nil {
//	                return ack.Generate(msg, ack.Error(err.Error()))
//	            }
//	        default:
//	            return ack.Generate(msg,
//	                ack.Reject("Unsupported message type"))
//	        }
//
//	        return ack.Generate(msg, ack.Accept())
//	    }
//
//	    server := mllp.NewServer(":2575", handler,
//	        mllp.WithReadTimeout(60*time.Second),
//	        mllp.WithWriteTimeout(30*time.Second),
//	    )
//
//	    log.Println("Starting MLLP server on :2575")
//	    if err := server.ListenAndServe(); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// # Constants
//
// MLLP framing constants are exported for custom implementations:
//
//	mllp.StartBlock      // 0x0B - vertical tab
//	mllp.EndBlock        // 0x1C - file separator
//	mllp.CarriageReturn  // 0x0D - carriage return
package mllp
