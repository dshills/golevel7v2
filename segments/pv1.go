package segments

import (
	"fmt"

	"github.com/dshills/golevel7/hl7"
)

// PV1 represents the Patient Visit segment.
// This segment contains information about the patient visit, including
// patient class, assigned locations, attending physicians, and visit-related
// administrative data.
//
// Field positions follow the HL7 standard where PV1-1 is the first field
// after the segment name.
type PV1 struct {
	// SetID is PV1-1: Set ID for the PV1 segment.
	SetID string `hl7:"PV1.1"`

	// PatientClass is PV1-2: Patient class (E=Emergency, I=Inpatient, O=Outpatient, etc.).
	PatientClass string `hl7:"PV1.2"`

	// AssignedPatientLocation is PV1-3: Assigned patient location (PL - Person Location).
	// Format: PointOfCare^Room^Bed^Facility^LocationStatus^PersonLocationType^Building^Floor
	AssignedPatientLocation string `hl7:"PV1.3"`

	// AdmissionType is PV1-4: Admission type code.
	AdmissionType string `hl7:"PV1.4"`

	// PreadmitNumber is PV1-5: Pre-admit number.
	PreadmitNumber string `hl7:"PV1.5"`

	// PriorPatientLocation is PV1-6: Prior patient location.
	PriorPatientLocation string `hl7:"PV1.6"`

	// AttendingDoctor is PV1-7: Attending doctor (XCN - Extended Composite ID Number and Name).
	AttendingDoctor string `hl7:"PV1.7"`

	// ReferringDoctor is PV1-8: Referring doctor.
	ReferringDoctor string `hl7:"PV1.8"`

	// ConsultingDoctor is PV1-9: Consulting doctor (can repeat).
	ConsultingDoctor string `hl7:"PV1.9"`

	// HospitalService is PV1-10: Hospital service code.
	HospitalService string `hl7:"PV1.10"`

	// TemporaryLocation is PV1-11: Temporary location.
	TemporaryLocation string `hl7:"PV1.11"`

	// PreadmitTestIndicator is PV1-12: Pre-admit test indicator.
	PreadmitTestIndicator string `hl7:"PV1.12"`

	// ReadmissionIndicator is PV1-13: Re-admission indicator.
	ReadmissionIndicator string `hl7:"PV1.13"`

	// AdmitSource is PV1-14: Admit source code.
	AdmitSource string `hl7:"PV1.14"`

	// AmbulatoryStatus is PV1-15: Ambulatory status (can repeat).
	AmbulatoryStatus string `hl7:"PV1.15"`

	// VIPIndicator is PV1-16: VIP indicator.
	VIPIndicator string `hl7:"PV1.16"`

	// AdmittingDoctor is PV1-17: Admitting doctor.
	AdmittingDoctor string `hl7:"PV1.17"`

	// PatientType is PV1-18: Patient type.
	PatientType string `hl7:"PV1.18"`

	// VisitNumber is PV1-19: Visit number (CX - Extended Composite ID with Check Digit).
	VisitNumber string `hl7:"PV1.19"`

	// FinancialClass is PV1-20: Financial class (can repeat).
	FinancialClass string `hl7:"PV1.20"`

	// ChargePriceIndicator is PV1-21: Charge price indicator.
	ChargePriceIndicator string `hl7:"PV1.21"`

	// CourtesyCode is PV1-22: Courtesy code.
	CourtesyCode string `hl7:"PV1.22"`

	// CreditRating is PV1-23: Credit rating.
	CreditRating string `hl7:"PV1.23"`

	// ContractCode is PV1-24: Contract code (can repeat).
	ContractCode string `hl7:"PV1.24"`

	// ContractEffectiveDate is PV1-25: Contract effective date (can repeat).
	ContractEffectiveDate string `hl7:"PV1.25"`

	// ContractAmount is PV1-26: Contract amount (can repeat).
	ContractAmount string `hl7:"PV1.26"`

	// ContractPeriod is PV1-27: Contract period (can repeat).
	ContractPeriod string `hl7:"PV1.27"`

	// InterestCode is PV1-28: Interest code.
	InterestCode string `hl7:"PV1.28"`

	// TransferToBadDebtCode is PV1-29: Transfer to bad debt code.
	TransferToBadDebtCode string `hl7:"PV1.29"`

	// TransferToBadDebtDate is PV1-30: Transfer to bad debt date.
	TransferToBadDebtDate string `hl7:"PV1.30"`

	// BadDebtAgencyCode is PV1-31: Bad debt agency code.
	BadDebtAgencyCode string `hl7:"PV1.31"`

	// BadDebtTransferAmount is PV1-32: Bad debt transfer amount.
	BadDebtTransferAmount string `hl7:"PV1.32"`

	// BadDebtRecoveryAmount is PV1-33: Bad debt recovery amount.
	BadDebtRecoveryAmount string `hl7:"PV1.33"`

	// DeleteAccountIndicator is PV1-34: Delete account indicator.
	DeleteAccountIndicator string `hl7:"PV1.34"`

	// DeleteAccountDate is PV1-35: Delete account date.
	DeleteAccountDate string `hl7:"PV1.35"`

	// DischargeDisposition is PV1-36: Discharge disposition.
	DischargeDisposition string `hl7:"PV1.36"`

	// DischargedToLocation is PV1-37: Discharged to location.
	DischargedToLocation string `hl7:"PV1.37"`

	// DietType is PV1-38: Diet type.
	DietType string `hl7:"PV1.38"`

	// ServicingFacility is PV1-39: Servicing facility.
	ServicingFacility string `hl7:"PV1.39"`

	// BedStatus is PV1-40: Bed status.
	BedStatus string `hl7:"PV1.40"`

	// AccountStatus is PV1-41: Account status.
	AccountStatus string `hl7:"PV1.41"`

	// PendingLocation is PV1-42: Pending location.
	PendingLocation string `hl7:"PV1.42"`

	// PriorTemporaryLocation is PV1-43: Prior temporary location.
	PriorTemporaryLocation string `hl7:"PV1.43"`

	// AdmitDateTime is PV1-44: Admit date/time.
	AdmitDateTime string `hl7:"PV1.44"`

	// DischargeDateTime is PV1-45: Discharge date/time.
	DischargeDateTime string `hl7:"PV1.45"`

	// CurrentPatientBalance is PV1-46: Current patient balance.
	CurrentPatientBalance string `hl7:"PV1.46"`

	// TotalCharges is PV1-47: Total charges.
	TotalCharges string `hl7:"PV1.47"`

	// TotalAdjustments is PV1-48: Total adjustments.
	TotalAdjustments string `hl7:"PV1.48"`

	// TotalPayments is PV1-49: Total payments.
	TotalPayments string `hl7:"PV1.49"`

	// AlternateVisitID is PV1-50: Alternate visit ID.
	AlternateVisitID string `hl7:"PV1.50"`

	// VisitIndicator is PV1-51: Visit indicator.
	VisitIndicator string `hl7:"PV1.51"`

	// OtherHealthcareProvider is PV1-52: Other healthcare provider.
	OtherHealthcareProvider string `hl7:"PV1.52"`
}

