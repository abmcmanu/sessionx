package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

var (
	ErrInvalidSession = errors.New("invalid or corrupted session")
)

type Manager struct {
	cfg Config
}

func NewManager(cfg Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(m.cfg.SecretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return base64.RawStdEncoding.EncodeToString(encrypted), nil
}

func (m *Manager) decrypt(encoded string) ([]byte, error) {
	raw, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(m.cfg.SecretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return nil, ErrInvalidSession
	}

	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (m *Manager) Load(r *http.Request) (*Session, error) {
	c, err := r.Cookie(m.cfg.CookieName)
	if err != nil {
		// no session — create new
		return m.New(), nil
	}

	decrypted, err := m.decrypt(c.Value)
	if err != nil {
		return m.New(), nil // treat as no session
	}

	var sess Session
	if err := json.Unmarshal(decrypted, &sess); err != nil {
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
		return err
	}

	encrypted, err := m.encrypt(raw)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     m.cfg.CookieName,
		Value:    encrypted,
		Path:     m.cfg.Path,
		HttpOnly: m.cfg.HttpOnly,
		Secure:   m.cfg.Secure,
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
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
