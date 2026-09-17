package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

const DbName = "troubleshoot_tracker.db"

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite", DbName)
	if err != nil {
		log.Fatalf("Gagal membuka database: %v", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS tickets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		customer TEXT,
		product TEXT,
		reported_to_ioh TEXT,
		date_ticket TEXT,
		ticket_number TEXT,
		log_issue TEXT,
		identified_issue TEXT,
		resolve_date TEXT,
		sla TEXT,
		note TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Gagal membuat tabel tracker: %v", err)
	}

	return db
}