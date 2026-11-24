package store

import "errors"

var ForbiddenErr = errors.New("forbidden. not authorized")

type Item struct {
	ID     int
	UserID int
	Name   string
}

type RequestPayload struct {
	Value string `json:"value"`
}
type Items map[string]Item

type ItemsStore interface {
	GetItems() (Items, error)
	InsertItems(value Item) error
	GetItem(key string) (Item, error)
	UpdateItem(key string, value string, userID int) error
	DeleteItem(key string, userID int) error
}
