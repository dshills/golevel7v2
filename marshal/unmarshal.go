package marshal

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dshills/golevel7/hl7"
)

// Unmarshal errors.
var (
	// ErrNotPointer indicates the target is not a pointer.
	ErrNotPointer = errors.New("target must be a pointer")
	// ErrNotStruct indicates the target is not a struct.
	ErrNotStruct = errors.New("target must be a struct")
	// ErrNilPointer indicates a nil pointer was provided.
	ErrNilPointer = errors.New("target pointer is nil")
	// ErrNilMessage indicates a nil message was provided.
	ErrNilMessage = errors.New("message is nil")
	// ErrUnsupportedType indicates an unsupported field type.
	ErrUnsupportedType = errors.New("unsupported field type")
)

// Unmarshaler populates Go structs from HL7 messages.
type Unmarshaler interface {
	// Unmarshal populates the struct pointed to by v with data from the HL7 message.
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
	//   var patient Patient
	//   err := unmarshaler.Unmarshal(msg, &patient)
	Unmarshal(msg hl7.Message, v interface{}) error
}

// unmarshaler is the concrete implementation of Unmarshaler.
type unmarshaler struct {
	config *marshalConfig
}

// NewUnmarshaler creates a new Unmarshaler with the given options.
func NewUnmarshaler(opts ...Option) Unmarshaler {
	cfg := defaultConfig()
	cfg.applyOptions(opts...)
	return &unmarshaler{config: cfg}
}

// Unmarshal populates the struct pointed to by v with data from the HL7 message.
func (u *unmarshaler) Unmarshal(msg hl7.Message, v interface{}) error {
	if msg == nil {
		return ErrNilMessage
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return ErrNotPointer
	}
	if rv.IsNil() {
		return ErrNilPointer
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	return u.unmarshalStruct(msg, rv)
}

// unmarshalStruct unmarshals message data into a struct value.
func (u *unmarshaler) unmarshalStruct(msg hl7.Message, rv reflect.Value) error {
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get and parse tag
		tag := fieldType.Tag.Get(u.config.tagName)
		if tag == "" {
			// Check if it's a nested struct without a tag
			if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
				if err := u.unmarshalStruct(msg, field); err != nil {
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

		// Get value from message
		if err := u.unmarshalField(msg, field, fieldType, tagInfo); err != nil {
			return fmt.Errorf("field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// unmarshalField unmarshals a single field from the message.
func (u *unmarshaler) unmarshalField(msg hl7.Message, field reflect.Value, fieldType reflect.StructField, tagInfo *tagInfo) error {
	// Handle slice types for repetitions
	if field.Kind() == reflect.Slice {
		return u.unmarshalSlice(msg, field, fieldType, tagInfo)
	}

	// Handle pointer types
	if field.Kind() == reflect.Ptr {
		return u.unmarshalPointer(msg, field, fieldType, tagInfo)
	}

	// Handle nested structs (but not time.Time)
	if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
		return u.unmarshalNestedStruct(msg, field, tagInfo)
	}

	// Get single value from message
	value, err := msg.Get(tagInfo.location)
	if err != nil {
		// Field not found is not an error for unmarshaling
		if errors.Is(err, hl7.ErrSegmentNotFound) ||
			errors.Is(err, hl7.ErrFieldNotFound) ||
			errors.Is(err, hl7.ErrComponentNotFound) ||
			errors.Is(err, hl7.ErrSubComponentNotFound) {
			return nil
		}
		return err
	}

	if value == "" {
		return nil
	}

	return u.setFieldValue(field, value, tagInfo)
}

// unmarshalSlice unmarshals a slice field (for repetitions).
func (u *unmarshaler) unmarshalSlice(msg hl7.Message, field reflect.Value, fieldType reflect.StructField, tagInfo *tagInfo) error {
	// Get all values for this location
	values, err := msg.GetAll(tagInfo.location)
	if err != nil {
		// Field not found is not an error
		if errors.Is(err, hl7.ErrSegmentNotFound) ||
			errors.Is(err, hl7.ErrFieldNotFound) {
			return nil
		}
		return err
	}

	if len(values) == 0 {
		return nil
	}

	// Create slice of appropriate type
	elemType := fieldType.Type.Elem()
	slice := reflect.MakeSlice(fieldType.Type, len(values), len(values))

	for i, value := range values {
		if value == "" {
			continue
		}

		elem := slice.Index(i)

		// Handle pointer elements
		if elemType.Kind() == reflect.Ptr {
			ptr := reflect.New(elemType.Elem())
			if err := u.setFieldValue(ptr.Elem(), value, tagInfo); err != nil {
				return err
			}
			elem.Set(ptr)
		} else {
			if err := u.setFieldValue(elem, value, tagInfo); err != nil {
				return err
			}
		}
	}

	field.Set(slice)
	return nil
}

// unmarshalPointer unmarshals a pointer field.
func (u *unmarshaler) unmarshalPointer(msg hl7.Message, field reflect.Value, fieldType reflect.StructField, tagInfo *tagInfo) error {
	value, err := msg.Get(tagInfo.location)
	if err != nil {
		// Field not found is not an error
		if errors.Is(err, hl7.ErrSegmentNotFound) ||
			errors.Is(err, hl7.ErrFieldNotFound) ||
			errors.Is(err, hl7.ErrComponentNotFound) ||
			errors.Is(err, hl7.ErrSubComponentNotFound) {
			return nil
		}
		return err
	}

	if value == "" {
		return nil
	}

	// Create new value and set
	ptr := reflect.New(fieldType.Type.Elem())
	if err := u.setFieldValue(ptr.Elem(), value, tagInfo); err != nil {
		return err
	}
	field.Set(ptr)
	return nil
}

// unmarshalNestedStruct handles nested struct fields.
func (u *unmarshaler) unmarshalNestedStruct(msg hl7.Message, field reflect.Value, tagInfo *tagInfo) error {
	// For nested structs with a location tag, we treat the location as a prefix
	// and the nested fields extend from that prefix
	rt := field.Type()

	for i := 0; i < field.NumField(); i++ {
		nestedField := field.Field(i)
		nestedFieldType := rt.Field(i)

		if !nestedField.CanSet() {
			continue
		}

		tag := nestedFieldType.Tag.Get(u.config.tagName)
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

		// Combine parent location with nested location if the nested location
		// doesn't start with a segment name
		location := nestedTagInfo.location
		if tagInfo.location != "" && !startsWithSegment(location) {
			location = tagInfo.location + "." + location
			nestedTagInfo.location = location
		}

		if err := u.unmarshalField(msg, nestedField, nestedFieldType, nestedTagInfo); err != nil {
			return err
		}
	}

	return nil
}

// startsWithSegment checks if a location string starts with a segment name.
func startsWithSegment(loc string) bool {
	if len(loc) < 3 {
		return false
	}
	// Segment names are 3 uppercase letters
	for i := 0; i < 3; i++ {
		c := loc[i]
		if i == 0 {
			if c < 'A' || c > 'Z' {
				return false
			}
		} else {
			if (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
				return false
			}
		}
	}
	// Must be followed by end or '.' or '['
	if len(loc) == 3 {
		return true
	}
	return loc[3] == '.' || loc[3] == '['
}

// setFieldValue sets the field value from a string, performing type conversion.
func (u *unmarshaler) setFieldValue(field reflect.Value, value string, tagInfo *tagInfo) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return u.setIntValue(field, value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return u.setUintValue(field, value)

	case reflect.Float32, reflect.Float64:
		return u.setFloatValue(field, value)

	case reflect.Bool:
		return u.setBoolValue(field, value)

	case reflect.Struct:
		// Check for time.Time
		if field.Type() == reflect.TypeOf(time.Time{}) {
			return u.setTimeValue(field, value, tagInfo)
		}
		return fmt.Errorf("%w: %s", ErrUnsupportedType, field.Type().String())

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedType, field.Type().String())
	}
}

// setIntValue sets an integer field value.
func (u *unmarshaler) setIntValue(field reflect.Value, value string) error {
	// Clean the value - remove any non-numeric characters except leading minus
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("cannot parse %q as int: %w", value, err)
	}
	field.SetInt(i)
	return nil
}

// setUintValue sets an unsigned integer field value.
func (u *unmarshaler) setUintValue(field reflect.Value, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	i, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fmt.Errorf("cannot parse %q as uint: %w", value, err)
	}
	field.SetUint(i)
	return nil
}

// setFloatValue sets a float field value.
func (u *unmarshaler) setFloatValue(field reflect.Value, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("cannot parse %q as float: %w", value, err)
	}
	field.SetFloat(f)
	return nil
}

