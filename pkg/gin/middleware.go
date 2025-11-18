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

		c.Next()

		_ = manager.Save(c.Writer, sess)
	}
}

func Get(c *gin.Context) *session.Session {
	if v, exists := c.Get(SessionKey); exists {
		return v.(*session.Session)
	}
	return nil
}