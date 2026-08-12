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
	UpdateTodosByCategory(userID uint, category string, completed bool) error

	DeleteTodo(id int) error
	DeleteAllTodos() error
	DeleteTodosByCategory(category string) error

	CountUserTodos(userID uint) int64
	GetUserIDByUsername(username string) (uint, error)

	DeleteUserTodos(userID uint) error
	DeleteAllTodosGlobal() error
	DeleteCategoryForUser(userID uint, category string) error
	DeleteCategoryGlobal(category string) error
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

func (r *TodoRepositoryImpl) UpdateTodosByCategory(userID uint, category string, completed bool) error {
	tx := r.DB.Model(&models.Todo{}).
		Where("user_id = ? AND category = ?", userID, category).
		Update("completed", completed)
	return tx.Error
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

func (r *TodoRepositoryImpl) CountUserTodos(userID uint) int64 {
	var count int64
	r.DB.Model(&models.Todo{}).Where("user_id = ?", userID).Count(&count)
	return count
}

func (r *TodoRepositoryImpl) GetUserIDByUsername(username string) (uint, error) {
	var user models.User
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (r *TodoRepositoryImpl) DeleteUserTodos(userID uint) error {
	return r.DB.Where("user_id = ?", userID).Delete(&models.Todo{}).Error
}
func (r *TodoRepositoryImpl) DeleteAllTodosGlobal() error {
	return r.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Todo{}).Error
}
func (r *TodoRepositoryImpl) DeleteCategoryForUser(userID uint, category string) error {
	return r.DB.Where("user_id = ? AND category = ?", userID, category).Delete(&models.Todo{}).Error
}
func (r *TodoRepositoryImpl) DeleteCategoryGlobal(category string) error {
	return r.DB.Where("category = ?", category).Delete(&models.Todo{}).Error
}
