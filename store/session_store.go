package store

import (
	"crypto/rand"
	"errors"
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
	store map[string]Session
	mu    sync.Mutex
}

func (ims *IMSessionStore) CreateSession(userID int, ttl time.Duration) (Session, error) {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	storeInitialized := ims.checkStoreInitilaization()
	if !storeInitialized {
		return Session{}, UninitializedStoreErr
	}

	id := rand.Text()
	session := Session{
		ID:        id,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
	ims.store[id] = session

	return session, nil
}

func (ims *IMSessionStore) GetSession(sessionID string) (Session, error) {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	storeInitialized := ims.checkStoreInitilaization()
	if !storeInitialized {
		return Session{}, UninitializedStoreErr
	}

	session, ok := ims.store[sessionID]
	if !ok {
		return Session{}, SessionNotFoundErr
	}
	return session, nil
}

func (ims *IMSessionStore) DeleteSession(sessionID string) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	storeInitialized := ims.checkStoreInitilaization()
	if !storeInitialized {
		return UninitializedStoreErr
	}

	if _, ok := ims.store[sessionID]; !ok {
		return SessionNotFoundErr
	}
	delete(ims.store, sessionID)
	return nil
}

func (ims *IMSessionStore) checkStoreInitilaization() bool {
	return ims.store != nil
}
