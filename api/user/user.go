package user

import (
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

func (u *UserServer) createUser(w http.ResponseWriter, r *http.Request) {
	payload, _ := r.Context().Value(payload_key).(RequestPayload)

	exists, err := u.store.CheckByEmail(payload.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists {
		w.WriteHeader(http.StatusConflict)
		return
	}
	//hash password
	//insert in the store
	//respose
}

func NewUserServer(store UserStore) *UserServer {
	u := new(UserServer)
	u.store = store

	router := http.NewServeMux()
	router.Handle("POST /users", http.HandlerFunc(u.createUser))

	u.Handler = router

	return u
}
