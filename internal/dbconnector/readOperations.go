package dbconnector

import "database/sql"

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
