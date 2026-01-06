package segments

import (
	"testing"

	"github.com/dshills/golevel7/hl7"
)

func TestParsePV1(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *PV1
		wantErr bool
	}{
		{
			name:  "complete PV1 segment",
			input: "PV1|1|I|ICU^Room1^BedA^Hospital||||1234^Smith^John^Dr^^^MD|5678^Jones^Mary^Dr|||SUR||||||||VN12345|||||||||||||||||||||||20230101080000|20230115120000",
			want: &PV1{
				SetID:                   "1",
				PatientClass:            "I",
				AssignedPatientLocation: "ICU^Room1^BedA^Hospital",
				AttendingDoctor:         "1234^Smith^John^Dr^^^MD",
				ReferringDoctor:         "5678^Jones^Mary^Dr",
				HospitalService:         "SUR",
				VisitNumber:             "VN12345",
				AdmitDateTime:           "20230101080000",
				DischargeDateTime:       "20230115120000",
			},
			wantErr: false,
		},
		{
			name:  "minimal PV1 segment",
			input: "PV1|1|O|ER^Room5",
			want: &PV1{
				SetID:                   "1",
				PatientClass:            "O",
				AssignedPatientLocation: "ER^Room5",
			},
			wantErr: false,
		},
		{
			name:  "emergency PV1",
			input: "PV1|1|E|ED^Trauma1^Bed1|||||||EMR|||2|||||9999^Emergency^Dr|||||||||||||||||||||20230615143000",
			want: &PV1{
				SetID:                   "1",
				PatientClass:            "E",
				AssignedPatientLocation: "ED^Trauma1^Bed1",
				HospitalService:         "EMR",
				ReadmissionIndicator:    "2",
				AdmittingDoctor:         "9999^Emergency^Dr",
				AdmitDateTime:           "20230615143000",
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

			got, err := ParsePV1(seg)

			if tt.wantErr {
				if err == nil {
					t.Error("ParsePV1() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParsePV1() unexpected error: %v", err)
			}

			// Check key fields
			if got.SetID != tt.want.SetID {
				t.Errorf("SetID = %q, want %q", got.SetID, tt.want.SetID)
			}
			if got.PatientClass != tt.want.PatientClass {
				t.Errorf("PatientClass = %q, want %q", got.PatientClass, tt.want.PatientClass)
			}
			if got.AssignedPatientLocation != tt.want.AssignedPatientLocation {
				t.Errorf("AssignedPatientLocation = %q, want %q", got.AssignedPatientLocation, tt.want.AssignedPatientLocation)
			}
		})
	}
}

func TestParsePV1_WrongSegment(t *testing.T) {
	input := "PID|1||12345|||Doe^John||19800101|M"
	seg, err := hl7.ParseSegment([]rune(input), hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("failed to parse segment: %v", err)
	}

	_, err = ParsePV1(seg)
	if err == nil {
		t.Error("ParsePV1() expected error for non-PV1 segment, got nil")
	}
}

func TestPV1_ToSegment(t *testing.T) {
	tests := []struct {
		name    string
		pv1     *PV1
		wantErr bool
	}{
		{
			name: "inpatient PV1",
			pv1: &PV1{
				SetID:                   "1",
				PatientClass:            "I",
				AssignedPatientLocation: "MED^101^A",
				AttendingDoctor:         "1234^Attending^Dr",
				HospitalService:         "MED",
				VisitNumber:             "VN001",
				AdmitDateTime:           "20230601080000",
			},
			wantErr: false,
		},
		{
			name: "outpatient PV1",
			pv1: &PV1{
				SetID:                   "1",
				PatientClass:            "O",
				AssignedPatientLocation: "CLINIC^Room2",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := tt.pv1.ToSegment(hl7.DefaultDelimiters())

			if tt.wantErr {
				if err == nil {
					t.Error("ToSegment() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ToSegment() unexpected error: %v", err)
			}

			if seg.Name() != "PV1" {
				t.Errorf("segment name = %q, want PV1", seg.Name())
			}

			// Parse back and verify
			parsed, err := ParsePV1(seg)
			if err != nil {
				t.Fatalf("failed to parse created segment: %v", err)
			}

			if parsed.PatientClass != tt.pv1.PatientClass {
				t.Errorf("PatientClass = %q, want %q", parsed.PatientClass, tt.pv1.PatientClass)
			}
			if parsed.AssignedPatientLocation != tt.pv1.AssignedPatientLocation {
				t.Errorf("AssignedPatientLocation = %q, want %q", parsed.AssignedPatientLocation, tt.pv1.AssignedPatientLocation)
			}
		})
	}
}

func TestPV1_RoundTrip(t *testing.T) {
	original := &PV1{
		SetID:                   "1",
		PatientClass:            "I",
		AssignedPatientLocation: "ICU^Room1^Bed1^Hospital",
		AdmissionType:           "E",
		AttendingDoctor:         "12345^Smith^John^Dr^^^MD",
		ReferringDoctor:         "67890^Jones^Mary^Dr",
		HospitalService:         "ICU",
		VisitNumber:             "V12345",
		AdmitDateTime:           "20230101080000",
		DischargeDateTime:       "20230110120000",
	}

	// Convert to segment
	seg, err := original.ToSegment(hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("ToSegment() error: %v", err)
	}

	// Parse back
	parsed, err := ParsePV1(seg)
	if err != nil {
		t.Fatalf("ParsePV1() error: %v", err)
	}

	// Verify fields
	if parsed.SetID != original.SetID {
		t.Errorf("SetID = %q, want %q", parsed.SetID, original.SetID)
	}
	if parsed.PatientClass != original.PatientClass {
		t.Errorf("PatientClass = %q, want %q", parsed.PatientClass, original.PatientClass)
	}
	if parsed.AssignedPatientLocation != original.AssignedPatientLocation {
		t.Errorf("AssignedPatientLocation = %q, want %q", parsed.AssignedPatientLocation, original.AssignedPatientLocation)
	}
	if parsed.AttendingDoctor != original.AttendingDoctor {
		t.Errorf("AttendingDoctor = %q, want %q", parsed.AttendingDoctor, original.AttendingDoctor)
	}
	if parsed.HospitalService != original.HospitalService {
		t.Errorf("HospitalService = %q, want %q", parsed.HospitalService, original.HospitalService)
	}
	if parsed.VisitNumber != original.VisitNumber {
		t.Errorf("VisitNumber = %q, want %q", parsed.VisitNumber, original.VisitNumber)
	}
}
