package session

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

var testKey = []byte("0123456789abcdef0123456789abcdef")

type memStore struct {
	mu   sync.Mutex
	data map[string]Session
}

func newMemStore() *memStore {
	return &memStore{data: map[string]Session{}}
}

func (s *memStore) Load(id string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.data[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &sess, nil
}

func (s *memStore) Save(sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[sess.ID] = *sess
	return nil
}

func (s *memStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, id)
	return nil
}

func (s *memStore) len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}

func newTestManager(t *testing.T, opts ...ConfigOption) *Manager {
	t.Helper()
	m, err := NewManager(DevConfig(testKey, opts...))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("cookie %q not set", name)
	return nil
}

func requestWith(c *http.Cookie) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if c != nil {
		r.AddCookie(c)
	}
	return r
}

func TestNewManagerKeyLength(t *testing.T) {
	for _, n := range []int{16, 24, 32} {
		if _, err := NewManager(DevConfig(make([]byte, n))); err != nil {
			t.Errorf("key length %d: unexpected error %v", n, err)
		}
	}
	for _, n := range []int{0, 5, 31, 64} {
		_, err := NewManager(DevConfig(make([]byte, n)))
		if !errors.Is(err, ErrInvalidSecretKey) {
			t.Errorf("key length %d: got %v, want ErrInvalidSecretKey", n, err)
		}
		var se *SessionError
		if !errors.As(err, &se) || se.Op != "NewManager" {
			t.Errorf("key length %d: expected SessionError with Op NewManager", n)
		}
	}
}

func TestConfigDefaults(t *testing.T) {
	prod := DefaultConfig(testKey)
	if !prod.Secure || !prod.HttpOnly || prod.SameSite != "Lax" || prod.MaxAge != 24*time.Hour {
		t.Errorf("unexpected default config: %+v", prod)
	}
	if DevConfig(testKey).Secure {
		t.Error("dev config should not be secure")
	}
	cfg := DefaultConfig(testKey, WithCookieName("x"), WithPath("/api"), WithDomain("example.com"),
		WithMaxAge(time.Minute), WithSecure(false), WithHttpOnly(false), WithSameSite("Strict"),
		WithRotationInterval(time.Second))
	if cfg.CookieName != "x" || cfg.Path != "/api" || cfg.Domain != "example.com" || cfg.MaxAge != time.Minute ||
		cfg.Secure || cfg.HttpOnly || cfg.SameSite != "Strict" || cfg.RotationInterval != time.Second {
		t.Errorf("options not applied: %+v", cfg)
	}
}

func TestCookieRoundTrip(t *testing.T) {
	m := newTestManager(t)

	sess, err := m.Load(requestWith(nil))
	if err != nil {
		t.Fatal(err)
	}
	sess.Data["user"] = "alice"
	sess.Data["count"] = 3

	rec := httptest.NewRecorder()
	if err := m.Save(rec, sess); err != nil {
		t.Fatal(err)
	}
	c := sessionCookie(t, rec, "sessionx")
	if !c.HttpOnly || c.Path != "/" || c.SameSite != http.SameSiteLaxMode {
		t.Errorf("unexpected cookie attributes: %+v", c)
	}
	if strings.Contains(c.Value, "alice") {
		t.Error("cookie value is not encrypted")
	}

	loaded, err := m.Load(requestWith(c))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != sess.ID {
		t.Errorf("id mismatch: got %q want %q", loaded.ID, sess.ID)
	}
	if loaded.Data["user"] != "alice" || loaded.Data["count"] != float64(3) {
		t.Errorf("data mismatch: %v", loaded.Data)
	}
}

func TestTamperedCookieStartsNewSession(t *testing.T) {
	m := newTestManager(t)
	sess := m.New()
	sess.Data["role"] = "user"
	rec := httptest.NewRecorder()
	if err := m.Save(rec, sess); err != nil {
		t.Fatal(err)
	}
	c := sessionCookie(t, rec, "sessionx")

	raw := []byte(c.Value)
	if raw[len(raw)/2] == 'A' {
		raw[len(raw)/2] = 'B'
	} else {
		raw[len(raw)/2] = 'A'
	}

	for _, value := range []string{string(raw), "not-base64!!", "", "AAAA"} {
		loaded, err := m.Load(requestWith(&http.Cookie{Name: "sessionx", Value: value}))
		if err != nil {
			t.Fatal(err)
		}
		if loaded.ID == sess.ID || len(loaded.Data) != 0 {
			t.Errorf("value %q: tampered cookie was accepted", value)
		}
	}
}

