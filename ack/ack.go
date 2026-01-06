package ack

import (
	"errors"
	"fmt"
	"time"

	"github.com/dshills/golevel7/hl7"
)

// Errors returned by the ACK builder.
var (
	// ErrNilMessage indicates a nil message was provided.
	ErrNilMessage = errors.New("nil message")

	// ErrMissingControlID indicates the original message has no control ID.
	ErrMissingControlID = errors.New("original message missing control ID (MSH-10)")

	// ErrMissingMSH indicates the original message has no MSH segment.
	ErrMissingMSH = errors.New("original message missing MSH segment")

	// ErrInvalidACKCode indicates an invalid acknowledgment code was provided.
	ErrInvalidACKCode = errors.New("invalid acknowledgment code")
)

// Builder creates HL7 acknowledgment messages from original messages.
// It handles the construction of MSH, MSA, and optional ERR segments.
type Builder interface {
	// Accept creates an acceptance ACK (AA) for the original message.
	// The ACK message will have:
	//   - MSH segment with swapped sending/receiving applications
	//   - MSA segment with code "AA" and original message control ID
	Accept(original hl7.Message) (hl7.Message, error)

	// Reject creates a rejection ACK (AR) for the original message.
	// The ACK message will have:
	//   - MSH segment with swapped sending/receiving applications
	//   - MSA segment with code "AR" and original message control ID
	//   - Optional reason text in MSA-3
	Reject(original hl7.Message, reason string) (hl7.Message, error)

	// Error creates an error ACK (AE) for the original message.
	// The ACK message will have:
	//   - MSH segment with swapped sending/receiving applications
	//   - MSA segment with code "AE" and original message control ID
	//   - Error message from err.Error() in MSA-3
	//   - ERR segment with error details
	Error(original hl7.Message, err error) (hl7.Message, error)

	// Custom creates an ACK with fully customized acknowledgment data.
	// Use this for advanced scenarios requiring specific error codes,
	// error locations, or non-standard acknowledgment handling.
	Custom(original hl7.Message, ack ACK) (hl7.Message, error)
}

// builder is the concrete implementation of Builder.
type builder struct {
	// messageFactory creates new messages and segments.
	// If nil, a default implementation is used.
	messageFactory MessageFactory

	// timeFunc returns the current time. Used for testing.
	timeFunc func() time.Time

	// controlIDFunc generates unique control IDs for ACK messages.
	// If nil, uses timestamp-based generation.
	controlIDFunc func() string
}

// MessageFactory creates HL7 messages and segments.
// This interface allows for dependency injection and testing.
type MessageFactory interface {
	// NewMessage creates a new empty message with the given delimiters.
	NewMessage(delims *hl7.Delimiters) hl7.Message

	// NewSegment creates a new segment with the given name.
	NewSegment(name string, delims *hl7.Delimiters) hl7.Segment
}

// Option configures a Builder.
type Option func(*builder)

// WithMessageFactory sets a custom message factory.
func WithMessageFactory(factory MessageFactory) Option {
	return func(b *builder) {
		b.messageFactory = factory
	}
}

// WithTimeFunc sets a custom time function for testing.
func WithTimeFunc(fn func() time.Time) Option {
	return func(b *builder) {
		b.timeFunc = fn
	}
}

// WithControlIDFunc sets a custom control ID generator.
func WithControlIDFunc(fn func() string) Option {
	return func(b *builder) {
		b.controlIDFunc = fn
	}
}

// NewBuilder creates a new ACK Builder with the given options.
func NewBuilder(opts ...Option) Builder {
	b := &builder{
		timeFunc: time.Now,
	}

	for _, opt := range opts {
		opt(b)
	}

	if b.controlIDFunc == nil {
		b.controlIDFunc = func() string {
			return fmt.Sprintf("ACK%d", b.timeFunc().UnixNano())
		}
	}

	return b
}

// Accept creates an acceptance ACK (AA) for the original message.
func (b *builder) Accept(original hl7.Message) (hl7.Message, error) {
	if original == nil {
		return nil, ErrNilMessage
	}

	controlID := original.ControlID()
	if controlID == "" {
		return nil, ErrMissingControlID
	}

	return b.Custom(original, NewAcceptACK(controlID))
}

// Reject creates a rejection ACK (AR) for the original message.
func (b *builder) Reject(original hl7.Message, reason string) (hl7.Message, error) {
	if original == nil {
		return nil, ErrNilMessage
	}

	controlID := original.ControlID()
	if controlID == "" {
		return nil, ErrMissingControlID
	}

	return b.Custom(original, NewRejectACK(controlID, reason))
}

