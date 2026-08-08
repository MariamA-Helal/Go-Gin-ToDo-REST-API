package router

import (
	"example/ToDo/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(h *handler.TodoHandler) *gin.Engine {
	router := gin.Default()

	// 1. GET
	router.GET("/todos", h.GetTodos)
	router.GET("/todos/:id", h.GetTodoByID)
	router.GET("/todos/category/:category", h.GetTodosByCategory)
	router.GET("/todos/status/:status", h.GetTodosByStatus)
	router.GET("/todos/search", h.GetTodosBySearch)

	// 2. POST
	router.POST("/todos", h.CreateTodo)

	// 3. PUT
	router.PUT("/todos/:id", h.EditTodo)
	router.PUT("/todos/category/:category", h.UpdateTodosByCategory)

	// 4. PATCH
	router.PATCH("/todos/:id/status", h.UpdateTodoStatus)

	// 5. DELETE
	router.DELETE("/todos/:id", h.DeleteTodo)
	router.DELETE("/todos", h.DeleteAllTodos)
	router.DELETE("/todos/category/:category", h.DeleteTodosByCategory)

	return router
}
