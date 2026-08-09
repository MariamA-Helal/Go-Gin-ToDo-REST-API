package handler

import (
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example/ToDo/models"
	"example/ToDo/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 1. Dependency Injection Setup
type TodoHandler struct {
	Repo TodoRepository
}

func NewTodoHandler(repo TodoRepository) *TodoHandler {
	return &TodoHandler{Repo: repo}
}

// ==========================================
// GET METHODS
// ==========================================

func (h *TodoHandler) GetTodos(c *gin.Context) {

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")

	limit, offset := 10, 0

	var todos []models.Todo
	repoImpl, ok := h.Repo.(*repository.TodoRepositoryImpl)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if userRole.(string) == "admin" {
		repoImpl.DB.Limit(limit).Offset(offset).Find(&todos)
	} else {
		repoImpl.DB.Where("user_id = ?", userID.(uint)).Limit(limit).Offset(offset).Find(&todos)
	}

	c.IndentedJSON(http.StatusOK, todos)
}

func (h *TodoHandler) GetTodoByID(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	todo, err := h.Repo.GetTodoByID(idInt)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	if todo.UserID != userID.(uint) && userRole.(string) != "admin" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Not your todo"})
		return
	}

	c.IndentedJSON(http.StatusOK, todo)
}

func (h *TodoHandler) GetTodosByCategory(c *gin.Context) {
	category := c.Param("category")

	// Validations based on your tests
	if len(strings.TrimSpace(category)) <= 1 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Category is too short"})
		return
	}
	if _, err := strconv.Atoi(category); err == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Category cannot be only numbers"})
		return
	}
	if len(strings.Fields(category)) > 3 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Category exceeds 3 words limit"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	var todos []models.Todo

	repoImpl, ok := h.Repo.(*repository.TodoRepositoryImpl)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if userRole.(string) == "admin" {
		repoImpl.DB.Where("category = ?", category).Find(&todos)
	} else {
		repoImpl.DB.Where("user_id = ? AND category = ?", userID.(uint), category).Find(&todos)
	}

	c.IndentedJSON(http.StatusOK, todos)
}

func (h *TodoHandler) GetTodosByStatus(c *gin.Context) {
	statusStr := c.Param("status")
	statusBool, err := strconv.ParseBool(statusStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Status must be a boolean (true or false)"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	var todos []models.Todo

	repoImpl, ok := h.Repo.(*repository.TodoRepositoryImpl)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if userRole.(string) == "admin" {
		repoImpl.DB.Where("completed = ?", statusBool).Find(&todos)
	} else {
		repoImpl.DB.Where("user_id = ? AND completed = ?", userID.(uint), statusBool).Find(&todos)
	}

	c.IndentedJSON(http.StatusOK, todos)

}

func (h *TodoHandler) GetTodosBySearch(c *gin.Context) {
	query := c.Query("q")
	if strings.TrimSpace(query) == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Search query cannot be empty"})
		return
	}
	if len(strings.Fields(query)) > 10 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Search query exceeds 10 words limit"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	var todos []models.Todo

	repoImpl, ok := h.Repo.(*repository.TodoRepositoryImpl)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	searchQuery := "%" + query + "%"
	if userRole.(string) == "admin" {
		repoImpl.DB.Where("title ILIKE ? OR category ILIKE ?", searchQuery, searchQuery).Find(&todos)
	} else {
		repoImpl.DB.Where("user_id = ? AND (title ILIKE ? OR category ILIKE ?)", userID.(uint), searchQuery, searchQuery).Find(&todos)
	}

	c.IndentedJSON(http.StatusOK, todos)

}

// ==========================================
// POST METHOD (CREATE)
// ==========================================
// 1. الـ DTO بتاع الـ Create (بنستقبل البيانات دي بس)
type CreateTodoDTO struct {
	Title    string     `json:"title" binding:"required"`
	Category string     `json:"category" binding:"required"`
	Priority string     `json:"priority" binding:"required"`
	DueDate  *time.Time `json:"due_date"`
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req CreateTodoDTO
	//1. Binding
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Malformed JSON or missing required fields"})
		return
	}

	// 2. Sanitization
	safeTitle := html.EscapeString(strings.TrimSpace(req.Title))
	safeCategory := html.EscapeString(strings.TrimSpace(req.Category))

	// 3. Validations
	if len(safeTitle) <= 1 || len(strings.Fields(safeTitle)) >= 10 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid title length"})
		return
	}
	if _, err := strconv.Atoi(safeTitle); err == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Title cannot be only numbers"})
		return
	}

	if len(safeCategory) <= 1 || len(strings.Fields(safeCategory)) > 3 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid category format"})
		return
	}
	if _, err := strconv.Atoi(safeCategory); err == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Category cannot be only numbers"})
		return
	}

	if req.Priority != "Low" && req.Priority != "Medium" && req.Priority != "High" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Priority must be 'Low', 'Medium', or 'High'"})
		return
	}

	if req.DueDate != nil && req.DueDate.Before(time.Now().UTC()) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Due date cannot be in the past"})
		return
	}

	userID, _ := c.Get("user_id")

	newTodo := models.Todo{
		Title:    safeTitle,
		Category: safeCategory,
		Priority: req.Priority,
		DueDate:  req.DueDate,
		UserID:   userID.(uint),
	}

	err := h.Repo.CreateTodo(&newTodo)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save todo"})
		return
	}

	c.IndentedJSON(http.StatusCreated, newTodo)
}

