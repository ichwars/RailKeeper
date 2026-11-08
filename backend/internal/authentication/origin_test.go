package authentication

import (
	"context"
	"errors"
	"testing"
	"time"

	"railkeeper/backend/internal/credentials"
)

type originMemoryStore struct {
	*memoryStore
	origin                                        Origin
	current                                       Account
	originFound                                   bool
	originErr, resetErr                           error
	lookupUsername, lookupHash, actorID, targetID string
}

func (store *originMemoryStore) FindOrigin(_ context.Context, username, hash string) (Origin, Account, bool, error) {
	store.lookupUsername, store.lookupHash = username, hash
	return store.origin, store.current, store.originFound, store.originErr
}

func (store *originMemoryStore) ResetLoginBudget(_ context.Context, actorID, targetID string, _ time.Time) error {
	store.actorID, store.targetID = actorID, targetID
	return store.resetErr
}

func originFixture() (*originMemoryStore, *passwordStub, Input, time.Time) {
	base, passwords, input := loginFixture()
	now := time.Unix(59, 0)
	input.OriginToken = credentials.RandomToken()
	store := &originMemoryStore{memoryStore: base, current: base.account, originFound: true}
	store.origin = Origin{UserID: base.account.ID, TokenHash: credentials.HashToken(input.OriginToken),
		BudgetKey: "protected-budget", StateHash: CredentialState(base.account), ExpiresAt: now.Add(time.Hour)}
	return store, passwords, input, now
}

func TestRecognizedOriginHasSeparateBudgetAndRequiresFullAuthentication(t *testing.T) {
	for _, validPassword := range []bool{false, true} {
		store, passwords, input, now := originFixture()
		passwords.valid = validPassword
		result, err := NewService(store, WithPasswords(passwords), WithClock(func() time.Time { return now })).Login(t.Context(), input)
		if store.lastAttempt.Scope != OriginBudgetScope || store.lastAttempt.Key != "protected-budget" || store.lookupHash == input.OriginToken {
			t.Fatal("browser proof did not select an independent hashed budget")
		}
		if !validPassword {
			if !errors.Is(err, ErrInvalidLogin) || result != nil || store.commits.Load() != 0 {
				t.Fatal("recognized origin bypassed password")
			}
			continue
		}
		if err != nil || result.OriginToken != input.OriginToken || store.lastCommit.AttemptScope != OriginBudgetScope || store.lastCommit.Origin == nil ||
			store.lastCommit.Origin.StateHash == store.origin.StateHash || !store.lastCommit.OriginRecognized {
			t.Fatal("recognized login lost its token or upgraded credential binding", err)
		}
	}
	store, passwords, input, now := originFixture()
	store.account.TwoFactorEnabled = true
	store.account.TwoFactorSecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	store.current = store.account
	store.origin.StateHash = CredentialState(store.account)
	if result, err := NewService(store, WithPasswords(passwords), WithClock(func() time.Time { return now })).Login(t.Context(), input); !errors.Is(err, ErrTwoFactorRequired) || result != nil || store.commits.Load() != 0 {
		t.Fatal("recognized browser skipped second factor")
	}
}

func TestOriginCannotFollowAReplacedAccountBetweenLookupAndVerification(t *testing.T) {
	store, passwords, input, now := originFixture()
	store.account.ID = "replacement-account"
	if result, err := NewService(store, WithPasswords(passwords), WithClock(func() time.Time { return now })).Login(t.Context(), input); !errors.Is(err, ErrInvalidLogin) || result != nil || store.commits.Load() != 0 {
		t.Fatal("old origin followed a replaced account", err)
	}
}

