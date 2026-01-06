package hl7

import (
	"testing"
)

func TestNewSegment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
	}{
		{
			name:     "uppercase name",
			input:    "PID",
			wantName: "PID",
		},
		{
			name:     "lowercase converted to uppercase",
			input:    "pid",
			wantName: "PID",
		},
		{
			name:     "mixed case converted",
			input:    "Msh",
			wantName: "MSH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg := NewSegment(tt.input)
			if seg.Name() != tt.wantName {
				t.Errorf("NewSegment(%q).Name() = %q, want %q", tt.input, seg.Name(), tt.wantName)
			}
			if seg.FieldCount() != 0 {
				t.Errorf("NewSegment(%q).FieldCount() = %d, want 0", tt.input, seg.FieldCount())
			}
		})
	}
}

func TestParseSegment_Regular(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name       string
		data       string
		wantName   string
		wantFields int
		wantField1 string
		wantField2 string
	}{
		{
			name:       "PID segment with fields",
			data:       "PID|1|12345|PatientID^^^Hospital",
			wantName:   "PID",
			wantFields: 3,
			wantField1: "1",
			wantField2: "12345",
		},
		{
			name:       "OBX segment",
			data:       "OBX|1|ST|Code^Description|Value",
			wantName:   "OBX",
			wantFields: 4,
			wantField1: "1",
			wantField2: "ST",
		},
		{
			name:       "segment with empty fields",
			data:       "PID|1||PatientID",
			wantName:   "PID",
			wantFields: 3,
			wantField1: "1",
			wantField2: "",
		},
		{
			name:       "segment name only",
			data:       "EVN",
			wantName:   "EVN",
			wantFields: 0,
		},
		{
			name:       "segment with trailing delimiter",
			data:       "PID|1|2|",
			wantName:   "PID",
			wantFields: 3,
			wantField1: "1",
			wantField2: "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := ParseSegment([]rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseSegment() error = %v", err)
			}

			if seg.Name() != tt.wantName {
				t.Errorf("Name() = %q, want %q", seg.Name(), tt.wantName)
			}

			if seg.FieldCount() != tt.wantFields {
				t.Errorf("FieldCount() = %d, want %d", seg.FieldCount(), tt.wantFields)
			}

			if tt.wantFields > 0 {
				field1, ok := seg.Field(1)
				if !ok {
					t.Error("Field(1) not found")
				} else if field1.Value() != tt.wantField1 {
					t.Errorf("Field(1).Value() = %q, want %q", field1.Value(), tt.wantField1)
				}
			}

			if tt.wantFields > 1 {
				field2, ok := seg.Field(2)
				if !ok {
					t.Error("Field(2) not found")
				} else if field2.Value() != tt.wantField2 {
					t.Errorf("Field(2).Value() = %q, want %q", field2.Value(), tt.wantField2)
				}
			}
		})
	}
}

