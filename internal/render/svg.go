package render

import (
	"bytes"
	"fmt"
	"strings"
)

// generateSVGWithWrapping creates an SVG representation with text wrapping support
func (r *Renderer) generateSVGWithWrapping(w, h int, bgHex, fgHex, text string, rounded, bold bool, fontSize float64, isQuoteOrJoke bool) ([]byte, error) {
	var buf bytes.Buffer

	// SVG header
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w, h, w, h))
	buf.WriteString("\n")

	// Check if bgHex contains a gradient (comma-separated colors)
	color1, color2 := parseGradientColors(bgHex)

	// Calculate radius for rounded shapes (use minimum dimension to ensure circle fits)
	radius := w
	if h < w {
		radius = h
	}
	radius = radius / 2

	if color1 != "" && color2 != "" {
		// Generate unique gradient ID based on colors to avoid conflicts
		gradientID := fmt.Sprintf("grad_%s_%s", color1, color2)

		// Define linear gradient
		buf.WriteString(fmt.Sprintf(`<defs><linearGradient id="%s" x1="0%%" y1="0%%" x2="100%%" y2="0%%">`, gradientID))
		buf.WriteString(fmt.Sprintf(`<stop offset="0%%" style="stop-color:#%s;stop-opacity:1" />`, color1))
		buf.WriteString(fmt.Sprintf(`<stop offset="100%%" style="stop-color:#%s;stop-opacity:1" />`, color2))
		buf.WriteString(`</linearGradient></defs>`)
		buf.WriteString("\n")

		// Background shape with gradient
		if rounded {
			buf.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="url(#%s)" />`, w/2, h/2, radius, gradientID))
		} else {
			buf.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="url(#%s)" />`, w, h, gradientID))
		}
	} else {
		// Solid color background
		if color1 != "" {
			bgHex = color1
		}
		if rounded {
			buf.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="#%s" />`, w/2, h/2, radius, bgHex))
		} else {
			buf.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#%s" />`, w, h, bgHex))
		}
	}
	buf.WriteString("\n")

	// Text element(s)
	fontWeight := "normal"
	if bold {
		fontWeight = "bold"
	}

	// Wrap text if it's a quote/joke (use wrapping for readability)
	// For short text like initials or dimensions, use single-line rendering
	if isQuoteOrJoke {
		lines := wrapTextForSVG(text, float64(w), fontSize)
		lineHeight := fontSize * 1.5
		totalHeight := float64(len(lines)) * lineHeight
		centerY := float64(h) / 2
		startY := centerY - (totalHeight-lineHeight)/2

		for i, line := range lines {
			y := startY + float64(i)*lineHeight
			buf.WriteString(fmt.Sprintf(`<text x="%d" y="%.0f" font-family="sans-serif" font-size="%.0f" font-weight="%s" fill="#%s" text-anchor="middle" dominant-baseline="middle">%s</text>`,
				w/2, y, fontSize, fontWeight, fgHex, escapeXML(line)))
			buf.WriteString("\n")
		}
	} else {
		// For initials/short text/dimensions, draw as single line
		buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="sans-serif" font-size="%.0f" font-weight="%s" fill="#%s" text-anchor="middle" dominant-baseline="middle">%s</text>`,
			w/2, h/2, fontSize, fontWeight, fgHex, escapeXML(text)))
		buf.WriteString("\n")
	}

	// Close SVG
	buf.WriteString("</svg>")

	return buf.Bytes(), nil
}

// escapeXML escapes special XML characters in text
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
