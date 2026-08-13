package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"example/ToDo/models"
	"example/ToDo/utils"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "0000")
	dbName := getEnv("DB_NAME", "todo_gorm_db")
	port := getEnv("DB_PORT", "5432")

	// 1. Connect to the default postgres database to check if our target database exists
	defaultDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable", host, user, password, port)
	dbRaw, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		log.Fatal("Failed to connect to postgres server:", err)
	}

	var exist int
	// 2. Check if the database exists
	checkQuery := fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", dbName)
	err = dbRaw.QueryRow(checkQuery).Scan(&exist)
	if err != nil {
		maxRepeat := 5
		for i := 1; i <= maxRepeat; i++ {
			createStmt := fmt.Sprintf("CREATE DATABASE %s", dbName)
			_, err = dbRaw.Exec(createStmt)

			if err == nil {
				fmt.Printf("Database '%s' created successfully on attempt %d\n", dbName, i)
				break
			}
			fmt.Printf("Attempt %d failed to create database: %v\n", i, err)

			if i == maxRepeat {
				log.Fatal("Failed to create database after 5 attempts. Stopping the program.")
			}

			time.Sleep(3 * time.Second) // Wait for 3 seconds before retrying
		}
	} else {
		fmt.Printf("Database '%s' already exists.\n", dbName)
	}
	dbRaw.Close()

	// 3. Create a new connection via GORM with Retry Loop for Docker stability
	gormDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbName, port)

	var db *gorm.DB
	maxDbRetries := 5
	for i := 1; i <= maxDbRetries; i++ {
		db, err = gorm.Open(postgres.Open(gormDSN), &gorm.Config{})
		if err == nil {
			fmt.Printf("Connected to '%s' via GORM successfully\n", dbName)
			break
		}
		fmt.Printf("Failed to connect to GORM (attempt %d/%d): %v. Retrying in 3 seconds...\n", i, maxDbRetries, err)

		if i == maxDbRetries {
			log.Fatal("Failed to connect to the database via GORM after multiple attempts:", err)
		}
		time.Sleep(3 * time.Second)
	}

	// 4. AutoMigrate
	err = db.AutoMigrate(&models.User{}, &models.Todo{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	fmt.Println("Tables User and Todos are auto-migrated successfully")

	SeedAdmin(db) // Seed the admin user if it doesn't exist

	// 5. Return the database connection
	return db
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func SeedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Where("username = ?", "admin").Count(&count)

	if count == 0 {
		hashedPassword, _ := utils.HashPassword("admin123")
		adminUser := models.User{
			Username: "admin",
			Password: hashedPassword,
			Role:     "admin",
		}
		db.Create(&adminUser)
		fmt.Println("Master Admin seeded successfully!")
	} else {
		fmt.Println("Master Admin already exists.")
	}
}
