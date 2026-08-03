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

var DB *gorm.DB

func ConnectDB() {

	// Database connection parameters
	defualtDSN := "host=localhost user=postgres password=0000 dbname=todo_app sslmode=disable"
	dbRaw, err := sql.Open("postgres", defualtDSN)
	if err != nil {
		log.Fatal(err)
	}

	var exist int
	// Check if the database exists
	//return 1 if it search the postgresql list of databases and find the database name, otherwise return err
	err = dbRaw.QueryRow("Select 1 From pg_database Where datname = 'todo_gorm_db'").Scan(&exist)
	if err != nil {
		maxRepeat := 3
		for i := 1; i <= maxRepeat; i++ {
			_, err = dbRaw.Exec("Create Database todo_gorm_db")

			if err == nil {
				fmt.Println("Database 'todo_gorm_db' created successfully on attempt", i)
				break
			}
			fmt.Printf("Attempt %d failed to create database: %v\n", i, err)

			if i == maxRepeat {
				log.Fatal("Failed to create database after 3 attempts. Stopping the program.")
			}

			time.Sleep(2 * time.Second) // Wait for 2 seconds before retrying
		}
	} else {
		fmt.Println("Database 'todo_gorm_db' is already exists.")
	}
	dbRaw.Close()

	// Create a new connection to create database table
	gormDNS := "host=localhost user=postgres password=0000 dbname=todo_gorm_db sslmode=disable"
	db, err := gorm.Open(postgres.Open(gormDNS), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	fmt.Println("Connected to 'todo_gorm_db' via GORM")

	err = db.AutoMigrate(&models.Todo{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	fmt.Println("Table Todos is auto-migrated successfully")

}
