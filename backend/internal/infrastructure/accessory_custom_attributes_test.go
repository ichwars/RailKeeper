package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

type blockingAccessoryProductMutationRepository struct {
	application.AccessoryRepository
	validationStarted chan struct{}
	releaseValidation chan struct{}
}

func (r *blockingAccessoryProductMutationRepository) CreateProduct(
	ctx context.Context,
	input application.CreateAccessoryProductInput,
	actor string,
	validate application.AccessoryProductMutationValidator,
) (*application.AccessoryProduct, error) {
	return r.AccessoryRepository.CreateProduct(ctx, input, actor, func(state application.AccessoryProductMutationState) error {
		close(r.validationStarted)
		select {
		case <-r.releaseValidation:
		case <-ctx.Done():
			return ctx.Err()
		}
		return validate(state)
	})
}

func (r *blockingAccessoryProductMutationRepository) UpdateProduct(
	ctx context.Context,
	id string,
	input application.UpdateAccessoryProductInput,
	actor string,
	validate application.AccessoryProductMutationValidator,
) (*application.AccessoryProduct, error) {
	return r.AccessoryRepository.UpdateProduct(ctx, id, input, actor, func(state application.AccessoryProductMutationState) error {
		close(r.validationStarted)
		select {
		case <-r.releaseValidation:
		case <-ctx.Done():
			return ctx.Err()
		}
		return validate(state)
	})
}

func TestAccessoryCustomFieldDefinitionChangeCannotRaceProductCreate(t *testing.T) {
	_, db := testAccessoryService(t)
	db.SetMaxOpenConns(4)
	masterData := application.NewMasterDataService(db)
	active := true
	if _, err := masterData.Create(t.Context(), "accessory_custom_field", application.MasterDataInput{
		Key: "length", Label: "Length", Active: &active,
		Metadata: map[string]any{"kind": "number", "unit": "mm"},
	}); err != nil {
		t.Fatal(err)
	}

	repository := &blockingAccessoryProductMutationRepository{
		AccessoryRepository: infrastructure.NewAccessoryRepository(db),
		validationStarted:   make(chan struct{}),
		releaseValidation:   make(chan struct{}),
	}
	accessories := application.NewAccessoryService(repository)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	value := 12.5
	unit := "mm"
	createResult := make(chan error, 1)
	go func() {
		_, err := accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
			Manufacturer: "Club", Name: "Measured article", Category: "other",
			ArticleType: domain.AccessoryArticleOther, Subtype: "other:other", PackageQuantity: 1,
			StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
			Attributes: []domain.AccessoryAttributeValue{{
				Key: "length", Kind: domain.AccessoryAttributeNumber, NumberValue: &value, Unit: &unit,
			}},
		}, "editor")
		createResult <- err
	}()

	select {
	case <-repository.validationStarted:
	case <-ctx.Done():
		t.Fatal("product mutation did not reach transactional validation")
	}
	updateStarted := make(chan struct{})
	updateResult := make(chan error, 1)
	go func() {
		close(updateStarted)
		updateCtx, updateCancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer updateCancel()
		_, err := masterData.Update(updateCtx, "accessory_custom_field", "length", application.MasterDataInput{
			Label: "Length", Active: &active,
			Metadata: map[string]any{"kind": "number", "unit": "cm"},
		})
		updateResult <- err
	}()
	<-updateStarted
	if err := <-updateResult; err == nil {
		t.Fatal("concurrent definition update committed while product validation was blocked")
	}
	close(repository.releaseValidation)

	if err := <-createResult; err != nil {
		t.Fatalf("serialized product create failed: %v", err)
	}
	if _, err := masterData.Update(ctx, "accessory_custom_field", "length", application.MasterDataInput{
		Label: "Length", Active: &active,
		Metadata: map[string]any{"kind": "number", "unit": "cm"},
	}); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("incompatible concurrent definition update error = %v", err)
	}
	definition, err := masterData.Get(ctx, "accessory_custom_field", "length")
	if err != nil {
		t.Fatal(err)
	}
	if definition.Metadata["unit"] != "mm" {
		t.Fatalf("incompatible definition update committed: %#v", definition.Metadata)
	}
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("database connection leaked or deadlocked: %v", err)
	}
}

