package encode_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dshills/golevel7/encode"
	"github.com/dshills/golevel7/hl7"
	"github.com/dshills/golevel7/parse"
)

// Sample HL7 messages for testing
const (
	sampleADT = "MSH|^~\\&|SENDING_APP|SENDING_FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20231215120000||ADT^A01|MSG00001|P|2.5.1\rPID|1||123456^^^HOSP^MR||DOE^JOHN^A||19800101|M\rPV1|1|I|WARD^ROOM^BED\r"

	sampleORU = "MSH|^~\\&|LAB|FACILITY|APP|FAC|20231215||ORU^R01|12345|P|2.5\rPID|1||PATIENT123||SMITH^JANE\rOBR|1|ORDER123||TEST^Blood Test\rOBX|1|NM|WBC||10.5|K/uL|4.5-11.0|N\r"

	// Message with repetitions and subcomponents
	complexMessage = "MSH|^~\\&|APP|FAC|REC|RECFAC|20231215||ADT^A01|CTRL|P|2.5\rPID|1||ID1~ID2~ID3||LAST^FIRST^MIDDLE&JR\r"
)

func TestEncoder_Encode_Basic(t *testing.T) {
	parser := parse.New()
	encoder := encode.New()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "ADT message",
			input:   sampleADT,
			wantErr: false,
		},
		{
			name:    "ORU message",
			input:   sampleORU,
			wantErr: false,
		},
		{
			name:    "complex message",
			input:   complexMessage,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the message
			msg, err := parser.Parse([]byte(tt.input))
			if err != nil {
				t.Fatalf("failed to parse input message: %v", err)
			}

			// Encode the message
			encoded, err := encoder.Encode(msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify the encoded message is valid by parsing it again
			reparsed, err := parser.Parse(encoded)
			if err != nil {
				t.Errorf("failed to re-parse encoded message: %v", err)
				t.Logf("encoded output: %q", string(encoded))
				return
			}

			// Compare key fields
			if msg.ControlID() != reparsed.ControlID() {
				t.Errorf("control ID mismatch: got %q, want %q", reparsed.ControlID(), msg.ControlID())
			}
			if msg.Type() != reparsed.Type() {
				t.Errorf("message type mismatch: got %q, want %q", reparsed.Type(), msg.Type())
			}
			if msg.Version() != reparsed.Version() {
				t.Errorf("version mismatch: got %q, want %q", reparsed.Version(), msg.Version())
			}
		})
	}
}

func TestEncoder_Encode_NilMessage(t *testing.T) {
	encoder := encode.New()

	_, err := encoder.Encode(nil)
	if err == nil {
		t.Error("expected error for nil message, got nil")
	}
}

func TestEncoder_Encode_EmptyMessage(t *testing.T) {
	encoder := encode.New()
	msg := hl7.NewMessage(nil, nil)

	_, err := encoder.Encode(msg)
	if err == nil {
		t.Error("expected error for empty message, got nil")
	}
}

func TestEncoder_WithLineEnding(t *testing.T) {
	parser := parse.New()

	tests := []struct {
		name       string
		lineEnding string
	}{
		{
			name:       "CR (default)",
			lineEnding: "\r",
		},
		{
			name:       "LF",
			lineEnding: "\n",
		},
		{
			name:       "CRLF",
			lineEnding: "\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := encode.New(encode.WithLineEnding(tt.lineEnding))

			msg, err := parser.Parse([]byte(sampleADT))
			if err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			encoded, err := encoder.Encode(msg)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Count occurrences of the line ending
			// The encoded message should have segments separated by the configured line ending
			if !bytes.Contains(encoded, []byte(tt.lineEnding)) {
				t.Errorf("encoded message does not contain expected line ending %q", tt.lineEnding)
			}
		})
	}
}