// Error creates an error ACK (AE) for the original message.
func (b *builder) Error(original hl7.Message, err error) (hl7.Message, error) {
	if original == nil {
		return nil, ErrNilMessage
	}

	controlID := original.ControlID()
	if controlID == "" {
		return nil, ErrMissingControlID
	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	ack := NewErrorACK(controlID, "207", errMsg) // 207 = Application internal error
	return b.Custom(original, ack)
}

// Custom creates an ACK with fully customized acknowledgment data.
func (b *builder) Custom(original hl7.Message, ack ACK) (hl7.Message, error) {
	if original == nil {
		return nil, ErrNilMessage
	}

	if !ack.Code.IsValid() {
		return nil, fmt.Errorf("%w: %s", ErrInvalidACKCode, ack.Code)
	}

	// Get the original MSH segment
	msh, ok := original.Segment("MSH")
	if !ok {
		return nil, ErrMissingMSH
	}

	// Use original message's delimiters
	delims := original.Delimiters()
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	// Build the ACK message
	return b.buildACKMessage(msh, delims, ack)
}

// buildACKMessage constructs the complete ACK message.
func (b *builder) buildACKMessage(originalMSH hl7.Segment, delims *hl7.Delimiters, ack ACK) (hl7.Message, error) {
	// Create a new message using the factory if available
	var msg hl7.Message
	if b.messageFactory != nil {
		msg = b.messageFactory.NewMessage(delims)
	} else {
		// Use default message implementation
		msg = newSimpleMessage(delims)
	}

	// Build and add MSH segment
	mshSeg, err := b.buildMSHSegment(originalMSH, delims, ack)
	if err != nil {
		return nil, fmt.Errorf("building MSH segment: %w", err)
	}
	if err := msg.AddSegment(mshSeg); err != nil {
		return nil, fmt.Errorf("adding MSH segment: %w", err)
	}

	// Build and add MSA segment
	msaSeg, err := b.buildMSASegment(delims, ack)
	if err != nil {
		return nil, fmt.Errorf("building MSA segment: %w", err)
	}
	if err := msg.AddSegment(msaSeg); err != nil {
		return nil, fmt.Errorf("adding MSA segment: %w", err)
	}

	// Build and add ERR segment if needed
	if ack.NeedsERRSegment() {
		errSeg, err := b.buildERRSegment(delims, ack)
		if err != nil {
			return nil, fmt.Errorf("building ERR segment: %w", err)
		}
		if err := msg.AddSegment(errSeg); err != nil {
			return nil, fmt.Errorf("adding ERR segment: %w", err)
		}
	}

	return msg, nil
}

// buildMSHSegment creates the MSH segment for the ACK message.
// It swaps sending and receiving applications from the original MSH.
func (b *builder) buildMSHSegment(originalMSH hl7.Segment, delims *hl7.Delimiters, _ ACK) (hl7.Segment, error) {
	var seg hl7.Segment
	if b.messageFactory != nil {
		seg = b.messageFactory.NewSegment("MSH", delims)
	} else {
		seg = newSimpleSegment("MSH", delims)
	}

	// MSH-1: Field separator (set implicitly by segment)
	// MSH-2: Encoding characters (set implicitly by segment)

	// Swap sending and receiving applications
	// Original MSH-3 (Sending App) -> ACK MSH-5 (Receiving App)
	// Original MSH-4 (Sending Facility) -> ACK MSH-6 (Receiving Facility)
	// Original MSH-5 (Receiving App) -> ACK MSH-3 (Sending App)
	// Original MSH-6 (Receiving Facility) -> ACK MSH-4 (Sending Facility)

	originalSendingApp, _ := originalMSH.Get("3")
	originalSendingFacility, _ := originalMSH.Get("4")
	originalReceivingApp, _ := originalMSH.Get("5")
	originalReceivingFacility, _ := originalMSH.Get("6")

	// MSH-3: Sending Application (was receiving)
	if err := seg.Set("3", originalReceivingApp); err != nil {
		return nil, fmt.Errorf("setting MSH-3: %w", err)
	}

	// MSH-4: Sending Facility (was receiving)
	if err := seg.Set("4", originalReceivingFacility); err != nil {
		return nil, fmt.Errorf("setting MSH-4: %w", err)
	}

	// MSH-5: Receiving Application (was sending)
	if err := seg.Set("5", originalSendingApp); err != nil {
		return nil, fmt.Errorf("setting MSH-5: %w", err)
	}

	// MSH-6: Receiving Facility (was sending)
	if err := seg.Set("6", originalSendingFacility); err != nil {
		return nil, fmt.Errorf("setting MSH-6: %w", err)
	}

	// MSH-7: Date/Time of Message
	timestamp := b.timeFunc().Format("20060102150405")
	if err := seg.Set("7", timestamp); err != nil {
		return nil, fmt.Errorf("setting MSH-7: %w", err)
	}

	// MSH-9: Message Type (ACK)
	// Format: ACK^<trigger event from original>
	originalMsgType, _ := originalMSH.Get("9")
	ackMsgType := "ACK"
	if originalMsgType != "" {
		// Extract trigger event (second component)
		// Parse using delimiters to get component
		if field, ok := originalMSH.Field(9); ok {
			if rep, ok := field.Repetition(0); ok {
				if comp, ok := rep.Component(2); ok {
					triggerEvent := comp.Value()
					if triggerEvent != "" {
						ackMsgType = fmt.Sprintf("ACK%c%s", delims.Component, triggerEvent)
					}
				}
			}
		}
	}
	if err := seg.Set("9", ackMsgType); err != nil {
		return nil, fmt.Errorf("setting MSH-9: %w", err)
	}

	// MSH-10: Message Control ID (unique for the ACK)
	controlID := b.controlIDFunc()
	if err := seg.Set("10", controlID); err != nil {
		return nil, fmt.Errorf("setting MSH-10: %w", err)
	}

	// MSH-11: Processing ID (copy from original)
	processingID, _ := originalMSH.Get("11")
	if processingID != "" {
		if err := seg.Set("11", processingID); err != nil {
			return nil, fmt.Errorf("setting MSH-11: %w", err)
		}
	}

	// MSH-12: Version ID (copy from original)
	versionID, _ := originalMSH.Get("12")
	if versionID != "" {
		if err := seg.Set("12", versionID); err != nil {
			return nil, fmt.Errorf("setting MSH-12: %w", err)
		}
	}

	return seg, nil
}

// buildMSASegment creates the MSA (Message Acknowledgment) segment.
func (b *builder) buildMSASegment(delims *hl7.Delimiters, ack ACK) (hl7.Segment, error) {
	var seg hl7.Segment
	if b.messageFactory != nil {
		seg = b.messageFactory.NewSegment("MSA", delims)
	} else {
		seg = newSimpleSegment("MSA", delims)
	}

	// MSA-1: Acknowledgment Code
	if err := seg.Set("1", string(ack.Code)); err != nil {
		return nil, fmt.Errorf("setting MSA-1: %w", err)
	}

	// MSA-2: Message Control ID (from original message)
	if err := seg.Set("2", ack.ControlID); err != nil {
		return nil, fmt.Errorf("setting MSA-2: %w", err)
	}

	// MSA-3: Text Message (optional)
	if ack.TextMessage != "" {
		if err := seg.Set("3", ack.TextMessage); err != nil {
			return nil, fmt.Errorf("setting MSA-3: %w", err)
		}
	}

	return seg, nil
}

// buildERRSegment creates the ERR (Error) segment for error/reject ACKs.
func (b *builder) buildERRSegment(delims *hl7.Delimiters, ack ACK) (hl7.Segment, error) {
	var seg hl7.Segment
	if b.messageFactory != nil {
		seg = b.messageFactory.NewSegment("ERR", delims)
	} else {
		seg = newSimpleSegment("ERR", delims)
	}

	// ERR-1: Error Code and Location (HL7 v2.3 and earlier)
	// For backward compatibility, we set this if ErrorLocation is provided
	if ack.ErrorLocation != "" {
		if err := seg.Set("1", ack.ErrorLocation); err != nil {
			return nil, fmt.Errorf("setting ERR-1: %w", err)
		}
	}

	// ERR-2: Error Location (HL7 v2.4+)
	// This is a more structured location in newer versions
	if ack.ErrorLocation != "" {
		if err := seg.Set("2", ack.ErrorLocation); err != nil {
			return nil, fmt.Errorf("setting ERR-2: %w", err)
		}
	}

	// ERR-3: HL7 Error Code (HL7 v2.5+)
	if ack.ErrorCode != "" {
		if err := seg.Set("3", ack.ErrorCode); err != nil {
			return nil, fmt.Errorf("setting ERR-3: %w", err)
		}
	}

	// ERR-4: Severity (HL7 v2.5+)
	if ack.Severity != "" {
		if err := seg.Set("4", ack.Severity); err != nil {
			return nil, fmt.Errorf("setting ERR-4: %w", err)
		}
	}

	// ERR-7: Diagnostic Information (HL7 v2.5+)
	if ack.ErrorMessage != "" {
		if err := seg.Set("7", ack.ErrorMessage); err != nil {
			return nil, fmt.Errorf("setting ERR-7: %w", err)
		}
	}

	return seg, nil
}
