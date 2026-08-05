package repository

import (
	"example/ToDo/database"
	"example/ToDo/models"

	"gorm.io/gorm"
)

func GetTodos() []models.Todo {
	var todos []models.Todo
	database.DB.Find(&todos)
	return todos
}

func GetTodoByID(id int) (models.Todo, error) {
	var todo models.Todo
	result := database.DB.First(&todo, id)
	return todo, result.Error
}

func GetTodosByCategory(category string) []models.Todo {
	var todos []models.Todo
	database.DB.Where("category = ?", category).Find(&todos)
	return todos
}

func GetTodosByStatus(status string) []models.Todo {
	var todos []models.Todo
	completed := false
	if status == "completed" {
		completed = true
	}
	database.DB.Where("completed = ?", completed).Find(&todos)
	return todos
}

func GetTodosBySearch(query string) []models.Todo {
	var todos []models.Todo
	database.DB.Where("title LIKE ?", "%"+query+"%").Find(&todos)
	return todos
}

func CreateTodo(todo *models.Todo) error {
	return database.DB.Create(todo).Error
}

func EditTodo(todo *models.Todo) {
	database.DB.Save(todo)
}

func UpdateTodoStatus(todo *models.Todo, completed bool) {
	database.DB.Model(todo).Update("completed", completed)
}

func UpdateTodosByCategory(category string, updatedData map[string]interface{}) {
	database.DB.Model(&models.Todo{}).Where("category = ?", category).Updates(updatedData)
}

func DeleteTodo(todo *models.Todo) {
	database.DB.Delete(todo)
}

func DeleteAllTodos() {
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Todo{})
}
