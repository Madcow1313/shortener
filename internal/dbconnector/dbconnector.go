package dbconnector

import (
	"database/sql"
	"shortener/internal/mylogger"

	_ "github.com/lib/pq"
)

const (
	createQuery         = "CREATE TABLE IF NOT EXISTS url (short varchar, origin varchar unique, user_id varchar, is_deleted boolean DEFAULT false)"
	insertQuery         = "INSERT INTO url (short, origin, user_id) VALUES ($1, $2, $3)"
	selectQuery         = "SELECT * FROM url"
	selectShortURL      = "SELECT short FROM url WHERE origin=$1"
	selectIsDeleted     = "SELECT is_deleted FROM url WHERE short=$1"
	updateOnDeleteQuery = "UPDATE url SET is_deleted = true WHERE short=$1"
)

type Connector struct {
	DatabaseDSN string
	LastResult  string
	IsDeleted   bool
	URLmap      map[string]string
	UserURLS    map[string][]string
	DB          *sql.DB
	Z           *mylogger.Mylogger
}

func NewConnector(databaseDSN string) *Connector {
	return &Connector{DatabaseDSN: databaseDSN}
}

func (c *Connector) ConnectToDB(connectFunc func(db *sql.DB, args ...interface{}) error) error {
	if c.DB == nil {
		db, err := sql.Open("postgres", c.DatabaseDSN)
		if err != nil {
			return err
		}
		c.DB = db
	}

	err := c.DB.Ping()
	if err != nil {
		return err
	}
	if connectFunc != nil {
		err = connectFunc(c.DB)
		return err
	}
	return nil
}

func (c *Connector) CreateTable(db *sql.DB) error {
	_, err := db.Exec(createQuery)
	if err != nil {
		return err
	}
	return nil
}
