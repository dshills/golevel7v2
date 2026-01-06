package parse

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dshills/golevel7/hl7"
)

// Sample HL7 messages for testing
const (
	simpleADT = "MSH|^~\\&|SENDING|FACILITY|RECEIVING|FACILITY|202301011200||ADT^A01|MSG001|P|2.5\rPID|1||12345^^^MRN||Doe^John^A||19800101|M\r"

	mllpFramedADT = "\x0BMSH|^~\\&|SENDING|FACILITY|RECEIVING|FACILITY|202301011200||ADT^A01|MSG001|P|2.5\rPID|1||12345^^^MRN||Doe^John^A||19800101|M\r\x1C\x0D"

	oru = "MSH|^~\\&|LAB|HOSPITAL|HIS|HOSPITAL|202301011200||ORU^R01|MSG002|P|2.5\rPID|1||67890^^^MRN||Smith^Jane||19750515|F\rOBR|1|ORD001|ACC001|CBC^Complete Blood Count\rOBX|1|NM|WBC^White Blood Cell Count||7.5|10*3/uL|4.5-11.0|N|||F\r"

	mshOnly = "MSH|^~\\&|SENDING|FACILITY|RECEIVING|FACILITY|202301011200||ACK|MSG003|P|2.5\r"

	noTerminator = "MSH|^~\\&|SENDING|FACILITY|RECEIVING|FACILITY|202301011200||ADT^A01|MSG004|P|2.5"

	emptySegment = "MSH|^~\\&|SENDING|FACILITY|||202301011200||ADT^A01|MSG005|P|2.5\r\rPID|1||12345\r"

	customDelimiters = "MSH$#~\\@$SENDING$FACILITY$RECEIVING$FACILITY$202301011200$$ADT#A01$MSG006$P$2.5\rPID$1$$12345###MRN$$Doe#John#A$$19800101$M\r"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []ParserOption
	}{
		{
			name: "default parser",
			opts: nil,
		},
		{
			name: "with strict mode",
			opts: []ParserOption{WithStrictMode(true)},
		},
		{
			name: "with custom max segments",
			opts: []ParserOption{WithMaxSegments(100)},
		},
		{
			name: "with multiple options",
			opts: []ParserOption{
				WithStrictMode(true),
				WithMaxSegments(500),
				WithMaxFieldLength(32768),
				WithAllowEmptySegments(true),
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := New(tt.opts...)
			if p == nil {
				t.Fatal("New() returned nil")
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		opts        []ParserOption
		wantErr     bool
		errContains string
		validate    func(*testing.T, hl7.Message)
	}{
		{
			name:    "simple ADT message",
			input:   simpleADT,
			wantErr: false,
			validate: func(t *testing.T, msg hl7.Message) {
				if msg == nil {
					t.Fatal("message is nil")
				}

				segs := msg.AllSegments()
				if len(segs) != 2 {
					t.Errorf("expected 2 segments, got %d", len(segs))
				}

				msh, ok := msg.Segment("MSH")
				if !ok {
					t.Fatal("MSH segment not found")
				}
				if msh.Name() != "MSH" {
					t.Errorf("expected MSH, got %s", msh.Name())
				}

				pid, ok := msg.Segment("PID")
				if !ok {
					t.Fatal("PID segment not found")
				}
				if pid.Name() != "PID" {
					t.Errorf("expected PID, got %s", pid.Name())
				}
			},
		},
		{
			name:    "MLLP framed message",
			input:   mllpFramedADT,
			wantErr: false,
			validate: func(t *testing.T, msg hl7.Message) {
				if msg == nil {
					t.Fatal("message is nil")
				}

				segs := msg.AllSegments()
				if len(segs) != 2 {
					t.Errorf("expected 2 segments, got %d", len(segs))
				}
			},
		},
		{
			name:    "ORU message with multiple segments",
			input:   oru,
			wantErr: false,
			validate: func(t *testing.T, msg hl7.Message) {
				if msg == nil {
					t.Fatal("message is nil")
				}

				segs := msg.AllSegments()
				if len(segs) != 4 {
					t.Errorf("expected 4 segments, got %d", len(segs))
				}

				obx, ok := msg.Segment("OBX")
				if !ok {
					t.Fatal("OBX segment not found")
				}
				if obx.Name() != "OBX" {
					t.Errorf("expected OBX, got %s", obx.Name())
				}
			},
		},
		{
			name:    "MSH only message",
			input:   mshOnly,
			wantErr: false,
			validate: func(t *testing.T, msg hl7.Message) {
				segs := msg.AllSegments()
				if len(segs) != 1 {
					t.Errorf("expected 1 segment, got %d", len(segs))
				}
			},
		},
		{
			name:    "message without final terminator",
			input:   noTerminator,
			wantErr: false,
			validate: func(t *testing.T, msg hl7.Message) {
				if msg == nil {
					t.Fatal("message is nil")
				}
				segs := msg.AllSegments()
				if len(segs) != 1 {
					t.Errorf("expected 1 segment, got %d", len(segs))
				}
			},
		},
		{
			name:        "empty input",
			input:       "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "whitespace only",
			input:       "   \r\n\t  ",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "no MSH segment",
			input:       "PID|1||12345^^^MRN||Doe^John\r",
			wantErr:     true,
			errContains: "MSH",
		},
		{
			name:        "MSH not first segment",
			input:       "PID|1||12345\rMSH|^~\\&|SENDING|FACILITY|||202301011200||ADT^A01|MSG|P|2.5\r",
			wantErr:     true,
			errContains: "MSH",
		},
		{
			name:    "empty segment - non-strict mode",
			input:   emptySegment,
			opts:    []ParserOption{WithStrictMode(false)},
			wantErr: false,
			validate: func(t *testing.T, msg hl7.Message) {
				// Empty segments should be skipped
				segs := msg.AllSegments()
				if len(segs) != 2 {
					t.Errorf("expected 2 segments (empty skipped), got %d", len(segs))
				}
			},
		},
		{
			name:        "empty segment - strict mode",
			input:       emptySegment,
			opts:        []ParserOption{WithStrictMode(true)},
			wantErr:     true,
			errContains: "empty segment",
		},
		{
			name:    "message with LF terminators",
			input:   "MSH|^~\\&|SENDING|FACILITY|||202301011200||ADT^A01|MSG|P|2.5\nPID|1||12345\n",
			opts:    []ParserOption{WithSegmentTerminator('\n')},
			wantErr: false,
			validate: func(t *testing.T, msg hl7.Message) {
				segs := msg.AllSegments()
				if len(segs) != 2 {
					t.Errorf("expected 2 segments, got %d", len(segs))
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := New(tt.opts...)
			msg, err := p.Parse([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errContains)) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, msg)
			}
		})
	}
}

func TestParser_ParseContext(t *testing.T) {
	t.Parallel()

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		p := New()
		_, err := p.ParseContext(ctx, []byte(simpleADT))

		if err == nil {
			t.Fatal("expected error for canceled context")
		}
		if !errors.Is(err, ErrContextCanceled) {
			t.Errorf("expected ErrContextCanceled, got %v", err)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Give it a moment to timeout
		time.Sleep(1 * time.Millisecond)

		p := New()
		_, err := p.ParseContext(ctx, []byte(simpleADT))

		if err == nil {
			t.Fatal("expected error for timed out context")
		}
	})

	t.Run("successful parse with context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		p := New()
		msg, err := p.ParseContext(ctx, []byte(simpleADT))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg == nil {
			t.Fatal("message is nil")
		}
	})
}

