package gpgx

import (
	"errors"
	"strings"
)

var ErrNoRows = errors.New("no rows in result set")

type PgError struct {
	Message string
}

func NewPgError(message string) *PgError {
	return &PgError{Message: message}
}

func (pe PgError) Error() string {
	return pe.Message
}

func (pe PgError) IsEmptyResult() bool {
	return strings.Contains(pe.Message, "no row was found") ||
		strings.Contains(pe.Message, ErrNoRows.Error())
}

func (pe PgError) IsFinal() bool {
	return strings.Contains(pe.Message, "rows final error")
}

func (pe PgError) ReturnedMultipleRows() bool {
	return strings.Contains(pe.Message, "expected 1 row, got:") ||
		strings.Contains(pe.Message, "query multiple result rows")
}
