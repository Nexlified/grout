package render

import (
	"fmt"

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
