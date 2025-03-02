package config

import (
	"database/sql"
	"fmt"
	"log"

	"time"

	_ "github.com/lib/pq"
)

var sqlOpen = sql.Open

func ConnectDatabase(dsn string, maxRetries int, retryInterval time.Duration) (*sql.DB, error) {

	var db *sql.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = sqlOpen("postgres", dsn)
		if err != nil {
			log.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(retryInterval)
			continue
		}

		if err = db.Ping(); err == nil {
			fmt.Println("Successfully connected!")
			return db, nil
		}
		log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryInterval)
	}
	return nil, fmt.Errorf("could not connect to database after %d retries: %v", maxRetries, err)
}
