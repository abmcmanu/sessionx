package session

type Store interface {
	Load(id string) (*Session, error)
	Save(sess *Session) error
	Delete(id string) error
}