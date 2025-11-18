package main

import (
	"bytes"
	"encoding/json"
	"github.com/abmcmanu/sessionx/pkg/session"
	sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
)

func ExampleTesting() {
	gin.SetMode(gin.TestMode)

	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.POST("/login", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		sess.Data["logged_in"] = true
		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	})

	r.GET("/profile", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		loggedIn, _ := sess.Data["logged_in"].(bool)

		if !loggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile page"})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/login", nil)
	r.ServeHTTP(w1, req1)

	cookies := w1.Result().Cookies()

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/profile", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	r.ServeHTTP(w2, req2)

	var response map[string]interface{}
	_ = json.Unmarshal(w2.Body.Bytes(), &response)
}

func ExampleJSONRequest() {
	gin.SetMode(gin.TestMode)

	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.POST("/api/data", func(c *gin.Context) {
		var data struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sess := sessiongin.Get(c)
		sess.Data["last_name"] = data.Name
		sess.Data["last_value"] = data.Value

		c.JSON(http.StatusOK, gin.H{
			"message": "Data saved to session",
			"data":    sess.Data,
		})
	})

	payload := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}
	jsonData, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/data", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
}

func ExampleMultipleRequests() {
	gin.SetMode(gin.TestMode)

	cfg := session.DevConfig([]byte("0123456789abcdef0123456789abcdef"))
	manager, _ := session.NewManager(cfg)

	r := gin.Default()
	r.Use(sessiongin.SessionMiddleware(manager))

	r.GET("/counter", func(c *gin.Context) {
		sess := sessiongin.Get(c)
		count, _ := sess.Data["count"].(float64)
		sess.Data["count"] = count + 1

		c.JSON(http.StatusOK, gin.H{"count": sess.Data["count"]})
	})

	var cookies []*http.Cookie

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/counter", nil)

		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}

		r.ServeHTTP(w, req)
		cookies = w.Result().Cookies()
	}
}