package render

import (
	"strings"
	"testing"
)

func TestGetInitials(t *testing.T) {
	cases := []struct {
		name  string
		input string
		exp   string
	}{
		{"empty", "", ""},
		{"single", "alice", "A"},
		{"two words", "alice baker", "AB"},
		{"extra words", "alice baker charlie", "AB"},
		{"mixed spacing", "  alice   baker  ", "AB"},
		{"non letters", "  -alice  123 baker", "-1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetInitials(tc.input); got != tc.exp {
				t.Fatalf("expected %q got %q", tc.exp, got)
			}
		})
	}
}

func TestGetContrastColorWithGradient(t *testing.T) {
	cases := []struct {
		name  string
		input string
		exp   string
	}{
		{"light gradient", "ffffff,cccccc", "000000"},
		{"dark gradient", "000000,333333", "ffffff"},
		{"red to blue", "ff0000,0000ff", "ffffff"},
		{"single color still works", "ffffff", "000000"},
		{"single dark color", "000000", "ffffff"},
		{"more than 2 colors uses first", "cccccc,00ff00,0000ff", "000000"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetContrastColor(tc.input); got != tc.exp {
				t.Fatalf("expected %q got %q", tc.exp, got)
			}
		})
	}
}

func TestDrawImageWithGradient(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	// Test that gradient image generation doesn't error
	_, err = r.DrawImageWithFormat(400, 300, "ff0000,0000ff", "ffffff", "Test", false, false, FormatPNG)
	if err != nil {
		t.Fatalf("failed to draw image with gradient: %v", err)
	}

	// Test with single color (existing behavior)
	_, err = r.DrawImageWithFormat(400, 300, "ff0000", "ffffff", "Test", false, false, FormatPNG)
	if err != nil {
		t.Fatalf("failed to draw image with solid color: %v", err)
	}

	// Test with more than 2 colors (should use first color)
	_, err = r.DrawImageWithFormat(400, 300, "ff0000,00ff00,0000ff", "ffffff", "Test", false, false, FormatPNG)
	if err != nil {
		t.Fatalf("failed to draw image with more than 2 colors: %v", err)
	}
}

func TestDrawImageWithSVGFormat(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	tests := []struct {
		name    string
		width   int
		height  int
		bg      string
		fg      string
		text    string
		rounded bool
	}{
		{"Simple SVG", 200, 200, "cccccc", "000000", "AB", false},
		{"Rounded SVG", 128, 128, "f0e9e9", "8b5d5d", "JD", true},
		{"Gradient SVG", 400, 300, "ff0000,0000ff", "ffffff", "Test", false},
		{"Large text SVG", 500, 300, "333333", "ffffff", "Hello World", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := r.DrawImageWithFormat(tt.width, tt.height, tt.bg, tt.fg, tt.text, tt.rounded, false, FormatSVG)
			if err != nil {
				t.Fatalf("failed to draw SVG: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("expected SVG data, got empty")
			}
			// Verify it starts with SVG tag
			svgStr := string(data)
			if !strings.HasPrefix(svgStr, "<svg") {
				t.Fatalf("expected SVG to start with <svg, got: %s", svgStr[:20])
			}
			// Verify it contains the text
			if !strings.Contains(svgStr, tt.text) {
				t.Fatalf("expected SVG to contain text '%s', got: %s", tt.text, svgStr)
			}
		})
	}
}

func TestDrawImageWithSVGBold(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	// Test with bold=false
	normalData, err := r.DrawImageWithFormat(200, 200, "cccccc", "000000", "AB", false, false, FormatSVG)
	if err != nil {
		t.Fatalf("failed to draw normal SVG: %v", err)
	}
	normalStr := string(normalData)
	if !strings.Contains(normalStr, `font-weight="normal"`) {
		t.Fatalf("expected normal font-weight, got: %s", normalStr)
	}

	// Test with bold=true
	boldData, err := r.DrawImageWithFormat(200, 200, "cccccc", "000000", "AB", false, true, FormatSVG)
	if err != nil {
		t.Fatalf("failed to draw bold SVG: %v", err)
	}
	boldStr := string(boldData)
	if !strings.Contains(boldStr, `font-weight="bold"`) {
		t.Fatalf("expected bold font-weight, got: %s", boldStr)
	}
}

func TestDrawPlaceholderImageWithQuote(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	tests := []struct {
		name          string
		width         int
		height        int
		text          string
		isQuoteOrJoke bool
		format        ImageFormat
	}{
		{"Short quote PNG", 800, 400, "The only way to do great work is to love what you do.", true, FormatPNG},
		{"Long quote PNG", 1000, 600, "Success is not final, failure is not fatal: It is the courage to continue that counts. Success is not final, failure is not fatal.", true, FormatPNG},
		{"Regular text", 400, 300, "400 x 300", false, FormatPNG},
		{"Short quote SVG", 800, 400, "Be yourself; everyone else is already taken.", true, FormatSVG},
		{"Long quote SVG", 1200, 500, "In three words I can sum up everything I've learned about life: it goes on. And on and on.", true, FormatSVG},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := r.DrawPlaceholderImage(tt.width, tt.height, "2c3e50", "ecf0f1", tt.text, tt.isQuoteOrJoke, tt.format)
			if err != nil {
				t.Fatalf("failed to draw placeholder: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("expected image data, got empty")
			}
		})
	}
}

func TestWrapTextForSVG(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    float64
		fontSize float64
		minLines int
		maxLines int
	}{
		{"Short text", "Hello World", 800, 24, 1, 1},
		{"Long text wraps", "The only way to do great work is to love what you do. Stay hungry, stay foolish.", 600, 24, 2, 5},
		{"Very long text", "Success is not final, failure is not fatal: It is the courage to continue that counts. Success is not final, failure is not fatal: It is the courage to continue that counts.", 800, 20, 3, 8},
		{"Small width forces wrapping", "This is a test of text wrapping", 300, 18, 2, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := wrapTextForSVG(tt.text, tt.width, tt.fontSize)
			if len(lines) < tt.minLines {
				t.Errorf("expected at least %d lines, got %d", tt.minLines, len(lines))
			}
			if len(lines) > tt.maxLines {
				t.Errorf("expected at most %d lines, got %d", tt.maxLines, len(lines))
			}
			// Verify all lines are non-empty
			for i, line := range lines {
				if strings.TrimSpace(line) == "" {
					t.Errorf("line %d is empty", i)
				}
			}
		})
	}
}
