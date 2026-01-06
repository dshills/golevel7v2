package validate

// RuleSet represents a collection of validation rules that can be combined and reused.
type RuleSet interface {
	// Rules returns all rules in this set.
	Rules() []Rule
	// Add adds rules to this set and returns the set for chaining.
	Add(rules ...Rule) RuleSet
	// Merge combines this set with another set and returns a new set containing all rules.
	Merge(other RuleSet) RuleSet
}

// ruleSet is the concrete implementation of RuleSet.
type ruleSet struct {
	rules []Rule
}

// NewRuleSet creates a new RuleSet with the given rules.
func NewRuleSet(rules ...Rule) RuleSet {
	rs := &ruleSet{
		rules: make([]Rule, 0, len(rules)),
	}
	rs.rules = append(rs.rules, rules...)
	return rs
}

// Rules returns all rules in this set.
func (rs *ruleSet) Rules() []Rule {
	if rs.rules == nil {
		return []Rule{}
	}
	// Return a copy to prevent external modification
	result := make([]Rule, len(rs.rules))
	copy(result, rs.rules)
	return result
}

// Add adds rules to this set and returns the set for chaining.
func (rs *ruleSet) Add(rules ...Rule) RuleSet {
	rs.rules = append(rs.rules, rules...)
	return rs
}

// Merge combines this set with another set and returns a new set containing all rules.
func (rs *ruleSet) Merge(other RuleSet) RuleSet {
	if other == nil {
		return NewRuleSet(rs.rules...)
	}
	combined := make([]Rule, 0, len(rs.rules)+len(other.Rules()))
	combined = append(combined, rs.rules...)
	combined = append(combined, other.Rules()...)
	return NewRuleSet(combined...)
}

// MSHRules returns a RuleSet containing standard MSH segment validation rules.
// Validates:
//   - MSH.9 (Message Type) is required
//   - MSH.10 (Message Control ID) is required
//   - MSH.12 (Version ID) is required
func MSHRules() RuleSet {
	return NewRuleSet(
		At("MSH.9").Required().WithDescription("Message Type is required").Build(),
		At("MSH.10").Required().WithDescription("Message Control ID is required").Build(),
		At("MSH.12").Required().WithDescription("Version ID is required").Build(),
	)
}

// PIDRules returns a RuleSet containing standard PID segment validation rules.
// Validates:
//   - PID.3 (Patient Identifier List) is required
func PIDRules() RuleSet {
	return NewRuleSet(
		At("PID.3").Required().WithDescription("Patient Identifier is required").Build(),
	)
}

// PV1Rules returns a RuleSet containing standard PV1 segment validation rules.
// Validates:
//   - PV1.2 (Patient Class) is required
func PV1Rules() RuleSet {
	return NewRuleSet(
		At("PV1.2").Required().WithDescription("Patient Class is required").Build(),
	)
}

// OBRRules returns a RuleSet containing standard OBR segment validation rules.
// Validates:
//   - OBR.4 (Universal Service Identifier) is required
func OBRRules() RuleSet {
	return NewRuleSet(
		At("OBR.4").Required().WithDescription("Universal Service Identifier is required").Build(),
	)
}

// OBXRules returns a RuleSet containing standard OBX segment validation rules.
// Validates:
//   - OBX.2 (Value Type) is required
//   - OBX.3 (Observation Identifier) is required
func OBXRules() RuleSet {
	return NewRuleSet(
		At("OBX.2").Required().WithDescription("Value Type is required").Build(),
		At("OBX.3").Required().WithDescription("Observation Identifier is required").Build(),
	)
}

// ADTRules returns a RuleSet for ADT (Admit/Discharge/Transfer) messages.
// Combines MSH and PID rules.
func ADTRules() RuleSet {
	return MSHRules().Merge(PIDRules())
}

// ORURules returns a RuleSet for ORU (Observation Result) messages.
// Combines MSH, PID, OBR, and OBX rules.
func ORURules() RuleSet {
	return MSHRules().
		Merge(PIDRules()).
		Merge(OBRRules()).
		Merge(OBXRules())
}

// ORMRules returns a RuleSet for ORM (Order) messages.
// Combines MSH, PID, and OBR rules.
func ORMRules() RuleSet {
	return MSHRules().
		Merge(PIDRules()).
		Merge(OBRRules())
}

// StandardRules returns a RuleSet containing the minimum standard rules
// that apply to all HL7 messages (MSH segment rules).
func StandardRules() RuleSet {
	return MSHRules()
}
