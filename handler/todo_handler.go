package handler

import (
	"net/http"
	"strconv"

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

func CreateTodo(c *gin.Context) {

	var newTodo models.Todo

	if err := c.BindJSON(&newTodo); err != nil {
		return
	}

	if newTodo.Title == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	repository.CreateTodo(&newTodo)
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