func TestParseSegment_MSH(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name       string
		data       string
		wantMSH1   string
		wantMSH2   string
		wantMSH3   string
		wantMSH4   string
		wantFields int
	}{
		{
			name:       "standard MSH",
			data:       "MSH|^~\\&|SendingApp|SendingFac|ReceivingApp|ReceivingFac",
			wantMSH1:   "|",
			wantMSH2:   "^~\\&",
			wantMSH3:   "SendingApp",
			wantMSH4:   "SendingFac",
			wantFields: 6,
		},
		{
			name:       "MSH with custom delimiters in data",
			data:       "MSH|^~\\&#|App^Sub|Facility",
			wantMSH1:   "|",
			wantMSH2:   "^~\\&#",
			wantMSH3:   "App^Sub", // Value() returns full encoded value including components
			wantMSH4:   "Facility",
			wantFields: 4,
		},
		{
			name:       "minimal MSH",
			data:       "MSH|^~\\&",
			wantMSH1:   "|",
			wantMSH2:   "^~\\&",
			wantFields: 2,
		},
		{
			name:       "MSH with empty fields",
			data:       "MSH|^~\\&|App||Fac",
			wantMSH1:   "|",
			wantMSH2:   "^~\\&",
			wantMSH3:   "App",
			wantMSH4:   "",
			wantFields: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := ParseSegment([]rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseSegment() error = %v", err)
			}

			if seg.Name() != "MSH" {
				t.Errorf("Name() = %q, want %q", seg.Name(), "MSH")
			}

			if seg.FieldCount() != tt.wantFields {
				t.Errorf("FieldCount() = %d, want %d", seg.FieldCount(), tt.wantFields)
			}

			// MSH-1: Field separator
			msh1, ok := seg.Field(1)
			if !ok {
				t.Error("MSH-1 not found")
			} else if msh1.Value() != tt.wantMSH1 {
				t.Errorf("MSH-1 = %q, want %q", msh1.Value(), tt.wantMSH1)
			}

			// MSH-2: Encoding characters
			if tt.wantFields >= 2 {
				msh2, ok := seg.Field(2)
				if !ok {
					t.Error("MSH-2 not found")
				} else if msh2.Value() != tt.wantMSH2 {
					t.Errorf("MSH-2 = %q, want %q", msh2.Value(), tt.wantMSH2)
				}
			}

			// MSH-3: Sending Application
			if tt.wantFields >= 3 && tt.wantMSH3 != "" {
				msh3, ok := seg.Field(3)
				if !ok {
					t.Error("MSH-3 not found")
				} else if msh3.Value() != tt.wantMSH3 {
					t.Errorf("MSH-3 = %q, want %q", msh3.Value(), tt.wantMSH3)
				}
			}

			// MSH-4: Sending Facility
			if tt.wantFields >= 4 {
				msh4, ok := seg.Field(4)
				if !ok {
					t.Error("MSH-4 not found")
				} else if msh4.Value() != tt.wantMSH4 {
					t.Errorf("MSH-4 = %q, want %q", msh4.Value(), tt.wantMSH4)
				}
			}
		})
	}
}

func TestSegment_FieldAccess(t *testing.T) {
	delims := DefaultDelimiters()
	data := "PID|1|12345|PatientID^^^Hospital|Name^Given"
	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	t.Run("valid field access", func(t *testing.T) {
		f, ok := seg.Field(1)
		if !ok {
			t.Error("Field(1) should exist")
		}
		if f.Value() != "1" {
			t.Errorf("Field(1).Value() = %q, want %q", f.Value(), "1")
		}
	})

	t.Run("field with components", func(t *testing.T) {
		f, ok := seg.Field(4)
		if !ok {
			t.Error("Field(4) should exist")
		}
		// Value() returns full encoded value including components
		if f.Value() != "Name^Given" {
			t.Errorf("Field(4).Value() = %q, want %q", f.Value(), "Name^Given")
		}

		// Check component access
		comp, ok := f.Component(1)
		if !ok {
			t.Error("Component(1) should exist")
		}
		if comp.Value() != "Name" {
			t.Errorf("Component(1).Value() = %q, want %q", comp.Value(), "Name")
		}

		comp2, ok := f.Component(2)
		if !ok {
			t.Error("Component(2) should exist")
		}
		if comp2.Value() != "Given" {
			t.Errorf("Component(2).Value() = %q, want %q", comp2.Value(), "Given")
		}
	})

	t.Run("non-existent field", func(t *testing.T) {
		_, ok := seg.Field(99)
		if ok {
			t.Error("Field(99) should not exist")
		}
	})

	t.Run("field 0 not valid", func(t *testing.T) {
		_, ok := seg.Field(0)
		if ok {
			t.Error("Field(0) should not exist (1-based indexing)")
		}
	})

	t.Run("negative field not valid", func(t *testing.T) {
		_, ok := seg.Field(-1)
		if ok {
			t.Error("Field(-1) should not exist")
		}
	})
}

