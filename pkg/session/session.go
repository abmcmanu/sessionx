package session

import "time"

type Session struct {
	ID        string
	Data      map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
	RotatedAt time.Time
}

func (s *Session) AddFlash(key string, value interface{}) {
	flashes, ok := s.Data["_flashes"].(map[string]interface{})
	if !ok {
		flashes = make(map[string]interface{})
		s.Data["_flashes"] = flashes
	}
	flashes[key] = value
}

func (s *Session) GetFlash(key string) (interface{}, bool) {
	flashes, ok := s.Data["_flashes"].(map[string]interface{})
	if !ok {
		return nil, false
	}

	value, exists := flashes[key]
	if exists {
		delete(flashes, key)
		if len(flashes) == 0 {
			delete(s.Data, "_flashes")
		}
	}

	return value, exists
}

func (s *Session) GetFlashes() map[string]interface{} {
	flashes, ok := s.Data["_flashes"].(map[string]interface{})
	if !ok {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	for k, v := range flashes {
		result[k] = v
	}

	delete(s.Data, "_flashes")
	return result
}

func (s *Session) HasFlash(key string) bool {
	flashes, ok := s.Data["_flashes"].(map[string]interface{})
	if !ok {
		return false
	}
	_, exists := flashes[key]
	return exists
}