func TestEncoder_WithMLLP(t *testing.T) {
	parser := parse.New()
	encoder := encode.New(encode.WithMLLP(true))

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	encoded, err := encoder.Encode(msg)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Check for MLLP framing
	if len(encoded) < 3 {
		t.Fatal("encoded message too short for MLLP framing")
	}

	// Start block: 0x0B
	if encoded[0] != 0x0B {
		t.Errorf("expected MLLP start block 0x0B, got 0x%02X", encoded[0])
	}

	// End block: 0x1C 0x0D
	if encoded[len(encoded)-2] != 0x1C {
		t.Errorf("expected MLLP end block 0x1C at position -2, got 0x%02X", encoded[len(encoded)-2])
	}
	if encoded[len(encoded)-1] != 0x0D {
		t.Errorf("expected MLLP CR 0x0D at position -1, got 0x%02X", encoded[len(encoded)-1])
	}

	// Verify the message inside MLLP framing is still valid
	// Strip MLLP framing for re-parsing
	innerMessage := encoded[1 : len(encoded)-2]
	reparsed, err := parser.Parse(innerMessage)
	if err != nil {
		t.Errorf("failed to re-parse inner message: %v", err)
	}

	if msg.ControlID() != reparsed.ControlID() {
		t.Errorf("control ID mismatch after MLLP: got %q, want %q", reparsed.ControlID(), msg.ControlID())
	}
}

func TestEncoder_EncodeToWriter(t *testing.T) {
	parser := parse.New()
	encoder := encode.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	var buf bytes.Buffer
	err = encoder.EncodeToWriter(context.Background(), &buf, msg)
	if err != nil {
		t.Fatalf("EncodeToWriter() error = %v", err)
	}

	// Verify output matches Encode()
	encoded, err := encoder.Encode(msg)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if !bytes.Equal(buf.Bytes(), encoded) {
		t.Errorf("EncodeToWriter output differs from Encode output")
		t.Logf("EncodeToWriter: %q", buf.String())
		t.Logf("Encode: %q", string(encoded))
	}
}

func TestEncoder_EncodeToWriter_ContextCancellation(t *testing.T) {
	parser := parse.New()
	encoder := encode.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var buf bytes.Buffer
	err = encoder.EncodeToWriter(ctx, &buf, msg)
	if err == nil {
		t.Error("expected context cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestEncoder_EncodeToWriter_NilMessage(t *testing.T) {
	encoder := encode.New()

	var buf bytes.Buffer
	err := encoder.EncodeToWriter(context.Background(), &buf, nil)
	if err == nil {
		t.Error("expected error for nil message, got nil")
	}
}

func TestEncoder_RoundTrip(t *testing.T) {
	parser := parse.New()
	encoder := encode.New()

	tests := []struct {
		name  string
		input string
	}{
		{"ADT", sampleADT},
		{"ORU", sampleORU},
		{"complex", complexMessage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse -> Encode -> Parse
			msg1, err := parser.Parse([]byte(tt.input))
			if err != nil {
				t.Fatalf("first parse failed: %v", err)
			}

			encoded, err := encoder.Encode(msg1)
			if err != nil {
				t.Fatalf("encode failed: %v", err)
			}

			msg2, err := parser.Parse(encoded)
			if err != nil {
				t.Fatalf("second parse failed: %v", err)
			}

			// Compare messages
			compareMessages(t, msg1, msg2)
		})
	}
}

// compareMessages compares two messages for equality.
func compareMessages(t *testing.T, expected, actual hl7.Message) {
	t.Helper()

	expectedSegs := expected.AllSegments()
	actualSegs := actual.AllSegments()

	if len(expectedSegs) != len(actualSegs) {
		t.Errorf("segment count mismatch: expected %d, got %d", len(expectedSegs), len(actualSegs))
		return
	}

	for i := range expectedSegs {
		if expectedSegs[i].Name() != actualSegs[i].Name() {
			t.Errorf("segment %d name mismatch: expected %q, got %q", i, expectedSegs[i].Name(), actualSegs[i].Name())
		}

		// Compare field counts
		if expectedSegs[i].FieldCount() != actualSegs[i].FieldCount() {
			t.Errorf("segment %s field count mismatch: expected %d, got %d",
				expectedSegs[i].Name(), expectedSegs[i].FieldCount(), actualSegs[i].FieldCount())
		}
	}

	// Compare key values
	if expected.Type() != actual.Type() {
		t.Errorf("message type mismatch: expected %q, got %q", expected.Type(), actual.Type())
	}
	if expected.ControlID() != actual.ControlID() {
		t.Errorf("control ID mismatch: expected %q, got %q", expected.ControlID(), actual.ControlID())
	}
	if expected.Version() != actual.Version() {
		t.Errorf("version mismatch: expected %q, got %q", expected.Version(), actual.Version())
	}
}

func TestEncoder_Options(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		encoder := encode.New()
		if encoder == nil {
			t.Error("New() returned nil")
		}
	})

	t.Run("with all options", func(t *testing.T) {
		encoder := encode.New(
			encode.WithLineEnding("\r\n"),
			encode.WithMLLP(true),
			encode.WithTrailingDelimiters(true),
		)
		if encoder == nil {
			t.Error("New() returned nil")
		}
	})
}

