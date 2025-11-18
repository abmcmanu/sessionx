package main

import (
	"fmt"
	"github.com/abmcmanu/sessionx/pkg/session"
	"net/http"
	"time"
)

// Example 1: Development configuration (localhost)
func ExampleDevelopment() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithMaxAge(1*time.Hour), // Session expires after 1 hour
	)
	manager := session.NewManager(cfg)

	http.ListenAndServe(":8080", manager.Middleware(handler()))
}

// Example 2: Production configuration (HTTPS required)
func ExampleProduction() {
	cfg := session.DefaultConfig(
		[]byte("your-production-secret-key-32bytes"),
		session.WithMaxAge(24*time.Hour),       // 24 hour sessions
		session.WithCookieName("prod_session"), // Custom cookie name
	)
	manager := session.NewManager(cfg)

	http.ListenAndServe(":443", manager.Middleware(handler()))
}

// Example 3: Short-lived sessions (banking app, admin panel)
func ExampleShortLivedSessions() {
	cfg := session.DefaultConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithMaxAge(30*time.Minute), // Expire after 30 minutes
		session.WithSameSite("Strict"),     // Maximum CSRF protection
	)
	manager := session.NewManager(cfg)

	http.ListenAndServe(":8080", manager.Middleware(handler()))
}

// Example 4: Multi-subdomain application
func ExampleSubdomains() {
	cfg := session.DefaultConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithDomain(".example.com"),  // Share across app.example.com, api.example.com
		session.WithMaxAge(7*24*time.Hour),  // 7 days
	)
	manager := session.NewManager(cfg)

	http.ListenAndServe(":8080", manager.Middleware(handler()))
}

// Example 5: Custom API path
func ExampleAPIPath() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithPath("/api"),              // Only for /api/* routes
		session.WithMaxAge(2*time.Hour),
		session.WithCookieName("api_session"),
	)
	manager := session.NewManager(cfg)

	http.ListenAndServe(":8080", manager.Middleware(handler()))
}

// Example 6: All options combined
func ExampleFullCustomization() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithMaxAge(4*time.Hour),
		session.WithCookieName("custom_session"),
		session.WithDomain(".myapp.com"),
		session.WithPath("/"),
		session.WithSameSite("Lax"),
		session.WithRotationInterval(15*time.Minute),
	)
	manager := session.NewManager(cfg)

	http.ListenAndServe(":8080", manager.Middleware(handler()))
}

// Example handler showing session usage
func handler() http.Handler {
	mux := http.NewServeMux()

	// Counter example
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)
		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		fmt.Fprintf(w, "Session ID: %s\n", sess.ID)
		fmt.Fprintf(w, "Visits: %.0f\n", sess.Data["count"])
	})

	// User authentication example
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)
		sess.Data["user_id"] = "12345"
		sess.Data["username"] = "john_doe"
		sess.Data["role"] = "admin"
		sess.Data["logged_in"] = true

		fmt.Fprintf(w, "Logged in successfully!\n")
	})

	// Protected route example
	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)
		loggedIn, _ := sess.Data["logged_in"].(bool)

		if !loggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, _ := sess.Data["username"].(string)
		fmt.Fprintf(w, "Welcome to dashboard, %s!\n", username)
	})

	// Logout example
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		// Option 1: Clear session data
		sess := session.Get(r)
		sess.Data = make(map[string]interface{})

		// Option 2: Destroy session completely (uncomment to use)
		// manager.Destroy(w)

		fmt.Fprintf(w, "Logged out successfully!\n")
	})

	return mux
}