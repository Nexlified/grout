package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"grout/internal/config"
	"grout/internal/render"
	"grout/internal/utils"
)

func (s *Service) handleAvatar(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	format := render.FormatSVG // Default to SVG

	if strings.HasPrefix(r.URL.Path, "/avatar/") {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 2 && parts[2] != "" {
			format, name = extractFormat(parts[2])
		}
	}
	if name == "" {
		name = "John Doe"
	}

	size := utils.ParseIntOrDefault(r.URL.Query().Get("size"), config.DefaultSize)
	rounded := r.URL.Query().Get("rounded") == "true"
	bold := r.URL.Query().Get("bold") == "true"

	// Accept both 'background' and 'bg' for consistency (background is primary)
	bgHex := r.URL.Query().Get("background")
	if bgHex == "" {
		bgHex = r.URL.Query().Get("bg")
	}
	if bgHex == "" {
		bgHex = config.DefaultAvatarBg
	}
	if strings.EqualFold(bgHex, "random") {
		bgHex = render.GenerateColorHash(name)
	}

	fgHex := r.URL.Query().Get("color")
	if fgHex == "" {
		fgHex = render.GetContrastColor(bgHex)
	}

	key := fmt.Sprintf("Avatar:%s:%d:%t:%t:%s:%s:%s", name, size, rounded, bold, bgHex, fgHex, format)
	s.serveImage(w, r, key, format, func() ([]byte, error) {
		return s.renderer.DrawImageWithFormat(size, size, bgHex, fgHex, render.GetInitials(name), rounded, bold, format)
	})
}
