package store

import "errors"

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
