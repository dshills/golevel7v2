package validate

import (
	"testing"
)

func TestNewRuleSet(t *testing.T) {
	// Empty ruleset
	rs := NewRuleSet()
	if rs == nil {
		t.Fatal("NewRuleSet() returned nil")
	}
	if len(rs.Rules()) != 0 {
		t.Errorf("NewRuleSet() with no args should have 0 rules, got %d", len(rs.Rules()))
	}

	// With rules
	rule1 := At("MSH.9").Required().Build()
	rule2 := At("MSH.10").Required().Build()
	rs2 := NewRuleSet(rule1, rule2)
	if len(rs2.Rules()) != 2 {
		t.Errorf("NewRuleSet() with 2 rules should have 2 rules, got %d", len(rs2.Rules()))
	}
}

func TestRuleSet_Rules(t *testing.T) {
	rule1 := At("MSH.9").Required().Build()
	rule2 := At("MSH.10").Required().Build()
	rs := NewRuleSet(rule1, rule2)

	rules := rs.Rules()
	if len(rules) != 2 {
		t.Errorf("Rules() = %d, want 2", len(rules))
	}

	// Verify returns a copy
	originalLen := len(rs.Rules())
	rules[0] = nil
	if len(rs.Rules()) != originalLen {
		t.Error("Rules() should return a copy, not the original slice")
	}
}

func TestRuleSet_Add(t *testing.T) {
	rs := NewRuleSet()

	// Add single rule
	rule1 := At("MSH.9").Required().Build()
	result := rs.Add(rule1)
	if result != rs {
		t.Error("Add() should return the same RuleSet for chaining")
	}
	if len(rs.Rules()) != 1 {
		t.Errorf("After Add(), Rules() = %d, want 1", len(rs.Rules()))
	}

	// Add multiple rules
	rule2 := At("MSH.10").Required().Build()
	rule3 := At("MSH.12").Required().Build()
	rs.Add(rule2, rule3)
	if len(rs.Rules()) != 3 {
		t.Errorf("After Add(), Rules() = %d, want 3", len(rs.Rules()))
	}
}

func TestRuleSet_Merge(t *testing.T) {
	rule1 := At("MSH.9").Required().Build()
	rule2 := At("MSH.10").Required().Build()
	rs1 := NewRuleSet(rule1)

	rule3 := At("PID.3").Required().Build()
	rs2 := NewRuleSet(rule2, rule3)

	// Merge creates new RuleSet
	merged := rs1.Merge(rs2)
	if len(merged.Rules()) != 3 {
		t.Errorf("Merge() = %d rules, want 3", len(merged.Rules()))
	}

	// Original sets unchanged
	if len(rs1.Rules()) != 1 {
		t.Errorf("Original rs1 changed after merge, got %d rules", len(rs1.Rules()))
	}
	if len(rs2.Rules()) != 2 {
		t.Errorf("Original rs2 changed after merge, got %d rules", len(rs2.Rules()))
	}
}

func TestRuleSet_MergeNil(t *testing.T) {
	rule1 := At("MSH.9").Required().Build()
	rs1 := NewRuleSet(rule1)

	// Merge with nil
	merged := rs1.Merge(nil)
	if len(merged.Rules()) != 1 {
		t.Errorf("Merge(nil) = %d rules, want 1", len(merged.Rules()))
	}
}

func TestMSHRules(t *testing.T) {
	rs := MSHRules()
	rules := rs.Rules()

	if len(rules) != 3 {
		t.Errorf("MSHRules() = %d rules, want 3", len(rules))
	}

	// Verify locations
	locations := make(map[string]bool)
	for _, rule := range rules {
		locations[rule.Location()] = true
	}

	expectedLocations := []string{"MSH.9", "MSH.10", "MSH.12"}
	for _, loc := range expectedLocations {
		if !locations[loc] {
			t.Errorf("MSHRules() missing rule for %s", loc)
		}
	}

	// Test validation
	m := newMockMessage()
	m.setField("MSH.9", "ADT^A01")
	m.setField("MSH.10", "12345")
	m.setField("MSH.12", "2.5")

	v := NewWithRuleSet(rs)
	result := v.Validate(m)
	if !result.Valid() {
		t.Errorf("MSHRules validation failed with valid message: %v", result.Errors())
	}
}

func TestPIDRules(t *testing.T) {
	rs := PIDRules()
	rules := rs.Rules()

	if len(rules) != 1 {
		t.Errorf("PIDRules() = %d rules, want 1", len(rules))
	}

	if rules[0].Location() != "PID.3" {
		t.Errorf("PIDRules() location = %s, want PID.3", rules[0].Location())
	}

	// Test validation
	m := newMockMessage()
	m.setField("PID.3", "123456")

	v := NewWithRuleSet(rs)
	result := v.Validate(m)
	if !result.Valid() {
		t.Errorf("PIDRules validation failed with valid message: %v", result.Errors())
	}
}

func TestPV1Rules(t *testing.T) {
	rs := PV1Rules()
	rules := rs.Rules()

	if len(rules) != 1 {
		t.Errorf("PV1Rules() = %d rules, want 1", len(rules))
	}

	if rules[0].Location() != "PV1.2" {
		t.Errorf("PV1Rules() location = %s, want PV1.2", rules[0].Location())
	}
}

