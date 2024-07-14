package dbconnector

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Connector struct {
	DatabaseDSN string
}

func NewConnector(databaseDSN string) Connector {
	return Connector{DatabaseDSN: databaseDSN}
}

func (c *Connector) Connect(connectFunc func()) error {
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
		connectFunc()
	}
	return nil
}