func TestCookieFromOtherKeyRejected(t *testing.T) {
	a := newTestManager(t)
	b, err := NewManager(DevConfig([]byte("fedcba9876543210fedcba9876543210")))
	if err != nil {
		t.Fatal(err)
	}
	sess := a.New()
	rec := httptest.NewRecorder()
	_ = a.Save(rec, sess)

	loaded, _ := b.Load(requestWith(sessionCookie(t, rec, "sessionx")))
	if loaded.ID == sess.ID {
		t.Error("cookie encrypted with another key was accepted")
	}
}

func TestExpiredSession(t *testing.T) {
	m := newTestManager(t, WithMaxAge(time.Hour))
	sess := m.New()
	sess.Data["x"] = 1
	sess.UpdatedAt = time.Now().Add(-2 * time.Hour)

	raw, err := json.Marshal(sess)
	if err != nil {
		t.Fatal(err)
	}
	value, err := m.encrypt(raw)
	if err != nil {
		t.Fatal(err)
	}

	loaded, _ := m.Load(requestWith(&http.Cookie{Name: "sessionx", Value: value}))
	if loaded.ID == sess.ID {
		t.Error("expired session was accepted")
	}
}

func TestAutomaticRotation(t *testing.T) {
	m := newTestManager(t, WithRotationInterval(time.Minute))
	sess := m.New()
	sess.RotatedAt = time.Now().Add(-2 * time.Minute)
	sess.Data["keep"] = "me"

	store := newMemStore()
	ms := newTestManager(t, WithRotationInterval(time.Minute), WithStore(store))
	_ = store.Save(sess)

	loaded, _ := ms.Load(requestWith(&http.Cookie{Name: "sessionx", Value: sess.ID}))
	if loaded.ID == sess.ID {
		t.Fatal("session was not rotated")
	}
	if loaded.Data["keep"] != "me" {
		t.Error("data lost on rotation")
	}
	if _, err := store.Load(sess.ID); err == nil {
		t.Error("old session id still present in store after rotation")
	}

	fresh := m.New()
	id := fresh.ID
	rec := httptest.NewRecorder()
	_ = m.Save(rec, fresh)
	again, _ := m.Load(requestWith(sessionCookie(t, rec, "sessionx")))
	if again.ID != id {
		t.Error("session rotated before interval elapsed")
	}
}

func TestManualRotation(t *testing.T) {
	m := newTestManager(t)
	sess := m.New()
	old := sess.ID
	before := sess.RotatedAt
	time.Sleep(time.Millisecond)
	m.Rotate(sess)
	if sess.ID == old || !sess.RotatedAt.After(before) {
		t.Error("Rotate did not change id or timestamp")
	}
}

func TestUniqueIDs(t *testing.T) {
	m := newTestManager(t)
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		id := m.New().ID
		if seen[id] {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = true
	}
}

func TestStoreMode(t *testing.T) {
	store := newMemStore()
	m := newTestManager(t, WithStore(store))

	sess, _ := m.Load(requestWith(nil))
	sess.Data["cart"] = []interface{}{"a", "b"}
	rec := httptest.NewRecorder()
	if err := m.Save(rec, sess); err != nil {
		t.Fatal(err)
	}
	c := sessionCookie(t, rec, "sessionx")
	if c.Value != sess.ID {
		t.Errorf("cookie should hold only the session id, got %q", c.Value)
	}

	loaded, _ := m.Load(requestWith(c))
	if loaded.ID != sess.ID || len(loaded.Data["cart"].([]interface{})) != 2 {
		t.Errorf("unexpected loaded session: %+v", loaded)
	}

	unknown, _ := m.Load(requestWith(&http.Cookie{Name: "sessionx", Value: "missing"}))
	if unknown.ID == "missing" {
		t.Error("unknown id should start a new session")
	}
}

func TestStoreNilDataIsInitialized(t *testing.T) {
	store := newMemStore()
	m := newTestManager(t, WithStore(store))
	_ = store.Save(&Session{ID: "abc", UpdatedAt: time.Now(), RotatedAt: time.Now()})

	sess, _ := m.Load(requestWith(&http.Cookie{Name: "sessionx", Value: "abc"}))
	if sess.Data == nil {
		t.Fatal("Data should be initialized")
	}
	sess.AddFlash("k", "v")
}

func TestCookieTooLarge(t *testing.T) {
	m := newTestManager(t)
	sess := m.New()
	sess.Data["blob"] = strings.Repeat("x", 5000)
	rec := httptest.NewRecorder()
	if err := m.Save(rec, sess); !errors.Is(err, ErrCookieTooLarge) {
		t.Fatalf("got %v, want ErrCookieTooLarge", err)
	}
	if rec.Header().Get("Set-Cookie") != "" {
		t.Error("oversized cookie should not be written")
	}
}

