package server

import "net/http"

type Store interface {
	ListItems()
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
	w.WriteHeader(http.StatusOK)
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
