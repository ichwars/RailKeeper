package authentication

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"railkeeper/backend/internal/credentials"
)

func testWorkLimiter(t *testing.T, queue int, wait time.Duration) *WorkLimiter {
	t.Helper()
	work, err := NewWorkLimiter(WorkLimits{Parallel: 1, Queue: queue, Wait: wait, MemoryMiB: 256})
	if err != nil {
		t.Fatal(err)
	}
	return work
}

func waitForQueued(t *testing.T, work *WorkLimiter, count int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	for {
		work.mu.Lock()
		actual := len(work.waiting)
		work.mu.Unlock()
		if actual == count {
			return
		}
		if ctx.Err() != nil {
			t.Fatalf("waiting count %d, want %d", actual, count)
		}
		runtime.Gosched()
	}
}

func TestWorkLimiterFIFOAndReservedCapacity(t *testing.T) {
	work := testWorkLimiter(t, 4, 2*time.Second)
	if err := work.acquire(t.Context()); err != nil {
		t.Fatal(err)
	}
	results := make(chan int, 3)
	for index := range 3 {
		go func() {
			if err := work.acquire(t.Context()); err != nil {
				results <- -1
				return
			}
			results <- index
		}()
		waitForQueued(t, work, index+1)
	}
	if err := work.acquire(t.Context()); !errors.Is(err, ErrLoginThrottled) {
		t.Fatalf("full queue returned %v", err)
	}
	for index := range 3 {
		work.release()
		if actual := <-results; actual != index {
			t.Fatalf("queued waiter %d granted, want %d", actual, index)
		}
	}
	work.release()
	if work.active != 0 || len(work.waiting) != 0 {
		t.Fatal("successful handover retained a slot")
	}
}

func TestWorkLimiterCancellationAndTimeoutReturnCapacity(t *testing.T) {
	for _, cancelWait := range []bool{true, false} {
		work := testWorkLimiter(t, 1, 20*time.Millisecond)
		if err := work.acquire(t.Context()); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(t.Context())
		result := make(chan error, 1)
		go func() { result <- work.acquire(ctx) }()
		if cancelWait {
			waitForQueued(t, work, 1)
			cancel()
		}
		err := <-result
		cancel()
		if cancelWait && !errors.Is(err, context.Canceled) || !cancelWait && !errors.Is(err, ErrLoginThrottled) {
			t.Fatalf("canceled=%v: %v", cancelWait, err)
		}
		work.release()
		if work.active != 0 || len(work.waiting) != 0 {
			t.Fatal("failed waiter retained capacity")
		}
		if err := work.acquire(t.Context()); err != nil {
			t.Fatal("capacity did not recover", err)
		}
		work.release()
	}
}

func TestWorkLimiterGrantCancellationRaceDoesNotLeak(t *testing.T) {
	for range 100 {
		work := testWorkLimiter(t, 1, time.Second)
		if err := work.acquire(t.Context()); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(t.Context())
		result := make(chan error, 1)
		go func() {
			err := work.acquire(ctx)
			if err == nil {
				work.release()
			}
			result <- err
		}()
		waitForQueued(t, work, 1)
		var racers sync.WaitGroup
		racers.Go(cancel)
		racers.Go(work.release)
		racers.Wait()
		err := <-result
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if work.active != 0 || len(work.waiting) != 0 {
			t.Fatal("racing cancellation retained capacity")
		}
	}
}

