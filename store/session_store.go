package store

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

var SessionNotFoundErr = errors.New("session not found or expired")

type Session struct {
	ID        string    `json:"session_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

type SessionStore interface {
	CreateSession(userID int, ttl time.Duration) (Session, error)
	GetSession(sessionID string) (Session, error)
	DeleteSession(sessionID string) error
}

type IMSessionStore struct {
	store  map[string]Session
	nextID int
	mu     sync.Mutex
}

func (ims *IMSessionStore) CreateSession(userID int, ttl time.Duration) (Session, error) {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	id := strconv.Itoa(ims.nextID)
	session := Session{
		ID:        id,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
	ims.store[id] = session
	ims.nextID++

	return session, nil
}

func (ims *IMSessionStore) GetSession(sessionID string) (Session, error) {
	session, ok := ims.store[sessionID]
	if !ok {
		return Session{}, SessionNotFoundErr
	}

	return session, nil
}

func (ims *IMSessionStore) DeleteSession(sessionID string) error {
	if _, ok := ims.store[sessionID]; !ok {
		return errors.New("could not delete. session not found")
	}
	delete(ims.store, sessionID)
	return nil
}
