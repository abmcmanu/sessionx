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
		r = r.WithContext(context.WithValue(r.Context(), Key, sess))

		wrapped := &responseWriterWrapper{
			ResponseWriter: w,
			session:        sess,
			manager:        m,
		}

		next.ServeHTTP(wrapped, r)

		wrapped.save()
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	session *Session
	manager *Manager
	saved   bool
}

func (rw *responseWriterWrapper) save() {
	if rw.saved {
		return
	}
	rw.saved = true
	_ = rw.manager.Save(rw.ResponseWriter, rw.session)
}

func (rw *responseWriterWrapper) WriteHeader(status int) {
	rw.save()
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	rw.save()
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriterWrapper) Flush() {
	rw.save()
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (rw *responseWriterWrapper) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func Get(r *http.Request) *Session {
	if sess, ok := r.Context().Value(Key).(*Session); ok {
		return sess
	}
	return nil
}
