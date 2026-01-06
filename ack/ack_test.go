package ack

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dshills/golevel7/hl7"
)

// mockMessage creates a simple mock message for testing.
func mockMessage(sendingApp, sendingFacility, receivingApp, receivingFacility, msgType, controlID, processingID, version string) hl7.Message {
	delims := hl7.DefaultDelimiters()
	msg := newSimpleMessage(delims)

	msh := newSimpleSegment("MSH", delims)
	msh.fields[3] = sendingApp
	msh.fields[4] = sendingFacility
	msh.fields[5] = receivingApp
	msh.fields[6] = receivingFacility
	msh.fields[7] = "20240101120000"
	msh.fields[9] = msgType
	msh.fields[10] = controlID
	msh.fields[11] = processingID
	msh.fields[12] = version

	_ = msg.AddSegment(msh)
	return msg
}

// mockADTMessage creates a mock ADT^A01 message for testing.
func mockADTMessage() hl7.Message {
	return mockMessage(
		"SENDING_APP",
		"SENDING_FACILITY",
		"RECEIVING_APP",
		"RECEIVING_FACILITY",
		"ADT^A01",
		"MSG001",
		"P",
		"2.5.1",
	)
}

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("NewBuilder() returned nil")
	}
}

func TestBuilder_Accept(t *testing.T) {
	// Use fixed time for deterministic testing
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	b := NewBuilder(
		WithTimeFunc(func() time.Time { return fixedTime }),
		WithControlIDFunc(func() string { return "ACK001" }),
	)

	original := mockADTMessage()
	ackMsg, err := b.Accept(original)
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	// Verify MSH segment exists
	msh, ok := ackMsg.Segment("MSH")
	if !ok {
		t.Fatal("ACK message missing MSH segment")
	}

	// Verify sending/receiving apps are swapped
	tests := []struct {
		field    string
		expected string
		desc     string
	}{
		{"3", "RECEIVING_APP", "MSH-3 (Sending App) should be original receiving app"},
		{"4", "RECEIVING_FACILITY", "MSH-4 (Sending Facility) should be original receiving facility"},
		{"5", "SENDING_APP", "MSH-5 (Receiving App) should be original sending app"},
		{"6", "SENDING_FACILITY", "MSH-6 (Receiving Facility) should be original sending facility"},
		{"7", "20240115103000", "MSH-7 should be current timestamp"},
		{"10", "ACK001", "MSH-10 should be ACK control ID"},
		{"11", "P", "MSH-11 should be copied from original"},
		{"12", "2.5.1", "MSH-12 should be copied from original"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := msh.Get(tt.field)
			if err != nil {
				t.Errorf("Get(%s) error = %v", tt.field, err)
				return
			}
			if got != tt.expected {
				t.Errorf("Get(%s) = %q, want %q", tt.field, got, tt.expected)
			}
		})
	}

	// Verify MSH-9 contains ACK
	msgType, _ := msh.Get("9")
	if !strings.HasPrefix(msgType, "ACK") {
		t.Errorf("MSH-9 = %q, want to start with 'ACK'", msgType)
	}

	// Verify MSA segment
	msa, ok := ackMsg.Segment("MSA")
	if !ok {
		t.Fatal("ACK message missing MSA segment")
	}

	msaTests := []struct {
		field    string
		expected string
		desc     string
	}{
		{"1", "AA", "MSA-1 should be AA (Application Accept)"},
		{"2", "MSG001", "MSA-2 should be original control ID"},
	}

	for _, tt := range msaTests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := msa.Get(tt.field)
			if err != nil {
				t.Errorf("Get(%s) error = %v", tt.field, err)
				return
			}
			if got != tt.expected {
				t.Errorf("Get(%s) = %q, want %q", tt.field, got, tt.expected)
			}
		})
	}

	// Accept should NOT have ERR segment
	_, hasERR := ackMsg.Segment("ERR")
	if hasERR {
		t.Error("Accept ACK should not have ERR segment")
	}
}

func TestBuilder_Reject(t *testing.T) {
	b := NewBuilder(
		WithControlIDFunc(func() string { return "ACK002" }),
	)

	original := mockADTMessage()
	reason := "Message format not supported"

	ackMsg, err := b.Reject(original, reason)
	if err != nil {
		t.Fatalf("Reject() error = %v", err)
	}

	// Verify MSA segment
	msa, ok := ackMsg.Segment("MSA")
	if !ok {
		t.Fatal("ACK message missing MSA segment")
	}

	// MSA-1 should be AR
	code, _ := msa.Get("1")
	if code != "AR" {
		t.Errorf("MSA-1 = %q, want AR", code)
	}

	// MSA-2 should be original control ID
	ctrlID, _ := msa.Get("2")
	if ctrlID != "MSG001" {
		t.Errorf("MSA-2 = %q, want MSG001", ctrlID)
	}

	// MSA-3 should be the reason
	text, _ := msa.Get("3")
	if text != reason {
		t.Errorf("MSA-3 = %q, want %q", text, reason)
	}
}

