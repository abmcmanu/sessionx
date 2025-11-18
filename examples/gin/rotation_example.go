package main

import (
	"github.com/abmcmanu/sessionx/pkg/session"
	sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func ExampleAutomaticRotation() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithRotationInterval(30*time.Second),
	)
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.GET("/", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		timeSinceRotation := time.Since(sess.RotatedAt)

		c.JSON(http.StatusOK, gin.H{
			"session_id":        sess.ID,
			"visits":            sess.Data["count"],
			"last_rotated":      sess.RotatedAt,
			"time_since_rotation": timeSinceRotation.Seconds(),
			"message":           "Session will rotate automatically after 30 seconds",
		})
	})

	_ = r.Run(":8080")
}

func ExampleManualRotationOnLogin() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.POST("/auth/login", func(c *gin.Context) {
		var loginData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&loginData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sess := sessiongin.Get(c)

		oldID := sess.ID

		sess.Data["user_id"] = "12345"
		sess.Data["username"] = loginData.Username
		sess.Data["logged_in"] = true

		manager.Rotate(sess)

		c.JSON(http.StatusOK, gin.H{
			"message":        "Login successful",
			"old_session_id": oldID,
			"new_session_id": sess.ID,
			"rotated":        true,
		})
	})

	r.GET("/auth/status", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		loggedIn, _ := sess.Data["logged_in"].(bool)

		c.JSON(http.StatusOK, gin.H{
			"logged_in":  loggedIn,
			"session_id": sess.ID,
			"rotated_at": sess.RotatedAt,
		})
	})

	_ = r.Run(":8080")
}

func ExampleRotationWithMetrics() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithRotationInterval(1*time.Minute),
		session.WithMaxAge(30*time.Minute),
	)
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.GET("/session/info", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		c.JSON(http.StatusOK, gin.H{
			"session_id": sess.ID,
			"created_at": sess.CreatedAt,
			"updated_at": sess.UpdatedAt,
			"rotated_at": sess.RotatedAt,
			"age":        time.Since(sess.CreatedAt).Seconds(),
			"rotation_age": time.Since(sess.RotatedAt).Seconds(),
			"data":       sess.Data,
		})
	})

	r.POST("/session/rotate", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		oldID := sess.ID

		manager.Rotate(sess)

		c.JSON(http.StatusOK, gin.H{
			"message":        "Session rotated manually",
			"old_session_id": oldID,
			"new_session_id": sess.ID,
		})
	})

	_ = r.Run(":8080")
}