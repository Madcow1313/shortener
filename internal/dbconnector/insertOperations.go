package dbconnector

import (
	"database/sql"

	"github.com/lib/pq"
)

func (c *Connector) InsertURL(db *sql.DB, key, value, userID string) error {
	_, err := db.Exec(insertQuery, key, value, userID)
	if err != nil {
		c.Z.LogError(err)
		return err
	}
	return nil
}

func (c *Connector) InsertBatchToDatabase(db *sql.DB, data map[string]string, userID string) error {
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(pq.CopyIn("url", "short", "origin", "user_id"))
	if err != nil {
		return err
	}
	for key, val := range data {
		_, err = stmt.Exec(key, val, userID)
		if err != nil {
			return err
		}
	}
	if _, err = stmt.Exec(); err != nil {
		return err
	}
	if stmt.Close() != nil {
		return err
	}
	return tx.Commit()
}
