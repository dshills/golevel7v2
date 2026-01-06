package segments

import (
	"fmt"

	"github.com/dshills/golevel7/hl7"
)

// OBR represents the Observation Request segment.
// This segment is used to transmit information about an observation request,
// such as laboratory test orders, radiology orders, or other clinical services.
//
// Field positions follow the HL7 standard where OBR-1 is the first field
// after the segment name.
type OBR struct {
	// SetID is OBR-1: Set ID for the OBR segment.
	SetID string `hl7:"OBR.1"`

	// PlacerOrderNumber is OBR-2: Placer order number (EI - Entity Identifier).
	PlacerOrderNumber string `hl7:"OBR.2"`

	// FillerOrderNumber is OBR-3: Filler order number.
	FillerOrderNumber string `hl7:"OBR.3"`

	// UniversalServiceIdentifier is OBR-4: Universal service identifier (CE - Coded Element).
	// This identifies the test/procedure being ordered.
	UniversalServiceIdentifier string `hl7:"OBR.4"`

	// Priority is OBR-5: Priority (deprecated, use TQ1-9).
	Priority string `hl7:"OBR.5"`

	// RequestedDateTime is OBR-6: Requested date/time (deprecated, use TQ1-7).
	RequestedDateTime string `hl7:"OBR.6"`

	// ObservationDateTime is OBR-7: Observation date/time.
	// The clinically relevant date/time of the observation.
	ObservationDateTime string `hl7:"OBR.7"`

	// ObservationEndDateTime is OBR-8: Observation end date/time.
	ObservationEndDateTime string `hl7:"OBR.8"`

	// CollectionVolume is OBR-9: Collection volume.
	CollectionVolume string `hl7:"OBR.9"`

	// CollectorIdentifier is OBR-10: Collector identifier (can repeat).
	CollectorIdentifier string `hl7:"OBR.10"`

	// SpecimenActionCode is OBR-11: Specimen action code.
	SpecimenActionCode string `hl7:"OBR.11"`

	// DangerCode is OBR-12: Danger code.
	DangerCode string `hl7:"OBR.12"`

	// RelevantClinicalInfo is OBR-13: Relevant clinical information.
	RelevantClinicalInfo string `hl7:"OBR.13"`

	// SpecimenReceivedDateTime is OBR-14: Specimen received date/time.
	SpecimenReceivedDateTime string `hl7:"OBR.14"`

	// SpecimenSource is OBR-15: Specimen source (deprecated, use SPM segment).
	SpecimenSource string `hl7:"OBR.15"`

	// OrderingProvider is OBR-16: Ordering provider (XCN - Extended Composite ID Number and Name).
	OrderingProvider string `hl7:"OBR.16"`

	// OrderCallbackPhoneNumber is OBR-17: Order callback phone number.
	OrderCallbackPhoneNumber string `hl7:"OBR.17"`

	// PlacerField1 is OBR-18: Placer field 1.
	PlacerField1 string `hl7:"OBR.18"`

	// PlacerField2 is OBR-19: Placer field 2.
	PlacerField2 string `hl7:"OBR.19"`

	// FillerField1 is OBR-20: Filler field 1.
	FillerField1 string `hl7:"OBR.20"`

	// FillerField2 is OBR-21: Filler field 2.
	FillerField2 string `hl7:"OBR.21"`

	// ResultsRptStatusChngDateTime is OBR-22: Results report/status change date/time.
	ResultsRptStatusChngDateTime string `hl7:"OBR.22"`

	// ChargeToPractice is OBR-23: Charge to practice.
	ChargeToPractice string `hl7:"OBR.23"`

	// DiagnosticServSectID is OBR-24: Diagnostic service section ID.
	DiagnosticServSectID string `hl7:"OBR.24"`

	// ResultStatus is OBR-25: Result status (F=Final, P=Preliminary, C=Correction, etc.).
	ResultStatus string `hl7:"OBR.25"`

	// ParentResult is OBR-26: Parent result.
	ParentResult string `hl7:"OBR.26"`

	// QuantityTiming is OBR-27: Quantity/timing (deprecated, use TQ1).
	QuantityTiming string `hl7:"OBR.27"`

	// ResultCopiesTo is OBR-28: Result copies to (can repeat).
	ResultCopiesTo string `hl7:"OBR.28"`

	// Parent is OBR-29: Parent (EIP - Entity Identifier Pair).
	Parent string `hl7:"OBR.29"`

	// TransportationMode is OBR-30: Transportation mode.
	TransportationMode string `hl7:"OBR.30"`

	// ReasonForStudy is OBR-31: Reason for study (can repeat).
	ReasonForStudy string `hl7:"OBR.31"`

	// PrincipalResultInterpreter is OBR-32: Principal result interpreter.
	PrincipalResultInterpreter string `hl7:"OBR.32"`

	// AssistantResultInterpreter is OBR-33: Assistant result interpreter (can repeat).
	AssistantResultInterpreter string `hl7:"OBR.33"`

	// Technician is OBR-34: Technician (can repeat).
	Technician string `hl7:"OBR.34"`

	// Transcriptionist is OBR-35: Transcriptionist (can repeat).
	Transcriptionist string `hl7:"OBR.35"`

	// ScheduledDateTime is OBR-36: Scheduled date/time.
	ScheduledDateTime string `hl7:"OBR.36"`

	// NumberOfSampleContainers is OBR-37: Number of sample containers.
	NumberOfSampleContainers string `hl7:"OBR.37"`

	// TransportLogisticsOfCollectedSample is OBR-38: Transport logistics of collected sample (can repeat).
	TransportLogisticsOfCollectedSample string `hl7:"OBR.38"`

	// CollectorComment is OBR-39: Collector's comment (can repeat).
	CollectorComment string `hl7:"OBR.39"`

	// TransportArrangementResponsibility is OBR-40: Transport arrangement responsibility.
	TransportArrangementResponsibility string `hl7:"OBR.40"`

	// TransportArranged is OBR-41: Transport arranged.
	TransportArranged string `hl7:"OBR.41"`

	// EscortRequired is OBR-42: Escort required.
	EscortRequired string `hl7:"OBR.42"`

	// PlannedPatientTransportComment is OBR-43: Planned patient transport comment (can repeat).
	PlannedPatientTransportComment string `hl7:"OBR.43"`

	// ProcedureCode is OBR-44: Procedure code.
	ProcedureCode string `hl7:"OBR.44"`

	// ProcedureCodeModifier is OBR-45: Procedure code modifier (can repeat).
	ProcedureCodeModifier string `hl7:"OBR.45"`

	// PlacerSupplementalServiceInformation is OBR-46: Placer supplemental service information (can repeat).
	PlacerSupplementalServiceInformation string `hl7:"OBR.46"`

	// FillerSupplementalServiceInformation is OBR-47: Filler supplemental service information (can repeat).
	FillerSupplementalServiceInformation string `hl7:"OBR.47"`

	// MedicallyNecessaryDuplicateProcedureReason is OBR-48: Medically necessary duplicate procedure reason.
	MedicallyNecessaryDuplicateProcedureReason string `hl7:"OBR.48"`

	// ResultHandling is OBR-49: Result handling.
	ResultHandling string `hl7:"OBR.49"`

	// ParentUniversalServiceIdentifier is OBR-50: Parent universal service identifier.
	ParentUniversalServiceIdentifier string `hl7:"OBR.50"`
}

