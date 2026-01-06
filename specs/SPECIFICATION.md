# GoLevel7 v2 - Technical Specification

## 1. Executive Summary

GoLevel7 v2 is a complete rewrite of the original golevel7 library for parsing, encoding, validating, and manipulating HL7 v2.x messages in Go. This version emphasizes strong interfaces, modern Go architecture patterns, and extensibility while fixing known limitations in the original implementation.

### Key Improvements Over v1
- **Interface-first design**: All major components defined by interfaces, enabling testing and extensibility
- **Fixed MSH segment handling**: Proper encoding/decoding of MSH segments
- **Repeating field/segment support**: Full support for HL7 repetition semantics
- **Context-aware operations**: All I/O operations accept `context.Context`
- **Functional options pattern**: Flexible, backwards-compatible configuration
- **Structured errors**: Rich error types with location and context information
- **Streaming support**: Memory-efficient processing of large message batches

---

## 2. HL7 v2.x Protocol Overview

### 2.1 Message Structure Hierarchy

```
Message
├── Segment (3-letter identifier, e.g., MSH, PID, OBR)
│   ├── Field (separated by |)
│   │   ├── Repetition (separated by ~)
│   │   │   ├── Component (separated by ^)
│   │   │   │   └── SubComponent (separated by &)
```

### 2.2 Standard Delimiters
| Delimiter | Character | Purpose |
|-----------|-----------|---------|
| Field | `\|` | Separates fields within a segment |
| Component | `^` | Separates components within a field |
| Repetition | `~` | Separates repeated fields |
| Escape | `\` | Escape character for special chars |
| SubComponent | `&` | Separates subcomponents |
| Segment Terminator | `\r` (CR) | Ends each segment |

### 2.3 MSH Segment Special Handling
The MSH (Message Segment Header) segment has unique parsing rules:
- MSH-1 is the field separator itself (not delimited)
- MSH-2 contains the encoding characters (^~\&)
- Field numbering starts at MSH-1, but actual parsing begins at MSH-3

### 2.4 Message Framing (MLLP)
Standard HL7 over TCP uses Minimal Lower Layer Protocol:
```
<VT> MESSAGE <FS><CR>
0x0B          0x1C 0x0D
```

---

## 3. Architecture Overview

### 3.1 Package Structure

```
golevel7v2/
├── hl7/                    # Core domain types and interfaces
│   ├── message.go          # Message interface and types
│   ├── segment.go          # Segment interface and types
│   ├── field.go            # Field interface and types
│   ├── component.go        # Component interface and types
│   ├── subcomponent.go     # SubComponent type
│   ├── location.go         # Location parsing and navigation
│   ├── delimiters.go       # Delimiter configuration
│   └── errors.go           # Domain error types
├── parse/                  # Parsing implementation
│   ├── parser.go           # Main parser interface and implementation
│   ├── lexer.go            # Tokenization
│   ├── options.go          # Parser options
│   └── scanner.go          # Streaming message scanner
├── encode/                 # Encoding implementation
│   ├── encoder.go          # Main encoder interface and implementation
│   ├── options.go          # Encoder options
│   └── writer.go           # Streaming message writer
├── marshal/                # Struct marshaling/unmarshaling
│   ├── marshal.go          # Marshal interface and implementation
│   ├── unmarshal.go        # Unmarshal interface and implementation
│   ├── tags.go             # Struct tag parsing
│   └── options.go          # Marshal options
├── validate/               # Message validation
│   ├── validator.go        # Validator interface and implementation
│   ├── rules.go            # Validation rule types
│   ├── ruleset.go          # Pre-built validation rulesets
│   └── options.go          # Validator options
├── ack/                    # Acknowledgment handling
│   ├── ack.go              # ACK/NAK generation
│   └── types.go            # ACK type definitions
├── mllp/                   # Network transport
│   ├── client.go           # MLLP client
│   ├── server.go           # MLLP server
│   ├── handler.go          # Message handler interface
│   └── options.go          # Transport options
├── segments/               # Common segment definitions
│   ├── msh.go              # MSH segment helpers
│   ├── pid.go              # PID segment helpers
│   ├── pv1.go              # PV1 segment helpers
│   └── ...                 # Other common segments
└── internal/               # Internal utilities
    ├── escape/             # Escape sequence handling
    └── runes/              # Rune manipulation utilities
