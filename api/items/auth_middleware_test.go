package items

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/habx47/gapi/store"
)

type StubSessionStore struct {
	expired bool
	store   map[string]store.Session
	store.SessionStore
}

func (ss *StubSessionStore) GetSession(sessionID string) (store.Session, error) {
	if ss.expired {
		return store.Session{
			ID:        "stub_session_id",
			UserID:    1,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now(),
		}, nil
	}
	return store.Session{
		ID:        "stub_session_id",
		UserID:    1,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(2 * time.Minute),
	}, nil
}

func (ss *StubSessionStore) DeleteSession(sessionID string) error {
	return nil
}

type StubHandler_auth struct {
	ID     string
	UserID int
}

func (sh *StubHandler_auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := r.Context().Value(sessionKey).(store.Session)
	sh.ID = session.ID
	sh.UserID = session.UserID
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("required cookie not set", func(t *testing.T) {
		stubSessionStore := StubSessionStore{}
		stubHandler := StubHandler_auth{}
		request := httptest.NewRequest(http.MethodGet, "/items", nil)
		response := httptest.NewRecorder()

		request.AddCookie(&http.Cookie{
			Name:  "invalid_session_id",
			Value: "this_does_not_matter",
		})

		AuthMiddleware(&stubSessionStore)(&stubHandler).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("expired cookie", func(t *testing.T) {
		stubSessionStore := StubSessionStore{expired: true}
		stubHandler := StubHandler_auth{}
		request := httptest.NewRequest(http.MethodGet, "/items", nil)
		response := httptest.NewRecorder()

		request.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: "stub_session_id",
		})

		AuthMiddleware(&stubSessionStore)(&stubHandler).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("successful authentication", func(t *testing.T) {
		stubSessionStore := StubSessionStore{}
		stubHandler := StubHandler_auth{}
		request := httptest.NewRequest(http.MethodGet, "/items", nil)
		response := httptest.NewRecorder()

		request.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: "stub_session_id",
		})

		AuthMiddleware(&stubSessionStore)(&stubHandler).ServeHTTP(response, request)

		assertResult(t, StubHandler_auth{ID: stubHandler.ID, UserID: stubHandler.UserID}, StubHandler_auth{ID: "stub_session_id", UserID: 1})
	})
}

func assertResult(t testing.TB, got, want StubHandler_auth) {
	t.Helper()
	if got != want {
		t.Errorf("got: %+v, want: %+v", got, want)
	}
}
