package items

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	session "github.com/habx47/gapi/api/internal"
)

type StubSessionStore struct {
	store map[string]session.Session
	session.SessionStore
}

func (ss *StubSessionStore) GetSession(sessionID string) (session.Session, error) {
	return session.Session{
		ID:        "stub_session_id",
		UserID:    1,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(2 * time.Minute),
	}, nil
}

func (ss *StubSessionStore) DeleteSession(sessionID string) error {
	return nil
}

type StubHandler struct {
	ID     string
	UserID int
}

func (sh *StubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := r.Context().Value(sessionKey).(session.Session)
	sh.ID = session.ID
	sh.UserID = session.UserID
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("required cookie not set", func(t *testing.T) {
		stubSessionStore := StubSessionStore{}
		stubHandler := StubHandler{}
		request := httptest.NewRequest(http.MethodGet, "/items", nil)
		response := httptest.NewRecorder()

		request.AddCookie(&http.Cookie{
			Name:  "invalid_session_id",
			Value: "this_does_not_matter",
		})

		AuthMiddleware(&stubSessionStore)(&stubHandler).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnauthorized)
	})
}
