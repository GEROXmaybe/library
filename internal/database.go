package internal

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS books (
		id TEXT PRIMARY KEY,
		title TEXT,
		author TEXT,
		isbn TEXT,
		year INTEGER,
		status TEXT
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT,
		email TEXT,
		registration_date DATETIME
	);

	CREATE TABLE IF NOT EXISTS issues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		book_id TEXT,
		user_id TEXT,
		issue_date DATETIME,
		due_date DATETIME,
		return_date DATETIME,
		FOREIGN KEY(book_id) REFERENCES books(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	_, err = DB.Exec(createTables)
	if err != nil {
		log.Fatalf("Ошибка создания таблиц: %v", err)
	}
}
