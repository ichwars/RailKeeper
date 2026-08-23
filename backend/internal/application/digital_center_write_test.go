package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDigitalCenterWritePreviewCreatesActorBoundHashedGrantFromDryRun(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)

	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"protocol", "name", "address"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Token == "" || preview.ExpiresAt != "2026-08-21T10:10:00Z" ||
		preview.Direction != DigitalCenterWriteRailKeeperToCenter ||
		!reflect.DeepEqual(preview.Fields, []string{"address", "name"}) || len(preview.Changes) != 2 {
		t.Fatalf("preview = %#v", preview)
	}
	tokenBytes, err := base64.RawURLEncoding.DecodeString(preview.Token)
	if err != nil || len(tokenBytes) != 32 {
		t.Fatalf("public token decodes to %d bytes, err=%v", len(tokenBytes), err)
	}
	tokenDigest := sha256.Sum256([]byte(preview.Token))
	if fixture.repository.grant.TokenHash != hex.EncodeToString(tokenDigest[:]) ||
		fixture.repository.grant.TokenHash == preview.Token ||
		fixture.repository.grant.ActorUserID != "admin-1" ||
		fixture.repository.grant.SessionID != fixture.session.ID ||
		fixture.repository.grant.WorkItemID != fixture.item.ID ||
		fixture.repository.grant.PreviewHash == "" {
		t.Fatalf("persisted grant = %#v", fixture.repository.grant)
	}
	if len(fixture.ecos.syncCalls) != 1 || !fixture.ecos.syncCalls[0].DryRun || fixture.ecos.syncCalls[0].Confirm ||
		fixture.ecos.syncCalls[0].Host != "trusted-center.local" || fixture.ecos.syncCalls[0].Port != 15471 {
		t.Fatalf("preview sync calls = %#v", fixture.ecos.syncCalls)
	}
}

func TestDigitalCenterWritePreviewRejectsUntrustedDryRunChanges(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.ecos.tamperedDesired = "Nicht der RailKeeper-Wert"

	_, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
	if !errors.Is(err, ErrDigitalCenterDeviceOutput) {
		t.Fatalf("error = %v, want invalid dry-run output", err)
	}
	if fixture.repository.grant.TokenHash != "" {
		t.Fatalf("unsafe dry-run created grant: %#v", fixture.repository.grant)
	}
}

func TestDigitalCenterWritePreviewRejectsUnsafeWorkspaceStateBeforeDeviceRead(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*digitalCenterWriteFixture)
		input  DigitalCenterWritePreviewInput
		want   error
	}{
		{name: "stale read", mutate: func(f *digitalCenterWriteFixture) {
			f.session.ReadCompletedAt = "2026-08-21T09:49:59Z"
			f.repository.session = f.session
		}, want: ErrDigitalCenterReadNotFresh},
		{name: "missing completed read", mutate: func(f *digitalCenterWriteFixture) {
			f.session.State = DigitalCenterSessionInterrupted
			f.session.ReadCompletedAt = ""
			f.repository.session = f.session
		}, want: ErrDigitalCenterReadNotFresh},
		{name: "unresolved conflict", mutate: func(f *digitalCenterWriteFixture) {
			f.item.CompareStatus = DigitalCompareConflict
			f.item.Conflicts = []map[string]any{{"vehicleId": "vehicle-2"}}
			f.repository.item = f.item
		}, want: ErrDigitalCenterConflictUnresolved},
		{name: "unsupported capability", mutate: func(f *digitalCenterWriteFixture) {
			f.session.Capabilities.WriteLocomotives = false
			f.repository.session = f.session
		}, want: ErrDigitalCenterCapabilityUnavailable},
		{name: "unsupported field", input: DigitalCenterWritePreviewInput{Fields: []string{"cv29"}},
			want: ErrDigitalCenterWriteFieldUnsupported},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newDigitalCenterWriteFixture(t)
			if test.mutate != nil {
				test.mutate(fixture)
			}
			_, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
				test.input, "admin-1")
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if len(fixture.ecos.syncCalls) != 0 || fixture.repository.grant.TokenHash != "" {
				t.Fatalf("unsafe preview touched device/grant: calls=%#v grant=%#v",
					fixture.ecos.syncCalls, fixture.repository.grant)
			}
		})
	}
}

