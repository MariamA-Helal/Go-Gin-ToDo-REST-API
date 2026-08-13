package handler

import (
	"example/ToDo/models"
)

type TodoRepository interface {
	// 1. Read Operations
	GetTodos(limit int, offset int) []models.Todo
	GetTodoByID(id int) (models.Todo, error)
	GetTodosByCategory(category string) []models.Todo
	GetTodosByStatus(status string) []models.Todo
	GetTodosBySearch(query string) []models.Todo

	// 2. Create & Update Operations (Using Pointers for data modification)

	CreateTodo(todo *models.Todo) error
	EditTodo(todo *models.Todo) error
	UpdateTodoStatus(todo *models.Todo) error
	UpdateTodosByCategory(userID uint, category string, completed bool) error

	// 3. Delete Operations
	DeleteTodo(id int) error
	DeleteAllTodos() error
	DeleteTodosByCategory(category string) error

	DeleteUserTodos(userID uint) error
	DeleteAllTodosGlobal() error
	DeleteCategoryForUser(userID uint, category string) error
	DeleteCategoryGlobal(category string) error

	CountUserTodos(userID uint) int64
	GetUserIDByUsername(username string) (uint, error)
}
