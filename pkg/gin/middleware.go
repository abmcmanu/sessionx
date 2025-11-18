package gin

import (
	"github.com/abmcmanu/sessionx/pkg/session"
	"github.com/gin-gonic/gin"
)

const SessionKey = "sessionx"

func SessionMiddleware(manager *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess, _ := manager.Load(c.Request)

		c.Set(SessionKey, sess)

		// Wrap the response writer to intercept Write/WriteHeader calls
		wrapped := &responseWriterWrapper{
			ResponseWriter: c.Writer,
			session:        sess,
			manager:        manager,
		}
		c.Writer = wrapped

		c.Next()

		// Ensure session is saved even if no response was written
		wrapped.ensureSaved()
	}
}

// responseWriterWrapper wraps gin.ResponseWriter to save session before writing response
type responseWriterWrapper struct {
	gin.ResponseWriter
	session *session.Session
	manager *session.Manager
	saved   bool
}

// WriteHeader saves the session before writing the status code
func (rw *responseWriterWrapper) WriteHeader(status int) {
	if !rw.saved {
		_ = rw.manager.Save(rw.ResponseWriter, rw.session)
		rw.saved = true
	}
	rw.ResponseWriter.WriteHeader(status)
}

// Write saves the session before writing the response body
func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.saved {
		_ = rw.manager.Save(rw.ResponseWriter, rw.session)
		rw.saved = true
	}
	return rw.ResponseWriter.Write(b)
}

// WriteString saves the session before writing a string
func (rw *responseWriterWrapper) WriteString(s string) (int, error) {
	if !rw.saved {
		_ = rw.manager.Save(rw.ResponseWriter, rw.session)
		rw.saved = true
	}
	return rw.ResponseWriter.WriteString(s)
}

func (rw *responseWriterWrapper) WriteHeaderNow() {
	if !rw.saved {
		_ = rw.manager.Save(rw.ResponseWriter, rw.session)
		rw.saved = true
	}
	rw.ResponseWriter.WriteHeaderNow()
}

// ensureSaved guarantees the session is saved even if no write occurred
func (rw *responseWriterWrapper) ensureSaved() {
	if !rw.saved {
		_ = rw.manager.Save(rw.ResponseWriter, rw.session)
		rw.saved = true
	}
}

// Get retrieves the session from the Gin context
func Get(c *gin.Context) *session.Session {
	if v, exists := c.Get(SessionKey); exists {
		return v.(*session.Session)
	}
	return nil
}
