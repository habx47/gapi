package user

import "net/http"

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
}

func NewUserServer(store UserStore) *UserServer {
	u := new(UserServer)
	u.store = store

	router := http.NewServeMux()
	router.Handle("POST /user", http.HandlerFunc(u.createUser))

	u.Handler = router

	return u
}
