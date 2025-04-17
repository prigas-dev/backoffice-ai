package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/joho/godotenv"
	"github.com/prigas-dev/backoffice-ai/http_server"
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

	ctx := context.Background()

	http_server.Start(ctx, db)
}

func main_Pages_Database() {
	db, err := gorm.Open(sqlite.Open("pages.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to pages.db: %v", err)
	}

	db.AutoMigrate(&Page{}, &Component{}, &PageComponent{})
}

func main_HTTP_Server() {
	HTML()
}

func main_Sqlite3_Execution_With_AI() {

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
