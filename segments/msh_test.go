package segments

import (
	"testing"

	"github.com/dshills/golevel7/hl7"
)

func TestParseMSH(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     *MSH
		wantErr  bool
		errMatch string
	}{
		{
			name:  "complete MSH segment",
			input: "MSH|^~\\&|SendApp|SendFac|RecvApp|RecvFac|20230101120000||ADT^A01|MSG00001|P|2.5.1",
			want: &MSH{
				FieldSeparator:       "|",
				EncodingCharacters:   "^~\\&",
				SendingApplication:   "SendApp",
				SendingFacility:      "SendFac",
				ReceivingApplication: "RecvApp",
				ReceivingFacility:    "RecvFac",
				DateTime:             "20230101120000",
				Security:             "",
				MessageType:          "ADT^A01",
				MessageControlID:     "MSG00001",
				ProcessingID:         "P",
				VersionID:            "2.5.1",
			},
			wantErr: false,
		},
		{
			name:  "MSH with security field",
			input: "MSH|^~\\&|SendApp|SendFac|RecvApp|RecvFac|20230101120000|SEC123|ORU^R01|MSG00002|P|2.5",
			want: &MSH{
				FieldSeparator:       "|",
				EncodingCharacters:   "^~\\&",
				SendingApplication:   "SendApp",
				SendingFacility:      "SendFac",
				ReceivingApplication: "RecvApp",
				ReceivingFacility:    "RecvFac",
				DateTime:             "20230101120000",
				Security:             "SEC123",
				MessageType:          "ORU^R01",
				MessageControlID:     "MSG00002",
				ProcessingID:         "P",
				VersionID:            "2.5",
			},
			wantErr: false,
		},
		{
			name:  "minimal MSH segment",
			input: "MSH|^~\\&|||||||ADT^A01|MSG00001|P|2.5.1",
			want: &MSH{
				FieldSeparator:     "|",
				EncodingCharacters: "^~\\&",
				MessageType:        "ADT^A01",
				MessageControlID:   "MSG00001",
				ProcessingID:       "P",
				VersionID:          "2.5.1",
			},
			wantErr: false,
		},
		{
			name:     "nil segment",
			input:    "",
			want:     nil,
			wantErr:  true,
			errMatch: "segment is nil",
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

			got, err := ParseMSH(seg)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseMSH() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseMSH() unexpected error: %v", err)
			}

			// Check key fields
			if got.FieldSeparator != tt.want.FieldSeparator {
				t.Errorf("FieldSeparator = %q, want %q", got.FieldSeparator, tt.want.FieldSeparator)
			}
			if got.EncodingCharacters != tt.want.EncodingCharacters {
				t.Errorf("EncodingCharacters = %q, want %q", got.EncodingCharacters, tt.want.EncodingCharacters)
			}
			if got.SendingApplication != tt.want.SendingApplication {
				t.Errorf("SendingApplication = %q, want %q", got.SendingApplication, tt.want.SendingApplication)
			}
			if got.SendingFacility != tt.want.SendingFacility {
				t.Errorf("SendingFacility = %q, want %q", got.SendingFacility, tt.want.SendingFacility)
			}
			if got.ReceivingApplication != tt.want.ReceivingApplication {
				t.Errorf("ReceivingApplication = %q, want %q", got.ReceivingApplication, tt.want.ReceivingApplication)
			}
			if got.ReceivingFacility != tt.want.ReceivingFacility {
				t.Errorf("ReceivingFacility = %q, want %q", got.ReceivingFacility, tt.want.ReceivingFacility)
			}
			if got.DateTime != tt.want.DateTime {
				t.Errorf("DateTime = %q, want %q", got.DateTime, tt.want.DateTime)
			}
			if got.Security != tt.want.Security {
				t.Errorf("Security = %q, want %q", got.Security, tt.want.Security)
			}
			if got.MessageType != tt.want.MessageType {
				t.Errorf("MessageType = %q, want %q", got.MessageType, tt.want.MessageType)
			}
			if got.MessageControlID != tt.want.MessageControlID {
				t.Errorf("MessageControlID = %q, want %q", got.MessageControlID, tt.want.MessageControlID)
			}
			if got.ProcessingID != tt.want.ProcessingID {
				t.Errorf("ProcessingID = %q, want %q", got.ProcessingID, tt.want.ProcessingID)
			}
			if got.VersionID != tt.want.VersionID {
				t.Errorf("VersionID = %q, want %q", got.VersionID, tt.want.VersionID)
			}
		})
	}
}

