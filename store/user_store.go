package store

import (
	"errors"
	"sync"
)

var EmailExistsErr = errors.New("email already exists")
var UserNotFoundErr = errors.New("user not found")
var UninitializedStoreErr = errors.New("required store[s] not initialized")
var StoreSyncErr = errors.New("store and emailToID are out of sync")

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

	storeInitialized := imu.checkStoreInitilaization()
	if !storeInitialized {
		return User{}, UninitializedStoreErr
	}

	_, exists := imu.checkEmailExists(u.Email)
	if exists {
		return User{}, EmailExistsErr
	}

	u.ID = imu.nextID
	imu.store[imu.nextID] = u
	imu.emailToID[u.Email] = imu.nextID
	imu.nextID++

	return u, nil
}

func (imu *IMUserStore) getByID(id int) (User, bool) {
	user, ok := imu.store[id]
	return user, ok
}

func (imu *IMUserStore) GetByEmail(email string) (User, error) {
	imu.mu.Lock()
	defer imu.mu.Unlock()

	storeInitialized := imu.checkStoreInitilaization()
	if !storeInitialized {
		return User{}, UninitializedStoreErr
	}

	id, exists := imu.checkEmailExists(email)
	if !exists {
		return User{}, UserNotFoundErr
	}

	user, exists := imu.getByID(id)
	//if user does not exist here, it means store and emailToID are out of sync
	if !exists {
		return User{}, StoreSyncErr
	}

	return user, nil
}

func (imu *IMUserStore) checkEmailExists(email string) (int, bool) {
	id, ok := imu.emailToID[email]
	return id, ok
}

func (imu *IMUserStore) checkStoreInitilaization() bool {
	return imu.store != nil && imu.emailToID != nil
}

func NewIMUserStore() *IMUserStore {
	s := new(IMUserStore)
	s.store = map[int]User{}
	s.emailToID = map[string]int{}
	s.nextID = 1

	return s
}
