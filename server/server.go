package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const jsonContentType = "application/json"

type Item string
type Items map[string]Item

type Store interface {
	getItems() (Items, error)
	insertItems(value Item) error
	getItem(key string) (Item, error)
	updateItem(key string, value Item) (bool, error)
	deleteItem(key string) (bool, error)
}

type Server struct {
	store Store
	http.Handler
}

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.getItems()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(items)
}

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value, err := s.validatePayload(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	insertErr := s.store.insertItems(value)
	if insertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) getItem(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/items/")
	value, err := s.store.getItem(key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(value) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	item := Items{key: value}
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(item)
}

func (s *Server) updateItem(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/items/")
	value, err := s.validatePayload(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	updated, err := s.store.updateItem(key, value)
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
	isDeleted, err := s.store.deleteItem(key)
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

func (s *Server) validatePayload(w http.ResponseWriter, r *http.Request) (Item, error) {
	var payload map[string]Item
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return "", err
	}

	value, ok := payload["value"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
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
