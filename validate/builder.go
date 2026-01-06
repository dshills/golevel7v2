package validate

import (
	"regexp"

	"github.com/dshills/golevel7/hl7"
)

// RuleBuilder provides a fluent interface for constructing validation rules.
type RuleBuilder interface {
	// Required adds a requirement that the field must be present and non-empty.
	Required() RuleBuilder
	// Value adds a requirement that the field must have an exact value.
	Value(expected string) RuleBuilder
	// Pattern adds a requirement that the field must match a regular expression.
	Pattern(pattern string) RuleBuilder
	// Length adds a requirement that the field length must be within bounds.
	// Use 0 for minLen or maxLen to indicate no bound on that side.
	Length(minLen, maxLen int) RuleBuilder
	// OneOf adds a requirement that the field value must be one of the allowed values.
	OneOf(values ...string) RuleBuilder
	// Custom adds a custom validation function.
	Custom(fn func(value string) error) RuleBuilder
	// WithDescription sets a custom description for the rule.
	WithDescription(desc string) RuleBuilder
	// Build constructs the final Rule from the builder configuration.
	Build() Rule
}

// ruleBuilder is the concrete implementation of RuleBuilder.
type ruleBuilder struct {
	location    string
	description string
	rules       []Rule
}

// At creates a new RuleBuilder for the specified HL7 location.
// The location follows HL7 path notation (e.g., "MSH.9", "PID.3.1").
func At(location string) RuleBuilder {
	return &ruleBuilder{
		location: location,
		rules:    make([]Rule, 0),
	}
}

// Required adds a requirement that the field must be present and non-empty.
func (b *ruleBuilder) Required() RuleBuilder {
	b.rules = append(b.rules, &requiredRule{
		location: b.location,
	})
	return b
}

// Value adds a requirement that the field must have an exact value.
func (b *ruleBuilder) Value(expected string) RuleBuilder {
	b.rules = append(b.rules, &valueRule{
		location: b.location,
		expected: expected,
	})
	return b
}

// Pattern adds a requirement that the field must match a regular expression.
// If the pattern is invalid, the rule will always fail with a pattern error.
func (b *ruleBuilder) Pattern(pattern string) RuleBuilder {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// Store a rule that will always fail with the compile error
		b.rules = append(b.rules, &invalidPatternRule{
			location: b.location,
			pattern:  pattern,
			err:      err,
		})
		return b
	}
	b.rules = append(b.rules, &patternRule{
		location: b.location,
		pattern:  compiled,
	})
	return b
}

// Length adds a requirement that the field length must be within bounds.
// Use 0 for minLen or maxLen to indicate no bound on that side.
func (b *ruleBuilder) Length(minLen, maxLen int) RuleBuilder {
	b.rules = append(b.rules, &lengthRule{
		location: b.location,
		min:      minLen,
		max:      maxLen,
	})
	return b
}

// OneOf adds a requirement that the field value must be one of the allowed values.
func (b *ruleBuilder) OneOf(values ...string) RuleBuilder {
	b.rules = append(b.rules, &oneOfRule{
		location: b.location,
		allowed:  values,
	})
	return b
}

// Custom adds a custom validation function.
func (b *ruleBuilder) Custom(fn func(value string) error) RuleBuilder {
	b.rules = append(b.rules, &customRule{
		location: b.location,
		fn:       fn,
	})
	return b
}

// WithDescription sets a custom description for the rule.
func (b *ruleBuilder) WithDescription(desc string) RuleBuilder {
	b.description = desc
	return b
}

// Build constructs the final Rule from the builder configuration.
// If no rules were added, returns a no-op rule that always passes.
// If only one rule was added, returns that rule directly.
// If multiple rules were added, returns a composite rule.
func (b *ruleBuilder) Build() Rule {
	if len(b.rules) == 0 {
		return &noopRule{
			location:    b.location,
			description: b.description,
		}
	}

	// Apply description to rules if set
	if b.description != "" {
		for _, rule := range b.rules {
			switch r := rule.(type) {
			case *requiredRule:
				r.description = b.description
			case *valueRule:
				r.description = b.description
			case *patternRule:
				r.description = b.description
			case *lengthRule:
				r.description = b.description
			case *oneOfRule:
				r.description = b.description
			case *customRule:
				r.description = b.description
			case *invalidPatternRule:
				r.description = b.description
			}
		}
	}

	if len(b.rules) == 1 {
		return b.rules[0]
	}

	return &compositeRule{
		location:    b.location,
		rules:       b.rules,
		description: b.description,
	}
}

// noopRule is a rule that always passes validation.
type noopRule struct {
	location    string
	description string
}

func (r *noopRule) Validate(_ hl7.Message) []ValidationError {
	return nil
}

func (r *noopRule) Location() string {
	return r.location
}

func (r *noopRule) Description() string {
	if r.description != "" {
		return r.description
	}
	return "no validation"
}

// invalidPatternRule is a rule that always fails because the pattern was invalid.
type invalidPatternRule struct {
	location    string
	pattern     string
	err         error
	description string
}

func (r *invalidPatternRule) Validate(_ hl7.Message) []ValidationError {
	return []ValidationError{{
		Location: r.location,
		Rule:     "pattern",
		Message:  "invalid pattern: " + r.err.Error(),
		Expected: r.pattern,
	}}
}

func (r *invalidPatternRule) Location() string {
	return r.location
}

func (r *invalidPatternRule) Description() string {
	if r.description != "" {
		return r.description
	}
	return "invalid pattern rule"
}
