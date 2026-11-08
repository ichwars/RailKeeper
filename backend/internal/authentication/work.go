package authentication

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"railkeeper/backend/internal/credentials"
)

type WorkLimits struct {
	Parallel  int
	Queue     int
	Wait      time.Duration
	MemoryMiB int
}

func DefaultWorkLimits() WorkLimits {
	return WorkLimits{Parallel: 2, Queue: 8, Wait: time.Second, MemoryMiB: 512}
}

func (limits WorkLimits) Validate() error {
	if limits.Parallel < 1 || limits.Parallel > 4 || limits.Queue < 0 || limits.Queue > 64 ||
		limits.Wait < 0 || limits.Wait > 2*time.Second || limits.MemoryMiB < credentials.MaxPasswordMemoryMiB ||
		limits.MemoryMiB > 1024 || limits.Parallel*credentials.MaxPasswordMemoryMiB > limits.MemoryMiB {
		return errors.New("login work limits exceed the bounded concurrency, queue, wait or password memory budget")
	}
	return nil
}

type workWaiter struct {
	ready   chan struct{}
	granted bool
}

// WorkLimiter bounds whole login attempts, including any password hash upgrade.
type WorkLimiter struct {
	mu      sync.Mutex
	limits  WorkLimits
	active  int
	waiting []*workWaiter
}

func NewWorkLimiter(limits WorkLimits) (*WorkLimiter, error) {
	if err := limits.Validate(); err != nil {
		return nil, err
	}
	return &WorkLimiter{limits: limits}, nil
}

func WithWorkLimiter(limiter *WorkLimiter) Option {
	return func(service *Service) {
		if limiter != nil {
			service.work = limiter
		}
	}
}

func (limiter *WorkLimiter) acquire(ctx context.Context) error {
	limiter.mu.Lock()
	if err := ctx.Err(); err != nil {
		limiter.mu.Unlock()
		return err
	}
	if limiter.active < limiter.limits.Parallel {
		limiter.active++
		limiter.mu.Unlock()
		return nil
	}
	if limiter.limits.Wait == 0 || len(limiter.waiting) >= limiter.limits.Queue {
		limiter.mu.Unlock()
		return ErrLoginBusy
	}
	waiter := &workWaiter{ready: make(chan struct{})}
	limiter.waiting = append(limiter.waiting, waiter)
	limiter.mu.Unlock()
	timer := time.NewTimer(limiter.limits.Wait)
	defer timer.Stop()
	select {
	case <-waiter.ready:
		if err := ctx.Err(); err != nil {
			limiter.release()
			return err
		}
		return nil
	case <-ctx.Done():
		limiter.cancel(waiter)
		return ctx.Err()
	case <-timer.C:
		limiter.cancel(waiter)
		return ErrLoginBusy
	}
}

func (limiter *WorkLimiter) cancel(waiter *workWaiter) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if waiter.granted {
		limiter.releaseLocked()
		return
	}
	for index, pending := range limiter.waiting {
		if pending == waiter {
			limiter.waiting = slices.Delete(limiter.waiting, index, index+1)
			return
		}
	}
}

func (limiter *WorkLimiter) release() {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	limiter.releaseLocked()
}

func (limiter *WorkLimiter) releaseLocked() {
	if len(limiter.waiting) == 0 {
		limiter.active--
		return
	}
	next := limiter.waiting[0]
	limiter.waiting = slices.Delete(limiter.waiting, 0, 1)
	next.granted = true
	close(next.ready)
}
