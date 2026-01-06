// Package validate provides validation rules and validators for HL7 v2.x messages.
package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dshills/golevel7/hl7"
)

// Rule defines a validation rule that can be applied to an HL7 message.
type Rule interface {
	// Validate applies this rule to the message and returns any validation errors.
	Validate(msg hl7.Message) []ValidationError
	// Location returns the HL7 path this rule applies to (e.g., "MSH.9").
	Location() string
	// Description returns a human-readable description of what this rule validates.
	Description() string
}

// ValidationError represents a validation failure.
type ValidationError struct {
	// Location is the HL7 path where validation failed.
	Location string
	// Rule is the name/type of the validation rule that failed.
	Rule string
	// Message describes what went wrong.
	Message string
	// Expected describes what was expected (optional).
	Expected string
	// Actual describes what was found (optional).
	Actual string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString("validation error")

	if e.Location != "" {
		sb.WriteString(" at ")
		sb.WriteString(e.Location)
	}

	if e.Rule != "" {
		sb.WriteString(" [")
		sb.WriteString(e.Rule)
		sb.WriteString("]")
	}

	if e.Message != "" {
		sb.WriteString(": ")
		sb.WriteString(e.Message)
	}

	switch {
	case e.Expected != "" && e.Actual != "":
		sb.WriteString(fmt.Sprintf(" (expected %s, got %s)", e.Expected, e.Actual))
	case e.Expected != "":
		sb.WriteString(fmt.Sprintf(" (expected %s)", e.Expected))
	case e.Actual != "":
		sb.WriteString(fmt.Sprintf(" (got %s)", e.Actual))
	}

	return sb.String()
}

// ValidationWarning represents a non-critical validation issue.
type ValidationWarning struct {
	// Location is the HL7 path where the warning was raised.
	Location string
	// Rule is the name/type of the validation rule that raised the warning.
	Rule string
	// Message describes the warning.
	Message string
}

// String returns a human-readable representation of the warning.
func (w ValidationWarning) String() string {
	var sb strings.Builder
	sb.WriteString("warning")

	if w.Location != "" {
		sb.WriteString(" at ")
		sb.WriteString(w.Location)
	}

	if w.Rule != "" {
		sb.WriteString(" [")
		sb.WriteString(w.Rule)
		sb.WriteString("]")
	}

	if w.Message != "" {
		sb.WriteString(": ")
		sb.WriteString(w.Message)
	}

	return sb.String()
}

// requiredRule validates that a field is present and non-empty.
type requiredRule struct {
	location    string
	description string
}

// Validate checks that the location exists and has a non-empty value.
func (r *requiredRule) Validate(msg hl7.Message) []ValidationError {
	if msg == nil {
		return []ValidationError{{
			Location: r.location,
			Rule:     "required",
			Message:  "message is nil",
		}}
	}

	value, err := msg.Get(r.location)
	if err != nil {
		return []ValidationError{{
			Location: r.location,
			Rule:     "required",
			Message:  fmt.Sprintf("field not found: %v", err),
		}}
	}

	if strings.TrimSpace(value) == "" {
		return []ValidationError{{
			Location: r.location,
			Rule:     "required",
			Message:  "field is required but empty",
		}}
	}

	return nil
}

// Location returns the HL7 path this rule applies to.
func (r *requiredRule) Location() string {
	return r.location
}

// Description returns a human-readable description of this rule.
func (r *requiredRule) Description() string {
	if r.description != "" {
		return r.description
	}
	return fmt.Sprintf("%s is required", r.location)
}

// valueRule validates that a field has an exact expected value.
type valueRule struct {
	location    string
	expected    string
	description string
}

// Validate checks that the location has the expected value.
func (r *valueRule) Validate(msg hl7.Message) []ValidationError {
	if msg == nil {
		return []ValidationError{{
			Location: r.location,
			Rule:     "value",
			Message:  "message is nil",
		}}
	}

	value, err := msg.Get(r.location)
	if err != nil {
		return []ValidationError{{
			Location: r.location,
			Rule:     "value",
			Message:  fmt.Sprintf("field not found: %v", err),
			Expected: r.expected,
		}}
	}

	if value != r.expected {
		return []ValidationError{{
			Location: r.location,
			Rule:     "value",
			Message:  "field value does not match expected",
			Expected: r.expected,
			Actual:   value,
		}}
	}

	return nil
}

// Location returns the HL7 path this rule applies to.
func (r *valueRule) Location() string {
	return r.location
}

// Description returns a human-readable description of this rule.
func (r *valueRule) Description() string {
	if r.description != "" {
		return r.description
	}
	return fmt.Sprintf("%s must equal %q", r.location, r.expected)
}

// patternRule validates that a field matches a regular expression pattern.
type patternRule struct {
	location    string
	pattern     *regexp.Regexp
	description string
}

// Validate checks that the location value matches the pattern.
func (r *patternRule) Validate(msg hl7.Message) []ValidationError {
	if msg == nil {
		return []ValidationError{{
			Location: r.location,
			Rule:     "pattern",
			Message:  "message is nil",
		}}
	}

	value, err := msg.Get(r.location)
	if err != nil {
		// If field doesn't exist, pattern validation passes (use required rule for presence)
		return nil
	}

	// Empty values pass pattern validation (use required rule for presence)
	if value == "" {
		return nil
	}

	if !r.pattern.MatchString(value) {
		return []ValidationError{{
			Location: r.location,
			Rule:     "pattern",
			Message:  "field value does not match pattern",
			Expected: r.pattern.String(),
			Actual:   value,
		}}
	}

	return nil
}

