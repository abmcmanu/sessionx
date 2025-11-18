# SessionX Redis Store

Redis-backed session storage for sessionx, enabling stateless and scalable session management across multiple servers.

## Features

- **Distributed Sessions**: Share sessions across multiple application servers
- **Scalability**: Handle high traffic with Redis cluster support
- **Auto-expiration**: Automatic TTL management in Redis
- **Persistence**: Sessions survive application restarts
- **Performance**: Fast in-memory operations with Redis

## Installation

```bash
go get github.com/abmcmanu/sessionx/optional/store/redis
```

## Requirements

- Redis 6.0 or higher
- Go 1.23 or higher

## Quick Start

```go
package main

import (
    "github.com/abmcmanu/sessionx/pkg/session"
    redisstore "github.com/abmcmanu/sessionx/optional/store/redis"
    "time"
)

func main() {
    // Create Redis store
    store, err := redisstore.NewRedisStore(redisstore.Options{
        Addr:   "localhost:6379",
        Prefix: "session:",
        TTL:    30 * time.Minute,
    })
    if err != nil {
        panic(err)
    }
    defer store.Close()

    // Configure session manager with Redis store
    cfg := session.DefaultConfig(
        []byte("your-32-byte-secret-key-here!"),
        session.WithStore(store),
        session.WithMaxAge(30*time.Minute),
    )

    manager, err := session.NewManager(cfg)
    if err != nil {
        panic(err)
    }

    // Use manager in your application...
}
```

## Configuration Options

```go
type Options struct {
    Addr     string        // Redis server address (default: "localhost:6379")
    Password string        // Redis password (optional)
    DB       int           // Redis database number (default: 0)
    Prefix   string        // Key prefix for sessions (default: "sessionx:")
    TTL      time.Duration // Session TTL in Redis (default: 24h)
}
```

## Usage Examples

### Basic Session Counter

```go
store, _ := redisstore.NewRedisStore(redisstore.Options{
    Addr:   "localhost:6379",
    Prefix: "app:",
    TTL:    1 * time.Hour,
})
defer store.Close()

cfg := session.DevConfig(
    secretKey,
    session.WithStore(store),
)

manager, _ := session.NewManager(cfg)

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)
    count, _ := sess.Data["count"].(float64)
    sess.Data["count"] = count + 1

    fmt.Fprintf(w, "Visits: %.0f (stored in Redis)", sess.Data["count"])
})
```

### Multi-Server Setup

```go
// Same configuration on all servers
store, _ := redisstore.NewRedisStore(redisstore.Options{
    Addr:   "redis-cluster.example.com:6379",
    Prefix: "production:",
    TTL:    24 * time.Hour,
})

cfg := session.DefaultConfig(
    []byte("same-secret-across-all-servers"),
    session.WithStore(store),
    session.WithDomain(".example.com"),
    session.WithMaxAge(24*time.Hour),
)

// Sessions will work across all servers sharing this Redis instance
```

### Authentication with Redis

```go
store, _ := redisstore.NewRedisStore(redisstore.Options{
    Addr:   "localhost:6379",
    Prefix: "auth:",
    TTL:    1 * time.Hour,
})
defer store.Close()

cfg := session.DefaultConfig(
    secretKey,
    session.WithStore(store),
    session.WithMaxAge(1*time.Hour),
)

manager, _ := session.NewManager(cfg)

// Login
http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
    sess := session.Get(r)

    // Authenticate user...
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

    username := sess.Data["username"]
    fmt.Fprintf(w, "Welcome %s!", username)
})
```

### Gin Integration

```go
import (
    sessiongin "github.com/abmcmanu/sessionx/pkg/gin"
    redisstore "github.com/abmcmanu/sessionx/optional/store/redis"
)

store, _ := redisstore.NewRedisStore(redisstore.Options{
    Addr:   "localhost:6379",
    Prefix: "gin:",
    TTL:    30 * time.Minute,
})
defer store.Close()

cfg := session.DevConfig(secretKey, session.WithStore(store))
manager, _ := session.NewManager(cfg)

r := gin.Default()
r.Use(sessiongin.SessionMiddleware(manager))

r.GET("/api/counter", func(c *gin.Context) {
    sess := sessiongin.Get(c)
    count, _ := sess.Data["count"].(float64)
    sess.Data["count"] = count + 1

    c.JSON(200, gin.H{
        "count":   sess.Data["count"],
        "storage": "redis",
    })
})
```