```

### 3.2 Dependency Graph

```
                    ┌─────────────┐
                    │    mllp     │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
    ┌─────────┐      ┌──────────┐      ┌──────────┐
    │  parse  │      │  encode  │      │ validate │
    └────┬────┘      └────┬─────┘      └────┬─────┘
         │                │                 │
         └────────────────┼─────────────────┘
                          │
                          ▼
                    ┌──────────┐
                    │ marshal  │
                    └────┬─────┘
                         │
                         ▼
                    ┌──────────┐
                    │   hl7    │  (core domain)
                    └──────────┘
```

---

## 4. Core Interfaces

### 4.1 Message Interface

```go
package hl7

// Message represents a complete HL7 v2.x message
type Message interface {
    // Segment access
    Segment(name string) (Segment, bool)
    Segments(name string) []Segment
    AllSegments() []Segment

    // Query interface using location syntax
    Get(location string) (string, error)
    GetAll(location string) ([]string, error)
    Set(location string, value string) error

    // Structured access
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
    Type() string           // e.g., "ADT^A01"
    ControlID() string      // MSH-10
    Version() string        // e.g., "2.4"
    Delimiters() *Delimiters
}
```

### 4.2 Segment Interface

```go
package hl7

// Segment represents a single HL7 segment
type Segment interface {
    // Identity
    Name() string           // 3-letter segment ID (MSH, PID, etc.)

    // Field access
    Field(seq int) (Field, bool)
    Fields(seq int) []Field  // For repeating fields
    AllFields() []Field
    FieldCount() int

    // Query interface
    Get(location string) (string, error)
    GetAll(location string) ([]string, error)
    Set(location string, value string) error

    // Field manipulation
    SetField(seq int, field Field) error
    AddField(field Field) error

    // Serialization
    Bytes(delims *Delimiters) []byte
    String() string
}
```

### 4.3 Field Interface

```go
package hl7

// Field represents a single HL7 field (may contain repetitions)
type Field interface {
    // Identity
    SeqNum() int

    // Value access (first repetition)
    Value() string
    Component(index int) (Component, bool)
    Components() []Component

    // Repetition access
    Repetition(index int) (Repetition, bool)
    Repetitions() []Repetition
    RepetitionCount() int

    // Query interface
    Get(location string) (string, error)
    Set(location string, value string) error

    // Serialization
    Bytes(delims *Delimiters) []byte
    String() string
}

// Repetition represents a single repetition within a field
type Repetition interface {
    Value() string
    Component(index int) (Component, bool)
    Components() []Component

    Bytes(delims *Delimiters) []byte
    String() string
}
```

### 4.4 Component Interface

```go
package hl7

// Component represents a component within a field
type Component interface {
    Value() string
    SubComponent(index int) (SubComponent, bool)
    SubComponents() []SubComponent

    Set(value string) error
    SetSubComponent(index int, value string) error

    Bytes(delims *Delimiters) []byte
    String() string
}

// SubComponent is the atomic unit of HL7 data
type SubComponent interface {
    Value() string
    Set(value string) error
    Bytes(delims *Delimiters) []byte
    String() string
}
```

### 4.5 Parser Interface

```go
package parse

// Parser parses raw bytes into HL7 messages
type Parser interface {
    // Parse a single message
    Parse(data []byte) (hl7.Message, error)

    // Parse with context (for cancellation)
    ParseContext(ctx context.Context, data []byte) (hl7.Message, error)
}

// Scanner provides streaming message parsing
type Scanner interface {
    // Scan advances to the next message
    Scan() bool

    // Message returns the current message
    Message() hl7.Message

    // Err returns any error encountered
    Err() error
}
```

### 4.6 Encoder Interface

```go
package encode

// Encoder encodes HL7 messages to bytes
type Encoder interface {
    // Encode a message to bytes
    Encode(msg hl7.Message) ([]byte, error)

    // EncodeToWriter writes encoded message to writer
    EncodeToWriter(ctx context.Context, w io.Writer, msg hl7.Message) error
}

// Writer provides streaming message encoding
type Writer interface {
    // Write encodes and writes a message
    Write(msg hl7.Message) error

    // Flush ensures all buffered data is written
    Flush() error

    // Close closes the writer
    Close() error
}
```

### 4.7 Marshal Interfaces

```go
package marshal

// Marshaler converts Go structs to HL7 messages
type Marshaler interface {
    Marshal(v interface{}) (hl7.Message, error)
    MarshalInto(msg hl7.Message, v interface{}) error
}

// Unmarshaler converts HL7 messages to Go structs
type Unmarshaler interface {
    Unmarshal(msg hl7.Message, v interface{}) error
}
```

### 4.8 Validator Interface

```go
package validate

