// Package marshal provides struct marshaling and unmarshaling for HL7 v2.x messages.
//
// The marshal package enables bidirectional conversion between Go structs and
// HL7 messages using struct tags to specify field locations. This provides a
// type-safe, idiomatic Go way to work with HL7 message data.
//
// # Struct Tags
//
// Use the "hl7" struct tag to map struct fields to HL7 locations:
//
//	type Patient struct {
//	    ID        string    `hl7:"PID.3.1"`
//	    LastName  string    `hl7:"PID.5.1"`
//	    FirstName string    `hl7:"PID.5.2"`
//	    DOB       time.Time `hl7:"PID.7"`
//	    Gender    string    `hl7:"PID.8"`
//	    Address   string    `hl7:"PID.11"`
//	}
//
// # Unmarshaling (HL7 to Struct)
//
// Extract data from an HL7 message into a Go struct:
//
//	// Parse the HL7 message
//	p := parse.New()
//	msg, err := p.Parse(hl7Data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Unmarshal into struct
//	var patient Patient
//	if err := marshal.Unmarshal(msg, &patient); err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Patient: %s %s (ID: %s)\n",
//	    patient.FirstName, patient.LastName, patient.ID)
//
// # Marshaling (Struct to HL7)
//
// Populate an HL7 message from a Go struct:
//
//	patient := Patient{
//	    ID:        "12345",
//	    LastName:  "SMITH",
//	    FirstName: "JOHN",
//	    DOB:       time.Date(1980, 6, 15, 0, 0, 0, 0, time.UTC),
//	    Gender:    "M",
//	}
//
//	// Create or parse a base message
//	msg := createBaseMessage()
//
//	// Marshal struct into message
//	if err := marshal.Marshal(&patient, msg); err != nil {
//	    log.Fatal(err)
//	}
//
// # Supported Types
//
// The marshaler supports these Go types:
//   - string: Direct mapping to HL7 text values
//   - int, int8, int16, int32, int64: Numeric values
//   - uint, uint8, uint16, uint32, uint64: Unsigned numeric values
//   - float32, float64: Floating-point values
//   - bool: Boolean values (true/false, yes/no, Y/N)
//   - time.Time: Date and time values (configurable format)
//   - *T: Pointers to any supported type (nil = empty field)
//   - []T: Slices for repeating fields
//
// # Marshaler Options
//
// Configure marshaling behavior with functional options:
//
//	// Use a custom struct tag name
//	m := marshal.NewMarshaler(marshal.WithTagName("custom"))
//
//	// Omit zero-value fields when marshaling
//	m := marshal.NewMarshaler(marshal.WithOmitEmpty(true))
//
//	// Set time format (default: "20060102150405")
//	m := marshal.NewMarshaler(marshal.WithTimeFormat("20060102"))
//
//	// Set timezone for time parsing
//	loc, _ := time.LoadLocation("America/New_York")
//	m := marshal.NewMarshaler(marshal.WithTimeLocation(loc))
//
// # Time Formats
//
// Common HL7 time formats:
//   - "20060102" - Date only (YYYYMMDD)
//   - "20060102150405" - Date and time (YYYYMMDDHHMMSS)
//   - "20060102150405.000" - With milliseconds
//   - "20060102150405-0700" - With timezone offset
//
// Example with custom time format:
//
//	type Appointment struct {
//	    ScheduledTime time.Time `hl7:"SCH.11"`
//	}
//
//	m := marshal.NewMarshaler(
//	    marshal.WithTimeFormat("200601021504"),
//	)
//
//	var appt Appointment
//	err := m.Unmarshal(msg, &appt)
//
// # Repeating Fields
//
// Handle repeating HL7 fields with slices:
//
//	type Patient struct {
//	    IDs       []string `hl7:"PID.3"` // Multiple patient identifiers
//	    Allergies []string `hl7:"AL1.3"` // Multiple allergies (across segments)
//	}
//
// # Nested Structs
//
// Use nested structs for complex data types:
//
//	type Name struct {
//	    Family string `hl7:"1"`  // Component 1
//	    Given  string `hl7:"2"`  // Component 2
//	    Middle string `hl7:"3"`  // Component 3
//	    Suffix string `hl7:"4"`  // Component 4
//	    Prefix string `hl7:"5"`  // Component 5
//	}
//
//	type Patient struct {
//	    Name Name `hl7:"PID.5"`  // Maps to PID-5 (patient name)
//	}
//
// # Example: ADT Message Processing
//
//	// Define structs for ADT message
//	type ADTMessage struct {
//	    MessageType    string    `hl7:"MSH.9"`
//	    SendingApp     string    `hl7:"MSH.3"`
//	    MessageTime    time.Time `hl7:"MSH.7"`
//	    PatientID      string    `hl7:"PID.3.1"`
//	    PatientName    string    `hl7:"PID.5"`
//	    DOB            time.Time `hl7:"PID.7"`
//	    Gender         string    `hl7:"PID.8"`
//	    AttendingDoc   string    `hl7:"PV1.7"`
//	    AdmitTime      time.Time `hl7:"PV1.44"`
//	}
//
//	// Parse and unmarshal
//	msg, _ := parse.New().Parse(data)
//	var adt ADTMessage
//	if err := marshal.Unmarshal(msg, &adt); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Process the structured data
//	fmt.Printf("ADT %s received for patient %s\n",
//	    adt.MessageType, adt.PatientName)
//
// # Error Handling
//
// Marshaling errors include field information:
//
//	err := marshal.Unmarshal(msg, &patient)
//	if err != nil {
//	    var fieldErr *marshal.FieldError
//	    if errors.As(err, &fieldErr) {
//	        fmt.Printf("Error in field %s at %s: %v\n",
//	            fieldErr.Field, fieldErr.Location, fieldErr.Cause)
//	    }
//	}
package marshal
