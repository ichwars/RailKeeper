package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"railkeeper/backend/internal/api"
)

func TestBugReportCapabilityCacheUsesTTLAndInvalidation(t *testing.T) {
	now := time.Date(2026, 8, 29, 8, 0, 0, 0, time.UTC)
	validationErr := error(nil)
	calls := 0
	cache := newBugReportCapabilityCache(api.BugReportCapability{ContactAvailable: true}, func(context.Context) error {
		calls++
		return validationErr
	})
	cache.now = func() time.Time { return now }

	if capability := cache.Current(t.Context()); !capability.Enabled || !capability.ContactAvailable || calls != 1 {
		t.Fatalf("initial capability=%#v calls=%d", capability, calls)
	}
	validationErr = errors.New("revoked")
	if capability := cache.Current(t.Context()); !capability.Enabled || calls != 1 {
		t.Fatalf("cached capability=%#v calls=%d", capability, calls)
	}
	cache.Invalidate()
	if capability := cache.Current(t.Context()); capability.Enabled || calls != 2 {
		t.Fatalf("invalidated capability=%#v calls=%d", capability, calls)
	}
	validationErr = nil
	now = now.Add(bugReportCapabilityTTL + time.Second)
	if capability := cache.Current(t.Context()); !capability.Enabled || calls != 3 {
		t.Fatalf("refreshed capability=%#v calls=%d", capability, calls)
	}
}

func TestBugReportCapabilityCacheIgnoresCanceledRequestContext(t *testing.T) {
	validatorCalled := false
	cache := newBugReportCapabilityCache(api.BugReportCapability{}, func(ctx context.Context) error {
		validatorCalled = true
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, bounded := ctx.Deadline(); !bounded {
			t.Fatal("capability validation context has no deadline")
		}
		return nil
	})
	requestContext, cancel := context.WithCancel(context.Background())
	cancel()

	capability := cache.Current(requestContext)
	if !validatorCalled || !capability.Enabled {
		t.Fatalf("validatorCalled=%t capability=%#v", validatorCalled, capability)
	}
}

func TestBugReportCapabilityCacheDoesNotPersistCancellationAsDisabled(t *testing.T) {
	calls := 0
	cache := newBugReportCapabilityCache(api.BugReportCapability{}, func(context.Context) error {
		calls++
		if calls == 1 {
			return context.Canceled
		}
		return nil
	})

	if capability := cache.Current(t.Context()); capability.Enabled {
		t.Fatalf("transient capability=%#v, want disabled response", capability)
	}
	if capability := cache.Current(t.Context()); !capability.Enabled || calls != 2 {
		t.Fatalf("retried capability=%#v calls=%d", capability, calls)
	}
}
