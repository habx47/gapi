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
var UserNotFoundErr = errors.New("user not found")

type User struct {
	ID         int    `json:"id"`
	Email      string `json:"email"`
	PasswdHash []byte `json:"-"`
}

type UserStore interface {
	Create(u User) (User, error)
	GetByID(id int) (User, error)
	GetByEmail(email string) (User, error)
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
	payload, _ := r.Context().Value(payload_key).(RequestPayload)
	user, err := u.store.GetByEmail(payload.Email)
	if err != nil {
		if errors.Is(err, UserNotFoundErr) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	matched, err := password.Verify(string(user.PasswdHash), payload.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !matched {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
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
