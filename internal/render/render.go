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
)

// DrawImage renders an image with provided options.
func (r *Renderer) DrawImage(w, h int, bgHex, fgHex, text string, rounded, bold bool) ([]byte, error) {
	return r.DrawImageWithFormat(w, h, bgHex, fgHex, text, rounded, bold, FormatWebP)
}

// DrawImageWithFormat renders an image in the specified format with provided options.
func (r *Renderer) DrawImageWithFormat(w, h int, bgHex, fgHex, text string, rounded, bold bool, format ImageFormat) ([]byte, error) {
	dc := gg.NewContext(w, h)

	bg := ParseHexColor(bgHex)
	fg := ParseHexColor(fgHex)
	dc.SetColor(bg)
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

	return encodeImage(dc.Image(), format)
}

// encodeImage encodes the image in the specified format
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
		fallthrough
	default:
		// Default to WebP for any unrecognized format
		if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 90}); err != nil {
			return nil, fmt.Errorf("encode webp: %w", err)
		}
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
