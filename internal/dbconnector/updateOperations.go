package dbconnector

import (
	"context"
	"database/sql"
)

func (c *Connector) UpdateIsDeletedColumn(db *sql.DB, ctx context.Context, urls chan string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(updateOnDeleteQuery)
	if err != nil {
		return err
	}
	for {
		val, ok := <-urls
		if _, err = stmt.Exec(val); err != nil {
			return err
		}
		if !ok {
			if stmt.Close() != nil {
				return err
			}
			return tx.Commit()
		}
	}
}
