package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/prigas-dev/backoffice-ai/http_server"
	"github.com/spf13/afero"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbPath := "./kanban.db"
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = os.MkdirAll("tmp/operations", 0755)
	if err != nil {
		panic(err)
	}
	operationsFs := afero.NewBasePathFs(afero.NewOsFs(), "tmp/operations")

	ctx := context.Background()

	http_server.Start(ctx, db, operationsFs)
}