func TestSegment_Get(t *testing.T) {
	delims := DefaultDelimiters()
	data := "PID|1|12345|PatientID^^^Hospital&Dept|Last^First^Middle"
	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	tests := []struct {
		name     string
		location string
		want     string
		wantErr  bool
	}{
		{
			name:     "field only",
			location: ".1",
			want:     "1",
		},
		{
			name:     "field only with leading dot",
			location: ".2",
			want:     "12345",
		},
		{
			name:     "field and component",
			location: ".4.1",
			want:     "Last",
		},
		{
			name:     "field and second component",
			location: ".4.2",
			want:     "First",
		},
		{
			name:     "field and third component",
			location: ".4.3",
			want:     "Middle",
		},
		{
			name:     "field, component, subcomponent",
			location: ".3.4.1",
			want:     "Hospital",
		},
		{
			name:     "field, component, second subcomponent",
			location: ".3.4.2",
			want:     "Dept",
		},
		{
			name:     "non-existent field returns empty",
			location: ".99",
			want:     "",
		},
		{
			name:     "non-existent component returns empty",
			location: ".1.99",
			want:     "",
		},
		{
			name:     "empty location error",
			location: "",
			wantErr:  true,
		},
		{
			name:     "just dot error",
			location: ".",
			wantErr:  true,
		},
		{
			name:     "invalid field number",
			location: ".abc",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := seg.Get(tt.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get(%q) error = %v, wantErr %v", tt.location, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.location, got, tt.want)
			}
		})
	}
}

func TestSegment_Set(t *testing.T) {
	t.Run("set existing field", func(t *testing.T) {
		seg := NewSegment("PID")
		_ = seg.AddField(NewField(1, "1"))
		_ = seg.AddField(NewField(2, "original"))

		err := seg.Set(".2", "modified")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		got, _ := seg.Get(".2")
		if got != "modified" {
			t.Errorf("Get(.2) = %q, want %q", got, "modified")
		}
	})

	t.Run("set new field expands segment", func(t *testing.T) {
		seg := NewSegment("PID")

		err := seg.Set(".5", "value")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if seg.FieldCount() != 5 {
			t.Errorf("FieldCount() = %d, want 5", seg.FieldCount())
		}

		got, _ := seg.Get(".5")
		if got != "value" {
			t.Errorf("Get(.5) = %q, want %q", got, "value")
		}
	})

	t.Run("set component within field", func(t *testing.T) {
		seg := NewSegment("PID")
		_ = seg.AddField(NewField(1, "1"))

		err := seg.Set(".1.2", "component2")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		got, _ := seg.Get(".1.2")
		if got != "component2" {
			t.Errorf("Get(.1.2) = %q, want %q", got, "component2")
		}
	})
}

func TestSegment_SetField(t *testing.T) {
	t.Run("set field at position", func(t *testing.T) {
		seg := NewSegment("PID")

		err := seg.SetField(3, NewField(3, "third"))
		if err != nil {
			t.Fatalf("SetField() error = %v", err)
		}

		if seg.FieldCount() != 3 {
			t.Errorf("FieldCount() = %d, want 3", seg.FieldCount())
		}

		f, ok := seg.Field(3)
		if !ok {
			t.Error("Field(3) should exist")
		}
		if f.Value() != "third" {
			t.Errorf("Field(3).Value() = %q, want %q", f.Value(), "third")
		}
	})

	t.Run("invalid sequence number", func(t *testing.T) {
		seg := NewSegment("PID")

		err := seg.SetField(0, NewField(0, "invalid"))
		if err == nil {
			t.Error("SetField(0, ...) should return error")
		}
	})
}

