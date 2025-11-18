package main

import (
	"fmt"
	"github.com/abmcmanu/sessionx/pkg/session"
	"net/http"
	"time"
)

func main() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithMaxAge(2*time.Hour),
		session.WithCookieName("my_session"),
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

		_, _ = fmt.Fprintf(w, "Visits: %.0f\n", sess.Data["count"])
	})

	_ = http.ListenAndServe(":8080", manager.Middleware(mux))
}