// setBoolValue sets a boolean field value.
// Accepts: "true", "false", "1", "0", "Y", "N", "yes", "no" (case-insensitive)
func (u *unmarshaler) setBoolValue(field reflect.Value, value string) error {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return nil
	}

	switch value {
	case "true", "1", "y", "yes":
		field.SetBool(true)
	case "false", "0", "n", "no":
		field.SetBool(false)
	default:
		return fmt.Errorf("cannot parse %q as bool", value)
	}
	return nil
}

// setTimeValue sets a time.Time field value.
func (u *unmarshaler) setTimeValue(field reflect.Value, value string, tagInfo *tagInfo) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	format := tagInfo.getTimeFormat(u.config.timeFormat)

	// Try parsing with the specified format
	t, err := time.ParseInLocation(format, value, u.config.timeLocation)
	if err != nil {
		// Try common HL7 formats if the specified format fails
		formats := []string{
			"20060102150405.0000-0700",
			"20060102150405.000-0700",
			"20060102150405-0700",
			"20060102150405.0000",
			"20060102150405.000",
			"20060102150405",
			"200601021504",
			"2006010215",
			"20060102",
		}

		for _, f := range formats {
			if len(value) == len(f) || (len(value) > len(f) && (f[len(f)-1] == '0' || f[len(f)-1] == '4')) {
				t, err = time.ParseInLocation(f, value, u.config.timeLocation)
				if err == nil {
					break
				}
			}
		}

		if err != nil {
			return fmt.Errorf("cannot parse %q as time with format %q: %w", value, format, err)
		}
	}

	field.Set(reflect.ValueOf(t))
	return nil
}
