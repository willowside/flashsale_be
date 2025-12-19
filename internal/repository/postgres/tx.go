package postgres

// import (
// 	"context"
// 	"flashsale/internal/repository"

// 	"github.com/jmoiron/sqlx"
// )

// type TxWrapper struct {
// 	tx *sqlx.Tx
// }

// func (t *TxWrapper) Exec(ctx context.Context, query string, args ...any) error {
// 	_, err := t.tx.ExecContext(ctx, query, args...)
// 	return err
// }

// func (t *TxWrapper) QueryRow(ctx context.Context, query string, args ...any) repository.Row {
// 	return t.tx.QueryRowxContext(ctx, query, args...)
// }

// func (t *TxWrapper) Commit() error {
// 	return t.tx.Commit()
// }

// func (t *TxWrapper) Rollback() error {
// 	return t.tx.Rollback()
// }

// type DBWrapper struct {
// 	db *sqlx.DB
// }

// func NewDBWrapper(db *sqlx.DB) *DBWrapper {
// 	return &DBWrapper{db: db}
// }

// func (d *DBWrapper) BeginTx(ctx context.Context) (repository.Tx, error) {
// 	tx, err := d.db.BeginTxx(ctx, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &TxWrapper{tx: tx}, nil
// }
