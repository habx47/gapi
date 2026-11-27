package root

import (
	"net/http"

	"github.com/habx47/gapi/api/items"
	"github.com/habx47/gapi/api/user"
	"github.com/habx47/gapi/store"
)

type Root struct {
	http.Handler
}

func NewRootHandler(userStore store.UserStore, itemsStore store.ItemsStore, sessionStore store.SessionStore) *Root {
	r := new(Root)
	userServer := user.NewUserServer(userStore, sessionStore)
	itemsServer := items.NewItemsServer(itemsStore)

	router := http.NewServeMux()
	router.Handle("/user", http.StripPrefix("/user", user.UserMiddleware(userServer)))
	router.Handle("/user/", http.StripPrefix("/user", user.UserMiddleware(userServer)))

	router.Handle("/items", http.StripPrefix("/items", items.AuthMiddleware(sessionStore)(itemsServer)))
	router.Handle("/items/", http.StripPrefix("/items", items.AuthMiddleware(sessionStore)(itemsServer)))

	r.Handler = router

	return r
}