func TestDestroy(t *testing.T) {
	store := newMemStore()
	m := newTestManager(t, WithStore(store), WithDomain("example.com"), WithSameSite("Strict"))

	var id string
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/logout" {
			_ = m.Destroy(w, r)
			_, _ = w.Write([]byte("bye"))
			return
		}
		sess := Get(r)
		sess.Data["user"] = "bob"
		id = sess.ID
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	c := sessionCookie(t, rec, "sessionx")
	if store.len() != 1 {
		t.Fatalf("expected 1 stored session, got %d", store.len())
	}

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	req.AddCookie(c)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if store.len() != 0 {
		t.Errorf("session %q still in store after destroy", id)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected exactly one Set-Cookie after destroy, got %d", len(cookies))
	}
	if cookies[0].MaxAge >= 0 || cookies[0].Domain != "example.com" || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Errorf("unexpected expiry cookie: %+v", cookies[0])
	}
}

func TestMiddlewarePersistsAcrossRequests(t *testing.T) {
	m := newTestManager(t)
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := Get(r)
		n, _ := sess.Data["n"].(float64)
		sess.Data["n"] = n + 1
		w.WriteHeader(http.StatusOK)
	}))

	var c *http.Cookie
	for i := 1; i <= 3; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, requestWith(c))
		if got := len(rec.Result().Header.Values("Set-Cookie")); got != 1 {
			t.Fatalf("request %d: expected 1 Set-Cookie, got %d", i, got)
		}
		c = sessionCookie(t, rec, "sessionx")
	}

	sess, _ := m.Load(requestWith(c))
	if sess.Data["n"] != float64(3) {
		t.Errorf("got n=%v, want 3", sess.Data["n"])
	}
}

func TestMiddlewareSavesWithoutWrite(t *testing.T) {
	m := newTestManager(t)
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Get(r).Data["x"] = true
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, requestWith(nil))
	sessionCookie(t, rec, "sessionx")
}

func TestMiddlewareFlushAndUnwrap(t *testing.T) {
	m := newTestManager(t)
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("flush failed: %v", err)
		}
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, requestWith(nil))
	if !rec.Flushed {
		t.Error("response was not flushed")
	}
	sessionCookie(t, rec, "sessionx")
}

func TestGetWithoutMiddleware(t *testing.T) {
	if Get(httptest.NewRequest(http.MethodGet, "/", nil)) != nil {
		t.Error("expected nil session")
	}
}

func TestFlashes(t *testing.T) {
	m := newTestManager(t)
	sess := m.New()

	if _, ok := sess.GetFlash("missing"); ok {
		t.Error("unexpected flash")
	}
	sess.AddFlash("success", "saved")
	sess.AddFlash("error", "oops")
	if !sess.HasFlash("success") {
		t.Error("HasFlash should be true")
	}

	rec := httptest.NewRecorder()
	_ = m.Save(rec, sess)
	loaded, _ := m.Load(requestWith(sessionCookie(t, rec, "sessionx")))

	v, ok := loaded.GetFlash("success")
	if !ok || v != "saved" {
		t.Errorf("got %v %v", v, ok)
	}
	if loaded.HasFlash("success") {
		t.Error("flash should be consumed")
	}
	all := loaded.GetFlashes()
	if len(all) != 1 || all["error"] != "oops" {
		t.Errorf("unexpected flashes %v", all)
	}
	if _, exists := loaded.Data["_flashes"]; exists {
		t.Error("flash storage should be removed once empty")
	}
	if len(loaded.GetFlashes()) != 0 {
		t.Error("flashes should be empty")
	}
}

func TestParseSameSite(t *testing.T) {
	cases := map[string]http.SameSite{
		"Strict": http.SameSiteStrictMode,
		"strict": http.SameSiteStrictMode,
		"None":   http.SameSiteNoneMode,
		"Lax":    http.SameSiteLaxMode,
		"":       http.SameSiteLaxMode,
		"bogus":  http.SameSiteLaxMode,
	}
	for in, want := range cases {
		if got := parseSameSite(in); got != want {
			t.Errorf("parseSameSite(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestSessionErrorFormat(t *testing.T) {
	err := newError("Save", ErrMarshalFailed)
	if err.Error() != "session Save: failed to marshal session data" {
		t.Errorf("unexpected message %q", err.Error())
	}
	if newError("x", nil) != nil {
		t.Error("newError(nil) should return nil")
	}
}
