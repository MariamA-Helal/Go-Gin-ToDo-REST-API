package repository

import (
	"example/ToDo/models"

	"gorm.io/gorm"
)

// 1. Struct and Constructor
type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) *todoRepository {
	return &todoRepository{db: db}
}

// ==========================================
// 2. READ OPERATIONS
// ==========================================

func (r *todoRepository) GetTodos(limit int, offset int) []models.Todo {
	var todos []models.Todo
	r.db.Limit(limit).Offset(offset).Find(&todos)
	return todos
}

func (r *todoRepository) GetTodoByID(id int) (models.Todo, error) {
	var todo models.Todo
	result := r.db.First(&todo, id)
	return todo, result.Error
}

func (r *todoRepository) GetTodosByCategory(category string) []models.Todo {
	var todos []models.Todo
	r.db.Where("category = ?", category).Find(&todos)
	return todos
}

func (r *todoRepository) GetTodosByStatus(status string) []models.Todo {
	var todos []models.Todo
	completed := status == "true"
	r.db.Where("completed = ?", completed).Find(&todos)
	return todos
}

func (r *todoRepository) GetTodosBySearch(query string) []models.Todo {
	var todos []models.Todo
	r.db.Where("title LIKE ?", "%"+query+"%").Find(&todos)
	return todos
}

// ==========================================
// 3. CREATE & UPDATE OPERATIONS
// ==========================================

func (r *todoRepository) CreateTodo(todo *models.Todo) error {
	return r.db.Create(todo).Error
}

func (r *todoRepository) EditTodo(todo *models.Todo) error {
	return r.db.Save(todo).Error
}

func (r *todoRepository) UpdateTodoStatus(todo *models.Todo) error {
	return r.db.Save(todo).Error
}

func (r *todoRepository) UpdateTodosByCategory(category string, updatedData map[string]interface{}) error {
	return r.db.Model(&models.Todo{}).Where("category = ?", category).Updates(updatedData).Error
}

// ==========================================
// 4. DELETE OPERATIONS
// ==========================================

func (r *todoRepository) DeleteTodo(id int) error {
	return r.db.Delete(&models.Todo{}, id).Error
}

func (r *todoRepository) DeleteAllTodos() error {
	return r.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Todo{}).Error
}

func (r *todoRepository) DeleteTodosByCategory(category string) error {
	return r.db.Where("category = ?", category).Delete(&models.Todo{}).Error
}
