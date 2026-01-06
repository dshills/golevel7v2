// Package hl7 provides core types and interfaces for HL7 v2.x message handling.
//
// The hl7 package defines the fundamental data structures for representing
// HL7 messages: Message, Segment, Field, Component, and SubComponent.
// All types are defined as interfaces to enable testing and extensibility.
//
// # Message Structure
//
// HL7 messages follow a hierarchical structure:
//   - Message contains Segments
//   - Segment contains Fields
//   - Field contains Repetitions (separated by ~)
//   - Repetition contains Components (separated by ^)
//   - Component contains SubComponents (separated by &)
//
// # Location Syntax
//
// The package uses a location string syntax to address values within messages.
// The format is: SEG[idx].field[rep].component.subcomponent
//
// Examples:
//   - "PID" - entire PID segment
//   - "PID.5" - field 5 of PID
//   - "PID.5.1" - component 1 of field 5
//   - "PID.5.1.2" - subcomponent 2 of component 1
//   - "PID[1].5" - field 5 of the second PID segment
//   - "PID.5[0].1" - component 1 of the first repetition of field 5
//
// Field, Component, and SubComponent indices are 1-based per HL7 convention.
// Segment and Repetition indices are 0-based.
//
// # Delimiters
//
// HL7 v2.x messages define their delimiters in the MSH segment:
//   - MSH-1: Field separator (typically |)
//   - MSH-2: Encoding characters (typically ^~\&)
//
// The default delimiters are:
//   - Field: |
//   - Component: ^
//   - Repetition: ~
//   - Escape: \
//   - SubComponent: &
//
// # Escape Sequences
//
// Special characters within data values are represented using escape sequences:
//   - \F\ for field separator (|)
//   - \S\ for component separator (^)
//   - \T\ for subcomponent separator (&)
//   - \R\ for repetition separator (~)
//   - \E\ for escape character (\)
//   - \Xhh...\ for hexadecimal data
//   - \.br\ for line breaks
//
// # Example Usage
//
// Accessing values in a parsed message:
//
//	msg, err := parser.Parse(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get patient name
//	name, err := msg.Get("PID.5")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get family name (component 1)
//	lastName, err := msg.Get("PID.5.1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get all patient identifiers (repeating field)
//	ids, err := msg.GetAll("PID.3")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Set a value
//	err = msg.Set("PID.5.1", "SMITH")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Working with segments:
//
//	// Get a specific segment
//	pid, ok := msg.Segment("PID")
//	if !ok {
//	    log.Fatal("PID segment not found")
//	}
//
//	// Get all OBX segments
//	observations := msg.Segments("OBX")
//	for _, obx := range observations {
//	    value, _ := obx.Get("5")
//	    fmt.Println("Observation:", value)
//	}
//
// Using Location for efficient repeated access:
//
//	// Parse location once, use many times
//	loc, err := hl7.ParseLocation("PID.5.1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, msg := range messages {
//	    name, _ := msg.GetAt(loc)
//	    fmt.Println("Patient:", name)
//	}
//
// # Interface Design
//
// All message structure types (Message, Segment, Field, etc.) are defined as
// interfaces rather than concrete types. This design enables:
//   - Easy mocking for unit tests
//   - Alternative implementations optimized for specific use cases
//   - Lazy parsing implementations for performance
//   - Custom storage backends
//
// The package also defines Parser, Encoder, and Validator interfaces that
// are implemented by the parse, encode, and validate packages respectively.
package hl7
