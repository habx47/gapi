package items

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

type itemIDContextKey string

var itemID itemIDContextKey = "itemID"

func ItemIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/")
		id, err := strconv.Atoi(key)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), itemID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
