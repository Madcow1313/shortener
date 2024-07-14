package dbconnector

import (
	"database/sql"
	"shortener/internal/mylogger"

	_ "github.com/lib/pq"
)

const (
	createQuery = "CREATE TABLE IF NOT EXISTS url (short varchar, origin varchar)"
	insertQuery = "INSERT INTO url (short, origin) VALUES ($1, $2)"
	selectQuery = "SELECT * FROM url"
)

type Connector struct {
	DatabaseDSN string
	LastResult  string
	URLmap      map[string]string
}

func NewConnector(databaseDSN string) Connector {
	return Connector{DatabaseDSN: databaseDSN}
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
	}
	c.URLmap = make(map[string]string)
	for rows.Next() {
		var short, origin string
		rows.Scan(&short, origin)
		c.URLmap[short] = origin
	}
	return nil
}