## How It Works

1. **Session Creation**: When a new session is created, only the session ID is stored in the cookie
2. **Session Storage**: The actual session data is serialized to JSON and stored in Redis
3. **Session Loading**: On each request, the session ID from the cookie is used to fetch data from Redis
4. **Session Saving**: Modified session data is saved back to Redis with automatic TTL refresh

### Cookie vs Redis Storage

| Aspect | Cookie-based (default) | Redis Store |
|--------|----------------------|-------------|
| Session data location | Encrypted in cookie | Redis server |
| Cookie size | Large (encrypted data) | Small (only ID) |
| Multi-server | Shared secret needed | Automatic |
| Scalability | Limited | High |
| Data limit | ~4KB | No practical limit |
| Performance | Fast | Very fast (in-memory) |

## Redis Key Structure

Sessions are stored in Redis with the following key format:

```
{prefix}{session_id}
```

Example: `session:abc123def456`

The value is a JSON-serialized session object:

```json
{
  "ID": "abc123def456",
  "Data": {
    "user_id": "12345",
    "logged_in": true
  },
  "CreatedAt": "2024-01-01T10:00:00Z",
  "UpdatedAt": "2024-01-01T10:15:00Z",
  "RotatedAt": "2024-01-01T10:00:00Z"
}
```

## TTL and Expiration

- **Redis TTL**: Automatically set on each save
- **Application MaxAge**: Checked on Load() for additional validation
- **TTL Refresh**: Updated on every session save

```go
store, _ := redisstore.NewRedisStore(redisstore.Options{
    TTL: 30 * time.Minute, // Redis auto-deletes after 30min
})

cfg := session.DefaultConfig(
    secretKey,
    session.WithStore(store),
    session.WithMaxAge(30*time.Minute), // Application-level validation
)
```

## Best Practices

1. **Use the same secret key** across all servers
2. **Set appropriate TTL** based on your security requirements
3. **Use Redis password** in production
4. **Monitor Redis memory** usage
5. **Use Redis Sentinel or Cluster** for high availability
6. **Set Domain attribute** for subdomain sharing

## Error Handling

```go
store, err := redisstore.NewRedisStore(redisstore.Options{
    Addr: "localhost:6379",
})
if err != nil {
    // Connection failed
    log.Fatal("Redis connection error:", err)
}

// Always close when done
defer store.Close()
```

## Performance Considerations

- Redis operations are fast (sub-millisecond)
- Network latency matters - keep Redis close to application servers
- Use connection pooling (handled automatically by go-redis)
- Consider Redis persistence settings for your use case

## Migration from Cookie-based

To migrate from cookie-based to Redis storage:

```go
// Before (cookie-based)
cfg := session.DevConfig(secretKey)

// After (Redis-based)
store, _ := redisstore.NewRedisStore(redisstore.Options{
    Addr: "localhost:6379",
})
cfg := session.DevConfig(
    secretKey,
    session.WithStore(store), // Add this line
)
```

Existing sessions will be recreated automatically on the first request.

## Troubleshooting

### Connection refused

```
Error: dial tcp [::1]:6379: connect: connection refused
```

**Solution**: Ensure Redis is running:

```bash
redis-server
# or
docker run -d -p 6379:6379 redis
```

### Authentication failed

```
Error: NOAUTH Authentication required
```

**Solution**: Add password to options:

```go
Options{
    Addr:     "localhost:6379",
    Password: "your-redis-password",
}
```

### Sessions not persisting

**Check**:
1. Redis is running and accessible
2. TTL is not too short
3. No Redis eviction policies removing keys
4. Sufficient Redis memory

## License

MIT License - see main sessionx repository