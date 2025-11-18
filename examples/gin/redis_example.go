package main

import (
	"github.com/abmcmanu/sessionx/pkg/session"
	sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
	redisstore "github.com/abmcmanu/sessionx/optional/store/redis"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func ExampleRedisBasic() {
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

func ExampleRedisAuthentication() {
	store, err := redisstore.NewRedisStore(redisstore.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Prefix:   "gin_auth:",
		TTL:      1 * time.Hour,
	})
	if err != nil {
		panic(err)
	}
	defer store.Close()

	cfg := session.DefaultConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithStore(store),
		session.WithMaxAge(1*time.Hour),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

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

		if loginData.Username == "admin" && loginData.Password == "password" {
			sess.Data["user_id"] = "12345"
			sess.Data["username"] = loginData.Username
			sess.Data["logged_in"] = true

			manager.Rotate(sess)
			sess.AddFlash("success", "Login successful!")

			c.JSON(http.StatusOK, gin.H{
				"message":    "Login successful",
				"session_id": sess.ID,
				"storage":    "redis",
			})
		} else {
			sess.AddFlash("error", "Invalid credentials")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
			})
		}
	})

	r.GET("/auth/profile", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		response := gin.H{}

		if successMsg, exists := sess.GetFlash("success"); exists {
			response["flash_success"] = successMsg
		}

		loggedIn, _ := sess.Data["logged_in"].(bool)
		if !loggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Not authenticated",
			})
			return
		}

		response["username"] = sess.Data["username"]
		response["user_id"] = sess.Data["user_id"]
		response["session_id"] = sess.ID
		response["storage"] = "redis"

		c.JSON(http.StatusOK, response)
	})

	r.POST("/auth/logout", func(c *gin.Context) {
		_ = manager.Destroy(c.Writer, c.Request)

		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out successfully",
			"info":    "Session deleted from Redis",
		})
	})

	_ = r.Run(":8080")
}

func ExampleRedisAPIWithFlash() {
	store, err := redisstore.NewRedisStore(redisstore.Options{
		Addr:   "localhost:6379",
		Prefix: "api_session:",
		TTL:    24 * time.Hour,
	})
	if err != nil {
		panic(err)
	}
	defer store.Close()

	cfg := session.DefaultConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithStore(store),
		session.WithMaxAge(24*time.Hour),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.POST("/api/order", func(c *gin.Context) {
		var orderData struct {
			ProductID string  `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Amount    float64 `json:"amount"`
		}

		if err := c.ShouldBindJSON(&orderData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sess := sessiongin.Get(c)

		sess.Data["last_order_id"] = "ORD-12345"
		sess.Data["last_order_amount"] = orderData.Amount

		sess.AddFlash("success", "Order placed successfully!")
		sess.AddFlash("info", "Confirmation email sent")

		c.JSON(http.StatusOK, gin.H{
			"message":  "Order created",
			"order_id": "ORD-12345",
		})
	})

	r.GET("/api/order/status", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		response := gin.H{
			"last_order_id":     sess.Data["last_order_id"],
			"last_order_amount": sess.Data["last_order_amount"],
		}

		if successMsg, exists := sess.GetFlash("success"); exists {
			response["flash_success"] = successMsg
		}

		if infoMsg, exists := sess.GetFlash("info"); exists {
			response["flash_info"] = infoMsg
		}

		c.JSON(http.StatusOK, response)
	})

	_ = r.Run(":8080")
}

func ExampleRedisMultiServer() {
	store, err := redisstore.NewRedisStore(redisstore.Options{
		Addr:   "redis-cluster.example.com:6379",
		Prefix: "app_session:",
		TTL:    24 * time.Hour,
	})
	if err != nil {
		panic(err)
	}
	defer store.Close()

	cfg := session.DefaultConfig(
		[]byte("shared-secret-key-across-servers"),
		session.WithStore(store),
		session.WithDomain(".example.com"),
		session.WithMaxAge(24*time.Hour),
		session.WithRotationInterval(15*time.Minute),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.GET("/api/info", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		c.JSON(http.StatusOK, gin.H{
			"server":     c.Request.Host,
			"session_id": sess.ID,
			"visits":     sess.Data["count"],
			"storage":    "redis",
			"message":    "Session works across multiple servers",
		})
	})

	r.GET("/api/session", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		c.JSON(http.StatusOK, gin.H{
			"session_id": sess.ID,
			"created_at": sess.CreatedAt,
			"updated_at": sess.UpdatedAt,
			"rotated_at": sess.RotatedAt,
			"data":       sess.Data,
			"storage":    "redis",
		})
	})

	_ = r.Run(":8080")
}

func ExampleRedisShoppingCart() {
	store, err := redisstore.NewRedisStore(redisstore.Options{
		Addr:   "localhost:6379",
		Prefix: "cart_session:",
		TTL:    2 * time.Hour,
	})
	if err != nil {
		panic(err)
	}
	defer store.Close()

	cfg := session.DevConfig(
		[]byte("0123456789abcdef0123456789abcdef"),
		session.WithStore(store),
		session.WithMaxAge(2*time.Hour),
	)

	manager, err := session.NewManager(cfg)
	if err != nil {
		panic(err)
	}

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
			"storage": "redis - persists across servers",
		})
	})

	r.GET("/cart", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		cart, _ := sess.Data["cart"].([]interface{})

		c.JSON(http.StatusOK, gin.H{
			"cart":    cart,
			"storage": "redis",
		})
	})

	r.DELETE("/cart/clear", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		delete(sess.Data, "cart")

		c.JSON(http.StatusOK, gin.H{
			"message": "Cart cleared",
		})
	})

	_ = r.Run(":8080")
}