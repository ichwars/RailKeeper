package decoder

func CheckCVRevision(existing *CVRecord, expected *string) error {
	if expected == nil {
		return nil
	} // Existing clients keep their legacy request contract.
	if *expected != "" && !ValidCVRevision(*expected) {
		return ErrCVValidation
	}
	if existing == nil {
		if *expected != "" {
			return ErrCVConflict
		}
	} else if CVRecordRevision(*existing) != *expected {
		return ErrCVConflict
	}
	return nil
}
