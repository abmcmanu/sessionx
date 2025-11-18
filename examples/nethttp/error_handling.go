package main

import (
	"errors"
	"fmt"
	"github.com/abmcmanu/sessionx/pkg/session"
	"log"
	"net/http"
)

func ExampleErrorHandling() {
	fmt.Println("Example 1: Invalid secret key")
	cfg1 := session.DevConfig([]byte("short"))
	manager1, err := session.NewManager(cfg1)
	if err != nil {
		if errors.Is(err, session.ErrInvalidSecretKey) {
			fmt.Println("✗ Error: Secret key must be 16, 24, or 32 bytes")
		}

		var sessionErr *session.SessionError
		if errors.As(err, &sessionErr) {
			fmt.Printf("✗ Operation '%s' failed: %v\n", sessionErr.Op, sessionErr.Err)
		}
		fmt.Println()
	} else {
		fmt.Println("✓ Manager created successfully (this shouldn't happen!)")
	}

	fmt.Println("Example 2: Valid secret key")
	cfg2 := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager2, err := session.NewManager(cfg2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Manager created successfully")
	fmt.Println()

	fmt.Println("Example 3: Production pattern with graceful error handling")
	productionExample()

	_ = manager2
}

func productionExample() {
	secretKey := []byte("your-32-byte-production-key-here")

	cfg := session.DefaultConfig(
		secretKey,
		session.WithMaxAge(86400),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		var sessionErr *session.SessionError
		if errors.As(err, &sessionErr) {
			log.Printf("Failed to create session manager in '%s': %v",
				sessionErr.Op, sessionErr.Err)
		}

		if errors.Is(err, session.ErrInvalidSecretKey) {
			log.Fatal("Configuration error: Secret key must be 16, 24, or 32 bytes. " +
				"Please check your environment variables.")
		}

		log.Fatalf("Failed to initialize session manager: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)
		if sess == nil {
			http.Error(w, "Session not found", http.StatusInternalServerError)
			log.Println("Warning: Session is nil - check middleware configuration")
			return
		}

		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		_, _ = fmt.Fprintf(w, "Visit count: %.0f\n", sess.Data["count"])
	})

	fmt.Println("✓ Server configured successfully")
	fmt.Println("  Ready to start on :8080")
	fmt.Println()

	_ = mux
}

func ExampleValidation() {
	secretKey := []byte("user-provided-key")

	keyLen := len(secretKey)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		fmt.Printf("✗ Invalid key length: %d bytes (expected 16, 24, or 32)\n", keyLen)
		return
	}

	fmt.Printf("✓ Valid key length: %d bytes\n", keyLen)

	cfg := session.DefaultConfig(secretKey)
	manager, err := session.NewManager(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✓ Manager created successfully")
	_ = manager
}

func ExampleErrorRecovery() {
	secretKey := []byte("might-be-invalid")

	cfg := session.DevConfig(secretKey)
	manager, err := session.NewManager(cfg)

	if err != nil {
		fmt.Println("✗ Failed to create manager, using default fallback")

		fallbackKey := []byte("0123456789abcdef0123456789abcdef")
		cfg = session.DevConfig(fallbackKey)
		manager, err = session.NewManager(cfg)

		if err != nil {
			log.Fatal("Fatal: Cannot create session manager even with fallback config")
		}

		fmt.Println("✓ Using fallback configuration")
	} else {
		fmt.Println("✓ Using provided configuration")
	}

	_ = manager
}