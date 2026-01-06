// Package segments provides typed helper structs for common HL7 v2.x segments.
//
// Each segment type provides:
//   - A struct with fields corresponding to HL7 field positions, tagged with `hl7:"SEG.N"` tags
//   - A ParseXXX function to extract data from an hl7.Segment interface into the typed struct
//   - A ToSegment method to convert the typed struct back into an hl7.Segment
//
// The segment helpers simplify working with common HL7 segments by providing type-safe access
// to field values without the need to remember field positions.
//
// # Supported Segments
//
// The following segments have typed helpers:
//   - MSH (Message Header) - msh.go
//   - PID (Patient Identification) - pid.go
//   - PV1 (Patient Visit) - pv1.go
//   - OBR (Observation Request) - obr.go
//   - OBX (Observation Result) - obx.go
//   - ORC (Common Order) - orc.go
//
// # Usage Example
//
// Parsing a segment:
//
//	msg, err := parser.Parse(data)
//	if err != nil {
//	    return err
//	}
//
//	pidSeg, ok := msg.Segment("PID")
//	if !ok {
//	    return errors.New("PID segment not found")
//	}
//
//	pid, err := segments.ParsePID(pidSeg)
//	if err != nil {
//	    return err
//	}
//
//	fmt.Println("Patient Name:", pid.PatientName)
//	fmt.Println("Date of Birth:", pid.DateOfBirth)
//
// Creating a segment:
//
//	msh := &segments.MSH{
//	    FieldSeparator:      "|",
//	    EncodingCharacters:  "^~\\&",
//	    SendingApplication:  "MyApp",
//	    SendingFacility:     "MyFacility",
//	    ReceivingApplication: "TargetApp",
//	    ReceivingFacility:   "TargetFacility",
//	    DateTime:            time.Now().Format("20060102150405"),
//	    MessageType:         "ADT^A01",
//	    MessageControlID:    "MSG00001",
//	    ProcessingID:        "P",
//	    VersionID:           "2.5.1",
//	}
//
//	seg, err := msh.ToSegment(hl7.DefaultDelimiters())
//	if err != nil {
//	    return err
//	}
//
// # Field Numbering
//
// Field numbers follow the HL7 standard where:
//   - MSH-1 is the field separator character itself (|)
//   - MSH-2 is the encoding characters (^~\&)
//   - Other segments start field numbering at 1 for the first data field after the segment name
//
// # Component Access
//
// For fields with components (e.g., PID-5 Patient Name), the helper stores the full field value.
// To access individual components, you can:
//   - Parse the field value using the component separator (^)
//   - Use the underlying hl7.Segment interface for component-level access
//
// # Repetitions
//
// Fields that can repeat store only the first repetition in the typed struct.
// For full repetition access, use the underlying hl7.Segment interface.
package segments
