package segments

import (
	"fmt"

	"github.com/dshills/golevel7/hl7"
)

// ORC represents the Common Order segment.
// This segment is used for transmitting common order information, including
// order control codes, placer/filler order numbers, order status, and
// timing/quantity information for orders.
//
// Field positions follow the HL7 standard where ORC-1 is the first field
// after the segment name.
type ORC struct {
	// OrderControl is ORC-1: Order control code.
	// Determines the function of the order segment (NW=New, CA=Cancel, SC=Status Changed, etc.).
	OrderControl string `hl7:"ORC.1"`

	// PlacerOrderNumber is ORC-2: Placer order number (EI - Entity Identifier).
	// The order number assigned by the ordering/placing application.
	PlacerOrderNumber string `hl7:"ORC.2"`

	// FillerOrderNumber is ORC-3: Filler order number.
	// The order number assigned by the filling application.
	FillerOrderNumber string `hl7:"ORC.3"`

	// PlacerGroupNumber is ORC-4: Placer group number.
	// Used to group related orders from the placer application.
	PlacerGroupNumber string `hl7:"ORC.4"`

	// OrderStatus is ORC-5: Order status (A=Some but not all, CA=Canceled, CM=Completed, etc.).
	OrderStatus string `hl7:"ORC.5"`

	// ResponseFlag is ORC-6: Response flag (E=Report exceptions only, R=Same as E, D=Detailed, etc.).
	ResponseFlag string `hl7:"ORC.6"`

	// QuantityTiming is ORC-7: Quantity/timing (deprecated, use TQ1 segment).
	QuantityTiming string `hl7:"ORC.7"`

	// Parent is ORC-8: Parent order (EIP - Entity Identifier Pair).
	Parent string `hl7:"ORC.8"`

	// DateTimeOfTransaction is ORC-9: Date/time of transaction.
	DateTimeOfTransaction string `hl7:"ORC.9"`

	// EnteredBy is ORC-10: Entered by (XCN - Extended Composite ID Number and Name).
	EnteredBy string `hl7:"ORC.10"`

	// VerifiedBy is ORC-11: Verified by.
	VerifiedBy string `hl7:"ORC.11"`

	// OrderingProvider is ORC-12: Ordering provider.
	OrderingProvider string `hl7:"ORC.12"`

	// EntererLocation is ORC-13: Enterer's location (PL - Person Location).
	EntererLocation string `hl7:"ORC.13"`

	// CallBackPhoneNumber is ORC-14: Call back phone number (can repeat).
	CallBackPhoneNumber string `hl7:"ORC.14"`

	// OrderEffectiveDateTime is ORC-15: Order effective date/time.
	OrderEffectiveDateTime string `hl7:"ORC.15"`

	// OrderControlCodeReason is ORC-16: Order control code reason.
	OrderControlCodeReason string `hl7:"ORC.16"`

	// EnteringOrganization is ORC-17: Entering organization.
	EnteringOrganization string `hl7:"ORC.17"`

	// EnteringDevice is ORC-18: Entering device.
	EnteringDevice string `hl7:"ORC.18"`

	// ActionBy is ORC-19: Action by (XCN).
	ActionBy string `hl7:"ORC.19"`

	// AdvancedBeneficiaryNoticeCode is ORC-20: Advanced beneficiary notice code.
	AdvancedBeneficiaryNoticeCode string `hl7:"ORC.20"`

	// OrderingFacilityName is ORC-21: Ordering facility name (can repeat).
	OrderingFacilityName string `hl7:"ORC.21"`

	// OrderingFacilityAddress is ORC-22: Ordering facility address (can repeat).
	OrderingFacilityAddress string `hl7:"ORC.22"`

	// OrderingFacilityPhoneNumber is ORC-23: Ordering facility phone number (can repeat).
	OrderingFacilityPhoneNumber string `hl7:"ORC.23"`

	// OrderingProviderAddress is ORC-24: Ordering provider address (can repeat).
	OrderingProviderAddress string `hl7:"ORC.24"`

	// OrderStatusModifier is ORC-25: Order status modifier.
	OrderStatusModifier string `hl7:"ORC.25"`

	// AdvancedBeneficiaryNoticeOverrideReason is ORC-26: Advanced beneficiary notice override reason.
	AdvancedBeneficiaryNoticeOverrideReason string `hl7:"ORC.26"`

	// FillerExpectedAvailabilityDateTime is ORC-27: Filler's expected availability date/time.
	FillerExpectedAvailabilityDateTime string `hl7:"ORC.27"`

	// ConfidentialityCode is ORC-28: Confidentiality code.
	ConfidentialityCode string `hl7:"ORC.28"`

	// OrderType is ORC-29: Order type.
	OrderType string `hl7:"ORC.29"`

	// EntererAuthorizationMode is ORC-30: Enterer authorization mode.
	EntererAuthorizationMode string `hl7:"ORC.30"`

	// ParentUniversalServiceIdentifier is ORC-31: Parent universal service identifier.
	ParentUniversalServiceIdentifier string `hl7:"ORC.31"`
}

