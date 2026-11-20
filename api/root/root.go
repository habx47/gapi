package root

import (
	"net/http"

	"github.com/habx47/gapi/api/items"
	"github.com/habx47/gapi/api/user"
)

type Root struct {
	http.Handler
}

func NewRootHandler(userStore user.UserStore, itemsStore items.ItemsStore) *Root {
	r := new(Root)

	router := http.NewServeMux()
	router.Handle("/user", http.StripPrefix("/user", user.UserMiddleware(user.NewUserServer(userStore))))
	router.Handle("/user/", http.StripPrefix("/user", user.UserMiddleware(user.NewUserServer(userStore))))

	router.Handle("/items", http.StripPrefix("/items", items.NewItemsServer(itemsStore)))
	router.Handle("/items/", http.StripPrefix("/items", items.NewItemsServer(itemsStore)))

	r.Handler = router

	return r
}
