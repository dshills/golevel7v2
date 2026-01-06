package segments

import (
	"errors"
	"fmt"

	"github.com/dshills/golevel7/hl7"
)

// MSH represents the Message Header segment.
// This segment is required in all HL7 v2.x messages and contains metadata
// about the message including sending/receiving applications, message type,
// and version information.
//
// Field positions follow the HL7 standard where MSH-1 is the field separator
// and MSH-2 contains the encoding characters.
type MSH struct {
	// FieldSeparator is MSH-1: The field separator character (typically "|").
	FieldSeparator string `hl7:"MSH.1"`

	// EncodingCharacters is MSH-2: The encoding characters (typically "^~\&").
	// Contains component separator, repetition separator, escape character, and subcomponent separator.
	EncodingCharacters string `hl7:"MSH.2"`

	// SendingApplication is MSH-3: Identifies the sending application.
	SendingApplication string `hl7:"MSH.3"`

	// SendingFacility is MSH-4: Identifies the sending facility.
	SendingFacility string `hl7:"MSH.4"`

	// ReceivingApplication is MSH-5: Identifies the receiving application.
	ReceivingApplication string `hl7:"MSH.5"`

	// ReceivingFacility is MSH-6: Identifies the receiving facility.
	ReceivingFacility string `hl7:"MSH.6"`

	// DateTime is MSH-7: Date/time of message creation (format: YYYYMMDDHHMMSS).
	DateTime string `hl7:"MSH.7"`

	// Security is MSH-8: Security information.
	Security string `hl7:"MSH.8"`

	// MessageType is MSH-9: Message type (e.g., "ADT^A01", "ORU^R01").
	// Format is typically MessageType^TriggerEvent^MessageStructure.
	MessageType string `hl7:"MSH.9"`

	// MessageControlID is MSH-10: Unique message identifier.
	MessageControlID string `hl7:"MSH.10"`

	// ProcessingID is MSH-11: Processing ID (P=Production, D=Debugging, T=Training).
	ProcessingID string `hl7:"MSH.11"`

	// VersionID is MSH-12: HL7 version number (e.g., "2.5.1").
	VersionID string `hl7:"MSH.12"`

	// SequenceNumber is MSH-13: Sequence number for message ordering.
	SequenceNumber string `hl7:"MSH.13"`

	// ContinuationPointer is MSH-14: Continuation pointer for fragmented messages.
	ContinuationPointer string `hl7:"MSH.14"`

	// AcceptAckType is MSH-15: Accept acknowledgment type (AL, NE, ER, SU).
	AcceptAckType string `hl7:"MSH.15"`

	// ApplicationAckType is MSH-16: Application acknowledgment type (AL, NE, ER, SU).
	ApplicationAckType string `hl7:"MSH.16"`

	// CountryCode is MSH-17: Country code (ISO 3166).
	CountryCode string `hl7:"MSH.17"`

	// CharacterSet is MSH-18: Character set identifier.
	CharacterSet string `hl7:"MSH.18"`

	// PrincipalLanguage is MSH-19: Principal language of message.
	PrincipalLanguage string `hl7:"MSH.19"`

	// AlternateCharacterSetHandling is MSH-20: Alternate character set handling scheme.
	AlternateCharacterSetHandling string `hl7:"MSH.20"`

	// MessageProfileIdentifier is MSH-21: Message profile identifier.
	MessageProfileIdentifier string `hl7:"MSH.21"`
}

// Errors for MSH segment operations.
var (
	ErrNilSegment    = errors.New("segment is nil")
	ErrNotMSHSegment = errors.New("segment is not MSH")
)

