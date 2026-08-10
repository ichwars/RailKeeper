package application

import (
	"context"
	"errors"
	"math"
	"testing"

	"railkeeper/backend/internal/domain"
)

type trackPlannerRepositorySpy struct {
	listedGauge     string
	createdRevision string
	created         CreatePlanTrackObjectInput
	updated         UpdatePlanTrackObjectInput
	deletedVersion  int
	plan            *TrackPlan
	plans           map[string]*TrackPlan
	planForObject   *TrackPlan
	materials       []TrackMaterialStatus
	reservations    []TrackPlanObjectReservation
	baseRevisionID  string
	affected        []TrackPlanAffectedConfiguration
	reserved        ReserveTrackPlanMaterialsInput
}

func (spy *trackPlannerRepositorySpy) ListGeometries(
	_ context.Context,
	gauge string,
) ([]domain.TrackGeometryDefinition, error) {
	spy.listedGauge = gauge
	return []domain.TrackGeometryDefinition{}, nil
}

func (spy *trackPlannerRepositorySpy) GetPlan(_ context.Context, revisionID string) (*TrackPlan, error) {
	if plan := spy.plans[revisionID]; plan != nil {
		return plan, nil
	}
	if spy.plan != nil {
		return spy.plan, nil
	}
	return &TrackPlan{RevisionID: revisionID, Objects: []domain.PlanTrackObject{}}, nil
}

func (spy *trackPlannerRepositorySpy) GetBaseRevisionID(_ context.Context, _ string) (string, error) {
	return spy.baseRevisionID, nil
}

func (spy *trackPlannerRepositorySpy) ListAffectedConfigurations(
	_ context.Context,
	_ string,
) ([]TrackPlanAffectedConfiguration, error) {
	return spy.affected, nil
}

func (spy *trackPlannerRepositorySpy) ReserveMaterials(
	_ context.Context,
	revisionID string,
	input ReserveTrackPlanMaterialsInput,
	_ string,
) (*TrackPlanReservationBatch, error) {
	spy.reserved = input
	return &TrackPlanReservationBatch{RevisionID: revisionID, Reservations: []TrackPlanObjectReservation{}}, nil
}

func (spy *trackPlannerRepositorySpy) GetPlanForObject(_ context.Context, _ string) (*TrackPlan, error) {
	if spy.planForObject == nil {
		return nil, ErrTrackPlanNotFound
	}
	return spy.planForObject, nil
}

func (spy *trackPlannerRepositorySpy) TrackMaterialAvailability(
	_ context.Context,
	_ []domain.TrackBOMLine,
) ([]TrackMaterialStatus, error) {
	return spy.materials, nil
}

func (spy *trackPlannerRepositorySpy) ListMaterialReservations(
	_ context.Context,
	_ string,
) ([]TrackPlanObjectReservation, error) {
	return spy.reservations, nil
}

func (spy *trackPlannerRepositorySpy) CreateObject(
	_ context.Context,
	revisionID string,
	input CreatePlanTrackObjectInput,
	_ string,
) (*domain.PlanTrackObject, error) {
	spy.createdRevision = revisionID
	spy.created = input
	return &domain.PlanTrackObject{
		RevisionID: revisionID, GeometryID: input.GeometryID,
		PositionXMM: input.PositionXMM, PositionYMM: input.PositionYMM,
		RotationDegrees: input.RotationDegrees, ElevationStartMM: input.ElevationStartMM,
		ElevationEndMM: input.ElevationEndMM,
	}, nil
}

func (spy *trackPlannerRepositorySpy) UpdateObject(
	_ context.Context,
	_ string,
	input UpdatePlanTrackObjectInput,
	_ string,
) (*domain.PlanTrackObject, error) {
	spy.updated = input
	return &domain.PlanTrackObject{
		PositionXMM: input.PositionXMM, PositionYMM: input.PositionYMM,
		RotationDegrees: input.RotationDegrees, ElevationStartMM: input.ElevationStartMM,
		ElevationEndMM: input.ElevationEndMM,
	}, nil
}

func (spy *trackPlannerRepositorySpy) DeleteObject(
	_ context.Context,
	_ string,
	expectedVersion int,
	_ string,
) error {
	spy.deletedVersion = expectedVersion
	return nil
}

