package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// ==========================================
// 2. TEST HELPERS
// ==========================================

func init() {
	// Suppress Gin debug output during testing
	gin.SetMode(gin.TestMode)
}

// setupTestEngine configures Gin and binds the mock repository
func setupTestEngine(mockRepo *MockTodoRepository, userID uint, role string) *gin.Engine {
	h := NewTodoHandler(mockRepo)
	router := gin.Default()

	// Dynamic Role and ID values
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	})

	// 1. GET
	router.GET("/todos", h.GetTodos)
	router.GET("/todos/:id", h.GetTodoByID)
	router.GET("/todos/category/:category", h.GetTodosByCategory)
	router.GET("/todos/status/:status", h.GetTodosByStatus)
	router.GET("/todos/search", h.GetTodosBySearch)

	// 2. POST
	router.POST("/todos", h.CreateTodo)

	// 3. PUT
	router.PUT("/todos/:id", h.EditTodo)
	router.PUT("/todos/category/:category", h.UpdateTodosByCategory)

	// 4. PATCH
	router.PATCH("/todos/:id/status", h.UpdateTodoStatus)

	// 5. DELETE
	router.DELETE("/todos/:id", h.DeleteTodo)
	router.DELETE("/todos", h.DeleteAllTodos)
	router.DELETE("/todos/category/:category", h.DeleteTodosByCategory)

	return router
}

// performRequest executes the HTTP request and returns the recorder
func performRequest(r http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		// If body is a string (e.g., for bad JSON), write it directly
		if strBody, ok := body.(string); ok {
			buf.WriteString(strBody)
		} else {
			_ = json.NewEncoder(&buf).Encode(body)
		}
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
