package user

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/habx47/gapi/password"
)

const JSONContentType = "application/json"

var EmailExistsErr = errors.New("email already exists")

type User struct {
	ID         int    `json:"id"`
	Email      string `json:"email"`
	PasswdHash []byte `json:"-"`
}

type UserStore interface {
	Create(u User) (User, error)
	GetByID(id int) (bool, error)
	GetByEmail(email string) (bool, error)
}

type UserServer struct {
	store UserStore
	http.Handler
}

func (u *UserServer) createUser(w http.ResponseWriter, r *http.Request) {
	payload, _ := r.Context().Value(payload_key).(RequestPayload)

	hashedPassword, err := password.Hash(payload.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := User{
		Email:      payload.Email,
		PasswdHash: []byte(hashedPassword),
	}

	createdUser, err := u.store.Create(user)
	if err != nil {
		if errors.Is(err, EmailExistsErr) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", JSONContentType)
	w.WriteHeader(http.StatusCreated)
	responseErr := json.NewEncoder(w).Encode(createdUser)
	if responseErr != nil {
		log.Printf("error sending response: %v", responseErr)
	}
}

func (u *UserServer) login(w http.ResponseWriter, r *http.Request) {
}

func NewUserServer(store UserStore) *UserServer {
	u := new(UserServer)
	u.store = store

	router := http.NewServeMux()
	router.Handle("POST /", http.HandlerFunc(u.createUser))
	router.Handle("POST /login", http.HandlerFunc(u.login))

	u.Handler = router

	return u
}
