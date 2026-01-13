package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/golang-lru/v2"

	"grout/internal/config"
	"grout/internal/handlers"
	"grout/internal/middleware"
	"grout/internal/render"
)

func main() {
	cfg := config.LoadServerConfig()

	renderer, err := render.New()
	if err != nil {
		log.Fatalf("init renderer: %v", err)
	}

	cache, err := lru.New[string, []byte](cfg.CacheSize)
	if err != nil {
		log.Fatalf("init cache: %v", err)
	}

	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPM, cfg.RateLimitBurst)

	svc := handlers.NewService(renderer, cache, cfg)
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux, rateLimiter)

	// Wrap with compression middleware
	handler := middleware.CompressionMiddleware(mux)

	fmt.Printf("Grout running on %s (rate limit: %d req/min, burst: %d)\n", cfg.Addr, cfg.RateLimitRPM, cfg.RateLimitBurst)
	log.Fatal(http.ListenAndServe(cfg.Addr, handler))
}