func TestParseMSH_WrongSegment(t *testing.T) {
	input := "PID|1||12345^^^Hospital^MR||Doe^John"
	seg, err := hl7.ParseSegment([]rune(input), hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("failed to parse segment: %v", err)
	}

	_, err = ParseMSH(seg)
	if err == nil {
		t.Error("ParseMSH() expected error for non-MSH segment, got nil")
	}
}

func TestMSH_ToSegment(t *testing.T) {
	tests := []struct {
		name    string
		msh     *MSH
		wantErr bool
	}{
		{
			name: "complete MSH",
			msh: &MSH{
				FieldSeparator:       "|",
				EncodingCharacters:   "^~\\&",
				SendingApplication:   "TestApp",
				SendingFacility:      "TestFac",
				ReceivingApplication: "RecvApp",
				ReceivingFacility:    "RecvFac",
				DateTime:             "20230615143000",
				MessageType:          "ADT^A01",
				MessageControlID:     "MSG12345",
				ProcessingID:         "P",
				VersionID:            "2.5.1",
			},
			wantErr: false,
		},
		{
			name: "minimal MSH",
			msh: &MSH{
				MessageType:      "ADT^A01",
				MessageControlID: "MSG00001",
				ProcessingID:     "P",
				VersionID:        "2.5",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := tt.msh.ToSegment(hl7.DefaultDelimiters())

			if tt.wantErr {
				if err == nil {
					t.Error("ToSegment() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ToSegment() unexpected error: %v", err)
			}

			if seg.Name() != "MSH" {
				t.Errorf("segment name = %q, want MSH", seg.Name())
			}

			// Parse the created segment back and verify values
			parsed, err := ParseMSH(seg)
			if err != nil {
				t.Fatalf("failed to parse created segment: %v", err)
			}

			if parsed.MessageType != tt.msh.MessageType {
				t.Errorf("MessageType = %q, want %q", parsed.MessageType, tt.msh.MessageType)
			}
			if parsed.MessageControlID != tt.msh.MessageControlID {
				t.Errorf("MessageControlID = %q, want %q", parsed.MessageControlID, tt.msh.MessageControlID)
			}
			if parsed.VersionID != tt.msh.VersionID {
				t.Errorf("VersionID = %q, want %q", parsed.VersionID, tt.msh.VersionID)
			}
		})
	}
}

func TestMSH_RoundTrip(t *testing.T) {
	original := &MSH{
		FieldSeparator:       "|",
		EncodingCharacters:   "^~\\&",
		SendingApplication:   "SendApp",
		SendingFacility:      "SendFac",
		ReceivingApplication: "RecvApp",
		ReceivingFacility:    "RecvFac",
		DateTime:             "20230615143000",
		Security:             "SEC",
		MessageType:          "ORU^R01",
		MessageControlID:     "CTL123",
		ProcessingID:         "P",
		VersionID:            "2.5.1",
	}

	// Convert to segment
	seg, err := original.ToSegment(hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("ToSegment() error: %v", err)
	}

	// Parse back
	parsed, err := ParseMSH(seg)
	if err != nil {
		t.Fatalf("ParseMSH() error: %v", err)
	}

	// Verify all fields match
	if parsed.SendingApplication != original.SendingApplication {
		t.Errorf("SendingApplication = %q, want %q", parsed.SendingApplication, original.SendingApplication)
	}
	if parsed.SendingFacility != original.SendingFacility {
		t.Errorf("SendingFacility = %q, want %q", parsed.SendingFacility, original.SendingFacility)
	}
	if parsed.ReceivingApplication != original.ReceivingApplication {
		t.Errorf("ReceivingApplication = %q, want %q", parsed.ReceivingApplication, original.ReceivingApplication)
	}
	if parsed.ReceivingFacility != original.ReceivingFacility {
		t.Errorf("ReceivingFacility = %q, want %q", parsed.ReceivingFacility, original.ReceivingFacility)
	}
	if parsed.DateTime != original.DateTime {
		t.Errorf("DateTime = %q, want %q", parsed.DateTime, original.DateTime)
	}
	if parsed.MessageType != original.MessageType {
		t.Errorf("MessageType = %q, want %q", parsed.MessageType, original.MessageType)
	}
	if parsed.MessageControlID != original.MessageControlID {
		t.Errorf("MessageControlID = %q, want %q", parsed.MessageControlID, original.MessageControlID)
	}
	if parsed.ProcessingID != original.ProcessingID {
		t.Errorf("ProcessingID = %q, want %q", parsed.ProcessingID, original.ProcessingID)
	}
	if parsed.VersionID != original.VersionID {
		t.Errorf("VersionID = %q, want %q", parsed.VersionID, original.VersionID)
	}
}
