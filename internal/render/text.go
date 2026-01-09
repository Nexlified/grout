package render

import (
	"strings"

	"github.com/fogleman/gg"

	"grout/internal/config"
)

// GetInitials returns up to two leading letters from the name.
func GetInitials(name string) string {
	parts := strings.Fields(name)
	initials := make([]rune, 0, 2)
	for _, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		initials = append(initials, runes[0])
		if len(initials) == 2 {
			break
		}
	}
	return strings.ToUpper(string(initials))
}

// wrapText breaks text into lines that fit within the given width with padding
func (r *Renderer) wrapText(dc *gg.Context, text string, imageWidth, fontSize float64) []string {
	// Calculate available width with padding (10% on each side = 80% usable)
	padding := imageWidth * 0.1
	maxWidth := imageWidth - (2 * padding)

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " " + word
		} else {
			testLine = word
		}

		// Measure the width of the test line
		width, _ := dc.MeasureString(testLine)

		if width <= maxWidth {
			currentLine = testLine
		} else {
			// Line is too long, save current line and start new one
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Single word is too long, add it anyway
				lines = append(lines, word)
				currentLine = ""
			}
		}
	}

	// Add the last line after processing all words
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		return []string{text}
	}

	return lines
}

// wrapTextForSVG breaks text into lines for SVG rendering (simpler version without measuring)
func wrapTextForSVG(text string, imageWidth, fontSize float64) []string {
	// Estimate character width as roughly 0.6 * fontSize
	charWidth := fontSize * 0.6
	padding := imageWidth * 0.1
	maxWidth := imageWidth - (2 * padding)
	maxCharsPerLine := int(maxWidth / charWidth)

	if maxCharsPerLine < config.MinCharsPerLine {
		maxCharsPerLine = config.MinCharsPerLine
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " " + word
		} else {
			testLine = word
		}

		if len(testLine) <= maxCharsPerLine {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Single word is too long, add it anyway
				lines = append(lines, word)
				currentLine = ""
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		return []string{text}
	}

	return lines
}

// drawMultiLineText draws multiple lines of text centered on the image
func drawMultiLineText(dc *gg.Context, lines []string, width, height, fontSize float64) {
	lineHeight := fontSize * 1.5 // 1.5x line spacing for readability

	// The actual text block height is one font-sized line plus spacing between lines.
	// This avoids counting extra leading above the first line and below the last line.
	totalHeight := fontSize + float64(len(lines)-1)*lineHeight

	// Calculate starting Y position to center the text block vertically
	// Use fontSize/2 to align the first line to the actual text height, not the line spacing.
	startY := (height-totalHeight)/2 + fontSize/2
	// Draw each line centered horizontally
	for i, line := range lines {
		y := startY + float64(i)*lineHeight
		dc.DrawStringAnchored(line, width/2, y, 0.5, 0.5)
	}
}
