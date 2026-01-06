// Package encode provides HL7 v2.x message encoding functionality.
//
// The encode package converts structured [hl7.Message] objects back to their
// wire format representation. It supports configurable line endings, optional
// MLLP (Minimal Lower Layer Protocol) framing, and streaming to io.Writer.
//
// # Basic Usage
//
// Encode a message to bytes:
//
//	enc := encode.New()
//	data, err := enc.Encode(msg)
//	if err != nil {
//	    log.Fatal("encode error:", err)
//	}
//	// data contains the HL7 message as bytes
//
// Encode directly to a writer (e.g., network connection):
//
//	ctx := context.Background()
//	err := enc.EncodeToWriter(ctx, conn, msg)
//	if err != nil {
//	    log.Fatal("encode error:", err)
//	}
//
// # Encoder Options
//
// The encoder supports functional options for configuration:
//
//	// Use CRLF line endings (for Windows compatibility)
//	enc := encode.New(encode.WithLineEnding("\r\n"))
//
//	// Enable MLLP framing for TCP transmission
//	enc := encode.New(encode.WithMLLP(true))
//
//	// Include trailing delimiters
//	enc := encode.New(encode.WithTrailingDelimiters(true))
//
//	// Combine multiple options
//	enc := encode.New(
//	    encode.WithMLLP(true),
//	    encode.WithLineEnding("\r"),
//	)
//
// # Line Endings
//
// HL7 v2.x specifies carriage return (CR, 0x0D) as the segment terminator.
// Some systems require different line endings:
//
//	// Standard HL7 (default)
//	enc := encode.New(encode.WithLineEnding("\r"))
//
//	// Windows-style CRLF
//	enc := encode.New(encode.WithLineEnding("\r\n"))
//
//	// Unix-style LF
//	enc := encode.New(encode.WithLineEnding("\n"))
//
// # MLLP Framing
//
// MLLP (Minimal Lower Layer Protocol) is the standard transport protocol
// for HL7 over TCP/IP. When enabled, messages are wrapped with:
//   - Start block: 0x0B (vertical tab)
//   - End block: 0x1C 0x0D (file separator + carriage return)
//
// Example with MLLP framing:
//
//	enc := encode.New(encode.WithMLLP(true))
//	data, _ := enc.Encode(msg)
//	// data starts with 0x0B, ends with 0x1C 0x0D
//
// MLLP frame structure:
//
//	<VT>MSH|^~\&|...<CR>PID|...<CR>...<FS><CR>
//	^                                     ^  ^
//	|                                     |  |
//	Start Block (0x0B)          End Block-+  +- CR (0x0D)
//	                               (0x1C)
//
// # Streaming Encoding
//
// For large messages or network transmission, use EncodeToWriter for
// efficient streaming with context cancellation support:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	err := enc.EncodeToWriter(ctx, conn, msg)
//	if err != nil {
//	    if errors.Is(err, context.DeadlineExceeded) {
//	        log.Println("encode timeout")
//	    } else {
//	        log.Println("encode error:", err)
//	    }
//	}
//
// # Error Handling
//
// Encoding errors are returned as *Error with detailed information:
//
//	data, err := enc.Encode(msg)
//	if err != nil {
//	    var encErr *encode.Error
//	    if errors.As(err, &encErr) {
//	        fmt.Printf("Encode failed: %s\n", encErr.Message)
//	        if encErr.Segment != "" {
//	            fmt.Printf("  at segment: %s\n", encErr.Segment)
//	        }
//	        if encErr.Cause != nil {
//	            fmt.Printf("  cause: %v\n", encErr.Cause)
//	        }
//	    }
//	}
//
// # Example: Complete Encoding Workflow
//
//	// Create encoder with MLLP for network transmission
//	enc := encode.New(
//	    encode.WithMLLP(true),
//	    encode.WithLineEnding("\r"),
//	)
//
//	// Parse incoming message, modify it, and send response
//	parser := parse.New()
//	msg, err := parser.Parse(incomingData)
//	if err != nil {
//	    return err
//	}
//
//	// Modify message
//	msg.Set("MSH.5", "RECEIVING_APP")
//	msg.Set("MSH.6", "RECEIVING_FACILITY")
//
//	// Encode and send
//	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
//	defer cancel()
//
//	if err := enc.EncodeToWriter(ctx, conn, msg); err != nil {
//	    return fmt.Errorf("failed to send message: %w", err)
//	}
//
// # MLLP Constants
//
// The package exports MLLP framing constants for use in custom implementations:
//
//	encode.MLLPStartBlock      // 0x0B - vertical tab
//	encode.MLLPEndBlock        // 0x1C - file separator
//	encode.MLLPCarriageReturn  // 0x0D - carriage return
package encode