// Location returns the HL7 path this rule applies to.
func (r *patternRule) Location() string {
	return r.location
}

// Description returns a human-readable description of this rule.
func (r *patternRule) Description() string {
	if r.description != "" {
		return r.description
	}
	return fmt.Sprintf("%s must match pattern %q", r.location, r.pattern.String())
}

// lengthRule validates that a field value length is within bounds.
type lengthRule struct {
	location    string
	min         int
	max         int
	description string
}

// Validate checks that the location value length is within min and max bounds.
func (r *lengthRule) Validate(msg hl7.Message) []ValidationError {
	if msg == nil {
		return []ValidationError{{
			Location: r.location,
			Rule:     "length",
			Message:  "message is nil",
		}}
	}

	value, err := msg.Get(r.location)
	if err != nil {
		// If field doesn't exist, length validation passes (use required rule for presence)
		return nil
	}

	length := len(value)

	if r.min > 0 && length < r.min {
		return []ValidationError{{
			Location: r.location,
			Rule:     "length",
			Message:  fmt.Sprintf("field length %d is less than minimum %d", length, r.min),
			Expected: fmt.Sprintf("minimum %d characters", r.min),
			Actual:   fmt.Sprintf("%d characters", length),
		}}
	}

	if r.max > 0 && length > r.max {
		return []ValidationError{{
			Location: r.location,
			Rule:     "length",
			Message:  fmt.Sprintf("field length %d exceeds maximum %d", length, r.max),
			Expected: fmt.Sprintf("maximum %d characters", r.max),
			Actual:   fmt.Sprintf("%d characters", length),
		}}
	}

	return nil
}

// Location returns the HL7 path this rule applies to.
func (r *lengthRule) Location() string {
	return r.location
}

// Description returns a human-readable description of this rule.
func (r *lengthRule) Description() string {
	if r.description != "" {
		return r.description
	}
	if r.min > 0 && r.max > 0 {
		return fmt.Sprintf("%s length must be between %d and %d", r.location, r.min, r.max)
	}
	if r.min > 0 {
		return fmt.Sprintf("%s length must be at least %d", r.location, r.min)
	}
	if r.max > 0 {
		return fmt.Sprintf("%s length must be at most %d", r.location, r.max)
	}
	return fmt.Sprintf("%s length validation", r.location)
}

// oneOfRule validates that a field value is one of the allowed values.
type oneOfRule struct {
	location    string
	allowed     []string
	description string
}

// Validate checks that the location value is in the allowed list.
func (r *oneOfRule) Validate(msg hl7.Message) []ValidationError {
	if msg == nil {
		return []ValidationError{{
			Location: r.location,
			Rule:     "oneOf",
			Message:  "message is nil",
		}}
	}

	value, err := msg.Get(r.location)
	if err != nil {
		// If field doesn't exist, oneOf validation passes (use required rule for presence)
		return nil
	}

	// Empty values pass oneOf validation (use required rule for presence)
	if value == "" {
		return nil
	}

	for _, allowed := range r.allowed {
		if value == allowed {
			return nil
		}
	}

	return []ValidationError{{
		Location: r.location,
		Rule:     "oneOf",
		Message:  "field value is not in allowed list",
		Expected: fmt.Sprintf("one of [%s]", strings.Join(r.allowed, ", ")),
		Actual:   value,
	}}
}

// Location returns the HL7 path this rule applies to.
func (r *oneOfRule) Location() string {
	return r.location
}

// Description returns a human-readable description of this rule.
func (r *oneOfRule) Description() string {
	if r.description != "" {
		return r.description
	}
	return fmt.Sprintf("%s must be one of [%s]", r.location, strings.Join(r.allowed, ", "))
}

// customRule validates a field using a custom validation function.
type customRule struct {
	location    string
	fn          func(string) error
	description string
}

// Validate applies the custom validation function to the field value.
func (r *customRule) Validate(msg hl7.Message) []ValidationError {
	if msg == nil {
		return []ValidationError{{
			Location: r.location,
			Rule:     "custom",
			Message:  "message is nil",
		}}
	}

	value, err := msg.Get(r.location)
	if err != nil {
		// If field doesn't exist, custom validation passes (use required rule for presence)
		return nil
	}

	if validationErr := r.fn(value); validationErr != nil {
		return []ValidationError{{
			Location: r.location,
			Rule:     "custom",
			Message:  validationErr.Error(),
			Actual:   value,
		}}
	}

	return nil
}

// Location returns the HL7 path this rule applies to.
func (r *customRule) Location() string {
	return r.location
}

// Description returns a human-readable description of this rule.
func (r *customRule) Description() string {
	if r.description != "" {
		return r.description
	}
	return fmt.Sprintf("%s custom validation", r.location)
}

// compositeRule combines multiple rules that all apply to the same location.
// All rules must pass for the composite to pass.
type compositeRule struct {
	location    string
	rules       []Rule
	description string
}

// Validate applies all contained rules and collects all errors.
func (r *compositeRule) Validate(msg hl7.Message) []ValidationError {
	var errors []ValidationError
	for _, rule := range r.rules {
		if errs := rule.Validate(msg); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}
	return errors
}

// Location returns the HL7 path this rule applies to.
func (r *compositeRule) Location() string {
	return r.location
}

// Description returns a human-readable description of this rule.
func (r *compositeRule) Description() string {
	if r.description != "" {
		return r.description
	}
	descriptions := make([]string, 0, len(r.rules))
	for _, rule := range r.rules {
		descriptions = append(descriptions, rule.Description())
	}
	return strings.Join(descriptions, "; ")
}
