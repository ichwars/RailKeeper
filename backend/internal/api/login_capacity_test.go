package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/infrastructure/authdata"
)

type heldLoginStore struct {
	authentication.Store
	entered chan struct{}
	release <-chan struct{}
}

func (store *heldLoginStore) FindAccount(ctx context.Context, name string) (authentication.Account, bool, error) {
	close(store.entered)
	<-store.release
	return store.Store.FindAccount(ctx, name)
}

type exhaustedLoginStore struct{ authentication.Store }

func (exhaustedLoginStore) AllowAttempt(context.Context, authentication.Attempt) (bool, error) {
	return false, nil
}

func loginCapacityRequest() *http.Request {
	return httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"synthetic-capacity-test","password":"synthetic-wrong-password"}`))
}

func TestBusyLoginSuggestsShortRetryWithoutChangingAccountWindow(t *testing.T) {
	database := testRouterDB(t)
	release := make(chan struct{})
	store := &heldLoginStore{Store: authdata.NewStore(database), entered: make(chan struct{}), release: release}
	limiter, err := authentication.NewWorkLimiter(authentication.WorkLimits{Parallel: 1, Queue: 0, MemoryMiB: 256})
	if err != nil {
		t.Fatal(err)
	}
	auth := application.NewAuthService(database, application.WithLoginStore(store), application.WithLoginWorkLimiter(limiter))
	router := NewRouter(Config{AuthService: auth})
	first, done := httptest.NewRecorder(), make(chan struct{})
	go func() { router.ServeHTTP(first, loginCapacityRequest()); close(done) }()
	t.Cleanup(func() { close(release); <-done })
	select {
	case <-store.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("initial login did not acquire work capacity")
	}
	busy := httptest.NewRecorder()
	router.ServeHTTP(busy, loginCapacityRequest())
	if busy.Code != http.StatusTooManyRequests || busy.Header().Get("Retry-After") != "1" || len(busy.Result().Cookies()) != 0 {
		t.Fatalf("busy login status=%d, retry=%q", busy.Code, busy.Header().Get("Retry-After"))
	}
	if strings.Contains(busy.Body.String(), "synthetic-capacity-test") {
		t.Fatal("busy response disclosed the attempted account")
	}
	budgetAuth := application.NewAuthService(database, application.WithLoginStore(exhaustedLoginStore{Store: authdata.NewStore(database)}))
	budget := httptest.NewRecorder()
	NewRouter(Config{AuthService: budgetAuth}).ServeHTTP(budget, loginCapacityRequest())
	if budget.Code != http.StatusTooManyRequests || budget.Header().Get("Retry-After") != "300" {
		t.Fatalf("account window changed: status=%d retry=%q", budget.Code, budget.Header().Get("Retry-After"))
	}
}