// ErrNotOBRSegment indicates the segment is not an OBR segment.
var ErrNotOBRSegment = fmt.Errorf("segment is not OBR")

// ParseOBR extracts field values from an hl7.Segment into an OBR struct.
// Returns an error if the segment is nil or not an OBR segment.
func ParseOBR(seg hl7.Segment) (*OBR, error) {
	if seg == nil {
		return nil, ErrNilSegment
	}

	if seg.Name() != "OBR" {
		return nil, fmt.Errorf("%w: got %s", ErrNotOBRSegment, seg.Name())
	}

	obr := &OBR{
		SetID:                                      getFieldValue(seg, 1),
		PlacerOrderNumber:                          getFieldValue(seg, 2),
		FillerOrderNumber:                          getFieldValue(seg, 3),
		UniversalServiceIdentifier:                 getFieldValue(seg, 4),
		Priority:                                   getFieldValue(seg, 5),
		RequestedDateTime:                          getFieldValue(seg, 6),
		ObservationDateTime:                        getFieldValue(seg, 7),
		ObservationEndDateTime:                     getFieldValue(seg, 8),
		CollectionVolume:                           getFieldValue(seg, 9),
		CollectorIdentifier:                        getFieldValue(seg, 10),
		SpecimenActionCode:                         getFieldValue(seg, 11),
		DangerCode:                                 getFieldValue(seg, 12),
		RelevantClinicalInfo:                       getFieldValue(seg, 13),
		SpecimenReceivedDateTime:                   getFieldValue(seg, 14),
		SpecimenSource:                             getFieldValue(seg, 15),
		OrderingProvider:                           getFieldValue(seg, 16),
		OrderCallbackPhoneNumber:                   getFieldValue(seg, 17),
		PlacerField1:                               getFieldValue(seg, 18),
		PlacerField2:                               getFieldValue(seg, 19),
		FillerField1:                               getFieldValue(seg, 20),
		FillerField2:                               getFieldValue(seg, 21),
		ResultsRptStatusChngDateTime:               getFieldValue(seg, 22),
		ChargeToPractice:                           getFieldValue(seg, 23),
		DiagnosticServSectID:                       getFieldValue(seg, 24),
		ResultStatus:                               getFieldValue(seg, 25),
		ParentResult:                               getFieldValue(seg, 26),
		QuantityTiming:                             getFieldValue(seg, 27),
		ResultCopiesTo:                             getFieldValue(seg, 28),
		Parent:                                     getFieldValue(seg, 29),
		TransportationMode:                         getFieldValue(seg, 30),
		ReasonForStudy:                             getFieldValue(seg, 31),
		PrincipalResultInterpreter:                 getFieldValue(seg, 32),
		AssistantResultInterpreter:                 getFieldValue(seg, 33),
		Technician:                                 getFieldValue(seg, 34),
		Transcriptionist:                           getFieldValue(seg, 35),
		ScheduledDateTime:                          getFieldValue(seg, 36),
		NumberOfSampleContainers:                   getFieldValue(seg, 37),
		TransportLogisticsOfCollectedSample:        getFieldValue(seg, 38),
		CollectorComment:                           getFieldValue(seg, 39),
		TransportArrangementResponsibility:         getFieldValue(seg, 40),
		TransportArranged:                          getFieldValue(seg, 41),
		EscortRequired:                             getFieldValue(seg, 42),
		PlannedPatientTransportComment:             getFieldValue(seg, 43),
		ProcedureCode:                              getFieldValue(seg, 44),
		ProcedureCodeModifier:                      getFieldValue(seg, 45),
		PlacerSupplementalServiceInformation:       getFieldValue(seg, 46),
		FillerSupplementalServiceInformation:       getFieldValue(seg, 47),
		MedicallyNecessaryDuplicateProcedureReason: getFieldValue(seg, 48),
		ResultHandling:                             getFieldValue(seg, 49),
		ParentUniversalServiceIdentifier:           getFieldValue(seg, 50),
	}

	return obr, nil
}

