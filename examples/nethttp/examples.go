package main

import (
	"fmt"
	"github.com/abmcmanu/sessionx/pkg/session"
	"net/http"
	"time"
)

func ExampleDevelopment() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithMaxAge(1*time.Hour),
	)
	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	_ = http.ListenAndServe(":8080", manager.Middleware(handler()))
}

func ExampleProduction() {
	cfg := session.DefaultConfig(
		[]byte("your-production-secret-key-32bytes"),
		session.WithMaxAge(24*time.Hour),
		session.WithCookieName("prod_session"),
	)
	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	_ = http.ListenAndServe(":443", manager.Middleware(handler()))
}

func ExampleShortLivedSessions() {
	cfg := session.DefaultConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithMaxAge(30*time.Minute),
		session.WithSameSite("Strict"),
	)
	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	_ = http.ListenAndServe(":8080", manager.Middleware(handler()))
}

func ExampleSubdomains() {
	cfg := session.DefaultConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithDomain(".example.com"),
		session.WithMaxAge(7*24*time.Hour),
	)
	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	_ = http.ListenAndServe(":8080", manager.Middleware(handler()))
}

func ExampleAPIPath() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithPath("/api"),
		session.WithMaxAge(2*time.Hour),
		session.WithCookieName("api_session"),
	)
	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	_ = http.ListenAndServe(":8080", manager.Middleware(handler()))
}

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
	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	_ = http.ListenAndServe(":8080", manager.Middleware(handler()))
}

func handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)
		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		_, _ = fmt.Fprintf(w, "Session ID: %s\n", sess.ID)
		_, _ = fmt.Fprintf(w, "Visits: %.0f\n", sess.Data["count"])
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)
		sess.Data["user_id"] = "12345"
		sess.Data["username"] = "john_doe"
		sess.Data["role"] = "admin"
		sess.Data["logged_in"] = true

		_, _ = fmt.Fprintf(w, "Logged in successfully!\n")
	})

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)
		loggedIn, _ := sess.Data["logged_in"].(bool)

		if !loggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, _ := sess.Data["username"].(string)
		_, _ = fmt.Fprintf(w, "Welcome to dashboard, %s!\n", username)
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)
		sess.Data = make(map[string]interface{})

		_, _ = fmt.Fprintf(w, "Logged out successfully!\n")
	})

	return mux
}