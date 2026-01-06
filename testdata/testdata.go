// Package testdata provides embedded HL7 test messages for testing the golevel7v2 library.
package testdata

import (
	"embed"
	"fmt"
	"path"
)

//go:embed *.hl7 malformed/*.hl7
var FS embed.FS

// Message file names
const (
	FileADTA01            = "adt_a01.hl7"
	FileORUR01            = "oru_r01.hl7"
	FileORMO01            = "orm_o01.hl7"
	FileADTA08            = "adt_a08.hl7"
	FileACKAA             = "ack_aa.hl7"
	FileComplex           = "complex.hl7"
	FileMissingMSH        = "malformed/missing_msh.hl7"
	FileEmpty             = "malformed/empty.hl7"
	FileInvalidDelimiters = "malformed/invalid_delimiters.hl7"
	FileTruncated         = "malformed/truncated.hl7"
)

// LoadADTA01 loads the ADT^A01 (Patient Admit) test message.
func LoadADTA01() ([]byte, error) {
	return FS.ReadFile(FileADTA01)
}

// LoadORUR01 loads the ORU^R01 (Observation Result) test message.
func LoadORUR01() ([]byte, error) {
	return FS.ReadFile(FileORUR01)
}

// LoadORMO01 loads the ORM^O01 (Order) test message.
func LoadORMO01() ([]byte, error) {
	return FS.ReadFile(FileORMO01)
}

// LoadADTA08 loads the ADT^A08 (Patient Update) test message.
func LoadADTA08() ([]byte, error) {
	return FS.ReadFile(FileADTA08)
}

// LoadACKAA loads the ACK (Application Accept) test message.
func LoadACKAA() ([]byte, error) {
	return FS.ReadFile(FileACKAA)
}

// LoadComplex loads the complex test message with repetitions,
// components, subcomponents, and escape sequences.
func LoadComplex() ([]byte, error) {
	return FS.ReadFile(FileComplex)
}

// LoadMissingMSH loads a malformed message without an MSH segment.
func LoadMissingMSH() ([]byte, error) {
	return FS.ReadFile(FileMissingMSH)
}

// LoadEmpty loads an empty file for testing empty input handling.
func LoadEmpty() ([]byte, error) {
	return FS.ReadFile(FileEmpty)
}

// LoadInvalidDelimiters loads a message with invalid MSH-2 delimiters.
func LoadInvalidDelimiters() ([]byte, error) {
	return FS.ReadFile(FileInvalidDelimiters)
}

// LoadTruncated loads a truncated/incomplete message.
func LoadTruncated() ([]byte, error) {
	return FS.ReadFile(FileTruncated)
}

// LoadFile loads any test file by name from the embedded filesystem.
func LoadFile(name string) ([]byte, error) {
	data, err := FS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("loading test file %s: %w", name, err)
	}
	return data, nil
}

// MustLoad loads a test file and panics on error.
// Useful for test setup where failure should halt the test.
func MustLoad(name string) []byte {
	data, err := LoadFile(name)
	if err != nil {
		panic(err)
	}
	return data
}

// ListFiles returns a list of all embedded test file names.
func ListFiles() ([]string, error) {
	var files []string

	// Read root directory
	entries, err := FS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("reading root directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Read subdirectory
			subEntries, err := FS.ReadDir(entry.Name())
			if err != nil {
				return nil, fmt.Errorf("reading directory %s: %w", entry.Name(), err)
			}
			for _, subEntry := range subEntries {
				if !subEntry.IsDir() {
					files = append(files, path.Join(entry.Name(), subEntry.Name()))
				}
			}
		} else {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ListMalformedFiles returns a list of malformed test file names.
func ListMalformedFiles() ([]string, error) {
	entries, err := FS.ReadDir("malformed")
	if err != nil {
		return nil, fmt.Errorf("reading malformed directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, path.Join("malformed", entry.Name()))
		}
	}

	return files, nil
}

// ListValidFiles returns a list of valid (non-malformed) test file names.
func ListValidFiles() ([]string, error) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("reading root directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
