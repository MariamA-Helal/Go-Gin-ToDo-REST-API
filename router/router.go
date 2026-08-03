package router

import (
	"example/ToDo/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/todos", handler.GetTodos)
	router.GET("/todos/:id", handler.GetTodoByID)
	router.POST("/todos", handler.CreateTodo)
	router.PUT("/todos/:id", handler.EditTodo)
	router.DELETE("/todos/:id", handler.DeleteTodo)
	router.PATCH("/todos/:id/status", handler.UpdateTodoStatus)

	return router
}
