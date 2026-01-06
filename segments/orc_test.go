package segments

import (
	"testing"

	"github.com/dshills/golevel7/hl7"
)

func TestParseORC(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *ORC
		wantErr bool
	}{
		{
			name:  "new order ORC",
			input: "ORC|NW|P001^Placer|F001^Filler||SC|||||||1234^Ordering^Dr||555-123-4567|20230615090000",
			want: &ORC{
				OrderControl:           "NW",
				PlacerOrderNumber:      "P001^Placer",
				FillerOrderNumber:      "F001^Filler",
				OrderStatus:            "SC",
				OrderingProvider:       "1234^Ordering^Dr",
				CallBackPhoneNumber:    "555-123-4567",
				OrderEffectiveDateTime: "20230615090000",
			},
			wantErr: false,
		},
		{
			name:  "cancel order ORC",
			input: "ORC|CA|P002|F002||CM|||||||5678^Provider^Jane",
			want: &ORC{
				OrderControl:      "CA",
				PlacerOrderNumber: "P002",
				FillerOrderNumber: "F002",
				OrderStatus:       "CM",
				OrderingProvider:  "5678^Provider^Jane",
			},
			wantErr: false,
		},
		{
			name:  "status changed ORC",
			input: "ORC|SC|P003|F003|G001|IP",
			want: &ORC{
				OrderControl:      "SC",
				PlacerOrderNumber: "P003",
				FillerOrderNumber: "F003",
				PlacerGroupNumber: "G001",
				OrderStatus:       "IP",
			},
			wantErr: false,
		},
		{
			name:  "ORC with organization info",
			input: "ORC|NW|P004||||||||||||||||Hospital XYZ|123 Medical Way^^City^ST^12345|555-999-8888",
			want: &ORC{
				OrderControl:                "NW",
				PlacerOrderNumber:           "P004",
				OrderingFacilityName:        "Hospital XYZ",
				OrderingFacilityAddress:     "123 Medical Way^^City^ST^12345",
				OrderingFacilityPhoneNumber: "555-999-8888",
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

			got, err := ParseORC(seg)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseORC() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseORC() unexpected error: %v", err)
			}

			// Check key fields
			if got.OrderControl != tt.want.OrderControl {
				t.Errorf("OrderControl = %q, want %q", got.OrderControl, tt.want.OrderControl)
			}
			if got.PlacerOrderNumber != tt.want.PlacerOrderNumber {
				t.Errorf("PlacerOrderNumber = %q, want %q", got.PlacerOrderNumber, tt.want.PlacerOrderNumber)
			}
			if got.FillerOrderNumber != tt.want.FillerOrderNumber {
				t.Errorf("FillerOrderNumber = %q, want %q", got.FillerOrderNumber, tt.want.FillerOrderNumber)
			}
			if got.OrderStatus != tt.want.OrderStatus {
				t.Errorf("OrderStatus = %q, want %q", got.OrderStatus, tt.want.OrderStatus)
			}
		})
	}
}

func TestParseORC_WrongSegment(t *testing.T) {
	input := "OBR|1|P001|F001|CBC^Complete Blood Count"
	seg, err := hl7.ParseSegment([]rune(input), hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("failed to parse segment: %v", err)
	}

	_, err = ParseORC(seg)
	if err == nil {
		t.Error("ParseORC() expected error for non-ORC segment, got nil")
	}
}

