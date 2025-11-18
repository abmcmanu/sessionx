package main

import (
	"fmt"
	"github.com/abmcmanu/sessionx/pkg/session"
	"net/http"
	"time"
)

func ExampleAutomaticRotation() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithRotationInterval(30*time.Second),
	)
	manager, _ := session.NewManager(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		timeSinceRotation := time.Since(sess.RotatedAt)

		_, _ = fmt.Fprintf(w, "Session ID: %s\n", sess.ID)
		_, _ = fmt.Fprintf(w, "Visits: %.0f\n", sess.Data["count"])
		_, _ = fmt.Fprintf(w, "Last rotated: %s ago\n", timeSinceRotation.Round(time.Second))
		_, _ = fmt.Fprintf(w, "\nSession will rotate automatically after 30 seconds\n")
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}

func ExampleManualRotation() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		sess.Data["user_id"] = "12345"
		sess.Data["username"] = "john_doe"
		sess.Data["logged_in"] = true

		manager.Rotate(sess)

		_, _ = fmt.Fprintf(w, "Logged in successfully\n")
		_, _ = fmt.Fprintf(w, "Session ID rotated for security: %s\n", sess.ID)
	})

	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)
		loggedIn, _ := sess.Data["logged_in"].(bool)

		if !loggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, _ := sess.Data["username"].(string)
		_, _ = fmt.Fprintf(w, "Welcome, %s!\n", username)
		_, _ = fmt.Fprintf(w, "Session ID: %s\n", sess.ID)
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}

func ExampleRotationInterval() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithRotationInterval(1*time.Minute),
		session.WithMaxAge(30*time.Minute),
	)
	manager, _ := session.NewManager(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		oldID := sess.ID
		sess.Data["counter"] = sess.Data["counter"].(float64) + 1

		_, _ = fmt.Fprintf(w, "Session ID: %s\n", sess.ID)
		_, _ = fmt.Fprintf(w, "Previous ID: %s\n", oldID)
		_, _ = fmt.Fprintf(w, "Counter: %.0f\n", sess.Data["counter"])
		_, _ = fmt.Fprintf(w, "\nCreated: %s\n", sess.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(w, "Rotated: %s\n", sess.RotatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(w, "Updated: %s\n", sess.UpdatedAt.Format(time.RFC3339))
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}