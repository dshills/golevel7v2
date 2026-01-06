// Package hl7 provides interfaces and types for parsing and encoding HL7 v2.x messages.
//
// The package follows a hierarchical structure that mirrors the HL7 v2.x message format:
// Message -> Segment -> Field -> Repetition -> Component -> SubComponent
//
// Each level in the hierarchy implements common operations for accessing and
// manipulating data, while respecting HL7 encoding rules and delimiters.
//
// Core interfaces (Message, Segment, Field, Repetition, Component, SubComponent)
// are defined in their respective implementation files. This file contains
// additional interfaces for parsing, encoding, validation, and message building.
package hl7

// Escaper handles HL7 escape sequence encoding and decoding.
//
// HL7 v2.x uses escape sequences to represent special characters within
// data values. The escape character (default \) introduces sequences like:
//   - \F\ for field separator (|)
//   - \S\ for component separator (^)
//   - \T\ for subcomponent separator (&)
//   - \R\ for repetition separator (~)
//   - \E\ for escape character (\)
//   - \Xhh...\ for hexadecimal data
//   - \.br\ for line breaks
//   - \H\ for start highlighting
//   - \N\ for normal text (end highlighting)
type Escaper interface {
	// Escape encodes special characters in the input string using HL7 escape sequences.
	// Characters that match the configured delimiters are replaced with their
	// corresponding escape sequences.
	Escape(value string) string

	// Unescape decodes HL7 escape sequences in the input string.
	// Escape sequences are replaced with their literal character equivalents.
	Unescape(value string) string

	// Delimiters returns the delimiter configuration used by this escaper.
	Delimiters() *Delimiters
}

// Parser defines the interface for parsing HL7 v2.x messages.
//
// Implementations handle the conversion of raw HL7 message bytes into
// the structured Message hierarchy.
type Parser interface {
	// Parse parses raw HL7 message data into a Message.
	// Returns an error if the data is not a valid HL7 message.
	Parse(data []byte) (Message, error)

	// ParseString parses an HL7 message string into a Message.
	// Convenience wrapper around Parse.
	ParseString(data string) (Message, error)
}

// Encoder defines the interface for encoding HL7 v2.x messages.
//
// Implementations handle the conversion of the structured Message
// hierarchy back into raw HL7 message format.
type Encoder interface {
	// Encode encodes a Message into HL7 format bytes.
	Encode(msg Message) ([]byte, error)

	// EncodeString encodes a Message into an HL7 format string.
	EncodeString(msg Message) (string, error)
}

// Validator defines the interface for validating HL7 messages.
//
// Implementations can check message structure, required fields,
// data types, and conformance to specific message profiles.
type Validator interface {
	// Validate checks the message for errors.
	// Returns a slice of validation errors, or nil if the message is valid.
	Validate(msg Message) []error

	// ValidateSegment validates a single segment.
	ValidateSegment(seg Segment) []error
}

// MessageBuilder provides a fluent interface for constructing HL7 messages.
//
// MessageBuilder allows programmatic creation of HL7 messages without
// manually managing the internal structure. It handles proper delimiter
// configuration and segment ordering.
type MessageBuilder interface {
	// SetDelimiters configures custom delimiters for the message.
	// If not called, default delimiters are used.
	SetDelimiters(delims *Delimiters) MessageBuilder

	// SetVersion sets the HL7 version in MSH-12.
	SetVersion(version string) MessageBuilder

	// SetType sets the message type in MSH-9.
	// Example: SetType("ADT", "A01")
	SetType(messageType, triggerEvent string) MessageBuilder

	// SetControlID sets the message control ID in MSH-10.
	SetControlID(controlID string) MessageBuilder

	// SetSendingApplication sets MSH-3.
	SetSendingApplication(app string) MessageBuilder

	// SetSendingFacility sets MSH-4.
	SetSendingFacility(facility string) MessageBuilder

	// SetReceivingApplication sets MSH-5.
	SetReceivingApplication(app string) MessageBuilder

	// SetReceivingFacility sets MSH-6.
	SetReceivingFacility(facility string) MessageBuilder

	// SetDateTime sets the message date/time in MSH-7.
	// The time should be formatted according to HL7 DTM format.
	SetDateTime(datetime string) MessageBuilder

	// AddSegment adds a segment to the message.
	AddSegment(seg Segment) MessageBuilder

	// Set sets a value at the specified location.
	Set(location string, value string) MessageBuilder

	// Build constructs and returns the Message.
	// Returns an error if the message configuration is invalid.
	Build() (Message, error)
}

