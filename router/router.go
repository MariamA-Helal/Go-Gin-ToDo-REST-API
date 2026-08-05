package router

import (
	"example/ToDo/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/todos", handler.GetTodos)
	router.GET("/todos/:id", handler.GetTodoByID)
	router.GET("/todos/category/:category", handler.GetTodosByCategory)
	router.GET("/todos/status/:status", handler.GetTodoByStatus)
	router.GET("/todos/search", handler.GetTodosBySearch)

	router.POST("/todos", handler.CreateTodo)
	router.PUT("/todos/:id", handler.EditTodo)
	router.PUT("/todos/category/:category", handler.UpdateTodosByCategory)

	router.DELETE("/todos/:id", handler.DeleteTodo)
	router.DELETE("/todos/", handler.DeleteAllTodo)

	router.PATCH("/todos/:id/status", handler.UpdateTodoStatus)

	return router
}
