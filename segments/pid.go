package segments

import (
	"fmt"

	"github.com/dshills/golevel7/hl7"
)

// PID represents the Patient Identification segment.
// This segment contains patient demographic information including identifiers,
// name, date of birth, address, and other patient-related data.
//
// Field positions follow the HL7 standard where PID-1 is the first field
// after the segment name.
type PID struct {
	// SetID is PID-1: Set ID for the PID segment (1-based sequence number).
	SetID string `hl7:"PID.1"`

	// PatientID is PID-2: Patient ID (external).
	// Note: This field is retained for backward compatibility; PID-3 is preferred.
	PatientID string `hl7:"PID.2"`

	// PatientIDList is PID-3: Patient identifier list.
	// This is the primary patient identifier field containing a list of IDs.
	PatientIDList string `hl7:"PID.3"`

	// AlternatePatientID is PID-4: Alternate patient ID.
	// Note: This field is retained for backward compatibility.
	AlternatePatientID string `hl7:"PID.4"`

	// PatientName is PID-5: Patient name (XPN - Extended Person Name).
	// Format: FamilyName^GivenName^MiddleName^Suffix^Prefix^Degree
	PatientName string `hl7:"PID.5"`

	// MotherMaidenName is PID-6: Mother's maiden name.
	MotherMaidenName string `hl7:"PID.6"`

	// DateOfBirth is PID-7: Date/time of birth (format: YYYYMMDD or YYYYMMDDHHMMSS).
	DateOfBirth string `hl7:"PID.7"`

	// Sex is PID-8: Administrative sex (M, F, O, U, A, N).
	Sex string `hl7:"PID.8"`

	// PatientAlias is PID-9: Patient alias (alternative names).
	PatientAlias string `hl7:"PID.9"`

	// Race is PID-10: Race code.
	Race string `hl7:"PID.10"`

	// PatientAddress is PID-11: Patient address (XAD - Extended Address).
	PatientAddress string `hl7:"PID.11"`

	// CountyCode is PID-12: County code.
	CountyCode string `hl7:"PID.12"`

	// PhoneNumberHome is PID-13: Home phone number (XTN - Extended Telecommunication Number).
	PhoneNumberHome string `hl7:"PID.13"`

	// PhoneNumberBusiness is PID-14: Business phone number.
	PhoneNumberBusiness string `hl7:"PID.14"`

	// PrimaryLanguage is PID-15: Primary language.
	PrimaryLanguage string `hl7:"PID.15"`

	// MaritalStatus is PID-16: Marital status code.
	MaritalStatus string `hl7:"PID.16"`

	// Religion is PID-17: Religion code.
	Religion string `hl7:"PID.17"`

	// PatientAccountNumber is PID-18: Patient account number.
	PatientAccountNumber string `hl7:"PID.18"`

	// SSNNumber is PID-19: SSN number (deprecated, use PID-3).
	SSNNumber string `hl7:"PID.19"`

	// DriversLicenseNumber is PID-20: Driver's license number.
	DriversLicenseNumber string `hl7:"PID.20"`

	// MothersIdentifier is PID-21: Mother's identifier.
	MothersIdentifier string `hl7:"PID.21"`

	// EthnicGroup is PID-22: Ethnic group code.
	EthnicGroup string `hl7:"PID.22"`

	// BirthPlace is PID-23: Birth place.
	BirthPlace string `hl7:"PID.23"`

	// MultipleBirthIndicator is PID-24: Multiple birth indicator (Y/N).
	MultipleBirthIndicator string `hl7:"PID.24"`

	// BirthOrder is PID-25: Birth order.
	BirthOrder string `hl7:"PID.25"`

	// Citizenship is PID-26: Citizenship code.
	Citizenship string `hl7:"PID.26"`

	// VeteransMilitaryStatus is PID-27: Veterans military status.
	VeteransMilitaryStatus string `hl7:"PID.27"`

	// Nationality is PID-28: Nationality code.
	Nationality string `hl7:"PID.28"`

	// PatientDeathDateTime is PID-29: Patient death date/time.
	PatientDeathDateTime string `hl7:"PID.29"`

	// PatientDeathIndicator is PID-30: Patient death indicator (Y/N).
	PatientDeathIndicator string `hl7:"PID.30"`

	// IdentityUnknownIndicator is PID-31: Identity unknown indicator.
	IdentityUnknownIndicator string `hl7:"PID.31"`

	// IdentityReliabilityCode is PID-32: Identity reliability code.
	IdentityReliabilityCode string `hl7:"PID.32"`

	// LastUpdateDateTime is PID-33: Last update date/time.
	LastUpdateDateTime string `hl7:"PID.33"`

	// LastUpdateFacility is PID-34: Last update facility.
	LastUpdateFacility string `hl7:"PID.34"`

	// SpeciesCode is PID-35: Species code (for veterinary).
	SpeciesCode string `hl7:"PID.35"`

	// BreedCode is PID-36: Breed code (for veterinary).
	BreedCode string `hl7:"PID.36"`

	// Strain is PID-37: Strain (for veterinary/laboratory).
	Strain string `hl7:"PID.37"`

	// ProductionClassCode is PID-38: Production class code.
	ProductionClassCode string `hl7:"PID.38"`

	// TribalCitizenship is PID-39: Tribal citizenship.
	TribalCitizenship string `hl7:"PID.39"`
}

