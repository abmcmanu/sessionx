package main

import (
	"net/http"
	"time"

	redisstore "github.com/abmcmanu/sessionx/optional/store/redis"
	sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
	"github.com/abmcmanu/sessionx/pkg/session"
	"github.com/gin-gonic/gin"
)

func main() {
	store, err := redisstore.NewRedisStore(redisstore.Options{
		Addr:   "localhost:6379",
		Prefix: "gin_session:",
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

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.GET("/api/counter", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		c.JSON(http.StatusOK, gin.H{
			"session_id": sess.ID,
			"count":      sess.Data["count"],
			"storage":    "redis",
		})
	})

	_ = r.Run(":8080")
}
