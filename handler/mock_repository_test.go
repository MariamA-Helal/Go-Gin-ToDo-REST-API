package handler

import (
	"example/ToDo/models"

	"github.com/stretchr/testify/mock"
)

// ==========================================
// 1. MOCK REPOSITORY SETUP
// ==========================================

// MockTodoRepository isolates the handler from the actual database
type MockTodoRepository struct {
	mock.Mock
}

func (m *MockTodoRepository) GetTodos(limit int, offset int) []models.Todo {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.Todo)
}

func (m *MockTodoRepository) GetTodoByID(id int) (models.Todo, error) {
	args := m.Called(id)
	return args.Get(0).(models.Todo), args.Error(1)
}

func (m *MockTodoRepository) GetTodosByCategory(category string) []models.Todo {
	args := m.Called(category)
	return args.Get(0).([]models.Todo)
}

func (m *MockTodoRepository) GetTodosByStatus(status string) []models.Todo {
	args := m.Called(status)
	return args.Get(0).([]models.Todo)
}

func (m *MockTodoRepository) GetTodosBySearch(title string) []models.Todo {
	args := m.Called(title)
	return args.Get(0).([]models.Todo)
}

func (m *MockTodoRepository) CreateTodo(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

func (m *MockTodoRepository) EditTodo(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

func (m *MockTodoRepository) UpdateTodoStatus(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

func (m *MockTodoRepository) UpdateTodosByCategory(userID uint, category string, completed bool) error {
	args := m.Called(userID, category, completed)
	return args.Error(0)
}

func (m *MockTodoRepository) DeleteTodo(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTodoRepository) DeleteAllTodos() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTodoRepository) DeleteTodosByCategory(category string) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockTodoRepository) GetUserIDByUsername(username string) (uint, error) {
	args := m.Called(username)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockTodoRepository) CountUserTodos(userID uint) int64 {
	args := m.Called(userID)
	return args.Get(0).(int64)
}

func (m *MockTodoRepository) DeleteUserTodos(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockTodoRepository) DeleteAllTodosGlobal() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTodoRepository) DeleteCategoryForUser(userID uint, category string) error {
	args := m.Called(userID, category)
	return args.Error(0)
}

func (m *MockTodoRepository) DeleteCategoryGlobal(category string) error {
	args := m.Called(category)
	return args.Error(0)
}
