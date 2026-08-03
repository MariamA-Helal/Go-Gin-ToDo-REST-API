package repository

import (
	"example/ToDo/database"
	"example/ToDo/models"
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

func CreateTodo(todo *models.Todo) {
	database.DB.Create(todo)
}

func EditTodo(todo *models.Todo) {
	database.DB.Save(todo)
}

func UpdateTodoStatus(todo *models.Todo, completed bool) {
	database.DB.Model(todo).Update("completed", completed)
}

func DeleteTodo(todo *models.Todo) {
	database.DB.Delete(todo)
}
