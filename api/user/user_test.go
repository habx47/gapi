package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/habx47/gapi/password"
	"github.com/habx47/gapi/store"
)

const TestEmail = "test@test.com"
const TestPassword = "abcabcabc"

var payload RequestPayload = RequestPayload{
	Email:    TestEmail,
	Password: TestPassword,
}

type StubUserStore struct {
	err          bool
	exists       bool
	invalidEmail bool
	store.UserStore
}

func (su *StubUserStore) Create(u store.User) (store.User, error) {
	if su.err {
		return store.User{}, errors.New("internal server error")
	}
	if su.exists {
		return store.User{}, store.EmailExistsErr
	}
	u.ID = 1
	return u, nil
}

func (su *StubUserStore) GetByEmail(email string) (store.User, error) {
	if su.invalidEmail {
		return store.User{}, store.UserNotFoundErr
	}

	passwordHash, _ := password.Hash(TestPassword)
	return store.User{
		ID:         1,
		Email:      TestEmail,
		PasswdHash: []byte(passwordHash),
	}, nil
}

type StubSessionStore struct {
	store.SessionStore
}

func (ss *StubSessionStore) CreateSession(userID int, ttl time.Duration) (store.Session, error) {
	return store.Session{
		ID:        "generatedsessionid",
		UserID:    1,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(2 * time.Minute),
	}, nil
}

type ResponseUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func TestCreateUser(t *testing.T) {
	t.Run("success on creating user", func(t *testing.T) {
		stubUserStore := StubUserStore{}
		stubSessionStore := StubSessionStore{}
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		response := httptest.NewRecorder()
		server := NewUserServer(&stubUserStore, &stubSessionStore)

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
		stubSessionStore := StubSessionStore{}
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		response := httptest.NewRecorder()
		server := NewUserServer(&stubUserStore, &stubSessionStore)

		ctx := createContext(request, payload)

		server.ServeHTTP(response, request.WithContext(ctx))

		assertStatus(t, response.Code, http.StatusConflict)
	})

	t.Run("internal server error on user creation", func(t *testing.T) {
		stubUserStore := StubUserStore{
			err: true,
		}
		stubSessionStore := StubSessionStore{}
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		response := httptest.NewRecorder()
		server := NewUserServer(&stubUserStore, &stubSessionStore)

		ctx := createContext(request, payload)

		server.ServeHTTP(response, request.WithContext(ctx))

		assertStatus(t, response.Code, http.StatusInternalServerError)
	})
}

func TestLogin(t *testing.T) {
	t.Run("wrong login credentials - email", func(t *testing.T) {
		stubUserStore := StubUserStore{invalidEmail: true}
		stubSessionStore := StubSessionStore{}
		loginPayload := RequestPayload{
			Email:    "test1@acme.com",
			Password: "doesnotmatterhere",
		}
		request := httptest.NewRequest(http.MethodPost, "/login", nil)
		response := httptest.NewRecorder()
		server := NewUserServer(&stubUserStore, &stubSessionStore)

		ctx := createContext(request, loginPayload)

		server.ServeHTTP(response, request.WithContext(ctx))

		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("wrong login credentials - password", func(t *testing.T) {
		stubUserStore := StubUserStore{}
		stubSessionStore := StubSessionStore{}
		loginPayload := RequestPayload{
			Email:    TestEmail,
			Password: "wrongpassword",
		}
		request := httptest.NewRequest(http.MethodPost, "/login", nil)
		response := httptest.NewRecorder()
		server := NewUserServer(&stubUserStore, &stubSessionStore)

		ctx := createContext(request, loginPayload)

		server.ServeHTTP(response, request.WithContext(ctx))

		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("success on login and session creation", func(t *testing.T) {
		stubUserStore := StubUserStore{}
		stubSessionStore := StubSessionStore{}
		loginPayload := RequestPayload{
			Email:    TestEmail,
			Password: TestPassword,
		}
		request := httptest.NewRequest(http.MethodPost, "/login", nil)
		response := httptest.NewRecorder()
		server := NewUserServer(&stubUserStore, &stubSessionStore)

		ctx := createContext(request, loginPayload)

		server.ServeHTTP(response, request.WithContext(ctx))

		var sessionCookie *http.Cookie
		for _, c := range response.Result().Cookies() {
			if c.Name == "session_id" {
				sessionCookie = c
				break
			}
		}

		assertStatus(t, response.Code, http.StatusNoContent)
		assertSessionCookie(t, sessionCookie)
		assertSessionCookieValue(t, sessionCookie.Value, "generatedsessionid")
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

func assertSessionCookie(t testing.TB, sessionCookie *http.Cookie) {
	t.Helper()
	if sessionCookie == nil {
		t.Fatal("expected session_id cookie to be set, but none was found")
	}
}

func assertSessionCookieValue(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}
