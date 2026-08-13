package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func raw() {
	// Database connection parameters
	connStr := "host=localhost user=postgres password=0000 dbname=todo_app sslmode=disable"

	// Open a connection to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Insert
	_, err = db.Exec("INSERT INTO todos (title) VALUES ($1)", "Learn PostgreSQL")
	if err != nil {
		log.Fatal(err)
	}

	// Query
	rows, err := db.Query("SELECT id, title, completed FROM todos")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title string
		var completed bool
		if err := rows.Scan(&id, &title, &completed); err != nil {
			log.Fatal(err)
		}
		log.Println(id, title, completed)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}
