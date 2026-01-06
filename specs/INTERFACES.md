# GoLevel7 v2 - Interface Quick Reference

This document provides a consolidated view of all public interfaces in GoLevel7 v2.

---

## Core Domain Interfaces (hl7 package)

### Message
```go
type Message interface {
    // Segment access
    Segment(name string) (Segment, bool)
    Segments(name string) []Segment
    AllSegments() []Segment

    // String query interface
    Get(location string) (string, error)
    GetAll(location string) ([]string, error)
    Set(location string, value string) error

    // Structured query interface
    GetAt(loc *Location) (string, error)
    GetAllAt(loc *Location) ([]string, error)
    SetAt(loc *Location, value string) error

    // Segment manipulation
    AddSegment(seg Segment) error
    InsertSegment(index int, seg Segment) error
    RemoveSegment(name string) bool

    // Serialization
    Bytes() []byte
    String() string

    // Metadata
    Type() string
    ControlID() string
    Version() string
    Delimiters() *Delimiters
}
```

### Segment
```go
type Segment interface {
    Name() string
    Field(seq int) (Field, bool)
    Fields(seq int) []Field
    AllFields() []Field
    FieldCount() int
    Get(location string) (string, error)
    GetAll(location string) ([]string, error)
    Set(location string, value string) error
    SetField(seq int, field Field) error
    AddField(field Field) error
    Bytes(delims *Delimiters) []byte
    String() string
}
```

### Field
```go
type Field interface {
    SeqNum() int
    Value() string
    Component(index int) (Component, bool)
    Components() []Component
    Repetition(index int) (Repetition, bool)
    Repetitions() []Repetition
    RepetitionCount() int
    Get(location string) (string, error)
    Set(location string, value string) error
    Bytes(delims *Delimiters) []byte
    String() string
}
```

### Repetition
```go
type Repetition interface {
    Value() string
    Component(index int) (Component, bool)
    Components() []Component
    Bytes(delims *Delimiters) []byte
    String() string
}
```

### Component
```go
type Component interface {
    Value() string
    SubComponent(index int) (SubComponent, bool)
    SubComponents() []SubComponent
    Set(value string) error
    SetSubComponent(index int, value string) error
    Bytes(delims *Delimiters) []byte
    String() string
}
```

### SubComponent
```go
type SubComponent interface {
    Value() string
    Set(value string) error
    Bytes(delims *Delimiters) []byte
    String() string
}
```

### Escaper
```go
type Escaper interface {
    Escape(value string, delims *Delimiters) string
    Unescape(value string, delims *Delimiters) string
}
```

---

## Parser Interfaces (parse package)

### Parser
```go
type Parser interface {
    Parse(data []byte) (hl7.Message, error)
    ParseContext(ctx context.Context, data []byte) (hl7.Message, error)
}
```

### Scanner
```go
type Scanner interface {
    Scan() bool
    Message() hl7.Message
    Err() error
}
```

---

## Encoder Interfaces (encode package)

### Encoder
```go
type Encoder interface {
    Encode(msg hl7.Message) ([]byte, error)
    EncodeToWriter(ctx context.Context, w io.Writer, msg hl7.Message) error
}
```

### Writer
```go
type Writer interface {
    Write(msg hl7.Message) error
    Flush() error
    Close() error
}
```

---

## Marshal Interfaces (marshal package)

### Marshaler
```go
type Marshaler interface {
    Marshal(v interface{}) (hl7.Message, error)
    MarshalInto(msg hl7.Message, v interface{}) error
}
```

### Unmarshaler
```go
type Unmarshaler interface {
    Unmarshal(msg hl7.Message, v interface{}) error
}
```

---

## Validation Interfaces (validate package)

### Validator
```go
type Validator interface {
    Validate(msg hl7.Message) ValidationResult
    ValidateSegment(seg hl7.Segment) ValidationResult
}
```

### ValidationResult
```go
type ValidationResult interface {
    Valid() bool
    Errors() []ValidationError
    Warnings() []ValidationWarning
}
```

### Rule
```go
type Rule interface {
    Validate(msg hl7.Message) []ValidationError
    Location() string
    Description() string
}
```

### RuleBuilder
```go
type RuleBuilder interface {
    Required() RuleBuilder
    Value(expected string) RuleBuilder
    Pattern(pattern string) RuleBuilder
    Length(min, max int) RuleBuilder
    OneOf(values ...string) RuleBuilder
    Custom(fn func(value string) error) RuleBuilder
    Build() Rule
}
```

### RuleSet
```go
type RuleSet interface {
    Rules() []Rule
    Add(rules ...Rule) RuleSet
    Merge(other RuleSet) RuleSet
}
```

---

## ACK Interfaces (ack package)

### Builder
```go
type Builder interface {
    Accept(original hl7.Message) (hl7.Message, error)
    Reject(original hl7.Message, reason string) (hl7.Message, error)
    Error(original hl7.Message, err error) (hl7.Message, error)
    Custom(original hl7.Message, ack ACK) (hl7.Message, error)
}
```

---

## MLLP Interfaces (mllp package)

### Client
```go
type Client interface {
    Send(ctx context.Context, msg hl7.Message) (hl7.Message, error)
    SendAsync(ctx context.Context, msg hl7.Message) error
    Close() error
}
```

### Server
```go
type Server interface {
    Serve(listener net.Listener) error
    Shutdown(ctx context.Context) error
}
```

### Handler
```go
type Handler interface {
    HandleMessage(ctx context.Context, msg hl7.Message) (hl7.Message, error)
}
```

---

## Usage Examples

### Basic Parsing
```go
parser := parse.New()
msg, err := parser.Parse(data)
if err != nil {
    log.Fatal(err)
}

// Query by location string
patientName, _ := msg.Get("PID.5")

// Query with structured location
loc, _ := hl7.ParseLocation("PID.5.1")
lastName, _ := msg.GetAt(loc)
```

### Struct Mapping
```go
type Patient struct {
    ID        string `hl7:"PID.3"`
    LastName  string `hl7:"PID.5.1"`
    FirstName string `hl7:"PID.5.2"`
}

unmarshaler := marshal.NewUnmarshaler()
var patient Patient
err := unmarshaler.Unmarshal(msg, &patient)
```

### Validation
```go
validator := validate.New(
    validate.At("MSH.9").Required().Build(),
    validate.At("PID.3").Required().Build(),
    validate.At("PID.5").Required().Build(),
)

result := validator.Validate(msg)
if !result.Valid() {
    for _, err := range result.Errors() {
        log.Printf("Validation error at %s: %v", err.Location(), err)
    }
}
```

### MLLP Server
```go
handler := mllp.HandlerFunc(func(ctx context.Context, msg hl7.Message) (hl7.Message, error) {
    // Process message
    log.Printf("Received: %s", msg.Type())

    // Return acknowledgment
    return ack.NewBuilder().Accept(msg)
})

server := mllp.NewServer(mllp.WithHandler(handler))
listener, _ := net.Listen("tcp", ":2575")
server.Serve(listener)
```
