package main

import (
	"github.com/abmcmanu/sessionx/pkg/session"
	sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func ExampleBasicAPI() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.GET("/api/counter", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		c.JSON(http.StatusOK, gin.H{"count": sess.Data["count"]})
	})

	_ = r.Run(":8080")
}

func ExampleAuthentication() {
	cfg := session.DefaultConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithMaxAge(30*time.Minute),
		session.WithSameSite("Strict"),
	)
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
		sess.Data["user_id"] = "12345"
		sess.Data["username"] = loginData.Username
		sess.Data["logged_in"] = true

		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	})

	r.POST("/auth/logout", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		sess.Data = make(map[string]interface{})

		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
	})

	_ = r.Run(":8080")
}

func ExampleAuthMiddleware() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	authRequired := func(c *gin.Context) {
		sess := sessiongin.Get(c)
		loggedIn, _ := sess.Data["logged_in"].(bool)

		if !loggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		c.Next()
	}

	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Public endpoint"})
	})

	protected := r.Group("/api")
	protected.Use(authRequired)
	{
		protected.GET("/profile", func(c *gin.Context) {
			sess := sessiongin.Get(c)
			c.JSON(http.StatusOK, gin.H{
				"username": sess.Data["username"],
				"user_id":  sess.Data["user_id"],
			})
		})

		protected.GET("/settings", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "User settings"})
		})
	}

	_ = r.Run(":8080")
}

func ExampleShoppingCart() {
	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithMaxAge(24*time.Hour),
	)
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.POST("/cart/add", func(c *gin.Context) {
		var item struct {
			ProductID string  `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
		}

		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sess := sessiongin.Get(c)

		cart, ok := sess.Data["cart"].([]interface{})
		if !ok {
			cart = []interface{}{}
		}

		cart = append(cart, map[string]interface{}{
			"product_id": item.ProductID,
			"quantity":   item.Quantity,
			"price":      item.Price,
		})

		sess.Data["cart"] = cart

		c.JSON(http.StatusOK, gin.H{
			"message": "Item added to cart",
			"cart":    cart,
		})
	})

	r.GET("/cart", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		cart, _ := sess.Data["cart"].([]interface{})

		c.JSON(http.StatusOK, gin.H{"cart": cart})
	})

	r.DELETE("/cart/clear", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		delete(sess.Data, "cart")

		c.JSON(http.StatusOK, gin.H{"message": "Cart cleared"})
	})

	_ = r.Run(":8080")
}

func ExampleMultiLanguage() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.POST("/language", func(c *gin.Context) {
		var data struct {
			Language string `json:"language"`
		}

		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sess := sessiongin.Get(c)
		sess.Data["language"] = data.Language

		c.JSON(http.StatusOK, gin.H{"message": "Language preference saved"})
	})

	r.GET("/welcome", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		lang, ok := sess.Data["language"].(string)
		if !ok {
			lang = "en"
		}

		messages := map[string]string{
			"en": "Welcome!",
			"fr": "Bienvenue!",
			"es": "¡Bienvenido!",
			"de": "Willkommen!",
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  messages[lang],
			"language": lang,
		})
	})

	_ = r.Run(":8080")
}