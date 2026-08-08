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
	GetTodosBySearch(title string) []models.Todo

	// 2. Create & Update Operations (Using Pointers for data modification)
	CreateTodo(todo *models.Todo) error
	EditTodo(todo *models.Todo) error
	UpdateTodoStatus(todo *models.Todo) error
	UpdateTodosByCategory(category string, updatedData map[string]interface{}) error

	// 3. Delete Operations
	DeleteTodo(id int) error
	DeleteAllTodos() error
	DeleteTodosByCategory(category string) error
}
