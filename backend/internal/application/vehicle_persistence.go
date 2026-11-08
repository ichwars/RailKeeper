package application

import (
	"context"
	"database/sql"
)

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