// ErrNotORCSegment indicates the segment is not an ORC segment.
var ErrNotORCSegment = fmt.Errorf("segment is not ORC")

// ParseORC extracts field values from an hl7.Segment into an ORC struct.
// Returns an error if the segment is nil or not an ORC segment.
func ParseORC(seg hl7.Segment) (*ORC, error) {
	if seg == nil {
		return nil, ErrNilSegment
	}

	if seg.Name() != "ORC" {
		return nil, fmt.Errorf("%w: got %s", ErrNotORCSegment, seg.Name())
	}

	orc := &ORC{
		OrderControl:                            getFieldValue(seg, 1),
		PlacerOrderNumber:                       getFieldValue(seg, 2),
		FillerOrderNumber:                       getFieldValue(seg, 3),
		PlacerGroupNumber:                       getFieldValue(seg, 4),
		OrderStatus:                             getFieldValue(seg, 5),
		ResponseFlag:                            getFieldValue(seg, 6),
		QuantityTiming:                          getFieldValue(seg, 7),
		Parent:                                  getFieldValue(seg, 8),
		DateTimeOfTransaction:                   getFieldValue(seg, 9),
		EnteredBy:                               getFieldValue(seg, 10),
		VerifiedBy:                              getFieldValue(seg, 11),
		OrderingProvider:                        getFieldValue(seg, 12),
		EntererLocation:                         getFieldValue(seg, 13),
		CallBackPhoneNumber:                     getFieldValue(seg, 14),
		OrderEffectiveDateTime:                  getFieldValue(seg, 15),
		OrderControlCodeReason:                  getFieldValue(seg, 16),
		EnteringOrganization:                    getFieldValue(seg, 17),
		EnteringDevice:                          getFieldValue(seg, 18),
		ActionBy:                                getFieldValue(seg, 19),
		AdvancedBeneficiaryNoticeCode:           getFieldValue(seg, 20),
		OrderingFacilityName:                    getFieldValue(seg, 21),
		OrderingFacilityAddress:                 getFieldValue(seg, 22),
		OrderingFacilityPhoneNumber:             getFieldValue(seg, 23),
		OrderingProviderAddress:                 getFieldValue(seg, 24),
		OrderStatusModifier:                     getFieldValue(seg, 25),
		AdvancedBeneficiaryNoticeOverrideReason: getFieldValue(seg, 26),
		FillerExpectedAvailabilityDateTime:      getFieldValue(seg, 27),
		ConfidentialityCode:                     getFieldValue(seg, 28),
		OrderType:                               getFieldValue(seg, 29),
		EntererAuthorizationMode:                getFieldValue(seg, 30),
		ParentUniversalServiceIdentifier:        getFieldValue(seg, 31),
	}

	return orc, nil
}

// ToSegment converts the ORC struct into an hl7.Segment.
// The delims parameter specifies the delimiters to use for encoding.
// If delims is nil, default delimiters are used.
func (o *ORC) ToSegment(delims *hl7.Delimiters) (hl7.Segment, error) {
	if delims == nil {
		delims = hl7.DefaultDelimiters()
	}

	fields := []string{
		o.OrderControl,
		o.PlacerOrderNumber,
		o.FillerOrderNumber,
		o.PlacerGroupNumber,
		o.OrderStatus,
		o.ResponseFlag,
		o.QuantityTiming,
		o.Parent,
		o.DateTimeOfTransaction,
		o.EnteredBy,
		o.VerifiedBy,
		o.OrderingProvider,
		o.EntererLocation,
		o.CallBackPhoneNumber,
		o.OrderEffectiveDateTime,
		o.OrderControlCodeReason,
		o.EnteringOrganization,
		o.EnteringDevice,
		o.ActionBy,
		o.AdvancedBeneficiaryNoticeCode,
		o.OrderingFacilityName,
		o.OrderingFacilityAddress,
		o.OrderingFacilityPhoneNumber,
		o.OrderingProviderAddress,
		o.OrderStatusModifier,
		o.AdvancedBeneficiaryNoticeOverrideReason,
		o.FillerExpectedAvailabilityDateTime,
		o.ConfidentialityCode,
		o.OrderType,
		o.EntererAuthorizationMode,
		o.ParentUniversalServiceIdentifier,
	}

	data := buildSegmentData("ORC", fields, delims)

	seg, err := hl7.ParseSegment([]rune(data), delims)
	if err != nil {
		return nil, fmt.Errorf("failed to create ORC segment: %w", err)
	}

	return seg, nil
}
