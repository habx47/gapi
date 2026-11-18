package user

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"strings"
)

type contextKey string

type RequestPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const payload_key contextKey = "payload"

func CreateUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := validateContentType(r.Header.Get("Content-Type"))
		if err != nil {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		payload, err := validatePayload(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		normalizeErr := normalizeEmail(&payload)
		if normalizeErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), payload_key, payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateContentType(header string) error {
	if header == "" {
		return errors.New("Content-Type header is requried")
	}

	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil || mediaType != jsonContentType {
		return errors.New("Content-Type must be application/json")
	}
	return nil
}

func validatePayload(r io.Reader) (RequestPayload, error) {
	var payload RequestPayload
	err := json.NewDecoder(r).Decode(&payload)
	if err != nil {
		return RequestPayload{}, err
	}

	if payload.Email == "" || payload.Password == "" {
		return RequestPayload{}, errors.New("json not as per required format: missing email or password key")
	}

	return payload, nil
}

func normalizeEmail(payload *RequestPayload) error {
	email := payload.Email
	cleaned := strings.ToLower(strings.TrimSpace(email))
	if cleaned == "" {
		return errors.New("empty email value in request body")
	}

	addr, err := mail.ParseAddress(cleaned)
	if err != nil {
		return errors.New("invalid email format")
	}

	payload.Email = addr.Address
	return nil
}
