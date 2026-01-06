package segments

import (
	"testing"

	"github.com/dshills/golevel7/hl7"
)

func TestParseOBR(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *OBR
		wantErr bool
	}{
		{
			name:  "complete OBR segment",
			input: "OBR|1|ORD001^Placer|FIL001^Filler|12345-6^Glucose^LN|||20230615080000|||||||||||1234^Ordering^Dr|||||||F",
			want: &OBR{
				SetID:                      "1",
				PlacerOrderNumber:          "ORD001^Placer",
				FillerOrderNumber:          "FIL001^Filler",
				UniversalServiceIdentifier: "12345-6^Glucose^LN",
				ObservationDateTime:        "20230615080000",
				OrderingProvider:           "1234^Ordering^Dr",
				ResultStatus:               "F",
			},
			wantErr: false,
		},
		{
			name:  "minimal OBR segment",
			input: "OBR|1||F123|80048^Basic Metabolic Panel^CPT",
			want: &OBR{
				SetID:                      "1",
				FillerOrderNumber:          "F123",
				UniversalServiceIdentifier: "80048^Basic Metabolic Panel^CPT",
			},
			wantErr: false,
		},
		{
			name:  "OBR with specimen info",
			input: "OBR|1|P001|F001|CBC^Complete Blood Count||20230601|20230601090000|20230601091500|||Nurse^Nancy|||20230601092000|Blood^Venous",
			want: &OBR{
				SetID:                      "1",
				PlacerOrderNumber:          "P001",
				FillerOrderNumber:          "F001",
				UniversalServiceIdentifier: "CBC^Complete Blood Count",
				RequestedDateTime:          "20230601",
				ObservationDateTime:        "20230601090000",
				ObservationEndDateTime:     "20230601091500",
				CollectorIdentifier:        "Nurse^Nancy",
				SpecimenReceivedDateTime:   "20230601092000",
				SpecimenSource:             "Blood^Venous",
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

			got, err := ParseOBR(seg)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseOBR() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseOBR() unexpected error: %v", err)
			}

			// Check key fields
			if got.SetID != tt.want.SetID {
				t.Errorf("SetID = %q, want %q", got.SetID, tt.want.SetID)
			}
			if got.PlacerOrderNumber != tt.want.PlacerOrderNumber {
				t.Errorf("PlacerOrderNumber = %q, want %q", got.PlacerOrderNumber, tt.want.PlacerOrderNumber)
			}
			if got.FillerOrderNumber != tt.want.FillerOrderNumber {
				t.Errorf("FillerOrderNumber = %q, want %q", got.FillerOrderNumber, tt.want.FillerOrderNumber)
			}
			if got.UniversalServiceIdentifier != tt.want.UniversalServiceIdentifier {
				t.Errorf("UniversalServiceIdentifier = %q, want %q", got.UniversalServiceIdentifier, tt.want.UniversalServiceIdentifier)
			}
		})
	}
}

func TestParseOBR_WrongSegment(t *testing.T) {
	input := "OBX|1|NM|12345^Glucose||100|mg/dL|70-100|N|||F"
	seg, err := hl7.ParseSegment([]rune(input), hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("failed to parse segment: %v", err)
	}

	_, err = ParseOBR(seg)
	if err == nil {
		t.Error("ParseOBR() expected error for non-OBR segment, got nil")
	}
}

func TestOBR_ToSegment(t *testing.T) {
	tests := []struct {
		name    string
		obr     *OBR
		wantErr bool
	}{
		{
			name: "lab order OBR",
			obr: &OBR{
				SetID:                      "1",
				PlacerOrderNumber:          "P12345",
				FillerOrderNumber:          "F67890",
				UniversalServiceIdentifier: "14749-6^Glucose",
				ObservationDateTime:        "20230615090000",
				OrderingProvider:           "1234^Smith^Dr",
				ResultStatus:               "F",
			},
			wantErr: false,
		},
		{
			name: "minimal OBR",
			obr: &OBR{
				SetID:                      "1",
				UniversalServiceIdentifier: "CBC^Complete Blood Count",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := tt.obr.ToSegment(hl7.DefaultDelimiters())

			if tt.wantErr {
				if err == nil {
					t.Error("ToSegment() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ToSegment() unexpected error: %v", err)
			}

			if seg.Name() != "OBR" {
				t.Errorf("segment name = %q, want OBR", seg.Name())
			}

			// Parse back and verify
			parsed, err := ParseOBR(seg)
			if err != nil {
				t.Fatalf("failed to parse created segment: %v", err)
			}

			if parsed.UniversalServiceIdentifier != tt.obr.UniversalServiceIdentifier {
				t.Errorf("UniversalServiceIdentifier = %q, want %q", parsed.UniversalServiceIdentifier, tt.obr.UniversalServiceIdentifier)
			}
		})
	}
}

func TestOBR_RoundTrip(t *testing.T) {
	original := &OBR{
		SetID:                      "1",
		PlacerOrderNumber:          "PLACER001^Lab",
		FillerOrderNumber:          "FILLER001^LabSys",
		UniversalServiceIdentifier: "24323-8^Comprehensive Metabolic Panel^LN",
		ObservationDateTime:        "20230615140000",
		SpecimenReceivedDateTime:   "20230615130000",
		OrderingProvider:           "99999^Ordering^Doctor^MD",
		DiagnosticServSectID:       "LAB",
		ResultStatus:               "F",
	}

	// Convert to segment
	seg, err := original.ToSegment(hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("ToSegment() error: %v", err)
	}

	// Parse back
	parsed, err := ParseOBR(seg)
	if err != nil {
		t.Fatalf("ParseOBR() error: %v", err)
	}

	// Verify fields
	if parsed.SetID != original.SetID {
		t.Errorf("SetID = %q, want %q", parsed.SetID, original.SetID)
	}
	if parsed.PlacerOrderNumber != original.PlacerOrderNumber {
		t.Errorf("PlacerOrderNumber = %q, want %q", parsed.PlacerOrderNumber, original.PlacerOrderNumber)
	}
	if parsed.FillerOrderNumber != original.FillerOrderNumber {
		t.Errorf("FillerOrderNumber = %q, want %q", parsed.FillerOrderNumber, original.FillerOrderNumber)
	}
	if parsed.UniversalServiceIdentifier != original.UniversalServiceIdentifier {
		t.Errorf("UniversalServiceIdentifier = %q, want %q", parsed.UniversalServiceIdentifier, original.UniversalServiceIdentifier)
	}
	if parsed.ObservationDateTime != original.ObservationDateTime {
		t.Errorf("ObservationDateTime = %q, want %q", parsed.ObservationDateTime, original.ObservationDateTime)
	}
	if parsed.ResultStatus != original.ResultStatus {
		t.Errorf("ResultStatus = %q, want %q", parsed.ResultStatus, original.ResultStatus)
	}
}
