package main

import (
	"fmt"
	"github.com/abmcmanu/sessionx/pkg/session"
	"net/http"
)

func main() {
	cfg := session.DefaultConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager := session.NewManager(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Ignore requests for favicon
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		sess := session.Get(r)
		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		_, err := fmt.Fprintf(w, "Visits: %.0f\n", sess.Data["count"])
		if err != nil {
			return
		}
	})

	err := http.ListenAndServe(":8080", manager.Middleware(mux))
	if err != nil {
		return
	}
}
