package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type StubStore struct {
	items Items
}

func (s *StubStore) GetItems() Items {
	return s.items
}

func TestGETStore(t *testing.T) {
	items := Items{"1": "item_one"}
	store := StubStore{
		items: items,
	}
	server := NewServer(&store)
	request := httptest.NewRequest(http.MethodGet, "/items", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	got := getItemsFromResponse(t, response.Body)

	assertStatus(t, response.Code, http.StatusOK)
	assertItems(t, got, items)
	assetContentType(t, response, jsonContentType)
}

func TestPOSTStore(t *testing.T) {
	store := StubStore{}
	server := NewServer(&store)
	request := httptest.NewRequest(http.MethodPost, "/items", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusAccepted)
}

func getItemsFromResponse(t testing.TB, body io.Reader) (items Items) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&items)
	if err != nil {
		t.Fatalf("unable to parse response from server %q into map of items, '%v'", body, err)
	}

	return
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got: %d, want: %d", got, want)
	}
}

func assetContentType(t testing.TB, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	if response.Result().Header.Get("content-type") != want {
		t.Errorf("response did not have content-type of %s, got: %v", want, response.Result().Header)
	}
}

func assertItems(t testing.TB, got, want Items) {}
