package dbconnector

import (
	"database/sql"
	"fmt"
	"shortener/internal/mylogger"

	"github.com/lib/pq"
)

const (
	createQuery         = "CREATE TABLE IF NOT EXISTS url (short varchar, origin varchar unique, user_id varchar, is_deleted boolean DEFAULT false)"
	insertQuery         = "INSERT INTO url (short, origin, user_id) VALUES ($1, $2, $3)"
	selectQuery         = "SELECT * FROM url"
	selectShortURL      = "SELECT short FROM url WHERE origin=$1"
	selectIsDeleted     = "SELECT is_deleted FROM url WHERE short=$1"
	updateOnDeleteQuery = "UPDATE url SET is_deleted = true WHERE short=$1 AND is_deleted=false"
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

func (c *Connector) Connect(connectFunc func(db *sql.DB, args ...interface{}) error) error {
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

func (c *Connector) InsertURL(db *sql.DB, key, value, userID string) error {
	_, err := db.Exec(insertQuery, key, value, userID)
	if err != nil {
		c.Z.LogError(err)
		return err
	}
	return nil
}

func (c *Connector) ReadFromDB(db *sql.DB) error {
	rows, err := db.Query(selectQuery)
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	c.URLmap = make(map[string]string)
	c.UserURLS = make(map[string][]string)
	for rows.Next() {
		var short, origin, userID string
		rows.Scan(&short, &origin, &userID)
		c.URLmap[short] = origin
		if _, ok := c.UserURLS[userID]; !ok {
			c.UserURLS[userID] = make([]string, 0)
		}
		c.UserURLS[userID] = append(c.UserURLS[userID], short)
	}
	return nil
}

func (c *Connector) UpdateOnDelete(db *sql.DB, urls chan string) error {
	tx, _ := db.Begin()
	stmt, err := tx.Prepare(pq.CopyIn("url", "short"))
	if err != nil {
		return err
	}
	counter := 0
	for {
		val, ok := <-urls
		stmt.Exec(val)
		counter++
		if !ok {
			stmt.Exec()
			tx.Commit()
			fmt.Println("done")
			break
		}
	}
	// stmt.Exec()
	// tx.Commit()
	return nil
}

func (c *Connector) InsertBatchToDatabase(db *sql.DB, data map[string]string, userID string) error {
	stmt, err := db.Prepare(insertQuery)
	if err != nil {
		return err
	}
	for key, val := range data {
		tx, err := db.Begin()
		if err != nil {
			continue
		}
		_, err = stmt.Exec(key, val, userID)
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

func (c *Connector) IsShortDeleted(db *sql.DB, short string) error {
	r, err := db.Query(selectIsDeleted, short)
	if err != nil || r.Err() != nil {
		return err
	}
	r.Next()
	err = r.Scan(&c.IsDeleted)
	if err != nil {
		return err
	}
	return nil
}
