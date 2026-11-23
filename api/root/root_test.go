package root

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	session "github.com/habx47/gapi/api/internal"
	"github.com/habx47/gapi/api/items"
	"github.com/habx47/gapi/api/user"
)

type StubUserStore struct {
	user.UserStore
}

func (s *StubUserStore) Create(u user.User) (user.User, error) {
	u.ID = 1
	return u, nil
}

func (s *StubUserStore) CheckByEmail(email string) (bool, error) {
	return false, nil
}

type StubItemsStore struct {
	item items.Item
	items.ItemsStore
}

func (s *StubItemsStore) InsertItems(item items.Item) error {
	item.ID = 1
	s.item = item
	return nil
}

type StubSessionStore struct {
	session.SessionStore
}

func (s *StubSessionStore) GetSession(sessionID string) (session.Session, error) {
	return session.Session{
		ID:        "stub_session_id",
		UserID:    1,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}, nil
}

type ResponseUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func TestRootHandler(t *testing.T) {
	t.Run("success on POST /user", func(t *testing.T) {
		stubUserStore := StubUserStore{}
		stubItemsStore := StubItemsStore{}
		stubSessionStore := StubSessionStore{}

		payload := user.RequestPayload{
			Email:    " Test@test.com ",
			Password: "abcabc",
		}
		jsonBytes, _ := json.Marshal(payload)

		request := httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(jsonBytes))
		response := httptest.NewRecorder()
		request.Header.Set("Content-Type", user.JSONContentType)

		server := NewRootHandler(&stubUserStore, &stubItemsStore, &stubSessionStore)
		server.ServeHTTP(response, request)

		var responseUser ResponseUser
		err := json.NewDecoder(response.Body).Decode(&responseUser)
		if err != nil {
			t.Fatalf("could not read response from the response: %v", err)
		}

		assertStatus(t, response.Code, http.StatusCreated)
		assertContentType(t, response, user.JSONContentType)
		assertResponseUser(t, responseUser, ResponseUser{ID: 1, Email: "test@test.com"})
	})

	t.Run("success on POST /items", func(t *testing.T) {
		stubUserStore := StubUserStore{}
		stubItemsStore := StubItemsStore{}
		StubSessionStore := StubSessionStore{}

		payload := items.RequestPayload{
			Value: "item_one",
		}
		jsonBytes, _ := json.Marshal(payload)

		request := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(jsonBytes))
		request.Header.Set("Content-Type", user.JSONContentType)
		request.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: "stub_session_id",
		})
		response := httptest.NewRecorder()

		server := NewRootHandler(&stubUserStore, &stubItemsStore, &StubSessionStore)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusCreated)
		assertItem(t, stubItemsStore.item, items.Item{ID: 1, UserID: 1, Name: "item_one"})
	})
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got: %d, want: %d", got, want)
	}
}

func assertContentType(t testing.TB, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	if response.Result().Header.Get("Content-Type") != want {
		t.Errorf("response did not have content-type of %s, got: %v", want, response.Result().Header)
	}
}

func assertResponseUser(t testing.TB, got, want ResponseUser) {
	t.Helper()
	if got != want {
		t.Errorf("got: %+v as response, want: %+v", got, want)
	}
}

func assertItem(t testing.TB, got, want items.Item) {
	t.Helper()
	if got != want {
		t.Errorf("got: %+v, want: %+v", got, want)
	}
}
