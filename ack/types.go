// Package ack provides functionality for building HL7 v2.x acknowledgment (ACK) messages.
//
// ACK messages are used to confirm receipt and acceptance of HL7 messages.
// The package supports all standard acknowledgment codes:
//   - AA (Application Accept): Message accepted successfully
//   - AE (Application Error): Message contains errors but was received
//   - AR (Application Reject): Message rejected, not processed
//   - CA (Commit Accept): Message committed to storage
//   - CE (Commit Error): Commit failed with errors
//   - CR (Commit Reject): Commit rejected
//
// An ACK message consists of:
//   - MSH segment: Message header with swapped sending/receiving applications
//   - MSA segment: Message acknowledgment with code and original message control ID
//   - ERR segment (optional): Error details when acknowledgment indicates an error
package ack

// Code represents an HL7 acknowledgment code.
// Acknowledgment codes indicate the result of processing the original message.
type Code string

// Standard HL7 acknowledgment codes.
const (
	// ApplicationAccept indicates the message was received, understood,
	// and accepted for processing.
	ApplicationAccept Code = "AA"

	// ApplicationError indicates the message was received and understood,
	// but contains errors that prevent processing.
	ApplicationError Code = "AE"

	// ApplicationReject indicates the message was received but rejected.
	// The message will not be processed.
	ApplicationReject Code = "AR"

	// CommitAccept indicates the message was received and committed
	// to safe storage for later processing.
	CommitAccept Code = "CA"

	// CommitError indicates the message was received but could not be
	// committed to safe storage due to errors.
	CommitError Code = "CE"

	// CommitReject indicates the message was received but rejected
	// from being committed to safe storage.
	CommitReject Code = "CR"
)

// String returns the string representation of the acknowledgment code.
func (c Code) String() string {
	return string(c)
}

// IsAccept returns true if the code represents an accept condition (AA or CA).
func (c Code) IsAccept() bool {
	return c == ApplicationAccept || c == CommitAccept
}

// IsError returns true if the code represents an error condition (AE or CE).
func (c Code) IsError() bool {
	return c == ApplicationError || c == CommitError
}

// IsReject returns true if the code represents a reject condition (AR or CR).
func (c Code) IsReject() bool {
	return c == ApplicationReject || c == CommitReject
}

// IsValid returns true if the code is a valid HL7 acknowledgment code.
func (c Code) IsValid() bool {
	switch c {
	case ApplicationAccept, ApplicationError, ApplicationReject,
		CommitAccept, CommitError, CommitReject:
		return true
	default:
		return false
	}
}

// ACK represents acknowledgment data used to construct an ACK message.
// It contains all the information needed to build MSA and optional ERR segments.
type ACK struct {
	// Code is the acknowledgment code (AA, AE, AR, CA, CE, CR).
	// This is placed in MSA-1.
	Code Code

	// ControlID is the message control ID from the original message (MSH-10).
	// This is placed in MSA-2 to correlate the ACK with the original message.
	ControlID string

	// TextMessage is an optional text message describing the acknowledgment.
	// This is placed in MSA-3.
	TextMessage string

	// ErrorCode is the error code for the ERR segment.
	// If non-empty, an ERR segment will be included in the ACK message.
	// Common values include:
	//   - "0"   : Message accepted
	//   - "100" : Segment sequence error
	//   - "101" : Required field missing
	//   - "102" : Data type error
	//   - "103" : Table value not found
	//   - "200" : Unsupported message type
	//   - "201" : Unsupported event code
	//   - "202" : Unsupported processing id
	//   - "203" : Unsupported version id
	//   - "204" : Unknown key identifier
	//   - "205" : Duplicate key identifier
	//   - "206" : Application record locked
	//   - "207" : Application internal error
	ErrorCode string

	// ErrorLocation is the HL7 location path where the error occurred.
	// Format: "SEG-Field-Component-SubComponent" (e.g., "PID-3-1").
	// This is placed in ERR-2 (Error Location) in HL7 v2.4+ or ERR-1 in earlier versions.
	ErrorLocation string

	// ErrorMessage provides additional details about the error.
	// This is placed in ERR-7 (Diagnostic Information) in HL7 v2.5+.
	ErrorMessage string

	// Severity indicates the error severity for the ERR segment.
	// Values: "E" (Error), "W" (Warning), "I" (Information).
	// This is placed in ERR-4 in HL7 v2.5+.
	Severity string
}

// NewAcceptACK creates an ACK struct for accepting a message.
func NewAcceptACK(controlID string) ACK {
	return ACK{
		Code:      ApplicationAccept,
		ControlID: controlID,
	}
}

// NewErrorACK creates an ACK struct for an error response.
func NewErrorACK(controlID string, errorCode string, message string) ACK {
	return ACK{
		Code:         ApplicationError,
		ControlID:    controlID,
		ErrorCode:    errorCode,
		TextMessage:  message,
		ErrorMessage: message,
		Severity:     "E",
	}
}

// NewRejectACK creates an ACK struct for rejecting a message.
func NewRejectACK(controlID string, reason string) ACK {
	return ACK{
		Code:        ApplicationReject,
		ControlID:   controlID,
		TextMessage: reason,
		Severity:    "E",
	}
}

// HasError returns true if the ACK includes error information.
func (a ACK) HasError() bool {
	return a.ErrorCode != "" || a.ErrorLocation != "" || a.ErrorMessage != ""
}

// NeedsERRSegment returns true if the ACK should include an ERR segment.
// An ERR segment is included when there is error information and the
// acknowledgment code indicates an error or reject condition.
func (a ACK) NeedsERRSegment() bool {
	return a.HasError() && (a.Code.IsError() || a.Code.IsReject())
}
