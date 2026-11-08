package application

import (
	"database/sql"

	"railkeeper/backend/internal/authentication/bootstrap"
	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/infrastructure/authdata"
)

var ErrAlreadySetup = bootstrap.ErrAlreadySetup
var ErrWeakSetup = bootstrap.ErrWeakSetup
var ErrInvalidSetupCode = bootstrap.ErrInvalidSetupCode

type SetupService = bootstrap.Service
type CreateAdminInput = bootstrap.CreateAdminInput

func NewSetupService(db *sql.DB) *SetupService { return bootstrap.New(authdata.NewSetupStore(db)) }
func randomID() string                         { return credentials.RandomID() }
