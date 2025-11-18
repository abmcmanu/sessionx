package session

import "time"

type Config struct {
	CookieName       string
	SecretKey        []byte // 32 bytes recommended
	MaxAge           time.Duration
	Path             string
	Domain           string
	Secure           bool
	HttpOnly         bool
	SameSite         string        // "Lax", "Strict", "None"
	RotationInterval time.Duration // auto-regen session ID
}

type ConfigOption func(*Config)

// DefaultConfig returns a production-ready configuration
func DefaultConfig(secretKey []byte, opts ...ConfigOption) Config {
	cfg := Config{
		CookieName:       "sessionx",
		SecretKey:        secretKey,
		MaxAge:           24 * time.Hour,
		Path:             "/",
		Secure:           true,
		HttpOnly:         true,
		SameSite:         "Lax",
		RotationInterval: 15 * time.Minute,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

// DevConfig returns a development-friendly configuration (Secure=false for localhost)
func DevConfig(secretKey []byte, opts ...ConfigOption) Config {
	cfg := Config{
		CookieName:       "sessionx",
		SecretKey:        secretKey,
		MaxAge:           24 * time.Hour,
		Path:             "/",
		Secure:           false, // Allow HTTP for localhost
		HttpOnly:         true,
		SameSite:         "Lax",
		RotationInterval: 15 * time.Minute,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

func WithMaxAge(d time.Duration) ConfigOption {
	return func(c *Config) {
		c.MaxAge = d
	}
}

func WithCookieName(name string) ConfigOption {
	return func(c *Config) {
		c.CookieName = name
	}
}

func WithDomain(domain string) ConfigOption {
	return func(c *Config) {
		c.Domain = domain
	}
}

func WithPath(path string) ConfigOption {
	return func(c *Config) {
		c.Path = path
	}
}

func WithSecure(secure bool) ConfigOption {
	return func(c *Config) {
		c.Secure = secure
	}
}

func WithHttpOnly(httpOnly bool) ConfigOption {
	return func(c *Config) {
		c.HttpOnly = httpOnly
	}
}

func WithSameSite(sameSite string) ConfigOption {
	return func(c *Config) {
		c.SameSite = sameSite
	}
}

func WithRotationInterval(d time.Duration) ConfigOption {
	return func(c *Config) {
		c.RotationInterval = d
	}
}
