package root

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	session "github.com/habx47/gapi/api/internal"
	"github.com/habx47/gapi/api/items"
	"github.com/habx47/gapi/api/user"
)

type StubUserStore struct {
	user.UserStore
}

type StubItemsStore struct {
	items.ItemsStore
}

type StubSessionStore struct {
	session.SessionStore
}

func (s *StubUserStore) Create(u user.User) (user.User, error) {
	u.ID = 1
	return u, nil
}

func (s *StubUserStore) CheckByEmail(email string) (bool, error) {
	return false, nil
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