func TestDigitalCenterWriteConfirmConsumesBeforeMutationVerifiesMapsAndAudits(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	fixture.ecos.beforeConfirmedWrite = func() {
		if !fixture.repository.consumed {
			t.Fatal("device mutation happened before atomic grant consumption")
		}
	}

	result, err := fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWriteConfirmInput{Token: preview.Token, Confirm: true, Fields: preview.Fields}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Applied || !result.Verified || result.Result != DigitalCenterWriteVerified {
		t.Fatalf("result = %#v", result)
	}
	if len(fixture.ecos.syncCalls) != 3 || !fixture.ecos.syncCalls[1].DryRun ||
		!fixture.ecos.syncCalls[2].Confirm || fixture.ecos.syncCalls[2].DryRun {
		t.Fatalf("sync calls = %#v", fixture.ecos.syncCalls)
	}
	confirmedDesired := fixture.ecos.syncCalls[2].Desired
	if confirmedDesired.Name != "Neue Lok" || confirmedDesired.Address != 0 || confirmedDesired.Protocol != "" {
		t.Fatalf("confirmed unsupported/unpreviewed fields: %#v", confirmedDesired)
	}
	if fixture.vehicles.mapping == nil || fixture.vehicles.mapping.ExternalID != "3" ||
		fixture.vehicles.mapping.ExternalName != "Neue Lok" ||
		fixture.vehicles.mapping.ExternalAddress != "18" ||
		fixture.vehicles.mapping.ExternalProtocol != "DCC" ||
		fixture.vehicles.mapping.SyncStatus != "synced" {
		t.Fatalf("verified mapping = %#v", fixture.vehicles.mapping)
	}
	if len(fixture.audit.entries) != 1 {
		t.Fatalf("audit entries = %#v", fixture.audit.entries)
	}
	entry := fixture.audit.entries[0]
	if entry.actor != "admin-1" || entry.action != "DigitalCenterSynchronization" ||
		entry.targetType != "digital_center_work_item" || entry.targetID != fixture.item.ID {
		t.Fatalf("audit entry = %#v", entry)
	}
	if strings.Contains(entry.details, "Alte Lok") || strings.Contains(entry.details, "Neue Lok") ||
		strings.Contains(entry.details, "<REPLY") {
		t.Fatalf("audit leaked raw values: %s", entry.details)
	}
	var details struct {
		Station  string   `json:"station"`
		ObjectID string   `json:"objectId"`
		Fields   []string `json:"fields"`
		Result   string   `json:"result"`
	}
	if err := json.Unmarshal([]byte(entry.details), &details); err != nil {
		t.Fatal(err)
	}
	if details.Station != "ecos" || details.ObjectID != "3" ||
		!reflect.DeepEqual(details.Fields, []string{"name"}) || details.Result != "verified" {
		t.Fatalf("audit details = %#v", details)
	}
}