func TestOBRRules(t *testing.T) {
	rs := OBRRules()
	rules := rs.Rules()

	if len(rules) != 1 {
		t.Errorf("OBRRules() = %d rules, want 1", len(rules))
	}

	if rules[0].Location() != "OBR.4" {
		t.Errorf("OBRRules() location = %s, want OBR.4", rules[0].Location())
	}
}

func TestOBXRules(t *testing.T) {
	rs := OBXRules()
	rules := rs.Rules()

	if len(rules) != 2 {
		t.Errorf("OBXRules() = %d rules, want 2", len(rules))
	}

	// Verify locations
	locations := make(map[string]bool)
	for _, rule := range rules {
		locations[rule.Location()] = true
	}

	if !locations["OBX.2"] {
		t.Error("OBXRules() missing rule for OBX.2")
	}
	if !locations["OBX.3"] {
		t.Error("OBXRules() missing rule for OBX.3")
	}
}

func TestADTRules(t *testing.T) {
	rs := ADTRules()
	rules := rs.Rules()

	// MSH (3) + PID (1) = 4 rules
	if len(rules) != 4 {
		t.Errorf("ADTRules() = %d rules, want 4", len(rules))
	}

	// Test validation with complete ADT message
	m := newMockMessage()
	m.setField("MSH.9", "ADT^A01")
	m.setField("MSH.10", "12345")
	m.setField("MSH.12", "2.5")
	m.setField("PID.3", "P123456")

	v := NewWithRuleSet(rs)
	result := v.Validate(m)
	if !result.Valid() {
		t.Errorf("ADTRules validation failed with valid message: %v", result.Errors())
	}
}

func TestORURules(t *testing.T) {
	rs := ORURules()
	rules := rs.Rules()

	// MSH (3) + PID (1) + OBR (1) + OBX (2) = 7 rules
	if len(rules) != 7 {
		t.Errorf("ORURules() = %d rules, want 7", len(rules))
	}

	// Test validation with complete ORU message
	m := newMockMessage()
	m.setField("MSH.9", "ORU^R01")
	m.setField("MSH.10", "12345")
	m.setField("MSH.12", "2.5")
	m.setField("PID.3", "P123456")
	m.setField("OBR.4", "CBC^Complete Blood Count")
	m.setField("OBX.2", "NM")
	m.setField("OBX.3", "WBC^White Blood Cell Count")

	v := NewWithRuleSet(rs)
	result := v.Validate(m)
	if !result.Valid() {
		t.Errorf("ORURules validation failed with valid message: %v", result.Errors())
	}
}

func TestORMRules(t *testing.T) {
	rs := ORMRules()
	rules := rs.Rules()

	// MSH (3) + PID (1) + OBR (1) = 5 rules
	if len(rules) != 5 {
		t.Errorf("ORMRules() = %d rules, want 5", len(rules))
	}
}

func TestStandardRules(t *testing.T) {
	rs := StandardRules()
	rules := rs.Rules()

	// Should be same as MSHRules
	if len(rules) != 3 {
		t.Errorf("StandardRules() = %d rules, want 3", len(rules))
	}
}

func TestRuleSetChaining(t *testing.T) {
	// Test chaining Add and Merge
	rs := NewRuleSet().
		Add(At("MSH.9").Required().Build()).
		Add(At("MSH.10").Required().Build())

	if len(rs.Rules()) != 2 {
		t.Errorf("Chained Add() = %d rules, want 2", len(rs.Rules()))
	}

	rs2 := NewRuleSet(At("PID.3").Required().Build())
	combined := rs.Merge(rs2)

	if len(combined.Rules()) != 3 {
		t.Errorf("Merge after chained Add() = %d rules, want 3", len(combined.Rules()))
	}
}

func TestRuleSetValidation(t *testing.T) {
	tests := []struct {
		name      string
		ruleSet   RuleSet
		setup     func(*mockMessage)
		wantValid bool
		wantCount int
	}{
		{
			name:    "MSH rules all valid",
			ruleSet: MSHRules(),
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "ADT^A01")
				m.setField("MSH.10", "12345")
				m.setField("MSH.12", "2.5")
			},
			wantValid: true,
			wantCount: 0,
		},
		{
			name:    "MSH rules missing message type",
			ruleSet: MSHRules(),
			setup: func(m *mockMessage) {
				m.setField("MSH.10", "12345")
				m.setField("MSH.12", "2.5")
			},
			wantValid: false,
			wantCount: 1,
		},
		{
			name:    "ADT rules missing patient ID",
			ruleSet: ADTRules(),
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "ADT^A01")
				m.setField("MSH.10", "12345")
				m.setField("MSH.12", "2.5")
				// PID.3 missing
			},
			wantValid: false,
			wantCount: 1,
		},
		{
			name:    "custom merged ruleset",
			ruleSet: MSHRules().Merge(NewRuleSet(At("MSH.11").Required().Build())),
			setup: func(m *mockMessage) {
				m.setField("MSH.9", "ADT^A01")
				m.setField("MSH.10", "12345")
				m.setField("MSH.12", "2.5")
				m.setField("MSH.11", "P")
			},
			wantValid: true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockMessage()
			tt.setup(m)

			v := NewWithRuleSet(tt.ruleSet)
			result := v.Validate(m)

			if result.Valid() != tt.wantValid {
				t.Errorf("Valid() = %v, want %v", result.Valid(), tt.wantValid)
			}
			if len(result.Errors()) != tt.wantCount {
				t.Errorf("Errors() = %d, want %d: %v", len(result.Errors()), tt.wantCount, result.Errors())
			}
		})
	}
}
