package redis

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/abmcmanu/sessionx/pkg/session"
)

func newTestStore(t *testing.T) *RedisStore {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set")
	}
	s, err := NewRedisStore(Options{Addr: addr, Prefix: "sessionx_test:" + t.Name() + ":", TTL: time.Minute})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestSaveLoadDelete(t *testing.T) {
	s := newTestStore(t)
	now := time.Now().Truncate(time.Second)
	sess := &session.Session{
		ID:        "abc",
		Data:      map[string]interface{}{"user": "alice"},
		CreatedAt: now,
		UpdatedAt: now,
		RotatedAt: now,
	}

	if err := s.Save(sess); err != nil {
		t.Fatal(err)
	}
	loaded, err := s.Load("abc")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != "abc" || loaded.Data["user"] != "alice" || !loaded.UpdatedAt.Equal(now) {
		t.Errorf("unexpected session %+v", loaded)
	}

	ttl := s.GetClient().TTL(s.ctx, s.prefix+"abc").Val()
	if ttl <= 0 || ttl > time.Minute {
		t.Errorf("unexpected ttl %v", ttl)
	}

	if err := s.Delete("abc"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load("abc"); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("got %v, want ErrSessionNotFound", err)
	}
}

func TestInvalidData(t *testing.T) {
	s := newTestStore(t)
	if err := s.GetClient().Set(s.ctx, s.prefix+"bad", "{not json", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load("bad"); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("got %v, want ErrInvalidSession", err)
	}
}

func TestDefaults(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set")
	}
	s, err := NewRedisStore(Options{Addr: addr})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	defer s.Close()
	if s.prefix != "sessionx:" || s.ttl != 24*time.Hour {
		t.Errorf("unexpected defaults prefix=%q ttl=%v", s.prefix, s.ttl)
	}
}

func TestUnreachableServer(t *testing.T) {
	if _, err := NewRedisStore(Options{Addr: "127.0.0.1:1"}); err == nil {
		t.Error("expected connection error")
	}
}

func TestWithManager(t *testing.T) {
	s := newTestStore(t)
	m, err := session.NewManager(session.DevConfig([]byte("0123456789abcdef"), session.WithStore(s)))
	if err != nil {
		t.Fatal(err)
	}
	sess := m.New()
	sess.Data["n"] = 1
	if err := s.Save(sess); err != nil {
		t.Fatal(err)
	}
	m.Rotate(sess)
	if _, err := s.Load(sess.ID); !errors.Is(err, ErrSessionNotFound) {
		t.Error("rotated id should not exist yet")
	}
}
