package render

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strconv"
	"strings"

	"github.com/chai2010/webp"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"

	"grout/internal/config"
)

// Renderer is responsible for drawing avatars and placeholders.
type Renderer struct {
	regular *truetype.Font
	bold    *truetype.Font
}

// New creates a renderer preloaded with embedded fonts.
func New() (*Renderer, error) {
	regular, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse regular font: %w", err)
	}
	bold, err := truetype.Parse(gobold.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse bold font: %w", err)
	}
	return &Renderer{regular: regular, bold: bold}, nil
}

// ImageFormat represents the output image format
type ImageFormat string

const (
	FormatPNG  ImageFormat = "png"
	FormatJPG  ImageFormat = "jpg"
	FormatJPEG ImageFormat = "jpeg"
	FormatGIF  ImageFormat = "gif"
	FormatWebP ImageFormat = "webp"
	FormatSVG  ImageFormat = "svg"
)

// parseGradientColors parses a comma-separated color string into two colors.
// Returns the two colors if valid gradient (exactly 2 colors).
// Returns first color and empty string if more than 2 colors.
// Returns empty strings if not a gradient.
func parseGradientColors(bgHex string) (string, string) {
	if !strings.Contains(bgHex, ",") {
		return "", ""
	}
	colors := strings.Split(bgHex, ",")
	if len(colors) == 2 {
		return strings.TrimSpace(colors[0]), strings.TrimSpace(colors[1])
	}
	if len(colors) > 2 {
		// More than 2 colors - return first color only
		return strings.TrimSpace(colors[0]), ""
	}
	return "", ""
}

// DrawImage renders an image with provided options.
func (r *Renderer) DrawImage(w, h int, bgHex, fgHex, text string, rounded, bold bool) ([]byte, error) {
	return r.DrawImageWithFormat(w, h, bgHex, fgHex, text, rounded, bold, FormatSVG)
}

// DrawPlaceholderImage renders a placeholder image with optimized font sizing for quotes/jokes
func (r *Renderer) DrawPlaceholderImage(w, h int, bgHex, fgHex, text string, isQuoteOrJoke bool, format ImageFormat) ([]byte, error) {
	// Calculate font size based on whether it's a quote/joke or regular placeholder
	var fontSize float64

	if isQuoteOrJoke {
		// For quotes/jokes, use dynamic sizing based on text length and image dimensions
		// Start with a base size relative to height
		fontSize = float64(h) * 0.08

		// Adjust based on text length
		textLen := len(text)
		if textLen > 200 {
			fontSize = float64(h) * 0.05
		} else if textLen > 100 {
			fontSize = float64(h) * 0.06
		}

		// Apply min/max bounds from config
		if fontSize < config.MinFontSize {
			fontSize = config.MinFontSize
		}
		if fontSize > config.MaxFontSize {
			fontSize = config.MaxFontSize
		}
	} else {
		// For regular placeholders (dimensions text, initials), use existing logic
		minDim := float64(w)
		if float64(h) < minDim {
			minDim = float64(h)
		}

		fontSize = minDim * 0.5
		if len(text) > config.MinTextLengthForWrapping {
			fontSize = minDim * 0.15
			if fontSize < 12 {
				fontSize = 12
			}
		}
	}

	// For SVG format, generate directly without rasterization
	if format == FormatSVG {
		return r.generateSVGWithWrapping(w, h, bgHex, fgHex, text, false, true, fontSize, isQuoteOrJoke)
	}

	// For raster formats, create the image using gg
	return r.drawRasterImageWithWrapping(w, h, bgHex, fgHex, text, false, true, fontSize, isQuoteOrJoke, format)
}

// DrawImageWithFormat renders an image in the specified format with provided options.
func (r *Renderer) DrawImageWithFormat(w, h int, bgHex, fgHex, text string, rounded, bold bool, format ImageFormat) ([]byte, error) {
	// Calculate font size for consistent rendering across formats
	minDim := float64(w)
	if float64(h) < minDim {
		minDim = float64(h)
	}

	fontSize := minDim * 0.5
	if len(text) > config.MinTextLengthForWrapping {
		fontSize = minDim * 0.15
		if fontSize < 12 {
			fontSize = 12
		}
	}

	// For SVG format, generate directly without rasterization
	if format == FormatSVG {
		return r.generateSVGWithWrapping(w, h, bgHex, fgHex, text, rounded, bold, fontSize, false)
	}

	// For raster formats, create the image using gg
	return r.drawRasterImageWithWrapping(w, h, bgHex, fgHex, text, rounded, bold, fontSize, false, format)
}