// Validator validates HL7 messages against rules
type Validator interface {
    // Validate checks a message against configured rules
    Validate(msg hl7.Message) ValidationResult

    // ValidateSegment checks a single segment
    ValidateSegment(seg hl7.Segment) ValidationResult
}

// ValidationResult contains validation outcomes
type ValidationResult interface {
    Valid() bool
    Errors() []ValidationError
    Warnings() []ValidationWarning
}

// ValidationError represents a validation failure
type ValidationError interface {
    error
    Location() string
    Code() string
    Severity() Severity
}
```

---

## 5. Domain Types

### 5.1 Location Type

```go
package hl7

// Location represents a position within an HL7 message
type Location struct {
    Segment      string // Segment name (e.g., "PID")
    SegmentIndex int    // For repeated segments (-1 for first/all)
    Field        int    // Field sequence number (1-based, HL7 standard)
    Repetition   int    // Repetition index (0-based, -1 for first/all)
    Component    int    // Component index (1-based, HL7 standard)
    SubComponent int    // SubComponent index (1-based, HL7 standard)
}

// ParseLocation parses a location string
// Format: SEG[idx].field[rep].component.subcomponent
// Examples:
//   - "PID" -> entire PID segment
//   - "PID.5" -> PID field 5
//   - "PID.5.1" -> PID field 5, component 1
//   - "PID.5[0].1" -> PID field 5, first repetition, component 1
//   - "PID[1].5" -> Second PID segment, field 5
func ParseLocation(s string) (*Location, error)
```

### 5.2 Delimiters Type

```go
package hl7

// Delimiters defines HL7 message delimiters
type Delimiters struct {
    Field        rune // Default: |
    Component    rune // Default: ^
    Repetition   rune // Default: ~
    Escape       rune // Default: \
    SubComponent rune // Default: &
    Truncation   rune // Default: # (v2.7+)
}

// DefaultDelimiters returns standard HL7 delimiters
func DefaultDelimiters() *Delimiters

// String returns the encoding characters (MSH-2 value)
func (d *Delimiters) String() string

// ParseDelimiters extracts delimiters from MSH-1 and MSH-2
func ParseDelimiters(mshSegment []byte) (*Delimiters, error)
```

### 5.3 Error Types

```go
package hl7

// ParseError indicates a parsing failure
type ParseError struct {
    Message  string
    Location string
    Line     int
    Column   int
    Cause    error
}

// LocationError indicates an invalid location
type LocationError struct {
    Location string
    Reason   string
}

// ValidationError indicates a validation failure
type ValidationError struct {
    Location string
    Rule     string
    Expected string
    Actual   string
    Severity Severity
}

type Severity int

const (
    SeverityError Severity = iota
    SeverityWarning
    SeverityInfo
)
```

---

## 6. Configuration Options

### 6.1 Parser Options

```go
package parse

type ParserOption func(*parserConfig)

// WithStrictMode enables strict HL7 parsing
func WithStrictMode(strict bool) ParserOption

// WithAllowEmptySegments allows segments with no fields
func WithAllowEmptySegments(allow bool) ParserOption

// WithCustomDelimiters uses non-standard delimiters
func WithCustomDelimiters(d *hl7.Delimiters) ParserOption

// WithMaxSegments limits segments per message (DoS protection)
func WithMaxSegments(max int) ParserOption

// WithMaxFieldLength limits field length (DoS protection)
func WithMaxFieldLength(max int) ParserOption

// WithSegmentTerminator sets custom segment terminator
func WithSegmentTerminator(term rune) ParserOption
```

### 6.2 Encoder Options

```go
package encode

type EncoderOption func(*encoderConfig)

// WithLineEnding sets the segment terminator
func WithLineEnding(ending string) EncoderOption

// WithMLLP wraps output in MLLP framing
func WithMLLP(enable bool) EncoderOption

// WithTrailingDelimiters includes trailing empty delimiters
func WithTrailingDelimiters(include bool) EncoderOption
```

### 6.3 Marshal Options

```go
package marshal

type MarshalOption func(*marshalConfig)

// WithTagName sets the struct tag name (default: "hl7")
func WithTagName(name string) MarshalOption

// WithOmitEmpty skips zero-value fields
func WithOmitEmpty(omit bool) MarshalOption

// WithTimeFormat sets the time format for time.Time fields
func WithTimeFormat(format string) MarshalOption

