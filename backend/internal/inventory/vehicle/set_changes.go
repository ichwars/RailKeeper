package vehicle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/inventory/numbering"
)

func (s *SetService) CreateSet(
	ctx context.Context,
	input CreateVehicleSetInput,
	actorUserID string,
) (*VehicleSet, error) {
	setInput := cleanVehicleSetInput(input.Set)
	if !isValidVehicleSetInput(setInput) || len(input.Members) < 2 || len(input.Members) > MaxVehicleSetMembers {
		return nil, ErrVehicleSetValidation
	}

	preparedMembers := make([]CreateVehicleInput, len(input.Members))
	memberIDs := make([]string, len(input.Members))
	for index, member := range input.Members {
		member = mergeVehicleSetMember(setInput, member, index)
		var err error
		member, err = ValidateCreateInput(member)
		if err != nil {
			return nil, err
		}
		memberIDs[index] = credentials.RandomID()
		preparedMembers[index] = member
	}

	setID, err := ResolveCreationIdentity(ctx, s.repository, input.CreationID)
	if err != nil {
		return nil, err
	}
	for index, member := range preparedMembers {
		var err error
		if len(member.Images) > 0 {
			member.Images, err = s.environment.LocalizeImages(ctx, memberIDs[index], member.Images)
			if err != nil {
				return nil, err
			}
		}
		preparedMembers[index] = member
		defer s.environment.CleanupImages(ctx, member.Images)
	}

	tx, err := s.repository.BeginSet(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create vehicle set: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if input.CreationID != "" {
		if err := tx.CheckCreationIdentity(ctx, setID); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	setInventoryNumber, err := ReserveVehicleSetNumber(ctx, tx)
	if err != nil {
		return nil, err
	}
	if err = tx.InsertSet(ctx, SetWrite{ID: setID, InventoryNumber: setInventoryNumber, Input: setInput, At: now}); err != nil {
		return nil, err
	}

	members := make([]Vehicle, 0, len(preparedMembers))
	for index, memberInput := range preparedMembers {
		member, createErr := InsertVehicle(ctx, tx, memberIDs[index], memberInput, actorUserID, now)
		if createErr != nil {
			return nil, createErr
		}
		if err = tx.InsertSetMember(ctx, SetMemberWrite{SetID: setID, VehicleID: member.ID, Position: index + 1, Label: member.Name}); err != nil {
			return nil, err
		}
		if err := tx.LinkCreatedSetMember(ctx, setID, member.ID); err != nil {
			return nil, err
		}
		member.VehicleSetID = setID
		member.VehicleSetName = setInput.Name
		member.VehicleSetPosition = index + 1
		member.VehicleSetSize = len(preparedMembers)
		member.VehicleSet = vehicleSetSummary(setID, setInventoryNumber, setInput, len(preparedMembers), index+1)
		members = append(members, *member)
	}

	if err = tx.AuditSetCreated(ctx, SetAudit{ID: credentials.RandomID(), Actor: actorUserID, SetID: setID, At: now}); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create vehicle set: %w", err)
	}

	for index := range members {
		images, loadErr := s.repository.MemberImages(ctx, members[index].ID)
		if loadErr != nil {
			return nil, loadErr
		}
		members[index].Images = images
	}

	result := vehicleSetFromInput(setID, setInventoryNumber, setInput, members, now)
	if err := s.environment.AttachMainImage(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SetService) GetSet(ctx context.Context, id string) (*VehicleSet, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrVehicleSetNotFound
	}

	set, err := s.repository.Set(ctx, id)
	if err != nil {
		return nil, err
	}
	memberIDs, err := s.repository.MemberIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	loadedMembers, err := s.environment.LoadMembers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load vehicle set members: %w", err)
	}

	membersByID := make(map[string]Vehicle, len(loadedMembers))
	for _, member := range loadedMembers {
		membersByID[member.ID] = member
	}
	set.Members = make([]Vehicle, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		member, found := membersByID[memberID]
		if !found {
			return nil, fmt.Errorf("get vehicle set member %s: %w", memberID, ErrVehicleNotFound)
		}
		set.Members = append(set.Members, member)
	}
	if err := s.environment.AttachMainImage(ctx, &set); err != nil {
		return nil, err
	}
	return &set, nil
}

func (s *SetService) UpdateSet(
	ctx context.Context,
	id string,
	input VehicleSetInput,
	actorUserID string,
) (*VehicleSet, error) {
	id = strings.TrimSpace(id)
	input = cleanVehicleSetInput(input)
	if id == "" {
		return nil, ErrVehicleSetNotFound
	}
	if !isValidVehicleSetInput(input) {
		return nil, ErrVehicleSetValidation
	}

	tx, err := s.repository.BeginSet(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update vehicle set: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339)
	affected, err := tx.UpdateSet(ctx, id, input, now)
	if err != nil {
		return nil, err
	}

	if affected == 0 {
		return nil, ErrVehicleSetNotFound
	}

	if err = tx.AuditSetUpdated(ctx, SetAudit{ID: credentials.RandomID(), Actor: actorUserID, SetID: id, At: now}); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update vehicle set: %w", err)
	}
	return s.GetSet(ctx, id)
}

func ReserveVehicleSetNumber(ctx context.Context, tx SetMutation) (string, error) {
	return numbering.Reserve(ctx, tx.Counter(), "Set", "", func(candidate string) error {
		used, err := tx.SetInventoryNumberInUse(ctx, strings.TrimSpace(candidate))
		if err != nil {
			return err
		}
		if used {
			return numbering.ErrConflict
		}
		return nil
	})
}
