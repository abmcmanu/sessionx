package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/abmcmanu/sessionx/pkg/session"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidSession  = errors.New("invalid session data")
)

type RedisStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
	ctx    context.Context
}

type Options struct {
	Addr     string
	Password string
	DB       int
	Prefix   string
	TTL      time.Duration
}

func NewRedisStore(opts Options) (*RedisStore, error) {
	if opts.Prefix == "" {
		opts.Prefix = "sessionx:"
	}

	if opts.TTL == 0 {
		opts.TTL = 24 * time.Hour
	}

	client := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStore{
		client: client,
		prefix: opts.Prefix,
		ttl:    opts.TTL,
		ctx:    ctx,
	}, nil
}

func (s *RedisStore) Load(id string) (*session.Session, error) {
	key := s.prefix + id

	data, err := s.client.Get(s.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	var sess session.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, ErrInvalidSession
	}

	return &sess, nil
}

func (s *RedisStore) Save(sess *session.Session) error {
	key := s.prefix + sess.ID

	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	return s.client.Set(s.ctx, key, data, s.ttl).Err()
}

func (s *RedisStore) Delete(id string) error {
	key := s.prefix + id
	return s.client.Del(s.ctx, key).Err()
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

func (s *RedisStore) SetTTL(ttl time.Duration) {
	s.ttl = ttl
}

func (s *RedisStore) GetClient() *redis.Client {
	return s.client
}