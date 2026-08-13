package handler

import (
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example/ToDo/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminTodoResponse
type AdminTodoResponse struct {
	ID            uint   `json:"real_id"`
	OwnerID       uint   `json:"owner_id"`
	OwnerUsername string `json:"owner_username"`
	TaskID        uint   `json:"task_id"`
	Title         string `json:"title"`
	Category      string `json:"category"`
	Priority      string `json:"priority"`
	Completed     bool   `json:"completed"`
}

// UserTodoResponse
type UserTodoResponse struct {
	TaskID    uint   `json:"task_id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Priority  string `json:"priority"`
	Completed bool   `json:"completed"`
}

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

// / GetTodos      godoc
// @Summary      Get all todos
// @Description  Retrieves a list of todos for the authenticated user
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit     query  int     false  "Pagination limit"
// @Param        page      query  int     false  "Pagination page"
// @Param        scope     query  string  false  "Set to 'personal' to get only your tasks (Admin only)"
// @Success      200       {object}  interface{}
// @Router       /todos [get]
func (h *TodoHandler) GetTodos(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	scope := c.Query("scope")

	limit, offset := 10, 0

	if limitQuery := c.Query("limit"); limitQuery != "" {
		if parsedLimit, err := strconv.Atoi(limitQuery); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100 // Clamping
			}
		}
	}

	if pageQuery := c.Query("page"); pageQuery != "" {
		if parsedPage, err := strconv.Atoi(pageQuery); err == nil && parsedPage > 0 {
			offset = (parsedPage - 1) * limit
		}
	}
	todos := h.Repo.GetTodos(limit, offset)

	if userRole.(string) == "admin" && scope != "personal" {
		var responses []AdminTodoResponse = []AdminTodoResponse{}
		for _, t := range todos {
			responses = append(responses, AdminTodoResponse{
				ID:            t.ID,
				OwnerID:       t.UserID,
				OwnerUsername: t.User.Username,
				TaskID:        t.UserTaskID,
				Title:         t.Title,
				Category:      t.Category,
				Priority:      t.Priority,
				Completed:     t.Completed,
			})
		}
		c.IndentedJSON(http.StatusOK, responses)
		return
	}

	var personalTodos []models.Todo
	for _, t := range todos {
		if t.UserID == userID.(uint) {
			personalTodos = append(personalTodos, t)
		}
	}

	if userRole.(string) == "admin" {
		var responses []AdminTodoResponse = []AdminTodoResponse{}
		for _, t := range personalTodos {
			responses = append(responses, AdminTodoResponse{
				ID:            t.ID,
				OwnerID:       t.UserID,
				OwnerUsername: t.User.Username,
				TaskID:        t.UserTaskID,
				Title:         t.Title,
				Category:      t.Category,
				Priority:      t.Priority,
				Completed:     t.Completed,
			})
		}
		c.IndentedJSON(http.StatusOK, responses)
	} else {
		var responses []UserTodoResponse = []UserTodoResponse{}
		for _, t := range personalTodos {
			responses = append(responses, UserTodoResponse{
				TaskID:    t.UserTaskID,
				Title:     t.Title,
				Category:  t.Category,
				Priority:  t.Priority,
				Completed: t.Completed,
			})
		}
		c.IndentedJSON(http.StatusOK, responses)
	}
}

// GetTodoByID   godoc
// @Summary      Get todo by ID
// @Description  Retrieves a specific todo by its ID for the authenticated user
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id        path   int     true   "Todo ID"
// @Param        scope     query  string  false  "Set to 'personal' to get only your tasks (Admin only)"
// @Success      200       {object}  interface{}
// @Failure      404       {object}  map[string]string
// @Router       /todos/{id} [get]
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

	if userRole.(string) == "admin" {
		response := AdminTodoResponse{
			ID:            todo.ID,
			OwnerID:       todo.UserID,
			OwnerUsername: todo.User.Username,
			TaskID:        todo.UserTaskID,
			Title:         todo.Title,
			Category:      todo.Category,
			Priority:      todo.Priority,
			Completed:     todo.Completed,
		}
		c.IndentedJSON(http.StatusOK, response)
	} else {
		response := UserTodoResponse{
			TaskID:    todo.UserTaskID,
			Title:     todo.Title,
			Category:  todo.Category,
			Priority:  todo.Priority,
			Completed: todo.Completed,
		}
		c.IndentedJSON(http.StatusOK, response)
	}
}

// GetTodosByCategory godoc
// @Summary      Get todos by category
// @Description  Admins get all in category (or personal if ?scope=personal). Users get only their own.
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        category  path   string  true   "Todo Category"
// @Param        scope     query  string  false  "Set to 'personal' to get only your tasks (Admin only)"
// @Success      200       {object}  interface{}
// @Failure      400       {object}  map[string]string
// @Router       /todos/category/{category} [get]
func (h *TodoHandler) GetTodosByCategory(c *gin.Context) {
	category := c.Param("category")

	// 1. Validations First
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

	// 2. Auth Context
	role, _ := c.Get("role")
	userID, _ := c.Get("user_id")
	scope := c.Query("scope")

	// 3. Fetch from Repo directly (بدل dbQuery)
	todos := h.Repo.GetTodosByCategory(category)

	// 4. Role-based Logic & DTO Mapping
	if role == "admin" && scope != "personal" {
		// Admin sees all in category
		var responses []AdminTodoResponse = []AdminTodoResponse{}
		for _, t := range todos {
			responses = append(responses, AdminTodoResponse{
				ID:            t.ID,
				OwnerID:       t.UserID,
				OwnerUsername: t.User.Username,
				TaskID:        t.UserTaskID,
				Title:         t.Title,
				Category:      t.Category,
				Priority:      t.Priority,
				Completed:     t.Completed,
			})
		}
		c.JSON(http.StatusOK, responses)
		return
	}

	var personalTodos []models.Todo
	for _, t := range todos {
		if t.UserID == userID.(uint) {
			personalTodos = append(personalTodos, t)
		}
	}

	if role == "admin" {
		var responses []AdminTodoResponse = []AdminTodoResponse{}
		for _, t := range personalTodos {
			responses = append(responses, AdminTodoResponse{
				ID:            t.ID,
				OwnerID:       t.UserID,
				OwnerUsername: t.User.Username,
				TaskID:        t.UserTaskID,
				Title:         t.Title,
				Category:      t.Category,
				Priority:      t.Priority,
				Completed:     t.Completed,
			})
		}
		c.JSON(http.StatusOK, responses)
	} else {
		var responses []UserTodoResponse = []UserTodoResponse{}
		for _, t := range personalTodos {
			responses = append(responses, UserTodoResponse{
				TaskID:    t.UserTaskID,
				Title:     t.Title,
				Category:  t.Category,
				Priority:  t.Priority,
				Completed: t.Completed,
			})
		}
		c.JSON(http.StatusOK, responses)
	}
}

// GetTodosByStatus godoc
// @Summary      Get todos by status
// @Description  Retrieves a list of todos filtered by their completion status
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        status  path   string  true   "Todo Status (true or false)"
// @Param        scope   query  string  false  "Set to 'personal' to get only your tasks (Admin only)"
// @Success      200     {object}  interface{}
// @Failure      400     {object}  map[string]string
// @Router       /todos/status/{status} [get]
func (h *TodoHandler) GetTodosByStatus(c *gin.Context) {
	statusStr := c.Param("status")

	// 1. Validation First
	_, err := strconv.ParseBool(statusStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Status must be a boolean (true or false)"})
		return
	}

	// 2. Auth Context
	role, _ := c.Get("role")
	userID, _ := c.Get("user_id")
	scope := c.Query("scope")

	// 3. Fetch from Repo directly (بدون استخدام DB Query)
	todos := h.Repo.GetTodosByStatus(statusStr)

	// 4. Role-based Logic & DTO Mapping
	if role == "admin" && scope != "personal" {
		// Admin sees all
		var responses []AdminTodoResponse = []AdminTodoResponse{}
		for _, t := range todos {
			responses = append(responses, AdminTodoResponse{
				ID:            t.ID,
				OwnerID:       t.UserID,
				OwnerUsername: t.User.Username,
				TaskID:        t.UserTaskID,
				Title:         t.Title,
				Category:      t.Category,
				Priority:      t.Priority,
				Completed:     t.Completed,
			})
		}
		c.JSON(http.StatusOK, responses)
		return
	}

	var personalTodos []models.Todo
	for _, t := range todos {
		if t.UserID == userID.(uint) {
			personalTodos = append(personalTodos, t)
		}
	}

	if role == "admin" {
		var responses []AdminTodoResponse = []AdminTodoResponse{}
		for _, t := range personalTodos {
			responses = append(responses, AdminTodoResponse{
				ID:            t.ID,
				OwnerID:       t.UserID,
				OwnerUsername: t.User.Username,
				TaskID:        t.UserTaskID,
				Title:         t.Title,
				Category:      t.Category,
				Priority:      t.Priority,
				Completed:     t.Completed,
			})
		}
		c.JSON(http.StatusOK, responses)
	} else {
		var responses []UserTodoResponse = []UserTodoResponse{}
		for _, t := range personalTodos {
			responses = append(responses, UserTodoResponse{
				TaskID:    t.UserTaskID,
				Title:     t.Title,
				Category:  t.Category,
				Priority:  t.Priority,
				Completed: t.Completed,
			})
		}
		c.JSON(http.StatusOK, responses)
	}
}

// GetTodosBySearch godoc
// @Summary      Search todos by title
// @Description  Retrieves a list of todos matching the search query text
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        q         query  string  true   "Search query text"
// @Param        scope     query  string  false  "Set to 'personal' to get only your tasks (Admin only)"
// @Success      200       {object}  interface{}
// @Failure      400       {object}  map[string]string
// @Router       /todos/search [get]
func (h *TodoHandler) GetTodosBySearch(c *gin.Context) {
	query := c.Query("q")

	// 1. Validations First
	if strings.TrimSpace(query) == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Search query cannot be empty"})
		return
	}
	if len(strings.Fields(query)) > 10 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Search query exceeds 10 words limit"})
		return
	}

	// 2. Auth Context
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	scope := c.Query("scope")

	// 3. Fetch from Repo directly
	todos := h.Repo.GetTodosBySearch(query)

	// 4. Role-based Logic & DTO Mapping
	if userRole.(string) == "admin" && scope != "personal" {
		var responses []AdminTodoResponse = []AdminTodoResponse{}
		for _, t := range todos {
			responses = append(responses, AdminTodoResponse{
				ID:            t.ID,
				OwnerID:       t.UserID,
				OwnerUsername: t.User.Username,
				TaskID:        t.UserTaskID,
				Title:         t.Title,
				Category:      t.Category,
				Priority:      t.Priority,
				Completed:     t.Completed,
			})
		}
		c.IndentedJSON(http.StatusOK, responses)
		return
	}

	var personalTodos []models.Todo
	for _, t := range todos {
		if t.UserID == userID.(uint) {
			personalTodos = append(personalTodos, t)
		}
	}

	if userRole.(string) == "admin" {
		var responses []AdminTodoResponse = []AdminTodoResponse{}
		for _, t := range personalTodos {
			responses = append(responses, AdminTodoResponse{
				ID:            t.ID,
				OwnerID:       t.UserID,
				OwnerUsername: t.User.Username,
				TaskID:        t.UserTaskID,
				Title:         t.Title,
				Category:      t.Category,
				Priority:      t.Priority,
				Completed:     t.Completed,
			})
		}
		c.IndentedJSON(http.StatusOK, responses)
	} else {
		var responses []UserTodoResponse = []UserTodoResponse{}
		for _, t := range personalTodos {
			responses = append(responses, UserTodoResponse{
				TaskID:    t.UserTaskID,
				Title:     t.Title,
				Category:  t.Category,
				Priority:  t.Priority,
				Completed: t.Completed,
			})
		}
		c.IndentedJSON(http.StatusOK, responses)
	}
}

// ==========================================
// POST METHOD (CREATE)
// ==========================================
// 1. الـ DTO بتاع الـ Create (بنستقبل البيانات دي بس)
type CreateTodoDTO struct {
	Title          string     `json:"title" binding:"required"`
	Category       string     `json:"category" binding:"required"`
	Priority       string     `json:"priority" binding:"required"`
	DueDate        *time.Time `json:"due_date"`
	TargetUsername string     `json:"target_username,omitempty"`
}

// CreateTodo godoc
// @Summary      Create a new todo
// @Description  Creates a new todo task for the authenticated user or assigns it to another user (Admin only)
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input    body      CreateTodoDTO  true  "Todo Data"
// @Success      201      {object}  models.Todo
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /todos [post]
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
	userRole, _ := c.Get("role")

	finalUserID := userID.(uint)

	if strings.TrimSpace(req.TargetUsername) != "" {
		if userRole.(string) == "admin" {
			targetID, err := h.Repo.GetUserIDByUsername(req.TargetUsername)
			if err != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Target user not found"})
				return
			}
			finalUserID = targetID
		} else {
			c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Only admins can assign tasks to other users"})
			return
		}
	}

	newTodo := models.Todo{
		Title:    safeTitle,
		Category: safeCategory,
		Priority: req.Priority,
		DueDate:  req.DueDate,
		UserID:   finalUserID,
	}

	count := h.Repo.CountUserTodos(newTodo.UserID)
	newTodo.UserTaskID = uint(count) + 1

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

// EditTodoDTO
type EditTodoDTO struct {
	Title     string     `json:"title" binding:"required"`
	Category  string     `json:"category" binding:"required"`
	Priority  string     `json:"priority" binding:"required"`
	DueDate   *time.Time `json:"due_date"`
	Completed bool       `json:"completed"`
}

// EditTodo godoc
// @Summary      Edit a todo
// @Description  Updates an existing todo's full details by its ID
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int          true  "Todo ID"
// @Param        input    body      EditTodoDTO  true  "Updated Todo Data"
// @Success      200      {object}  models.Todo
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /todos/{id} [put]
func (h *TodoHandler) EditTodo(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req EditTodoDTO
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid body or missing required fields"})
		return
	}

	if len(strings.TrimSpace(req.Title)) == 0 {
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

	if req.Completed != todoToUpdate.Completed {

		if todoToUpdate.UserID != userID.(uint) {
			c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Only the task owner can change the completion status"})
			return
		}

		if req.Completed {
			now := time.Now().UTC()
			todoToUpdate.CompletedAt = &now
		} else {
			todoToUpdate.CompletedAt = nil
		}
		todoToUpdate.Completed = req.Completed
	}

	todoToUpdate.Title = req.Title
	todoToUpdate.Category = req.Category
	todoToUpdate.Priority = req.Priority
	todoToUpdate.DueDate = req.DueDate

	err = h.Repo.EditTodo(&todoToUpdate)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}

	c.IndentedJSON(http.StatusOK, todoToUpdate)
}

// UpdateTodoStatus godoc
// @Summary      Update a todo's status
// @Description  Partially updates the completion status of a specific todo by ID
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int             true  "Todo ID"
// @Param        input    body      map[string]bool true  "Status Update (e.g., {\"completed\": true})"
// @Success      200      {object}  models.Todo
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /todos/{id}/status [patch]
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

	if todoToUpdate.UserID != userID.(uint) {
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

	err = h.Repo.UpdateTodoStatus(&todoToUpdate)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.IndentedJSON(http.StatusOK, todoToUpdate)
}

// UpdateTodosByCategory godoc
// @Summary      Bulk update status by category
// @Description  Updates the completion status of your personal todos under a specific category
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        category path      string           true  "Todo Category"
// @Param        input    body      map[string]bool  true  "Status Update (e.g., {\"completed\": true})"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /todos/category/{category} [put]
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

	completed, ok := val.(bool)
	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "'completed' must be a boolean"})
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

	err := h.Repo.UpdateTodosByCategory(userID.(uint), category, completed)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Todos category updated"})
}

// ==========================================
// DELETE METHODS
// ==========================================

// DeleteTodo godoc
// @Summary      Delete a todo
// @Description  Deletes a specific todo by its ID (Admin can delete anyone's, User can delete only theirs)
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int  true  "Todo ID"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /todos/{id} [delete]
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

// DeleteAllTodos godoc
// @Summary      Delete all todos
// @Description  Deletes todos. Admins can delete all globally, their own (?scope=personal), or a specific user's (?target_username=...). Users delete only theirs.
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        scope            query  string  false  "Set to 'personal' to delete only your tasks (Admin only)"
// @Param        target_username  query  string  false  "Target username to delete their tasks (Admin only)"
// @Success      200              {object}  map[string]string
// @Failure      403              {object}  map[string]string
// @Failure      404              {object}  map[string]string
// @Router       /todos [delete]
func (h *TodoHandler) DeleteAllTodos(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	scope := c.Query("scope")
	targetUsername := c.Query("target_username")

	var err error

	if userRole.(string) == "admin" {

		if scope == "personal" {
			err = h.Repo.DeleteUserTodos(userID.(uint))
		} else if strings.TrimSpace(targetUsername) != "" {
			targetID, findErr := h.Repo.GetUserIDByUsername(targetUsername)

			if findErr != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Target user not found"})
				return
			}
			err = h.Repo.DeleteUserTodos(targetID)
		} else {
			err = h.Repo.DeleteAllTodosGlobal()
		}
	} else {
		if scope != "" || targetUsername != "" {
			c.IndentedJSON(http.StatusForbidden, gin.H{"error": "You don't have permission to use admin filters"})
			return
		}
		err = h.Repo.DeleteUserTodos(userID.(uint))
	}

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todos"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Todos deleted successfully"})
}

// DeleteTodosByCategory godoc
// @Summary      Delete todos by category
// @Description  Deletes todos under a category. Admins can delete globally, personally, or for a specific user.
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        category         path   string  true   "Todo Category"
// @Param        scope            query  string  false  "Set to 'personal' to delete only your tasks (Admin only)"
// @Param        target_username  query  string  false  "Target username to delete their tasks (Admin only)"
// @Success      200              {object}  map[string]string
// @Failure      400              {object}  map[string]string
// @Failure      403              {object}  map[string]string
// @Failure      404              {object}  map[string]string
// @Router       /todos/category/{category} [delete]
func (h *TodoHandler) DeleteTodosByCategory(c *gin.Context) {
	category := c.Param("category")

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
	scope := c.Query("scope")
	targetUsername := c.Query("target_username")

	var err error

	if userRole.(string) == "admin" {
		if scope == "personal" {
			err = h.Repo.DeleteCategoryForUser(userID.(uint), category)
		} else if strings.TrimSpace(targetUsername) != "" {
			targetID, findErr := h.Repo.GetUserIDByUsername(targetUsername)
			if findErr != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Target user not found"})
				return
			}
			err = h.Repo.DeleteCategoryForUser(targetID, category)
		} else {
			err = h.Repo.DeleteCategoryGlobal(category)
		}
	} else {
		if scope != "" || targetUsername != "" {
			c.IndentedJSON(http.StatusForbidden, gin.H{"error": "You don't have permission to use admin filters"})
			return
		}
		err = h.Repo.DeleteCategoryForUser(userID.(uint), category)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
