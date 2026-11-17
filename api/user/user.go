package user

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

const jsonContentType = "application/json"

type User struct {
	ID         int
	Email      string
	PasswdHash []byte
}

type UserStore interface {
	Create(u User) (User, error)
	CheckByID(id int) (bool, error)
	CheckByEmail(email string) (bool, error)
	Delete(id int) (bool, error)
}

type UserServer struct {
	store UserStore
	http.Handler
}

type RequestPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *UserServer) createUser(w http.ResponseWriter, r *http.Request) {
	err := u.validateContentType(r.Header.Get("Content-Type"))
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	payload, err := validatePayload(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//validate email
	//hash password
	//insert in the store
	//respose
}

func (u *UserServer) validateContentType(header string) error {
	if header == "" {
		return errors.New("Content-Type header is requried")
	}

	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil || mediaType != jsonContentType {
		return errors.New("Content-Type must be application/json")
	}
	return nil
}

func (u *UserServer) validatePayload(r io.Reader) (RequestPayload, error) {
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

func NewUserServer(store UserStore) *UserServer {
	u := new(UserServer)
	u.store = store

	router := http.NewServeMux()
	router.Handle("POST /users", http.HandlerFunc(u.createUser))

	u.Handler = router

	return u
}