// SegmentBuilder provides a fluent interface for constructing HL7 segments.
//
// SegmentBuilder allows programmatic creation of segments without
// manually managing fields and components.
type SegmentBuilder interface {
	// SetName sets the segment name (e.g., "PID", "OBX").
	SetName(name string) SegmentBuilder

	// SetDelimiters configures custom delimiters for the segment.
	SetDelimiters(delims *Delimiters) SegmentBuilder

	// SetField sets the value at the specified 1-based field index.
	SetField(index int, value string) SegmentBuilder

	// SetComponent sets a component value at the specified field and component indices.
	// Field index is 1-based, component index is 1-based.
	SetComponent(fieldIndex, componentIndex int, value string) SegmentBuilder

	// SetSubComponent sets a subcomponent value at the specified indices.
	// All indices are 1-based.
	SetSubComponent(fieldIndex, componentIndex, subComponentIndex int, value string) SegmentBuilder

	// AddRepetition adds a repetition to the specified field.
	AddRepetition(fieldIndex int, value string) SegmentBuilder

	// Build constructs and returns the Segment.
	// Returns an error if the segment configuration is invalid.
	Build() (Segment, error)
}

// Walker provides methods for traversing HL7 message structures.
//
// Walker enables visitor-pattern style traversal of messages, useful for
// transformations, validations, or data extraction across entire messages.
type Walker interface {
	// Walk traverses the message and calls the visitor function for each element.
	// The visitor receives the current location and value.
	// If the visitor returns an error, traversal stops and the error is returned.
	Walk(msg Message, visitor func(loc *Location, value string) error) error

	// WalkSegment traverses a single segment.
	WalkSegment(seg Segment, visitor func(loc *Location, value string) error) error
}

// AckGenerator generates HL7 acknowledgment messages.
//
// AckGenerator creates ACK messages in response to received messages,
// following HL7 acknowledgment rules and conventions.
type AckGenerator interface {
	// GenerateACK creates an acknowledgment message for the given message.
	// The ackCode should be one of: AA (accept), AE (error), AR (reject).
	GenerateACK(msg Message, ackCode string, textMessage string) (Message, error)

	// GenerateNACK creates a negative acknowledgment for an unparseable message.
	// Used when the original message cannot be parsed.
	GenerateNACK(controlID string, ackCode string, textMessage string) (Message, error)
}

// SegmentIterator provides iteration over segments in a message.
//
// SegmentIterator allows efficient traversal of segments without
// allocating a slice for all segments.
type SegmentIterator interface {
	// Next advances the iterator and returns true if there is another segment.
	Next() bool

	// Segment returns the current segment.
	// Must be called after Next returns true.
	Segment() Segment

	// Err returns any error encountered during iteration.
	Err() error
}

// StreamParser provides streaming parsing of HL7 messages.
//
// StreamParser allows parsing of HL7 messages from an io.Reader without
// loading the entire message into memory first.
type StreamParser interface {
	// ParseNext parses the next message from the stream.
	// Returns io.EOF when no more messages are available.
	ParseNext() (Message, error)

	// Reset resets the parser to read from a new reader.
	Reset(r interface{})
}

// MessageMatcher defines criteria for matching HL7 messages.
//
// MessageMatcher is used for routing, filtering, or selecting messages
// based on their content.
type MessageMatcher interface {
	// Match returns true if the message matches the criteria.
	Match(msg Message) bool

	// MatchSegment returns true if the segment matches the criteria.
	MatchSegment(seg Segment) bool
}

// Transformer transforms HL7 messages according to defined rules.
//
// Transformer can modify, filter, or convert messages during processing.
type Transformer interface {
	// Transform applies transformations to the message.
	// Returns a new message with the transformations applied.
	// The original message is not modified.
	Transform(msg Message) (Message, error)

	// TransformSegment applies transformations to a single segment.
	TransformSegment(seg Segment) (Segment, error)
}

// MessageHandler processes HL7 messages.
//
// MessageHandler is used for implementing message processing pipelines.
type MessageHandler interface {
	// Handle processes the message and returns a response message.
	// The response may be an ACK, a result message, or nil if no response is needed.
	Handle(msg Message) (Message, error)
}

// FieldEncoder provides custom encoding for specific field types.
//
// FieldEncoder is used when field values require special formatting
// beyond the default string representation.
type FieldEncoder interface {
	// EncodeField encodes a value for the specified field location.
	// Returns the encoded string representation.
	EncodeField(loc *Location, value interface{}) (string, error)

	// DecodeField decodes a field value into the specified type.
	// Returns the decoded value.
	DecodeField(loc *Location, encoded string, target interface{}) error
}

// DateTimeEncoder handles HL7 date/time formatting and parsing.
//
// DateTimeEncoder provides conversion between Go time.Time values and
// HL7 DTM (date/time) format strings.
type DateTimeEncoder interface {
	// Format formats a time value as an HL7 DTM string.
	// The precision parameter controls the output format:
	// - "Y" for year only (YYYY)
	// - "M" for year-month (YYYYMM)
	// - "D" for date (YYYYMMDD)
	// - "H" for date-hour (YYYYMMDDHH)
	// - "m" for date-hour-minute (YYYYMMDDHHmm)
	// - "S" for full timestamp (YYYYMMDDHHmmss)
	// - "s" for timestamp with fractional seconds (YYYYMMDDHHmmss.SSSS)
	Format(t interface{}, precision string) (string, error)

	// Parse parses an HL7 DTM string into a time value.
	Parse(dtm string) (interface{}, error)
}
