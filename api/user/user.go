package user

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	session "github.com/habx47/gapi/api/internal"
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
	store        UserStore
	sessionStore session.SessionStore
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

	ttl := 24 * time.Hour
	session, err := u.sessionStore.CreateSession(user.ID, ttl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func NewUserServer(store UserStore, sessionStore session.SessionStore) *UserServer {
	u := new(UserServer)
	u.store = store
	u.sessionStore = sessionStore

	router := http.NewServeMux()
	router.Handle("POST /", http.HandlerFunc(u.createUser))
	router.Handle("POST /login", http.HandlerFunc(u.login))

	u.Handler = router

	return u
}
