package items

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

const jsonContentType = "application/json"

type Item struct {
	ID     int
	UserID int
	Name   string
}

type RequestPayload struct {
	Value string `json:"value"`
}
type Items map[string]Item

type ItemsStore interface {
	GetItems() (Items, error)
	InsertItems(value Item) error
	GetItem(key string) (Item, bool, error)
	UpdateItem(key string, value string) (bool, error)
	DeleteItem(key string) (bool, error)
}

type ItemsServer struct {
	store ItemsStore
	http.Handler
}

func (s *ItemsServer) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.GetItems()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(items)
}

func (s *ItemsServer) createItem(w http.ResponseWriter, r *http.Request) {
	err := s.validateContentType(r.Header.Get("Content-Type"))
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	payload, err := s.validatePayload(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	item := Item{
		Name: payload.Value,
		//needs UserID extracted from r.Context().Value() and put here
	}
	insertErr := s.store.InsertItems(item)
	if insertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *ItemsServer) getItem(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/items/")
	value, found, err := s.store.GetItem(key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	item := Items{key: value}
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(item)
}

func (s *ItemsServer) updateItem(w http.ResponseWriter, r *http.Request) {
	err := s.validateContentType(r.Header.Get("Content-Type"))
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/items/")
	payload, err := s.validatePayload(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	updated, err := s.store.UpdateItem(key, payload.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !updated {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *ItemsServer) removeItem(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/items/")
	isDeleted, err := s.store.DeleteItem(key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isDeleted {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *ItemsServer) validateContentType(header string) error {
	if header == "" {
		return errors.New("Content-Type header is required")
	}

	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil || mediaType != jsonContentType {
		return errors.New("Content-Type must be applicaton/json")
	}
	return nil
}

func (s *ItemsServer) validatePayload(body io.ReadCloser) (RequestPayload, error) {
	var payload RequestPayload
	err := json.NewDecoder(body).Decode(&payload)
	if err != nil {
		return RequestPayload{}, err
	}

	if payload.Value == "" {
		return RequestPayload{}, errors.New("json not as per required format: missing value key")
	}
	return payload, nil
}

func NewItemsServer(store ItemsStore) *ItemsServer {
	s := new(ItemsServer)
	s.store = store

	router := http.NewServeMux()
	router.Handle("GET /items", http.HandlerFunc(s.listItems))
	router.Handle("GET /items/{id}", http.HandlerFunc(s.getItem))
	router.Handle("POST /items", http.HandlerFunc(s.createItem))
	router.Handle("PUT /items/{id}", http.HandlerFunc(s.updateItem))
	router.Handle("DELETE /items/{id}", http.HandlerFunc(s.removeItem))

	s.Handler = router

	return s
}
