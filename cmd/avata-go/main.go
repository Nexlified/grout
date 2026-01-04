package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/golang-lru/v2"

	"go-avatars/internal/config"
	"go-avatars/internal/handlers"
	"go-avatars/internal/render"
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

	svc := handlers.NewService(renderer, cache, cfg)
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux)

	fmt.Println("Grout running on", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, mux))
}
