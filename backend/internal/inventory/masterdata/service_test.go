package masterdata

import (
	"context"
	"errors"
	"testing"
)

type mutationRepository struct {
	Repository
	entry                    MasterDataEntry
	loads, writes            int
	writeFailure, getFailure error
}

func (r *mutationRepository) Snapshot(context.Context) (map[string][]MasterDataEntry, error) {
	r.loads++
	return map[string][]MasterDataEntry{r.entry.Type: {r.entry}}, nil
}
func (r *mutationRepository) Update(_ context.Context, mutation Mutation) error {
	if r.writeFailure != nil {
		return r.writeFailure
	}
	r.writes++
	r.entry = mutation.Entry
	return nil
}
func (r *mutationRepository) Get(context.Context, string, string) (*MasterDataEntry, error) {
	if r.getFailure != nil {
		return nil, r.getFailure
	}
	return &r.entry, nil
}

func TestCommittedMasterDataWriteInvalidatesCacheEvenWhenReadbackFails(t *testing.T) {
	sentinel := errors.New("synthetic readback failure")
	repository := &mutationRepository{entry: MasterDataEntry{ID: "gauge:recorded", Type: "gauge", Key: "recorded", Label: "Before", Active: true}}
	service := New(repository, nil)
	if err := service.WarmCache(t.Context()); err != nil {
		t.Fatal(err)
	}
	repository.getFailure = sentinel
	value, err := service.Update(t.Context(), "gauge", "recorded", MasterDataInput{Label: "After"})
	if !errors.Is(err, sentinel) || value != nil || repository.writes != 1 {
		t.Fatal("readback failure was acknowledged or lost the committed write", value, err)
	}
	current, err := service.ListAll(t.Context(), false)
	if err != nil || repository.loads != 2 || current["gauge"][0].Label != "After" {
		t.Fatal("readback failure left the previous cached record visible", current, err)
	}
}

func TestRejectedMasterDataWriteKeepsExistingCacheAvailable(t *testing.T) {
	sentinel := errors.New("synthetic write failure")
	repository := &mutationRepository{entry: MasterDataEntry{ID: "gauge:recorded", Type: "gauge", Key: "recorded", Label: "Before", Active: true}, writeFailure: sentinel}
	service := New(repository, nil)
	if err := service.WarmCache(t.Context()); err != nil {
		t.Fatal(err)
	}
	if value, err := service.Update(t.Context(), "gauge", "recorded", MasterDataInput{Label: "After"}); !errors.Is(err, sentinel) || value != nil || repository.writes != 0 {
		t.Fatal("rejected write changed or acknowledged a record", value, err)
	}
	current, err := service.ListAll(t.Context(), false)
	if err != nil || repository.loads != 1 || current["gauge"][0].Label != "Before" {
		t.Fatal("rejected write discarded or changed the valid previous cache", current, err)
	}
}

func TestReconciliationRetainsBundledIdentityAndRejectsOmittedReferences(t *testing.T) {
	bundled := MasterDataEntry{ID: "original-bundled-id", Type: "gauge", Key: "bundled", Label: "Recorded bundled", Origin: MasterDataOriginBundled}
	custom := MasterDataEntry{ID: "original-custom-id", Type: "manufacturer", Key: "custom", Label: "Recorded custom", Origin: MasterDataOriginCustom}
	current := map[Identity]MasterDataEntry{IdentityOf(bundled): bundled, IdentityOf(custom): custom}
	for _, referenced := range []bool{false, true} {
		calls := 0
		desired, _, err := BuildDesiredEntries(t.Context(), func(_ context.Context, entry MasterDataEntry) (bool, error) {
			calls++
			if entry.ID != custom.ID {
				t.Fatal("bundled record was tested as a deletion candidate")
			}
			return referenced, nil
		}, current, nil, "2026-11-08T13:00:00Z", false, false)
		if calls != 1 {
			t.Fatal("omitted custom identity did not receive exactly one reference check", calls)
		}
		if referenced {
			if !errors.Is(err, ErrMasterDataInUse) || desired != nil {
				t.Fatal("referenced omission produced an applicable import plan", desired, err)
			}
		} else if err != nil || len(desired) != 1 || desired[IdentityOf(bundled)].ID != bundled.ID {
			t.Fatal("reconciliation changed an omitted bundled identity", desired, err)
		}
	}
}

func TestRelationReconciliationRetainsBundledLinksAndRejectsMissingEndpoints(t *testing.T) {
	parent := MasterDataEntry{Type: "category", Key: "parent", Origin: MasterDataOriginBundled}
	child := MasterDataEntry{Type: "gattung", Key: "child", Origin: MasterDataOriginBundled}
	current := map[Identity]MasterDataEntry{IdentityOf(parent): parent, IdentityOf(child): child}
	relation := MasterDataRelation{ID: "recorded-relation", ParentType: parent.Type, ParentKey: parent.Key, ChildType: child.Type, ChildKey: child.Key}
	result, count, err := BuildDesiredRelations(nil, []MasterDataRelation{relation}, current, current)
	if err != nil || count != 0 || len(result) != 1 || result[0].ID != relation.ID {
		t.Fatal("bundled relation identity was lost", result, err)
	}
	missing := map[Identity]MasterDataEntry{IdentityOf(parent): parent}
	if result, _, err := BuildDesiredRelations([]MasterDataRelation{relation}, nil, current, missing); !errors.Is(err, ErrMasterDataValidation) || result != nil {
		t.Fatal("missing relation endpoint was accepted", result, err)
	}
}
