package main

import (
	"example/ToDo/api"
	
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/todos", api.GetTodos)
	router.GET("/todos/:id", api.GetTodoByID)
	router.POST("/todos", api.CreatTodos)
	router.PUT("/todos/:id", api.EditTodo)
	router.DELETE("/todos/:id", api.DeleteTodo)

	router.PATCH("/todos/:id/status", UpdateTodo)

	router.Run(":8080")
}
