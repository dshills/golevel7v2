// Package ack provides HL7 v2.x acknowledgment (ACK) message generation.
//
// The ack package creates ACK messages in response to incoming HL7 messages.
// It supports both positive acknowledgments (AA - Application Accept) and
// negative acknowledgments (AE - Application Error, AR - Application Reject).
//
// # ACK Message Structure
//
// An ACK message consists of:
//   - MSH: Message header (mirrored from original with swapped sender/receiver)
//   - MSA: Message acknowledgment segment containing:
//   - MSA.1: Acknowledgment code (AA, AE, AR)
//   - MSA.2: Message control ID (from original MSH.10)
//   - MSA.3: Text message (optional description)
//   - ERR: Error segment (optional, for AE/AR responses)
//
// # Basic Usage
//
// Generate a positive acknowledgment:
//
//	// Parse incoming message
//	msg, err := parser.Parse(incomingData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Generate ACK
//	ackMsg, err := ack.Generate(msg, ack.Accept())
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Encode and send response
//	response, _ := encoder.Encode(ackMsg)
//	conn.Write(response)
//
// Generate a negative acknowledgment:
//
//	ackMsg, err := ack.Generate(msg, ack.Error("Patient ID not found"))
//
// Generate a rejection:
//
//	ackMsg, err := ack.Generate(msg, ack.Reject("Invalid message format"))
//
// # Acknowledgment Codes
//
// HL7 defines three acknowledgment codes:
//
//	AA - Application Accept
//	    Message was received, validated, and accepted for processing.
//
//	AE - Application Error
//	    Message was received but contained errors. The sending application
//	    should review and correct the message.
//
//	AR - Application Reject
//	    Message was rejected due to system or communication errors.
//	    The sending application may retry the transmission.
//
// # ACK Options
//
// Configure ACK generation with functional options:
//
//	// Accept with custom text
//	ackMsg, err := ack.Generate(msg,
//	    ack.Accept(),
//	    ack.WithText("Message processed successfully"),
//	)
//
//	// Error with error code
//	ackMsg, err := ack.Generate(msg,
//	    ack.Error("Validation failed"),
//	    ack.WithErrorCode("100"),
//	    ack.WithErrorLocation("PID.3"),
//	)
//
//	// Set custom sending application
//	ackMsg, err := ack.Generate(msg,
//	    ack.Accept(),
//	    ack.WithSendingApp("MY_APP"),
//	    ack.WithSendingFacility("MY_FACILITY"),
//	)
//
// # Error Segments
//
// For AE and AR responses, include error details:
//
//	ackMsg, err := ack.Generate(msg,
//	    ack.Error("Validation error"),
//	    ack.WithErrorSegment(ack.ErrorSegment{
//	        ErrorCode:        "101",
//	        ErrorCodeText:    "Required field missing",
//	        ErrorLocation:    "PID.3.1",
//	        DiagnosticInfo:   "Patient ID is required",
//	        UserMessage:      "Please provide a patient identifier",
//	    }),
//	)
//
// # Original Mode vs Enhanced Mode
//
// HL7 supports two acknowledgment modes:
//
// Original Mode (default):
//   - Single ACK for each message
//   - MSA segment only
//
// Enhanced Mode:
//   - Separate Accept and Application acknowledgments
//   - More detailed error reporting
//   - Additional ERR segment details
//
// Example enhanced mode:
//
//	ackMsg, err := ack.Generate(msg,
//	    ack.Accept(),
//	    ack.WithEnhancedMode(true),
//	    ack.WithAcceptAckType("CA"),  // Commit Accept
//	)
//
// # Message Control ID
//
// The ACK message control ID can be auto-generated or specified:
//
//	// Auto-generate (default)
//	ackMsg, err := ack.Generate(msg, ack.Accept())
//
//	// Use specific control ID
//	ackMsg, err := ack.Generate(msg,
//	    ack.Accept(),
//	    ack.WithControlID("ACK-12345"),
//	)
//
// # Example: Complete ACK Workflow
//
//	func handleHL7Message(data []byte) ([]byte, error) {
//	    // Parse incoming message
//	    parser := parse.New()
//	    msg, err := parser.Parse(data)
//	    if err != nil {
//	        // Generate reject for parse errors
//	        // Note: May need minimal message info for ACK
//	        return nil, fmt.Errorf("parse error: %w", err)
//	    }
//
//	    // Validate message
//	    validator := validate.NewValidator(
//	        validate.Required("PID.3.1"),
//	        validate.Required("PID.5"),
//	    )
//
//	    if errors := validator.Validate(msg); len(errors) > 0 {
//	        // Generate error ACK
//	        ackMsg, _ := ack.Generate(msg,
//	            ack.Error("Validation failed"),
//	            ack.WithText(errors[0].Error()),
//	            ack.WithErrorLocation(errors[0].Location),
//	        )
//	        return encode.New().Encode(ackMsg)
//	    }
//
//	    // Process message...
//	    if err := processMessage(msg); err != nil {
//	        // Generate error ACK for processing errors
//	        ackMsg, _ := ack.Generate(msg,
//	            ack.Error("Processing failed"),
//	            ack.WithText(err.Error()),
//	        )
//	        return encode.New().Encode(ackMsg)
//	    }
//
//	    // Generate success ACK
//	    ackMsg, _ := ack.Generate(msg,
//	        ack.Accept(),
//	        ack.WithText("Message processed successfully"),
//	    )
//	    return encode.New().Encode(ackMsg)
//	}
//
// # Example ACK Message
//
// For an incoming ADT^A01 message, a successful ACK looks like:
//
//	MSH|^~\&|RECEIVING_APP|RECEIVING_FAC|SENDING_APP|SENDING_FAC|20240115120000||ACK^A01|ACK12345|P|2.5.1
//	MSA|AA|MSG12345|Message accepted
//
// An error ACK:
//
//	MSH|^~\&|RECEIVING_APP|RECEIVING_FAC|SENDING_APP|SENDING_FAC|20240115120000||ACK^A01|ACK12346|P|2.5.1
//	MSA|AE|MSG12345|Patient ID not found
//	ERR|||100|E||||Patient identifier is required in PID-3
package ack
