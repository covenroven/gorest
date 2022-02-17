package database

import (
	"database/sql"
	"fmt"

	"github.com/covenroven/gorest/config"
	_ "github.com/lib/pq"
)

// Open connection to database
func InitDB() (*sql.DB, error) {
	connInfo := fmt.Sprintf(`
		host=%s port=%s user=%s password=%s dbname=%s sslmode=disable
	`, config.DB_HOST, config.DB_PORT, config.DB_USER, config.DB_PASS, config.DB_NAME)
	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
