package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/habx47/gapi/api/root"
	"github.com/habx47/gapi/store"
)

func main() {
	userStore := store.IMUserStore{
		Store:     map[int]store.User{},
		EmailToID: map[string]int{},
	}
	itemStore := store.IMItemsStore{
		Store: store.Items{},
	}
	sessionStore := store.IMSessionStore{
		Store: map[string]store.Session{},
	}

	root := root.NewRootHandler(&userStore, &itemStore, &sessionStore)

	fmt.Println("started listening on port: 8080")
	err := http.ListenAndServe(":8080", root)
	if err != nil {
		log.Fatalf("err starting the server: %v", err)
	}
}
