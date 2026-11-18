package user

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type StubHandler struct{}

func (s *StubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestUserMiddleware(t *testing.T) {
	t.Run("invalid content-type", func(t *testing.T) {
		stubHandler := StubHandler{}
		request := httptest.NewRequest(http.MethodPost, "/users", nil)
		response := httptest.NewRecorder()

		request.Header.Set("Content-Type", "text/plain")

		CreateUserMiddleware(&stubHandler).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnsupportedMediaType)

	})
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got: %d, want: %d", got, want)
	}
}
