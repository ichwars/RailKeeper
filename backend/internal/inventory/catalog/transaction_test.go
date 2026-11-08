package catalog

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mutationProbe struct {
	before                                                *VehicleCatalogContext
	failAt                                                string
	contextReads, projections, audits, commits, rollbacks int
}

var errStorage = errors.New("synthetic catalog storage failure")

func (p *mutationProbe) fail(step string) error {
	if p.failAt == step {
		return errStorage
	}
	return nil
}
func (p *mutationProbe) Begin(context.Context) (Mutation, error) { return p, p.fail("begin") }
func (p *mutationProbe) Context(context.Context, string, string) (*VehicleCatalogContext, error) {
	p.contextReads++
	if err := p.fail("read"); err != nil {
		return nil, err
	}
	if p.contextReads > 1 {
		if err := p.fail("readback"); err != nil {
			return nil, err
		}
	}
	return p.before, nil
}
func (p *mutationProbe) RevisionExists(context.Context, string) (bool, error) {
	return p.failAt == "duplicate", p.fail("exists")
}
func (p *mutationProbe) CreateArticle(context.Context, string, string) error {
	return p.fail("article")
}
func (p *mutationProbe) InsertRevision(_ context.Context, id, article string, fields VehicleCatalogFields, note, now string) (*VehicleCatalogRevision, error) {
	return &VehicleCatalogRevision{ID: id, ArticleID: article, Fields: fields, Note: note, CreatedAt: now}, p.fail("revision")
}
func (p *mutationProbe) Project(context.Context, *VehicleCatalogContext, *VehicleCatalogRevision, Projection, string) error {
	p.projections++
	return p.fail("projection")
}
func (p *mutationProbe) Audit(context.Context, *VehicleCatalogContext, string, string, string, string) error {
	p.audits++
	return p.fail("audit")
}
func (p *mutationProbe) Report(context.Context) ([]RawAssignment, error) {
	return nil, p.fail("report")
}
func (p *mutationProbe) LinkCount(context.Context) (int, error) { return 0, p.fail("count") }
func (p *mutationProbe) Backfill(context.Context) error         { return p.fail("backfill") }
func (p *mutationProbe) AuditBackfill(context.Context, int, string, string) error {
	return p.fail("backfill-audit")
}
func (p *mutationProbe) Commit() error   { p.commits++; return p.fail("commit") }
func (p *mutationProbe) Rollback() error { p.rollbacks++; return nil }

func TestCatalogMutationDoesNotAcknowledgeAnyFailedStep(t *testing.T) {
	for _, step := range []string{"begin", "read", "exists", "article", "revision", "projection", "audit", "readback", "commit"} {
		t.Run(step, func(t *testing.T) {
			probe := &mutationProbe{failAt: step, before: &VehicleCatalogContext{Revision: strings.Repeat("a", 64), CurrentFields: recordedFields()}}
			value, err := New(probe).CreateCatalogRevision(t.Context(), "vehicle", "vehicle", VehicleCatalogRevisionInput{
				ID: strings.Repeat("1", 32), ExpectedRevision: probe.before.Revision, Fields: recordedFields(), SelectedFields: []string{}, Independent: true,
			}, "actor")
			if !errors.Is(err, errStorage) || value != nil {
				t.Fatal("failed mutation acknowledged", value, err)
			}
			if step != "begin" && probe.rollbacks != 1 {
				t.Fatal("mutation not rolled back", probe)
			}
			if step != "commit" && probe.commits != 0 {
				t.Fatal("failed mutation committed", probe)
			}
		})
	}
}

func TestAdoptingAlreadySelectedRevisionMakesNoWrites(t *testing.T) {
	chosen := VehicleCatalogRevision{ID: "chosen", Fields: recordedFields()}
	probe := &mutationProbe{before: &VehicleCatalogContext{Revision: strings.Repeat("a", 64), Authority: "catalog", Pinned: &chosen, Revisions: []VehicleCatalogRevision{chosen}, CurrentFields: recordedFields()}}
	value, err := New(probe).AdoptCatalogRevision(t.Context(), "vehicle", "vehicle", VehicleCatalogAdoptionInput{
		ExpectedRevision: probe.before.Revision, RevisionID: chosen.ID, SelectedFields: []string{"name"},
	}, "actor")
	if err != nil || value != probe.before || probe.projections != 0 || probe.audits != 0 || probe.commits != 0 || probe.rollbacks != 1 {
		t.Fatal("repeated adoption changed data", value, err, probe)
	}
}
