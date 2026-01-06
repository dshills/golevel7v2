package parse

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/dshills/golevel7/hl7"
)

func TestNewScanner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []ParserOption
	}{
		{
			name: "default scanner",
			opts: nil,
		},
		{
			name: "with parser options",
			opts: []ParserOption{WithStrictMode(true)},
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := strings.NewReader("")
			s := NewScanner(r, tt.opts...)
			if s == nil {
				t.Fatal("NewScanner() returned nil")
			}
		})
	}
}

func TestScanner_Scan_SingleMessage(t *testing.T) {
	t.Parallel()

	input := "MSH|^~\\&|SENDING|FACILITY|||202301011200||ADT^A01|MSG001|P|2.5\rPID|1||12345\r"
	r := strings.NewReader(input)
	s := NewScanner(r)

	// First scan should succeed
	if !s.Scan() {
		t.Fatalf("expected Scan() to return true, got error: %v", s.Err())
	}

	msg := s.Message()
	if msg == nil {
		t.Fatal("Message() returned nil")
	}

	segs := msg.AllSegments()
	if len(segs) != 2 {
		t.Errorf("expected 2 segments, got %d", len(segs))
	}

	// Second scan should return false (no more messages)
	if s.Scan() {
		t.Fatal("expected Scan() to return false for empty reader")
	}

	if s.Err() != nil {
		t.Errorf("unexpected error: %v", s.Err())
	}
}

func TestScanner_Scan_MultipleMessages(t *testing.T) {
	t.Parallel()

	// Multiple messages separated by double CR
	input := "MSH|^~\\&|APP1|FAC1|||202301011200||ADT^A01|MSG001|P|2.5\rPID|1||12345\r\r" +
		"MSH|^~\\&|APP2|FAC2|||202301011201||ORU^R01|MSG002|P|2.5\rOBX|1|NM|WBC||7.5\r"

	r := strings.NewReader(input)
	s := NewScanner(r)

	// First message
	if !s.Scan() {
		t.Fatalf("first Scan() failed: %v", s.Err())
	}

	msg1 := s.Message()
	if msg1 == nil {
		t.Fatal("first message is nil")
	}
	if msg1.ControlID() != "MSG001" {
		t.Errorf("expected MSG001, got %s", msg1.ControlID())
	}

	// Second message
	if !s.Scan() {
		t.Fatalf("second Scan() failed: %v", s.Err())
	}

	msg2 := s.Message()
	if msg2 == nil {
		t.Fatal("second message is nil")
	}
	if msg2.ControlID() != "MSG002" {
		t.Errorf("expected MSG002, got %s", msg2.ControlID())
	}

	// No more messages
	if s.Scan() {
		t.Fatal("expected no more messages")
	}
}

func TestScanner_Scan_MLLPFramed(t *testing.T) {
	t.Parallel()

	// Single MLLP-framed message
	input := []byte{
		0x0B, // Start byte
		'M', 'S', 'H', '|', '^', '~', '\\', '&', '|', 'S', 'E', 'N', 'D', '|', 'F', 'A', 'C', '|', '|', '|', '2', '0', '2', '3', '|', '|', 'A', 'D', 'T', '|', 'M', 'S', 'G', '|', 'P', '|', '2', '.', '5', '\r',
		0x1C, 0x0D, // End bytes
	}

	r := bytes.NewReader(input)
	s := NewScanner(r)

	if !s.Scan() {
		t.Fatalf("Scan() failed: %v", s.Err())
	}

	msg := s.Message()
	if msg == nil {
		t.Fatal("Message() returned nil")
	}

	msh, ok := msg.Segment("MSH")
	if !ok {
		t.Fatal("MSH segment not found")
	}
	if msh.Name() != "MSH" {
		t.Errorf("expected MSH, got %s", msh.Name())
	}
}

func TestScanner_Scan_MultipleMLLP(t *testing.T) {
	t.Parallel()

	// Two MLLP-framed messages
	msg1 := []byte{0x0B, 'M', 'S', 'H', '|', '^', '~', '\\', '&', '|', '|', '|', '|', '|', '2', '0', '2', '3', '|', '|', 'A', 'D', 'T', '|', 'M', '1', '|', 'P', '|', '2', '.', '5', '\r', 0x1C, 0x0D}
	msg2 := []byte{0x0B, 'M', 'S', 'H', '|', '^', '~', '\\', '&', '|', '|', '|', '|', '|', '2', '0', '2', '3', '|', '|', 'O', 'R', 'U', '|', 'M', '2', '|', 'P', '|', '2', '.', '5', '\r', 0x1C, 0x0D}

	input := make([]byte, 0, len(msg1)+len(msg2))
	input = append(input, msg1...)
	input = append(input, msg2...)
	r := bytes.NewReader(input)
	s := NewScanner(r)

	// First message
	if !s.Scan() {
		t.Fatalf("first Scan() failed: %v", s.Err())
	}
	if s.Message().ControlID() != "M1" {
		t.Errorf("expected M1, got %s", s.Message().ControlID())
	}

	// Second message
	if !s.Scan() {
		t.Fatalf("second Scan() failed: %v", s.Err())
	}
	if s.Message().ControlID() != "M2" {
		t.Errorf("expected M2, got %s", s.Message().ControlID())
	}

	// No more messages
	if s.Scan() {
		t.Fatal("expected no more messages")
	}
}

