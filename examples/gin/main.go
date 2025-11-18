package main

import (
	"github.com/abmcmanu/sessionx/pkg/session"
	sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithMaxAge(2*time.Hour),
		session.WithCookieName("gin_session"),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.Use(sessiongin.SessionMiddleware(manager))

	r.GET("/", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		c.JSON(http.StatusOK, gin.H{
			"session_id": sess.ID,
			"visits":     sess.Data["count"],
			"created_at": sess.CreatedAt,
			"updated_at": sess.UpdatedAt,
		})
	})

	r.POST("/login", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		sess.Data["user_id"] = "12345"
		sess.Data["username"] = "john_doe"
		sess.Data["role"] = "admin"
		sess.Data["logged_in"] = true

		c.JSON(http.StatusOK, gin.H{
			"message": "Logged in successfully",
			"user":    sess.Data["username"],
		})
	})

	r.GET("/dashboard", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		loggedIn, _ := sess.Data["logged_in"].(bool)

		if !loggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		username, _ := sess.Data["username"].(string)
		role, _ := sess.Data["role"].(string)

		c.JSON(http.StatusOK, gin.H{
			"message":  "Welcome to dashboard",
			"username": username,
			"role":     role,
		})
	})

	r.POST("/logout", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		sess.Data = make(map[string]interface{})

		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out successfully",
		})
	})

	_ = r.Run(":8080")
}