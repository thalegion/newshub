package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type ConnectionProvider struct {
	user     string
	password string
	host     string
	port     string
	dbname   string
}

func NewConnectionProvider(user, password, host, port, dbname string) *ConnectionProvider {
	return &ConnectionProvider{
		user:     user,
		password: password,
		host:     host,
		port:     port,
		dbname:   dbname,
	}
}

func (connectionProvider ConnectionProvider) GetConnection() *sql.DB {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		connectionProvider.user,
		connectionProvider.password,
		connectionProvider.host,
		connectionProvider.port,
		connectionProvider.dbname,
	)

	// Open a database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	return db
}