func TestBuilder_Error(t *testing.T) {
	b := NewBuilder(
		WithControlIDFunc(func() string { return "ACK003" }),
	)

	original := mockADTMessage()
	testErr := errors.New("database connection failed")

	ackMsg, err := b.Error(original, testErr)
	if err != nil {
		t.Fatalf("Error() error = %v", err)
	}

	// Verify MSA segment
	msa, ok := ackMsg.Segment("MSA")
	if !ok {
		t.Fatal("ACK message missing MSA segment")
	}

	// MSA-1 should be AE
	code, _ := msa.Get("1")
	if code != "AE" {
		t.Errorf("MSA-1 = %q, want AE", code)
	}

	// MSA-2 should be original control ID
	ctrlID, _ := msa.Get("2")
	if ctrlID != "MSG001" {
		t.Errorf("MSA-2 = %q, want MSG001", ctrlID)
	}

	// MSA-3 should contain error message
	text, _ := msa.Get("3")
	if text != testErr.Error() {
		t.Errorf("MSA-3 = %q, want %q", text, testErr.Error())
	}

	// Error ACK should have ERR segment
	errSeg, hasERR := ackMsg.Segment("ERR")
	if !hasERR {
		t.Fatal("Error ACK should have ERR segment")
	}

	// ERR-3 should have error code
	errCode, _ := errSeg.Get("3")
	if errCode != "207" {
		t.Errorf("ERR-3 = %q, want 207", errCode)
	}

	// ERR-4 should have severity
	severity, _ := errSeg.Get("4")
	if severity != "E" {
		t.Errorf("ERR-4 = %q, want E", severity)
	}

	// ERR-7 should have diagnostic info
	diagInfo, _ := errSeg.Get("7")
	if diagInfo != testErr.Error() {
		t.Errorf("ERR-7 = %q, want %q", diagInfo, testErr.Error())
	}
}

func TestBuilder_Custom(t *testing.T) {
	b := NewBuilder(
		WithControlIDFunc(func() string { return "ACK004" }),
	)

	original := mockADTMessage()

	customACK := ACK{
		Code:          CommitAccept,
		ControlID:     "MSG001",
		TextMessage:   "Message committed successfully",
		ErrorCode:     "",
		ErrorLocation: "",
	}

	ackMsg, err := b.Custom(original, customACK)
	if err != nil {
		t.Fatalf("Custom() error = %v", err)
	}

	// Verify MSA segment
	msa, ok := ackMsg.Segment("MSA")
	if !ok {
		t.Fatal("ACK message missing MSA segment")
	}

	// MSA-1 should be CA
	code, _ := msa.Get("1")
	if code != "CA" {
		t.Errorf("MSA-1 = %q, want CA", code)
	}

	// MSA-3 should have text message
	text, _ := msa.Get("3")
	if text != customACK.TextMessage {
		t.Errorf("MSA-3 = %q, want %q", text, customACK.TextMessage)
	}
}

func TestBuilder_CustomWithERRSegment(t *testing.T) {
	b := NewBuilder(
		WithControlIDFunc(func() string { return "ACK005" }),
	)

	original := mockADTMessage()

	customACK := ACK{
		Code:          ApplicationError,
		ControlID:     "MSG001",
		TextMessage:   "Validation failed",
		ErrorCode:     "101",
		ErrorLocation: "PID-3-1",
		ErrorMessage:  "Patient ID is required",
		Severity:      "E",
	}

	ackMsg, err := b.Custom(original, customACK)
	if err != nil {
		t.Fatalf("Custom() error = %v", err)
	}

	// Verify ERR segment exists
	errSeg, ok := ackMsg.Segment("ERR")
	if !ok {
		t.Fatal("ACK message should have ERR segment")
	}

	tests := []struct {
		field    string
		expected string
		desc     string
	}{
		{"1", "PID-3-1", "ERR-1 should have error location"},
		{"2", "PID-3-1", "ERR-2 should have error location"},
		{"3", "101", "ERR-3 should have error code"},
		{"4", "E", "ERR-4 should have severity"},
		{"7", "Patient ID is required", "ERR-7 should have diagnostic info"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := errSeg.Get(tt.field)
			if err != nil {
				t.Errorf("Get(%s) error = %v", tt.field, err)
				return
			}
			if got != tt.expected {
				t.Errorf("Get(%s) = %q, want %q", tt.field, got, tt.expected)
			}
		})
	}
}