// ==========================================
// PUT & PATCH METHODS (UPDATE)
// ==========================================

func (h *TodoHandler) EditTodo(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var editData models.Todo
	if err := c.BindJSON(&editData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

	if len(strings.TrimSpace(editData.Title)) == 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	todoToUpdate, err := h.Repo.GetTodoByID(idInt)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	if todoToUpdate.UserID != userID.(uint) && userRole.(string) != "admin" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "You can only edit your own tasks"})
		return
	}

	todoToUpdate.Title = editData.Title
	todoToUpdate.Category = editData.Category
	todoToUpdate.Priority = editData.Priority

	if editData.Completed && !todoToUpdate.Completed {
		now := time.Now().UTC()
		todoToUpdate.CompletedAt = &now
	} else if !editData.Completed && todoToUpdate.Completed {
		todoToUpdate.CompletedAt = nil
	}
	todoToUpdate.Completed = editData.Completed

	h.Repo.EditTodo(&todoToUpdate)
	c.IndentedJSON(http.StatusOK, todoToUpdate)
}

func (h *TodoHandler) UpdateTodoStatus(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var status struct {
		Completed *bool `json:"completed"`
	}

	if err := c.BindJSON(&status); err != nil || status.Completed == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid body or missing status"})
		return
	}

	todoToUpdate, err := h.Repo.GetTodoByID(idInt)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	if todoToUpdate.UserID != userID.(uint) && userRole.(string) != "admin" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "You can only edit your own tasks"})
		return
	}

	if *status.Completed && !todoToUpdate.Completed {
		now := time.Now().UTC()
		todoToUpdate.CompletedAt = &now
	} else if !*status.Completed && todoToUpdate.Completed {
		todoToUpdate.CompletedAt = nil
	}

	todoToUpdate.Completed = *status.Completed
	h.Repo.UpdateTodoStatus(&todoToUpdate)

	c.IndentedJSON(http.StatusOK, todoToUpdate)
}

func (h *TodoHandler) UpdateTodosByCategory(c *gin.Context) {
	category := c.Param("category")

	var updateData map[string]interface{}
	if err := c.BindJSON(&updateData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

	val, exists := updateData["completed"]
	if !exists {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Missing 'completed' field"})
		return
	}

	if len(strings.TrimSpace(category)) <= 1 || len(strings.Fields(category)) > 3 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid category format"})
		return
	}
	if _, err := strconv.Atoi(category); err == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Category cannot be only numbers"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")

	repoImpl, ok := h.Repo.(*repository.TodoRepositoryImpl)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var tx *gorm.DB
	if userRole.(string) == "admin" {
		tx = repoImpl.DB.Model(&models.Todo{}).Where("category = ?", category).Update("completed", val)
	} else {
		tx = repoImpl.DB.Model(&models.Todo{}).Where("user_id = ? AND category = ?", userID.(uint), category).Update("completed", val)
	}

	if tx.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Todos category updated"})
}

// ==========================================
// DELETE METHODS
// ==========================================
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	todoToDelete, err := h.Repo.GetTodoByID(idInt)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	if todoToDelete.UserID != userID.(uint) && userRole.(string) != "admin" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Forbidden: You don't have permission to modify this todo"})
		return
	}

	err = h.Repo.DeleteTodo(idInt)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Todo deleted"})
}

func (h *TodoHandler) DeleteAllTodos(c *gin.Context) {

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")

	repoImpl, ok := h.Repo.(*repository.TodoRepositoryImpl)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var err error
	if userRole.(string) == "admin" {
		err = repoImpl.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Todo{}).Error
	} else {
		err = repoImpl.DB.Where("user_id = ?", userID.(uint)).Delete(&models.Todo{}).Error
	}

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todos"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Todos deleted successfully"})

}

func (h *TodoHandler) DeleteTodosByCategory(c *gin.Context) {
	category := c.Param("category")

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")

	repoImpl, ok := h.Repo.(*repository.TodoRepositoryImpl)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var tx *gorm.DB
	if userRole.(string) == "admin" {
		tx = repoImpl.DB.Where("category = ?", category).Delete(&models.Todo{})
	} else {
		tx = repoImpl.DB.Where("user_id = ? AND category = ?", userID.(uint), category).Delete(&models.Todo{})
	}

	if tx.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	if len(strings.TrimSpace(category)) <= 1 || len(strings.Fields(category)) > 3 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid category format"})
		return
	}
	if _, err := strconv.Atoi(category); err == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Category cannot be only numbers"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
