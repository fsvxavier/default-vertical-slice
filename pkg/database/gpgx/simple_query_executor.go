package gpgx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type SimpleQueryExecutor struct {
	pgConn *PgConnection

	tx *pgx.Tx
}

func (sqe SimpleQueryExecutor) WithTx(tx *pgx.Tx) *SimpleQueryExecutor {
	sqe.tx = tx
	return &sqe
}

func (sqe *SimpleQueryExecutor) ExecQueryWithContext(ctx context.Context, dst any, multiple bool, query string, params ...any) error {
	if ctx == nil {
		ctx = context.TODO()
	}

	f := func() error {
		return sqe.pgConn.queryFor(ctx, sqe.tx, dst, multiple, query, params...)
	}

	err := f()
	if _, ok := err.(*NotConnectedError); ok {
		fmt.Println("DB connection lost. Reconnecting ...")
		err = sqe.pgConn.Reconnect(ctx)
		if err != nil {
			fmt.Printf("DB connection unable to reconnect - %s", err)
			return err
		}
		return f()
	}

	if dst == nil {
		if pgErr, ok := err.(*PgError); ok {
			if pgErr != nil && pgErr.IsEmptyResult() {
				return nil
			}
		}
	}

	if ctm, ok := dst.(ContextualModel); ok {
		ctm.SetContext(ctx)
	}

	return err
}
