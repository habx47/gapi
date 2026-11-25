package store

import (
	"errors"
	"reflect"
	"testing"
)

func TestCreate(t *testing.T) {
	t.Run("success on creating a user", func(t *testing.T) {
		stubStore := IMUserStore{
			store:     map[int]User{},
			emailToID: map[string]int{},
			nextID:    1,
		}
		stubUser := User{
			Email:      "test@test.com",
			PasswdHash: []byte("testpassword"),
		}

		user, err := stubStore.Create(stubUser)
		if err != nil {
			t.Fatalf("expected no error, but got: %v", err)
		}

		assertUser(t, user, User{Email: "test@test.com", PasswdHash: []byte("testpassword"), ID: 1})
	})

	t.Run("no stores initialzed", func(t *testing.T) {
		stubStore := IMUserStore{}
		stubUser := User{
			Email:      "test@test.com",
			PasswdHash: []byte("testpassword"),
		}

		_, err := stubStore.Create(stubUser)
		if err == nil {
			t.Fatalf("expected an error: %v, but got nil", UninitializedStoreErr)
		}

		assertErr(t, err, UninitializedStoreErr)
	})

	t.Run("email already exists", func(t *testing.T) {
		stubStore := IMUserStore{
			store:     map[int]User{},
			emailToID: map[string]int{"test@test.com": 23},
			nextID:    23,
		}
		stubUser := User{
			Email:      "test@test.com",
			PasswdHash: []byte("testpassword"),
		}

		_, err := stubStore.Create(stubUser)
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}

		assertErr(t, err, EmailExistsErr)
	})
}

func TestLogin(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		stubUser := User{
			Email:      "test@test.com",
			PasswdHash: []byte("testpassword"),
			ID:         1,
		}
		stubStore := IMUserStore{
			store:     map[int]User{1: stubUser},
			emailToID: map[string]int{"test@test.com": 1},
			nextID:    2,
		}

		user, err := stubStore.GetByEmail("test@test.com")
		if err != nil {
			t.Fatalf("expected no error, but got: %v", err)
		}

		assertUser(t, user, stubUser)
	})

	t.Run("user not found", func(t *testing.T) {
		stubUser := User{
			Email:      "test@test.com",
			PasswdHash: []byte("testpassword"),
			ID:         1,
		}
		stubStore := IMUserStore{
			store:     map[int]User{1: stubUser},
			emailToID: map[string]int{"test@test.com": 1},
			nextID:    2,
		}

		_, err := stubStore.GetByEmail("test_two@test.com")
		if err == nil {
			t.Fatalf("expected an error: %v, but got nil", UserNotFoundErr)
		}

		assertErr(t, err, UserNotFoundErr)
	})

	t.Run("store sync error", func(t *testing.T) {
		stubUser := User{
			Email:      "test@test.com",
			PasswdHash: []byte("testpassword"),
			ID:         1,
		}
		stubStore := IMUserStore{
			store:     map[int]User{1: stubUser},
			emailToID: map[string]int{"test@test.com": 1, "test2@test.com": 2},
			nextID:    3,
		}

		_, err := stubStore.GetByEmail("test2@test.com")
		if err == nil {
			t.Fatalf("expected an error: %v, but got nil", StoreSyncErr)
		}

		assertErr(t, err, StoreSyncErr)
	})
}

func assertUser(t testing.TB, got, want User) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %+v, want: %+v as created user", got, want)
	}
}

func assertErr(t testing.TB, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("got: %v, want: %v as err value", got, want)
	}
}
