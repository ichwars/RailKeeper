package authdata_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/infrastructure/authdata"
)

type heldLoginCommit struct {
	*authdata.Store
	entered chan struct{}
	release chan struct{}
}

func (store *heldLoginCommit) CommitLogin(ctx context.Context, proof authentication.Commit) error {
	close(store.entered)
	<-store.release
	return store.Store.CommitLogin(ctx, proof)
}

func TestInFlightLoginCannotRestoreARevokedBrowserProof(t *testing.T) {
	for _, action := range []string{"admin-reset", "role-change"} {
		t.Run(action, func(t *testing.T) {
			db, auth, id := originDatabase(t)
			initial, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password"})
			if err != nil {
				t.Fatal(err)
			}
			store := &heldLoginCommit{Store: authdata.NewStore(db), entered: make(chan struct{}), release: make(chan struct{})}
			var once sync.Once
			release := func() { once.Do(func() { close(store.release) }) }
			defer release()
			results := make(chan error, 1)
			go func() {
				result, err := authentication.NewService(store).Login(t.Context(), authentication.Input{Username: "admin", Password: "very-secure-password", OriginToken: initial.OriginToken})
				if result != nil {
					results <- errors.New("revoked proof yielded a session")
					return
				}
				results <- err
			}()
			select {
			case <-store.entered:
			case err := <-results:
				t.Fatal("login did not reach commit", err)
			case <-time.After(5 * time.Second):
				t.Fatal("login never reached controlled commit")
			}
			if action == "admin-reset" {
				err = auth.ResetLoginBudget(t.Context(), id, id)
			} else {
				_, err = auth.UpdateUser(t.Context(), id, id, application.UpdateUserInput{Username: "admin", Email: "admin@example.test", Roles: []string{"Admin"}})
			}
			if err != nil {
				t.Fatal(err)
			}
			release()
			if err := <-results; !errors.Is(err, authentication.ErrInvalidLogin) {
				t.Fatal("revoked recognition survived commit", err)
			}
			if _, _, found, err := store.FindOrigin(t.Context(), "admin", credentials.HashToken(initial.OriginToken)); err != nil || found {
				t.Fatal("revoked token was recreated", err)
			}
			var sessions int
			if err := db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&sessions); err != nil || sessions != 1 {
				t.Fatal("in-flight login created a session", sessions, err)
			}
			fresh, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password"})
			if err != nil || fresh.OriginToken == initial.OriginToken {
				t.Fatal("fresh complete login did not replace the revoked proof", err)
			}
		})
	}
}
