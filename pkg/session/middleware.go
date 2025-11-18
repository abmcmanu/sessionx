package session

import (
	"context"
	"net/http"
)

type ContextKey string

var Key ContextKey = "sessionx"

func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, _ := m.Load(r)

		ctx := context.WithValue(r.Context(), Key, sess)
		r = r.WithContext(ctx)

		wrapped := &responseWriterWrapper{
			ResponseWriter: w,
			session:        sess,
			manager:        m,
		}

		next.ServeHTTP(wrapped, r)

		wrapped.ensureSaved()
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	session *Session
	manager *Manager
	saved   bool
}

func (rw *responseWriterWrapper) WriteHeader(status int) {
	if !rw.saved {
		_ = rw.manager.Save(rw.ResponseWriter, rw.session)
		rw.saved = true
	}
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.saved {
		_ = rw.manager.Save(rw.ResponseWriter, rw.session)
		rw.saved = true
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriterWrapper) ensureSaved() {
	if !rw.saved {
		_ = rw.manager.Save(rw.ResponseWriter, rw.session)
		rw.saved = true
	}
}

func Get(r *http.Request) *Session {
	if v := r.Context().Value(Key); v != nil {
		return v.(*Session)
	}
	return nil
}
