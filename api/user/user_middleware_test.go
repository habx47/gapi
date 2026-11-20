package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type StubHandler struct {
	Email, Password string
}

func (s *StubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, _ := r.Context().Value(payload_key).(RequestPayload)
	s.Email = payload.Email
	s.Password = payload.Password
}

func TestUserMiddleware(t *testing.T) {
	t.Run("invalid content-type", func(t *testing.T) {
		stubHandler := StubHandler{}
		request := httptest.NewRequest(http.MethodPost, "/users", nil)
		response := httptest.NewRecorder()

		request.Header.Set("Content-Type", "text/plain")

		UserMiddleware(&stubHandler).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnsupportedMediaType)
	})

	t.Run("missing email/password key in payload", func(t *testing.T) {
		stubHandler := StubHandler{}
		payload := RequestPayload{Password: "abc"}
		jsonPayload, _ := json.Marshal(payload)

		request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(jsonPayload))
		response := httptest.NewRecorder()

		request.Header.Set("Content-Type", JSONContentType)

		UserMiddleware(&stubHandler).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("invalid email", func(t *testing.T) {
		stubHandler := StubHandler{}
		payload := RequestPayload{Email: "test.test.com", Password: "abc"}
		jsonPayload, _ := json.Marshal(payload)

		request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(jsonPayload))
		response := httptest.NewRecorder()

		request.Header.Set("Content-Type", JSONContentType)

		UserMiddleware(&stubHandler).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("successful parsing", func(t *testing.T) {
		stubHandler := StubHandler{}
		payload := RequestPayload{Email: " HellO@test.com ", Password: "abc"}
		jsonPayload, _ := json.Marshal(payload)

		request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(jsonPayload))
		response := httptest.NewRecorder()

		request.Header.Set("Content-Type", JSONContentType)

		UserMiddleware(&stubHandler).ServeHTTP(response, request)
		want := RequestPayload{Email: "hello@test.com", Password: "abc"}

		assertPayload(t, RequestPayload(stubHandler), want)
	})
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got: %d, want: %d", got, want)
	}
}

func assertPayload(t testing.TB, got, want RequestPayload) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct payload, got: %+v, want: %+v", got, want)
	}
}
