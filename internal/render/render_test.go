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
