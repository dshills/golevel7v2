package segments

import (
	"testing"

	"github.com/dshills/golevel7/hl7"
)

func TestParseOBX(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *OBX
		wantErr bool
	}{
		{
			name:  "numeric OBX segment",
			input: "OBX|1|NM|2345-7^Glucose^LN||120|mg/dL|70-100|H|||F|||20230615090000",
			want: &OBX{
				SetID:                   "1",
				ValueType:               "NM",
				ObservationIdentifier:   "2345-7^Glucose^LN",
				ObservationValue:        "120",
				Units:                   "mg/dL",
				ReferencesRange:         "70-100",
				AbnormalFlags:           "H",
				ObservationResultStatus: "F",
				DateTimeOfObservation:   "20230615090000",
			},
			wantErr: false,
		},
		{
			name:  "text OBX segment",
			input: "OBX|1|TX|1234-5^Clinical Note||Patient presents with mild symptoms.||||||F",
			want: &OBX{
				SetID:                   "1",
				ValueType:               "TX",
				ObservationIdentifier:   "1234-5^Clinical Note",
				ObservationValue:        "Patient presents with mild symptoms.",
				ObservationResultStatus: "F",
			},
			wantErr: false,
		},
		{
			name:  "coded element OBX",
			input: "OBX|2|CE|9999-1^Blood Type^LN||A+^A Positive^HL70005||||N||F",
			want: &OBX{
				SetID:                   "2",
				ValueType:               "CE",
				ObservationIdentifier:   "9999-1^Blood Type^LN",
				ObservationValue:        "A+^A Positive^HL70005",
				NatureOfAbnormalTest:    "N",
				ObservationResultStatus: "F",
			},
			wantErr: false,
		},
		{
			name:  "OBX with sub-ID",
			input: "OBX|1|NM|12345^Test^LN|1.1|50|units||N|||F",
			want: &OBX{
				SetID:                   "1",
				ValueType:               "NM",
				ObservationIdentifier:   "12345^Test^LN",
				ObservationSubID:        "1.1",
				ObservationValue:        "50",
				Units:                   "units",
				AbnormalFlags:           "N",
				ObservationResultStatus: "F",
			},
			wantErr: false,
		},
		{
			name:    "nil segment",
			input:   "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seg hl7.Segment
			var err error

			if tt.input != "" {
				seg, err = hl7.ParseSegment([]rune(tt.input), hl7.DefaultDelimiters())
				if err != nil {
					t.Fatalf("failed to parse segment: %v", err)
				}
			}

			got, err := ParseOBX(seg)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseOBX() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseOBX() unexpected error: %v", err)
			}

			// Check key fields
			if got.SetID != tt.want.SetID {
				t.Errorf("SetID = %q, want %q", got.SetID, tt.want.SetID)
			}
			if got.ValueType != tt.want.ValueType {
				t.Errorf("ValueType = %q, want %q", got.ValueType, tt.want.ValueType)
			}
			if got.ObservationIdentifier != tt.want.ObservationIdentifier {
				t.Errorf("ObservationIdentifier = %q, want %q", got.ObservationIdentifier, tt.want.ObservationIdentifier)
			}
			if got.ObservationValue != tt.want.ObservationValue {
				t.Errorf("ObservationValue = %q, want %q", got.ObservationValue, tt.want.ObservationValue)
			}
			if got.ObservationResultStatus != tt.want.ObservationResultStatus {
				t.Errorf("ObservationResultStatus = %q, want %q", got.ObservationResultStatus, tt.want.ObservationResultStatus)
			}
		})
	}
}

func TestParseOBX_WrongSegment(t *testing.T) {
	input := "OBR|1|P001|F001|CBC^Complete Blood Count"
	seg, err := hl7.ParseSegment([]rune(input), hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("failed to parse segment: %v", err)
	}

	_, err = ParseOBX(seg)
	if err == nil {
		t.Error("ParseOBX() expected error for non-OBX segment, got nil")
	}
}

