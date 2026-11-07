package authentication

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type memoryStore struct {
	account                                Account
	found, allowed                         bool
	allowErr, readErr, rolesErr, commitErr error
	attempts, lookups, commits, audits     atomic.Int32
	mu                                     sync.Mutex
	lastAttempt                            Attempt
	lastCommit                             Commit
	lastUsername                           string
}

func (s *memoryStore) AllowAttempt(_ context.Context, attempt Attempt) (bool, error) {
	s.attempts.Add(1)
	s.mu.Lock()
	s.lastAttempt = attempt
	s.mu.Unlock()
	return s.allowed, s.allowErr
}
func (s *memoryStore) FindAccount(_ context.Context, username string) (Account, bool, error) {
	s.lookups.Add(1)
	s.mu.Lock()
	s.lastUsername = username
	s.mu.Unlock()
	return s.account, s.found, s.readErr
}
func (s *memoryStore) Roles(context.Context, string) ([]string, error) {
	return []string{"Viewer"}, s.rolesErr
}
func (s *memoryStore) CommitLogin(_ context.Context, proof Commit) error {
	s.commits.Add(1)
	s.mu.Lock()
	s.lastCommit = proof
	s.mu.Unlock()
	return s.commitErr
}
func (s *memoryStore) RecordFailure(context.Context, Audit) error { s.audits.Add(1); return nil }

type passwordStub struct {
	valid                   bool
	beforeVerify            func()
	hashes                  chan string
	verifications, upgrades atomic.Int32
	err                     error
}

func (p *passwordStub) Verify(_, hash string) bool {
	p.verifications.Add(1)
	if p.hashes != nil {
		p.hashes <- hash
	}
	if p.beforeVerify != nil {
		p.beforeVerify()
	}
	return p.valid
}
func (p *passwordStub) Upgrade(string, string) (string, error) {
	p.upgrades.Add(1)
	return "upgraded-hash", p.err
}
func (*passwordStub) DummyHash() string { return "dummy-public-hash" }

func loginFixture() (*memoryStore, *passwordStub, Input) {
	return &memoryStore{account: Account{ID: "account", Username: "Operator", PasswordHash: "old-hash"}, found: true, allowed: true},
		&passwordStub{valid: true}, Input{Username: " Operator ", Password: "supplied password"}
}

func TestUnknownAndIncorrectAccountsBothPerformPasswordWork(t *testing.T) {
	for _, found := range []bool{false, true} {
		store, passwords, input := loginFixture()
		store.found, passwords.valid, passwords.hashes = found, false, make(chan string, 1)
		result, err := NewService(store, WithPasswords(passwords)).Login(t.Context(), input)
		if !errors.Is(err, ErrInvalidLogin) || result != nil || passwords.verifications.Load() != 1 || passwords.upgrades.Load() != 0 || store.commits.Load() != 0 {
			t.Fatal("failed or missing account bypassed work or changed credentials")
		}
		want := "dummy-public-hash"
		if found {
			want = "old-hash"
		}
		if hash := <-passwords.hashes; hash != want || store.audits.Load() != 1 {
			t.Fatal("missing account did not use the dummy verification path")
		}
	}
}

func TestLimitsAndCancellationStopBeforePasswordVerification(t *testing.T) {
	sentinel := errors.New("synthetic budget failure")
	for _, failure := range []error{nil, sentinel} {
		store, passwords, input := loginFixture()
		store.allowed, store.allowErr = false, failure
		_, err := NewService(store, WithPasswords(passwords)).Login(t.Context(), input)
		want := failure
		if want == nil {
			want = ErrLoginThrottled
		}
		if !errors.Is(err, want) || store.lookups.Load() != 0 || passwords.verifications.Load() != 0 {
			t.Fatal("budget rejection triggered password work or masked a storage failure")
		}
	}
	store, passwords, input := loginFixture()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := NewService(store, WithPasswords(passwords)).Login(ctx, input); !errors.Is(err, context.Canceled) || store.attempts.Load() != 0 {
		t.Fatal("cancelled login touched the store")
	}
}

