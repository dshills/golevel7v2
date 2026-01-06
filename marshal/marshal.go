package marshal

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/dshills/golevel7/hl7"
)

// Marshal errors.
var (
	// ErrNotStructValue indicates the value is not a struct.
	ErrNotStructValue = errors.New("value must be a struct or pointer to struct")
)

// Marshaler converts Go structs to HL7 messages.
type Marshaler interface {
	// Marshal creates a new HL7 message from the struct.
	// The struct fields should be tagged with hl7 tags specifying the location path.
	//
	// Example:
	//   type Patient struct {
	//       ID        string    `hl7:"PID.3"`
	//       LastName  string    `hl7:"PID.5.1"`
	//       FirstName string    `hl7:"PID.5.2"`
	//       DOB       time.Time `hl7:"PID.7,format=20060102"`
	//   }
	//
	//   patient := Patient{ID: "12345", LastName: "Smith"}
	//   msg, err := marshaler.Marshal(patient)
	Marshal(v interface{}) (hl7.Message, error)

	// MarshalInto populates an existing HL7 message with data from the struct.
	// This allows updating specific fields while preserving other message content.
	//
	// Example:
	//   msg, _ := parser.Parse(rawMessage)
	//   patient := Patient{ID: "12345", LastName: "Smith"}
	//   err := marshaler.MarshalInto(msg, patient)
	MarshalInto(msg hl7.Message, v interface{}) error
}

// marshaler is the concrete implementation of Marshaler.
type marshaler struct {
	config *marshalConfig
}

// NewMarshaler creates a new Marshaler with the given options.
func NewMarshaler(opts ...Option) Marshaler {
	cfg := defaultConfig()
	cfg.applyOptions(opts...)
	return &marshaler{config: cfg}
}

// Marshal creates a new HL7 message from the struct.
func (m *marshaler) Marshal(v interface{}) (hl7.Message, error) {
	rv, err := m.getStructValue(v)
	if err != nil {
		return nil, err
	}

	// Create a new message
	msg := hl7.NewEmptyMessage()

	// Marshal struct into message
	if err := m.marshalStruct(msg, rv); err != nil {
		return nil, err
	}

	return msg, nil
}

// MarshalInto populates an existing HL7 message with data from the struct.
func (m *marshaler) MarshalInto(msg hl7.Message, v interface{}) error {
	if msg == nil {
		return ErrNilMessage
	}

	rv, err := m.getStructValue(v)
	if err != nil {
		return err
	}

	return m.marshalStruct(msg, rv)
}

// getStructValue extracts the reflect.Value of a struct from an interface.
func (m *marshaler) getStructValue(v interface{}) (reflect.Value, error) {
	rv := reflect.ValueOf(v)

	// Handle pointers
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return reflect.Value{}, ErrNilPointer
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, ErrNotStructValue
	}

	return rv, nil
}

// marshalStruct marshals a struct value into an HL7 message.
func (m *marshaler) marshalStruct(msg hl7.Message, rv reflect.Value) error {
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Get and parse tag
		tag := fieldType.Tag.Get(m.config.tagName)
		if tag == "" {
			// Check if it's a nested struct without a tag
			if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
				if err := m.marshalStruct(msg, field); err != nil {
					return err
				}
			}
			continue
		}

		tagInfo, err := parseTag(tag)
		if err != nil {
			return fmt.Errorf("field %s: %w", fieldType.Name, err)
		}

		if tagInfo.ignore || !tagInfo.hasLocation() {
			continue
		}

		// Check if we should skip zero values
		if tagInfo.shouldOmit(m.config.omitEmpty) && isZeroValue(field) {
			continue
		}

		// Marshal field into message
		if err := m.marshalField(msg, field, fieldType, tagInfo); err != nil {
			return fmt.Errorf("field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// marshalField marshals a single field into the message.
func (m *marshaler) marshalField(msg hl7.Message, field reflect.Value, fieldType reflect.StructField, tagInfo *tagInfo) error {
	// Handle slice types for repetitions
	if field.Kind() == reflect.Slice {
		return m.marshalSlice(msg, field, tagInfo)
	}

	// Handle pointer types
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil // Skip nil pointers
		}
		field = field.Elem()
	}

	// Handle nested structs (but not time.Time)
	if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
		return m.marshalNestedStruct(msg, field, tagInfo)
	}

	// Convert field value to string
	value, err := m.fieldToString(field, tagInfo)
	if err != nil {
		return err
	}

	if value == "" && tagInfo.shouldOmit(m.config.omitEmpty) {
		return nil
	}

	// Set value in message, creating segment if necessary
	return m.setMessageValue(msg, tagInfo.location, value)
}

