// Package parse provides HL7 v2.x message parsing functionality.
//
// The parse package converts raw HL7 message bytes or strings into structured
// [hl7.Message] objects. It handles delimiter detection, segment splitting,
// field parsing, escape sequence processing, and validation.
//
// # Basic Usage
//
// Parse a message from bytes:
//
//	p := parse.New()
//	msg, err := p.Parse(data)
//	if err != nil {
//	    log.Fatal("parse error:", err)
//	}
//
//	// Access message data
//	msgType := msg.Type()         // e.g., "ADT^A01"
//	controlID := msg.ControlID()  // e.g., "12345"
//	version := msg.Version()      // e.g., "2.5.1"
//
// Parse a message from a string:
//
//	msg, err := p.ParseString(hl7String)
//	if err != nil {
//	    log.Fatal("parse error:", err)
//	}
//
// # Parser Options
//
// The parser supports functional options for configuration:
//
//	// Enable strict parsing mode
//	p := parse.New(parse.WithStrictMode(true))
//
//	// Allow empty segments
//	p := parse.New(parse.WithAllowEmptySegments(true))
//
//	// Use custom delimiters (for non-standard messages)
//	delims := &hl7.Delimiters{
//	    Field:        '|',
//	    Component:    '^',
//	    Repetition:   '~',
//	    Escape:       '\\',
//	    SubComponent: '&',
//	}
//	p := parse.New(parse.WithCustomDelimiters(delims))
//
//	// Set DoS protection limits
//	p := parse.New(
//	    parse.WithMaxSegments(500),
//	    parse.WithMaxFieldLength(32768),
//	)
//
//	// Use a different segment terminator (default is CR)
//	p := parse.New(parse.WithSegmentTerminator('\n'))
//
// # Delimiter Detection
//
// By default, the parser automatically detects delimiters from the MSH segment:
//   - MSH-1 (position 3) contains the field separator
//   - MSH-2 (positions 4-7) contains the encoding characters
//
// For standard HL7 messages, delimiters are typically:
//
//	MSH|^~\&|...
//
// Where | is the field separator, ^ is component, ~ is repetition,
// \ is escape, and & is subcomponent.
//
// # Strict Mode
//
// When strict mode is enabled, the parser performs additional validation:
//   - Segment names must be exactly 3 uppercase alphanumeric characters
//   - MSH segment must be first
//   - Required fields in MSH must be present
//   - Field lengths are validated against maximums
//
// In non-strict mode (default), the parser is more lenient and will
// accept messages with minor formatting issues.
//
// # DoS Protection
//
// The parser includes built-in protection against denial-of-service attacks:
//   - Maximum segment count (default: 1000)
//   - Maximum field length (default: 65536 bytes)
//
// These limits prevent maliciously crafted messages from consuming
// excessive memory or CPU time.
//
// # Error Handling
//
// Parse errors include detailed information about what went wrong:
//
//	msg, err := p.Parse(data)
//	if err != nil {
//	    // Error contains location and description
//	    fmt.Println("Parse failed:", err)
//	}
//
// Common error conditions:
//   - Missing or invalid MSH segment
//   - Invalid delimiters
//   - Segment count exceeds maximum
//   - Field length exceeds maximum
//   - Invalid escape sequences (in strict mode)
//
// # Example: Complete Parsing Workflow
//
//	// Create parser with options
//	p := parse.New(
//	    parse.WithStrictMode(true),
//	    parse.WithMaxSegments(500),
//	)
//
//	// Parse message
//	msg, err := p.Parse(rawHL7Data)
//	if err != nil {
//	    return fmt.Errorf("failed to parse HL7 message: %w", err)
//	}
//
//	// Extract message metadata
//	fmt.Printf("Message Type: %s\n", msg.Type())
//	fmt.Printf("Control ID: %s\n", msg.ControlID())
//	fmt.Printf("Version: %s\n", msg.Version())
//
//	// Access patient data
//	patientID, _ := msg.Get("PID.3.1")
//	patientName, _ := msg.Get("PID.5")
//	dob, _ := msg.Get("PID.7")
//
//	fmt.Printf("Patient: %s (ID: %s, DOB: %s)\n",
//	    patientName, patientID, dob)
//
//	// Iterate over observations
//	for _, obx := range msg.Segments("OBX") {
//	    obsID, _ := obx.Get("3")
//	    value, _ := obx.Get("5")
//	    units, _ := obx.Get("6")
//	    fmt.Printf("  %s: %s %s\n", obsID, value, units)
//	}
package parse