func TestScanner_Scan_EmptyReader(t *testing.T) {
	t.Parallel()

	r := strings.NewReader("")
	s := NewScanner(r)

	if s.Scan() {
		t.Fatal("expected Scan() to return false for empty reader")
	}

	if s.Message() != nil {
		t.Error("expected Message() to return nil")
	}

	// EOF is not an error
	if s.Err() != nil {
		t.Errorf("unexpected error: %v", s.Err())
	}
}

func TestScanner_Scan_InvalidMessage(t *testing.T) {
	t.Parallel()

	// Invalid message (no MSH)
	input := "PID|1||12345\r"
	r := strings.NewReader(input)
	s := NewScanner(r)

	if s.Scan() {
		t.Fatal("expected Scan() to return false for invalid message")
	}

	if s.Err() == nil {
		t.Fatal("expected error for invalid message")
	}
}

func TestScanner_Err(t *testing.T) {
	t.Parallel()

	// Valid message should have no error
	input := "MSH|^~\\&|SEND|FAC|||2023||ADT|MSG|P|2.5\r"
	r := strings.NewReader(input)
	s := NewScanner(r)

	s.Scan()

	if s.Err() != nil {
		t.Errorf("unexpected error: %v", s.Err())
	}
}

func TestScanner_Message_BeforeScan(t *testing.T) {
	t.Parallel()

	r := strings.NewReader("MSH|^~\\&|SEND|FAC|||2023||ADT|MSG|P|2.5\r")
	s := NewScanner(r)

	// Message should be nil before Scan is called
	if s.Message() != nil {
		t.Error("Message() should return nil before Scan()")
	}
}

func TestNewScannerWithOptions(t *testing.T) {
	t.Parallel()

	r := strings.NewReader("MSH|^~\\&|SEND|FAC|||2023||ADT|MSG|P|2.5\r")

	parserOpts := []ParserOption{WithStrictMode(true)}
	scannerOpts := []ScannerOption{WithMaxMessageSize(1024)}

	s := NewScannerWithOptions(r, parserOpts, scannerOpts...)

	if s == nil {
		t.Fatal("NewScannerWithOptions() returned nil")
	}

	if !s.Scan() {
		t.Fatalf("Scan() failed: %v", s.Err())
	}

	if s.Message() == nil {
		t.Fatal("Message() returned nil")
	}
}

func TestScanner_MaxMessageSize(t *testing.T) {
	t.Parallel()

	// Create a message that exceeds max size
	var sb strings.Builder
	sb.WriteString("MSH|^~\\&|SEND|FAC|||2023||ADT|MSG|P|2.5\r")
	for i := 0; i < 100; i++ {
		sb.WriteString("NTE|1|" + strings.Repeat("X", 100) + "\r")
	}
	input := sb.String()

	r := strings.NewReader(input)
	s := NewScannerWithOptions(r, nil, WithMaxMessageSize(1000))

	if s.Scan() {
		t.Fatal("expected Scan() to fail for oversized message")
	}

	if s.Err() == nil {
		t.Fatal("expected error for oversized message")
	}
}

func TestScanner_PlainMessage_NewMSHBoundary(t *testing.T) {
	t.Parallel()

	// Two messages where second starts with MSH right after CR
	input := "MSH|^~\\&|APP1|FAC1|||2023||ADT^A01|MSG001|P|2.5\rPID|1||12345\r" +
		"MSH|^~\\&|APP2|FAC2|||2023||ORU^R01|MSG002|P|2.5\rOBX|1|NM|WBC||7.5\r"

	r := strings.NewReader(input)
	s := NewScanner(r)

	// First message
	if !s.Scan() {
		t.Fatalf("first Scan() failed: %v", s.Err())
	}

	msg1 := s.Message()
	if msg1 == nil {
		t.Fatal("first message is nil")
	}
	if msg1.ControlID() != "MSG001" {
		t.Errorf("first message: expected MSG001, got %s", msg1.ControlID())
	}

	// Second message
	if !s.Scan() {
		t.Fatalf("second Scan() failed: %v", s.Err())
	}

	msg2 := s.Message()
	if msg2 == nil {
		t.Fatal("second message is nil")
	}
	if msg2.ControlID() != "MSG002" {
		t.Errorf("second message: expected MSG002, got %s", msg2.ControlID())
	}

	// No more messages
	if s.Scan() {
		t.Fatal("expected no more messages")
	}
}

// errReader always returns an error after reading some data
type errReader struct {
	data    string
	pos     int
	errAt   int
	errType error
}

