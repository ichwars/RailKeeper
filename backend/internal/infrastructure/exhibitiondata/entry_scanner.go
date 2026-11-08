package exhibitiondata

import (
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/exhibition"
)

type exhibitionEntryScanner interface{ Scan(dest ...any) error }

func scanExhibitionEntry(row exhibitionEntryScanner) (exhibition.ExhibitionEntry, error) {
	var entry exhibition.ExhibitionEntry
	var imageURL, decoderNumber, functionKeys, notes sql.NullString
	var dtDecoder int
	var analog int
	if err := row.Scan(
		&entry.ID,
		&entry.ListID,
		&entry.Owner,
		&imageURL,
		&entry.LocomotiveName,
		&entry.VehicleID,
		&entry.Gattung,
		&entry.Series,
		&entry.Manufacturer,
		&entry.Epoch,
		&entry.RailwayCompany,
		&entry.DayScope,
		&dtDecoder,
		&decoderNumber,
		&entry.DecoderType,
		&entry.Adapter,
		&entry.InterfaceName,
		&entry.SXAddress,
		&analog,
		&functionKeys,
		&notes,
		&entry.SortOrder,
		&entry.Availability,
		&entry.Revision,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	); err != nil {
		return exhibition.ExhibitionEntry{}, fmt.Errorf("scan exhibition entry: %w", err)
	}
	entry.ImageURL = imageURL.String
	entry.DTDecoder = dtDecoder == 1
	entry.DecoderNumber = decoderNumber.String
	entry.Analog = analog == 1
	entry.FunctionKeys = functionKeys.String
	entry.Notes = notes.String
	return entry, nil
}