func TestDigitalCenterCreatePreviewAndConfirmVerifiesMapsAndAudits(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.item.CenterObjectID = "42"
	fixture.item.Name = "BR 18 201 Roco S"
	fixture.item.Address = 4405
	fixture.item.Protocol = "DCC"
	fixture.item.CompareStatus = DigitalCompareMissing
	fixture.item.StationStatus = "missing"
	fixture.item.Center = map[string]any{}
	fixture.item.RailKeeper = map[string]any{
		"vehicleId": "vehicle-1", "name": "BR 18 201 Roco S", "decoderAddress": 4405, "protocol": "DCC",
	}
	fixture.repository.item = fixture.item
	fixture.ecos.verificationName = "BR 18 201 Roco S"
	fixture.ecos.verificationAddress = 4405
	fixture.ecos.verificationProtocol = "DCC"
	fixture.ecos.createObjectID = 1002

	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Operation: DigitalCenterWriteCreate,
			Fields: []string{"name", "address", "protocol"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Operation != DigitalCenterWriteCreate || preview.ObjectID != "" ||
		!reflect.DeepEqual(preview.Fields, []string{"address", "name", "protocol"}) || len(preview.Changes) != 3 {
		t.Fatalf("preview = %#v", preview)
	}
	if len(fixture.ecos.syncCalls) != 0 || len(fixture.ecos.createCalls) != 0 {
		t.Fatalf("create preview mutated/read target: sync=%#v create=%#v", fixture.ecos.syncCalls,
			fixture.ecos.createCalls)
	}

	fixture.ecos.beforeConfirmedWrite = func() {
		if !fixture.repository.consumed {
			t.Fatal("device creation happened before atomic grant consumption")
		}
	}
	result, err := fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWriteConfirmInput{Operation: DigitalCenterWriteCreate, Token: preview.Token,
			Confirm: true, Fields: preview.Fields}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Operation != DigitalCenterWriteCreate || result.ObjectID != "1002" || !result.Applied ||
		!result.Verified || result.Result != DigitalCenterWriteVerified {
		t.Fatalf("result = %#v", result)
	}
	if len(fixture.ecos.createCalls) != 1 || !fixture.ecos.createCalls[0].Confirm ||
		fixture.ecos.createCalls[0].Desired.Name != "BR 18 201 Roco S" ||
		fixture.ecos.createCalls[0].Desired.Address != 4405 || fixture.ecos.createCalls[0].Desired.Protocol != "DCC" {
		t.Fatalf("create calls = %#v", fixture.ecos.createCalls)
	}
	if fixture.vehicles.mapping == nil || fixture.vehicles.mapping.ExternalID != "1002" ||
		fixture.vehicles.previousExternalID != "42" || fixture.repository.item.CenterObjectID != "1002" ||
		fixture.repository.item.CompareStatus != DigitalCompareOK {
		t.Fatalf("mapping=%#v work item=%#v", fixture.vehicles.mapping, fixture.repository.item)
	}
	if len(fixture.audit.entries) != 1 || !strings.Contains(fixture.audit.entries[0].details, `"operation":"create"`) {
		t.Fatalf("audit entries = %#v", fixture.audit.entries)
	}
}

func TestDigitalCenterCreateVerificationMismatchRebindsMappingAsLinked(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.item.CenterObjectID = "42"
	fixture.item.Name = "BR 18"
	fixture.item.Address = 4405
	fixture.item.Protocol = "DCC"
	fixture.item.CompareStatus = DigitalCompareMissing
	fixture.item.StationStatus = "missing"
	fixture.item.Center = map[string]any{}
	fixture.item.RailKeeper = map[string]any{
		"vehicleId": "vehicle-1", "name": "BR 18", "decoderAddress": 4405, "protocol": "DCC",
	}
	fixture.repository.item = fixture.item
	fixture.ecos.createObjectID = 1002
	fixture.ecos.verificationName = "Abweichender Name"
	fixture.ecos.verificationAddress = 4405
	fixture.ecos.verificationProtocol = "DCC"

	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Operation: DigitalCenterWriteCreate,
			Fields: []string{"name", "address", "protocol"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWriteConfirmInput{Operation: DigitalCenterWriteCreate, Token: preview.Token,
			Confirm: true, Fields: preview.Fields}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Result != DigitalCenterWriteVerificationFailed || result.Verified {
		t.Fatalf("result=%#v", result)
	}
	if fixture.vehicles.mapping == nil || fixture.vehicles.mapping.ExternalID != "1002" ||
		fixture.vehicles.mapping.SyncStatus != "linked" || fixture.vehicles.previousExternalID != "42" {
		t.Fatalf("mapping=%#v previous=%q", fixture.vehicles.mapping, fixture.vehicles.previousExternalID)
	}
}

func TestDigitalCenterWriteConfirmRejectsFalseActorExpiredMismatchedAndConsumedGrants(t *testing.T) {
	tests := []struct {
		name       string
		confirm    bool
		actor      string
		consumeErr error
		mutate     func(*digitalCenterWriteFixture)
		want       error
	}{
		{name: "confirmation missing", confirm: false, actor: "admin-1", want: ErrDigitalCenterConfirmationRequired},
		{name: "actor mismatch", confirm: true, actor: "admin-2",
			consumeErr: ErrDigitalCenterGrantActorMismatch, want: ErrDigitalCenterGrantActorMismatch},
		{name: "expired", confirm: true, actor: "admin-1",
			consumeErr: ErrDigitalCenterGrantExpired, want: ErrDigitalCenterGrantExpired},
		{name: "consumed", confirm: true, actor: "admin-1",
			consumeErr: ErrDigitalCenterGrantConsumed, want: ErrDigitalCenterGrantConsumed},
		{name: "unknown token", confirm: true, actor: "admin-1",
			consumeErr: sql.ErrNoRows, want: ErrDigitalCenterGrantMismatch},
		{name: "endpoint mismatch", confirm: true, actor: "admin-1", mutate: func(f *digitalCenterWriteFixture) {
			f.repository.grant.WorkItemID = "other-item"
		}, want: ErrDigitalCenterGrantMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newDigitalCenterWriteFixture(t)
			preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
				DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
			if err != nil {
				t.Fatal(err)
			}
			fixture.repository.consumeErr = test.consumeErr
			if test.mutate != nil {
				test.mutate(fixture)
			}
			_, err = fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID,
				DigitalCenterWriteConfirmInput{Token: preview.Token, Confirm: test.confirm, Fields: preview.Fields},
				test.actor)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			for _, call := range fixture.ecos.syncCalls[1:] {
				if call.Confirm {
					t.Fatalf("rejected grant sent confirmed device call: %#v", call)
				}
			}
			if fixture.vehicles.mapping != nil || len(fixture.audit.entries) != 0 {
				t.Fatalf("rejected grant mutated local state: mapping=%#v audit=%#v",
					fixture.vehicles.mapping, fixture.audit.entries)
			}
		})
	}
}

func TestDigitalCenterWriteConfirmRejectsStalePreviewAndConsumesGrant(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	fixture.ecos.currentName = "Zwischenzeitlich geändert"

	_, err = fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWriteConfirmInput{Token: preview.Token, Confirm: true, Fields: preview.Fields}, "admin-1")
	if !errors.Is(err, ErrDigitalCenterPreviewStale) || !fixture.repository.consumed {
		t.Fatalf("stale error=%v consumed=%v", err, fixture.repository.consumed)
	}
	if fixture.ecos.confirmCalls != 0 || fixture.vehicles.mapping != nil {
		t.Fatalf("stale preview wrote device/local mapping: confirms=%d mapping=%#v",
			fixture.ecos.confirmCalls, fixture.vehicles.mapping)
	}
}

