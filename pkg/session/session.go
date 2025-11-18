package session

import "time"

type Session struct {
	ID        string
	Data      map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}
