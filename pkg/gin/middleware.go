package gin

import (
	"context"

	"github.com/abmcmanu/sessionx/pkg/session"
	"github.com/gin-gonic/gin"
)

const SessionKey = "sessionx"

func SessionMiddleware(manager *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess, _ := manager.Load(c.Request)
		c.Set(SessionKey, sess)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), session.Key, sess))

		wrapped := &responseWriterWrapper{
			ResponseWriter: c.Writer,
			session:        sess,
			manager:        manager,
		}
		c.Writer = wrapped

		c.Next()

		wrapped.save()
	}
}

type responseWriterWrapper struct {
	gin.ResponseWriter
	session *session.Session
	manager *session.Manager
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

func (rw *responseWriterWrapper) WriteHeaderNow() {
	rw.save()
	rw.ResponseWriter.WriteHeaderNow()
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	rw.save()
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriterWrapper) WriteString(s string) (int, error) {
	rw.save()
	return rw.ResponseWriter.WriteString(s)
}

func Get(c *gin.Context) *session.Session {
	if v, ok := c.Get(SessionKey); ok {
		if sess, ok := v.(*session.Session); ok {
			return sess
		}
	}
	return nil
}