func TestDigitalCenterWriteConfirmTreatsAlreadyAppliedPreviewAsStale(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	fixture.ecos.currentName = "Neue Lok"

	_, err = fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWriteConfirmInput{Token: preview.Token, Confirm: true, Fields: preview.Fields}, "admin-1")
	if !errors.Is(err, ErrDigitalCenterPreviewStale) || !fixture.repository.consumed {
		t.Fatalf("already-applied error=%v consumed=%v", err, fixture.repository.consumed)
	}
	if fixture.ecos.confirmCalls != 0 || fixture.vehicles.mapping != nil {
		t.Fatalf("already-applied preview wrote device/local mapping: confirms=%d mapping=%#v",
			fixture.ecos.confirmCalls, fixture.vehicles.mapping)
	}
}

func TestDigitalCenterWriteVerificationMismatchIsVisibleAndDoesNotMap(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	fixture.ecos.verificationName = "Nicht übernommen"

	result, err := fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWriteConfirmInput{Token: preview.Token, Confirm: true, Fields: preview.Fields}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Applied || result.Verified || result.Result != DigitalCenterWriteVerificationFailed ||
		fixture.vehicles.mapping != nil {
		t.Fatalf("result=%#v mapping=%#v", result, fixture.vehicles.mapping)
	}
	if result.WorkItem == nil || result.WorkItem.Name != "Nicht übernommen" ||
		result.WorkItem.CompareStatus != DigitalCompareDeviation ||
		fixture.repository.item.Name != "Nicht übernommen" || fixture.repository.item.StationStatus != "read" {
		t.Fatalf("result item=%#v persisted item=%#v", result.WorkItem, fixture.repository.item)
	}
	if len(fixture.audit.entries) != 1 || !strings.Contains(fixture.audit.entries[0].details,
		`"result":"verification_failed"`) {
		t.Fatalf("verification audit = %#v", fixture.audit.entries)
	}
}

func TestDigitalCenterWriteGrantCannotBeReused(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	input := DigitalCenterWriteConfirmInput{Token: preview.Token, Confirm: true, Fields: preview.Fields}
	if _, err := fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID, input, "admin-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID, input, "admin-1"); !errors.Is(err, ErrDigitalCenterGrantConsumed) {
		t.Fatalf("second confirm error = %v", err)
	}
	if fixture.ecos.confirmCalls != 1 {
		t.Fatalf("confirmed writes = %d, want one", fixture.ecos.confirmCalls)
	}
}

func TestDigitalCenterWriteGrantAllowsOnlyOneConcurrentConfirmation(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	input := DigitalCenterWriteConfirmInput{Token: preview.Token, Confirm: true, Fields: preview.Fields}
	errorsSeen := make(chan error, 2)
	start := make(chan struct{})
	for range 2 {
		go func() {
			<-start
			_, confirmErr := fixture.service.ConfirmWrite(
				t.Context(), fixture.session.ID, fixture.item.ID, input, "admin-1",
			)
			errorsSeen <- confirmErr
		}()
	}
	close(start)

	succeeded := 0
	consumed := 0
	for range 2 {
		switch confirmErr := <-errorsSeen; {
		case confirmErr == nil:
			succeeded++
		case errors.Is(confirmErr, ErrDigitalCenterGrantConsumed):
			consumed++
		default:
			t.Fatalf("concurrent confirmation error = %v", confirmErr)
		}
	}
	if succeeded != 1 || consumed != 1 || fixture.ecos.confirmCalls != 1 {
		t.Fatalf("succeeded=%d consumed=%d device writes=%d", succeeded, consumed, fixture.ecos.confirmCalls)
	}
}

func TestDigitalCenterConfirmWritePausesWritesVerifiesAndResumes(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.ecos.liveStatus = ECoSLiveStatus{Provider: "ecos", State: ECoSLiveRunning, Connected: true}
	preview := previewDigitalCenterWriteFixture(t, fixture, []string{"name", "address"})
	fixture.ecos.events = nil

	result, err := confirmDigitalCenterWriteFixture(t, fixture, preview)
	if err != nil {
		t.Fatal(err)
	}
	wantEvents := []string{"pause", "dry-run", "list", "write", "read-target", "resume"}
	if !reflect.DeepEqual(fixture.ecos.events, wantEvents) {
		t.Fatalf("events=%#v want=%#v", fixture.ecos.events, wantEvents)
	}
	if result.Result != DigitalCenterWriteVerified || result.WorkItem == nil ||
		!result.LiveMonitor.WasRunning || !result.LiveMonitor.Restarted {
		t.Fatalf("result=%#v", result)
	}
}

