package hl7

import (
	"errors"
	"testing"
)

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{
			name:     "error severity",
			severity: SeverityError,
			want:     "ERROR",
		},
		{
			name:     "warning severity",
			severity: SeverityWarning,
			want:     "WARNING",
		},
		{
			name:     "info severity",
			severity: SeverityInfo,
			want:     "INFO",
		},
		{
			name:     "unknown severity",
			severity: Severity(99),
			want:     "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ParseError
		want string
	}{
		{
			name: "with location and message",
			err: &ParseError{
				Message:  "unexpected character",
				Location: "PID-3-1",
			},
			want: "parse error at PID-3-1: unexpected character",
		},
		{
			name: "with line and column",
			err: &ParseError{
				Message: "invalid segment",
				Line:    5,
				Column:  10,
			},
			want: "parse error at line 5, column 10: invalid segment",
		},
		{
			name: "with cause",
			err: &ParseError{
				Message: "failed to parse",
				Cause:   ErrInvalidMessage,
			},
			want: "parse error: failed to parse: invalid message",
		},
		{
			name: "location takes precedence over line/column",
			err: &ParseError{
				Message:  "test",
				Location: "MSH-9",
				Line:     1,
				Column:   5,
			},
			want: "parse error at MSH-9: test",
		},
		{
			name: "minimal error",
			err:  &ParseError{},
			want: "parse error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ParseError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	cause := ErrInvalidMessage
	err := &ParseError{
		Message: "test",
		Cause:   cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("ParseError.Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Verify errors.Is works through Unwrap
	if !errors.Is(err, ErrInvalidMessage) {
		t.Error("errors.Is(err, ErrInvalidMessage) = false, want true")
	}
}

func TestLocationError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *LocationError
		want string
	}{
		{
			name: "with reason",
			err: &LocationError{
				Location: "PID-3-1-2-3-4",
				Reason:   "too many components",
			},
			want: `invalid location "PID-3-1-2-3-4": too many components`,
		},
		{
			name: "without reason",
			err: &LocationError{
				Location: "INVALID",
			},
			want: `invalid location "INVALID"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("LocationError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocationError_Unwrap(t *testing.T) {
	err := &LocationError{
		Location: "INVALID",
		Reason:   "test",
	}

	if unwrapped := err.Unwrap(); unwrapped != ErrInvalidLocation {
		t.Errorf("LocationError.Unwrap() = %v, want %v", unwrapped, ErrInvalidLocation)
	}

	// Verify errors.Is works
	if !errors.Is(err, ErrInvalidLocation) {
		t.Error("errors.Is(err, ErrInvalidLocation) = false, want true")
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ValidationError
		want string
	}{
		{
			name: "full error with all fields",
			err: &ValidationError{
				Location: "PID-5-1",
				Rule:     "required",
				Expected: "non-empty value",
				Actual:   "empty string",
				Severity: SeverityError,
			},
			want: `[ERROR] validation failed at PID-5-1: rule "required", expected non-empty value but got empty string`,
		},
		{
			name: "warning severity",
			err: &ValidationError{
				Location: "OBX-2",
				Rule:     "format",
				Severity: SeverityWarning,
			},
			want: `[WARNING] validation failed at OBX-2: rule "format"`,
		},
		{
			name: "expected only",
			err: &ValidationError{
				Location: "MSH-9",
				Expected: "ADT^A01",
				Severity: SeverityError,
			},
			want: "[ERROR] validation failed at MSH-9, expected ADT^A01",
		},
		{
			name: "actual only",
			err: &ValidationError{
				Location: "MSH-9",
				Actual:   "INVALID",
				Severity: SeverityError,
			},
			want: "[ERROR] validation failed at MSH-9, got INVALID",
		},
		{
			name: "minimal error",
			err: &ValidationError{
				Severity: SeverityInfo,
			},
			want: "[INFO] validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	err := &ValidationError{
		Location: "PID-3",
		Rule:     "test",
	}

	if unwrapped := err.Unwrap(); unwrapped != nil {
		t.Errorf("ValidationError.Unwrap() = %v, want nil", unwrapped)
	}
}

func TestSegmentError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *SegmentError
		want string
	}{
		{
			name: "with field and reason",
			err: &SegmentError{
				Segment: "PID",
				Field:   3,
				Reason:  "invalid format",
			},
			want: "segment PID field 3: invalid format",
		},
		{
			name: "segment only with reason",
			err: &SegmentError{
				Segment: "OBX",
				Reason:  "missing required fields",
			},
			want: "segment OBX: missing required fields",
		},
		{
			name: "with cause",
			err: &SegmentError{
				Segment: "MSH",
				Cause:   ErrInvalidMSH,
			},
			want: "segment MSH: invalid MSH segment",
		},
		{
			name: "with reason and cause",
			err: &SegmentError{
				Segment: "MSH",
				Field:   9,
				Reason:  "parse failed",
				Cause:   ErrFieldNotFound,
			},
			want: "segment MSH field 9: parse failed: field not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("SegmentError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSegmentError_Unwrap(t *testing.T) {
	cause := ErrSegmentNotFound
	err := &SegmentError{
		Segment: "ZZZ",
		Reason:  "custom segment",
		Cause:   cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("SegmentError.Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Verify errors.Is works
	if !errors.Is(err, ErrSegmentNotFound) {
		t.Error("errors.Is(err, ErrSegmentNotFound) = false, want true")
	}
}

func TestFieldError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *FieldError
		want string
	}{
		{
			name: "with reason",
			err: &FieldError{
				Sequence: 5,
				Reason:   "exceeds maximum length",
			},
			want: "field 5: exceeds maximum length",
		},
		{
			name: "with cause",
			err: &FieldError{
				Sequence: 3,
				Cause:    ErrComponentNotFound,
			},
			want: "field 3: component not found",
		},
		{
			name: "with reason and cause",
			err: &FieldError{
				Sequence: 7,
				Reason:   "component access failed",
				Cause:    ErrSubComponentNotFound,
			},
			want: "field 7: component access failed: subcomponent not found",
		},
		{
			name: "minimal",
			err: &FieldError{
				Sequence: 1,
			},
			want: "field 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("FieldError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldError_Unwrap(t *testing.T) {
	cause := ErrFieldNotFound
	err := &FieldError{
		Sequence: 99,
		Reason:   "out of range",
		Cause:    cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("FieldError.Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Verify errors.Is works
	if !errors.Is(err, ErrFieldNotFound) {
		t.Error("errors.Is(err, ErrFieldNotFound) = false, want true")
	}
}

func TestErrors_As(t *testing.T) {
	t.Run("ParseError", func(t *testing.T) {
		original := &ParseError{
			Message:  "test",
			Location: "PID-3",
			Line:     1,
			Column:   5,
		}
		wrapped := errors.Join(errors.New("outer"), original)

		var target *ParseError
		if !errors.As(wrapped, &target) {
			t.Error("errors.As() returned false, want true")
		}
		if target.Location != "PID-3" {
			t.Errorf("target.Location = %v, want PID-3", target.Location)
		}
	})

	t.Run("LocationError", func(t *testing.T) {
		original := &LocationError{
			Location: "INVALID",
			Reason:   "bad format",
		}
		wrapped := errors.Join(errors.New("outer"), original)

		var target *LocationError
		if !errors.As(wrapped, &target) {
			t.Error("errors.As() returned false, want true")
		}
		if target.Location != "INVALID" {
			t.Errorf("target.Location = %v, want INVALID", target.Location)
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		original := &ValidationError{
			Location: "OBX-5",
			Rule:     "format",
			Severity: SeverityWarning,
		}
		wrapped := errors.Join(errors.New("outer"), original)

		var target *ValidationError
		if !errors.As(wrapped, &target) {
			t.Error("errors.As() returned false, want true")
		}
		if target.Rule != "format" {
			t.Errorf("target.Rule = %v, want format", target.Rule)
		}
		if target.Severity != SeverityWarning {
			t.Errorf("target.Severity = %v, want SeverityWarning", target.Severity)
		}
	})

	t.Run("SegmentError", func(t *testing.T) {
		original := &SegmentError{
			Segment: "PID",
			Field:   3,
			Reason:  "test",
		}
		wrapped := errors.Join(errors.New("outer"), original)

		var target *SegmentError
		if !errors.As(wrapped, &target) {
			t.Error("errors.As() returned false, want true")
		}
		if target.Segment != "PID" {
			t.Errorf("target.Segment = %v, want PID", target.Segment)
		}
	})

	t.Run("FieldError", func(t *testing.T) {
		original := &FieldError{
			Sequence: 5,
			Reason:   "test",
		}
		wrapped := errors.Join(errors.New("outer"), original)

		var target *FieldError
		if !errors.As(wrapped, &target) {
			t.Error("errors.As() returned false, want true")
		}
		if target.Sequence != 5 {
			t.Errorf("target.Sequence = %v, want 5", target.Sequence)
		}
	})
}

func TestErrors_Is_SentinelErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		wantIs bool
	}{
		{
			name:   "ErrInvalidLocation direct",
			err:    ErrInvalidLocation,
			target: ErrInvalidLocation,
			wantIs: true,
		},
		{
			name:   "ErrSegmentNotFound direct",
			err:    ErrSegmentNotFound,
			target: ErrSegmentNotFound,
			wantIs: true,
		},
		{
			name:   "ErrFieldNotFound direct",
			err:    ErrFieldNotFound,
			target: ErrFieldNotFound,
			wantIs: true,
		},
		{
			name:   "ErrComponentNotFound direct",
			err:    ErrComponentNotFound,
			target: ErrComponentNotFound,
			wantIs: true,
		},
		{
			name:   "ErrSubComponentNotFound direct",
			err:    ErrSubComponentNotFound,
			target: ErrSubComponentNotFound,
			wantIs: true,
		},
		{
			name:   "ErrInvalidMessage direct",
			err:    ErrInvalidMessage,
			target: ErrInvalidMessage,
			wantIs: true,
		},
		{
			name:   "ErrEmptyMessage direct",
			err:    ErrEmptyMessage,
			target: ErrEmptyMessage,
			wantIs: true,
		},
		{
			name:   "ErrMissingMSH direct",
			err:    ErrMissingMSH,
			target: ErrMissingMSH,
			wantIs: true,
		},
		{
			name:   "ErrInvalidMSH direct",
			err:    ErrInvalidMSH,
			target: ErrInvalidMSH,
			wantIs: true,
		},
		{
			name:   "different errors",
			err:    ErrInvalidLocation,
			target: ErrSegmentNotFound,
			wantIs: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.wantIs {
				t.Errorf("errors.Is() = %v, want %v", got, tt.wantIs)
			}
		})
	}
}

func TestErrorChaining(t *testing.T) {
	// Build a chain: SegmentError -> FieldError -> ErrComponentNotFound
	fieldErr := &FieldError{
		Sequence: 3,
		Reason:   "component access failed",
		Cause:    ErrComponentNotFound,
	}

	segmentErr := &SegmentError{
		Segment: "PID",
		Field:   5,
		Reason:  "field processing failed",
		Cause:   fieldErr,
	}

	parseErr := &ParseError{
		Message:  "message parsing failed",
		Location: "PID-5-3",
		Cause:    segmentErr,
	}

	// Verify the full chain is accessible
	if !errors.Is(parseErr, ErrComponentNotFound) {
		t.Error("errors.Is(parseErr, ErrComponentNotFound) = false, want true")
	}

	// Verify we can extract intermediate errors
	var extractedFieldErr *FieldError
	if !errors.As(parseErr, &extractedFieldErr) {
		t.Error("errors.As(parseErr, &FieldError) = false, want true")
	}
	if extractedFieldErr.Sequence != 3 {
		t.Errorf("extractedFieldErr.Sequence = %v, want 3", extractedFieldErr.Sequence)
	}

	var extractedSegmentErr *SegmentError
	if !errors.As(parseErr, &extractedSegmentErr) {
		t.Error("errors.As(parseErr, &SegmentError) = false, want true")
	}
	if extractedSegmentErr.Segment != "PID" {
		t.Errorf("extractedSegmentErr.Segment = %v, want PID", extractedSegmentErr.Segment)
	}
}
