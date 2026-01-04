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

// DrawImageWithFormat renders an image in the specified format with provided options.
func (r *Renderer) DrawImageWithFormat(w, h int, bgHex, fgHex, text string, rounded, bold bool, format ImageFormat) ([]byte, error) {
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

	minDim := float64(w)
	if float64(h) < minDim {
		minDim = float64(h)
	}

	fontSize := minDim * 0.5
	if len(text) > 2 {
		fontSize = minDim * 0.15
		if fontSize < 12 {
			fontSize = 12
		}
	}

	font := r.regular
	if bold {
		font = r.bold
	}
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: fontSize}))
	dc.SetColor(fg)
	dc.DrawStringAnchored(text, float64(w)/2, float64(h)/2, 0.5, 0.5)

	return encodeImage(dc.Image(), format, w, h, bgHex, fgHex, text, rounded, bold, font, fontSize)
}

// encodeImage encodes the image in the specified format
func encodeImage(img image.Image, format ImageFormat, w, h int, bgHex, fgHex, text string, rounded, bold bool, font *truetype.Font, fontSize float64) ([]byte, error) {
	var buf bytes.Buffer

	switch format {
	case FormatSVG:
		// Generate SVG directly without rasterizing
		return generateSVG(w, h, bgHex, fgHex, text, rounded, bold, fontSize)
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
		fallthrough
	default:
		// Default to WebP for any unrecognized format
		if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 90}); err != nil {
			return nil, fmt.Errorf("encode webp: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// generateSVG creates an SVG representation of the image
func generateSVG(w, h int, bgHex, fgHex, text string, rounded, bold bool, fontSize float64) ([]byte, error) {
	var buf bytes.Buffer

	// SVG header
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w, h, w, h))
	buf.WriteString("\n")

	// Check if bgHex contains a gradient (comma-separated colors)
	color1, color2 := parseGradientColors(bgHex)
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
			buf.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="url(#%s)" />`, w/2, h/2, w/2, gradientID))
		} else {
			buf.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="url(#%s)" />`, w, h, gradientID))
		}
	} else {
		// Solid color background
		if color1 != "" {
			bgHex = color1
		}
		if rounded {
			buf.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="#%s" />`, w/2, h/2, w/2, bgHex))
		} else {
			buf.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#%s" />`, w, h, bgHex))
		}
	}
	buf.WriteString("\n")

	// Text element
	// SVG text is positioned by baseline, so we need to adjust
	// Using dominant-baseline="middle" and text-anchor="middle" for centering
	fontWeight := "normal"
	if bold {
		fontWeight = "bold"
	}
	buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Arial, Helvetica, sans-serif" font-size="%.0f" font-weight="%s" fill="#%s" text-anchor="middle" dominant-baseline="middle">%s</text>`,
		w/2, h/2, fontSize, fontWeight, fgHex, escapeXML(text)))
	buf.WriteString("\n")

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
