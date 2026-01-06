package validate

import (
	"errors"
	"testing"
)

func TestAt(t *testing.T) {
	builder := At("MSH.9")
	if builder == nil {
		t.Fatal("At() returned nil")
	}
}

func TestRuleBuilder_Required(t *testing.T) {
	rule := At("MSH.9").Required().Build()

	if rule.Location() != "MSH.9" {
		t.Errorf("Location() = %q, want %q", rule.Location(), "MSH.9")
	}

	// Test with value present
	m := newMockMessage()
	m.setField("MSH.9", "ADT^A01")
	errs := rule.Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0", len(errs))
	}

	// Test with value absent
	m2 := newMockMessage()
	errs = rule.Validate(m2)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for missing field")
	}
}

func TestRuleBuilder_Value(t *testing.T) {
	rule := At("MSH.12").Value("2.5").Build()

	// Test with matching value
	m := newMockMessage()
	m.setField("MSH.12", "2.5")
	errs := rule.Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0", len(errs))
	}

	// Test with non-matching value
	m2 := newMockMessage()
	m2.setField("MSH.12", "2.4")
	errs = rule.Validate(m2)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for mismatched value")
	}
}

func TestRuleBuilder_Pattern(t *testing.T) {
	rule := At("PID.3").Pattern(`^\d{6}$`).Build()

	// Test with matching pattern
	m := newMockMessage()
	m.setField("PID.3", "123456")
	errs := rule.Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0", len(errs))
	}

	// Test with non-matching pattern
	m2 := newMockMessage()
	m2.setField("PID.3", "12345")
	errs = rule.Validate(m2)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for pattern mismatch")
	}
}

func TestRuleBuilder_InvalidPattern(t *testing.T) {
	rule := At("PID.3").Pattern(`[invalid`).Build()

	// Invalid pattern should always fail
	m := newMockMessage()
	m.setField("PID.3", "anything")
	errs := rule.Validate(m)
	if len(errs) == 0 {
		t.Error("Validate() should fail for invalid pattern")
	}
	if len(errs) > 0 && errs[0].Rule != "pattern" {
		t.Errorf("Error rule = %q, want %q", errs[0].Rule, "pattern")
	}
}

func TestRuleBuilder_Length(t *testing.T) {
	rule := At("MSH.10").Length(5, 20).Build()

	// Test within bounds
	m := newMockMessage()
	m.setField("MSH.10", "12345678")
	errs := rule.Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0", len(errs))
	}

	// Test below minimum
	m2 := newMockMessage()
	m2.setField("MSH.10", "123")
	errs = rule.Validate(m2)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for length below minimum")
	}

	// Test above maximum
	m3 := newMockMessage()
	m3.setField("MSH.10", "123456789012345678901")
	errs = rule.Validate(m3)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for length above maximum")
	}
}

func TestRuleBuilder_OneOf(t *testing.T) {
	rule := At("PV1.2").OneOf("I", "O", "E", "P", "R").Build()

	// Test with allowed value
	m := newMockMessage()
	m.setField("PV1.2", "I")
	errs := rule.Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0", len(errs))
	}

	// Test with disallowed value
	m2 := newMockMessage()
	m2.setField("PV1.2", "X")
	errs = rule.Validate(m2)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for disallowed value")
	}
}

func TestRuleBuilder_Custom(t *testing.T) {
	validateDate := func(v string) error {
		if len(v) != 8 {
			return errors.New("date must be YYYYMMDD format")
		}
		return nil
	}

	rule := At("PID.7").Custom(validateDate).Build()

	// Test with valid format
	m := newMockMessage()
	m.setField("PID.7", "19850101")
	errs := rule.Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0", len(errs))
	}

	// Test with invalid format
	m2 := newMockMessage()
	m2.setField("PID.7", "1985-01-01")
	errs = rule.Validate(m2)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for invalid date format")
	}
}

func TestRuleBuilder_WithDescription(t *testing.T) {
	rule := At("MSH.9").Required().WithDescription("Message Type is mandatory").Build()

	desc := rule.Description()
	if desc != "Message Type is mandatory" {
		t.Errorf("Description() = %q, want %q", desc, "Message Type is mandatory")
	}
}

func TestRuleBuilder_MultipleRules(t *testing.T) {
	rule := At("MSH.10").
		Required().
		Length(1, 20).
		Pattern(`^[A-Z0-9]+$`).
		Build()

	// Test all rules pass
	m := newMockMessage()
	m.setField("MSH.10", "CTRL12345")
	errs := rule.Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0", len(errs))
	}

	// Test required fails
	m2 := newMockMessage()
	errs = rule.Validate(m2)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for missing required field")
	}

	// Test length fails
	m3 := newMockMessage()
	m3.setField("MSH.10", "A123456789012345678901")
	errs = rule.Validate(m3)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for length violation")
	}

	// Test pattern fails
	m4 := newMockMessage()
	m4.setField("MSH.10", "ctrl123")
	errs = rule.Validate(m4)
	if len(errs) == 0 {
		t.Error("Validate() returned 0 errors, want errors for pattern violation")
	}
}

func TestRuleBuilder_NoRules(t *testing.T) {
	rule := At("MSH.9").Build()

	// No-op rule should always pass
	m := newMockMessage()
	errs := rule.Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0 for no-op rule", len(errs))
	}

	if rule.Location() != "MSH.9" {
		t.Errorf("Location() = %q, want %q", rule.Location(), "MSH.9")
	}
}

func TestRuleBuilder_Chaining(t *testing.T) {
	// Test that all builder methods return the builder for chaining
	builder := At("MSH.9")

	// Chain all methods
	result := builder.
		Required().
		Value("ADT^A01").
		Pattern(`^ADT`).
		Length(1, 50).
		OneOf("ADT^A01", "ADT^A04").
		Custom(func(_ string) error { return nil }).
		WithDescription("Test rule")

	if result == nil {
		t.Fatal("Chaining returned nil")
	}

	rule := result.Build()
	if rule == nil {
		t.Fatal("Build() returned nil")
	}
}

func TestNoopRule(t *testing.T) {
	rule := &noopRule{location: "TEST.1"}

	// Should always return no errors
	m := newMockMessage()
	errs := rule.Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0", len(errs))
	}

	// Should work with nil message
	errs = rule.Validate(nil)
	if len(errs) != 0 {
		t.Errorf("Validate(nil) returned %d errors, want 0", len(errs))
	}

	if rule.Location() != "TEST.1" {
		t.Errorf("Location() = %q, want %q", rule.Location(), "TEST.1")
	}

	if rule.Description() != "no validation" {
		t.Errorf("Description() = %q, want %q", rule.Description(), "no validation")
	}
}

func TestInvalidPatternRule(t *testing.T) {
	rule := &invalidPatternRule{
		location: "TEST.1",
		pattern:  "[invalid",
		err:      errors.New("missing closing bracket"),
	}

	m := newMockMessage()
	m.setField("TEST.1", "value")
	errs := rule.Validate(m)
	if len(errs) != 1 {
		t.Errorf("Validate() returned %d errors, want 1", len(errs))
	}

	if errs[0].Rule != "pattern" {
		t.Errorf("Error rule = %q, want %q", errs[0].Rule, "pattern")
	}

	if rule.Location() != "TEST.1" {
		t.Errorf("Location() = %q, want %q", rule.Location(), "TEST.1")
	}
}
