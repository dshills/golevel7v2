package validate

import (
	"github.com/dshills/golevel7/hl7"
)

// ValidationResult represents the outcome of validating an HL7 message.
type ValidationResult interface {
	// Valid returns true if no validation errors occurred.
	Valid() bool
	// Errors returns all validation errors encountered.
	Errors() []ValidationError
	// Warnings returns all validation warnings encountered.
	Warnings() []ValidationWarning
}

// Validator validates HL7 messages against a set of rules.
type Validator interface {
	// Validate applies all rules to the message and returns the result.
	Validate(msg hl7.Message) ValidationResult
	// ValidateSegment validates a specific segment against applicable rules.
	ValidateSegment(seg hl7.Segment) ValidationResult
}

// validationResult is the concrete implementation of ValidationResult.
type validationResult struct {
	errors   []ValidationError
	warnings []ValidationWarning
}

// Valid returns true if no validation errors occurred.
func (r *validationResult) Valid() bool {
	return len(r.errors) == 0
}

// Errors returns all validation errors encountered.
func (r *validationResult) Errors() []ValidationError {
	if r.errors == nil {
		return []ValidationError{}
	}
	// Return a copy to prevent external modification
	result := make([]ValidationError, len(r.errors))
	copy(result, r.errors)
	return result
}

// Warnings returns all validation warnings encountered.
func (r *validationResult) Warnings() []ValidationWarning {
	if r.warnings == nil {
		return []ValidationWarning{}
	}
	// Return a copy to prevent external modification
	result := make([]ValidationWarning, len(r.warnings))
	copy(result, r.warnings)
	return result
}

// validator is the concrete implementation of Validator.
type validator struct {
	rules []Rule
}

// New creates a new Validator with the specified rules.
func New(rules ...Rule) Validator {
	return &validator{
		rules: rules,
	}
}

// NewWithRuleSet creates a new Validator from a RuleSet.
func NewWithRuleSet(rs RuleSet) Validator {
	return &validator{
		rules: rs.Rules(),
	}
}

// Validate applies all rules to the message and returns the result.
func (v *validator) Validate(msg hl7.Message) ValidationResult {
	result := &validationResult{
		errors:   make([]ValidationError, 0),
		warnings: make([]ValidationWarning, 0),
	}

	if msg == nil {
		result.errors = append(result.errors, ValidationError{
			Rule:    "validator",
			Message: "message is nil",
		})
		return result
	}

	for _, rule := range v.rules {
		if errs := rule.Validate(msg); len(errs) > 0 {
			result.errors = append(result.errors, errs...)
		}
	}

	return result
}

// ValidateSegment validates a specific segment against applicable rules.
// Only rules whose location starts with the segment name will be applied.
func (v *validator) ValidateSegment(seg hl7.Segment) ValidationResult {
	result := &validationResult{
		errors:   make([]ValidationError, 0),
		warnings: make([]ValidationWarning, 0),
	}

	if seg == nil {
		result.errors = append(result.errors, ValidationError{
			Rule:    "validator",
			Message: "segment is nil",
		})
		return result
	}

	segName := seg.Name()

	// Create a wrapper that allows rules to query just this segment
	wrapper := &segmentWrapper{seg: seg}

	for _, rule := range v.rules {
		loc := rule.Location()
		// Check if this rule applies to the segment
		if len(loc) >= len(segName) && loc[:len(segName)] == segName {
			// Check for exact match or continuation with dot
			if len(loc) == len(segName) || loc[len(segName)] == '.' || loc[len(segName)] == '[' {
				if errs := rule.Validate(wrapper); len(errs) > 0 {
					result.errors = append(result.errors, errs...)
				}
			}
		}
	}

	return result
}

// segmentWrapper wraps a Segment to implement the Message interface subset
// needed for validation. This allows segment-level validation using the same
// rule implementations.
type segmentWrapper struct {
	seg hl7.Segment
}

// Get implements the Get method needed by rules.
func (w *segmentWrapper) Get(location string) (string, error) {
	return w.seg.Get(location)
}

// GetAll implements the GetAll method for completeness.
func (w *segmentWrapper) GetAll(location string) ([]string, error) {
	return w.seg.GetAll(location)
}

// Segment returns the wrapped segment if the name matches.
func (w *segmentWrapper) Segment(name string) (hl7.Segment, bool) {
	if w.seg.Name() == name {
		return w.seg, true
	}
	return nil, false
}

// Segments returns the wrapped segment if the name matches.
func (w *segmentWrapper) Segments(name string) []hl7.Segment {
	if w.seg.Name() == name {
		return []hl7.Segment{w.seg}
	}
	return nil
}

// AllSegments returns just the wrapped segment.
func (w *segmentWrapper) AllSegments() []hl7.Segment {
	return []hl7.Segment{w.seg}
}

// Set is not supported for segment wrapper.
func (w *segmentWrapper) Set(_ string, _ string) error {
	return nil
}

// GetAt implements structured query.
func (w *segmentWrapper) GetAt(loc *hl7.Location) (string, error) {
	return w.seg.Get(loc.String())
}

// GetAllAt implements structured query.
func (w *segmentWrapper) GetAllAt(loc *hl7.Location) ([]string, error) {
	return w.seg.GetAll(loc.String())
}

// SetAt is not supported for segment wrapper.
func (w *segmentWrapper) SetAt(_ *hl7.Location, _ string) error {
	return nil
}

// AddSegment is not supported for segment wrapper.
func (w *segmentWrapper) AddSegment(_ hl7.Segment) error {
	return nil
}

// InsertSegment is not supported for segment wrapper.
func (w *segmentWrapper) InsertSegment(_ int, _ hl7.Segment) error {
	return nil
}

// RemoveSegment is not supported for segment wrapper.
func (w *segmentWrapper) RemoveSegment(_ string) bool {
	return false
}

// Bytes returns the segment bytes.
func (w *segmentWrapper) Bytes() []byte {
	return w.seg.Bytes(nil)
}

// String returns the segment string.
func (w *segmentWrapper) String() string {
	return w.seg.String()
}

// Type returns empty string as this is just a segment wrapper.
func (w *segmentWrapper) Type() string {
	return ""
}

// ControlID returns empty string as this is just a segment wrapper.
func (w *segmentWrapper) ControlID() string {
	return ""
}

// Version returns empty string as this is just a segment wrapper.
func (w *segmentWrapper) Version() string {
	return ""
}

// Delimiters returns nil as this is just a segment wrapper.
func (w *segmentWrapper) Delimiters() *hl7.Delimiters {
	return nil
}

// Ensure segmentWrapper implements the Message interface methods needed by rules.
var _ hl7.Message = (*segmentWrapper)(nil)
