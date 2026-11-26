package store

import (
	"crypto/rand"
	"testing"
	"time"
)

func TestCreateSession(t *testing.T) {
	t.Run("success on creating session", func(t *testing.T) {
		stubStore := IMSessionStore{
			Store: map[string]Session{},
		}
		userID := 1
		ttl := 24 * time.Hour
		session, err := stubStore.CreateSession(userID, ttl)
		if err != nil {
			t.Fatalf("expected no error, but got: %v", err)
		}

		if session.UserID != userID {
			t.Errorf("got: %v, want: %v as userID", session.UserID, userID)
		}
	})

	t.Run("store not initialized", func(t *testing.T) {
		stubStore := IMSessionStore{}
		userID := 1
		ttl := 24 * time.Hour
		_, err := stubStore.CreateSession(userID, ttl)
		if err == nil {
			t.Fatalf("expected an error: %v, but got nil", UninitializedStoreErr)
		}
	})
}

func TestGetSession(t *testing.T) {
	t.Run("sucess on getting a sesssion", func(t *testing.T) {
		id := rand.Text()
		stubSession := Session{
			ID:        id,
			UserID:    1,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}
		stubStore := IMSessionStore{
			Store: map[string]Session{id: stubSession},
		}

		session, err := stubStore.GetSession(id)
		if err != nil {
			t.Fatalf("expected no error, but got: %v", err)
		}

		assertSession(t, session, stubSession)
	})

	t.Run("store not initialized", func(t *testing.T) {
		stubStore := IMSessionStore{}

		_, err := stubStore.GetSession("sessionid")
		if err == nil {
			t.Fatalf("expected an error: %v, but nil", UninitializedStoreErr)
		}

		assertErr(t, err, UninitializedStoreErr)
	})

	t.Run("non-existent session-id", func(t *testing.T) {
		id := rand.Text()
		stubSession := Session{
			ID:        id,
			UserID:    1,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}
		stubStore := IMSessionStore{
			Store: map[string]Session{id: stubSession},
		}

		_, err := stubStore.GetSession("different_id")
		if err == nil {
			t.Fatalf("expected an error: %v, but got nil", SessionNotFoundErr)
		}

		assertErr(t, err, SessionNotFoundErr)
	})
}

func assertSession(t testing.TB, got, want Session) {
	t.Helper()
	if got != want {
		t.Errorf("got: %+v, want: %+v as value of session", got, want)
	}
}