func TestBuilder_NilMessage(t *testing.T) {
	b := NewBuilder()

	_, err := b.Accept(nil)
	if !errors.Is(err, ErrNilMessage) {
		t.Errorf("Accept(nil) error = %v, want %v", err, ErrNilMessage)
	}

	_, err = b.Reject(nil, "reason")
	if !errors.Is(err, ErrNilMessage) {
		t.Errorf("Reject(nil) error = %v, want %v", err, ErrNilMessage)
	}

	_, err = b.Error(nil, errors.New("test"))
	if !errors.Is(err, ErrNilMessage) {
		t.Errorf("Error(nil) error = %v, want %v", err, ErrNilMessage)
	}

	_, err = b.Custom(nil, ACK{Code: ApplicationAccept})
	if !errors.Is(err, ErrNilMessage) {
		t.Errorf("Custom(nil) error = %v, want %v", err, ErrNilMessage)
	}
}

func TestBuilder_MissingControlID(t *testing.T) {
	b := NewBuilder()

	// Create message without control ID
	original := mockMessage("APP", "FAC", "APP2", "FAC2", "ADT^A01", "", "P", "2.5")

	_, err := b.Accept(original)
	if !errors.Is(err, ErrMissingControlID) {
		t.Errorf("Accept() error = %v, want %v", err, ErrMissingControlID)
	}
}

func TestBuilder_InvalidACKCode(t *testing.T) {
	b := NewBuilder()
	original := mockADTMessage()

	_, err := b.Custom(original, ACK{
		Code:      Code("XX"), // Invalid code
		ControlID: "MSG001",
	})
	if !errors.Is(err, ErrInvalidACKCode) {
		t.Errorf("Custom() error = %v, want %v", err, ErrInvalidACKCode)
	}
}

func TestBuilder_MSHFieldsSwapped(t *testing.T) {
	b := NewBuilder(
		WithControlIDFunc(func() string { return "ACK006" }),
	)

	// Create original message with distinct values
	original := mockMessage(
		"HOSPITAL_HIS",
		"MAIN_CAMPUS",
		"LAB_LIS",
		"LAB_WEST",
		"ORM^O01",
		"ORDER123",
		"T",
		"2.4",
	)

	ackMsg, err := b.Accept(original)
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	msh, _ := ackMsg.Segment("MSH")

	// Verify the swap
	swapTests := []struct {
		field    string
		expected string
		desc     string
	}{
		{"3", "LAB_LIS", "Sending App should be original Receiving App"},
		{"4", "LAB_WEST", "Sending Facility should be original Receiving Facility"},
		{"5", "HOSPITAL_HIS", "Receiving App should be original Sending App"},
		{"6", "MAIN_CAMPUS", "Receiving Facility should be original Sending Facility"},
	}

	for _, tt := range swapTests {
		t.Run(tt.desc, func(t *testing.T) {
			got, _ := msh.Get(tt.field)
			if got != tt.expected {
				t.Errorf("MSH-%s = %q, want %q", tt.field, got, tt.expected)
			}
		})
	}
}

func TestBuilder_MessageBytes(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)

	b := NewBuilder(
		WithTimeFunc(func() time.Time { return fixedTime }),
		WithControlIDFunc(func() string { return "ACK123" }),
	)

	original := mockADTMessage()
	ackMsg, err := b.Accept(original)
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	bytes := ackMsg.Bytes()
	msgStr := string(bytes)

	// Verify message structure
	if !strings.HasPrefix(msgStr, "MSH|") {
		t.Error("Message should start with MSH|")
	}

	if !strings.Contains(msgStr, "MSA|AA|MSG001") {
		t.Error("Message should contain MSA segment with AA code and original control ID")
	}

	// Verify segments are terminated properly
	segments := strings.Split(strings.TrimSuffix(msgStr, "\r"), "\r")
	if len(segments) < 2 {
		t.Errorf("Expected at least 2 segments, got %d", len(segments))
	}
}

