package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

// Connect initializes the database connection using the SUPABASE_DB_CONN environment variable.
func Connect() {
	connStr := os.Getenv("SUPABASE_DB_CONN")
	if connStr == "" {
		log.Fatal("SUPABASE_DB_CONN environment variable is not set")
	}

	// Supabase Pooler (Transaction Mode) does not support prepared statements.
	// Force simple protocol if not already specified.
	if !strings.Contains(connStr, "simple_protocol") {
		if strings.Contains(connStr, "?") {
			connStr += "&default_query_exec_mode=simple_protocol"
		} else {
			connStr += "?default_query_exec_mode=simple_protocol"
		}
	}

	var err error
	DB, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	fmt.Println("Successfully connected to Supabase database")
}
