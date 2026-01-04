package config

import (
	"flag"
	"os"
	"strconv"
)

const (
	DefaultSize      = 128
	DefaultBgColor   = "cccccc"
	DefaultFontColor = "969696"
	DefaultAvatarBg  = "f0e9e9"
	DefaultAvatarFg  = "8b5d5d"
	DefaultAddr      = ":8080"
	DefaultDomain    = "localhost:8080"
	CacheSize        = 2000
)

// ServerConfig represents runtime server settings.
type ServerConfig struct {
	Addr      string
	Domain    string
	CacheSize int
}

var (
	addrFlag      = flag.String("addr", "", "HTTP listen address (env ADDR)")
	domainFlag    = flag.String("domain", "", "Public domain for example URLs (env DOMAIN)")
	cacheSizeFlag = flag.Int("cache-size", 0, "LRU cache size (env CACHE_SIZE)")
)

// DefaultServerConfig returns sane defaults for local development.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{Addr: DefaultAddr, Domain: DefaultDomain, CacheSize: CacheSize}
}

// LoadServerConfig reads defaults, then env, then flags.
func LoadServerConfig() ServerConfig {
	cfg := DefaultServerConfig()

	if addr := os.Getenv("ADDR"); addr != "" {
		cfg.Addr = addr
	}
	if domain := os.Getenv("DOMAIN"); domain != "" {
		cfg.Domain = domain
	}
	if cacheEnv := os.Getenv("CACHE_SIZE"); cacheEnv != "" {
		if n, err := strconv.Atoi(cacheEnv); err == nil && n > 0 {
			cfg.CacheSize = n
		}
	}

	if !flag.Parsed() {
		flag.Parse()
	}

	if addrFlag != nil && *addrFlag != "" {
		cfg.Addr = *addrFlag
	}
	if domainFlag != nil && *domainFlag != "" {
		cfg.Domain = *domainFlag
	}
	if cacheSizeFlag != nil && *cacheSizeFlag > 0 {
		cfg.CacheSize = *cacheSizeFlag
	}

	return cfg
}
