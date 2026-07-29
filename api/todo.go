package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// In-memory storage initialized as empty (Requirement 4)
var todos = []todo{}

// Tracking next ID globally to prevent ID collisions after deletions
var nextID = 1

func GetTodos(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, todos)
}

func GetTodoByID(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	for _, t := range todos {
		if t.ID == idInt {
			c.IndentedJSON(http.StatusOK, t)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
}

func CreatTodos(c *gin.Context) {

	var newTodo todo

	if err := c.BindJSON(&newTodo); err != nil {
		return
	}

	if newTodo.Title == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	// Assign an incremental ID safely
	newTodo.ID = nextID
	nextID++

	todos = append(todos, newTodo)
	c.IndentedJSON(http.StatusCreated, newTodo)
}

func EditTodo(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var editData todo

	if err := c.BindJSON(&editData); err != nil {
		return
	}

	if editData.Title == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	// Full update using PUT
	for i, t := range todos {
		if t.ID == idInt {
			todos[i].Title = editData.Title
			todos[i].Completed = editData.Completed
			c.IndentedJSON(http.StatusOK, todos[i])
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
}

// BONUS: PATCH endpoint for partial updates (Updating status only)
func UpdateTodo(c *gin.Context) {
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

	for i, t := range todos {
		if t.ID == idInt {
			todos[i].Completed = status.Completed
			c.IndentedJSON(http.StatusOK, todos[i])
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})

}

func DeleteTodo(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	for i, t := range todos {
		if t.ID == idInt {
			todos = append(todos[:i], todos[i+1:]...)
			c.IndentedJSON(http.StatusOK, gin.H{"message": "Todo deleted"})
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
}