// ParseMSH extracts field values from an hl7.Segment into an MSH struct.
// Returns an error if the segment is nil or not an MSH segment.
func ParseMSH(seg hl7.Segment) (*MSH, error) {
	if seg == nil {
		return nil, ErrNilSegment
	}

	if seg.Name() != "MSH" {
		return nil, fmt.Errorf("%w: got %s", ErrNotMSHSegment, seg.Name())
	}

	msh := &MSH{}

	// MSH-1: Field separator (special handling - it's the delimiter itself)
	if f, ok := seg.Field(1); ok {
		msh.FieldSeparator = f.Value()
	}

	// MSH-2: Encoding characters (special handling - contains delimiter chars)
	if f, ok := seg.Field(2); ok {
		msh.EncodingCharacters = f.Value()
	}

	// MSH-3 through MSH-21
	if f, ok := seg.Field(3); ok {
		msh.SendingApplication = f.Value()
	}
	if f, ok := seg.Field(4); ok {
		msh.SendingFacility = f.Value()
	}
	if f, ok := seg.Field(5); ok {
		msh.ReceivingApplication = f.Value()
	}
	if f, ok := seg.Field(6); ok {
		msh.ReceivingFacility = f.Value()
	}
	if f, ok := seg.Field(7); ok {
		msh.DateTime = f.Value()
	}
	if f, ok := seg.Field(8); ok {
		msh.Security = f.Value()
	}
	if f, ok := seg.Field(9); ok {
		msh.MessageType = f.Value()
	}
	if f, ok := seg.Field(10); ok {
		msh.MessageControlID = f.Value()
	}
	if f, ok := seg.Field(11); ok {
		msh.ProcessingID = f.Value()
	}
	if f, ok := seg.Field(12); ok {
		msh.VersionID = f.Value()
	}
	if f, ok := seg.Field(13); ok {
		msh.SequenceNumber = f.Value()
	}
	if f, ok := seg.Field(14); ok {
		msh.ContinuationPointer = f.Value()
	}
	if f, ok := seg.Field(15); ok {
		msh.AcceptAckType = f.Value()
	}
	if f, ok := seg.Field(16); ok {
		msh.ApplicationAckType = f.Value()
	}
	if f, ok := seg.Field(17); ok {
		msh.CountryCode = f.Value()
	}
	if f, ok := seg.Field(18); ok {
		msh.CharacterSet = f.Value()
	}
	if f, ok := seg.Field(19); ok {
		msh.PrincipalLanguage = f.Value()
	}
	if f, ok := seg.Field(20); ok {
		msh.AlternateCharacterSetHandling = f.Value()
	}
	if f, ok := seg.Field(21); ok {
		msh.MessageProfileIdentifier = f.Value()
	}

	return msh, nil
}

// ToSegment converts the MSH struct into an hl7.Segment.
// The delims parameter specifies the delimiters to use for encoding.
// If delims is nil, default delimiters are used.
func (m *MSH) ToSegment(delims *hl7.Delimiters) (hl7.Segment, error) {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	// Build segment data as a string
	// MSH is special: MSH-1 is the field separator itself, MSH-2 is encoding chars
	fieldSep := string(delims.Field)
	if m.FieldSeparator != "" {
		fieldSep = m.FieldSeparator
	}

	encChars := delims.EncodingCharacters()
	if m.EncodingCharacters != "" {
		encChars = m.EncodingCharacters
	}

	// Build the segment string
	// Format: MSH|^~\&|field3|field4|...
	data := "MSH" + fieldSep + encChars

	// Append remaining fields (MSH-3 through MSH-21)
	fields := []string{
		m.SendingApplication,
		m.SendingFacility,
		m.ReceivingApplication,
		m.ReceivingFacility,
		m.DateTime,
		m.Security,
		m.MessageType,
		m.MessageControlID,
		m.ProcessingID,
		m.VersionID,
		m.SequenceNumber,
		m.ContinuationPointer,
		m.AcceptAckType,
		m.ApplicationAckType,
		m.CountryCode,
		m.CharacterSet,
		m.PrincipalLanguage,
		m.AlternateCharacterSetHandling,
		m.MessageProfileIdentifier,
	}

	// Find the last non-empty field to avoid trailing delimiters
	lastNonEmpty := -1
	for i := len(fields) - 1; i >= 0; i-- {
		if fields[i] != "" {
			lastNonEmpty = i
			break
		}
	}

	// Append fields up to and including the last non-empty field
	for i := 0; i <= lastNonEmpty; i++ {
		data += fieldSep + fields[i]
	}

	// Parse the constructed string back into a segment
	seg, err := hl7.ParseSegment([]rune(data), delims)
	if err != nil {
		return nil, fmt.Errorf("failed to create MSH segment: %w", err)
	}

	return seg, nil
}