// ErrNotPIDSegment indicates the segment is not a PID segment.
var ErrNotPIDSegment = fmt.Errorf("segment is not PID")

// ParsePID extracts field values from an hl7.Segment into a PID struct.
// Returns an error if the segment is nil or not a PID segment.
func ParsePID(seg hl7.Segment) (*PID, error) {
	if seg == nil {
		return nil, ErrNilSegment
	}

	if seg.Name() != "PID" {
		return nil, fmt.Errorf("%w: got %s", ErrNotPIDSegment, seg.Name())
	}

	pid := &PID{
		SetID:                    getFieldValue(seg, 1),
		PatientID:                getFieldValue(seg, 2),
		PatientIDList:            getFieldValue(seg, 3),
		AlternatePatientID:       getFieldValue(seg, 4),
		PatientName:              getFieldValue(seg, 5),
		MotherMaidenName:         getFieldValue(seg, 6),
		DateOfBirth:              getFieldValue(seg, 7),
		Sex:                      getFieldValue(seg, 8),
		PatientAlias:             getFieldValue(seg, 9),
		Race:                     getFieldValue(seg, 10),
		PatientAddress:           getFieldValue(seg, 11),
		CountyCode:               getFieldValue(seg, 12),
		PhoneNumberHome:          getFieldValue(seg, 13),
		PhoneNumberBusiness:      getFieldValue(seg, 14),
		PrimaryLanguage:          getFieldValue(seg, 15),
		MaritalStatus:            getFieldValue(seg, 16),
		Religion:                 getFieldValue(seg, 17),
		PatientAccountNumber:     getFieldValue(seg, 18),
		SSNNumber:                getFieldValue(seg, 19),
		DriversLicenseNumber:     getFieldValue(seg, 20),
		MothersIdentifier:        getFieldValue(seg, 21),
		EthnicGroup:              getFieldValue(seg, 22),
		BirthPlace:               getFieldValue(seg, 23),
		MultipleBirthIndicator:   getFieldValue(seg, 24),
		BirthOrder:               getFieldValue(seg, 25),
		Citizenship:              getFieldValue(seg, 26),
		VeteransMilitaryStatus:   getFieldValue(seg, 27),
		Nationality:              getFieldValue(seg, 28),
		PatientDeathDateTime:     getFieldValue(seg, 29),
		PatientDeathIndicator:    getFieldValue(seg, 30),
		IdentityUnknownIndicator: getFieldValue(seg, 31),
		IdentityReliabilityCode:  getFieldValue(seg, 32),
		LastUpdateDateTime:       getFieldValue(seg, 33),
		LastUpdateFacility:       getFieldValue(seg, 34),
		SpeciesCode:              getFieldValue(seg, 35),
		BreedCode:                getFieldValue(seg, 36),
		Strain:                   getFieldValue(seg, 37),
		ProductionClassCode:      getFieldValue(seg, 38),
		TribalCitizenship:        getFieldValue(seg, 39),
	}

	return pid, nil
}

// ToSegment converts the PID struct into an hl7.Segment.
// The delims parameter specifies the delimiters to use for encoding.
// If delims is nil, default delimiters are used.
func (p *PID) ToSegment(delims *hl7.Delimiters) (hl7.Segment, error) {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	fields := []string{
		p.SetID,
		p.PatientID,
		p.PatientIDList,
		p.AlternatePatientID,
		p.PatientName,
		p.MotherMaidenName,
		p.DateOfBirth,
		p.Sex,
		p.PatientAlias,
		p.Race,
		p.PatientAddress,
		p.CountyCode,
		p.PhoneNumberHome,
		p.PhoneNumberBusiness,
		p.PrimaryLanguage,
		p.MaritalStatus,
		p.Religion,
		p.PatientAccountNumber,
		p.SSNNumber,
		p.DriversLicenseNumber,
		p.MothersIdentifier,
		p.EthnicGroup,
		p.BirthPlace,
		p.MultipleBirthIndicator,
		p.BirthOrder,
		p.Citizenship,
		p.VeteransMilitaryStatus,
		p.Nationality,
		p.PatientDeathDateTime,
		p.PatientDeathIndicator,
		p.IdentityUnknownIndicator,
		p.IdentityReliabilityCode,
		p.LastUpdateDateTime,
		p.LastUpdateFacility,
		p.SpeciesCode,
		p.BreedCode,
		p.Strain,
		p.ProductionClassCode,
		p.TribalCitizenship,
	}

	data := buildSegmentData("PID", fields, delims)

	seg, err := hl7.ParseSegment([]rune(data), delims)
	if err != nil {
		return nil, fmt.Errorf("failed to create PID segment: %w", err)
	}

	return seg, nil
}
