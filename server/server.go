package server

import (
	"encoding/json"
	"net/http"
)

const jsonContentType = "application/json"

type Items map[string]string

type Store interface {
	ListItems() Items
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
		s.createItem(w)
	}
}

func (s *Server) listItems(w http.ResponseWriter) {
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(s.store.ListItems())
}

func (s *Server) createItem(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
}

func NewServer(store Store) *Server {
	s := new(Server)
	s.store = store

	router := http.NewServeMux()
	router.Handle("/items", http.HandlerFunc(s.storeHandler))

	s.Handler = router

	return s
}
