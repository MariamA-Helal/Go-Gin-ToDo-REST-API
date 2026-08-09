package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"example/ToDo/middleware"
	"example/ToDo/utils"

	"github.com/gin-gonic/gin"
)

// Setup a dummy router with our middleware to test JWT and Roles
func setupMiddlewareTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	protected := r.Group("/protected")
	protected.Use(middleware.RequireAuth())
	{
		protected.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		protected.DELETE("/todo/:owner_id", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			userRole, _ := c.Get("role")
			ownerID := c.Param("owner_id")

			if userRole != "admin" && fmt.Sprintf("%v", userID) != ownerID {
				c.Status(http.StatusForbidden)
				return
			}
			c.Status(http.StatusOK)
		})
	}
	return r
}

func TestJWT_AccessWithoutToken(t *testing.T) {
	r := setupMiddlewareTestRouter()
	req, _ := http.NewRequest("GET", "/protected/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for no token, got %d", w.Code)
	}
}

func TestJWT_AccessWithInvalidToken(t *testing.T) {
	r := setupMiddlewareTestRouter()
	req, _ := http.NewRequest("GET", "/protected/test", nil)
	req.Header.Set("Authorization", "Bearer InvalidTokenStringHere")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", w.Code)
	}
}

func TestJWT_AccessWithValidToken(t *testing.T) {
	r := setupMiddlewareTestRouter()
	token, _ := utils.GenerateToken(1, "mariam", "user")

	req, _ := http.NewRequest("GET", "/protected/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for valid token, got %d", w.Code)
	}
}

func TestRole_UserCannotDeleteOtherUsersTodo(t *testing.T) {
	r := setupMiddlewareTestRouter()
	token, _ := utils.GenerateToken(1, "mariam", "user")

	req, _ := http.NewRequest("DELETE", "/protected/todo/2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden, got %d", w.Code)
	}
}

func TestRole_AdminCanDeleteAnyTodo(t *testing.T) {
	r := setupMiddlewareTestRouter()
	token, _ := utils.GenerateToken(99, "admin_user", "admin")

	req, _ := http.NewRequest("DELETE", "/protected/todo/2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for Admin deleting any todo, got %d", w.Code)
	}
}
