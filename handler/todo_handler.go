package handler

import (
	"net/http"
	"strconv"
	"time"

	"example/ToDo/models"
	"example/ToDo/repository"

	"github.com/gin-gonic/gin"
)

func GetTodos(c *gin.Context) {
	todos := repository.GetTodos()
	c.IndentedJSON(http.StatusOK, todos)
}

func GetTodoByID(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	todo, err := repository.GetTodoByID(idInt)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, todo)
}

func GetTodosByCategory(c *gin.Context) {
	category := c.Param("category")
	todos := repository.GetTodosByCategory(category)
	c.IndentedJSON(http.StatusOK, todos)
}

func GetTodoByStatus(c *gin.Context) {
	status := c.Param("status")
	todos := repository.GetTodosByStatus(status)
	c.IndentedJSON(http.StatusOK, todos)
}

func GetTodosBySearch(c *gin.Context) {
	query := c.Query("q")
	todos := repository.GetTodosBySearch(query)
	c.IndentedJSON(http.StatusOK, todos)
}

func CreateTodo(c *gin.Context) {

	var newTodo models.Todo

	if err := c.BindJSON(&newTodo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Missing Required Fields or Invalid Data Format"})
		return
	}

	if newTodo.DueDate != nil && newTodo.DueDate.Before(time.Now().UTC()) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Due date cannot be before creation date"})
		return
	}

	if newTodo.Priority != "Low" && newTodo.Priority != "Medium" && newTodo.Priority != "High" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Priority must be 'Low', 'Medium', or 'High'"})
		return
	}

	if newTodo.Completed {
		now := time.Now().UTC()
		newTodo.CompletedAt = &now
	}

	err := repository.CreateTodo(&newTodo)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save todo to the database"})
		return
	}
	c.IndentedJSON(http.StatusCreated, newTodo)
}

func EditTodo(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var editData models.Todo

	if err := c.BindJSON(&editData); err != nil {
		return
	}

	if editData.Title == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	// Full update using PUT
	todoToUpdate, err := repository.GetTodoByID(idInt)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	todoToUpdate.Title = editData.Title
	todoToUpdate.Completed = editData.Completed
	repository.EditTodo(&todoToUpdate)

	c.IndentedJSON(http.StatusOK, todoToUpdate)
}

// BONUS: PATCH endpoint for partial updates (Updating status only)
func UpdateTodoStatus(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var status struct {
		Completed bool `json:"completed"`
	}

	if err := c.BindJSON(&status); err != nil {
		return
	}

	todoToUpdate, err := repository.GetTodoByID(idInt)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	repository.UpdateTodoStatus(&todoToUpdate, status.Completed)
	updatedTodo, _ := repository.GetTodoByID(idInt)
	c.IndentedJSON(http.StatusOK, updatedTodo)
}

func UpdateTodosByCategory(c *gin.Context) {
	category := c.Param("category")
	var updateData struct {
		Completed bool `json:"completed"`
	}

	if err := c.BindJSON(&updateData); err != nil {
		return
	}

	todos := repository.GetTodosByCategory(category)
	for _, todo := range todos {
		repository.UpdateTodoStatus(&todo, updateData.Completed)
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Todos updated"})
}

func DeleteTodo(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	todoToDelete, err := repository.GetTodoByID(idInt)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	repository.DeleteTodo(&todoToDelete)
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Todo deleted"})
}

func DeleteAllTodo(c *gin.Context) {
	repository.DeleteAllTodos()
	c.IndentedJSON(http.StatusOK, gin.H{"message": "All todos deleted"})
}
