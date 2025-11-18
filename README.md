# SessionX 

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight, secure, and flexible session management library for Go web applications. SessionX provides encrypted cookie-based sessions with optional Redis backend for scalability.

## ✨ Features

- 🔒 **Secure by Default**: AES-GCM encryption, HttpOnly, Secure, SameSite cookies
- 🚀 **Flexible Storage**: Cookie-based (default) or Redis for distributed systems
- 🔄 **Session Rotation**: Automatic and manual session ID rotation
- 💬 **Flash Messages**: Built-in support for one-time notifications
- 🎯 **Framework Agnostic**: Works with net/http, Gin, and any Go web framework
- 📦 **Modular Design**: Optional Redis store, no forced dependencies
- 🛠️ **Developer Friendly**: Simple API, functional options pattern
- ⚡ **Production Ready**: Comprehensive error handling, expiration validation

## 📋 Table of Contents 

- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Configuration](#️-configuration)
- [Usage Examples](#-usage-examples)
- [Session Rotation](#-session-rotation)
- [Flash Messages](#-flash-messages)
- [Redis Store](#️-redis-store)
- [Framework Integration](#-framework-integration)
- [API Reference](#-api-reference)
- [Security](#-security)
- [Testing](#-testing)
- [Architecture](#-architecture)
- [Why SessionX?](#-why-sessionx)
- [Contributing](#-contributing)
- [License](#-license)


## 🚀 Installation

### Core Package

```bash
go get github.com/abmcmanu/sessionx
```

### Optional Packages

```bash
# Gin framework integration
go get github.com/abmcmanu/sessionx/pkg/gin

# Redis store for scalability
go get github.com/abmcmanu/sessionx/optional/store/redis
```

## ⚡ Quick Start

### Basic Example (net/http)

```go
package main

import (
    "fmt"
    "github.com/abmcmanu/sessionx/pkg/session"
    "net/http"
    "time"
)

func main() {
    // Create session manager
    cfg := session.DevConfig(
        []byte("your-32-byte-secret-key-here!"),
        session.WithMaxAge(2*time.Hour),
    )

    manager, err := session.NewManager(cfg)
    if err != nil {
        panic(err)
    }

    // Setup routes
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        sess := session.Get(r)

        count, _ := sess.Data["count"].(float64)
        sess.Data["count"] = count + 1

        fmt.Fprintf(w, "Visits: %.0f", sess.Data["count"])
    })

    // Apply middleware
    http.ListenAndServe(":8080", manager.Middleware(mux))
}
```

### Gin Framework

```go
import (
    "github.com/abmcmanu/sessionx/pkg/session"
    sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
    "github.com/gin-gonic/gin"
)

func main() {
    cfg := session.DevConfig([]byte("your-32-byte-secret-key-here!"))
    manager, _ := session.NewManager(cfg)

    r := gin.Default()
    r.Use(sessiongin.SessionMiddleware(manager))

    r.GET("/", func(c *gin.Context) {
        sess := sessiongin.Get(c)
        count, _ := sess.Data["count"].(float64)
        sess.Data["count"] = count + 1

        c.JSON(200, gin.H{"visits": sess.Data["count"]})
    })

    r.Run(":8080")
}
```

## ⚙️ Configuration

### Configuration Helpers

SessionX provides two configuration helpers:

```go
// Development (allows HTTP for localhost)
cfg := session.DevConfig(secretKey)

// Production (requires HTTPS)
cfg := session.DefaultConfig(secretKey)
```

### Functional Options

Customize your configuration with options:

```go
cfg := session.DefaultConfig(
    []byte("your-32-byte-secret-key-here!"),
    session.WithMaxAge(24*time.Hour),
    session.WithCookieName("my_session"),
    session.WithDomain(".example.com"),
    session.WithSameSite("Strict"),
    session.WithRotationInterval(15*time.Minute),
)
```

### Available Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithMaxAge(d time.Duration)` | Session lifetime | 24 hours |
| `WithCookieName(name string)` | Cookie name | "sessionx" |
| `WithDomain(domain string)` | Cookie domain | "" |
| `WithPath(path string)` | Cookie path | "/" |
| `WithSecure(bool)` | Secure flag | DevConfig: false, DefaultConfig: true |
| `WithHttpOnly(bool)` | HttpOnly flag | true |
| `WithSameSite(string)` | SameSite attribute | "Lax" |
| `WithRotationInterval(d time.Duration)` | Auto-rotation interval | 15 minutes |
| `WithStore(store Store)` | External store (Redis) | nil (cookie-based) |

## 📖 Usage Examples

### Authentication

```go
// Login
http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)

    // Validate credentials...
    sess.Data["user_id"] = "12345"
    sess.Data["username"] = "john"
    sess.Data["logged_in"] = true

    // Rotate session ID for security
    manager.Rotate(sess)

    http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
})

// Protected route
http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)

    if loggedIn, _ := sess.Data["logged_in"].(bool); !loggedIn {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    fmt.Fprintf(w, "Welcome %s!", sess.Data["username"])
})

// Logout
http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
    manager.Destroy(w, r)
    http.Redirect(w, r, "/", http.StatusSeeOther)
})
```

### Shopping Cart

```go
http.HandleFunc("/cart/add", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)

    cart, ok := sess.Data["cart"].([]interface{})
    if !ok {
        cart = []interface{}{}
    }

    cart = append(cart, map[string]interface{}{
        "product_id": r.FormValue("product_id"),
        "quantity":   r.FormValue("quantity"),
    })

    sess.Data["cart"] = cart
})
```

## 🔄 Session Rotation

Session rotation prevents session fixation attacks by changing the session ID.

### Automatic Rotation

```go
cfg := session.DefaultConfig(
    secretKey,
    session.WithRotationInterval(15*time.Minute),
)
```

Sessions automatically rotate after the specified interval.

### Manual Rotation

```go
// Rotate on security-critical operations
http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)

    // Authenticate user...
    sess.Data["user_id"] = "12345"

    // Rotate immediately after login
    manager.Rotate(sess)

    http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
})
```

## 💬 Flash Messages

Flash messages are one-time notifications that survive a single redirect.

```go
// Set flash message
http.HandleFunc("/action", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)

    sess.AddFlash("success", "Operation completed!")
    sess.AddFlash("info", "Check your email")

    http.Redirect(w, r, "/result", http.StatusSeeOther)
})

// Read flash messages (auto-deleted after reading)
http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)

    if msg, exists := sess.GetFlash("success"); exists {
        fmt.Fprintf(w, "Success: %s\n", msg)
    }

    // Or get all flashes at once
    flashes := sess.GetFlashes()
    for key, value := range flashes {
        fmt.Fprintf(w, "[%s]: %v\n", key, value)
    }
})
```

### Flash API

- `AddFlash(key, value)` - Add a flash message
- `GetFlash(key)` - Get and remove a single flash
- `GetFlashes()` - Get and remove all flashes
- `HasFlash(key)` - Check if flash exists without removing

## 🗄️ Redis Store

For multi-server deployments, use Redis to store sessions.

### Setup

```go
import redisstore "github.com/abmcmanu/sessionx/optional/store/redis"

store, err := redisstore.NewRedisStore(redisstore.Options{
    Addr:   "localhost:6379",
    Prefix: "session:",
    TTL:    30 * time.Minute,
})
if err != nil {
    panic(err)
}
defer store.Close()

cfg := session.DefaultConfig(
    secretKey,
    session.WithStore(store),
)

manager, _ := session.NewManager(cfg)
```

### Multi-Server Configuration

```go
// Same configuration on all servers
store, _ := redisstore.NewRedisStore(redisstore.Options{
    Addr:   "redis-cluster.example.com:6379",
    Prefix: "prod:",
    TTL:    24 * time.Hour,
})

cfg := session.DefaultConfig(
    []byte("same-secret-across-all-servers"),
    session.WithStore(store),
    session.WithDomain(".example.com"),
)
```

### Benefits

- ✅ Sessions work across multiple servers
- ✅ Sessions survive application restarts
- ✅ Better scalability for high traffic
- ✅ Smaller cookies (only session ID)

See [Redis Store Documentation](./optional/store/redis/README.md) for details.

## 🎨 Framework Integration

### net/http (Standard Library)

```go
mux := http.NewServeMux()
mux.HandleFunc("/", handler)

http.ListenAndServe(":8080", manager.Middleware(mux))
```

Access session: `sess := session.Get(r)`

### Gin

```go
import sessiongin "github.com/abmcmanu/sessionx/pkg/gin"

r := gin.Default()
r.Use(sessiongin.SessionMiddleware(manager))

r.GET("/", handler)
```

Access session: `sess := sessiongin.Get(c)`

### Other Frameworks

SessionX is framework-agnostic. Wrap your handler/middleware to call `manager.Load()` and `manager.Save()`.

## 📚 API Reference

### Manager

```go
// Create manager
func NewManager(cfg Config) (*Manager, error)

// Load session from request
func (m *Manager) Load(r *http.Request) (*Session, error)

// Create new session
func (m *Manager) New() *Session

// Save session to response
func (m *Manager) Save(w http.ResponseWriter, sess *Session) error

// Destroy session
func (m *Manager) Destroy(w http.ResponseWriter, r *http.Request) error

// Rotate session ID
func (m *Manager) Rotate(sess *Session)
```

### Session

```go
type Session struct {
    ID        string
    Data      map[string]interface{}
    CreatedAt time.Time
    UpdatedAt time.Time
    RotatedAt time.Time
}

// Flash messages
func (s *Session) AddFlash(key string, value interface{})
func (s *Session) GetFlash(key string) (interface{}, bool)
func (s *Session) GetFlashes() map[string]interface{}
func (s *Session) HasFlash(key string) bool
```

### Store Interface

Implement this interface for custom storage backends:

```go
type Store interface {
    Load(id string) (*Session, error)
    Save(sess *Session) error
    Delete(id string) error
}
```

## 🔒 Security

### Best Practices

1. **Use strong secret keys** (32 bytes recommended)
   ```go
   secretKey := []byte("your-32-byte-secret-key-here!")
   ```

2. **Enable HTTPS in production**
   ```go
   cfg := session.DefaultConfig(secretKey) // Secure: true
   ```

3. **Set appropriate SameSite**
   ```go
   session.WithSameSite("Strict") // or "Lax", "None"
   ```

4. **Rotate sessions on privilege changes**
   ```go
   manager.Rotate(sess) // After login, role changes
   ```

5. **Set reasonable session lifetimes**
   ```go
   session.WithMaxAge(30*time.Minute) // Sensitive apps
   session.WithMaxAge(24*time.Hour)   // Regular apps
   ```

### Security Features

- ✅ AES-GCM encryption (authenticated encryption)
- ✅ Automatic session expiration
- ✅ Session ID rotation
- ✅ HttpOnly cookies (prevents XSS)
- ✅ Secure flag for HTTPS
- ✅ SameSite attribute (CSRF protection)
- ✅ Domain validation

See [SECURITY.md](./SECURITY.md) for vulnerability reporting.

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./pkg/session
go test ./pkg/gin
go test ./optional/store/redis
```

## 📊 Architecture

### Cookie-based Mode (Default)

```
Request → Load (decrypt cookie) → Handler (modify session) → Save (encrypt cookie) → Response
```

- Session data encrypted in cookie
- No server-side storage needed
- Works with single server
- ~4KB cookie limit

### Redis Mode

```
Request → Load (fetch from Redis) → Handler (modify session) → Save (store in Redis) → Response
                      ↓                                                      ↓
                  Session ID in cookie                            Session data in Redis
```

- Only session ID in cookie
- Session data in Redis
- Works across multiple servers
- No practical size limit

## 🌟 Why SessionX?

| Feature | SessionX | gorilla/sessions | scs |
|---------|----------|------------------|-----|
| Cookie encryption | ✅ AES-GCM | ✅ AES | ❌ |
| Redis support | ✅ Optional | ✅ Required | ✅ |
| Flash messages | ✅ Built-in | ✅ | ❌ |
| Session rotation | ✅ Automatic/Manual | ❌ | ❌ |
| Gin integration | ✅ Separate package | ❌ | ❌ |
| Functional options | ✅ | ❌ | ✅ |
| Go version | 1.23+ | 1.18+ | 1.20+ |

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [cookie-session](https://github.com/expressjs/cookie-session) (Node.js)
- Built with [go-redis](https://github.com/redis/go-redis)
- Framework integrations: [Gin](https://github.com/gin-gonic/gin)

## 📮 Support

- 🐛 [Report bugs](https://github.com/abmcmanu/sessionx/issues)
- 💡 [Request features](https://github.com/abmcmanu/sessionx/issues)
- 📖 [Documentation](https://github.com/abmcmanu/sessionx)
- 🔒 [Security issues](./SECURITY.md)

---

Made with ❤️ by [abmcmanu](https://github.com/abmcmanu)