// marshalSlice marshals a slice field into the message (for repetitions).
func (m *marshaler) marshalSlice(msg hl7.Message, field reflect.Value, tagInfo *tagInfo) error {
	if field.Len() == 0 {
		return nil
	}

	// For now, we handle the first element
	// Full repetition support requires message interface changes
	// to support setting multiple repetitions

	for i := 0; i < field.Len(); i++ {
		elem := field.Index(i)

		// Handle pointer elements
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}

		value, err := m.fieldToString(elem, tagInfo)
		if err != nil {
			return err
		}

		if value == "" {
			continue
		}

		// For the first element, set normally
		// For subsequent elements, we would need SetRepetition or similar
		if i == 0 {
			if err := m.setMessageValue(msg, tagInfo.location, value); err != nil {
				return err
			}
		}
		// Note: Additional repetitions require message interface support
		// which may not be fully implemented yet
	}

	return nil
}

// marshalNestedStruct handles nested struct fields.
func (m *marshaler) marshalNestedStruct(msg hl7.Message, field reflect.Value, tagInfo *tagInfo) error {
	rt := field.Type()

	for i := 0; i < field.NumField(); i++ {
		nestedField := field.Field(i)
		nestedFieldType := rt.Field(i)

		if !nestedFieldType.IsExported() {
			continue
		}

		tag := nestedFieldType.Tag.Get(m.config.tagName)
		if tag == "" {
			continue
		}

		nestedTagInfo, err := parseTag(tag)
		if err != nil {
			return fmt.Errorf("nested field %s: %w", nestedFieldType.Name, err)
		}

		if nestedTagInfo.ignore || !nestedTagInfo.hasLocation() {
			continue
		}

		// Combine parent location with nested location
		location := nestedTagInfo.location
		if tagInfo.location != "" && !startsWithSegment(location) {
			location = tagInfo.location + "." + location
			nestedTagInfo.location = location
		}

		if nestedTagInfo.shouldOmit(m.config.omitEmpty) && isZeroValue(nestedField) {
			continue
		}

		if err := m.marshalField(msg, nestedField, nestedFieldType, nestedTagInfo); err != nil {
			return err
		}
	}

	return nil
}

// setMessageValue sets a value in the message, creating the segment if necessary.
func (m *marshaler) setMessageValue(msg hl7.Message, location, value string) error {
	// Parse the location to extract segment name
	loc, err := hl7.ParseLocation(location)
	if err != nil {
		return err
	}

	// Check if segment exists, create if not
	_, found := msg.Segment(loc.Segment)
	if !found {
		// Create and add the segment
		seg := hl7.NewSegment(loc.Segment)
		if err := msg.AddSegment(seg); err != nil {
			return err
		}
	}

	// Now set the value
	return msg.Set(location, value)
}

// fieldToString converts a field value to its string representation.
func (m *marshaler) fieldToString(field reflect.Value, tagInfo *tagInfo) (string, error) {
	// Handle pointer
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return "", nil
		}
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.String:
		return field.String(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10), nil

	case reflect.Float32:
		return strconv.FormatFloat(field.Float(), 'f', -1, 32), nil

	case reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'f', -1, 64), nil

	case reflect.Bool:
		if field.Bool() {
			return "Y", nil
		}
		return "N", nil

	case reflect.Struct:
		// Check for time.Time
		if field.Type() == reflect.TypeOf(time.Time{}) {
			return m.timeToString(field.Interface().(time.Time), tagInfo), nil
		}
		return "", fmt.Errorf("%w: %s", ErrUnsupportedType, field.Type().String())

	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedType, field.Type().String())
	}
}

// timeToString formats a time.Time value as a string.
func (m *marshaler) timeToString(t time.Time, tagInfo *tagInfo) string {
	if t.IsZero() {
		return ""
	}

	format := tagInfo.getTimeFormat(m.config.timeFormat)
	return t.In(m.config.timeLocation).Format(format)
}

// isZeroValue checks if a value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// Special case for time.Time
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		// Check all fields
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}
