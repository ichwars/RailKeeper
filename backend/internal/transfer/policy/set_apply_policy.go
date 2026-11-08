package policy

import (
	"fmt"
	"slices"

	"railkeeper/backend/internal/transfer"
)

type TransferVehicleSetApplyPolicy struct {
	Action           string
	TargetSetID      string
	MemberRecordKeys map[string]bool
}

func TransferVehicleSetApplyPolicies(
	previews []transfer.DataTransferVehicleSetPreview,
	issues []transfer.DataTransferIssue,
) (map[string]TransferVehicleSetApplyPolicy, map[string]TransferVehicleSetApplyPolicy, error) {
	setPolicies := make(map[string]TransferVehicleSetApplyPolicy, len(previews))
	memberPolicies := map[string]TransferVehicleSetApplyPolicy{}
	for _, preview := range previews {
		if preview.RecordKey == "" {
			return nil, nil, applyConflict("vehicle set record key is missing")
		}
		if _, duplicate := setPolicies[preview.RecordKey]; duplicate {
			return nil, nil, applyConflict("duplicate vehicle set record key")
		}
		action := preview.ProposedAction
		resolved := false
		for _, issue := range issues {
			if issue.Area != transfer.TransferVehicles || issue.RecordKey != preview.RecordKey {
				continue
			}
			switch issue.Code {
			case "duplicate_vehicle_set_inventory_number", "vehicle_set_member_external_conflict":
				action = issue.SelectedResolution
				resolved = true
			}
		}
		if preview.Classification == "warning" && !resolved {
			return nil, nil, applyConflict("vehicle set conflict has no resolution")
		}
		if !slices.Contains([]string{"create", "replace", "copy", "skip"}, action) {
			return nil, nil, applyConflict("unsupported vehicle set resolution")
		}
		if action == "replace" && preview.TargetID == "" {
			return nil, nil, applyConflict("vehicle set replacement target is missing")
		}
		policy := TransferVehicleSetApplyPolicy{
			Action: action, TargetSetID: preview.TargetID, MemberRecordKeys: map[string]bool{},
		}
		for _, recordKey := range preview.MemberRecordKeys {
			if recordKey == "" || memberPolicies[recordKey].Action != "" {
				return nil, nil, applyConflict("vehicle belongs to more than one imported set")
			}
			policy.MemberRecordKeys[recordKey] = true
		}
		if len(policy.MemberRecordKeys) == 0 {
			return nil, nil, applyConflict("vehicle set has no member records")
		}
		setPolicies[preview.RecordKey] = policy
		for recordKey := range policy.MemberRecordKeys {
			memberPolicies[recordKey] = policy
		}
	}
	return setPolicies, memberPolicies, nil
}
func applyConflict(message string) error {
	return fmt.Errorf("%w: %s", ErrDataTransferConflict, message)
}
