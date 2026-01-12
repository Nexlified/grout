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
