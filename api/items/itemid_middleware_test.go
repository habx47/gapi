package items

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type StubHandler_id struct {
	id int
}

func (sh *StubHandler_id) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, _ := r.Context().Value(itemID).(int)
	sh.id = id
}

func TestItemIDMiddleware(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		stubHandler := StubHandler_id{}
		request := httptest.NewRequest(http.MethodGet, "/invalidid", nil)
		response := httptest.NewRecorder()

		ItemIDMiddleware(&stubHandler).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("valid id", func(t *testing.T) {
		stubHandler := StubHandler_id{}
		request := httptest.NewRequest(http.MethodGet, "/1", nil)
		response := httptest.NewRecorder()

		ItemIDMiddleware(&stubHandler).ServeHTTP(response, request)

		assertResult_id(t, stubHandler, StubHandler_id{id: 1})
	})
}

func assertResult_id(t testing.TB, got, want StubHandler_id) {
	t.Helper()
	if got != want {
		t.Errorf("got: %+v, want: %+v", got, want)
	}
}