// WithTimeLocation sets the timezone for time parsing
func WithTimeLocation(loc *time.Location) MarshalOption
```

---

## 7. Struct Tags

### 7.1 Tag Syntax

```go
type Patient struct {
    // Basic field mapping
    ID        string    `hl7:"PID.3"`
    Name      string    `hl7:"PID.5"`

    // Component access
    LastName  string    `hl7:"PID.5.1"`
    FirstName string    `hl7:"PID.5.2"`

    // Repetition handling
    Addresses []Address `hl7:"PID.11"`

    // Time parsing with format
    DOB       time.Time `hl7:"PID.7,format=20060102"`

    // Optional fields
    SSN       string    `hl7:"PID.19,omitempty"`

    // Nested structs
    Visit     Visit     `hl7:"PV1"`
}

type Address struct {
    Street  string `hl7:".1"`  // Relative location
    City    string `hl7:".3"`
    State   string `hl7:".4"`
    Zip     string `hl7:".5"`
}
```

### 7.2 Supported Tag Options

| Option | Description | Example |
|--------|-------------|---------|
| `omitempty` | Skip if zero value | `hl7:"PID.19,omitempty"` |
| `format` | Time format | `hl7:"PID.7,format=20060102"` |
| `-` | Skip field | `hl7:"-"` |

---

## 8. MLLP Transport

### 8.1 Client Interface

```go
package mllp

// Client sends HL7 messages over MLLP
type Client interface {
    // Send sends a message and waits for ACK
    Send(ctx context.Context, msg hl7.Message) (hl7.Message, error)

    // SendAsync sends without waiting for response
    SendAsync(ctx context.Context, msg hl7.Message) error

    // Close closes the connection
    Close() error
}

type ClientOption func(*clientConfig)

func WithTimeout(d time.Duration) ClientOption
func WithRetry(attempts int, backoff time.Duration) ClientOption
func WithTLS(config *tls.Config) ClientOption
func WithKeepAlive(enable bool) ClientOption
```

### 8.2 Server Interface

```go
package mllp

// Handler processes incoming HL7 messages
type Handler interface {
    HandleMessage(ctx context.Context, msg hl7.Message) (hl7.Message, error)
}

// HandlerFunc is a function adapter for Handler
type HandlerFunc func(ctx context.Context, msg hl7.Message) (hl7.Message, error)

// Server accepts MLLP connections
type Server interface {
    // Serve starts accepting connections
    Serve(listener net.Listener) error

    // Shutdown gracefully stops the server
    Shutdown(ctx context.Context) error
}

type ServerOption func(*serverConfig)

func WithHandler(h Handler) ServerOption
func WithMaxConnections(max int) ServerOption
func WithReadTimeout(d time.Duration) ServerOption
func WithWriteTimeout(d time.Duration) ServerOption
func WithTLSConfig(config *tls.Config) ServerOption
```

---

## 9. Acknowledgment Handling

### 9.1 ACK Types

```go
package ack

type AckCode string

const (
    ApplicationAccept       AckCode = "AA"
    ApplicationError        AckCode = "AE"
    ApplicationReject       AckCode = "AR"
    CommitAccept            AckCode = "CA"
    CommitError             AckCode = "CE"
    CommitReject            AckCode = "CR"
)

// ACK represents an acknowledgment message structure
type ACK struct {
    Code          AckCode
    ControlID     string  // Original message control ID
    TextMessage   string  // Optional text message
    ErrorCode     string  // ERR segment error code
    ErrorLocation string  // Location of error
}
```

### 9.2 ACK Builder

```go
package ack

// Builder creates ACK messages
type Builder interface {
    // Accept creates an AA acknowledgment
    Accept(original hl7.Message) (hl7.Message, error)

    // Reject creates an AR acknowledgment
    Reject(original hl7.Message, reason string) (hl7.Message, error)

    // Error creates an AE acknowledgment
    Error(original hl7.Message, err error) (hl7.Message, error)

    // Custom creates a custom acknowledgment
    Custom(original hl7.Message, ack ACK) (hl7.Message, error)
}
```

---

## 10. Validation Rules

### 10.1 Rule Types

```go
package validate

// Rule defines a validation rule
type Rule interface {
    // Validate checks the rule against a message
    Validate(msg hl7.Message) []ValidationError

    // Location returns the rule's target location
    Location() string

    // Description returns a human-readable description
    Description() string
}

// RuleBuilder provides fluent rule construction
type RuleBuilder interface {
    // Required marks the location as required
    Required() RuleBuilder

    // Value checks for a specific value
    Value(expected string) RuleBuilder

    // Pattern checks against a regex pattern
    Pattern(pattern string) RuleBuilder

    // Length checks value length
    Length(min, max int) RuleBuilder

    // OneOf checks value is in allowed set
    OneOf(values ...string) RuleBuilder

    // Custom adds a custom validation function
    Custom(fn func(value string) error) RuleBuilder

    // Build creates the rule
    Build() Rule
}

