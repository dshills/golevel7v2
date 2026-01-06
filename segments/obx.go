package segments

import (
	"fmt"

	"github.com/dshills/golevel7/hl7"
)

// OBX represents the Observation Result segment.
// This segment is used to transmit a single observation or observation fragment.
// It contains the actual results of observations, including laboratory test results,
// vital signs, or other clinical measurements.
//
// Field positions follow the HL7 standard where OBX-1 is the first field
// after the segment name.
type OBX struct {
	// SetID is OBX-1: Set ID for the OBX segment.
	SetID string `hl7:"OBX.1"`

	// ValueType is OBX-2: Value type (CE, CWE, NM, ST, TX, etc.).
	// Indicates the data type of the observation value in OBX-5.
	ValueType string `hl7:"OBX.2"`

	// ObservationIdentifier is OBX-3: Observation identifier (CE/CWE).
	// Identifies the observation being reported (e.g., LOINC code).
	ObservationIdentifier string `hl7:"OBX.3"`

	// ObservationSubID is OBX-4: Observation sub-ID.
	// Used to distinguish between multiple OBX segments with the same observation identifier.
	ObservationSubID string `hl7:"OBX.4"`

	// ObservationValue is OBX-5: Observation value (varies based on OBX-2).
	// The actual result value. Can repeat for some value types.
	ObservationValue string `hl7:"OBX.5"`

	// Units is OBX-6: Units (CE/CWE).
	// The units of measurement for the observation value.
	Units string `hl7:"OBX.6"`

	// ReferencesRange is OBX-7: Reference range.
	// The normal reference range for the observation.
	ReferencesRange string `hl7:"OBX.7"`

	// AbnormalFlags is OBX-8: Abnormal flags (can repeat).
	// Indicates the normalcy status of the result (L=Low, H=High, N=Normal, etc.).
	AbnormalFlags string `hl7:"OBX.8"`

	// Probability is OBX-9: Probability.
	// For probabilistic results.
	Probability string `hl7:"OBX.9"`

	// NatureOfAbnormalTest is OBX-10: Nature of abnormal test (can repeat).
	// Indicates the nature of the abnormality (A=Age, S=Sex, R=Race, etc.).
	NatureOfAbnormalTest string `hl7:"OBX.10"`

	// ObservationResultStatus is OBX-11: Observation result status.
	// Status of the observation (F=Final, P=Preliminary, C=Correction, etc.).
	ObservationResultStatus string `hl7:"OBX.11"`

	// EffectiveDateOfReferenceRange is OBX-12: Effective date of reference range.
	EffectiveDateOfReferenceRange string `hl7:"OBX.12"`

	// UserDefinedAccessChecks is OBX-13: User defined access checks.
	UserDefinedAccessChecks string `hl7:"OBX.13"`

	// DateTimeOfObservation is OBX-14: Date/time of the observation.
	DateTimeOfObservation string `hl7:"OBX.14"`

	// ProducersID is OBX-15: Producer's ID.
	// Identifies the producer of the observation (lab, device, etc.).
	ProducersID string `hl7:"OBX.15"`

	// ResponsibleObserver is OBX-16: Responsible observer.
	ResponsibleObserver string `hl7:"OBX.16"`

	// ObservationMethod is OBX-17: Observation method (can repeat).
	ObservationMethod string `hl7:"OBX.17"`

	// EquipmentInstanceIdentifier is OBX-18: Equipment instance identifier (can repeat).
	EquipmentInstanceIdentifier string `hl7:"OBX.18"`

	// DateTimeOfAnalysis is OBX-19: Date/time of the analysis.
	DateTimeOfAnalysis string `hl7:"OBX.19"`

	// ObservationSite is OBX-20: Observation site (can repeat).
	ObservationSite string `hl7:"OBX.20"`

	// ObservationInstanceIdentifier is OBX-21: Observation instance identifier.
	ObservationInstanceIdentifier string `hl7:"OBX.21"`

	// MoodCode is OBX-22: Mood code.
	MoodCode string `hl7:"OBX.22"`

	// PerformingOrganizationName is OBX-23: Performing organization name.
	PerformingOrganizationName string `hl7:"OBX.23"`

	// PerformingOrganizationAddress is OBX-24: Performing organization address.
	PerformingOrganizationAddress string `hl7:"OBX.24"`

	// PerformingOrganizationMedicalDirector is OBX-25: Performing organization medical director.
	PerformingOrganizationMedicalDirector string `hl7:"OBX.25"`
}

