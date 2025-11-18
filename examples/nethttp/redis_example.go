package main

import (
	"fmt"
	"github.com/abmcmanu/sessionx/pkg/session"
	redisstore "github.com/abmcmanu/sessionx/optional/store/redis"
	"net/http"
	"time"
)

func ExampleRedisBasic() {
	store, err := redisstore.NewRedisStore(redisstore.Options{
		Addr:   "localhost:6379",
		Prefix: "session:",
		TTL:    30 * time.Minute,
	})
	if err != nil {
		panic(err)
	}
	defer store.Close()

	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithStore(store),
		session.WithMaxAge(30*time.Minute),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		sess := session.Get(r)

		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		_, _ = fmt.Fprintf(w, "Session ID: %s\n", sess.ID)
		_, _ = fmt.Fprintf(w, "Visits: %.0f\n", sess.Data["count"])
		_, _ = fmt.Fprintf(w, "Stored in Redis!\n")
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}

func ExampleRedisAuthentication() {
	store, err := redisstore.NewRedisStore(redisstore.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Prefix:   "auth_session:",
		TTL:      1 * time.Hour,
	})
	if err != nil {
		panic(err)
	}
	defer store.Close()

	cfg := session.DefaultConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithStore(store),
		session.WithMaxAge(1*time.Hour),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

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
			sess.AddFlash("success", "Login successful!")

			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		} else {
			sess.AddFlash("error", "Invalid credentials")
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
		_, _ = fmt.Fprintf(w, "Session ID: %s\n", sess.ID)
		_, _ = fmt.Fprintf(w, "Session stored in Redis for scalability\n")
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sess := session.Get(r)
		_ = manager.Destroy(w, r)

		_, _ = fmt.Fprintf(w, "Logged out. Session deleted from Redis.\n")
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}

func ExampleRedisMultiServer() {
	store, err := redisstore.NewRedisStore(redisstore.Options{
		Addr:   "redis-cluster.example.com:6379",
		Prefix: "app_session:",
		TTL:    24 * time.Hour,
	})
	if err != nil {
		panic(err)
	}
	defer store.Close()

	cfg := session.DefaultConfig(
		[]byte("shared-secret-key-across-servers"),
		session.WithStore(store),
		session.WithDomain(".example.com"),
		session.WithMaxAge(24*time.Hour),
		session.WithRotationInterval(15*time.Minute),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		hostname := r.Host

		_, _ = fmt.Fprintf(w, "Server: %s\n", hostname)
		_, _ = fmt.Fprintf(w, "Session ID: %s\n", sess.ID)
		_, _ = fmt.Fprintf(w, "Visits: %.0f\n", sess.Data["count"])
		_, _ = fmt.Fprintf(w, "\nThis session works across multiple servers!\n")
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}

func ExampleRedisWithCustomTTL() {
	store, err := redisstore.NewRedisStore(redisstore.Options{
		Addr:   "localhost:6379",
		Prefix: "short_session:",
		TTL:    5 * time.Minute,
	})
	if err != nil {
		panic(err)
	}
	defer store.Close()

	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithStore(store),
		session.WithMaxAge(5*time.Minute),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Get(r)

		sess.Data["created"] = sess.CreatedAt.Format(time.RFC3339)
		sess.Data["expires_in"] = "5 minutes"

		_, _ = fmt.Fprintf(w, "Session expires in 5 minutes\n")
		_, _ = fmt.Fprintf(w, "Created: %s\n", sess.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(w, "Updated: %s\n", sess.UpdatedAt.Format(time.RFC3339))
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}