// ToSegment converts the OBR struct into an hl7.Segment.
// The delims parameter specifies the delimiters to use for encoding.
// If delims is nil, default delimiters are used.
func (o *OBR) ToSegment(delims *hl7.Delimiters) (hl7.Segment, error) {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	fields := []string{
		o.SetID,
		o.PlacerOrderNumber,
		o.FillerOrderNumber,
		o.UniversalServiceIdentifier,
		o.Priority,
		o.RequestedDateTime,
		o.ObservationDateTime,
		o.ObservationEndDateTime,
		o.CollectionVolume,
		o.CollectorIdentifier,
		o.SpecimenActionCode,
		o.DangerCode,
		o.RelevantClinicalInfo,
		o.SpecimenReceivedDateTime,
		o.SpecimenSource,
		o.OrderingProvider,
		o.OrderCallbackPhoneNumber,
		o.PlacerField1,
		o.PlacerField2,
		o.FillerField1,
		o.FillerField2,
		o.ResultsRptStatusChngDateTime,
		o.ChargeToPractice,
		o.DiagnosticServSectID,
		o.ResultStatus,
		o.ParentResult,
		o.QuantityTiming,
		o.ResultCopiesTo,
		o.Parent,
		o.TransportationMode,
		o.ReasonForStudy,
		o.PrincipalResultInterpreter,
		o.AssistantResultInterpreter,
		o.Technician,
		o.Transcriptionist,
		o.ScheduledDateTime,
		o.NumberOfSampleContainers,
		o.TransportLogisticsOfCollectedSample,
		o.CollectorComment,
		o.TransportArrangementResponsibility,
		o.TransportArranged,
		o.EscortRequired,
		o.PlannedPatientTransportComment,
		o.ProcedureCode,
		o.ProcedureCodeModifier,
		o.PlacerSupplementalServiceInformation,
		o.FillerSupplementalServiceInformation,
		o.MedicallyNecessaryDuplicateProcedureReason,
		o.ResultHandling,
		o.ParentUniversalServiceIdentifier,
	}

	data := buildSegmentData("OBR", fields, delims)

	seg, err := hl7.ParseSegment([]rune(data), delims)
	if err != nil {
		return nil, fmt.Errorf("failed to create OBR segment: %w", err)
	}

	return seg, nil
}
