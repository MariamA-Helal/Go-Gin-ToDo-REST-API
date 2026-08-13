package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"example/ToDo/models"
	"example/ToDo/utils"

	"github.com/gin-gonic/gin"
)

// Mock Repo for Auth
type mockUserRepository struct {
	users map[string]*models.User
}

func (m *mockUserRepository) CreateUser(user *models.User) error {
	if _, exists := m.users[user.Username]; exists {
		return errors.New("username already exists")
	}
	user.ID = uint(len(m.users) + 1)
	m.users[user.Username] = user
	return nil
}

func (m *mockUserRepository) GetUserByUsername(username string) (*models.User, error) {
	if user, exists := m.users[username]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) UpgradeUserRole(userID uint, role string) error {
	return nil
}

func setupAuthTestRouter() (*gin.Engine, *mockUserRepository) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	mockRepo := &mockUserRepository{users: make(map[string]*models.User)}
	authHandler := NewAuthHandler(mockRepo)
	r.POST("/signup", authHandler.Signup)
	r.POST("/login", authHandler.Login)
	return r, mockRepo
}

func TestSignup_Success(t *testing.T) {
	r, _ := setupAuthTestRouter()
	body := []byte(`{"username": "mariam", "password": "password123"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", w.Code)
	}
}

func TestSignup_DuplicateUsername(t *testing.T) {
	r, mockRepo := setupAuthTestRouter()
	mockRepo.users["mariam"] = &models.User{Username: "mariam"}

	body := []byte(`{"username": "mariam", "password": "newpassword"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for duplicate, got %d", w.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	r, mockRepo := setupAuthTestRouter()
	hashedPassword, _ := utils.HashPassword("password123")
	mockRepo.users["mariam"] = &models.User{Username: "mariam", Password: hashedPassword, Role: "user"}

	body := []byte(`{"username": "mariam", "password": "password123"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	r, mockRepo := setupAuthTestRouter()
	hashedPassword, _ := utils.HashPassword("password123")
	mockRepo.users["mariam"] = &models.User{Username: "mariam", Password: hashedPassword, Role: "user"}

	body := []byte(`{"username": "mariam", "password": "wrongpassword"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", w.Code)
	}
}

// ---- Missing Signup Edge Cases ----

func TestSignup_Failures(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        []byte
		expectedStatus int
	}{
		{
			name:           "Missing Password",
			reqBody:        []byte(`{"username": "mariam"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Username",
			reqBody:        []byte(`{"password": "password123"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty Fields",
			reqBody:        []byte(`{"username": "", "password": ""}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Malformed JSON",
			reqBody:        []byte(`{"username": "mariam", "password": "password123"`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty Request Body (No JSON)",
			reqBody:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Whitespace Only Fields",
			reqBody:        []byte(`{"username": "   ", "password": "   "}`),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := setupAuthTestRouter()
			req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(tc.reqBody))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected %d for %s, got %d", tc.expectedStatus, tc.name, w.Code)
			}
		})
	}
}

// ---- Missing Login Edge Cases ----

func TestLogin_Failures(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        []byte
		expectedStatus int
	}{
		// 1.Validations
		{
			name:           "Missing Password",
			reqBody:        []byte(`{"username": "mariam"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Username",
			reqBody:        []byte(`{"password": "password123"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty Fields",
			reqBody:        []byte(`{"username": "", "password": ""}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Malformed JSON",
			reqBody:        []byte(`{"username": "mariam", "password": "password123"`),
			expectedStatus: http.StatusBadRequest,
		},

		// 2.(Security & Authentication)
		{
			name:           "Unregistered User (Not Found)",
			reqBody:        []byte(`{"username": "ghost_user", "password": "password123"}`),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Wrong Password",
			reqBody:        []byte(`{"username": "mariam", "password": "wrongpassword"}`),
			expectedStatus: http.StatusUnauthorized,
		},

		{
			name:           "Empty Request Body (No JSON)",
			reqBody:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Whitespace Only Fields",
			reqBody:        []byte(`{"username": "   ", "password": "   "}`),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, mockRepo := setupAuthTestRouter()

			hashedPassword, _ := utils.HashPassword("password123")
			mockRepo.users["mariam"] = &models.User{
				Username: "mariam",
				Password: hashedPassword,
				Role:     "user",
			}

			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(tc.reqBody))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected %d for %s, got %d", tc.expectedStatus, tc.name, w.Code)
			}
		})
	}
}