func TestORC_ToSegment(t *testing.T) {
	tests := []struct {
		name    string
		orc     *ORC
		wantErr bool
	}{
		{
			name: "new order",
			orc: &ORC{
				OrderControl:           "NW",
				PlacerOrderNumber:      "PLACER123",
				FillerOrderNumber:      "FILLER456",
				OrderStatus:            "SC",
				OrderingProvider:       "1234^Doctor^Test",
				OrderEffectiveDateTime: "20230615100000",
			},
			wantErr: false,
		},
		{
			name: "minimal ORC",
			orc: &ORC{
				OrderControl:      "NW",
				PlacerOrderNumber: "P001",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := tt.orc.ToSegment(hl7.DefaultDelimiters())

			if tt.wantErr {
				if err == nil {
					t.Error("ToSegment() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ToSegment() unexpected error: %v", err)
			}

			if seg.Name() != "ORC" {
				t.Errorf("segment name = %q, want ORC", seg.Name())
			}

			// Parse back and verify
			parsed, err := ParseORC(seg)
			if err != nil {
				t.Fatalf("failed to parse created segment: %v", err)
			}

			if parsed.OrderControl != tt.orc.OrderControl {
				t.Errorf("OrderControl = %q, want %q", parsed.OrderControl, tt.orc.OrderControl)
			}
			if parsed.PlacerOrderNumber != tt.orc.PlacerOrderNumber {
				t.Errorf("PlacerOrderNumber = %q, want %q", parsed.PlacerOrderNumber, tt.orc.PlacerOrderNumber)
			}
		})
	}
}

func TestORC_RoundTrip(t *testing.T) {
	original := &ORC{
		OrderControl:                "NW",
		PlacerOrderNumber:           "PLACER001^HospitalA",
		FillerOrderNumber:           "FILLER001^LabB",
		PlacerGroupNumber:           "GROUP001",
		OrderStatus:                 "IP",
		DateTimeOfTransaction:       "20230615140000",
		EnteredBy:                   "USER001^Clerk^Admin",
		OrderingProvider:            "DR001^Physician^Test^MD",
		OrderEffectiveDateTime:      "20230615150000",
		OrderingFacilityName:        "Main Hospital",
		OrderingFacilityAddress:     "100 Hospital Dr^^City^ST^12345",
		OrderingFacilityPhoneNumber: "555-123-4567",
	}

	// Convert to segment
	seg, err := original.ToSegment(hl7.DefaultDelimiters())
	if err != nil {
		t.Fatalf("ToSegment() error: %v", err)
	}

	// Parse back
	parsed, err := ParseORC(seg)
	if err != nil {
		t.Fatalf("ParseORC() error: %v", err)
	}

	// Verify fields
	if parsed.OrderControl != original.OrderControl {
		t.Errorf("OrderControl = %q, want %q", parsed.OrderControl, original.OrderControl)
	}
	if parsed.PlacerOrderNumber != original.PlacerOrderNumber {
		t.Errorf("PlacerOrderNumber = %q, want %q", parsed.PlacerOrderNumber, original.PlacerOrderNumber)
	}
	if parsed.FillerOrderNumber != original.FillerOrderNumber {
		t.Errorf("FillerOrderNumber = %q, want %q", parsed.FillerOrderNumber, original.FillerOrderNumber)
	}
	if parsed.PlacerGroupNumber != original.PlacerGroupNumber {
		t.Errorf("PlacerGroupNumber = %q, want %q", parsed.PlacerGroupNumber, original.PlacerGroupNumber)
	}
	if parsed.OrderStatus != original.OrderStatus {
		t.Errorf("OrderStatus = %q, want %q", parsed.OrderStatus, original.OrderStatus)
	}
	if parsed.OrderingProvider != original.OrderingProvider {
		t.Errorf("OrderingProvider = %q, want %q", parsed.OrderingProvider, original.OrderingProvider)
	}
	if parsed.OrderEffectiveDateTime != original.OrderEffectiveDateTime {
		t.Errorf("OrderEffectiveDateTime = %q, want %q", parsed.OrderEffectiveDateTime, original.OrderEffectiveDateTime)
	}
}

func TestORC_OrderControlCodes(t *testing.T) {
	orderControls := []struct {
		code        string
		description string
	}{
		{"NW", "New order"},
		{"CA", "Cancel order request"},
		{"OC", "Order canceled"},
		{"SC", "Status changed"},
		{"HD", "Hold order request"},
		{"RL", "Release previous hold"},
		{"XO", "Change order request"},
		{"CH", "Child order"},
		{"PA", "Parent order"},
		{"DC", "Discontinue order request"},
		{"OD", "Order discontinued"},
		{"RF", "Refill order request"},
		{"RE", "Release hold"},
	}

	for _, oc := range orderControls {
		t.Run(oc.description, func(t *testing.T) {
			original := &ORC{
				OrderControl:      oc.code,
				PlacerOrderNumber: "TEST001",
			}

			seg, err := original.ToSegment(hl7.DefaultDelimiters())
			if err != nil {
				t.Fatalf("ToSegment() error: %v", err)
			}

			parsed, err := ParseORC(seg)
			if err != nil {
				t.Fatalf("ParseORC() error: %v", err)
			}

			if parsed.OrderControl != oc.code {
				t.Errorf("OrderControl = %q, want %q", parsed.OrderControl, oc.code)
			}
		})
	}
}
