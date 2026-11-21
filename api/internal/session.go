package session

import (
	"errors"
	"time"
)

var SessionNotFoundErr = errors.New("session not found or expired")

type Session struct {
	ID        string    `json:"session_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s Session) isExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

type SessionStore interface {
	CreateSession(userID int, ttl time.Duration) (Session, error)
	GetSession(sessionID string) (Session, error)
	DeleteSession(sessionID string) error
}
