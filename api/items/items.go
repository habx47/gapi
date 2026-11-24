package items

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/habx47/gapi/store"
)

const JSONContentType = "application/json"

type RequestPayload struct {
	Value string `json:"value"`
}

type ItemsServer struct {
	store store.ItemsStore
	http.Handler
}

func (s *ItemsServer) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.GetItems()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", JSONContentType)
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

	session := getSession(r.Context())

	item := store.Item{
		Name:   payload.Value,
		UserID: session.UserID,
	}
	insertErr := s.store.InsertItems(item)
	if insertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *ItemsServer) getItem(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/")
	value, err := s.store.GetItem(key)
	if err != nil {
		if errors.Is(err, store.ForbiddenErr) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	item := store.Items{key: value}
	w.Header().Set("content-type", JSONContentType)
	json.NewEncoder(w).Encode(item)
}

func (s *ItemsServer) updateItem(w http.ResponseWriter, r *http.Request) {
	err := s.validateContentType(r.Header.Get("Content-Type"))
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/")
	payload, err := s.validatePayload(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	session := getSession(r.Context())
	updateErr := s.store.UpdateItem(key, payload.Value, session.UserID)
	if updateErr != nil {
		if errors.Is(updateErr, store.ForbiddenErr) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *ItemsServer) removeItem(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/")
	session := getSession(r.Context())
	err := s.store.DeleteItem(key, session.UserID)
	if err != nil {
		if errors.Is(err, store.ForbiddenErr) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *ItemsServer) validateContentType(header string) error {
	if header == "" {
		return errors.New("Content-Type header is required")
	}

	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil || mediaType != JSONContentType {
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

func getSession(ctx context.Context) store.Session {
	session, _ := ctx.Value(sessionKey).(store.Session)
	return session
}

func NewItemsServer(store store.ItemsStore) *ItemsServer {
	s := new(ItemsServer)
	s.store = store

	router := http.NewServeMux()
	router.Handle("GET /", http.HandlerFunc(s.listItems))
	router.Handle("GET /{id}", http.HandlerFunc(s.getItem))
	router.Handle("POST /", http.HandlerFunc(s.createItem))
	router.Handle("PUT /{id}", http.HandlerFunc(s.updateItem))
	router.Handle("DELETE /{id}", http.HandlerFunc(s.removeItem))

	s.Handler = router

	return s
}
