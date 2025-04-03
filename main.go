package main

import (
	"database/sql"
	"log"
	"os"
)

func main() {

	dbPath := "./kanban.db"
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	prompt := os.Args[1]

	Sqliter(db, prompt)
}
