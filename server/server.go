package server

import (
	"encoding/json"
	"net/http"
)

const jsonContentType = "application/json"

type Items map[string]string

type Store interface {
	GetItems() Items
	InsertItem(payload Items)
}

type Server struct {
	store Store
	http.Handler
}

func (s *Server) storeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listItems(w)
	case http.MethodPost:
		s.createItem(w, r)
	}
}

func (s *Server) listItems(w http.ResponseWriter) {
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(s.store.GetItems())
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
	s.store.InsertItem(payload)
}

func NewServer(store Store) *Server {
	s := new(Server)
	s.store = store

	router := http.NewServeMux()
	router.Handle("/items", http.HandlerFunc(s.storeHandler))

	s.Handler = router

	return s
}
