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

	router := http.NewServeMux()
	router.Handle("/user", http.StripPrefix("/user", user.UserMiddleware(user.NewUserServer(userStore, sessionStore))))
	router.Handle("/user/", http.StripPrefix("/user", user.UserMiddleware(user.NewUserServer(userStore, sessionStore))))

	router.Handle("/items", http.StripPrefix("/items", items.AuthMiddleware(sessionStore)(items.NewItemsServer(itemsStore))))
	router.Handle("/items/", http.StripPrefix("/items", items.AuthMiddleware(sessionStore)(items.NewItemsServer(itemsStore))))

	r.Handler = router

	return r
}
