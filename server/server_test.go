package server

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type StubStore struct {
	items Items
}

func (s *StubStore) getItems() Items {
	return s.items
}

func (s *StubStore) insertItems(payload Items) {
	s.items = maps.Clone(payload)
}

func (s *StubStore) getItem(id string) Item {
	value := s.items[id]
	return value
}

func (s *StubStore) updateItem(key string, value Item) bool {
	s.items[key] = value
	return true
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
	assertStoreState(t, got, items)
	assetContentType(t, response, jsonContentType)
}

func TestGETItem(t *testing.T) {
	items := Items{"1": "item_one", "2": "item_two"}
	store := StubStore{
		items: items,
	}
	server := NewServer(&store)
	request := httptest.NewRequest(http.MethodGet, "/items/1", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	got := getItemsFromResponse(t, response.Body)
	want := Items{"1": "item_one"}

	assertStatus(t, response.Code, http.StatusOK)
	assertStoreState(t, got, want)
	assetContentType(t, response, jsonContentType)
}

func TestPOSTStore(t *testing.T) {
	t.Run("content-type: application/json", func(t *testing.T) {
		payload := Items{"1": "item_one"}
		jsonBytes, _ := json.Marshal(payload)

		store := StubStore{}
		server := NewServer(&store)
		request := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(jsonBytes))
		request.Header.Set("content-type", jsonContentType)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusAccepted)
		assertStoreState(t, store.items, payload)
	})

	t.Run("content-type: text/plain", func(t *testing.T) {
		store := StubStore{}
		server := NewServer(&store)
		request := httptest.NewRequest(http.MethodPost, "/items", nil)
		request.Header.Set("content-type", "text/plain")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})
}

func TestPUTItem(t *testing.T) {
	payload := Items{"value": "item_one_updated"}
	jsonBytes, _ := json.Marshal(payload)

	store := StubStore{items: Items{"1": "item_one"}}
	updatedStore := StubStore{items: Items{"1": "item_one_updated"}}
	server := NewServer(&store)
	request := httptest.NewRequest(http.MethodPut, "/items/1", bytes.NewReader(jsonBytes))
	request.Header.Set("content-type", jsonContentType)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusNoContent)
	assertStoreState(t, store.items, updatedStore.items)
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

func assertStoreState(t testing.TB, got, want Items) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}
