package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

const jsonContentType = "application/json"

type Item string
type Items map[string]Item

type Store interface {
	getItems() Items
	insertItems(payload Items)
	getItem(key string) Item
	updateItem(key string, value Item) bool
}

type Server struct {
	store Store
	http.Handler
}

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(s.store.getItems())
}

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var payload Items
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	s.store.insertItems(payload)
}

func (s *Server) getItem(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/items/")
	value := s.store.getItem(key)

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
	var payload map[string]Item
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, ok := payload["value"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_ = s.store.updateItem(key, value)
	w.WriteHeader(http.StatusNoContent)
}

func NewServer(store Store) *Server {
	s := new(Server)
	s.store = store

	router := http.NewServeMux()
	router.Handle("GET /items", http.HandlerFunc(s.listItems))
	router.Handle("GET /items/{id}", http.HandlerFunc(s.getItem))
	router.Handle("POST /items", http.HandlerFunc(s.createItem))
	router.Handle("PUT /items/{id}", http.HandlerFunc(s.updateItem))

	s.Handler = router

	return s
}
