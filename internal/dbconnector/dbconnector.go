package dbconnector

import (
	"database/sql"
	"shortener/internal/mylogger"

	_ "github.com/lib/pq"
)

const (
	createQuery    = "CREATE TABLE IF NOT EXISTS url (short varchar, origin varchar unique)"
	insertQuery    = "INSERT INTO url (short, origin) VALUES ($1, $2)"
	selectQuery    = "SELECT * FROM url"
	selectShortURL = "SELECT short FROM url WHERE origin=$1"
)

type Connector struct {
	DatabaseDSN string
	LastResult  string
	URLmap      map[string]string
}

func NewConnector(databaseDSN string) *Connector {
	return &Connector{DatabaseDSN: databaseDSN}
}

func (c *Connector) Connect(connectFunc func(db *sql.DB, args ...interface{}) error) error {
	db, err := sql.Open("postgres", c.DatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		return err
	}
	if connectFunc != nil {
		err = connectFunc(db)
		return err
	}
	return nil
}

func (c *Connector) CreateTable(db *sql.DB) error {
	_, err := db.Exec(createQuery)
	if err != nil {
		mylogger.LogError(err)
		return err
	}
	return nil
}

func (c *Connector) InsertURL(db *sql.DB, key, value string) error {
	_, err := db.Exec(insertQuery, key, value)
	if err != nil {
		mylogger.LogError(err)
		return err
	}
	return nil
}

func (c *Connector) ReadFromDB(db *sql.DB) error {
	rows, err := db.Query(selectQuery)
	if err != nil {
		mylogger.LogError(err)
		return err
	}
	if rows.Err() != nil {
		mylogger.LogError(rows.Err())
		return err
	}
	c.URLmap = make(map[string]string)
	for rows.Next() {
		var short, origin string
		rows.Scan(&short, origin)
		c.URLmap[short] = origin
	}
	return nil
}

func (c *Connector) InsertBatchToDatabase(db *sql.DB, data map[string]string) error {
	stmt, err := db.Prepare(insertQuery)
	if err != nil {
		return err
	}
	for key, val := range data {
		tx, err := db.Begin()
		if err != nil {
			continue
		}
		_, err = stmt.Exec(key, val)
		if err != nil {
			tx.Rollback()
			continue
		}
		tx.Commit()
	}
	return err
}

func (c *Connector) SelectShortURL(db *sql.DB, origin string) error {
	r, err := db.Query(selectShortURL, origin)
	if err != nil || r.Err() != nil {
		return err
	}
	r.Next()
	err = r.Scan(&c.LastResult)
	if err != nil {
		return err
	}
	return nil
}
