package root

import (
	"net/http"

	session "github.com/habx47/gapi/api/internal"
	"github.com/habx47/gapi/api/items"
	"github.com/habx47/gapi/api/user"
)

type Root struct {
	http.Handler
}

func NewRootHandler(userStore user.UserStore, itemsStore items.ItemsStore, sessionStore session.SessionStore) *Root {
	r := new(Root)

	router := http.NewServeMux()
	router.Handle("/user", http.StripPrefix("/user", user.UserMiddleware(user.NewUserServer(userStore, sessionStore))))
	router.Handle("/user/", http.StripPrefix("/user", user.UserMiddleware(user.NewUserServer(userStore, sessionStore))))

	router.Handle("/items", http.StripPrefix("/items", items.AuthMiddleware(sessionStore)(items.NewItemsServer(itemsStore))))
	router.Handle("/items/", http.StripPrefix("/items", items.AuthMiddleware(sessionStore)(items.NewItemsServer(itemsStore))))

	r.Handler = router

	return r
}