// At starts building a rule for a location
func At(location string) RuleBuilder
```

### 10.2 Pre-built Rulesets

```go
package validate

// RuleSet is a collection of validation rules
type RuleSet interface {
    Rules() []Rule
    Add(rules ...Rule) RuleSet
    Merge(other RuleSet) RuleSet
}

// Pre-built rulesets for common message types
func MSHRules() RuleSet          // Standard MSH requirements
func PIDRules() RuleSet          // Patient identification
func PV1Rules() RuleSet          // Patient visit
func ORMRules() RuleSet          // Order messages
func ORURules() RuleSet          // Observation results
func ADTRules() RuleSet          // ADT messages
```

---

## 11. Escape Sequence Handling

### 11.1 Standard Escape Sequences

| Sequence | Meaning | Character |
|----------|---------|-----------|
| `\F\` | Field separator | `\|` |
| `\S\` | Component separator | `^` |
| `\T\` | Subcomponent separator | `&` |
| `\R\` | Repetition separator | `~` |
| `\E\` | Escape character | `\` |
| `\Xdd...\` | Hex encoded data | varies |
| `\.br\` | Line break | `\n` |

### 11.2 Escape Interface

```go
package hl7

// Escaper handles escape sequence encoding/decoding
type Escaper interface {
    // Escape encodes special characters
    Escape(value string, delims *Delimiters) string

    // Unescape decodes escape sequences
    Unescape(value string, delims *Delimiters) string
}
```

---

## 12. Error Handling Strategy

### 12.1 Error Categories

1. **Parse Errors**: Invalid message format, unexpected characters
2. **Location Errors**: Invalid location syntax, out-of-bounds access
3. **Validation Errors**: Failed validation rules
4. **Marshal Errors**: Type conversion failures, missing required fields
5. **Transport Errors**: Connection failures, timeouts

### 12.2 Error Wrapping

All errors support `errors.Is()` and `errors.As()` for type checking:

```go
msg, err := parser.Parse(data)
if err != nil {
    var parseErr *hl7.ParseError
    if errors.As(err, &parseErr) {
        log.Printf("Parse error at line %d: %s", parseErr.Line, parseErr.Message)
    }
}
```

---

## 13. Performance Considerations

### 13.1 Memory Management
- Reusable parser/encoder instances
- Buffer pooling for large messages
- Lazy parsing of nested structures
- Zero-copy string handling where possible

### 13.2 Concurrency
- Thread-safe parsing (stateless parsers)
- Connection pooling in MLLP client
- Worker pool for server message handling

### 13.3 Benchmarks (Target Metrics)
| Operation | Target | Notes |
|-----------|--------|-------|
| Parse small message (<1KB) | <10μs | Common ADT messages |
| Parse large message (>100KB) | <1ms | ORU with many OBX |
| Encode message | <5μs | |
| Unmarshal to struct | <20μs | Reflection overhead |
| Validate message | <50μs | Depends on ruleset |

---

## 14. Testing Strategy

### 14.1 Test Categories
- **Unit tests**: Each package has comprehensive unit tests
- **Integration tests**: Cross-package functionality
- **Fuzz tests**: Parser robustness testing
- **Benchmark tests**: Performance regression detection
- **Conformance tests**: HL7 specification compliance

### 14.2 Test Data
- Sample messages for each message type
- Edge cases (empty fields, max lengths, Unicode)
- Real-world anonymized messages
- Malformed messages for error handling

---

## 15. Compatibility

### 15.1 HL7 Version Support
- Primary: HL7 v2.4, v2.5, v2.5.1
- Secondary: HL7 v2.3, v2.3.1
- Experimental: HL7 v2.6, v2.7

### 15.2 Go Version Support
- Minimum: Go 1.21
- Tested: Go 1.21, 1.22, 1.23

### 15.3 Platform Support
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

---

## 16. API Stability

### 16.1 Semantic Versioning
- Major version: Breaking API changes
- Minor version: New features, backward compatible
- Patch version: Bug fixes

### 16.2 Deprecation Policy
- Deprecated APIs marked with `// Deprecated:` comments
- Minimum 2 minor versions before removal
- Migration guides provided

---

## 17. Documentation Requirements

### 17.1 Code Documentation
- All exported types, functions, and methods documented
- Examples for common use cases
- Package-level documentation with overview

### 17.2 External Documentation
- Quick start guide
- API reference (generated from code)
- Message type guides
- Migration guide from v1
