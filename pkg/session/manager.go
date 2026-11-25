package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxCookieSize = 4096

type Manager struct {
	cfg  Config
	aead cipher.AEAD
}

func NewManager(cfg Config) (*Manager, error) {
	keyLen := len(cfg.SecretKey)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, newError("NewManager", ErrInvalidSecretKey)
	}

	block, err := aes.NewCipher(cfg.SecretKey)
	if err != nil {
		return nil, newError("NewManager", ErrInvalidSecretKey)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, newError("NewManager", err)
	}

	return &Manager{cfg: cfg, aead: aead}, nil
}

func (m *Manager) encrypt(data []byte) (string, error) {
	nonce := make([]byte, m.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", newError("encrypt", ErrEncryptionFailed)
	}

	encrypted := m.aead.Seal(nonce, nonce, data, nil)
	return base64.RawStdEncoding.EncodeToString(encrypted), nil
}

func (m *Manager) decrypt(encoded string) ([]byte, error) {
	raw, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, newError("decrypt", ErrInvalidSession)
	}

	nonceSize := m.aead.NonceSize()
	if len(raw) < nonceSize+m.aead.Overhead() {
		return nil, newError("decrypt", ErrInvalidSession)
	}

	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	decrypted, err := m.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, newError("decrypt", ErrDecryptionFailed)
	}

	return decrypted, nil
}

func (m *Manager) Load(r *http.Request) (*Session, error) {
	c, err := r.Cookie(m.cfg.CookieName)
	if err != nil {
		return m.New(), nil
	}

	var sess *Session

	if m.cfg.Store != nil {
		sess, err = m.cfg.Store.Load(c.Value)
		if err != nil {
			return m.New(), nil
		}
	} else {
		decrypted, err := m.decrypt(c.Value)
		if err != nil {
			return m.New(), nil
		}

		var s Session
		if err := json.Unmarshal(decrypted, &s); err != nil {
			return m.New(), nil
		}
		sess = &s
	}

	if sess == nil {
		return m.New(), nil
	}
	if sess.Data == nil {
		sess.Data = map[string]interface{}{}
	}

	if m.cfg.MaxAge > 0 && time.Since(sess.UpdatedAt) > m.cfg.MaxAge {
		if m.cfg.Store != nil {
			_ = m.cfg.Store.Delete(sess.ID)
		}
		return m.New(), nil
	}

	if m.cfg.RotationInterval > 0 && time.Since(sess.RotatedAt) > m.cfg.RotationInterval {
		m.Rotate(sess)
	}

	return sess, nil
}

func (m *Manager) New() *Session {
	now := time.Now()
	return &Session{
		ID:        m.newID(),
		Data:      map[string]interface{}{},
		CreatedAt: now,
		UpdatedAt: now,
		RotatedAt: now,
	}
}

func (m *Manager) Save(w http.ResponseWriter, sess *Session) error {
	if sess == nil || sess.destroyed {
		return nil
	}

	sess.UpdatedAt = time.Now()

	var cookieValue string

	if m.cfg.Store != nil {
		if err := m.cfg.Store.Save(sess); err != nil {
			return err
		}
		cookieValue = sess.ID
	} else {
		raw, err := json.Marshal(sess)
		if err != nil {
			return newError("Save", ErrMarshalFailed)
		}

		encrypted, err := m.encrypt(raw)
		if err != nil {
			return err
		}
		cookieValue = encrypted
	}

	cookie := m.cookie(cookieValue, int(m.cfg.MaxAge.Seconds()))
	encoded := cookie.String()
	if len(encoded) > maxCookieSize {
		return newError("Save", ErrCookieTooLarge)
	}

	w.Header().Add("Set-Cookie", encoded)
	return nil
}

func (m *Manager) cookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     m.cfg.CookieName,
		Value:    value,
		Path:     m.cfg.Path,
		Domain:   m.cfg.Domain,
		HttpOnly: m.cfg.HttpOnly,
		Secure:   m.cfg.Secure,
		SameSite: parseSameSite(m.cfg.SameSite),
		MaxAge:   maxAge,
	}
}

func (m *Manager) Destroy(w http.ResponseWriter, r *http.Request) error {
	sess := Get(r)
	if sess != nil {
		sess.destroyed = true
	}

	if m.cfg.Store != nil {
		if sess != nil {
			_ = m.cfg.Store.Delete(sess.ID)
		}
		if c, err := r.Cookie(m.cfg.CookieName); err == nil {
			_ = m.cfg.Store.Delete(c.Value)
		}
	}

	http.SetCookie(w, m.cookie("", -1))
	return nil
}

func (m *Manager) Rotate(sess *Session) {
	if m.cfg.Store != nil && sess.ID != "" {
		_ = m.cfg.Store.Delete(sess.ID)
	}
	sess.ID = m.newID()
	sess.RotatedAt = time.Now()
}

func (m *Manager) newID() string {
	b := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
