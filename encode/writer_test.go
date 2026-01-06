package encode_test

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/dshills/golevel7/encode"
	"github.com/dshills/golevel7/hl7"
	"github.com/dshills/golevel7/parse"
)

func TestWriter_Write(t *testing.T) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	err = writer.Write(msg)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	err = writer.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	// Verify the output is valid HL7
	reparsed, err := parser.Parse(buf.Bytes())
	if err != nil {
		t.Errorf("failed to parse written output: %v", err)
		t.Logf("output: %q", buf.String())
		return
	}

	if msg.ControlID() != reparsed.ControlID() {
		t.Errorf("control ID mismatch: expected %q, got %q", msg.ControlID(), reparsed.ControlID())
	}
}

func TestWriter_Write_MultipleMessages(t *testing.T) {
	parser := parse.New()

	msg1, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse ADT: %v", err)
	}

	msg2, err := parser.Parse([]byte(sampleORU))
	if err != nil {
		t.Fatalf("failed to parse ORU: %v", err)
	}

	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	// Write multiple messages
	if err := writer.Write(msg1); err != nil {
		t.Fatalf("Write(msg1) error = %v", err)
	}
	if err := writer.Write(msg2); err != nil {
		t.Fatalf("Write(msg2) error = %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// The buffer should contain both messages
	// They should be parseable (though together they form an invalid HL7 message)
	if buf.Len() == 0 {
		t.Error("expected non-empty buffer after writing two messages")
	}
}

func TestWriter_Write_NilMessage(t *testing.T) {
	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	err := writer.Write(nil)
	if err == nil {
		t.Error("expected error for nil message, got nil")
	}
}

func TestWriter_Write_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	msg := hl7.NewMessage(nil, nil)
	err := writer.Write(msg)
	if err == nil {
		t.Error("expected error for empty message, got nil")
	}
}

func TestWriter_Flush(t *testing.T) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	// Write without flush should buffer
	if err := writer.Write(msg); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Flush should write to underlying buffer
	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty buffer after flush")
	}
}

func TestWriter_Close(t *testing.T) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	if err := writer.Write(msg); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Writing after close should fail
	err = writer.Write(msg)
	if err == nil {
		t.Error("expected error when writing after close, got nil")
	}
}

func TestWriter_Close_Idempotent(t *testing.T) {
	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	// Close multiple times should not error
	if err := writer.Close(); err != nil {
		t.Errorf("first Close() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Errorf("second Close() error = %v", err)
	}
}

func TestWriter_Flush_AfterClose(t *testing.T) {
	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	err := writer.Flush()
	if err == nil {
		t.Error("expected error when flushing after close, got nil")
	}
}

func TestWriter_WithMLLP(t *testing.T) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	var buf bytes.Buffer
	writer := encode.NewWriter(&buf, encode.WithMLLP(true))

	if err := writer.Write(msg); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	output := buf.Bytes()
	if len(output) < 3 {
		t.Fatal("output too short for MLLP framing")
	}

	// Check MLLP framing
	if output[0] != 0x0B {
		t.Errorf("expected MLLP start block 0x0B, got 0x%02X", output[0])
	}
	if output[len(output)-2] != 0x1C {
		t.Errorf("expected MLLP end block 0x1C, got 0x%02X", output[len(output)-2])
	}
	if output[len(output)-1] != 0x0D {
		t.Errorf("expected MLLP CR 0x0D, got 0x%02X", output[len(output)-1])
	}
}

func TestWriter_WithLineEnding(t *testing.T) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	tests := []struct {
		name       string
		lineEnding string
	}{
		{"CR", "\r"},
		{"LF", "\n"},
		{"CRLF", "\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := encode.NewWriter(&buf, encode.WithLineEnding(tt.lineEnding))

			if err := writer.Write(msg); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}

			if !bytes.Contains(buf.Bytes(), []byte(tt.lineEnding)) {
				t.Errorf("output does not contain expected line ending %q", tt.lineEnding)
			}
		})
	}
}

func TestWriter_Concurrent(t *testing.T) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	// Write from multiple goroutines
	var wg sync.WaitGroup
	numGoroutines := 10
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := writer.Write(msg); err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Errorf("concurrent Write() error = %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Buffer should have content from all writes
	if buf.Len() == 0 {
		t.Error("expected non-empty buffer after concurrent writes")
	}
}

// slowWriter simulates a slow writer for testing.
type slowWriter struct {
	buf *bytes.Buffer
}

func (w *slowWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func TestWriter_SlowWriter(t *testing.T) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	var buf bytes.Buffer
	writer := encode.NewWriter(&slowWriter{buf: &buf})

	if err := writer.Write(msg); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty buffer")
	}
}

// failingWriter always fails on write for testing.
type failingWriter struct {
	err error
}

func (w *failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestWriter_WriteError(t *testing.T) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	writeErr := errors.New("write failed")
	fw := &failingWriter{err: writeErr}
	writer := encode.NewWriter(fw)

	// Write the message - the error may be caught during Write or Flush
	writeError := writer.Write(msg)

	// If Write succeeded, the error should be caught on Flush/Close
	if writeError == nil {
		err = writer.Close()
		if err == nil {
			t.Error("expected write error during Write or Close, got nil")
		}
	}
}

func TestWriter_RoundTrip(t *testing.T) {
	parser := parse.New()

	tests := []string{
		sampleADT,
		sampleORU,
		complexMessage,
	}

	for i, input := range tests {
		t.Run(string(rune('0'+i)), func(t *testing.T) {
			msg1, err := parser.Parse([]byte(input))
			if err != nil {
				t.Fatalf("first parse failed: %v", err)
			}

			var buf bytes.Buffer
			writer := encode.NewWriter(&buf)

			if err := writer.Write(msg1); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}

			msg2, err := parser.Parse(buf.Bytes())
			if err != nil {
				t.Fatalf("second parse failed: %v", err)
			}

			compareMessages(t, msg1, msg2)
		})
	}
}

func BenchmarkWriter_Write(b *testing.B) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		b.Fatalf("failed to parse input: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		writer := encode.NewWriter(&buf)
		if err := writer.Write(msg); err != nil {
			b.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriter_Write_Reuse(b *testing.B) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		b.Fatalf("failed to parse input: %v", err)
	}

	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := writer.Write(msg); err != nil {
			b.Fatal(err)
		}
		if err := writer.Flush(); err != nil {
			b.Fatal(err)
		}
		buf.Reset()
	}
}

func BenchmarkWriter_WriteMultiple(b *testing.B) {
	parser := parse.New()

	msg, err := parser.Parse([]byte(sampleADT))
	if err != nil {
		b.Fatalf("failed to parse input: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		writer := encode.NewWriter(&buf)
		for j := 0; j < 10; j++ {
			if err := writer.Write(msg); err != nil {
				b.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

// TestWriter_ImplementsInterface verifies the Writer implements io.Closer.
func TestWriter_ImplementsInterface(t *testing.T) {
	var buf bytes.Buffer
	writer := encode.NewWriter(&buf)

	// Writer should have a Close method
	var closer io.Closer = writer
	if closer == nil {
		t.Error("Writer does not implement io.Closer")
	}
}
