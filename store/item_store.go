package store

import (
	"errors"
	"sync"
)

var ForbiddenErr = errors.New("forbidden. not authorized")

type Item struct {
	ID     int
	UserID int
	Name   string
}

type Items map[int]Item

type ItemsStore interface {
	GetItems() (Items, error)
	InsertItems(value Item) error
	GetItem(key int) (Item, error)
	UpdateItem(key int, value string, userID int) error
	DeleteItem(key int, userID int) error
}

type IMItemsStore struct {
	store  Items
	nextID int
	mu     sync.Mutex
}

func (imi *IMItemsStore) GetItems() (Items, error) {
	imi.mu.Lock()
	defer imi.mu.Unlock()

	storeInitialization := imi.checkStoreInitilaization()
	if !storeInitialization {
		return Items{}, errors.New("store not initialized")
	}

	return imi.store, nil
}

func (imi *IMItemsStore) InsertItems(value Item) error {
	imi.mu.Lock()
	defer imi.mu.Unlock()

	storeInitialization := imi.checkStoreInitilaization()
	if !storeInitialization {
		return errors.New("store not initialized")
	}

	value.ID = imi.nextID
	imi.store[imi.nextID] = value
	imi.nextID++
	return nil
}

func (imi *IMItemsStore) GetItem(key int) (Item, error) {
	imi.mu.Lock()
	defer imi.mu.Unlock()

	storeInitialization := imi.checkStoreInitilaization()
	if !storeInitialization {
		return Item{}, errors.New("store not initialized")
	}

	item, ok := imi.store[key]
	if !ok {
		return Item{}, ForbiddenErr
	}
	return item, nil
}

func (imi *IMItemsStore) UpdateItem(key int, value string, userID int) error {
	item, err := imi.GetItem(key)
	if err != nil {
		return err
	}

	if userID != item.UserID {
		return ForbiddenErr
	}

	item.Name = value
	imi.mu.Lock()
	imi.store[key] = item
	imi.mu.Unlock()
	return nil
}

func (imi *IMItemsStore) DeleteItem(key int, userID int) error {
	item, err := imi.GetItem(key)
	if err != nil {
		return err
	}

	if userID != item.UserID {
		return ForbiddenErr
	}

	imi.mu.Lock()
	delete(imi.store, key)
	imi.mu.Unlock()
	return nil
}

func (imi *IMItemsStore) checkStoreInitilaization() bool {
	return imi.store != nil
}