func TestTrackPlannerServiceNormalizesInputs(t *testing.T) {
	repository := &trackPlannerRepositorySpy{}
	service := NewTrackPlannerService(repository)

	if _, err := service.ListGeometries(t.Context(), " TT "); err != nil {
		t.Fatal(err)
	}
	if repository.listedGauge != "TT" {
		t.Fatalf("unexpected normalized gauge: %q", repository.listedGauge)
	}

	created, err := service.CreateObject(t.Context(), " revision-1 ", CreatePlanTrackObjectInput{
		GeometryID: " geometry-1 ", PositionXMM: 20, PositionYMM: 30, RotationDegrees: -15,
		ElevationStartMM: -2.5, ElevationEndMM: 4.5,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if repository.createdRevision != "revision-1" || repository.created.GeometryID != "geometry-1" ||
		created.RotationDegrees != 345 || created.ElevationStartMM != -2.5 || created.ElevationEndMM != 4.5 {
		t.Fatalf("unexpected normalized create: revision=%q input=%#v object=%#v",
			repository.createdRevision, repository.created, created)
	}
	repository.planForObject = &TrackPlan{RevisionID: "revision-1", Status: domain.PlanRevisionDraft,
		Objects: []domain.PlanTrackObject{trackPlannerTestG1("object-1", 25, 35, 15)}}

	updated, err := service.UpdateObject(t.Context(), " object-1 ", UpdatePlanTrackObjectInput{
		PositionXMM: 25, PositionYMM: 35, RotationDegrees: 735, ElevationStartMM: 3,
		ElevationEndMM: 7.15, ExpectedVersion: 2,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.RotationDegrees != 15 || repository.updated.ExpectedVersion != 2 ||
		updated.ElevationStartMM != 3 || updated.ElevationEndMM != 7.15 {
		t.Fatalf("unexpected normalized update: %#v", repository.updated)
	}
}

func TestTrackPlannerServiceRejectsInvalidInputs(t *testing.T) {
	service := NewTrackPlannerService(&trackPlannerRepositorySpy{})

	createInputs := []CreatePlanTrackObjectInput{
		{},
		{GeometryID: "geometry-1", PositionXMM: math.NaN()},
		{GeometryID: "geometry-1", PositionYMM: math.Inf(1)},
		{GeometryID: "geometry-1", RotationDegrees: math.NaN()},
		{GeometryID: "geometry-1", ElevationStartMM: math.Inf(-1)},
		{GeometryID: "geometry-1", ElevationEndMM: math.NaN()},
	}
	for _, input := range createInputs {
		if _, err := service.CreateObject(t.Context(), "revision-1", input, "planner"); !errors.Is(err, ErrTrackPlanValidation) {
			t.Fatalf("expected create validation error for %#v, got %v", input, err)
		}
	}

	if _, err := service.GetPlan(t.Context(), " "); !errors.Is(err, ErrTrackPlanValidation) {
		t.Fatalf("expected plan validation error, got %v", err)
	}
	if _, err := service.UpdateObject(t.Context(), "object-1", UpdatePlanTrackObjectInput{}, "planner"); !errors.Is(err, ErrTrackPlanValidation) {
		t.Fatalf("expected update validation error, got %v", err)
	}
	for _, input := range []UpdatePlanTrackObjectInput{
		{ExpectedVersion: 1, ElevationStartMM: math.NaN()},
		{ExpectedVersion: 1, ElevationEndMM: math.Inf(1)},
	} {
		if _, err := service.UpdateObject(t.Context(), "object-1", input, "planner"); !errors.Is(err, ErrTrackPlanValidation) {
			t.Fatalf("expected update validation error for %#v, got %v", input, err)
		}
	}
	if err := service.DeleteObject(t.Context(), "object-1", 0, "planner"); !errors.Is(err, ErrTrackPlanValidation) {
		t.Fatalf("expected delete validation error, got %v", err)
	}
}

func TestTrackPlannerServicePreviewsFlexPathWithoutMutation(t *testing.T) {
	flex := domain.PlanTrackObject{
		ID: "flex-1", Version: 3,
		Geometry: domain.TrackGeometryDefinition{
			ID: "flex-definition", Kind: domain.TrackGeometryFlex, LengthMM: 664,
			MinimumRadiusMM: float64Pointer(543),
		},
	}
	repository := &trackPlannerRepositorySpy{planForObject: &TrackPlan{
		Objects: []domain.PlanTrackObject{flex},
		Limits:  domain.TrackPlanLimits{MinimumFlexRadiusMM: float64Pointer(700)},
	}}
	service := NewTrackPlannerService(repository)
	preview, err := service.PreviewFlexPath(t.Context(), " flex-1 ", FlexTrackPreviewInput{
		EndXMM: 500, EndYMM: 100, EndDirectionDegrees: 20, ExpectedVersion: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Path.EndXMM != 500 || preview.EffectiveLengthMM <= 0 ||
		preview.RadiusLimitMM != 700 || !preview.Applicable {
		t.Fatalf("unexpected flex preview: %#v", preview)
	}
	if repository.updated.ExpectedVersion != 0 {
		t.Fatalf("preview mutated repository: %#v", repository.updated)
	}
	if _, err := service.PreviewFlexPath(t.Context(), "flex-1", FlexTrackPreviewInput{
		EndXMM: 500, EndYMM: 100, ExpectedVersion: 2,
	}); !errors.Is(err, ErrTrackPlanConflict) {
		t.Fatalf("expected preview conflict, got %v", err)
	}
}

func float64Pointer(value float64) *float64 { return &value }

func TestTrackPlannerSnapAppliesNearestCompatiblePose(t *testing.T) {
	moving := trackPlannerTestG1("moving", 172, 2, 2)
	target := trackPlannerTestG1("target", 0, 0, 0)
	repository := &trackPlannerRepositorySpy{planForObject: &TrackPlan{
		RevisionID: "revision-1", Status: domain.PlanRevisionDraft,
		Objects: []domain.PlanTrackObject{target, moving},
	}}
	service := NewTrackPlannerService(repository)

	updated, err := service.UpdateObject(t.Context(), moving.ID, UpdatePlanTrackObjectInput{
		PositionXMM: moving.PositionXMM, PositionYMM: moving.PositionYMM,
		RotationDegrees: moving.RotationDegrees, ExpectedVersion: 1,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.PositionXMM != 166 || updated.PositionYMM != 0 || updated.RotationDegrees != 0 {
		t.Fatalf("expected authoritative snapped pose, got %#v", updated)
	}
}

func TestTrackPlannerSnapLeavesPoseOutsideTolerance(t *testing.T) {
	moving := trackPlannerTestG1("moving", 174.01, 0, 0)
	target := trackPlannerTestG1("target", 0, 0, 0)
	repository := &trackPlannerRepositorySpy{planForObject: &TrackPlan{
		RevisionID: "revision-1", Status: domain.PlanRevisionDraft,
		Objects: []domain.PlanTrackObject{target, moving},
	}}
	service := NewTrackPlannerService(repository)

	updated, err := service.UpdateObject(t.Context(), moving.ID, UpdatePlanTrackObjectInput{
		PositionXMM: moving.PositionXMM, PositionYMM: moving.PositionYMM, ExpectedVersion: 1,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.PositionXMM != 174.01 {
		t.Fatalf("pose outside tolerance changed: %#v", updated)
	}
}

func TestTrackPlannerAnalyzePlanCombinesGeometryAndMaterials(t *testing.T) {
	objects := []domain.PlanTrackObject{
		trackPlannerTestG1("track-1", 0, 0, 0),
		trackPlannerTestG1("track-2", 166, 0, 0),
	}
	objects[0].ElevationEndMM = 4.15
	limit := 2.0
	repository := &trackPlannerRepositorySpy{
		materials: []TrackMaterialStatus{{
			GeometryID: "tillig-g1", Manufacturer: "Tillig", ArticleNumber: "83101",
			RequiredQuantity: 2, PhysicalQuantity: 3, ReservedQuantity: 1,
			AvailableQuantity: 2, MissingQuantity: 0,
		}},
		reservations: []TrackPlanObjectReservation{{TrackObjectID: "track-1"}},
	}
	repositoryPlan := &TrackPlan{RevisionID: "revision-1", Status: domain.PlanRevisionDraft, Objects: objects,
		Limits: domain.TrackPlanLimits{MaxGradePercent: &limit}}
	repository.plan = repositoryPlan
	service := NewTrackPlannerService(repository)
	analysis, err := service.AnalyzePlan(t.Context(), repositoryPlan.RevisionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Connections) != 1 || len(analysis.BOM) != 1 || len(analysis.Grades) != 2 ||
		analysis.Grades[0].GradePercent != 2.5 || len(analysis.Materials) != 1 ||
		len(analysis.Reservations) != 1 ||
		analysis.Materials[0].AvailableQuantity != 2 {
		t.Fatalf("unexpected combined analysis: %#v", analysis)
	}
	gradeIssues := 0
	for _, issue := range analysis.Issues {
		if issue.Code == domain.TrackPlanIssueGradeLimitExceeded {
			gradeIssues++
		}
	}
	if gradeIssues != 1 {
		t.Fatalf("expected one configured grade limit warning, got %#v", analysis.Issues)
	}
}

func TestTrackPlannerChangePreviewComparesBaseRevisionAndAffectedConfigurations(t *testing.T) {
	base := trackPlannerTestG1("base-object", 0, 0, 0)
	base.LineageID = "lineage-1"
	current := trackPlannerTestG1("current-object", 10, 0, 0)
	current.LineageID = base.LineageID
	added := trackPlannerTestG1("added-object", 176, 0, 0)
	added.LineageID = "lineage-2"
	repository := &trackPlannerRepositorySpy{
		baseRevisionID: "revision-base",
		plans: map[string]*TrackPlan{
			"revision-base": {RevisionID: "revision-base", Status: domain.PlanRevisionPublished,
				Objects: []domain.PlanTrackObject{base}},
			"revision-current": {RevisionID: "revision-current", Status: domain.PlanRevisionDraft,
				Objects: []domain.PlanTrackObject{current, added}},
		},
		affected: []TrackPlanAffectedConfiguration{{ID: "configuration-1", Name: "Ausstellung"}},
	}

	preview, err := NewTrackPlannerService(repository).ChangePreview(t.Context(), " revision-current ")
	if err != nil {
		t.Fatal(err)
	}
	if preview.RevisionID != "revision-current" || preview.BaseRevisionID != "revision-base" ||
		len(preview.ObjectChanges) != 2 || len(preview.MaterialDeltas) != 1 ||
		preview.MaterialDeltas[0].Delta != 1 || len(preview.Issues.Added) != 1 ||
		len(preview.Issues.Resolved) != 1 || len(preview.AffectedConfigurations) != 1 ||
		preview.AffectedConfigurations[0].Name != "Ausstellung" {
		t.Fatalf("unexpected track plan change preview: %#v", preview)
	}
}

func TestTrackPlannerChangePreviewWithoutBaseTreatsObjectsAsAdded(t *testing.T) {
	current := trackPlannerTestG1("current-object", 0, 0, 0)
	current.LineageID = "lineage-1"
	repository := &trackPlannerRepositorySpy{plans: map[string]*TrackPlan{
		"revision-current": {RevisionID: "revision-current", Status: domain.PlanRevisionDraft,
			Objects: []domain.PlanTrackObject{current}},
	}}

	preview, err := NewTrackPlannerService(repository).ChangePreview(t.Context(), "revision-current")
	if err != nil {
		t.Fatal(err)
	}
	if preview.BaseRevisionID != "" || len(preview.ObjectChanges) != 1 ||
		preview.ObjectChanges[0].Type != domain.TrackPlanObjectAdded ||
		len(preview.AffectedConfigurations) != 0 {
		t.Fatalf("unexpected first-revision preview: %#v", preview)
	}
}

func TestTrackPlannerChangePreviewUsesCurrentLayoutGradeLimit(t *testing.T) {
	base := trackPlannerTestG1("base-object", 0, 0, 0)
	base.LineageID = "lineage-1"
	base.ElevationEndMM = 6.64
	current := trackPlannerTestG1("current-object", 0, 0, 0)
	current.LineageID = base.LineageID
	current.ElevationEndMM = 4.98
	limit := 3.0
	repository := &trackPlannerRepositorySpy{
		baseRevisionID: "revision-base",
		plans: map[string]*TrackPlan{
			"revision-base": {
				RevisionID: "revision-base", Status: domain.PlanRevisionPublished,
				Objects: []domain.PlanTrackObject{base},
			},
			"revision-current": {
				RevisionID: "revision-current", Status: domain.PlanRevisionDraft,
				Objects: []domain.PlanTrackObject{current},
				Limits:  domain.TrackPlanLimits{MaxGradePercent: &limit},
			},
		},
	}

	preview, err := NewTrackPlannerService(repository).ChangePreview(t.Context(), "revision-current")
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range preview.Issues.Resolved {
		if issue.Code == domain.TrackPlanIssueGradeLimitExceeded {
			return
		}
	}
	t.Fatalf("expected resolved grade limit warning, got %#v", preview.Issues)
}

func TestTrackPlannerChangePreviewUsesCurrentLayoutClearanceLimit(t *testing.T) {
	baseLower := trackPlannerTestG1("base-lower", 0, 0, 0)
	baseLower.LineageID = "lineage-lower"
	baseUpper := trackPlannerTestG1("base-upper", 83, -83, 90)
	baseUpper.LineageID = "lineage-upper"
	baseUpper.ElevationStartMM, baseUpper.ElevationEndMM = 25, 25
	currentLower := trackPlannerTestG1("current-lower", 0, 0, 0)
	currentLower.LineageID = baseLower.LineageID
	currentUpper := trackPlannerTestG1("current-upper", 83, -83, 90)
	currentUpper.LineageID = baseUpper.LineageID
	currentUpper.ElevationStartMM, currentUpper.ElevationEndMM = 40, 40
	limit := 40.0
	repository := &trackPlannerRepositorySpy{
		baseRevisionID: "revision-base",
		plans: map[string]*TrackPlan{
			"revision-base": {
				RevisionID: "revision-base", Status: domain.PlanRevisionPublished,
				Objects: []domain.PlanTrackObject{baseLower, baseUpper},
			},
			"revision-current": {
				RevisionID: "revision-current", Status: domain.PlanRevisionDraft,
				Objects: []domain.PlanTrackObject{currentLower, currentUpper},
				Limits:  domain.TrackPlanLimits{MinimumTrackClearanceMM: &limit},
			},
		},
	}

	preview, err := NewTrackPlannerService(repository).ChangePreview(t.Context(), "revision-current")
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range preview.Issues.Resolved {
		if issue.Code == domain.TrackPlanIssueInsufficientClearance {
			return
		}
	}
	t.Fatalf("expected resolved clearance warning, got %#v", preview.Issues)
}

func TestTrackPlannerReserveMaterialsRequiresConfirmationAndValidUniqueObjects(t *testing.T) {
	service := NewTrackPlannerService(&trackPlannerRepositorySpy{})
	valid := ReserveTrackPlanMaterialsInput{Confirmed: true, Items: []TrackPlanReservationInput{{
		TrackObjectID: "object-1", ProductID: "product-1", LocationID: "location-1",
		ExpectedObjectVersion: 2,
	}}}
	invalid := []ReserveTrackPlanMaterialsInput{
		{},
		{Confirmed: true},
		{Confirmed: true, Items: []TrackPlanReservationInput{{TrackObjectID: "object-1"}}},
		{Confirmed: true, Items: []TrackPlanReservationInput{
			valid.Items[0], valid.Items[0],
		}},
	}
	for _, input := range invalid {
		if _, err := service.ReserveMaterials(t.Context(), "revision-1", input, "planner"); !errors.Is(err, ErrTrackPlanValidation) {
			t.Fatalf("expected reservation validation error for %#v, got %v", input, err)
		}
	}
	if _, err := service.ReserveMaterials(t.Context(), " ", valid, "planner"); !errors.Is(err, ErrTrackPlanValidation) {
		t.Fatalf("expected revision validation error, got %v", err)
	}
}

func TestTrackPlannerReserveMaterialsNormalizesBatch(t *testing.T) {
	repository := &trackPlannerRepositorySpy{materials: []TrackMaterialStatus{{GeometryID: "geometry-1"}}}
	service := NewTrackPlannerService(repository)
	batch, err := service.ReserveMaterials(t.Context(), " revision-1 ", ReserveTrackPlanMaterialsInput{
		Confirmed: true,
		Items: []TrackPlanReservationInput{{
			TrackObjectID: " object-1 ", ProductID: " product-1 ", LocationID: " location-1 ",
			AssetID: " asset-1 ", ExpectedObjectVersion: 2,
		}},
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if batch.RevisionID != "revision-1" || len(batch.Materials) != 1 ||
		repository.reserved.Items[0].TrackObjectID != "object-1" ||
		repository.reserved.Items[0].AssetID != "asset-1" {
		t.Fatalf("unexpected normalized reservation batch: %#v %#v", batch, repository.reserved)
	}
}

func trackPlannerTestG1(id string, x, y, rotation float64) domain.PlanTrackObject {
	return domain.PlanTrackObject{
		ID: id, GeometryID: "tillig-g1", PositionXMM: x, PositionYMM: y, RotationDegrees: rotation,
		Version: 1,
		Geometry: domain.TrackGeometryDefinition{
			ID: "tillig-g1", LibraryID: "tillig-v1", ArticleNumber: "83101", Name: "Gleisstück G1",
			Kind: domain.TrackGeometryStraight, LengthMM: 166, Status: domain.TrackGeometryVerified,
			Geometry: domain.TrackGeometry{SchemaVersion: 1,
				Ports: []domain.TrackPort{
					{ID: "a", DirectionDegrees: 180}, {ID: "b", XMM: 166, DirectionDegrees: 0},
				},
				Routes: []domain.TrackRoute{{ID: "main", Points: []domain.TrackPoint{{}, {XMM: 166}}}},
			},
		},
	}
}
