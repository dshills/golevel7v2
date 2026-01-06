// Package validate provides validation rules and validators for HL7 v2.x messages.
//
// The validate package enables comprehensive validation of HL7 messages against
// configurable rules. It supports required field checking, value constraints,
// pattern matching, length validation, and custom validation functions.
//
// # Basic Usage
//
// Create a validator with rules and validate a message:
//
//	v := validate.NewValidator(
//	    validate.Required("MSH.9"),    // Message type required
//	    validate.Required("MSH.10"),   // Control ID required
//	    validate.Required("PID.3.1"),  // Patient ID required
//	)
//
//	errors := v.Validate(msg)
//	if len(errors) > 0 {
//	    for _, err := range errors {
//	        log.Printf("Validation error: %v", err)
//	    }
//	}
//
// # Built-in Validation Rules
//
// The package provides several built-in rule types:
//
// Required - Ensures a field is present and non-empty:
//
//	validate.Required("PID.3.1")
//	validate.RequiredWithDesc("PID.5", "Patient name is required")
//
// Value - Ensures a field has a specific value:
//
//	validate.Value("MSH.9.1", "ADT")     // Message type must be ADT
//	validate.Value("MSH.11", "P")        // Processing ID must be Production
//
// Pattern - Validates against a regular expression:
//
//	// Date format: YYYYMMDD
//	validate.Pattern("PID.7", `^\d{8}$`)
//
//	// Phone number format
//	validate.Pattern("PID.13", `^\(\d{3}\)\d{3}-\d{4}$`)
//
// Length - Validates field length:
//
//	validate.Length("PID.3.1", 1, 20)    // ID between 1-20 chars
//	validate.MinLength("PID.5", 1)       // Name at least 1 char
//	validate.MaxLength("NTE.3", 65536)   // Note max 64KB
//
// OneOf - Validates against a list of allowed values:
//
//	validate.OneOf("PID.8", "M", "F", "O", "U")  // Gender codes
//	validate.OneOf("MSH.11", "P", "T", "D")      // Processing IDs
//
// Custom - Validates with a custom function:
//
//	validate.Custom("PID.7", func(value string) error {
//	    _, err := time.Parse("20060102", value)
//	    if err != nil {
//	        return fmt.Errorf("invalid date format")
//	    }
//	    return nil
//	})
//
// # Combining Rules
//
// Combine multiple rules for a field:
//
//	// Patient ID must be present, 1-20 chars, alphanumeric
//	v := validate.NewValidator(
//	    validate.Required("PID.3.1"),
//	    validate.Length("PID.3.1", 1, 20),
//	    validate.Pattern("PID.3.1", `^[A-Z0-9]+$`),
//	)
//
// Use composite rules for complex validation:
//
//	// All rules must pass
//	patientIDRules := validate.All("PID.3.1",
//	    validate.Required("PID.3.1"),
//	    validate.Length("PID.3.1", 1, 20),
//	    validate.Pattern("PID.3.1", `^[A-Z0-9]+$`),
//	)
//
// # Message Type Specific Validation
//
// Create validators for specific message types:
//
//	// ADT^A01 (Admit) validator
//	adtA01Validator := validate.NewValidator(
//	    // MSH requirements
//	    validate.Required("MSH.9"),
//	    validate.Value("MSH.9.1", "ADT"),
//	    validate.Value("MSH.9.2", "A01"),
//
//	    // PID requirements
//	    validate.Required("PID.3.1"),   // Patient ID
//	    validate.Required("PID.5"),     // Patient Name
//	    validate.Required("PID.7"),     // DOB
//	    validate.OneOf("PID.8", "M", "F", "O", "U"),
//
//	    // PV1 requirements for admit
//	    validate.Required("PV1.2"),     // Patient class
//	    validate.Required("PV1.3"),     // Assigned location
//	    validate.Required("PV1.44"),    // Admit date/time
//	)
//
// # Validation Results
//
// Validation errors contain detailed information:
//
//	errors := v.Validate(msg)
//	for _, err := range errors {
//	    fmt.Printf("Location: %s\n", err.Location)
//	    fmt.Printf("Rule: %s\n", err.Rule)
//	    fmt.Printf("Message: %s\n", err.Message)
//	    if err.Expected != "" {
//	        fmt.Printf("Expected: %s\n", err.Expected)
//	    }
//	    if err.Actual != "" {
//	        fmt.Printf("Actual: %s\n", err.Actual)
//	    }
//	}
//
// # Creating Custom Rules
//
// Implement the Rule interface for custom validation logic:
//
//	type Rule interface {
//	    Validate(msg hl7.Message) []ValidationError
//	    Location() string
//	    Description() string
//	}
//
// Example custom rule:
//
//	type dateRangeRule struct {
//	    location string
//	    min, max time.Time
//	}
//
//	func (r *dateRangeRule) Validate(msg hl7.Message) []ValidationError {
//	    value, err := msg.Get(r.location)
//	    if err != nil || value == "" {
//	        return nil // Let required rule handle presence
//	    }
//
//	    date, err := time.Parse("20060102", value)
//	    if err != nil {
//	        return []ValidationError{{
//	            Location: r.location,
//	            Rule:     "dateRange",
//	            Message:  "invalid date format",
//	        }}
//	    }
//
//	    if date.Before(r.min) || date.After(r.max) {
//	        return []ValidationError{{
//	            Location: r.location,
//	            Rule:     "dateRange",
//	            Message:  "date out of range",
//	            Expected: fmt.Sprintf("%s to %s",
//	                r.min.Format("2006-01-02"),
//	                r.max.Format("2006-01-02")),
//	            Actual:   date.Format("2006-01-02"),
//	        }}
//	    }
//
//	    return nil
//	}
//
// # Example: ORU Message Validation
//
//	// Validator for ORU^R01 (Lab Results)
//	oruValidator := validate.NewValidator(
//	    // Message header
//	    validate.Required("MSH.9"),
//	    validate.Value("MSH.9.1", "ORU"),
//	    validate.Value("MSH.9.2", "R01"),
//
//	    // Patient identification
//	    validate.Required("PID.3.1"),
//	    validate.Required("PID.5"),
//
//	    // Order information
//	    validate.Required("OBR.4"),     // Universal service ID
//	    validate.Required("OBR.7"),     // Observation date/time
//
//	    // Each OBX needs these fields
//	    validate.Required("OBX.3"),     // Observation identifier
//	    validate.Required("OBX.5"),     // Observation value
//	    validate.OneOf("OBX.11", "F", "C", "P", "I"), // Result status
//	)
//
//	// Validate incoming lab result
//	msg, _ := parse.New().Parse(labData)
//	if errors := oruValidator.Validate(msg); len(errors) > 0 {
//	    return fmt.Errorf("invalid ORU message: %d validation errors", len(errors))
//	}
package validate
