package main

import (
	"example/ToDo/database"
	"example/ToDo/handler"
	"example/ToDo/repository"
	"example/ToDo/router"
	"fmt"
)

func main() {
	// 1. Seeding the database and connecting to it
	db := database.ConnectDB()

	// 2. Repositories
	todoRepo := repository.NewTodoRepository(db)
	userRepo := repository.NewUserRepository(db) // الـ Repo الجديد

	// 3. Identify New Handlers
	todoHandler := handler.NewTodoHandler(todoRepo)
	authHandler := handler.NewAuthHandler(userRepo) // الـ Handler الجديد

	// 4. Handlers Setup
	r := router.SetupRouter(todoHandler, authHandler)

	fmt.Println("Server is running on port 8080...")
	r.Run(":8080")
}