func TestOBX_ToSegment(t *testing.T) {
	tests := []struct {
		name    string
		obx     *OBX
		wantErr bool
	}{
		{
			name: "numeric result",
			obx: &OBX{
				SetID:                   "1",
				ValueType:               "NM",
				ObservationIdentifier:   "14749-6^Glucose",
				ObservationValue:        "95",
				Units:                   "mg/dL",
				ReferencesRange:         "70-100",
				AbnormalFlags:           "N",
				ObservationResultStatus: "F",
			},
			wantErr: false,
		},
		{
			name: "text result",
			obx: &OBX{
				SetID:                   "1",
				ValueType:               "TX",
				ObservationIdentifier:   "Note^Clinical Note",
				ObservationValue:        "Normal findings.",
				ObservationResultStatus: "F",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := tt.obx.ToSegment(hl7.DefaultDelimiters())

			if tt.wantErr {
				if err == nil {
					t.Error("ToSegment() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ToSegment() unexpected error: %v", err)
			}

			if seg.Name() != "OBX" {
				t.Errorf("segment name = %q, want OBX", seg.Name())
			}

			// Parse back and verify
			parsed, err := ParseOBX(seg)
			if err != nil {
				t.Fatalf("failed to parse created segment: %v", err)
			}

			if parsed.ValueType != tt.obx.ValueType {
				t.Errorf("ValueType = %q, want %q", parsed.ValueType, tt.obx.ValueType)
			}
			if parsed.ObservationValue != tt.obx.ObservationValue {
				t.Errorf("ObservationValue = %q, want %q", parsed.ObservationValue, tt.obx.ObservationValue)
			}
			if parsed.ObservationResultStatus != tt.obx.ObservationResultStatus {
				t.Errorf("ObservationResultStatus = %q, want %q", parsed.ObservationResultStatus, tt.obx.ObservationResultStatus)
			}
		})
	}
}

func TestOBX_RoundTrip(t *testing.T) {
	original := &OBX{
		SetID:                   "1",
		ValueType:               "NM",
		ObservationIdentifier:   "2345-7^Glucose^LN",
		ObservationSubID:        "1",
		ObservationValue:        "105",
		Units:                   "mg/dL",
		ReferencesRange:         "70-100",
		AbnormalFlags:           "H",
		ObservationResultStatus: "F",
		DateTimeOfObservation:   "20230615143000",
		ProducersID:             "LAB001^Main Laboratory",
	}

	// Convert to segment
	seg, err := original.ToSegment(hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("ToSegment() error: %v", err)
	}

	// Parse back
	parsed, err := ParseOBX(seg)
	if err != nil {
		t.Fatalf("ParseOBX() error: %v", err)
	}

	// Verify fields
	if parsed.SetID != original.SetID {
		t.Errorf("SetID = %q, want %q", parsed.SetID, original.SetID)
	}
	if parsed.ValueType != original.ValueType {
		t.Errorf("ValueType = %q, want %q", parsed.ValueType, original.ValueType)
	}
	if parsed.ObservationIdentifier != original.ObservationIdentifier {
		t.Errorf("ObservationIdentifier = %q, want %q", parsed.ObservationIdentifier, original.ObservationIdentifier)
	}
	if parsed.ObservationValue != original.ObservationValue {
		t.Errorf("ObservationValue = %q, want %q", parsed.ObservationValue, original.ObservationValue)
	}
	if parsed.Units != original.Units {
		t.Errorf("Units = %q, want %q", parsed.Units, original.Units)
	}
	if parsed.ReferencesRange != original.ReferencesRange {
		t.Errorf("ReferencesRange = %q, want %q", parsed.ReferencesRange, original.ReferencesRange)
	}
	if parsed.AbnormalFlags != original.AbnormalFlags {
		t.Errorf("AbnormalFlags = %q, want %q", parsed.AbnormalFlags, original.AbnormalFlags)
	}
	if parsed.ObservationResultStatus != original.ObservationResultStatus {
		t.Errorf("ObservationResultStatus = %q, want %q", parsed.ObservationResultStatus, original.ObservationResultStatus)
	}
}

func TestOBX_VariousValueTypes(t *testing.T) {
	valueTypes := []struct {
		name      string
		valueType string
		value     string
	}{
		{"Numeric", "NM", "123.45"},
		{"String", "ST", "Simple String"},
		{"Text", "TX", "Long text value that could span multiple lines"},
		{"Coded Element", "CE", "CODE^Description^CodingSystem"},
		{"Coded with Exceptions", "CWE", "CODE^Description^System^^Alt^AltSystem"},
		{"Date/Time", "TS", "20230615143000"},
		{"Structured Numeric", "SN", ">100"},
	}

	for _, vt := range valueTypes {
		t.Run(vt.name, func(t *testing.T) {
			original := &OBX{
				SetID:                   "1",
				ValueType:               vt.valueType,
				ObservationIdentifier:   "TEST-1^Test",
				ObservationValue:        vt.value,
				ObservationResultStatus: "F",
			}

			seg, err := original.ToSegment(hl7.DefaultDelimiters())
			if err != nil {
				t.Fatalf("ToSegment() error: %v", err)
			}

			parsed, err := ParseOBX(seg)
			if err != nil {
				t.Fatalf("ParseOBX() error: %v", err)
			}

			if parsed.ValueType != vt.valueType {
				t.Errorf("ValueType = %q, want %q", parsed.ValueType, vt.valueType)
			}
			if parsed.ObservationValue != vt.value {
				t.Errorf("ObservationValue = %q, want %q", parsed.ObservationValue, vt.value)
			}
		})
	}
}
