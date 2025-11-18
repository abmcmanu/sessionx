package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Manager struct {
	cfg Config
}

func NewManager(cfg Config) (*Manager, error) {
	keyLen := len(cfg.SecretKey)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, newError("NewManager", ErrInvalidSecretKey)
	}

	return &Manager{cfg: cfg}, nil
}

func (m *Manager) encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(m.cfg.SecretKey)
	if err != nil {
		return "", newError("encrypt", ErrEncryptionFailed)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", newError("encrypt", ErrEncryptionFailed)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", newError("encrypt", ErrEncryptionFailed)
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return base64.RawStdEncoding.EncodeToString(encrypted), nil
}

func (m *Manager) decrypt(encoded string) ([]byte, error) {
	raw, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, newError("decrypt", ErrInvalidSession)
	}

	block, err := aes.NewCipher(m.cfg.SecretKey)
	if err != nil {
		return nil, newError("decrypt", ErrDecryptionFailed)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, newError("decrypt", ErrDecryptionFailed)
	}

	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return nil, newError("decrypt", ErrInvalidSession)
	}

	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
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

	decrypted, err := m.decrypt(c.Value)
	if err != nil {
		return m.New(), nil
	}

	var sess Session
	if err := json.Unmarshal(decrypted, &sess); err != nil {
		return m.New(), nil
	}

	if m.cfg.MaxAge > 0 && time.Since(sess.UpdatedAt) > m.cfg.MaxAge {
		return m.New(), nil
	}

	return &sess, nil
}

func (m *Manager) New() *Session {
	return &Session{
		ID:        m.newID(),
		Data:      map[string]interface{}{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (m *Manager) Save(w http.ResponseWriter, sess *Session) error {
	sess.UpdatedAt = time.Now()

	raw, err := json.Marshal(sess)
	if err != nil {
		return newError("Save", ErrMarshalFailed)
	}

	encrypted, err := m.encrypt(raw)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     m.cfg.CookieName,
		Value:    encrypted,
		Path:     m.cfg.Path,
		Domain:   m.cfg.Domain,
		HttpOnly: m.cfg.HttpOnly,
		Secure:   m.cfg.Secure,
		SameSite: parseSameSite(m.cfg.SameSite),
		MaxAge:   int(m.cfg.MaxAge.Seconds()),
	}

	w.Header().Add("Set-Cookie", cookie.String())
	return nil
}

func (m *Manager) Destroy(w http.ResponseWriter) {
	expired := &http.Cookie{
		Name:     m.cfg.CookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     m.cfg.Path,
		HttpOnly: true,
	}
	http.SetCookie(w, expired)
}

func (m *Manager) newID() string {
	b := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func parseSameSite(s string) http.SameSite {
	switch s {
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	case "Lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}
