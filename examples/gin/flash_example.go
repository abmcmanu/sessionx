package main

import (
	"github.com/abmcmanu/sessionx/pkg/session"
	sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ExampleFlashMessages() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.POST("/api/action", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		sess.AddFlash("success", "Action completed successfully!")
		sess.AddFlash("info", "3 items updated")

		c.JSON(http.StatusOK, gin.H{
			"message":  "Action completed",
			"redirect": "/api/result",
		})
	})

	r.GET("/api/result", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		flashes := sess.GetFlashes()

		c.JSON(http.StatusOK, gin.H{
			"flashes": flashes,
			"message": "Flashes will be empty on next request",
		})
	})

	_ = r.Run(":8080")
}

func ExampleAuthenticationFlash() {
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

		if loginData.Username == "admin" && loginData.Password == "password" {
			sess.Data["user_id"] = "12345"
			sess.Data["username"] = loginData.Username
			sess.Data["logged_in"] = true

			manager.Rotate(sess)
			sess.AddFlash("success", "Welcome back! Login successful.")

			c.JSON(http.StatusOK, gin.H{
				"message":  "Login successful",
				"redirect": "/dashboard",
			})
		} else {
			sess.AddFlash("error", "Invalid credentials")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":    "Invalid credentials",
				"redirect": "/login",
			})
		}
	})

	r.GET("/dashboard", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		response := gin.H{
			"message": "Dashboard",
		}

		if successMsg, exists := sess.GetFlash("success"); exists {
			response["flash_success"] = successMsg
		}

		loggedIn, _ := sess.Data["logged_in"].(bool)
		if !loggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":    "Not authenticated",
				"redirect": "/login",
			})
			return
		}

		response["username"] = sess.Data["username"]

		c.JSON(http.StatusOK, response)
	})

	r.GET("/login", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		response := gin.H{
			"message": "Login page",
		}

		if errorMsg, exists := sess.GetFlash("error"); exists {
			response["flash_error"] = errorMsg
		}

		c.JSON(http.StatusOK, response)
	})

	_ = r.Run(":8080")
}

func ExampleFormValidationFlash() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.POST("/api/profile/update", func(c *gin.Context) {
		var profileData struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}

		if err := c.ShouldBindJSON(&profileData); err != nil {
			sess := sessiongin.Get(c)
			sess.AddFlash("error", "Invalid request data")

			c.JSON(http.StatusBadRequest, gin.H{
				"error":    err.Error(),
				"redirect": "/profile",
			})
			return
		}

		sess := sessiongin.Get(c)

		if profileData.Email == "" {
			sess.AddFlash("error", "Email is required")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Email is required",
				"redirect": "/profile",
			})
			return
		}

		if profileData.Name == "" {
			sess.AddFlash("error", "Name is required")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Name is required",
				"redirect": "/profile",
			})
			return
		}

		sess.Data["email"] = profileData.Email
		sess.Data["name"] = profileData.Name
		sess.AddFlash("success", "Profile updated successfully!")

		c.JSON(http.StatusOK, gin.H{
			"message":  "Profile updated",
			"redirect": "/profile",
		})
	})

	r.GET("/profile", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		response := gin.H{
			"email": sess.Data["email"],
			"name":  sess.Data["name"],
		}

		if errorMsg, exists := sess.GetFlash("error"); exists {
			response["flash_error"] = errorMsg
		}

		if successMsg, exists := sess.GetFlash("success"); exists {
			response["flash_success"] = successMsg
		}

		c.JSON(http.StatusOK, response)
	})

	_ = r.Run(":8080")
}

func ExampleMultipleFlashTypes() {
	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.POST("/api/bulk-action", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		sess.AddFlash("success", "10 items updated")
		sess.AddFlash("warning", "2 items skipped")
		sess.AddFlash("error", "1 item failed")
		sess.AddFlash("info", "Processing completed in 2.5s")

		c.JSON(http.StatusOK, gin.H{
			"message":  "Bulk action completed",
			"redirect": "/api/results",
		})
	})

	r.GET("/api/results", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		response := gin.H{
			"message": "Results page",
		}

		if success, exists := sess.GetFlash("success"); exists {
			response["success"] = success
		}

		if warning, exists := sess.GetFlash("warning"); exists {
			response["warning"] = warning
		}

		if errorMsg, exists := sess.GetFlash("error"); exists {
			response["error"] = errorMsg
		}

		if info, exists := sess.GetFlash("info"); exists {
			response["info"] = info
		}

		c.JSON(http.StatusOK, response)
	})

	r.GET("/api/results/check", func(c *gin.Context) {
		sess := sessiongin.Get(c)

		c.JSON(http.StatusOK, gin.H{
			"has_success": sess.HasFlash("success"),
			"has_error":   sess.HasFlash("error"),
			"has_warning": sess.HasFlash("warning"),
			"has_info":    sess.HasFlash("info"),
		})
	})

	_ = r.Run(":8080")
}