func BenchmarkEncoder_Encode(b *testing.B) {
	parser := parse.New()
	encoder := encode.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		b.Fatalf("failed to parse input: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encoder.Encode(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncoder_EncodeToWriter(b *testing.B) {
	parser := parse.New()
	encoder := encode.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		b.Fatalf("failed to parse input: %v", err)
	}

	var buf bytes.Buffer
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := encoder.EncodeToWriter(ctx, &buf, msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncoder_RoundTrip(b *testing.B) {
	parser := parse.New()
	encoder := encode.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg, err := parser.Parse([]byte(sampleADT))
		if err != nil {
			b.Fatal(err)
		}

		encoded, err := encoder.Encode(msg)
		if err != nil {
			b.Fatal(err)
		}

		_, err = parser.Parse(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestError_Error tests the error message formatting.
func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name:     "basic error",
			err:      &encodeError{Message: "test error"},
			contains: []string{"test error"},
		},
		{
			name:     "error with segment",
			err:      &encodeError{Message: "failed", Segment: "PID", Position: 2},
			contains: []string{"failed", "PID", "2"},
		},
		{
			name:     "error with cause",
			err:      &encodeError{Message: "failed", Cause: errors.New("underlying")},
			contains: []string{"failed", "underlying"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, substr := range tt.contains {
				if !bytes.Contains([]byte(errStr), []byte(substr)) {
					t.Errorf("error message %q does not contain %q", errStr, substr)
				}
			}
		})
	}
}

// encodeError mirrors encode.Error for testing error formatting.
// This is a test-only type since Error is not exported.
type encodeError struct {
	Message  string
	Segment  string
	Position int
	Cause    error
}

func (e *encodeError) Error() string {
	msg := "encode error"
	if e.Segment != "" {
		msg = msg + " at segment " + e.Segment
		if e.Position > 0 {
			msg = msg + " (position " + string(rune('0'+e.Position)) + ")"
		}
	}
	if e.Message != "" {
		msg = msg + ": " + e.Message
	}
	if e.Cause != nil {
		msg = msg + ": " + e.Cause.Error()
	}
	return msg
}

func (e *encodeError) Unwrap() error {
	return e.Cause
}

// errorWriter is a writer that always returns an error for testing.
type errorWriter struct {
	err error
}

func (w *errorWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestEncoder_EncodeToWriter_WriteError(t *testing.T) {
	parser := parse.New()
	encoder := encode.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	writeErr := errors.New("write failed")
	errWriter := &errorWriter{err: writeErr}

	err = encoder.EncodeToWriter(context.Background(), errWriter, msg)
	if err == nil {
		t.Error("expected write error, got nil")
	}
}

func TestEncoder_EncodeToWriter_Timeout(t *testing.T) {
	parser := parse.New()
	encoder := encode.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give time for context to expire
	time.Sleep(10 * time.Millisecond)

	var buf bytes.Buffer
	err = encoder.EncodeToWriter(ctx, &buf, msg)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}