func TestCanceledHandoffPassesCapacityToNextWaiter(t *testing.T) {
	work := testWorkLimiter(t, 3, time.Second)
	if err := work.acquire(t.Context()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	first := make(chan error, 1)
	second := make(chan error, 1)
	go func() { first <- work.acquire(ctx) }()
	waitForQueued(t, work, 1)
	go func() { second <- work.acquire(t.Context()) }()
	waitForQueued(t, work, 2)
	// Cancellation may win its select before the worker hands over the permit.
	// Hold the state lock until that handover has happened, then verify recovery.
	work.mu.Lock()
	cancel()
	work.releaseLocked()
	work.mu.Unlock()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := <-second; err != nil {
		t.Fatal("next waiter lost the canceled handover", err)
	}
	work.release()
	if work.active != 0 || len(work.waiting) != 0 {
		t.Fatal("canceled handover retained capacity")
	}
}

func TestSharedLimiterRetainsCanceledDerivationUntilItFinishes(t *testing.T) {
	store, passwords, input := loginFixture()
	store.found = false
	work := testWorkLimiter(t, 1, time.Second)
	entered, release := make(chan struct{}, 2), make(chan struct{})
	passwords.beforeVerify = func() { entered <- struct{}{}; <-release }
	first := NewService(store, WithWorkLimiter(work), WithPasswords(passwords))
	second := NewService(store, WithWorkLimiter(work), WithPasswords(passwords))
	ctx, cancel := context.WithCancel(t.Context())
	results := make(chan error, 2)
	go func() { _, err := first.Login(ctx, input); results <- err }()
	<-entered
	cancel()
	go func() { _, err := second.Login(t.Context(), input); results <- err }()
	waitForQueued(t, work, 1)
	if store.attempts.Load() != 1 {
		t.Fatal("cancellation freed an in-progress derivation")
	}
	close(release)
	for range 2 {
		if err := <-results; !errors.Is(err, ErrInvalidLogin) {
			t.Fatal(err)
		}
	}
	if work.active != 0 {
		t.Fatal("completed derivation retained capacity")
	}
	// An omitted injected limiter must retain the bounded default.
	defaultService := NewService(store, WithWorkLimiter(nil))
	if defaultService.work.limits != DefaultWorkLimits() {
		t.Fatal("nil limiter disabled the default")
	}
}

func TestRealPasswordLoginKeepsCompatibleHashAndSessionCommit(t *testing.T) {
	store, _, input := loginFixture()
	hash, err := credentials.Hash(input.Password)
	if err != nil {
		t.Fatal(err)
	}
	store.account.PasswordHash = hash
	store.account.TwoFactorEnabled = false
	service := NewService(store)
	if _, err := service.Login(t.Context(), input); err != nil {
		t.Fatal(err)
	}
	if store.commits.Load() != 1 || store.lastCommit.UpgradedHash != "" {
		t.Fatal("work limiting changed the compatible password or atomic session commit")
	}
}

func TestWorkLimitsRejectUnsafeConfiguration(t *testing.T) {
	valid := DefaultWorkLimits()
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*WorkLimits){
		func(v *WorkLimits) { v.Parallel = 0 }, func(v *WorkLimits) { v.Parallel = 5 },
		func(v *WorkLimits) { v.Queue = -1 }, func(v *WorkLimits) { v.Queue = 65 },
		func(v *WorkLimits) { v.Wait = -1 }, func(v *WorkLimits) { v.Wait = 3 * time.Second },
		func(v *WorkLimits) { v.MemoryMiB = 255 }, func(v *WorkLimits) { v.MemoryMiB = 1025 },
		func(v *WorkLimits) { v.MemoryMiB = 256 },
	} {
		limits := valid
		change(&limits)
		if _, err := NewWorkLimiter(limits); err == nil {
			t.Fatalf("unsafe limits accepted: %+v", limits)
		}
	}
	work := testWorkLimiter(t, 0, 0)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := work.acquire(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := work.acquire(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := work.acquire(t.Context()); !errors.Is(err, ErrLoginThrottled) {
		t.Fatal("zero wait did not reject excess work", err)
	}
	work.release()
}

type measuredPasswords struct {
	passwordPolicy
	active atomic.Int32
	peak   atomic.Int32
}

func (policy *measuredPasswords) Verify(password, hash string) bool {
	active := policy.active.Add(1)
	defer policy.active.Add(-1)
	for peak := policy.peak.Load(); active > peak; peak = policy.peak.Load() {
		if policy.peak.CompareAndSwap(peak, active) {
			break
		}
	}
	return policy.passwordPolicy.Verify(password, hash)
}

func TestRealPasswordBurstRetainsBoundedWorkAndRecovers(t *testing.T) {
	store, _, input := loginFixture()
	store.found = false
	passwords := &measuredPasswords{}
	service := NewService(store, WithPasswords(passwords))
	start := make(chan struct{})
	errorsFound := make(chan error, 32)
	var work sync.WaitGroup
	for range 32 {
		work.Go(func() { <-start; _, err := service.Login(t.Context(), input); errorsFound <- err })
	}
	close(start)
	work.Wait()
	close(errorsFound)
	for err := range errorsFound {
		if !errors.Is(err, ErrLoginThrottled) && !errors.Is(err, ErrInvalidLogin) {
			t.Fatal("unexpected burst outcome", err)
		}
	}
	if passwords.peak.Load() < 1 || passwords.peak.Load() > 2 || service.work.active != 0 || len(service.work.waiting) != 0 {
		t.Fatal("real password work exceeded or retained capacity")
	}
	if _, err := service.Login(t.Context(), input); !errors.Is(err, ErrInvalidLogin) {
		t.Fatal("login did not recover after burst", err)
	}
	t.Logf("32 real-password requests; peak concurrent derivations=%d; no pending work", passwords.peak.Load())
}

// Maximum accepted hashes stay usable under repeated overload. This exercises
// real 256 MiB derivations, rather than estimating cost from the current policy.
func TestMaximumPasswordMemoryRepeatedBurstsRecover(t *testing.T) {
	store, _, input := loginFixture()
	store.account.PasswordHash = strings.Replace(credentials.DummyHash(), "m=65536,", "m=262144,", 1)
	passwords := &measuredPasswords{}
	service := NewService(store, WithPasswords(passwords))
	baseline := runtime.NumGoroutine()
	for round := range 3 {
		start := make(chan struct{})
		results := make(chan error, 16)
		var workers sync.WaitGroup
		for range 16 {
			workers.Go(func() { <-start; _, err := service.Login(t.Context(), input); results <- err })
		}
		close(start)
		workers.Wait()
		close(results)
		for err := range results {
			if !errors.Is(err, ErrInvalidLogin) && !errors.Is(err, ErrLoginBusy) {
				t.Fatal("unexpected sustained-load result", err)
			}
		}
		if service.work.active != 0 || len(service.work.waiting) != 0 || runtime.NumGoroutine() > baseline+32 {
			t.Fatalf("work leaked after maximum-memory round %d", round)
		}
	}
	if passwords.peak.Load() > 2 || passwords.peak.Load() < 1 {
		t.Fatal("unbounded maximum-memory derivations", passwords.peak.Load())
	}
	if _, err := service.Login(t.Context(), input); !errors.Is(err, ErrInvalidLogin) {
		t.Fatal("no recovery after sustained maximum-memory work", err)
	}
	t.Logf("3 bursts of 16, 256 MiB per accepted hash, peak derivations=%d, active/waiting=0", passwords.peak.Load())
}
