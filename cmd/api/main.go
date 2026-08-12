package main

import (
	"example/ToDo/database"
	"example/ToDo/handler"
	"example/ToDo/repository"
	"example/ToDo/router"
	"fmt"
)

// @title           Todo Application API
// @version         1.0
// @description     A robust Todo REST API with JWT Auth and RBAC.
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Seeding the database and connecting to it
	db := database.ConnectDB()

	// 2. Repositories
	todoRepo := repository.NewTodoRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 3. Identify New Handlers
	todoHandler := handler.NewTodoHandler(todoRepo)
	authHandler := handler.NewAuthHandler(userRepo)

	// 4. Handlers Setup
	r := router.SetupRouter(todoHandler, authHandler)

	fmt.Println("Server is running on port 8080...")
	r.Run(":8080")
}

// ثالثاً: حل مشكلة الـ Wiring في main.go
//البشمهندس يقصد إنك لازم تمرري الـ AuthHandler والـ TodoHandler مع بعض لدالة SetupRouter.
//التعديل في ملف cmd/api/main.go:
//طيب انا بالفعل عامله كده اصلا تفتكر في مشكله ايه تانيه ناقصه
