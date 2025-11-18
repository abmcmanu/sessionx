# SessionX Examples - Gin Framework

This directory contains examples of using SessionX with the Gin web framework.

## Files

- **main.go** - Basic API example with JSON responses
- **examples.go** - Multiple use cases (auth, cart, i18n, middleware)
- **testing_example.go** - Testing patterns with httptest

## Quick Start

Install dependencies:

```bash
go get github.com/gin-gonic/gin
go get github.com/abmcmanu/sessionx/pkg/session
go get github.com/abmcmanu/sessionx/pkg/gin
```

Run the basic example:

```bash
cd examples/gin
go run main.go
```

Test with curl:

```bash
# Visit counter
curl http://localhost:8080/

# Login
curl -X POST http://localhost:8080/login

# Dashboard (requires login)
curl http://localhost:8080/dashboard

# Logout
curl -X POST http://localhost:8080/logout
```

## Gin Integration

### Setup

```go
import (
    "github.com/abmcmanu/sessionx/pkg/session"
    sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
    "github.com/gin-gonic/gin"
)

func main() {
    cfg := session.DevConfig([]byte("your-32-byte-secret-key-here!"))
    manager, err := session.NewManager(cfg)
    if err != nil {
        panic(err)
    }

    r := gin.Default()

    // Add session middleware
    r.Use(sessiongin.SessionMiddleware(manager))

    r.Run(":8080")
}
```

### Access Session

```go
r.GET("/", func(c *gin.Context) {
    sess := sessiongin.Get(c)

    // Read data
    count, _ := sess.Data["count"].(float64)

    // Write data
    sess.Data["count"] = count + 1

    c.JSON(http.StatusOK, gin.H{"count": sess.Data["count"]})
})
```

## Usage Examples

### 1. Visit Counter

```go
r.GET("/api/counter", func(c *gin.Context) {
    sess := sessiongin.Get(c)
    count, _ := sess.Data["count"].(float64)
    sess.Data["count"] = count + 1

    c.JSON(http.StatusOK, gin.H{"count": sess.Data["count"]})
})
```

### 2. User Authentication

```go
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
```

### 3. Auth Middleware

```go
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        sess := sessiongin.Get(c)
        loggedIn, _ := sess.Data["logged_in"].(bool)

        if !loggedIn {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Authentication required",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// Usage
protected := r.Group("/api")
protected.Use(AuthRequired())
{
    protected.GET("/profile", profileHandler)
    protected.GET("/settings", settingsHandler)
}
```

### 4. Shopping Cart

```go
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
```

### 5. Language Preference

```go
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
    }

    c.JSON(http.StatusOK, gin.H{
        "message": messages[lang],
        "language": lang,
    })
})
```

## Configuration Examples

### Development (Localhost)

```go
cfg := session.DevConfig(
    []byte("0123456789abcdef0123456789abcdef"),
    session.WithMaxAge(1*time.Hour),
    session.WithCookieName("gin_session"),
)
```

### Production (HTTPS)

```go
cfg := session.DefaultConfig(
    []byte("your-production-secret-key-32bytes"),
    session.WithMaxAge(24*time.Hour),
    session.WithSameSite("Strict"),
)
```

### Short-lived Sessions

```go
cfg := session.DefaultConfig(
    secretKey,
    session.WithMaxAge(30*time.Minute),
    session.WithSameSite("Strict"),
)
```

## API Patterns

### JSON Response with Session Data

```go
r.GET("/api/session", func(c *gin.Context) {
    sess := sessiongin.Get(c)

    c.JSON(http.StatusOK, gin.H{
        "session_id": sess.ID,
        "data":       sess.Data,
        "created_at": sess.CreatedAt,
        "updated_at": sess.UpdatedAt,
    })
})
```

### Error Handling

```go
r.GET("/api/protected", func(c *gin.Context) {
    sess := sessiongin.Get(c)

    if sess == nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Session error",
        })
        c.Abort()
        return
    }

    // Handler logic...
})
```

### Conditional Logic

```go
r.GET("/api/user", func(c *gin.Context) {
    sess := sessiongin.Get(c)

    userID, exists := sess.Data["user_id"].(string)
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Not authenticated",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{"user_id": userID})
})
```

## Best Practices

1. **Always check session existence**
   ```go
   sess := sessiongin.Get(c)
   if sess == nil {
       c.JSON(http.StatusInternalServerError, gin.H{"error": "Session error"})
       c.Abort()
       return
   }
   ```

2. **Use middleware for authentication**
   ```go
   protected := r.Group("/api")
   protected.Use(AuthRequired())
   ```

3. **Type assertions with ok check**
   ```go
   username, ok := sess.Data["username"].(string)
   if !ok {
       // Handle missing or wrong type
   }
   ```

4. **Use c.Abort() to stop middleware chain**
   ```go
   if !authenticated {
       c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
       c.Abort()
       return
   }
   ```

5. **Clear sensitive data on logout**
   ```go
   sess.Data = make(map[string]interface{})
   ```

## Common Issues

### Session not persisting

Make sure you're using `DevConfig()` for localhost or serving over HTTPS:

```go
// For localhost
cfg := session.DevConfig(secretKey)

// For production HTTPS
cfg := session.DefaultConfig(secretKey)
```

### CORS issues with cookies

Configure CORS to allow credentials:

```go
import "github.com/gin-contrib/cors"

r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowCredentials: true,
    AllowHeaders:     []string{"Content-Type"},
}))
```

### Session data not available

Ensure middleware is registered before routes:

```go
r.Use(sessiongin.SessionMiddleware(manager)) // ← Before routes

r.GET("/", handler) // ← After middleware
```