func TestAccessoryCustomFieldImportCannotRaceProductCreate(t *testing.T) {
	_, db := testAccessoryService(t)
	db.SetMaxOpenConns(4)
	masterData := application.NewMasterDataService(db)
	active := true
	if _, err := masterData.Create(t.Context(), "accessory_custom_field", application.MasterDataInput{
		Key: "length", Label: "Length", Active: &active,
		Metadata: map[string]any{"kind": "number", "unit": "mm"},
	}); err != nil {
		t.Fatal(err)
	}
	doc, err := masterData.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	delete(doc.Entries, "accessory_custom_field")

	repository := &blockingAccessoryProductMutationRepository{
		AccessoryRepository: infrastructure.NewAccessoryRepository(db),
		validationStarted:   make(chan struct{}),
		releaseValidation:   make(chan struct{}),
	}
	accessories := application.NewAccessoryService(repository)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	value := 12.5
	unit := "mm"
	createResult := make(chan error, 1)
	go func() {
		_, err := accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
			Manufacturer: "Club", Name: "Measured article", Category: "other",
			ArticleType: domain.AccessoryArticleOther, Subtype: "other:other", PackageQuantity: 1,
			StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
			Attributes: []domain.AccessoryAttributeValue{{
				Key: "length", Kind: domain.AccessoryAttributeNumber, NumberValue: &value, Unit: &unit,
			}},
		}, "editor")
		createResult <- err
	}()

	select {
	case <-repository.validationStarted:
	case <-ctx.Done():
		t.Fatal("product mutation did not reach transactional validation")
	}
	importResult := make(chan error, 1)
	go func() {
		importCtx, importCancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer importCancel()
		_, err := masterData.Import(importCtx, doc)
		importResult <- err
	}()
	if err := <-importResult; err == nil {
		t.Fatal("concurrent import deleted the definition while product validation was blocked")
	}
	if _, err := masterData.Get(ctx, "accessory_custom_field", "length"); err != nil {
		t.Fatalf("concurrent import changed the definition: %v", err)
	}
	close(repository.releaseValidation)

	if err := <-createResult; err != nil {
		t.Fatalf("serialized product create failed: %v", err)
	}
	if _, err := masterData.Import(ctx, doc); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("definition-omitting import after product create error = %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("database connection leaked or deadlocked: %v", err)
	}
}

func TestAccessoryCustomFieldDefinitionChangeCannotRaceProductUpdate(t *testing.T) {
	baseAccessories, db := testAccessoryService(t)
	db.SetMaxOpenConns(4)
	masterData := application.NewMasterDataService(db)
	active := true
	if _, err := masterData.Create(t.Context(), "accessory_custom_field", application.MasterDataInput{
		Key: "length", Label: "Length", Active: &active,
		Metadata: map[string]any{"kind": "number", "unit": "mm"},
	}); err != nil {
		t.Fatal(err)
	}
	value := 5.0
	unit := "mm"
	input := application.CreateAccessoryProductInput{
		Manufacturer: "Club", Name: "Measured article", Category: "other",
		ArticleType: domain.AccessoryArticleOther, Subtype: "other:other", PackageQuantity: 1,
		StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
		Attributes: []domain.AccessoryAttributeValue{{
			Key: "length", Kind: domain.AccessoryAttributeNumber, NumberValue: &value, Unit: &unit,
		}},
	}
	product, err := baseAccessories.CreateProduct(t.Context(), input, "editor")
	if err != nil {
		t.Fatal(err)
	}
	value = 12.5

	repository := &blockingAccessoryProductMutationRepository{
		AccessoryRepository: infrastructure.NewAccessoryRepository(db),
		validationStarted:   make(chan struct{}),
		releaseValidation:   make(chan struct{}),
	}
	accessories := application.NewAccessoryService(repository)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	updateResult := make(chan error, 1)
	go func() {
		input.Name = "Edited measured article"
		_, err := accessories.UpdateProduct(ctx, product.ID, application.UpdateAccessoryProductInput{
			CreateAccessoryProductInput: input,
		}, "editor")
		updateResult <- err
	}()

	select {
	case <-repository.validationStarted:
	case <-ctx.Done():
		t.Fatal("product update did not reach transactional validation")
	}
	definitionResult := make(chan error, 1)
	go func() {
		updateCtx, updateCancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer updateCancel()
		_, err := masterData.Update(updateCtx, "accessory_custom_field", "length", application.MasterDataInput{
			Label: "Length", Active: &active,
			Metadata: map[string]any{"kind": "number", "unit": "mm", "max": 10},
		})
		definitionResult <- err
	}()
	if err := <-definitionResult; err == nil {
		t.Fatal("concurrent definition update committed while product update validation was blocked")
	}
	close(repository.releaseValidation)

	if err := <-updateResult; err != nil {
		t.Fatalf("serialized product update failed: %v", err)
	}
	if _, err := masterData.Update(ctx, "accessory_custom_field", "length", application.MasterDataInput{
		Label: "Length", Active: &active,
		Metadata: map[string]any{"kind": "number", "unit": "mm", "max": 10},
	}); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("incompatible definition update after product update error = %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("database connection leaked or deadlocked: %v", err)
	}
}