// ErrNotPV1Segment indicates the segment is not a PV1 segment.
var ErrNotPV1Segment = fmt.Errorf("segment is not PV1")

// ParsePV1 extracts field values from an hl7.Segment into a PV1 struct.
// Returns an error if the segment is nil or not a PV1 segment.
func ParsePV1(seg hl7.Segment) (*PV1, error) {
	if seg == nil {
		return nil, ErrNilSegment
	}

	if seg.Name() != "PV1" {
		return nil, fmt.Errorf("%w: got %s", ErrNotPV1Segment, seg.Name())
	}

	pv1 := &PV1{
		SetID:                   getFieldValue(seg, 1),
		PatientClass:            getFieldValue(seg, 2),
		AssignedPatientLocation: getFieldValue(seg, 3),
		AdmissionType:           getFieldValue(seg, 4),
		PreadmitNumber:          getFieldValue(seg, 5),
		PriorPatientLocation:    getFieldValue(seg, 6),
		AttendingDoctor:         getFieldValue(seg, 7),
		ReferringDoctor:         getFieldValue(seg, 8),
		ConsultingDoctor:        getFieldValue(seg, 9),
		HospitalService:         getFieldValue(seg, 10),
		TemporaryLocation:       getFieldValue(seg, 11),
		PreadmitTestIndicator:   getFieldValue(seg, 12),
		ReadmissionIndicator:    getFieldValue(seg, 13),
		AdmitSource:             getFieldValue(seg, 14),
		AmbulatoryStatus:        getFieldValue(seg, 15),
		VIPIndicator:            getFieldValue(seg, 16),
		AdmittingDoctor:         getFieldValue(seg, 17),
		PatientType:             getFieldValue(seg, 18),
		VisitNumber:             getFieldValue(seg, 19),
		FinancialClass:          getFieldValue(seg, 20),
		ChargePriceIndicator:    getFieldValue(seg, 21),
		CourtesyCode:            getFieldValue(seg, 22),
		CreditRating:            getFieldValue(seg, 23),
		ContractCode:            getFieldValue(seg, 24),
		ContractEffectiveDate:   getFieldValue(seg, 25),
		ContractAmount:          getFieldValue(seg, 26),
		ContractPeriod:          getFieldValue(seg, 27),
		InterestCode:            getFieldValue(seg, 28),
		TransferToBadDebtCode:   getFieldValue(seg, 29),
		TransferToBadDebtDate:   getFieldValue(seg, 30),
		BadDebtAgencyCode:       getFieldValue(seg, 31),
		BadDebtTransferAmount:   getFieldValue(seg, 32),
		BadDebtRecoveryAmount:   getFieldValue(seg, 33),
		DeleteAccountIndicator:  getFieldValue(seg, 34),
		DeleteAccountDate:       getFieldValue(seg, 35),
		DischargeDisposition:    getFieldValue(seg, 36),
		DischargedToLocation:    getFieldValue(seg, 37),
		DietType:                getFieldValue(seg, 38),
		ServicingFacility:       getFieldValue(seg, 39),
		BedStatus:               getFieldValue(seg, 40),
		AccountStatus:           getFieldValue(seg, 41),
		PendingLocation:         getFieldValue(seg, 42),
		PriorTemporaryLocation:  getFieldValue(seg, 43),
		AdmitDateTime:           getFieldValue(seg, 44),
		DischargeDateTime:       getFieldValue(seg, 45),
		CurrentPatientBalance:   getFieldValue(seg, 46),
		TotalCharges:            getFieldValue(seg, 47),
		TotalAdjustments:        getFieldValue(seg, 48),
		TotalPayments:           getFieldValue(seg, 49),
		AlternateVisitID:        getFieldValue(seg, 50),
		VisitIndicator:          getFieldValue(seg, 51),
		OtherHealthcareProvider: getFieldValue(seg, 52),
	}

	return pv1, nil
}

