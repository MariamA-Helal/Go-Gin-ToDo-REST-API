package router

import (
	_ "example/ToDo/docs"

	"example/ToDo/handler"
	"example/ToDo/middleware"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func SetupRouter(h *handler.TodoHandler, authHandler *handler.AuthHandler) *gin.Engine {
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Authentication routes
	router.POST("/signup", authHandler.Signup)
	router.POST("/login", authHandler.Login)

	todos := router.Group("/todos")
	todos.Use(middleware.RequireAuth())
	{
		todos := router.Group("/todos")
		todos.Use(middleware.RequireAuth())
		{
			// 1. GET
			todos.GET("", h.GetTodos)
			todos.GET("/:id", h.GetTodoByID)
			todos.GET("/category/:category", h.GetTodosByCategory)
			todos.GET("/status/:status", h.GetTodosByStatus)
			todos.GET("/search", h.GetTodosBySearch)

			// 2. POST
			todos.POST("", h.CreateTodo)

			// 3. PUT
			todos.PUT("/:id", h.EditTodo)
			todos.PUT("/category/:category", h.UpdateTodosByCategory)

			// 4. PATCH
			todos.PATCH("/:id/status", h.UpdateTodoStatus)

			// 5. DELETE
			todos.DELETE("/:id", h.DeleteTodo)
			todos.DELETE("", h.DeleteAllTodos)
			todos.DELETE("/category/:category", h.DeleteTodosByCategory)

			// 6. Admin
			userUpgradeRoutes := router.Group("/user")
			userUpgradeRoutes.Use(middleware.RequireAuth())
			{
				userUpgradeRoutes.POST("/request-upgrade", authHandler.RequestUpgrade)
				userUpgradeRoutes.GET("/my-secret-key", authHandler.GetMySecretKey)
				userUpgradeRoutes.POST("/upgrade", authHandler.UpgradeToAdmin)
			}

			// 7. Master
			masterRoutes := router.Group("/master")
			masterRoutes.Use(middleware.RequireAuth())
			{
				masterRoutes.PUT("/approve-upgrade/:id", authHandler.ApproveUpgrade)
			}
		}
	}

	return router
}
