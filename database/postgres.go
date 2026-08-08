package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"example/ToDo/models"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {

	// 1. Connect to the default postgres database to check if our target database exists
	defaultDSN := "host=localhost user=postgres password=0000 dbname=postgres sslmode=disable"
	dbRaw, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		log.Fatal("Failed to connect to postgres server:", err)
	}

	var exist int
	// 2. Check if the database exists
	err = dbRaw.QueryRow("SELECT 1 FROM pg_database WHERE datname = 'todo_gorm_db'").Scan(&exist)
	if err != nil {
		maxRepeat := 3
		for i := 1; i <= maxRepeat; i++ {
			_, err = dbRaw.Exec("CREATE DATABASE todo_gorm_db")

			if err == nil {
				fmt.Printf("Database 'todo_gorm_db' created successfully on attempt %d\n", i)
				break
			}
			fmt.Printf("Attempt %d failed to create database: %v\n", i, err)

			if i == maxRepeat {
				log.Fatal("Failed to create database after 3 attempts. Stopping the program.")
			}

			time.Sleep(2 * time.Second) // Wait for 2 seconds before retrying
		}
	} else {
		fmt.Println("Database 'todo_gorm_db' already exists.")
	}
	dbRaw.Close()

	// 3. Create a new connection via GORM
	gormDSN := "host=localhost user=postgres password=0000 dbname=todo_gorm_db sslmode=disable"
	db, err := gorm.Open(postgres.Open(gormDSN), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to the database via GORM:", err)
	}
	fmt.Println("Connected to 'todo_gorm_db' via GORM")

	// 4. AutoMigrate
	err = db.AutoMigrate(&models.Todo{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	fmt.Println("Table Todos is auto-migrated successfully")

	// 5. Return the database connection
	return db
}
