package decoder

// CVRecord contains only persisted fields, including nullable index/source identities.
type CVRecord struct {
	ID, VehicleID                   string
	CVNumber                        int
	IndexCV31, IndexCV32            *int
	Value                           int
	Description, Category, Protocol string
	DecoderProfile, SourceFileID    string
	SourceLine                      *int
	SourceExportedAt                string
	CreatedAt, UpdatedAt            string
}
