package store

import "database/sql"

func InitDB() (*sql.DB, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	if err := Migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
