package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const TestEmail = "test@test.com"
const TestPassword = "abc"

var payload RequestPayload = RequestPayload{
	Email:    TestEmail,
	Password: TestPassword,
}

type StubUserStore struct {
	err    bool
	exists bool
	UserStore
}

func (s *StubUserStore) Create(u User) (User, error) {
	if s.err {
		return User{}, errors.New("internal server error")
	}
	if s.exists {
		return User{}, EmailExistsErr
	}
	u.ID = 1
	return u, nil
}

type ResponseUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func TestCreateUser(t *testing.T) {
	t.Run("success on creating user", func(t *testing.T) {
		stubUserStore := StubUserStore{}
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		response := httptest.NewRecorder()
		server := NewUserServer(&stubUserStore)

		ctx := createContext(request, payload)

		server.ServeHTTP(response, request.WithContext(ctx))

		var responseUser ResponseUser
		err := json.NewDecoder(response.Body).Decode(&responseUser)
		if err != nil {
			t.Fatalf("could not read response from the response: %v", err)
		}

		assertStatus(t, response.Code, http.StatusCreated)
		assertContentType(t, response, JSONContentType)
		assertResponseUser(t, responseUser, ResponseUser{ID: 1, Email: TestEmail})
	})

	t.Run("email already exists", func(t *testing.T) {
		stubUserStore := StubUserStore{
			exists: true,
		}
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		response := httptest.NewRecorder()
		server := NewUserServer(&stubUserStore)

		ctx := createContext(request, payload)

		server.ServeHTTP(response, request.WithContext(ctx))

		assertStatus(t, response.Code, http.StatusConflict)
	})

	t.Run("internal server error on user creation", func(t *testing.T) {
		stubUserStore := StubUserStore{
			err: true,
		}
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		response := httptest.NewRecorder()
		server := NewUserServer(&stubUserStore)

		ctx := createContext(request, payload)

		server.ServeHTTP(response, request.WithContext(ctx))

		assertStatus(t, response.Code, http.StatusInternalServerError)
	})
}

func createContext(req *http.Request, value RequestPayload) context.Context {
	return context.WithValue(req.Context(), payload_key, value)
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
