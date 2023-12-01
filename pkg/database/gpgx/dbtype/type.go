package dbtype

import (
	"github.com/jackc/pgx/v5/pgtype"
)

func NewText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}
