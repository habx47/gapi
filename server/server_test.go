package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type StubStore struct {
}

func (s *StubStore) ListItems() {}

func TestGETStore(t *testing.T) {
	store := StubStore{}
	server := NewServer(&store)
	request := httptest.NewRequest(http.MethodGet, "/items", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusOK)
}

func TestPOSTStore(t *testing.T) {
	store := StubStore{}
	server := NewServer(&store)
	request := httptest.NewRequest(http.MethodPost, "/items", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusAccepted)
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got: %d, want: %d", got, want)
	}
}