// ToSegment converts the PV1 struct into an hl7.Segment.
// The delims parameter specifies the delimiters to use for encoding.
// If delims is nil, default delimiters are used.
func (p *PV1) ToSegment(delims *hl7.Delimiters) (hl7.Segment, error) {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	fields := []string{
		p.SetID,
		p.PatientClass,
		p.AssignedPatientLocation,
		p.AdmissionType,
		p.PreadmitNumber,
		p.PriorPatientLocation,
		p.AttendingDoctor,
		p.ReferringDoctor,
		p.ConsultingDoctor,
		p.HospitalService,
		p.TemporaryLocation,
		p.PreadmitTestIndicator,
		p.ReadmissionIndicator,
		p.AdmitSource,
		p.AmbulatoryStatus,
		p.VIPIndicator,
		p.AdmittingDoctor,
		p.PatientType,
		p.VisitNumber,
		p.FinancialClass,
		p.ChargePriceIndicator,
		p.CourtesyCode,
		p.CreditRating,
		p.ContractCode,
		p.ContractEffectiveDate,
		p.ContractAmount,
		p.ContractPeriod,
		p.InterestCode,
		p.TransferToBadDebtCode,
		p.TransferToBadDebtDate,
		p.BadDebtAgencyCode,
		p.BadDebtTransferAmount,
		p.BadDebtRecoveryAmount,
		p.DeleteAccountIndicator,
		p.DeleteAccountDate,
		p.DischargeDisposition,
		p.DischargedToLocation,
		p.DietType,
		p.ServicingFacility,
		p.BedStatus,
		p.AccountStatus,
		p.PendingLocation,
		p.PriorTemporaryLocation,
		p.AdmitDateTime,
		p.DischargeDateTime,
		p.CurrentPatientBalance,
		p.TotalCharges,
		p.TotalAdjustments,
		p.TotalPayments,
		p.AlternateVisitID,
		p.VisitIndicator,
		p.OtherHealthcareProvider,
	}

	data := buildSegmentData("PV1", fields, delims)

	seg, err := hl7.ParseSegment([]rune(data), delims)
	if err != nil {
		return nil, fmt.Errorf("failed to create PV1 segment: %w", err)
	}

	return seg, nil
}
