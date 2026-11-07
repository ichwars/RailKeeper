package application

import "database/sql"

// Legacy account and 2FA mutations retain their existing transaction owner.
func requireAuthMutation(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrInvalidLogin
	}
	return nil
}