func (r *errReader) Read(p []byte) (n int, err error) {
	if r.pos >= r.errAt {
		return 0, r.errType
	}

	remaining := r.data[r.pos:]
	toRead := len(remaining)
	if toRead > len(p) {
		toRead = len(p)
	}
	if r.pos+toRead > r.errAt {
		toRead = r.errAt - r.pos
	}

	n = copy(p, remaining[:toRead])
	r.pos += n

	if r.pos >= r.errAt {
		return n, r.errType
	}

	return n, nil
}

func TestScanner_ReadError(t *testing.T) {
	t.Parallel()

	// Create a reader that errors partway through
	testErr := io.ErrUnexpectedEOF
	r := &errReader{
		data:    "MSH|^~\\&|SEND|FAC|||2023||ADT|MSG|P|2.5\rPID|1||12345\r",
		errAt:   5, // Error after reading "MSH|^"
		errType: testErr,
	}

	s := NewScanner(r)

	// Scan should fail due to the read error
	result := s.Scan()

	// Either result is false with error, or parsing fails
	if result && s.Err() == nil {
		t.Fatal("expected Scan() to fail or return error")
	}
}

// Benchmark tests
func BenchmarkScanner_SingleMessage(b *testing.B) {
	input := "MSH|^~\\&|SENDING|FACILITY|||202301011200||ADT^A01|MSG001|P|2.5\rPID|1||12345^^^MRN||Doe^John^A||19800101|M\r"

	for i := 0; i < b.N; i++ {
		r := strings.NewReader(input)
		s := NewScanner(r)
		for s.Scan() {
			_ = s.Message()
		}
		if s.Err() != nil {
			b.Fatal(s.Err())
		}
	}
}

func BenchmarkScanner_MultipleMessages(b *testing.B) {
	input := "MSH|^~\\&|APP1|FAC1|||2023||ADT|MSG001|P|2.5\rPID|1||12345\r\r" +
		"MSH|^~\\&|APP2|FAC2|||2023||ORU|MSG002|P|2.5\rOBX|1|NM|WBC||7.5\r\r" +
		"MSH|^~\\&|APP3|FAC3|||2023||ACK|MSG003|P|2.5\r"

	for i := 0; i < b.N; i++ {
		r := strings.NewReader(input)
		s := NewScanner(r)
		count := 0
		for s.Scan() {
			_ = s.Message()
			count++
		}
		if s.Err() != nil {
			b.Fatal(s.Err())
		}
	}
}

func BenchmarkScanner_MLLPFramed(b *testing.B) {
	msg := []byte{
		0x0B,
		'M', 'S', 'H', '|', '^', '~', '\\', '&', '|', 'S', 'E', 'N', 'D', '|', 'F', 'A', 'C', '|', '|', '|', '2', '0', '2', '3', '|', '|', 'A', 'D', 'T', '|', 'M', 'S', 'G', '|', 'P', '|', '2', '.', '5', '\r',
		'P', 'I', 'D', '|', '1', '|', '|', '1', '2', '3', '4', '5', '\r',
		0x1C, 0x0D,
	}

	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(msg)
		s := NewScanner(r)
		for s.Scan() {
			_ = s.Message()
		}
		if s.Err() != nil {
			b.Fatal(s.Err())
		}
	}
}

func BenchmarkScanner_LargeMessage(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("MSH|^~\\&|SENDING|FACILITY|||202301011200||ORU^R01|MSG|P|2.5\r")
	sb.WriteString("PID|1||12345^^^MRN||Doe^John^A||19800101|M\r")
	for i := 0; i < 100; i++ {
		sb.WriteString("OBX|1|NM|WBC||7.5|10*3/uL|4.5-11.0|N|||F\r")
	}
	input := sb.String()

	for i := 0; i < b.N; i++ {
		r := strings.NewReader(input)
		s := NewScanner(r)
		for s.Scan() {
			_ = s.Message()
		}
		if s.Err() != nil {
			b.Fatal(s.Err())
		}
	}
}

// Example tests
func ExampleNewScanner() {
	input := "MSH|^~\\&|SEND|FAC|||2023||ADT^A01|MSG001|P|2.5\rPID|1||12345\r"
	r := strings.NewReader(input)

	scanner := NewScanner(r)
	for scanner.Scan() {
		msg := scanner.Message()
		_ = msg // Process message
	}

	if err := scanner.Err(); err != nil {
		_ = err // Handle error in real code
	}
}

func ExampleScanner_multipleMessages() {
	// Stream containing multiple HL7 messages
	input := "MSH|^~\\&|APP|FAC|||2023||ADT|MSG1|P|2.5\rPID|1||12345\r\r" +
		"MSH|^~\\&|APP|FAC|||2023||ORU|MSG2|P|2.5\rOBX|1|NM|WBC||7.5\r"

	r := strings.NewReader(input)
	scanner := NewScanner(r)

	messages := []hl7.Message{}
	for scanner.Scan() {
		messages = append(messages, scanner.Message())
	}

	// messages now contains both parsed messages
	_ = messages
}
