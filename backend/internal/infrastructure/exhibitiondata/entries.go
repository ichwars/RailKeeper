package exhibitiondata

import (
	"context"
	"fmt"

	"railkeeper/backend/internal/exhibition"
)

func (m *mutation) InsertEntry(ctx context.Context, write exhibition.EntryWrite) error {
	input, id, listID, now := write.Input, write.ID, write.ListID, write.At
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO exhibition_entries(
  id, list_id, owner, image_url, locomotive_name, vehicle_id,
  gattung, series, manufacturer, epoch, railway_company, day_scope,
  dt_decoder, decoder_number, decoder_type, adapter, interface_name, sx_address, analog,
  function_keys, notes, sort_order, availability, revision, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
`, id, listID, input.Owner, input.ImageURL, input.LocomotiveName, input.VehicleID, input.Gattung,
		input.Series, input.Manufacturer, input.Epoch, input.RailwayCompany, input.DayScope,
		boolToInt(input.DTDecoder), input.DecoderNumber, input.DecoderType, input.Adapter,
		input.InterfaceName, input.SXAddress, boolToInt(input.Analog), input.FunctionKeys,
		input.Notes, input.SortOrder, input.Availability, now, now); err != nil {
		return fmt.Errorf("create exhibition entry: %w", err)
	}
	return nil
}
func (m *mutation) UpdateEntry(ctx context.Context, write exhibition.EntryWrite) (bool, error) {
	input, entryID, listID, now := write.Input, write.ID, write.ListID, write.At
	result, err := m.tx.ExecContext(ctx, `
UPDATE exhibition_entries
SET owner=?, image_url=?, locomotive_name=?,
    vehicle_id=?, gattung=?, series=?, manufacturer=?, epoch=?, railway_company=?, day_scope=?,
    dt_decoder=?, decoder_number=?, decoder_type=?, adapter=?, interface_name=?, sx_address=?, analog=?,
    function_keys=?, notes=?, sort_order=?, availability=?, revision=revision+1, updated_at=?
WHERE id=? AND list_id=? AND (?=0 OR revision=?)
`, input.Owner, input.ImageURL, input.LocomotiveName, input.VehicleID, input.Gattung, input.Series,
		input.Manufacturer, input.Epoch, input.RailwayCompany, input.DayScope, boolToInt(input.DTDecoder),
		input.DecoderNumber, input.DecoderType, input.Adapter, input.InterfaceName, input.SXAddress,
		boolToInt(input.Analog), input.FunctionKeys, input.Notes, input.SortOrder, input.Availability,
		now, entryID, listID, input.ExpectedRevision, input.ExpectedRevision)
	if err != nil {
		return false, fmt.Errorf("update exhibition entry: %w", err)
	}
	return changed(result), nil
}
func (m *mutation) DeleteEntry(ctx context.Context, listID, entryID string) (bool, error) {
	result, err := m.tx.ExecContext(ctx, `DELETE FROM exhibition_entries WHERE id=? AND list_id=?`, entryID, listID)
	if err != nil {
		return false, fmt.Errorf("delete exhibition entry: %w", err)
	}
	return changed(result), nil
}
func (m *mutation) TouchUnlockedList(ctx context.Context, listID, now string) (bool, error) {
	result, err := m.tx.ExecContext(ctx, `
UPDATE exhibition_lists SET revision=revision+1, updated_at=? WHERE id=? AND locked=0
`, now, listID)
	if err != nil {
		return false, fmt.Errorf("touch exhibition list: %w", err)
	}
	return changed(result), nil
}