func TestSegment_Bytes_Regular(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "simple segment",
			data: "PID|1|12345|Name",
			want: "PID|1|12345|Name",
		},
		{
			name: "segment with components",
			data: "PID|1|ID^Type|Name^Given",
			want: "PID|1|ID^Type|Name^Given",
		},
		{
			name: "segment name only",
			data: "EVN",
			want: "EVN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := ParseSegment([]rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseSegment() error = %v", err)
			}

			got := string(seg.Bytes(delims))
			if got != tt.want {
				t.Errorf("Bytes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSegment_Bytes_MSH(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "standard MSH",
			data: "MSH|^~\\&|SendingApp|SendingFac",
			want: "MSH|^~\\&|SendingApp|SendingFac",
		},
		{
			name: "minimal MSH",
			data: "MSH|^~\\&",
			want: "MSH|^~\\&",
		},
		{
			name: "MSH with many fields",
			data: "MSH|^~\\&|App|Fac|RecvApp|RecvFac|20240101120000||ADT^A01|123|P|2.5",
			want: "MSH|^~\\&|App|Fac|RecvApp|RecvFac|20240101120000||ADT^A01|123|P|2.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg, err := ParseSegment([]rune(tt.data), delims)
			if err != nil {
				t.Fatalf("ParseSegment() error = %v", err)
			}

			got := string(seg.Bytes(delims))
			if got != tt.want {
				t.Errorf("Bytes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSegment_MSH_FieldNumbering(t *testing.T) {
	delims := DefaultDelimiters()
	data := "MSH|^~\\&|SendApp|SendFac|RecvApp"

	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	tests := []struct {
		field int
		want  string
	}{
		{1, "|"},       // MSH-1: Field separator
		{2, "^~\\&"},   // MSH-2: Encoding characters
		{3, "SendApp"}, // MSH-3: Sending Application
		{4, "SendFac"}, // MSH-4: Sending Facility
		{5, "RecvApp"}, // MSH-5: Receiving Application
	}

	for _, tt := range tests {
		t.Run("MSH-"+string(rune('0'+tt.field)), func(t *testing.T) {
			f, ok := seg.Field(tt.field)
			if !ok {
				t.Errorf("Field(%d) not found", tt.field)
				return
			}
			if f.Value() != tt.want {
				t.Errorf("Field(%d).Value() = %q, want %q", tt.field, f.Value(), tt.want)
			}
		})
	}

	// Also test via Get
	t.Run("Get MSH-3", func(t *testing.T) {
		got, err := seg.Get(".3")
		if err != nil {
			t.Errorf("Get(.3) error = %v", err)
		}
		if got != "SendApp" {
			t.Errorf("Get(.3) = %q, want %q", got, "SendApp")
		}
	})
}

func TestSegment_EmptyFields(t *testing.T) {
	delims := DefaultDelimiters()

	t.Run("middle empty fields", func(t *testing.T) {
		data := "PID|1||3||5"
		seg, err := ParseSegment([]rune(data), delims)
		if err != nil {
			t.Fatalf("ParseSegment() error = %v", err)
		}

		if seg.FieldCount() != 5 {
			t.Errorf("FieldCount() = %d, want 5", seg.FieldCount())
		}

		f2, ok := seg.Field(2)
		if !ok {
			t.Error("Field(2) should exist")
		}
		if f2.Value() != "" {
			t.Errorf("Field(2).Value() = %q, want empty", f2.Value())
		}

		f3, ok := seg.Field(3)
		if !ok {
			t.Error("Field(3) should exist")
		}
		if f3.Value() != "3" {
			t.Errorf("Field(3).Value() = %q, want %q", f3.Value(), "3")
		}
	})

	t.Run("trailing empty field", func(t *testing.T) {
		data := "PID|1|2|"
		seg, err := ParseSegment([]rune(data), delims)
		if err != nil {
			t.Fatalf("ParseSegment() error = %v", err)
		}

		if seg.FieldCount() != 3 {
			t.Errorf("FieldCount() = %d, want 3", seg.FieldCount())
		}

		f3, ok := seg.Field(3)
		if !ok {
			t.Error("Field(3) should exist")
		}
		if f3.Value() != "" {
			t.Errorf("Field(3).Value() = %q, want empty", f3.Value())
		}
	})
}

func TestSegment_String(t *testing.T) {
	delims := DefaultDelimiters()
	data := "PID|1|PatientID|Name"

	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	got := seg.String()
	if got != data {
		t.Errorf("String() = %q, want %q", got, data)
	}
}

func TestSegment_AllFields(t *testing.T) {
	delims := DefaultDelimiters()
	data := "PID|1|2|3"

	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	fields := seg.AllFields()
	if len(fields) != 3 {
		t.Errorf("len(AllFields()) = %d, want 3", len(fields))
	}

	// Verify it returns a copy
	fields[0] = NewField(1, "modified")
	f1, _ := seg.Field(1)
	if f1.Value() == "modified" {
		t.Error("AllFields() should return a copy, not modify original")
	}
}

func TestSegment_Fields(t *testing.T) {
	delims := DefaultDelimiters()
	data := "PID|1|2|3"

	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	t.Run("existing field", func(t *testing.T) {
		fields := seg.Fields(1)
		if len(fields) != 1 {
			t.Errorf("len(Fields(1)) = %d, want 1", len(fields))
		}
	})

	t.Run("non-existent field", func(t *testing.T) {
		fields := seg.Fields(99)
		if fields != nil {
			t.Errorf("Fields(99) = %v, want nil", fields)
		}
	})
}

func TestSegment_GetAll(t *testing.T) {
	delims := DefaultDelimiters()
	data := "PID|1|ID1|Name^Given"

	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	t.Run("field only", func(t *testing.T) {
		results, err := seg.GetAll(".1")
		if err != nil {
			t.Errorf("GetAll(.1) error = %v", err)
		}
		if len(results) != 1 || results[0] != "1" {
			t.Errorf("GetAll(.1) = %v, want [\"1\"]", results)
		}
	})

	t.Run("field with component", func(t *testing.T) {
		results, err := seg.GetAll(".3.1")
		if err != nil {
			t.Errorf("GetAll(.3.1) error = %v", err)
		}
		if len(results) != 1 || results[0] != "Name" {
			t.Errorf("GetAll(.3.1) = %v, want [\"Name\"]", results)
		}
	})

	t.Run("non-existent field", func(t *testing.T) {
		results, err := seg.GetAll(".99")
		if err != nil {
			t.Errorf("GetAll(.99) error = %v", err)
		}
		if results != nil {
			t.Errorf("GetAll(.99) = %v, want nil", results)
		}
	})
}

func TestSegment_AddField(t *testing.T) {
	seg := NewSegment("PID")

	err := seg.AddField(NewField(1, "first"))
	if err != nil {
		t.Fatalf("AddField() error = %v", err)
	}

	err = seg.AddField(NewField(2, "second"))
	if err != nil {
		t.Fatalf("AddField() error = %v", err)
	}

	if seg.FieldCount() != 2 {
		t.Errorf("FieldCount() = %d, want 2", seg.FieldCount())
	}

	f1, _ := seg.Field(1)
	if f1.Value() != "first" {
		t.Errorf("Field(1).Value() = %q, want %q", f1.Value(), "first")
	}

	f2, _ := seg.Field(2)
	if f2.Value() != "second" {
		t.Errorf("Field(2).Value() = %q, want %q", f2.Value(), "second")
	}
}

func TestParseSegment_Errors(t *testing.T) {
	delims := DefaultDelimiters()

	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    "",
			wantErr: true,
		},
		{
			name:    "too short segment name",
			data:    "PI",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSegment([]rune(tt.data), delims)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSegment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSegment_ComponentsAndSubcomponents(t *testing.T) {
	delims := DefaultDelimiters()
	// Use proper component^subcomponent structure for testing
	// Field 2 has a component with subcomponents: "Comp1^ID&SubID"
	// Field 3 has multiple components: "Name^Given^Middle"
	data := "PID|1|Comp1^ID&SubID|Name^Given^Middle"

	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	t.Run("subcomponent access", func(t *testing.T) {
		// Component 2 has subcomponents: "ID&SubID"
		val, err := seg.Get(".2.2.1")
		if err != nil {
			t.Errorf("Get(.2.2.1) error = %v", err)
		}
		if val != "ID" {
			t.Errorf("Get(.2.2.1) = %q, want %q", val, "ID")
		}

		val, err = seg.Get(".2.2.2")
		if err != nil {
			t.Errorf("Get(.2.2.2) error = %v", err)
		}
		if val != "SubID" {
			t.Errorf("Get(.2.2.2) = %q, want %q", val, "SubID")
		}
	})

	t.Run("multiple components", func(t *testing.T) {
		f, ok := seg.Field(3)
		if !ok {
			t.Fatal("Field(3) not found")
		}

		comps := f.Components()
		if len(comps) != 3 {
			t.Errorf("len(Components()) = %d, want 3", len(comps))
		}
	})
}

func TestSegment_NilDelimiters(t *testing.T) {
	// Should use default delimiters when nil is passed
	data := "PID|1|2|3"

	seg, err := ParseSegment([]rune(data), nil)
	if err != nil {
		t.Fatalf("ParseSegment() with nil delims error = %v", err)
	}

	if seg.FieldCount() != 3 {
		t.Errorf("FieldCount() = %d, want 3", seg.FieldCount())
	}

	// Bytes should also handle nil
	got := string(seg.Bytes(nil))
	if got != data {
		t.Errorf("Bytes(nil) = %q, want %q", got, data)
	}
}

func TestSegment_MSH_Components(t *testing.T) {
	delims := DefaultDelimiters()
	// MSH-9 typically has components like ADT^A01
	data := "MSH|^~\\&|SendApp|SendFac|RecvApp|RecvFac|20240101120000||ADT^A01|123|P|2.5"

	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	// MSH-9 should be "ADT^A01" - Value() returns full encoded value
	msh9, ok := seg.Field(9)
	if !ok {
		t.Fatal("MSH-9 not found")
	}

	if msh9.Value() != "ADT^A01" {
		t.Errorf("MSH-9.Value() = %q, want %q", msh9.Value(), "ADT^A01")
	}

	// MSH-9.1 should be "ADT"
	val, err := seg.Get(".9.1")
	if err != nil {
		t.Errorf("Get(.9.1) error = %v", err)
	}
	if val != "ADT" {
		t.Errorf("Get(.9.1) = %q, want %q", val, "ADT")
	}

	// MSH-9.2 should be "A01"
	val, err = seg.Get(".9.2")
	if err != nil {
		t.Errorf("Get(.9.2) error = %v", err)
	}
	if val != "A01" {
		t.Errorf("Get(.9.2) = %q, want %q", val, "A01")
	}
}

func TestSegment_FieldWithRepetitions(t *testing.T) {
	delims := DefaultDelimiters()
	// PID-3 can have multiple patient IDs separated by ~
	data := "PID|1||ID1~ID2~ID3||Name"

	seg, err := ParseSegment([]rune(data), delims)
	if err != nil {
		t.Fatalf("ParseSegment() error = %v", err)
	}

	f3, ok := seg.Field(3)
	if !ok {
		t.Fatal("Field(3) not found")
	}

	// Check repetition count
	if f3.RepetitionCount() != 3 {
		t.Errorf("RepetitionCount() = %d, want 3", f3.RepetitionCount())
	}

	// Check individual repetitions
	rep0, ok := f3.Repetition(0)
	if !ok || rep0.Value() != "ID1" {
		t.Errorf("Repetition(0).Value() = %q, want %q", rep0.Value(), "ID1")
	}

	rep1, ok := f3.Repetition(1)
	if !ok || rep1.Value() != "ID2" {
		t.Errorf("Repetition(1).Value() = %q, want %q", rep1.Value(), "ID2")
	}

	rep2, ok := f3.Repetition(2)
	if !ok || rep2.Value() != "ID3" {
		t.Errorf("Repetition(2).Value() = %q, want %q", rep2.Value(), "ID3")
	}
}

func TestParseSegmentLocation(t *testing.T) {
	tests := []struct {
		name             string
		location         string
		wantField        int
		wantComponent    int
		wantSubcomponent int
		wantErr          bool
	}{
		{
			name:      "field only",
			location:  ".5",
			wantField: 5,
		},
		{
			name:      "field without leading dot",
			location:  "5",
			wantField: 5,
		},
		{
			name:          "field and component",
			location:      ".5.1",
			wantField:     5,
			wantComponent: 1,
		},
		{
			name:             "field, component, subcomponent",
			location:         ".5.1.2",
			wantField:        5,
			wantComponent:    1,
			wantSubcomponent: 2,
		},
		{
			name:     "empty location",
			location: "",
			wantErr:  true,
		},
		{
			name:     "just dot",
			location: ".",
			wantErr:  true,
		},
		{
			name:     "invalid field",
			location: ".abc",
			wantErr:  true,
		},
		{
			name:     "negative field",
			location: ".-1",
			wantErr:  true,
		},
		{
			name:     "zero field",
			location: ".0",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := parseSegmentLocation(tt.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSegmentLocation(%q) error = %v, wantErr %v", tt.location, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if loc.field != tt.wantField {
				t.Errorf("field = %d, want %d", loc.field, tt.wantField)
			}
			if loc.component != tt.wantComponent {
				t.Errorf("component = %d, want %d", loc.component, tt.wantComponent)
			}
			if loc.subcomponent != tt.wantSubcomponent {
				t.Errorf("subcomponent = %d, want %d", loc.subcomponent, tt.wantSubcomponent)
			}
		})
	}
}
