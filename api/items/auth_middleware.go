package items

import (
	"context"
	"fmt"
	"net/http"

	"github.com/habx47/gapi/store"
)

type sessionContextKey string

const sessionKey sessionContextKey = "session"

func AuthMiddleware(sessionStore store.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			cookie, err := r.Cookie("session_id")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			session, err := sessionStore.GetSession(cookie.Value)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if session.IsExpired() {
				err := sessionStore.DeleteSession(cookie.Value)
				if err != nil {
					fmt.Printf("error deleting expired session with id: %v", session.ID)
				}
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), sessionKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