func TestParser_MaxSegments(t *testing.T) {
	t.Parallel()

	// Build a message with many segments
	var sb strings.Builder
	sb.WriteString("MSH|^~\\&|SENDING|FACILITY|||202301011200||ADT^A01|MSG|P|2.5\r")
	for i := 0; i < 10; i++ {
		sb.WriteString("PID|1||12345\r")
	}

	input := sb.String()

	tests := []struct {
		name        string
		maxSegments int
		wantErr     bool
	}{
		{
			name:        "within limit",
			maxSegments: 100,
			wantErr:     false,
		},
		{
			name:        "at limit",
			maxSegments: 11, // 1 MSH + 10 PID
			wantErr:     false,
		},
		{
			name:        "exceeds limit",
			maxSegments: 5,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := New(WithMaxSegments(tt.maxSegments))
			_, err := p.Parse([]byte(input))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !errors.Is(err, ErrTooManySegments) {
					t.Errorf("expected ErrTooManySegments, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParser_MaxFieldLength(t *testing.T) {
	t.Parallel()

	// Create a message with a long field
	longValue := strings.Repeat("X", 100)
	input := "MSH|^~\\&|SENDING|FACILITY|||202301011200||ADT^A01|MSG|P|2.5\rPID|1||" + longValue + "\r"

	tests := []struct {
		name           string
		maxFieldLength int
		wantErr        bool
	}{
		{
			name:           "within limit",
			maxFieldLength: 200,
			wantErr:        false,
		},
		{
			name:           "at limit",
			maxFieldLength: 100,
			wantErr:        false,
		},
		{
			name:           "exceeds limit",
			maxFieldLength: 50,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := New(WithMaxFieldLength(tt.maxFieldLength))
			_, err := p.Parse([]byte(input))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), "field") {
					t.Errorf("error should mention field, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParser_CustomDelimiters(t *testing.T) {
	t.Parallel()

	customDelims := &hl7.Delimiters{
		Field:        '$',
		Component:    '#',
		Repetition:   '~',
		Escape:       '\\',
		SubComponent: '@',
		Truncation:   '%',
	}

	p := New(WithCustomDelimiters(customDelims))
	msg, err := p.Parse([]byte(customDelimiters))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg == nil {
		t.Fatal("message is nil")
	}

	segs := msg.AllSegments()
	if len(segs) != 2 {
		t.Errorf("expected 2 segments, got %d", len(segs))
	}
}

func TestStripMLLP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "no MLLP framing",
			input:    []byte("MSH|^~\\&|"),
			expected: []byte("MSH|^~\\&|"),
		},
		{
			name:     "full MLLP framing",
			input:    []byte{0x0B, 'M', 'S', 'H', '|', 0x1C, 0x0D},
			expected: []byte{'M', 'S', 'H', '|'},
		},
		{
			name:     "start byte only",
			input:    []byte{0x0B, 'M', 'S', 'H', '|'},
			expected: []byte{'M', 'S', 'H', '|'},
		},
		{
			name:     "end bytes only (FS CR)",
			input:    []byte{'M', 'S', 'H', '|', 0x1C, 0x0D},
			expected: []byte{'M', 'S', 'H', '|'},
		},
		{
			name:     "FS without CR",
			input:    []byte{'M', 'S', 'H', '|', 0x1C},
			expected: []byte{'M', 'S', 'H', '|'},
		},
		{
			name:     "empty input",
			input:    []byte{},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := stripMLLP(tt.input)
			if string(result) != string(tt.expected) {
				t.Errorf("stripMLLP() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParser_MessageValues(t *testing.T) {
	t.Parallel()

	p := New()
	msg, err := p.Parse([]byte(simpleADT))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test message type - note: Type() returns the first component of MSH-9
	// The full message type with trigger is "ADT^A01" but Value() returns "ADT"
	msgType := msg.Type()
	if msgType != "ADT" && msgType != "ADT^A01" {
		t.Errorf("expected message type ADT or ADT^A01, got %s", msgType)
	}

	// Test control ID
	controlID := msg.ControlID()
	if controlID != "MSG001" {
		t.Errorf("expected control ID MSG001, got %s", controlID)
	}

	// Test version
	version := msg.Version()
	if version != "2.5" {
		t.Errorf("expected version 2.5, got %s", version)
	}

	// Test delimiters
	delims := msg.Delimiters()
	if delims == nil {
		t.Fatal("delimiters is nil")
		return
	}
	if delims.Field != '|' {
		t.Errorf("expected field delimiter |, got %c", delims.Field)
	}
	if delims.Component != '^' {
		t.Errorf("expected component delimiter ^, got %c", delims.Component)
	}
}

// Benchmark tests
func BenchmarkParser_Parse_SimpleADT(b *testing.B) {
	p := New()
	data := []byte(simpleADT)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParser_Parse_ORU(b *testing.B) {
	p := New()
	data := []byte(oru)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParser_Parse_MLLP(b *testing.B) {
	p := New()
	data := []byte(mllpFramedADT)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParser_Parse_LargeMessage(b *testing.B) {
	// Build a larger message
	var sb strings.Builder
	sb.WriteString("MSH|^~\\&|SENDING|FACILITY|||202301011200||ORU^R01|MSG|P|2.5\r")
	sb.WriteString("PID|1||12345^^^MRN||Doe^John^A||19800101|M\r")
	for i := 0; i < 100; i++ {
		sb.WriteString("OBX|1|NM|WBC||7.5|10*3/uL|4.5-11.0|N|||F\r")
	}
	data := []byte(sb.String())

	p := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
