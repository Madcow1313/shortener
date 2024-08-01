package dbconnector

import (
	"database/sql"
)

func (c *Connector) UpdateIsDeletedColumn(db *sql.DB, urls chan string) error {
	tx, err := db.Begin()
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
			if _, err = stmt.Exec(); err != nil {
				return err
			}
			if stmt.Close() != nil {
				return err
			}
			return tx.Commit()
		}
	}
}
