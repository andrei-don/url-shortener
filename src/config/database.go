package config

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var sqlOpen = sql.Open

func ConnectDatabase(dsn string) (*sql.DB, error) {
	db, err := sqlOpen("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database is not reachable: %v", err)
	}

	fmt.Println("Successfully connected!")
	return db, nil
}