func TestDigitalCenterConfirmWriteWaitsForPublishedStoppedLiveSession(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.ecos.liveStatus = ECoSLiveStatus{Provider: "ecos", State: ECoSLiveStopped, Connected: false}
	preview := previewDigitalCenterWriteFixture(t, fixture, []string{"name"})
	fixture.ecos.events = nil

	result, err := confirmDigitalCenterWriteFixture(t, fixture, preview)
	if err != nil {
		t.Fatal(err)
	}
	if len(fixture.ecos.events) == 0 || fixture.ecos.events[0] != "pause" ||
		result.LiveMonitor.WasRunning || result.LiveMonitor.Restarted {
		t.Fatalf("events=%#v live=%#v", fixture.ecos.events, result.LiveMonitor)
	}
}

func TestDigitalCenterConfirmWriteReturnsUnknownWithoutRetry(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	preview := previewDigitalCenterWriteFixture(t, fixture, []string{"name"})
	fixture.ecos.events = nil
	fixture.ecos.writeErr = ErrECoSWriteStateUnknown

	result, err := confirmDigitalCenterWriteFixture(t, fixture, preview)
	if err != nil || result.Result != DigitalCenterWriteUnknown || fixture.ecos.writeCalls != 1 {
		t.Fatalf("result=%#v err=%v calls=%d", result, err, fixture.ecos.writeCalls)
	}
}

func TestDigitalCenterConfirmWriteRecordsUnknownWhenVerificationReadFails(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	preview := previewDigitalCenterWriteFixture(t, fixture, []string{"name"})
	fixture.ecos.readErr = errors.New("connection closed")

	result, err := confirmDigitalCenterWriteFixture(t, fixture, preview)
	if err != nil || result.Result != DigitalCenterWriteUnknown || len(fixture.repository.messages) != 1 ||
		fixture.repository.messages[0].Code != DigitalCenterMessageWriteUnknown {
		t.Fatalf("result=%#v err=%v messages=%#v", result, err, fixture.repository.messages)
	}
}

func TestDigitalCenterConfirmWritePreservesDeviceOutcomeWhenLocalPersistenceFails(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*digitalCenterWriteFixture)
	}{
		{name: "audit", mutate: func(fixture *digitalCenterWriteFixture) {
			fixture.audit.err = errors.New("audit unavailable")
		}},
		{name: "mapping", mutate: func(fixture *digitalCenterWriteFixture) {
			fixture.vehicles.err = errors.New("mapping unavailable")
		}},
		{name: "work item", mutate: func(fixture *digitalCenterWriteFixture) {
			fixture.repository.updateErr = errors.New("work item unavailable")
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newDigitalCenterWriteFixture(t)
			preview := previewDigitalCenterWriteFixture(t, fixture, []string{"name"})
			test.mutate(fixture)

			result, err := confirmDigitalCenterWriteFixture(t, fixture, preview)
			if err != nil || result.Result != DigitalCenterWriteUnknown || !result.Applied ||
				!result.Verified || fixture.ecos.writeCalls != 1 {
				t.Fatalf("result=%#v err=%v writes=%d", result, err, fixture.ecos.writeCalls)
			}
		})
	}
}

