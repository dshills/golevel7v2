package escape

import (
	"testing"

	"github.com/dshills/golevel7/hl7"
)

func TestNew(t *testing.T) {
	t.Run("with nil delimiters uses defaults", func(t *testing.T) {
		e := New(nil)
		if e == nil {
			t.Fatal("New(nil) returned nil")
			return
		}
		if e.delims == nil {
			t.Fatal("Escaper has nil delimiters")
			return
		}
		if e.delims.Field != '|' {
			t.Errorf("expected field delimiter '|', got %q", e.delims.Field)
		}
	})

	t.Run("with custom delimiters", func(t *testing.T) {
		custom := &hl7.Delimiters{
			Field:        '#',
			Component:    '@',
			Repetition:   '!',
			Escape:       '~',
			SubComponent: '%',
		}
		e := New(custom)
		if e.delims.Field != '#' {
			t.Errorf("expected field delimiter '#', got %q", e.delims.Field)
		}
		if e.delims.Escape != '~' {
			t.Errorf("expected escape character '~', got %q", e.delims.Escape)
		}
	})
}

func TestEscape(t *testing.T) {
	e := New(hl7.DefaultDelimiters())

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no special characters",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "field separator",
			input: "Hello|World",
			want:  `Hello\F\World`,
		},
		{
			name:  "component separator",
			input: "Hello^World",
			want:  `Hello\S\World`,
		},
		{
			name:  "subcomponent separator",
			input: "Hello&World",
			want:  `Hello\T\World`,
		},
		{
			name:  "repetition separator",
			input: "Hello~World",
			want:  `Hello\R\World`,
		},
		{
			name:  "escape character",
			input: `Hello\World`,
			want:  `Hello\E\World`,
		},
		{
			name:  "multiple special characters",
			input: "A|B^C&D~E",
			want:  `A\F\B\S\C\T\D\R\E`,
		},
		{
			name:  "all delimiters",
			input: `|^&~\`,
			want:  `\F\\S\\T\\R\\E\`,
		},
		{
			name:  "special chars at start and end",
			input: "|text|",
			want:  `\F\text\F\`,
		},
		{
			name:  "consecutive special chars",
			input: "||^^",
			want:  `\F\\F\\S\\S\`,
		},
		{
			name:  "unicode characters preserved",
			input: "Hello|World\u00e9",
			want:  `Hello\F\World` + "\u00e9",
		},
		{
			name:  "numbers and punctuation",
			input: "123-456|789",
			want:  `123-456\F\789`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Escape(tt.input)
			if got != tt.want {
				t.Errorf("Escape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscape_CustomDelimiters(t *testing.T) {
	custom := &hl7.Delimiters{
		Field:        '#',
		Component:    '@',
		Repetition:   '!',
		Escape:       '~',
		SubComponent: '%',
	}
	e := New(custom)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "custom field separator",
			input: "Hello#World",
			want:  "Hello~F~World",
		},
		{
			name:  "custom component separator",
			input: "Hello@World",
			want:  "Hello~S~World",
		},
		{
			name:  "custom escape character",
			input: "Hello~World",
			want:  "Hello~E~World",
		},
		{
			name:  "standard chars not escaped",
			input: "Hello|^&World",
			want:  "Hello|^&World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Escape(tt.input)
			if got != tt.want {
				t.Errorf("Escape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUnescape(t *testing.T) {
	e := New(hl7.DefaultDelimiters())

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no escape sequences",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "field separator",
			input: `Hello\F\World`,
			want:  "Hello|World",
		},
		{
			name:  "component separator",
			input: `Hello\S\World`,
			want:  "Hello^World",
		},
		{
			name:  "subcomponent separator",
			input: `Hello\T\World`,
			want:  "Hello&World",
		},
		{
			name:  "repetition separator",
			input: `Hello\R\World`,
			want:  "Hello~World",
		},
		{
			name:  "escape character",
			input: `Hello\E\World`,
			want:  `Hello\World`,
		},
		{
			name:  "multiple escape sequences",
			input: `A\F\B\S\C\T\D\R\E`,
			want:  "A|B^C&D~E",
		},
		{
			name:  "all standard escapes",
			input: `\F\\S\\T\\R\\E\`,
			want:  `|^&~\`,
		},
		{
			name:  "line break",
			input: `Hello\.br\World`,
			want:  "Hello\nWorld",
		},
		{
			name:  "skip line",
			input: `Line1\.sk\Line2`,
			want:  "Line1\nLine2",
		},
		{
			name:  "spacing",
			input: `Hello\.sp\World`,
			want:  "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Unescape(tt.input)
			if got != tt.want {
				t.Errorf("Unescape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUnescape_HexEncoding(t *testing.T) {
	e := New(hl7.DefaultDelimiters())

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single byte hex",
			input: `\X41\`,
			want:  "A",
		},
		{
			name:  "lowercase hex",
			input: `\x41\`,
			want:  "A",
		},
		{
			name:  "multiple bytes hex",
			input: `\X48454C4C4F\`,
			want:  "HELLO",
		},
		{
			name:  "hex with surrounding text",
			input: `Start\X2D\End`,
			want:  "Start-End",
		},
		{
			name:  "space as hex",
			input: `Hello\X20\World`,
			want:  "Hello World",
		},
		{
			name:  "newline as hex",
			input: `Line1\X0A\Line2`,
			want:  "Line1\nLine2",
		},
		{
			name:  "carriage return as hex",
			input: `Line1\X0D\Line2`,
			want:  "Line1\rLine2",
		},
		{
			name:  "multiple hex sequences",
			input: `\X41\\X42\\X43\`,
			want:  "ABC",
		},
		{
			name:  "unicode character hex",
			input: `\XC3A9\`,
			want:  "\u00e9", // e with acute accent (UTF-8: C3 A9)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Unescape(tt.input)
			if got != tt.want {
				t.Errorf("Unescape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUnescape_MalformedSequences(t *testing.T) {
	e := New(hl7.DefaultDelimiters())

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unclosed escape at end",
			input: `Hello\F`,
			want:  `Hello\F`,
		},
		{
			name:  "single escape character",
			input: `Hello\`,
			want:  `Hello\`,
		},
		{
			name:  "escape at very end",
			input: `\`,
			want:  `\`,
		},
		{
			name:  "unrecognized single char",
			input: `Hello\Q\World`,
			want:  `Hello\Q\World`,
		},
		{
			name:  "unrecognized multi char",
			input: `Hello\XYZ\World`,
			want:  `Hello\XYZ\World`,
		},
		{
			name:  "invalid hex odd length",
			input: `\X4\`,
			want:  `\X4\`,
		},
		{
			name:  "invalid hex non-hex chars",
			input: `\XGH\`,
			want:  `\XGH\`,
		},
		{
			name:  "empty escape sequence",
			input: `Hello\\World`,
			want:  `Hello\\World`,
		},
		{
			name:  "mixed valid and invalid",
			input: `\F\text\Q\\S\`,
			want:  `|text\Q\^`,
		},
		{
			name:  "nested escape attempt",
			input: `\F\F\`,
			want:  `|F\`,
		},
		{
			name:  "partial hex sequence",
			input: `\X4`,
			want:  `\X4`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Unescape(tt.input)
			if got != tt.want {
				t.Errorf("Unescape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUnescape_CustomDelimiters(t *testing.T) {
	custom := &hl7.Delimiters{
		Field:        '#',
		Component:    '@',
		Repetition:   '!',
		Escape:       '~',
		SubComponent: '%',
	}
	e := New(custom)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "custom field separator",
			input: "Hello~F~World",
			want:  "Hello#World",
		},
		{
			name:  "custom component separator",
			input: "Hello~S~World",
			want:  "Hello@World",
		},
		{
			name:  "custom escape character",
			input: "Hello~E~World",
			want:  "Hello~World",
		},
		{
			name:  "standard escapes unchanged",
			input: `Hello\F\World`,
			want:  `Hello\F\World`, // backslash is not the escape char
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Unescape(tt.input)
			if got != tt.want {
				t.Errorf("Unescape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	e := New(hl7.DefaultDelimiters())

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "no special chars",
			input: "Hello World 123",
		},
		{
			name:  "field separator only",
			input: "Hello|World",
		},
		{
			name:  "all delimiters",
			input: `|^&~\`,
		},
		{
			name:  "complex mixed content",
			input: "Patient|Smith^John&Jr~Address|123 Main St",
		},
		{
			name:  "unicode with delimiters",
			input: "Caf\u00e9|Paris^France",
		},
		{
			name:  "multiple escapes needed",
			input: "|||^^^&&&~~~",
		},
		{
			name:  "delimiter at boundaries",
			input: "|middle|",
		},
		{
			name:  "backslash heavy",
			input: `path\to\file`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escaped := e.Escape(tt.input)
			unescaped := e.Unescape(escaped)
			if unescaped != tt.input {
				t.Errorf("Round trip failed:\n  input:     %q\n  escaped:   %q\n  unescaped: %q", tt.input, escaped, unescaped)
			}
		})
	}
}

func TestEncodeHex(t *testing.T) {
	e := New(hl7.DefaultDelimiters())

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single character",
			input: "A",
			want:  `\X41\`,
		},
		{
			name:  "multiple characters",
			input: "ABC",
			want:  `\X414243\`,
		},
		{
			name:  "space",
			input: " ",
			want:  `\X20\`,
		},
		{
			name:  "newline",
			input: "\n",
			want:  `\X0A\`,
		},
		{
			name:  "unicode character",
			input: "\u00e9",
			want:  `\XC3A9\`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.EncodeHex(tt.input)
			if got != tt.want {
				t.Errorf("EncodeHex(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHexRoundTrip(t *testing.T) {
	e := New(hl7.DefaultDelimiters())

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "ascii text",
			input: "Hello World",
		},
		{
			name:  "special characters",
			input: "|^&~\\",
		},
		{
			name:  "binary data simulation",
			input: "\x00\x01\x02\x03",
		},
		{
			name:  "unicode text",
			input: "Caf\u00e9 au lait",
		},
		{
			name:  "mixed content",
			input: "Test\n\r\t123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := e.EncodeHex(tt.input)
			decoded := e.Unescape(encoded)
			if decoded != tt.input {
				t.Errorf("Hex round trip failed:\n  input:   %q\n  encoded: %q\n  decoded: %q", tt.input, encoded, decoded)
			}
		})
	}
}

func TestFormattingEscapes(t *testing.T) {
	e := New(hl7.DefaultDelimiters())

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "word wrap start",
			input: `text\.fi\more`,
			want:  "textmore",
		},
		{
			name:  "word wrap end",
			input: `text\.nf\more`,
			want:  "textmore",
		},
		{
			name:  "indent",
			input: `text\.in\more`,
			want:  "textmore",
		},
		{
			name:  "temporary indent",
			input: `text\.ti\more`,
			want:  "textmore",
		},
		{
			name:  "center",
			input: `text\.ce\more`,
			want:  "textmore",
		},
		{
			name:  "combined formatting",
			input: `\.fi\Line1\.br\Line2\.nf\`,
			want:  "Line1\nLine2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Unescape(tt.input)
			if got != tt.want {
				t.Errorf("Unescape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	e := New(hl7.DefaultDelimiters())

	t.Run("very long string without escapes", func(t *testing.T) {
		input := make([]byte, 10000)
		for i := range input {
			input[i] = 'a'
		}
		inputStr := string(input)
		got := e.Escape(inputStr)
		if got != inputStr {
			t.Error("long string without special chars should be unchanged")
		}
	})

	t.Run("string of only delimiters", func(t *testing.T) {
		input := "||||"
		want := `\F\\F\\F\\F\`
		got := e.Escape(input)
		if got != want {
			t.Errorf("Escape(%q) = %q, want %q", input, got, want)
		}
		unescaped := e.Unescape(got)
		if unescaped != input {
			t.Errorf("Round trip failed for delimiter-only string")
		}
	})

	t.Run("single delimiter character", func(t *testing.T) {
		for _, tc := range []struct {
			char     string
			escaped  string
			unescapr string
		}{
			{"|", `\F\`, "|"},
			{"^", `\S\`, "^"},
			{"&", `\T\`, "&"},
			{"~", `\R\`, "~"},
			{`\`, `\E\`, `\`},
		} {
			got := e.Escape(tc.char)
			if got != tc.escaped {
				t.Errorf("Escape(%q) = %q, want %q", tc.char, got, tc.escaped)
			}
			unesc := e.Unescape(got)
			if unesc != tc.unescapr {
				t.Errorf("Unescape(%q) = %q, want %q", got, unesc, tc.unescapr)
			}
		}
	})

	t.Run("escape sequence at string boundaries", func(t *testing.T) {
		// Escape at start
		got := e.Unescape(`\F\text`)
		if got != "|text" {
			t.Errorf("expected '|text', got %q", got)
		}

		// Escape at end
		got = e.Unescape(`text\F\`)
		if got != "text|" {
			t.Errorf("expected 'text|', got %q", got)
		}

		// Only escape
		got = e.Unescape(`\F\`)
		if got != "|" {
			t.Errorf("expected '|', got %q", got)
		}
	})

	t.Run("unicode handling", func(t *testing.T) {
		input := "\u4e2d\u6587|\u65e5\u672c\u8a9e" // Chinese|Japanese
		escaped := e.Escape(input)
		unescaped := e.Unescape(escaped)
		if unescaped != input {
			t.Errorf("Unicode round trip failed:\n  input:     %q\n  escaped:   %q\n  unescaped: %q", input, escaped, unescaped)
		}
	})

	t.Run("emoji handling", func(t *testing.T) {
		input := "Hello|\U0001F600|World" // Hello|emoji|World
		escaped := e.Escape(input)
		unescaped := e.Unescape(escaped)
		if unescaped != input {
			t.Errorf("Emoji round trip failed:\n  input:     %q\n  escaped:   %q\n  unescaped: %q", input, escaped, unescaped)
		}
	})
}

// Benchmarks

func BenchmarkEscape_NoSpecialChars(b *testing.B) {
	e := New(hl7.DefaultDelimiters())
	input := "This is a normal string with no special characters"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Escape(input)
	}
}

func BenchmarkEscape_WithSpecialChars(b *testing.B) {
	e := New(hl7.DefaultDelimiters())
	input := "Patient|Smith^John&Jr~Address|123 Main St^City&State"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Escape(input)
	}
}

func BenchmarkUnescape_NoEscapes(b *testing.B) {
	e := New(hl7.DefaultDelimiters())
	input := "This is a normal string with no escape sequences"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Unescape(input)
	}
}

func BenchmarkUnescape_WithEscapes(b *testing.B) {
	e := New(hl7.DefaultDelimiters())
	input := `Patient\F\Smith\S\John\T\Jr\R\Address\F\123 Main St`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Unescape(input)
	}
}

func BenchmarkUnescape_HexEncoding(b *testing.B) {
	e := New(hl7.DefaultDelimiters())
	input := `Start\X48454C4C4F\Middle\X574F524C44\End`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Unescape(input)
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	e := New(hl7.DefaultDelimiters())
	input := "Patient|Smith^John&Jr~Address|123 Main St"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		escaped := e.Escape(input)
		e.Unescape(escaped)
	}
}
