package main

import (
	"example/ToDo/database"
	"example/ToDo/router"
)

func main() {
	// 1.Database connection
	database.ConnectDB()

	// 2.Router setup
	r := router.SetupRouter()

	// 3.Start the server
	r.Run(":8080")
}
