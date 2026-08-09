package repository

import (
	"example/ToDo/models"

	"gorm.io/gorm"
)

// 1. Interface
type TodoRepository interface {
	GetTodos(limit int, offset int) []models.Todo
	GetTodoByID(id int) (models.Todo, error)
	GetTodosByCategory(category string) []models.Todo
	GetTodosByStatus(status string) []models.Todo
	GetTodosBySearch(query string) []models.Todo
	CreateTodo(todo *models.Todo) error
	EditTodo(todo *models.Todo) error
	UpdateTodoStatus(todo *models.Todo) error
	UpdateTodosByCategory(category string, updatedData map[string]interface{}) error
	DeleteTodo(id int) error
	DeleteAllTodos() error
	DeleteTodosByCategory(category string) error
}

// 2. Struct and Constructor
type TodoRepositoryImpl struct {
	DB *gorm.DB
}

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &TodoRepositoryImpl{DB: db}
}

// ==========================================
// 3. READ OPERATIONS
// ==========================================

func (r *TodoRepositoryImpl) GetTodos(limit int, offset int) []models.Todo {
	var todos []models.Todo
	r.DB.Limit(limit).Offset(offset).Find(&todos)
	return todos
}

func (r *TodoRepositoryImpl) GetTodoByID(id int) (models.Todo, error) {
	var todo models.Todo
	result := r.DB.First(&todo, id)
	return todo, result.Error
}

func (r *TodoRepositoryImpl) GetTodosByCategory(category string) []models.Todo {
	var todos []models.Todo
	r.DB.Where("category = ?", category).Find(&todos)
	return todos
}

func (r *TodoRepositoryImpl) GetTodosByStatus(status string) []models.Todo {
	var todos []models.Todo
	completed := status == "true"
	r.DB.Where("completed = ?", completed).Find(&todos)
	return todos
}

func (r *TodoRepositoryImpl) GetTodosBySearch(query string) []models.Todo {
	var todos []models.Todo
	r.DB.Where("title LIKE ?", "%"+query+"%").Find(&todos)
	return todos
}

// ==========================================
// 4. CREATE & UPDATE OPERATIONS
// ==========================================

func (r *TodoRepositoryImpl) CreateTodo(todo *models.Todo) error {
	return r.DB.Create(todo).Error
}

func (r *TodoRepositoryImpl) EditTodo(todo *models.Todo) error {
	return r.DB.Save(todo).Error
}

func (r *TodoRepositoryImpl) UpdateTodoStatus(todo *models.Todo) error {
	return r.DB.Save(todo).Error
}

func (r *TodoRepositoryImpl) UpdateTodosByCategory(category string, updatedData map[string]interface{}) error {
	return r.DB.Model(&models.Todo{}).Where("category = ?", category).Updates(updatedData).Error
}

// ==========================================
// 5. DELETE OPERATIONS
// ==========================================

func (r *TodoRepositoryImpl) DeleteTodo(id int) error {
	return r.DB.Delete(&models.Todo{}, id).Error
}

func (r *TodoRepositoryImpl) DeleteAllTodos() error {
	return r.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Todo{}).Error
}

func (r *TodoRepositoryImpl) DeleteTodosByCategory(category string) error {
	return r.DB.Where("category = ?", category).Delete(&models.Todo{}).Error
}
