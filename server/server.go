package server

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

const jsonContentType = "application/json"

type Item string
type Items map[string]Item

type Store interface {
	GetItems() (Items, error)
	InsertItems(value Item) error
	GetItem(key string) (Item, bool, error)
	UpdateItem(key string, value Item) (bool, error)
	DeleteItem(key string) (bool, error)
}

type Server struct {
	store Store
	http.Handler
}

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.GetItems()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(items)
}

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	err := s.validateContentType(r.Header.Get("Content-Type"))
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	value, err := s.validatePayload(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	insertErr := s.store.InsertItems(value)
	if insertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) getItem(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) updateItem(w http.ResponseWriter, r *http.Request) {
	err := s.validateContentType(r.Header.Get("Content-Type"))
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/items/")
	value, err := s.validatePayload(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	updated, err := s.store.UpdateItem(key, value)
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

func (s *Server) removeItem(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) validateContentType(header string) error {
	if header == "" {
		return errors.New("Content-Type header is required")
	}

	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil || mediaType != jsonContentType {
		return errors.New("Content-Type must be applicaton/json")
	}
	return nil
}

func (s *Server) validatePayload(body io.ReadCloser) (Item, error) {
	var payload map[string]Item
	err := json.NewDecoder(body).Decode(&payload)
	if err != nil {
		return "", err
	}

	value, ok := payload["value"]
	if !ok {
		return "", errors.New("json not as per required format: missing value key")
	}
	return value, nil
}

func NewServer(store Store) *Server {
	s := new(Server)
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
