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

func DefaultConfig(secretKey []byte) Config {
	return Config{
		CookieName:       "sessionx",
		SecretKey:        secretKey,
		MaxAge:           24 * time.Hour,
		Path:             "/",
		Secure:           true,
		HttpOnly:         true,
		SameSite:         "Lax",
		RotationInterval: 15 * time.Minute,
	}
}
