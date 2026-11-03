package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"railkeeper/backend/internal/api"
)

const (
	bugReportCapabilityTTL               = 30 * time.Second
	bugReportCapabilityValidationTimeout = 10 * time.Second
)

type bugReportCapabilityCache struct {
	mu         sync.Mutex
	base       api.BugReportCapability
	validator  func(context.Context) error
	ttl        time.Duration
	now        func() time.Time
	capability api.BugReportCapability
	expiresAt  time.Time
	ready      bool
}

func newBugReportCapabilityCache(
	base api.BugReportCapability,
	validator func(context.Context) error,
) *bugReportCapabilityCache {
	return &bugReportCapabilityCache{
		base: base, validator: validator, ttl: bugReportCapabilityTTL, now: time.Now,
	}
}

func (cache *bugReportCapabilityCache) Current(_ context.Context) api.BugReportCapability {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	now := cache.now()
	if cache.ready && now.Before(cache.expiresAt) {
		return cache.capability
	}
	capability := cache.base
	validationContext, cancel := context.WithTimeout(context.Background(), bugReportCapabilityValidationTimeout)
	defer cancel()
	validationErr := error(nil)
	if cache.validator == nil {
		validationErr = errors.New("bug report capability validator is unavailable")
	} else {
		validationErr = cache.validator(validationContext)
	}
	if errors.Is(validationErr, context.Canceled) || errors.Is(validationErr, context.DeadlineExceeded) {
		if cache.ready {
			cache.expiresAt = now
			return cache.capability
		}
		return capability
	}
	capability.Enabled = validationErr == nil
	cache.capability = capability
	cache.expiresAt = now.Add(cache.ttl)
	cache.ready = true
	return capability
}

func (cache *bugReportCapabilityCache) Invalidate() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.ready = false
	cache.expiresAt = time.Time{}
}
