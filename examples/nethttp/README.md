# SessionX Examples - net/http

This directory contains examples of using SessionX with Go's standard `net/http` package.

## Files

- **main.go** - Basic usage example with visit counter
- **examples.go** - Multiple configuration examples for different scenarios

## Quick Start

Run the basic example:

```bash
cd examples/nethttp
go run main.go
```

Visit `http://localhost:8080` to see the visit counter increment.

## Configuration Examples

### 1. Development (Localhost)

Perfect for local development - `Secure=false` allows HTTP cookies.

```go
cfg := session.DevConfig(
    []byte("0123456789abcdef0123456789abcdef"),
    session.WithMaxAge(1*time.Hour),
)
```

### 2. Production (HTTPS)

Full security enabled - requires HTTPS for cookies.

```go
cfg := session.DefaultConfig(
    []byte("your-production-secret-key-32bytes"),
    session.WithMaxAge(24*time.Hour),
    session.WithCookieName("prod_session"),
)
```

### 3. Short-lived Sessions

For sensitive operations (banking, admin panels).

```go
cfg := session.DefaultConfig(
    secretKey,
    session.WithMaxAge(30*time.Minute),
    session.WithSameSite("Strict"), // Maximum CSRF protection
)
```

### 4. Multi-subdomain Application

Share sessions across subdomains (app.example.com, api.example.com).

```go
cfg := session.DefaultConfig(
    secretKey,
    session.WithDomain(".example.com"),
    session.WithMaxAge(7*24*time.Hour),
)
```

### 5. API with Custom Path

Limit sessions to specific paths only.

```go
cfg := session.DevConfig(
    secretKey,
    session.WithPath("/api"),
    session.WithCookieName("api_session"),
)
```

## Available Configuration Options

| Option | Description | Example |
|--------|-------------|---------|
| `WithMaxAge(d)` | Session lifetime | `WithMaxAge(2*time.Hour)` |
| `WithCookieName(name)` | Custom cookie name | `WithCookieName("session")` |
| `WithDomain(domain)` | Cookie domain | `WithDomain(".example.com")` |
| `WithPath(path)` | Cookie path | `WithPath("/api")` |
| `WithSecure(bool)` | HTTPS only | `WithSecure(true)` |
| `WithHttpOnly(bool)` | Protect against XSS | `WithHttpOnly(true)` |
| `WithSameSite(policy)` | CSRF protection | `WithSameSite("Strict")` |
| `WithRotationInterval(d)` | ID rotation | `WithRotationInterval(30*time.Minute)` |

## Usage Patterns

### Storing User Data

```go
mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)
    sess.Data["user_id"] = "12345"
    sess.Data["username"] = "john_doe"
    sess.Data["logged_in"] = true
})
```

### Reading Session Data

```go
mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)
    username, ok := sess.Data["username"].(string)
    if !ok {
        http.Error(w, "Not logged in", http.StatusUnauthorized)
        return
    }
    fmt.Fprintf(w, "Welcome, %s!", username)
})
```

### Destroying Session

```go
mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
    manager.Destroy(w)
    fmt.Fprintf(w, "Logged out!")
})
```

## Security Best Practices

1. **Always use a strong secret key (32 bytes)**
   ```go
   secretKey := []byte("your-32-byte-secret-key-here!")
   ```

2. **Use `DefaultConfig()` in production** - Ensures `Secure=true`

3. **Set appropriate `MaxAge`** - Shorter for sensitive apps

4. **Use `SameSite="Strict"`** for maximum CSRF protection

5. **Enable session rotation** for long-lived sessions
   ```go
   session.WithRotationInterval(15*time.Minute)
   ```

## Common Issues

### Sessions not persisting on Safari

Make sure you're using `DevConfig()` for localhost or serve over HTTPS:

```go
// For localhost development
cfg := session.DevConfig(secretKey)

// For production with HTTPS
cfg := session.DefaultConfig(secretKey)
```

### Session increments by 2 instead of 1

Filter out browser requests like favicon:

```go
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    // Your handler code here
})
```

### Sessions not shared across subdomains

Set the domain with a leading dot:

```go
session.WithDomain(".example.com") // Works for all *.example.com
```