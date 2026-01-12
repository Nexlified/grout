package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"grout/internal/config"
	"grout/internal/content"
	"grout/internal/render"
	"grout/internal/utils"
)

func (s *Service) handlePlaceholder(w http.ResponseWriter, r *http.Request) {
	width, height := config.DefaultSize, config.DefaultSize
	pathMetric := strings.TrimPrefix(r.URL.Path, "/placeholder/")

	// Extract format from path
	format, pathMetric := extractFormat(pathMetric)

	if matches := placeholderRegex.FindStringSubmatch(pathMetric); len(matches) == 3 {
		width = utils.ParseIntOrDefault(matches[1], config.DefaultSize)
		height = utils.ParseIntOrDefault(matches[2], config.DefaultSize)
	} else {
		width = utils.ParseIntOrDefault(r.URL.Query().Get("w"), config.DefaultSize)
		height = utils.ParseIntOrDefault(r.URL.Query().Get("h"), config.DefaultSize)
	}

	// Check for quote or joke parameter
	quoteParam := r.URL.Query().Get("quote")
	jokeParam := r.URL.Query().Get("joke")
	category := r.URL.Query().Get("category")

	text := r.URL.Query().Get("text")
	isQuoteOrJoke := false

	// Priority: quote > joke > text > default
	// Only render quote/joke if minimum width requirement is met
	if (quoteParam == "true" || quoteParam == "1") && width >= config.MinWidthForQuoteJoke {
		if s.contentManager != nil {
			randomQuote, err := s.contentManager.GetRandom(content.ContentTypeQuote, category)
			if err == nil {
				text = randomQuote
				isQuoteOrJoke = true
			} else {
				// If error (e.g., invalid category), fall back to text or default
				if text == "" {
					text = fmt.Sprintf("%d x %d", width, height)
				}
			}
		}
	} else if (jokeParam == "true" || jokeParam == "1") && width >= config.MinWidthForQuoteJoke {
		if s.contentManager != nil {
			randomJoke, err := s.contentManager.GetRandom(content.ContentTypeJoke, category)
			if err == nil {
				text = randomJoke
				isQuoteOrJoke = true
			} else {
				// If error (e.g., invalid category), fall back to text or default
				if text == "" {
					text = fmt.Sprintf("%d x %d", width, height)
				}
			}
		}
	} else if text == "" {
		text = fmt.Sprintf("%d x %d", width, height)
	}

	// Accept both 'background' and 'bg' for consistency (background is primary)
	bgHex := r.URL.Query().Get("background")
	if bgHex == "" {
		bgHex = r.URL.Query().Get("bg")
	}
	if bgHex == "" {
		bgHex = config.DefaultBgColor
	}
	fgHex := r.URL.Query().Get("color")
	if fgHex == "" {
		fgHex = render.GetContrastColor(bgHex)
	}

	key := fmt.Sprintf("PH:%d:%d:%s:%s:%s:%s", width, height, bgHex, fgHex, text, format)
	s.serveImage(w, r, key, format, func() ([]byte, error) {
		return s.renderer.DrawPlaceholderImage(width, height, bgHex, fgHex, text, isQuoteOrJoke, format)
	})
}
