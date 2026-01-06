# GoLevel7

[![Go Reference](https://pkg.go.dev/badge/github.com/dshills/golevel7.svg)](https://pkg.go.dev/github.com/dshills/golevel7)
[![Go Report Card](https://goreportcard.com/badge/github.com/dshills/golevel7)](https://goreportcard.com/report/github.com/dshills/golevel7)

A comprehensive Go library for parsing, encoding, and manipulating HL7 v2.x healthcare messages.

## Features

- **Full HL7 v2.x Support**: Parse and encode all HL7 v2.x message types
- **Streaming Parser**: Memory-efficient parsing with MLLP frame support
- **Struct Marshaling**: Map HL7 data to Go structs using tags
- **Rule-Based Validation**: Flexible validation with built-in and custom rules
- **MLLP Network Support**: Client/server implementation for HL7 transport
- **ACK Generation**: Automatic acknowledgment message creation
- **Escape Sequence Handling**: Full support for HL7 escape sequences
- **DoS Protection**: Built-in limits for segment count and field length

## Installation

```bash
go get github.com/dshills/golevel7
```

Requires Go 1.21 or later.

## Quick Start

### Parsing a Message

```go
package main

import (
    "fmt"
    "log"

    "github.com/dshills/golevel7/parse"
)

func main() {
    hl7Data := []byte(`MSH|^~\&|SENDING|FACILITY|||202401151200||ADT^A01|MSG001|P|2.5
PID|1||12345^^^MRN||Smith^John^A||19800115|M`)

    // Create parser and parse message
    p := parse.New()
    msg, err := p.Parse(hl7Data)
    if err != nil {
        log.Fatal(err)
    }

    // Access message data
    fmt.Printf("Message Type: %s\n", msg.Type())
    fmt.Printf("Control ID: %s\n", msg.ControlID())

    // Get specific fields
    patientID, _ := msg.Get("PID.3.1")
    patientName, _ := msg.Get("PID.5")
    fmt.Printf("Patient: %s (ID: %s)\n", patientName, patientID)
}
```

### Struct Marshaling

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/dshills/golevel7/marshal"
    "github.com/dshills/golevel7/parse"
)

type Patient struct {
    ID        string    `hl7:"PID.3.1"`
    LastName  string    `hl7:"PID.5.1"`
    FirstName string    `hl7:"PID.5.2"`
    DOB       time.Time `hl7:"PID.7"`
    Gender    string    `hl7:"PID.8"`
}

func main() {
    // Parse the message
    p := parse.New()
    msg, err := p.Parse(hl7Data)
    if err != nil {
        log.Fatal(err)
    }

    // Unmarshal into struct
    var patient Patient
    if err := marshal.Unmarshal(msg, &patient); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Patient: %s %s (DOB: %s)\n",
        patient.FirstName, patient.LastName, patient.DOB.Format("2006-01-02"))
}
```

### Validation

```go
package main

import (
    "fmt"

    "github.com/dshills/golevel7/parse"
    "github.com/dshills/golevel7/validate"
)

func main() {
    p := parse.New()
    msg, _ := p.Parse(hl7Data)

    // Create validator with rules
    v := validate.NewValidator(
        validate.Required("MSH.9"),         // Message type required
        validate.Required("PID.3.1"),       // Patient ID required
        validate.Required("PID.5"),         // Patient name required
        validate.OneOf("PID.8", "M", "F", "O", "U"), // Valid gender codes
        validate.Pattern("PID.7", `^\d{8}$`), // DOB format YYYYMMDD
    )

    // Validate message
    if errors := v.Validate(msg); len(errors) > 0 {
        for _, err := range errors {
            fmt.Printf("Validation error at %s: %s\n", err.Location, err.Message)
        }
    }
}
```

### MLLP Server

```go
package main

import (
    "context"
    "log"

    "github.com/dshills/golevel7/ack"
    "github.com/dshills/golevel7/hl7"
    "github.com/dshills/golevel7/mllp"
)

func main() {
    handler := func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
        log.Printf("Received: %s", msg.Type())

        // Process message...

        // Return acknowledgment
        return ack.Generate(msg, ack.Accept())
    }

    server := mllp.NewServer(":2575", handler,
        mllp.WithReadTimeout(30*time.Second),
        mllp.WithWriteTimeout(30*time.Second),
    )

    log.Println("Starting MLLP server on :2575")
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

### MLLP Client

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/dshills/golevel7/mllp"
    "github.com/dshills/golevel7/parse"
)

func main() {
    // Connect to MLLP server
    client, err := mllp.Dial("localhost:2575",
        mllp.WithDialTimeout(10*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Parse and send message
    p := parse.New()
    msg, _ := p.Parse(hl7Data)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    ackMsg, err := client.Send(ctx, msg)
    if err != nil {
        log.Fatal(err)
    }

    // Check acknowledgment
    ackCode, _ := ackMsg.Get("MSA.1")
    log.Printf("ACK Code: %s", ackCode)
}
```

## Package Overview

### `hl7` - Core Types

The core package provides the fundamental HL7 data types:

- `Message` - Complete HL7 message with segment access
- `Segment` - Individual segment (e.g., MSH, PID, OBX)
- `Field` - Field within a segment with repetition support
- `Component` - Component within a field
- `SubComponent` - Subcomponent within a component
- `Delimiters` - HL7 delimiter configuration
- `Location` - HL7 location addressing (e.g., "PID.3.1")

```go
// Create a new message
msg := hl7.NewEmptyMessage()

// Set values using location strings
msg.Set("MSH.9", "ADT^A01")
msg.Set("PID.3.1", "12345")
msg.Set("PID.5", "Smith^John^A")

// Get values
msgType, _ := msg.Get("MSH.9")
patientID, _ := msg.Get("PID.3.1")

// Access segments
pid, ok := msg.Segment("PID")
if ok {
    name, _ := pid.Get("5")
}

// Iterate segments
for _, seg := range msg.AllSegments() {
    fmt.Printf("Segment: %s\n", seg.Name())
}
```

### `parse` - Message Parsing

Parse HL7 messages with configurable options:

```go
// Basic parsing
p := parse.New()
msg, err := p.Parse(data)

// Strict mode with DoS protection
p := parse.New(
    parse.WithStrictMode(true),
    parse.WithMaxSegments(500),
    parse.WithMaxFieldLength(32768),
)

// Custom delimiters
delims := &hl7.Delimiters{
    Field:        '|',
    Component:    '^',
    Repetition:   '~',
    Escape:       '\\',
    SubComponent: '&',
}
p := parse.New(parse.WithCustomDelimiters(delims))

// Different segment terminator
p := parse.New(parse.WithSegmentTerminator('\n'))
```

**Streaming with Scanner:**

```go
reader := strings.NewReader(multipleMessages)
scanner := parse.NewScanner(reader)

for scanner.Scan() {
    msg := scanner.Message()
    fmt.Printf("Parsed: %s\n", msg.Type())
}

if err := scanner.Err(); err != nil {
    log.Fatal(err)
}
```

### `encode` - Message Encoding

Encode messages back to HL7 format:

```go
encoder := encode.New()

// Encode to bytes
data, err := encoder.Encode(msg)

// Encode to string
str, err := encoder.EncodeString(msg)

// With MLLP framing
encoder := encode.New(encode.WithMLLPFraming(true))
framedData, _ := encoder.Encode(msg)
```

### `marshal` - Struct Marshaling

Map between Go structs and HL7 messages:

```go
type ADTMessage struct {
    MessageType    string    `hl7:"MSH.9"`
    SendingApp     string    `hl7:"MSH.3"`
    MessageTime    time.Time `hl7:"MSH.7"`
    PatientID      string    `hl7:"PID.3.1"`
    PatientName    string    `hl7:"PID.5"`
    DOB            time.Time `hl7:"PID.7"`
    Gender         string    `hl7:"PID.8"`
}

// Unmarshal (HL7 -> struct)
var adt ADTMessage
err := marshal.Unmarshal(msg, &adt)

// Marshal (struct -> HL7)
err := marshal.Marshal(&adt, msg)
```

**Marshaler Options:**

```go
m := marshal.NewMarshaler(
    marshal.WithTagName("custom"),           // Custom tag name
    marshal.WithOmitEmpty(true),             // Skip zero values
    marshal.WithTimeFormat("20060102"),      // Date format
    marshal.WithTimeLocation(time.UTC),      // Timezone
)
```

**Supported Types:**
- `string` - Direct text mapping
- `int`, `int8`, `int16`, `int32`, `int64` - Integers
- `uint`, `uint8`, `uint16`, `uint32`, `uint64` - Unsigned integers
- `float32`, `float64` - Floating-point numbers
- `bool` - Boolean values
- `time.Time` - Date/time values
- `*T` - Pointers (nil = empty field)
- `[]T` - Slices for repeating fields

### `validate` - Message Validation

Validate messages with built-in and custom rules:

```go
v := validate.NewValidator(
    // Required fields
    validate.Required("MSH.9"),
    validate.Required("PID.3.1"),

    // Exact value
    validate.Value("MSH.9.1", "ADT"),

    // Pattern matching
    validate.Pattern("PID.7", `^\d{8}$`),

    // Length constraints
    validate.Length("PID.3.1", 1, 20),
    validate.MinLength("PID.5", 1),
    validate.MaxLength("NTE.3", 65536),

    // Allowed values
    validate.OneOf("PID.8", "M", "F", "O", "U"),

    // Custom validation
    validate.Custom("PID.7", func(value string) error {
        _, err := time.Parse("20060102", value)
        return err
    }),
)

errors := v.Validate(msg)
```

### `ack` - Acknowledgment Generation

Generate ACK/NAK responses:

```go
// Accept
ackMsg, _ := ack.Generate(msg, ack.Accept())

// Accept with text
ackMsg, _ := ack.Generate(msg,
    ack.Accept(),
    ack.WithText("Message processed successfully"),
)

// Error
ackMsg, _ := ack.Generate(msg,
    ack.Error("Validation failed"),
    ack.WithErrorCode("100"),
    ack.WithErrorLocation("PID.3"),
)

// Reject
ackMsg, _ := ack.Generate(msg,
    ack.Reject("Unsupported message type"),
)
```

**ACK Codes:**
- `AA` - Application Accept
- `AE` - Application Error
- `AR` - Application Reject

### `mllp` - Network Transport

MLLP client/server for HL7 over TCP:

**Server:**

```go
handler := func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
    // Process message
    return ack.Generate(msg, ack.Accept())
}

server := mllp.NewServer(":2575", handler,
    mllp.WithReadTimeout(60*time.Second),
    mllp.WithWriteTimeout(30*time.Second),
    mllp.WithMaxMessageSize(1024*1024),
    mllp.WithTLS(tlsConfig),
)

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
server.Shutdown(ctx)
```

**Client:**

```go
client, err := mllp.Dial("localhost:2575",
    mllp.WithDialTimeout(10*time.Second),
    mllp.WithClientTLS(tlsConfig),
    mllp.WithRetry(3, time.Second),
)
defer client.Close()

ackMsg, err := client.Send(ctx, msg)
```

### `segments` - Segment Helpers

Helper functions for common segments:

```go
// MSH helpers
msgType := segments.GetMessageType(msg)      // "ADT^A01"
controlID := segments.GetControlID(msg)      // "MSG001"
version := segments.GetVersion(msg)          // "2.5.1"

// PID helpers
patientID := segments.GetPatientID(msg)      // "12345"
patientName := segments.GetPatientName(msg)  // "Smith^John^A"

// OBX helpers
for _, obx := range msg.Segments("OBX") {
    obsID := segments.GetObservationID(obx)
    value := segments.GetObservationValue(obx)
    units := segments.GetObservationUnits(obx)
}
```

## HL7 Location Syntax

Access HL7 data using location strings:

| Location | Description |
|----------|-------------|
| `PID` | PID segment |
| `PID.3` | PID field 3 |
| `PID.3.1` | PID field 3, component 1 |
| `PID.3.1.2` | PID field 3, component 1, subcomponent 2 |
| `PID.3[0]` | PID field 3, first repetition |
| `PID.3[1].1` | PID field 3, second repetition, component 1 |

## Error Handling

All packages return detailed errors:

```go
msg, err := parser.Parse(data)
if err != nil {
    var parseErr *hl7.ParseError
    if errors.As(err, &parseErr) {
        fmt.Printf("Parse error at line %d: %s\n", parseErr.Line, parseErr.Message)
    }
}

errors := validator.Validate(msg)
for _, err := range errors {
    fmt.Printf("Validation error at %s: %s\n", err.Location, err.Message)
}
```

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Run benchmarks
make bench

# Generate coverage report
make coverage

# Format code
make fmt

# Run all pre-commit checks
make pre-commit

# Show all available commands
make help
```

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass (`make test`)
2. Code passes linting (`make lint`)
3. New features include tests
4. Code follows existing patterns

## License

MIT License - see LICENSE file for details.

## Acknowledgments

This library implements the HL7 v2.x specification for healthcare message interchange.
For more information about HL7, visit [hl7.org](https://www.hl7.org/).