// ErrNotOBXSegment indicates the segment is not an OBX segment.
var ErrNotOBXSegment = fmt.Errorf("segment is not OBX")

// ParseOBX extracts field values from an hl7.Segment into an OBX struct.
// Returns an error if the segment is nil or not an OBX segment.
func ParseOBX(seg hl7.Segment) (*OBX, error) {
	if seg == nil {
		return nil, ErrNilSegment
	}

	if seg.Name() != "OBX" {
		return nil, fmt.Errorf("%w: got %s", ErrNotOBXSegment, seg.Name())
	}

	obx := &OBX{
		SetID:                                 getFieldValue(seg, 1),
		ValueType:                             getFieldValue(seg, 2),
		ObservationIdentifier:                 getFieldValue(seg, 3),
		ObservationSubID:                      getFieldValue(seg, 4),
		ObservationValue:                      getFieldValue(seg, 5),
		Units:                                 getFieldValue(seg, 6),
		ReferencesRange:                       getFieldValue(seg, 7),
		AbnormalFlags:                         getFieldValue(seg, 8),
		Probability:                           getFieldValue(seg, 9),
		NatureOfAbnormalTest:                  getFieldValue(seg, 10),
		ObservationResultStatus:               getFieldValue(seg, 11),
		EffectiveDateOfReferenceRange:         getFieldValue(seg, 12),
		UserDefinedAccessChecks:               getFieldValue(seg, 13),
		DateTimeOfObservation:                 getFieldValue(seg, 14),
		ProducersID:                           getFieldValue(seg, 15),
		ResponsibleObserver:                   getFieldValue(seg, 16),
		ObservationMethod:                     getFieldValue(seg, 17),
		EquipmentInstanceIdentifier:           getFieldValue(seg, 18),
		DateTimeOfAnalysis:                    getFieldValue(seg, 19),
		ObservationSite:                       getFieldValue(seg, 20),
		ObservationInstanceIdentifier:         getFieldValue(seg, 21),
		MoodCode:                              getFieldValue(seg, 22),
		PerformingOrganizationName:            getFieldValue(seg, 23),
		PerformingOrganizationAddress:         getFieldValue(seg, 24),
		PerformingOrganizationMedicalDirector: getFieldValue(seg, 25),
	}

	return obx, nil
}

// ToSegment converts the OBX struct into an hl7.Segment.
// The delims parameter specifies the delimiters to use for encoding.
// If delims is nil, default delimiters are used.
func (o *OBX) ToSegment(delims *hl7.Delimiters) (hl7.Segment, error) {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	fields := []string{
		o.SetID,
		o.ValueType,
		o.ObservationIdentifier,
		o.ObservationSubID,
		o.ObservationValue,
		o.Units,
		o.ReferencesRange,
		o.AbnormalFlags,
		o.Probability,
		o.NatureOfAbnormalTest,
		o.ObservationResultStatus,
		o.EffectiveDateOfReferenceRange,
		o.UserDefinedAccessChecks,
		o.DateTimeOfObservation,
		o.ProducersID,
		o.ResponsibleObserver,
		o.ObservationMethod,
		o.EquipmentInstanceIdentifier,
		o.DateTimeOfAnalysis,
		o.ObservationSite,
		o.ObservationInstanceIdentifier,
		o.MoodCode,
		o.PerformingOrganizationName,
		o.PerformingOrganizationAddress,
		o.PerformingOrganizationMedicalDirector,
	}

	data := buildSegmentData("OBX", fields, delims)

	seg, err := hl7.ParseSegment([]rune(data), delims)
	if err != nil {
		return nil, fmt.Errorf("failed to create OBX segment: %w", err)
	}

	return seg, nil
}