func TestInvalidOriginNeverSelectsProtectedBudget(t *testing.T) {
	for _, change := range []func(*originMemoryStore, *Input, time.Time){
		func(s *originMemoryStore, _ *Input, _ time.Time) { s.originFound = false },
		func(s *originMemoryStore, _ *Input, n time.Time) { s.origin.ExpiresAt = n },
		func(s *originMemoryStore, _ *Input, _ time.Time) { s.current.PasswordHash = "changed" },
		func(s *originMemoryStore, _ *Input, _ time.Time) { s.current.Username = "another" },
		func(s *originMemoryStore, _ *Input, _ time.Time) { s.origin.UserID = "another" },
		func(_ *originMemoryStore, i *Input, _ time.Time) { i.OriginToken = "forged" },
		func(_ *originMemoryStore, i *Input, _ time.Time) { i.OriginToken = credentials.RandomToken() },
	} {
		store, passwords, input, now := originFixture()
		change(store, &input, now)
		result, err := NewService(store, WithPasswords(passwords), WithClock(func() time.Time { return now })).Login(t.Context(), input)
		if err != nil || store.lastAttempt.Scope != AccountBudgetScope || result.OriginToken == input.OriginToken ||
			!result.OriginExpiresAt.Equal(now.Add(OriginTTL)) {
			t.Fatal("invalid origin retained recognition", err)
		}
	}
	store, passwords, input, now := originFixture()
	store.originErr = errors.New("synthetic storage failure")
	if _, err := NewService(store, WithPasswords(passwords), WithClock(func() time.Time { return now })).Login(t.Context(), input); !errors.Is(err, store.originErr) || passwords.verifications.Load() != 0 {
		t.Fatal("masked origin lookup failure")
	}
}

func TestOriginRecoveryUsesThePersistencePermissionBoundary(t *testing.T) {
	store, _, _, _ := originFixture()
	service := NewService(store)
	store.resetErr = ErrForbidden
	if err := service.ResetLoginBudget(t.Context(), "actor", "target"); !errors.Is(err, ErrForbidden) || store.actorID != "actor" || store.targetID != "target" {
		t.Fatal("recovery bypassed store permission")
	}
	base, _, _ := loginFixture()
	if err := NewService(base).ResetLoginBudget(t.Context(), "actor", "target"); !errors.Is(err, ErrForbidden) {
		t.Fatal("unsupported recovery was allowed")
	}
}

func TestMixedQueueReservesRecognizedCapacityAndAlternatesClasses(t *testing.T) {
	work := testWorkLimiter(t, 4, 2*time.Second)
	if err := work.acquire(t.Context()); err != nil {
		t.Fatal(err)
	}
	results := make(chan string, 4)
	for index, name := range []string{"unknown-1", "unknown-2", "unknown-3"} {
		go func() {
			if err := work.acquireOrigin(t.Context(), false); err != nil {
				results <- "failed"
				return
			}
			results <- name
		}()
		waitForQueued(t, work, index+1)
	}
	if err := work.acquireOrigin(t.Context(), false); !errors.Is(err, ErrLoginBusy) {
		t.Fatal("unknown flood used reserved slot")
	}
	go func() {
		if err := work.acquireOrigin(t.Context(), true); err != nil {
			results <- "failed"
			return
		}
		results <- "known"
	}()
	waitForQueued(t, work, 4)
	for _, want := range []string{"known", "unknown-1", "unknown-2", "unknown-3"} {
		work.release()
		if actual := <-results; actual != want {
			t.Fatalf("grant %s, want %s", actual, want)
		}
	}
	work.release()
}

type blockedOriginStore struct {
	*originMemoryStore
	entered chan struct{}
}

func (store *blockedOriginStore) FindOrigin(ctx context.Context, _, _ string) (Origin, Account, bool, error) {
	store.entered <- struct{}{}
	<-ctx.Done()
	return Origin{}, Account{}, false, ctx.Err()
}

func TestOriginChecksHaveIndependentBoundedCapacityAndRecoverOnCancellation(t *testing.T) {
	store, _, input, _ := originFixture()
	blocked := &blockedOriginStore{originMemoryStore: store, entered: make(chan struct{}, 2)}
	service := NewService(blocked)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	results := make(chan error, 2)
	for range 2 {
		go func() { _, err := service.Login(ctx, input); results <- err }()
	}
	for range 2 {
		<-blocked.entered
	}
	if _, err := service.Login(t.Context(), input); !errors.Is(err, ErrLoginBusy) {
		t.Fatal("unbounded pre-work origin lookup", err)
	}
	cancel()
	for range 2 {
		if err := <-results; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
	if service.work.originChecks != 0 {
		t.Fatal("cancelled origin lookup retained capacity")
	}
	if _, err := service.Login(t.Context(), input); !errors.Is(err, ErrLoginBusy) {
		t.Fatal("bounded lookup deadline masked", err)
	}
}
