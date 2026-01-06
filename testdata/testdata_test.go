package testdata_test

import (
	"bytes"
	"testing"

	"github.com/dshills/golevel7/testdata"
)

func TestLoadADTA01(t *testing.T) {
	data, err := testdata.LoadADTA01()
	if err != nil {
		t.Fatalf("LoadADTA01() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("LoadADTA01() returned empty data")
	}
	if !bytes.HasPrefix(data, []byte("MSH|^~\\&|")) {
		t.Error("LoadADTA01() message does not start with expected MSH segment")
	}
	// Verify CR line endings
	if !bytes.Contains(data, []byte("\r")) {
		t.Error("LoadADTA01() message missing CR line endings")
	}
	// Should contain ADT^A01
	if !bytes.Contains(data, []byte("ADT^A01")) {
		t.Error("LoadADTA01() message does not contain ADT^A01")
	}
}

func TestLoadORUR01(t *testing.T) {
	data, err := testdata.LoadORUR01()
	if err != nil {
		t.Fatalf("LoadORUR01() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("LoadORUR01() returned empty data")
	}
	// Should contain ORU^R01
	if !bytes.Contains(data, []byte("ORU^R01")) {
		t.Error("LoadORUR01() message does not contain ORU^R01")
	}
	// Should contain multiple OBX segments
	obxCount := bytes.Count(data, []byte("\rOBX|"))
	if obxCount < 2 {
		t.Errorf("LoadORUR01() expected multiple OBX segments, got %d", obxCount)
	}
}

func TestLoadORMO01(t *testing.T) {
	data, err := testdata.LoadORMO01()
	if err != nil {
		t.Fatalf("LoadORMO01() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("LoadORMO01() returned empty data")
	}
	// Should contain ORM^O01
	if !bytes.Contains(data, []byte("ORM^O01")) {
		t.Error("LoadORMO01() message does not contain ORM^O01")
	}
	// Should contain ORC segment
	if !bytes.Contains(data, []byte("\rORC|")) {
		t.Error("LoadORMO01() message does not contain ORC segment")
	}
}

func TestLoadADTA08(t *testing.T) {
	data, err := testdata.LoadADTA08()
	if err != nil {
		t.Fatalf("LoadADTA08() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("LoadADTA08() returned empty data")
	}
	// Should contain ADT^A08
	if !bytes.Contains(data, []byte("ADT^A08")) {
		t.Error("LoadADTA08() message does not contain ADT^A08")
	}
}

func TestLoadACKAA(t *testing.T) {
	data, err := testdata.LoadACKAA()
	if err != nil {
		t.Fatalf("LoadACKAA() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("LoadACKAA() returned empty data")
	}
	// Should contain ACK message type
	if !bytes.Contains(data, []byte("||ACK|")) {
		t.Error("LoadACKAA() message does not contain ACK message type")
	}
	// Should contain MSA segment with AA
	if !bytes.Contains(data, []byte("\rMSA|AA|")) {
		t.Error("LoadACKAA() message does not contain MSA|AA segment")
	}
}

func TestLoadComplex(t *testing.T) {
	data, err := testdata.LoadComplex()
	if err != nil {
		t.Fatalf("LoadComplex() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("LoadComplex() returned empty data")
	}
	// Should contain repetitions (tilde)
	if !bytes.Contains(data, []byte("~")) {
		t.Error("LoadComplex() message does not contain repetition separators")
	}
	// Should contain escape sequences
	if !bytes.Contains(data, []byte("\\T\\")) {
		t.Error("LoadComplex() message does not contain \\T\\ escape sequence")
	}
	if !bytes.Contains(data, []byte("\\F\\")) {
		t.Error("LoadComplex() message does not contain \\F\\ escape sequence")
	}
	// Should contain subcomponents (ampersand)
	if !bytes.Contains(data, []byte("&")) {
		t.Error("LoadComplex() message does not contain subcomponent separators")
	}
}

func TestLoadMalformedFiles(t *testing.T) {
	tests := []struct {
		name     string
		loadFunc func() ([]byte, error)
	}{
		{"MissingMSH", testdata.LoadMissingMSH},
		{"Empty", testdata.LoadEmpty},
		{"InvalidDelimiters", testdata.LoadInvalidDelimiters},
		{"Truncated", testdata.LoadTruncated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.loadFunc()
			if err != nil {
				t.Fatalf("Load%s() error = %v", tt.name, err)
			}
			// Files should load without error (validation is separate)
			_ = data
		})
	}
}

func TestMissingMSHContent(t *testing.T) {
	data, err := testdata.LoadMissingMSH()
	if err != nil {
		t.Fatalf("LoadMissingMSH() error = %v", err)
	}
	// Should NOT start with MSH
	if bytes.HasPrefix(data, []byte("MSH|")) {
		t.Error("LoadMissingMSH() should not start with MSH segment")
	}
	// Should start with PID
	if !bytes.HasPrefix(data, []byte("PID|")) {
		t.Error("LoadMissingMSH() should start with PID segment")
	}
}

func TestEmptyContent(t *testing.T) {
	data, err := testdata.LoadEmpty()
	if err != nil {
		t.Fatalf("LoadEmpty() error = %v", err)
	}
	if len(data) != 0 {
		t.Errorf("LoadEmpty() expected empty data, got %d bytes", len(data))
	}
}

func TestInvalidDelimitersContent(t *testing.T) {
	data, err := testdata.LoadInvalidDelimiters()
	if err != nil {
		t.Fatalf("LoadInvalidDelimiters() error = %v", err)
	}
	// Should start with MSH but have invalid delimiter field
	if !bytes.HasPrefix(data, []byte("MSH|INVALID|")) {
		t.Error("LoadInvalidDelimiters() should have INVALID in MSH-2 position")
	}
}

func TestTruncatedContent(t *testing.T) {
	data, err := testdata.LoadTruncated()
	if err != nil {
		t.Fatalf("LoadTruncated() error = %v", err)
	}
	// Should not end with CR (incomplete)
	if len(data) > 0 && data[len(data)-1] == '\r' {
		t.Error("LoadTruncated() should not end with CR (message is truncated)")
	}
}

func TestListFiles(t *testing.T) {
	files, err := testdata.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if len(files) < 10 {
		t.Errorf("ListFiles() expected at least 10 files, got %d", len(files))
	}
	// Should include both valid and malformed files
	foundValid := false
	foundMalformed := false
	for _, f := range files {
		if f == "adt_a01.hl7" {
			foundValid = true
		}
		if f == "malformed/missing_msh.hl7" {
			foundMalformed = true
		}
	}
	if !foundValid {
		t.Error("ListFiles() missing adt_a01.hl7")
	}
	if !foundMalformed {
		t.Error("ListFiles() missing malformed/missing_msh.hl7")
	}
}

func TestListValidFiles(t *testing.T) {
	files, err := testdata.ListValidFiles()
	if err != nil {
		t.Fatalf("ListValidFiles() error = %v", err)
	}
	if len(files) < 6 {
		t.Errorf("ListValidFiles() expected at least 6 files, got %d", len(files))
	}
	for _, f := range files {
		if bytes.HasPrefix([]byte(f), []byte("malformed/")) {
			t.Errorf("ListValidFiles() returned malformed file: %s", f)
		}
	}
}

func TestListMalformedFiles(t *testing.T) {
	files, err := testdata.ListMalformedFiles()
	if err != nil {
		t.Fatalf("ListMalformedFiles() error = %v", err)
	}
	if len(files) < 4 {
		t.Errorf("ListMalformedFiles() expected at least 4 files, got %d", len(files))
	}
	for _, f := range files {
		if !bytes.HasPrefix([]byte(f), []byte("malformed/")) {
			t.Errorf("ListMalformedFiles() returned non-malformed file: %s", f)
		}
	}
}

func TestMustLoad(t *testing.T) {
	// Should not panic for valid file
	data := testdata.MustLoad(testdata.FileADTA01)
	if len(data) == 0 {
		t.Error("MustLoad() returned empty data")
	}
}

func TestMustLoadPanicsOnInvalidFile(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoad() expected panic for invalid file")
		}
	}()
	testdata.MustLoad("nonexistent.hl7")
}

func TestLoadFile(t *testing.T) {
	data, err := testdata.LoadFile(testdata.FileADTA01)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("LoadFile() returned empty data")
	}
}

func TestLoadFileError(t *testing.T) {
	_, err := testdata.LoadFile("nonexistent.hl7")
	if err == nil {
		t.Error("LoadFile() expected error for nonexistent file")
	}
}

// TestCRLineEndings verifies all valid messages use CR (0x0D) as segment terminator
func TestCRLineEndings(t *testing.T) {
	files, err := testdata.ListValidFiles()
	if err != nil {
		t.Fatalf("ListValidFiles() error = %v", err)
	}

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			data, err := testdata.LoadFile(f)
			if err != nil {
				t.Fatalf("LoadFile(%s) error = %v", f, err)
			}
			if len(data) == 0 {
				t.Skip("empty file")
			}
			// Should contain CR
			if !bytes.Contains(data, []byte{0x0D}) {
				t.Errorf("file %s missing CR (0x0D) line endings", f)
			}
			// Should NOT contain LF (0x0A)
			if bytes.Contains(data, []byte{0x0A}) {
				t.Errorf("file %s contains LF (0x0A), should only have CR", f)
			}
		})
	}
}
