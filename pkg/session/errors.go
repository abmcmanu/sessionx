package session

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSession   = errors.New("invalid or corrupted session")
	ErrSessionExpired   = errors.New("session has expired")
	ErrInvalidSecretKey = errors.New("secret key must be 16, 24, or 32 bytes")
	ErrDecryptionFailed = errors.New("failed to decrypt session data")
	ErrMarshalFailed    = errors.New("failed to marshal session data")
	ErrUnmarshalFailed  = errors.New("failed to unmarshal session data")
	ErrEncryptionFailed = errors.New("failed to encrypt session data")
)

type SessionError struct {
	Op  string
	Err error
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("session %s: %v", e.Op, e.Err)
}

func (e *SessionError) Unwrap() error {
	return e.Err
}

func (e *SessionError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

func newError(op string, err error) error {
	if err == nil {
		return nil
	}
	return &SessionError{Op: op, Err: err}
}