func TestDigitalCenterConfirmWriteKeepsVerifiedResultWhenResumeFails(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.ecos.liveStatus = ECoSLiveStatus{Provider: "ecos", State: ECoSLiveRunning, Connected: true}
	preview := previewDigitalCenterWriteFixture(t, fixture, []string{"name"})
	fixture.ecos.events = nil
	fixture.ecos.resumeErr = errors.New("restart failed")

	result, err := confirmDigitalCenterWriteFixture(t, fixture, preview)
	if err != nil || result.Result != DigitalCenterWriteVerified || result.LiveMonitor.Restarted {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestDigitalCenterConfirmWriteAbortsBeforeMutationWhenPauseFails(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.ecos.liveStatus = ECoSLiveStatus{Provider: "ecos", State: ECoSLiveRunning, Connected: true}
	preview := previewDigitalCenterWriteFixture(t, fixture, []string{"name"})
	fixture.ecos.events = nil
	fixture.ecos.pauseErr = errors.New("live connection did not close")

	_, err := confirmDigitalCenterWriteFixture(t, fixture, preview)
	if !errors.Is(err, ErrDigitalCenterLivePauseFailed) || fixture.ecos.writeCalls != 0 {
		t.Fatalf("error=%v writes=%d", err, fixture.ecos.writeCalls)
	}
}

func TestDigitalCenterConfirmWriteRechecksAddressOwnershipAfterPreview(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	preview := previewDigitalCenterWriteFixture(t, fixture, []string{"address"})
	fixture.ecos.events = nil
	fixture.ecos.locomotives = append(fixture.ecos.locomotives,
		ECoSLocomotive{ObjectID: 2002, Name: "Other", Address: 18, Protocol: "DCC"})

	_, err := confirmDigitalCenterWriteFixture(t, fixture, preview)
	var conflict *DigitalCenterAddressConflictError
	if !errors.As(err, &conflict) || fixture.ecos.writeCalls != 0 {
		t.Fatalf("error=%v writes=%d", err, fixture.ecos.writeCalls)
	}
}

func TestDigitalCenterWriteVerificationNormalizesTargetedReply(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.ecos.verificationName = "  Neue Lok  "
	fixture.ecos.verificationProtocol = "DCC128"
	target := digitalCenterWriteTarget{
		center: fixture.serviceCenter(), objectID: 3, fields: []string{"name", "protocol"},
		desired: ECoSLocomotiveSyncDesired{Name: "Neue Lok", Protocol: "DCC"},
	}
	locomotive, verified, err := fixture.service.verifyDigitalCenterWrite(t.Context(), target)
	if err != nil || !verified || locomotive.Name != "Neue Lok" || locomotive.Protocol != "DCC" {
		t.Fatalf("locomotive=%#v verified=%v err=%v", locomotive, verified, err)
	}
}

func TestDigitalCenterConfirmedNameWriteKeepsOtherDeviationsVisible(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.ecos.verificationAddress = 99
	preview := previewDigitalCenterWriteFixture(t, fixture, []string{"name"})

	result, err := confirmDigitalCenterWriteFixture(t, fixture, preview)
	if err != nil || result.Result != DigitalCenterWriteVerified || result.WorkItem == nil ||
		result.WorkItem.CompareStatus != DigitalCompareDeviation {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func previewDigitalCenterWriteFixture(
	t *testing.T,
	fixture *digitalCenterWriteFixture,
	fields []string,
) DigitalCenterWritePreview {
	t.Helper()
	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: fields}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	return preview
}

func confirmDigitalCenterWriteFixture(
	t *testing.T,
	fixture *digitalCenterWriteFixture,
	preview DigitalCenterWritePreview,
) (DigitalCenterWriteConfirmation, error) {
	t.Helper()
	return fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWriteConfirmInput{Token: preview.Token, Confirm: true, Fields: preview.Fields}, "admin-1")
}

type digitalCenterWriteFixture struct {
	service    *DigitalCenterWorkspaceService
	repository *digitalCenterWriteRepositoryStub
	ecos       *digitalCenterWriteECoSStub
	vehicles   *digitalCenterWriteVehicleStub
	audit      *digitalCenterWriteAuditStub
	session    DigitalCenterReadSession
	item       DigitalCenterWorkItem
}

func (fixture *digitalCenterWriteFixture) serviceCenter() DigitalCenterSummary {
	return DigitalCenterSummary{Provider: "ecos", Host: fixture.session.Host, Port: fixture.session.Port}
}

func newDigitalCenterWriteFixture(t *testing.T) *digitalCenterWriteFixture {
	t.Helper()
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	session := DigitalCenterReadSession{
		ID: "session-1", Provider: "ecos", State: DigitalCenterSessionReady,
		Host: "trusted-center.local", Port: 15471,
		Capabilities:  DigitalCenterCapabilities{ReadLocomotives: true, WriteLocomotives: true},
		ReadStartedAt: now.Add(-time.Second).Format(time.RFC3339), ReadCompletedAt: now.Format(time.RFC3339),
	}
	item := DigitalCenterWorkItem{
		ID: "item-1", SessionID: session.ID, CenterObjectID: "3", VehicleID: "vehicle-1",
		Name: "Alte Lok", Address: 3, Protocol: "DCC", CompareStatus: DigitalCompareDeviation,
		StationStatus: "read",
		Center:        map[string]any{"objectId": 3, "name": "Alte Lok", "decoderAddress": 3, "protocol": "DCC"},
		RailKeeper:    map[string]any{"vehicleId": "vehicle-1", "name": "Neue Lok", "decoderAddress": 18, "protocol": "DCC"},
		Proposed:      map[string]any{},
		Conflicts:     []map[string]any{},
	}
	repository := &digitalCenterWriteRepositoryStub{session: session, item: item}
	ecos := &digitalCenterWriteECoSStub{currentName: "Alte Lok", verificationName: "Neue Lok"}
	vehicles := &digitalCenterWriteVehicleStub{}
	audit := &digitalCenterWriteAuditStub{}
	service := NewDigitalCenterWorkspaceService(
		repository,
		&workspaceSettingsReaderStub{value: DigitalCenterSettings{
			Provider: "ecos",
			ECoS:     DigitalProviderSettings{Enabled: true, Host: "trusted-center.local", Port: "15471"},
		}},
		ecos,
		nil,
		vehicles,
		audit,
	)
	service.now = func() time.Time { return now }
	return &digitalCenterWriteFixture{
		service: service, repository: repository, ecos: ecos, vehicles: vehicles, audit: audit,
		session: session, item: item,
	}
}

type digitalCenterWriteRepositoryStub struct {
	mu         sync.Mutex
	session    DigitalCenterReadSession
	item       DigitalCenterWorkItem
	grant      DigitalCenterWriteGrant
	consumed   bool
	consumeErr error
	messages   []DigitalCenterSessionMessage
	updateErr  error
}

func (stub *digitalCenterWriteRepositoryStub) CreateSession(
	context.Context, DigitalCenterReadSession,
) (DigitalCenterReadSession, error) {
	return DigitalCenterReadSession{}, errors.New("not used")
}

func (*digitalCenterWriteRepositoryStub) UpdateSession(context.Context, DigitalCenterReadSession) error {
	return errors.New("not used")
}

func (stub *digitalCenterWriteRepositoryStub) GetSession(context.Context, string) (DigitalCenterReadSession, error) {
	return stub.session, nil
}

func (*digitalCenterWriteRepositoryStub) ReplaceWorkItems(context.Context, string, []DigitalCenterWorkItem) error {
	return errors.New("not used")
}

func (stub *digitalCenterWriteRepositoryStub) ListWorkItems(context.Context, string) ([]DigitalCenterWorkItem, error) {
	return []DigitalCenterWorkItem{stub.item}, nil
}

func (stub *digitalCenterWriteRepositoryStub) GetWorkItem(
	context.Context, string, string,
) (DigitalCenterWorkItem, error) {
	return stub.item, nil
}

func (stub *digitalCenterWriteRepositoryStub) UpdateWorkItem(
	_ context.Context,
	item DigitalCenterWorkItem,
) (DigitalCenterWorkItem, error) {
	if stub.updateErr != nil {
		return DigitalCenterWorkItem{}, stub.updateErr
	}
	stub.item = item
	return item, nil
}

func (stub *digitalCenterWriteRepositoryStub) AddMessage(
	_ context.Context,
	message DigitalCenterSessionMessage,
) error {
	stub.messages = append(stub.messages, message)
	return nil
}

func (*digitalCenterWriteRepositoryStub) ListMessages(
	context.Context, string,
) ([]DigitalCenterSessionMessage, error) {
	return []DigitalCenterSessionMessage{}, nil
}

func (stub *digitalCenterWriteRepositoryStub) CreateWriteGrant(_ context.Context, grant DigitalCenterWriteGrant) error {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	stub.grant = grant
	return nil
}

func (stub *digitalCenterWriteRepositoryStub) ConsumeWriteGrant(
	_ context.Context, tokenHash string, actor string,
) (DigitalCenterWriteGrant, error) {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	if stub.consumeErr != nil {
		return DigitalCenterWriteGrant{}, stub.consumeErr
	}
	if tokenHash != stub.grant.TokenHash {
		return DigitalCenterWriteGrant{}, ErrDigitalCenterGrantMismatch
	}
	if actor != stub.grant.ActorUserID {
		return DigitalCenterWriteGrant{}, ErrDigitalCenterGrantActorMismatch
	}
	if stub.consumed {
		return DigitalCenterWriteGrant{}, ErrDigitalCenterGrantConsumed
	}
	stub.consumed = true
	grant := stub.grant
	grant.ConsumedAt = "2026-08-21T10:00:00Z"
	return grant, nil
}

type digitalCenterWriteECoSStub struct {
	currentName          string
	verificationName     string
	verificationProtocol string
	verificationAddress  int
	locomotives          []ECoSLocomotive
	tamperedDesired      string
	syncCalls            []ECoSLocomotiveSyncInput
	confirmCalls         int
	probeCalls           int
	beforeConfirmedWrite func()
	events               []string
	liveStatus           ECoSLiveStatus
	resumeErr            error
	pauseErr             error
	writeErr             error
	writeCalls           int
	readErr              error
	createCalls          []ECoSLocomotiveCreateInput
	createObjectID       int
}

func (stub *digitalCenterWriteECoSStub) ProbeLocomotiveRaw(
	context.Context, ECoSConnectionInput,
) (*ECoSRawProbe, error) {
	stub.probeCalls++
	return &ECoSRawProbe{Locomotives: []ECoSRawLocomotive{{
		ObjectID: 3, Name: stub.verificationName, Address: 18, Protocol: "DCC128",
	}}}, nil
}

func (stub *digitalCenterWriteECoSStub) SyncLocomotive(
	_ context.Context, input ECoSLocomotiveSyncInput,
) (*ECoSLocomotiveSyncResult, error) {
	if input.Confirm {
		stub.events = append(stub.events, "write")
		stub.writeCalls++
		if stub.writeErr != nil {
			return nil, stub.writeErr
		}
	} else {
		stub.events = append(stub.events, "dry-run")
	}
	stub.syncCalls = append(stub.syncCalls, input)
	changes := []ECoSLocomotiveSyncChange{}
	if input.Desired.Name != "" && input.Desired.Name != stub.currentName {
		changes = append(changes, ECoSLocomotiveSyncChange{
			Field: "name", Current: stub.currentName, Desired: input.Desired.Name,
		})
	}
	if input.Desired.Address > 0 && input.Desired.Address != 3 {
		changes = append(changes, ECoSLocomotiveSyncChange{Field: "address", Current: "3", Desired: "18"})
	}
	if input.Desired.Protocol != "" && input.Desired.Protocol != "DCC128" {
		changes = append(changes, ECoSLocomotiveSyncChange{
			Field: "protocol", Current: "DCC128", Desired: input.Desired.Protocol,
		})
	}
	if stub.tamperedDesired != "" && len(changes) > 0 {
		changes[0].Desired = stub.tamperedDesired
	}
	if input.Confirm {
		if stub.beforeConfirmedWrite != nil {
			stub.beforeConfirmedWrite()
		}
		stub.confirmCalls++
		stub.currentName = input.Desired.Name
	}
	return &ECoSLocomotiveSyncResult{
		ObjectID: input.ObjectID, DryRun: input.DryRun || !input.Confirm,
		Applied: input.Confirm && len(changes) > 0, Desired: input.Desired, Changes: changes,
	}, nil
}

func (stub *digitalCenterWriteECoSStub) CreateLocomotive(
	_ context.Context, input ECoSLocomotiveCreateInput,
) (*ECoSLocomotiveCreateResult, error) {
	stub.events = append(stub.events, "create")
	stub.createCalls = append(stub.createCalls, input)
	if stub.beforeConfirmedWrite != nil {
		stub.beforeConfirmedWrite()
	}
	if stub.writeErr != nil {
		return nil, stub.writeErr
	}
	return &ECoSLocomotiveCreateResult{
		ObjectID: stub.createObjectID, Desired: input.Desired, Applied: input.Confirm,
	}, nil
}

func (stub *digitalCenterWriteECoSStub) LiveStatus() ECoSLiveStatus {
	return stub.liveStatus
}

func (stub *digitalCenterWriteECoSStub) StopLive() ECoSLiveStatus {
	stub.liveStatus.State = ECoSLiveStopped
	stub.liveStatus.Connected = false
	return stub.liveStatus
}

func (stub *digitalCenterWriteECoSStub) PauseLive(context.Context) (ECoSLiveStatus, error) {
	stub.events = append(stub.events, "pause")
	if stub.pauseErr != nil {
		return stub.liveStatus, stub.pauseErr
	}
	return stub.StopLive(), nil
}

func (stub *digitalCenterWriteECoSStub) StartLive(
	context.Context,
	ECoSConnectionInput,
) (*ECoSLiveStatus, error) {
	stub.events = append(stub.events, "resume")
	if stub.resumeErr != nil {
		return nil, stub.resumeErr
	}
	stub.liveStatus.State = ECoSLiveRunning
	stub.liveStatus.Connected = true
	return &stub.liveStatus, nil
}

func (stub *digitalCenterWriteECoSStub) StartLiveWithInterruption(
	ctx context.Context,
	input ECoSConnectionInput,
	_ func(),
) (*ECoSLiveStatus, error) {
	return stub.StartLive(ctx, input)
}

func (stub *digitalCenterWriteECoSStub) ReadLocomotive(
	_ context.Context,
	_ ECoSConnectionInput,
	objectID int,
) (ECoSLocomotive, error) {
	stub.events = append(stub.events, "read-target")
	if stub.readErr != nil {
		return ECoSLocomotive{}, stub.readErr
	}
	protocol := stub.verificationProtocol
	if protocol == "" {
		protocol = "DCC"
	}
	address := stub.verificationAddress
	if address == 0 {
		address = 18
	}
	return ECoSLocomotive{
		ObjectID: objectID, Name: stub.verificationName, Address: address, Protocol: protocol,
	}, nil
}

type digitalCenterWriteVehicleStub struct {
	mapping            *VehicleExternalMapInput
	previousExternalID string
	err                error
}

func (*digitalCenterWriteVehicleStub) ListReadOnly(context.Context, string) ([]Vehicle, error) {
	return []Vehicle{}, nil
}

func (stub *digitalCenterWriteVehicleStub) UpsertExternalMapping(
	_ context.Context, _ string, input VehicleExternalMapInput, _ string,
) (*VehicleExternalMap, error) {
	if stub.err != nil {
		return nil, stub.err
	}
	copy := input
	stub.mapping = &copy
	return &VehicleExternalMap{Provider: input.Provider, ExternalID: input.ExternalID}, nil
}

func (stub *digitalCenterWriteVehicleStub) RebindExternalMapping(
	_ context.Context, _ string, previousExternalID string, input VehicleExternalMapInput, _ string,
) (*VehicleExternalMap, error) {
	stub.previousExternalID = previousExternalID
	return stub.UpsertExternalMapping(context.Background(), "", input, "")
}

type digitalCenterWriteAuditEntry struct {
	actor, action, targetType, targetID, details string
}

type digitalCenterWriteAuditStub struct {
	entries []digitalCenterWriteAuditEntry
	err     error
}

func (stub *digitalCenterWriteAuditStub) RecordAudit(
	_ context.Context, actor, action, targetType, targetID, details string,
) error {
	stub.entries = append(stub.entries, digitalCenterWriteAuditEntry{
		actor: actor, action: action, targetType: targetType, targetID: targetID, details: details,
	})
	return stub.err
}
