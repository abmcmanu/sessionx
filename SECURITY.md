# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of SessionX seriously. If you discover a security vulnerability, please follow these guidelines:

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to:

- **Email**: manueolinga@gmail.com
- **Subject**: [SECURITY] SessionX Vulnerability Report

### What to Include

When reporting a vulnerability, please include:

1. **Description**: Clear description of the vulnerability
2. **Impact**: What can an attacker achieve?
3. **Steps to Reproduce**: Detailed steps to reproduce the issue
4. **Proof of Concept**: Code sample demonstrating the vulnerability (if applicable)
5. **Affected Versions**: Which versions are affected?
6. **Suggested Fix**: Your ideas for fixing the issue (optional)

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity, typically:
  - Critical: 1-7 days
  - High: 7-14 days
  - Medium: 14-30 days
  - Low: 30-90 days

### Disclosure Policy

- We follow **Coordinated Disclosure**
- We will keep you informed of the progress
- We will credit you in the security advisory (unless you prefer to remain anonymous)
- Please allow us time to fix the issue before public disclosure

## Security Best Practices

When using SessionX:

1. **Use Strong Secret Keys**
   ```go
   // Generate a secure 32-byte key
   secretKey := make([]byte, 32)
   _, err := rand.Read(secretKey)
   ```

2. **Enable HTTPS in Production**
   ```go
   cfg := session.DefaultConfig(secretKey) // Secure: true
   ```

3. **Set SameSite Attribute**
   ```go
   session.WithSameSite("Strict") // Prevents CSRF
   ```

4. **Rotate Sessions on Privilege Changes**
   ```go
   manager.Rotate(sess) // After login, role changes
   ```

5. **Use Appropriate Session Lifetimes**
   ```go
   session.WithMaxAge(30*time.Minute) // For sensitive operations
   ```

6. **Validate Session Data**
   ```go
   userID, ok := sess.Data["user_id"].(string)
   if !ok || userID == "" {
       // Handle invalid session
   }
   ```

## Known Security Considerations

### Cookie Size Limits

- Cookie-based sessions limited to ~4KB
- Use Redis store for larger session data
- Avoid storing sensitive data directly in sessions

### Session Fixation

- SessionX includes automatic rotation
- Always rotate after authentication
- Configurable rotation intervals

### XSS Protection

- HttpOnly flag prevents JavaScript access
- Always sanitize user input
- Never echo session data to HTML without escaping

### CSRF Protection

- SameSite attribute helps prevent CSRF
- Consider additional CSRF tokens for state-changing operations
- Use "Strict" or "Lax" SameSite mode

## Security Features

SessionX includes these security features:

-  **AES-GCM Encryption**: Authenticated encryption for cookie data
-  **Session Expiration**: Automatic timeout and validation
-  **Session Rotation**: Prevents session fixation attacks
-  **HttpOnly Cookies**: Prevents XSS attacks
-  **Secure Flag**: Ensures HTTPS-only transmission
-  **SameSite Attribute**: CSRF protection
-  **Secret Key Validation**: Enforces proper key lengths (16/24/32 bytes)

## Secure Configuration Examples

### Production Configuration

```go
cfg := session.DefaultConfig(
    secretKey,
    session.WithMaxAge(30*time.Minute),
    session.WithSameSite("Strict"),
    session.WithDomain(".yourapp.com"),
    session.WithRotationInterval(15*time.Minute),
)
```

### High-Security Application

```go
cfg := session.DefaultConfig(
    secretKey,
    session.WithMaxAge(15*time.Minute),       // Short lifetime
    session.WithSameSite("Strict"),           // Strict CSRF
    session.WithRotationInterval(5*time.Minute), // Frequent rotation
)
```

## Vulnerability History

No vulnerabilities have been reported yet.

## Credits

We would like to thank the following people for responsibly disclosing security issues:

- (No reports yet)

## Contact

For security-related questions that are not vulnerabilities, you can open a GitHub issue or discussion.

---

Thank you for helping keep SessionX and its users safe!