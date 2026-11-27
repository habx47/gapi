package items

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

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
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	payload, err := s.validatePayload(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	session, sessionExists := getSession(r.Context())
	if !sessionExists {
		w.WriteHeader(http.StatusForbidden)
		return
	}

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
	id, intIDExists := getItemID(r.Context())
	if !intIDExists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value, err := s.store.GetItem(id)
	if err != nil {
		if errors.Is(err, store.ForbiddenErr) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	item := store.Items{id: value}
	w.Header().Set("content-type", JSONContentType)
	json.NewEncoder(w).Encode(item)
}

func (s *ItemsServer) updateItem(w http.ResponseWriter, r *http.Request) {
	err := s.validateContentType(r.Header.Get("Content-Type"))
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	payload, err := s.validatePayload(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	session, sessionExists := getSession(r.Context())
	if !sessionExists {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	id, intIDExists := getItemID(r.Context())
	if !intIDExists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	updateErr := s.store.UpdateItem(id, payload.Value, session.UserID)
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
	session, sessionExists := getSession(r.Context())
	if !sessionExists {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	id, intIDExists := getItemID(r.Context())
	if !intIDExists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := s.store.DeleteItem(id, session.UserID)
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
	err := decodeJson(body, &payload)
	if err != nil {
		return RequestPayload{}, err
	}

	if payload.Value == "" {
		return RequestPayload{}, errors.New("json not as per required format: missing value key")
	}
	return payload, nil
}

func decodeJson(r io.Reader, dst any) error {
	decoder := json.NewDecoder(r)

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain only one json value")
		}
		return err
	}
	return nil
}

func getSession(ctx context.Context) (store.Session, bool) {
	session, ok := ctx.Value(sessionKey).(store.Session)
	return session, ok
}

func getItemID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(itemID).(int)
	return id, ok
}

func NewItemsServer(store store.ItemsStore) *ItemsServer {
	s := new(ItemsServer)
	s.store = store

	router := http.NewServeMux()
	router.Handle("GET /", http.HandlerFunc(s.listItems))
	router.Handle("GET /{id}", ItemIDMiddleware(http.HandlerFunc(s.getItem)))
	router.Handle("POST /", http.HandlerFunc(s.createItem))
	router.Handle("PUT /{id}", ItemIDMiddleware(http.HandlerFunc(s.updateItem)))
	router.Handle("DELETE /{id}", ItemIDMiddleware(http.HandlerFunc(s.removeItem)))

	s.Handler = router

	return s
}
