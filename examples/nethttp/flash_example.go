package main

import (
	"fmt"
	"github.com/abmcmanu/sessionx/pkg/session"
	"net/http"
)

func ExampleFlashMessages() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/set-flash", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		sess.AddFlash("success", "Your profile has been updated!")
		sess.AddFlash("info", "You have 3 new messages")
		sess.AddFlash("warning", "Your subscription expires in 7 days")

		http.Redirect(w, r, "/show-flash", http.StatusSeeOther)
	})

	mux.HandleFunc("/show-flash", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		flashes := sess.GetFlashes()

		if len(flashes) == 0 {
			_, _ = fmt.Fprintf(w, "No flash messages\n")
			return
		}

		_, _ = fmt.Fprintf(w, "Flash Messages:\n")
		for key, value := range flashes {
			_, _ = fmt.Fprintf(w, "  [%s]: %v\n", key, value)
		}
		_, _ = fmt.Fprintf(w, "\nRefresh this page - messages will be gone!\n")
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}

func ExampleFlashAfterLogin() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sess := session.Get(r)

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "admin" && password == "password" {
			sess.Data["user_id"] = "12345"
			sess.Data["username"] = username
			sess.Data["logged_in"] = true

			manager.Rotate(sess)
			sess.AddFlash("success", "Welcome back! Login successful.")

			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		} else {
			sess.AddFlash("error", "Invalid username or password")
			http.Redirect(w, r, "/login-form", http.StatusSeeOther)
		}
	})

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		if successMsg, exists := sess.GetFlash("success"); exists {
			_, _ = fmt.Fprintf(w, "[SUCCESS] %s\n\n", successMsg)
		}

		loggedIn, _ := sess.Data["logged_in"].(bool)
		if !loggedIn {
			http.Redirect(w, r, "/login-form", http.StatusSeeOther)
			return
		}

		username, _ := sess.Data["username"].(string)
		_, _ = fmt.Fprintf(w, "Dashboard - Welcome %s!\n", username)
	})

	mux.HandleFunc("/login-form", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		if errorMsg, exists := sess.GetFlash("error"); exists {
			_, _ = fmt.Fprintf(w, "[ERROR] %s\n\n", errorMsg)
		}

		_, _ = fmt.Fprintf(w, "Login Form\n")
		_, _ = fmt.Fprintf(w, "POST to /login with username=admin&password=password\n")
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}

func ExampleFormValidation() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sess := session.Get(r)

		email := r.FormValue("email")
		name := r.FormValue("name")

		if email == "" {
			sess.AddFlash("error", "Email is required")
			http.Redirect(w, r, "/form", http.StatusSeeOther)
			return
		}

		if name == "" {
			sess.AddFlash("error", "Name is required")
			http.Redirect(w, r, "/form", http.StatusSeeOther)
			return
		}

		sess.AddFlash("success", "Form submitted successfully!")
		http.Redirect(w, r, "/form", http.StatusSeeOther)
	})

	mux.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		if errorMsg, exists := sess.GetFlash("error"); exists {
			_, _ = fmt.Fprintf(w, "[ERROR] %s\n\n", errorMsg)
		}

		if successMsg, exists := sess.GetFlash("success"); exists {
			_, _ = fmt.Fprintf(w, "[SUCCESS] %s\n\n", successMsg)
		}

		_, _ = fmt.Fprintf(w, "Form Page\n")
		_, _ = fmt.Fprintf(w, "POST to /submit with email and name\n")
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}

func ExampleCheckFlashBeforeRedirect() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		sess.AddFlash("notification", "New feature available!")

		if sess.HasFlash("notification") {
			_, _ = fmt.Fprintf(w, "Flash message 'notification' is set\n")
		}

		http.Redirect(w, r, "/result", http.StatusSeeOther)
	})

	mux.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		if notification, exists := sess.GetFlash("notification"); exists {
			_, _ = fmt.Fprintf(w, "Notification: %s\n", notification)
		} else {
			_, _ = fmt.Fprintf(w, "No notifications\n")
		}
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}