func TestACKCode_Methods(t *testing.T) {
	tests := []struct {
		code     Code
		isAccept bool
		isError  bool
		isReject bool
		isValid  bool
	}{
		{ApplicationAccept, true, false, false, true},
		{ApplicationError, false, true, false, true},
		{ApplicationReject, false, false, true, true},
		{CommitAccept, true, false, false, true},
		{CommitError, false, true, false, true},
		{CommitReject, false, false, true, true},
		{Code("XX"), false, false, false, false},
		{Code(""), false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := tt.code.IsAccept(); got != tt.isAccept {
				t.Errorf("IsAccept() = %v, want %v", got, tt.isAccept)
			}
			if got := tt.code.IsError(); got != tt.isError {
				t.Errorf("IsError() = %v, want %v", got, tt.isError)
			}
			if got := tt.code.IsReject(); got != tt.isReject {
				t.Errorf("IsReject() = %v, want %v", got, tt.isReject)
			}
			if got := tt.code.IsValid(); got != tt.isValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

func TestACK_NeedsERRSegment(t *testing.T) {
	tests := []struct {
		name     string
		ack      ACK
		expected bool
	}{
		{
			name:     "Accept without error",
			ack:      ACK{Code: ApplicationAccept},
			expected: false,
		},
		{
			name:     "Accept with error info (should not have ERR)",
			ack:      ACK{Code: ApplicationAccept, ErrorCode: "100"},
			expected: false, // Accept codes don't need ERR
		},
		{
			name:     "Error with error code",
			ack:      ACK{Code: ApplicationError, ErrorCode: "100"},
			expected: true,
		},
		{
			name:     "Reject with error location",
			ack:      ACK{Code: ApplicationReject, ErrorLocation: "PID-3"},
			expected: true,
		},
		{
			name:     "Error without error info",
			ack:      ACK{Code: ApplicationError},
			expected: false,
		},
		{
			name:     "Commit error with message",
			ack:      ACK{Code: CommitError, ErrorMessage: "Storage full"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ack.NeedsERRSegment(); got != tt.expected {
				t.Errorf("NeedsERRSegment() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewAcceptACK(t *testing.T) {
	ack := NewAcceptACK("CTRL123")

	if ack.Code != ApplicationAccept {
		t.Errorf("Code = %v, want %v", ack.Code, ApplicationAccept)
	}
	if ack.ControlID != "CTRL123" {
		t.Errorf("ControlID = %v, want CTRL123", ack.ControlID)
	}
}

func TestNewErrorACK(t *testing.T) {
	ack := NewErrorACK("CTRL456", "102", "Data type error")

	if ack.Code != ApplicationError {
		t.Errorf("Code = %v, want %v", ack.Code, ApplicationError)
	}
	if ack.ControlID != "CTRL456" {
		t.Errorf("ControlID = %v, want CTRL456", ack.ControlID)
	}
	if ack.ErrorCode != "102" {
		t.Errorf("ErrorCode = %v, want 102", ack.ErrorCode)
	}
	if ack.TextMessage != "Data type error" {
		t.Errorf("TextMessage = %v, want 'Data type error'", ack.TextMessage)
	}
	if ack.Severity != "E" {
		t.Errorf("Severity = %v, want E", ack.Severity)
	}
}

func TestNewRejectACK(t *testing.T) {
	ack := NewRejectACK("CTRL789", "Unsupported message type")

	if ack.Code != ApplicationReject {
		t.Errorf("Code = %v, want %v", ack.Code, ApplicationReject)
	}
	if ack.ControlID != "CTRL789" {
		t.Errorf("ControlID = %v, want CTRL789", ack.ControlID)
	}
	if ack.TextMessage != "Unsupported message type" {
		t.Errorf("TextMessage = %v, want 'Unsupported message type'", ack.TextMessage)
	}
}

// TestBuilder_WithMessageFactory tests using a custom message factory
func TestBuilder_WithMessageFactory(t *testing.T) {
	factory := &testMessageFactory{}

	b := NewBuilder(
		WithMessageFactory(factory),
		WithControlIDFunc(func() string { return "ACK007" }),
	)

	original := mockADTMessage()
	_, err := b.Accept(original)
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	// Verify factory was used
	if !factory.newMessageCalled {
		t.Error("MessageFactory.NewMessage() was not called")
	}
	if factory.newSegmentCalls < 2 { // At least MSH and MSA
		t.Errorf("MessageFactory.NewSegment() called %d times, want at least 2", factory.newSegmentCalls)
	}
}

// testMessageFactory is a mock MessageFactory for testing
type testMessageFactory struct {
	newMessageCalled bool
	newSegmentCalls  int
}

func (f *testMessageFactory) NewMessage(delims *hl7.Delimiters) hl7.Message {
	f.newMessageCalled = true
	return newSimpleMessage(delims)
}

func (f *testMessageFactory) NewSegment(name string, delims *hl7.Delimiters) hl7.Segment {
	f.newSegmentCalls++
	return newSimpleSegment(name, delims)
}