func TestTwoFactorMustCompleteBeforeUpgradeAndSessionCommit(t *testing.T) {
	now := time.Unix(59, 0)
	for _, code := range []string{"", "000000", "287082"} {
		store, passwords, input := loginFixture()
		store.account.TwoFactorEnabled = true
		store.account.TwoFactorSecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
		input.TwoFactorCode = code
		result, err := NewService(store, WithPasswords(passwords), WithClock(func() time.Time { return now })).Login(t.Context(), input)
		if code != "287082" {
			want := ErrInvalidLogin
			if code == "" {
				want = ErrTwoFactorRequired
			}
			if !errors.Is(err, want) || result != nil || passwords.upgrades.Load() != 0 || store.commits.Load() != 0 {
				t.Fatal("incomplete two factor authentication caused a credential or session change")
			}
			continue
		}
		if err != nil || passwords.upgrades.Load() != 1 || store.commits.Load() != 1 || result == nil {
			t.Fatalf("complete proof did not commit: %v", err)
		}
		proof := store.lastCommit
		if proof.TwoFactorStep != 1 || proof.UpgradedHash != "upgraded-hash" || proof.Account.PasswordHash != "old-hash" ||
			proof.AttemptKey != store.lastAttempt.Key || store.lastUsername != "Operator" || !proof.Now.Equal(now) ||
			result.SessionToken == result.CSRFToken || !result.ExpiresAt.Equal(now.Add(12*time.Hour)) {
			t.Fatal("commit lost verified state, time, account scope or independent session tokens")
		}
	}
}

func TestFailedCommitReturnsNoSessionAndIsNeverRetried(t *testing.T) {
	for _, failure := range []error{errors.New("synthetic commit failure"), ErrInvalidLogin} {
		store, passwords, input := loginFixture()
		store.commitErr = failure
		result, err := NewService(store, WithPasswords(passwords)).Login(t.Context(), input)
		if !errors.Is(err, failure) || result != nil || store.commits.Load() != 1 {
			t.Fatal("failed commit returned a usable result or was retried")
		}
	}
}

func TestTechnicalFailuresRemainDistinctFromRejectedCredentials(t *testing.T) {
	for _, stage := range []string{"lookup", "upgrade", "roles"} {
		t.Run(stage, func(t *testing.T) {
			store, passwords, input := loginFixture()
			failure := errors.New("synthetic " + stage + " failure")
			switch stage {
			case "lookup":
				store.readErr = failure
			case "upgrade":
				passwords.err = failure
			case "roles":
				store.rolesErr = failure
			}
			result, err := NewService(store, WithPasswords(passwords)).Login(t.Context(), input)
			if !errors.Is(err, failure) || errors.Is(err, ErrInvalidLogin) || result != nil || store.commits.Load() != 0 {
				t.Fatal("technical failure was masked or committed a session")
			}
		})
	}
}

func TestConcurrentPasswordWorkIsBoundedWithoutTimingAssertions(t *testing.T) {
	store, passwords, input := loginFixture()
	entered, release := make(chan struct{}, 2), make(chan struct{})
	passwords.valid = false
	passwords.beforeVerify = func() { entered <- struct{}{}; <-release }
	service := NewService(store, WithPasswords(passwords))
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Go(func() { _, err := service.Login(t.Context(), input); results <- err })
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	defer workers.Wait()
	defer close(release)
	for range 2 {
		select {
		case <-entered:
		case <-ctx.Done():
			t.Fatal("password worker did not start")
		}
	}
	extra := make(chan error, 1)
	go func() { _, err := service.Login(ctx, input); extra <- err }()
	select {
	case err := <-extra:
		if !errors.Is(err, ErrLoginThrottled) || store.attempts.Load() != 2 {
			t.Fatal("more than two concurrent password checks entered the store")
		}
	case <-ctx.Done():
		t.Fatal("overloaded login did not return promptly while two checks were blocked")
	}
}