// drawRasterImageWithWrapping renders a raster image with text wrapping support
func (r *Renderer) drawRasterImageWithWrapping(w, h int, bgHex, fgHex, text string, rounded, bold bool, fontSize float64, isQuoteOrJoke bool, format ImageFormat) ([]byte, error) {
	dc := gg.NewContext(w, h)

	// Check if bgHex contains a gradient (comma-separated colors)
	color1, color2 := parseGradientColors(bgHex)
	if color1 != "" && color2 != "" {
		// Create linear gradient from left to right
		gradient := gg.NewLinearGradient(0, 0, float64(w), 0)
		gradient.AddColorStop(0, ParseHexColor(color1))
		gradient.AddColorStop(1, ParseHexColor(color2))
		dc.SetFillStyle(gradient)
	} else {
		// Solid color (use first color if comma-separated but invalid)
		if color1 != "" {
			dc.SetColor(ParseHexColor(color1))
		} else {
			dc.SetColor(ParseHexColor(bgHex))
		}
	}

	fg := ParseHexColor(fgHex)
	if rounded {
		dc.DrawCircle(float64(w)/2, float64(h)/2, float64(w)/2)
		dc.Fill()
	} else {
		dc.DrawRectangle(0, 0, float64(w), float64(h))
		dc.Fill()
	}

	font := r.regular
	if bold {
		font = r.bold
	}
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: fontSize}))
	dc.SetColor(fg)

	// Wrap text if it's a quote/joke (use wrapping for readability)
	// For short text like initials or dimensions, use single-line rendering
	if isQuoteOrJoke {
		lines := r.wrapText(dc, text, float64(w), fontSize)
		drawMultiLineText(dc, lines, float64(w), float64(h), fontSize)
	} else {
		// For initials/short text/dimensions, draw as single line
		dc.DrawStringAnchored(text, float64(w)/2, float64(h)/2, 0.5, 0.5)
	}

	return encodeImage(dc.Image(), format)
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

// encodeImage encodes a rasterized image in the specified format (PNG, JPEG, GIF, WebP)
func encodeImage(img image.Image, format ImageFormat) ([]byte, error) {
	var buf bytes.Buffer

	switch format {
	case FormatPNG:
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("encode png: %w", err)
		}
	case FormatJPG, FormatJPEG:
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, fmt.Errorf("encode jpeg: %w", err)
		}
	case FormatGIF:
		if err := gif.Encode(&buf, img, nil); err != nil {
			return nil, fmt.Errorf("encode gif: %w", err)
		}
	case FormatWebP:
		if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 90}); err != nil {
			return nil, fmt.Errorf("encode webp: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported raster format: %s", format)
	}

	return buf.Bytes(), nil
}

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

// escapeXML escapes special XML characters in text
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// ParseHexColor converts #rgb/#rrggbb strings to RGBA.
func ParseHexColor(s string) color.Color {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return color.RGBA{200, 200, 200, 255}
	}
	rgb, err := hexDecode(s)
	if err != nil {
		return color.RGBA{200, 200, 200, 255}
	}
	return color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 255}
}

func hexDecode(s string) ([]uint8, error) {
	b := make([]uint8, 3)
	for i := 0; i < 3; i++ {
		part := s[i*2 : i*2+2]
		val, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return nil, err
		}
		b[i] = uint8(val)
	}
	return b, nil
}

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

// GenerateColorHash returns a deterministic color hex from input.
func GenerateColorHash(seed string) string {
	hash := md5.Sum([]byte(seed))
	return fmt.Sprintf("%02x%02x%02x", hash[0], hash[1], hash[2])
}

// GetContrastColor determines if white or black text should be used
func GetContrastColor(bgHex string) string {
	// Handle gradient colors by averaging the two colors
	color1, color2 := parseGradientColors(bgHex)
	if color1 != "" && color2 != "" {
		c1 := ParseHexColor(color1).(color.RGBA)
		c2 := ParseHexColor(color2).(color.RGBA)
		// Average the two colors
		r := (float64(c1.R) + float64(c2.R)) / 2.0 / 255.0
		g := (float64(c1.G) + float64(c2.G)) / 2.0 / 255.0
		b := (float64(c1.B) + float64(c2.B)) / 2.0 / 255.0
		luminance := (0.2126 * r) + (0.7152 * g) + (0.0722 * b)
		if luminance > 0.5 {
			return "000000"
		}
		return "ffffff"
	}

	// Parse single color (or use first color if gradient parsing failed)
	if color1 != "" {
		bgHex = color1
	}

	// 1. Parse the background color
	c := ParseHexColor(bgHex).(color.RGBA)

	// 2. Normalize RGB values to 0-1 range
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	// 3. Calculate Relative Luminance
	// Formula: 0.2126*R + 0.7152*G + 0.0722*B
	luminance := (0.2126 * r) + (0.7152 * g) + (0.0722 * b)

	// 4. Return Black for light backgrounds, White for dark
	if luminance > 0.5 {
		return "000000" // Dark text
	}
	return "ffffff" // Light text
}
