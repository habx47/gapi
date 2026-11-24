package store

import (
	"errors"
	"sync"
)

var EmailExistsErr = errors.New("email already exists")
var UserNotFoundErr = errors.New("user not found")

type User struct {
	ID         int    `json:"id"`
	Email      string `json:"email"`
	PasswdHash []byte `json:"-"`
}

type UserStore interface {
	Create(u User) (User, error)
	GetByEmail(email string) (User, error)
}

type IMUserStore struct {
	store     map[int]User
	emailToID map[string]int
	mu        sync.Mutex
	nextID    int
}

func (imu *IMUserStore) Create(u User) (User, error) {
	imu.mu.Lock()
	defer imu.mu.Unlock()

	_, err := imu.checkEmailExists(u.Email)
	if err != nil {
		return User{}, err
	}

	u.ID = imu.nextID
	imu.store[imu.nextID] = u
	imu.emailToID[u.Email] = imu.nextID
	imu.nextID++

	return u, nil
}

func (imu *IMUserStore) getByID(id int) (User, error) {
	imu.mu.Lock()
	defer imu.mu.Unlock()

	user, ok := imu.store[id]
	if !ok {
		return User{}, UserNotFoundErr
	}

	return user, nil
}

func (imu *IMUserStore) GetByEmail(email string) (User, error) {
	imu.mu.Lock()
	defer imu.mu.Unlock()

	id, err := imu.checkEmailExists(email)
	if err == nil {
		return User{}, UserNotFoundErr
	}

	user, err := imu.getByID(id)
	if err != nil {
		return User{}, UserNotFoundErr
	}

	return user, nil
}

func (imu *IMUserStore) checkEmailExists(email string) (int, error) {
	id, ok := imu.emailToID[email]
	if !ok {
		return id, nil
	}
	return id, EmailExistsErr
}
