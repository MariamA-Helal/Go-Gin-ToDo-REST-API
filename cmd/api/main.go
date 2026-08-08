package main

import (
	"example/ToDo/database"
	"example/ToDo/handler"
	"example/ToDo/repository"
	"example/ToDo/router"
)

func main() {
	// 1. Database connection
	db := database.ConnectDB()

	// 2. Dependency Injection Setup
	realRepo := repository.NewTodoRepository(db)

	todoHandler := handler.NewTodoHandler(realRepo)

	// 3. Router setup
	r := router.SetupRouter(todoHandler)

	// 4. Start the server
	r.Run(":8080")
}
