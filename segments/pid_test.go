package segments

import (
	"testing"

	"github.com/dshills/golevel7/hl7"
)

func TestParsePID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *PID
		wantErr bool
	}{
		{
			name:  "complete PID segment",
			input: "PID|1||12345^^^Hospital^MR~98765^^^SSN^SS||Doe^John^Q^Jr^Dr||19800115|M|||123 Main St^^Anytown^ST^12345||555-123-4567||EN|M|Christian|ACCT001|123-45-6789",
			want: &PID{
				SetID:                "1",
				PatientIDList:        "12345^^^Hospital^MR~98765^^^SSN^SS",
				PatientName:          "Doe^John^Q^Jr^Dr",
				DateOfBirth:          "19800115",
				Sex:                  "M",
				PatientAddress:       "123 Main St^^Anytown^ST^12345",
				PhoneNumberHome:      "555-123-4567",
				PrimaryLanguage:      "EN",
				MaritalStatus:        "M",
				Religion:             "Christian",
				PatientAccountNumber: "ACCT001",
				SSNNumber:            "123-45-6789",
			},
			wantErr: false,
		},
		{
			name:  "minimal PID segment",
			input: "PID|1||12345^^^Hospital^MR||Smith^Jane||19900501|F",
			want: &PID{
				SetID:         "1",
				PatientIDList: "12345^^^Hospital^MR",
				PatientName:   "Smith^Jane",
				DateOfBirth:   "19900501",
				Sex:           "F",
			},
			wantErr: false,
		},
		{
			name:  "PID with mother's maiden name",
			input: "PID|1||67890|||Johnson^Mary|19750320|F||||||||||||||19500101",
			want: &PID{
				SetID:            "1",
				PatientIDList:    "67890",
				MotherMaidenName: "Johnson^Mary",
				DateOfBirth:      "19750320",
				Sex:              "F",
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

			got, err := ParsePID(seg)

			if tt.wantErr {
				if err == nil {
					t.Error("ParsePID() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParsePID() unexpected error: %v", err)
			}

			// Check key fields
			if got.SetID != tt.want.SetID {
				t.Errorf("SetID = %q, want %q", got.SetID, tt.want.SetID)
			}
			if got.PatientIDList != tt.want.PatientIDList {
				t.Errorf("PatientIDList = %q, want %q", got.PatientIDList, tt.want.PatientIDList)
			}
			if got.PatientName != tt.want.PatientName {
				t.Errorf("PatientName = %q, want %q", got.PatientName, tt.want.PatientName)
			}
			if got.DateOfBirth != tt.want.DateOfBirth {
				t.Errorf("DateOfBirth = %q, want %q", got.DateOfBirth, tt.want.DateOfBirth)
			}
			if got.Sex != tt.want.Sex {
				t.Errorf("Sex = %q, want %q", got.Sex, tt.want.Sex)
			}
		})
	}
}

func TestParsePID_WrongSegment(t *testing.T) {
	input := "MSH|^~\\&|App|Fac|||20230101||ADT^A01|MSG001|P|2.5.1"
	seg, err := hl7.ParseSegment([]rune(input), hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("failed to parse segment: %v", err)
	}

	_, err = ParsePID(seg)
	if err == nil {
		t.Error("ParsePID() expected error for non-PID segment, got nil")
	}
}

func TestPID_ToSegment(t *testing.T) {
	tests := []struct {
		name    string
		pid     *PID
		wantErr bool
	}{
		{
			name: "complete PID",
			pid: &PID{
				SetID:                "1",
				PatientIDList:        "12345^^^Hospital^MR",
				PatientName:          "Doe^John^Middle",
				DateOfBirth:          "19800115",
				Sex:                  "M",
				PatientAddress:       "123 Main St^^City^ST^12345",
				PhoneNumberHome:      "555-123-4567",
				PatientAccountNumber: "ACCT001",
			},
			wantErr: false,
		},
		{
			name: "minimal PID",
			pid: &PID{
				SetID:       "1",
				PatientName: "Smith^Jane",
				Sex:         "F",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := tt.pid.ToSegment(hl7.DefaultDelimiters())

			if tt.wantErr {
				if err == nil {
					t.Error("ToSegment() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ToSegment() unexpected error: %v", err)
			}

			if seg.Name() != "PID" {
				t.Errorf("segment name = %q, want PID", seg.Name())
			}

			// Parse back and verify
			parsed, err := ParsePID(seg)
			if err != nil {
				t.Fatalf("failed to parse created segment: %v", err)
			}

			if parsed.PatientName != tt.pid.PatientName {
				t.Errorf("PatientName = %q, want %q", parsed.PatientName, tt.pid.PatientName)
			}
			if parsed.Sex != tt.pid.Sex {
				t.Errorf("Sex = %q, want %q", parsed.Sex, tt.pid.Sex)
			}
		})
	}
}

func TestPID_RoundTrip(t *testing.T) {
	original := &PID{
		SetID:                "1",
		PatientIDList:        "MRN12345^^^Hospital^MR",
		PatientName:          "TestPatient^First^Middle^Jr",
		DateOfBirth:          "19850620",
		Sex:                  "M",
		Race:                 "W",
		PatientAddress:       "456 Oak Ave^^Springfield^IL^62701",
		PhoneNumberHome:      "217-555-1234",
		PhoneNumberBusiness:  "217-555-5678",
		PrimaryLanguage:      "EN",
		MaritalStatus:        "S",
		PatientAccountNumber: "ACCT98765",
	}

	// Convert to segment
	seg, err := original.ToSegment(hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("ToSegment() error: %v", err)
	}

	// Parse back
	parsed, err := ParsePID(seg)
	if err != nil {
		t.Fatalf("ParsePID() error: %v", err)
	}

	// Verify fields
	if parsed.SetID != original.SetID {
		t.Errorf("SetID = %q, want %q", parsed.SetID, original.SetID)
	}
	if parsed.PatientIDList != original.PatientIDList {
		t.Errorf("PatientIDList = %q, want %q", parsed.PatientIDList, original.PatientIDList)
	}
	if parsed.PatientName != original.PatientName {
		t.Errorf("PatientName = %q, want %q", parsed.PatientName, original.PatientName)
	}
	if parsed.DateOfBirth != original.DateOfBirth {
		t.Errorf("DateOfBirth = %q, want %q", parsed.DateOfBirth, original.DateOfBirth)
	}
	if parsed.Sex != original.Sex {
		t.Errorf("Sex = %q, want %q", parsed.Sex, original.Sex)
	}
	if parsed.PatientAddress != original.PatientAddress {
		t.Errorf("PatientAddress = %q, want %q", parsed.PatientAddress, original.PatientAddress)
	}
}
