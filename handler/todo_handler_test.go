package handler

import (
	"net/http"
	"testing"
	"time"

	"example/ToDo/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ==========================================
// 3. TABLE-DRIVEN TESTS (The Core)
// ==========================================

// ---- Test: GET /todos (Get All) ----
func TestGetTodos(t *testing.T) {
	tests := []struct {
		name           string
		query          string // Query params like ?limit=20&page=2
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		{
			name:  "Good Case: Default limit and page (No query params)",
			query: "",
			mockSetup: func(m *MockTodoRepository) {
				// Handler defaults to limit=10, offset=0
				m.On("GetTodos", 10, 0).Return([]models.Todo{{ID: 1, Title: "Task 1"}})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Good Case: Custom limit and page",
			query: "?limit=5&page=3",
			mockSetup: func(m *MockTodoRepository) {
				// page 3 with limit 5 means offset is (3-1)*5 = 10
				m.On("GetTodos", 5, 10).Return([]models.Todo{{ID: 11, Title: "Task 11"}})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Edge Case: Limit exceeds 100 (Clamping)",
			query: "?limit=500&page=1",
			mockSetup: func(m *MockTodoRepository) {
				// Handler should force limit down to 100
				m.On("GetTodos", 100, 0).Return([]models.Todo{})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Bad Case: Invalid characters in query params",
			query: "?limit=abc&page=xyz",
			mockSetup: func(m *MockTodoRepository) {
				// Handler should ignore garbage and fallback to defaults (10 and 1)
				m.On("GetTodos", 10, 0).Return([]models.Todo{})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Edge Case: Page is way out of bounds",
			query: "?limit=10&page=9999",
			mockSetup: func(m *MockTodoRepository) {
				// offset = (9999-1)*10 = 99980
				// If page is empty, DB just returns an empty array, which is correct behavior.
				m.On("GetTodos", 10, 99980).Return([]models.Todo{})
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo, 1, "admin")

			// Append the query string to the URL
			w := performRequest(router, "GET", "/todos"+tc.query, nil)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: GET /todos/:id ----
func TestGetTodoByID(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string // Passed as a string to allow testing invalid characters
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		{
			name:   "Good Case: Valid ID exists",
			todoID: "1",
			mockSetup: func(m *MockTodoRepository) {
				// Mock the database returning a task for ID 1
				m.On("GetTodoByID", 1).Return(models.Todo{ID: 1, Title: "Learn Go Testing"}, nil)
			},
			expectedStatus: http.StatusOK, // Expected 200 OK
		},
		{
			name:   "Edge Case: ID Not Found or Deleted",
			todoID: "999", // A very large number or a previously deleted ID
			mockSetup: func(m *MockTodoRepository) {
				// Mock the database returning a "record not found" error
				m.On("GetTodoByID", 999).Return(models.Todo{}, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound, // Expected 404 Not Found
		},
		{
			name:   "Bad Case: Invalid Characters in ID",
			todoID: "ABC", // User sends characters instead of an integer
			mockSetup: func(m *MockTodoRepository) {
				// No mock setup needed; handler should reject before calling the DB
			},
			expectedStatus: http.StatusBadRequest, // Expected 400 Bad Request
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Setup the mock repository
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)

			// 2. Initialize the router with the mock repo
			router := setupTestEngine(mockRepo, 1, "admin")

			// 3. Execute the request by appending the ID to the URL
			w := performRequest(router, "GET", "/todos/"+tc.todoID, nil)

			// 4. Assert the expected status code and verify mock expectations
			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: GET /todos/category/:category ----
func TestGetTodosByCategory(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		{
			name:     "Good Case: Valid Category (1 to 3 words)",
			category: "backend work",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodosByCategory", "backend work").Return([]models.Todo{{ID: 1, Category: "backend work"}})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Bad Case: Category is int/numbers",
			category: "12345",
			mockSetup: func(m *MockTodoRepository) {
				// Handler should reject numbers as category
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Edge Case: Category > 3 words",
			category: "my very long category name",
			mockSetup: func(m *MockTodoRepository) {
				// Handler should reject this before calling DB
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Bad Case: Category is empty",
			category: " ", // Empty space
			mockSetup: func(m *MockTodoRepository) {
				// Handler should reject empty strings
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Bad Case: Category is a single character",
			category: "a", // Only one letter
			mockSetup: func(m *MockTodoRepository) {
				// Handler should enforce minimum length
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo, 1, "admin")

			w := performRequest(router, "GET", "/todos/category/"+tc.category, nil)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: GET /todos/status/:status ----
func TestGetTodosByStatus(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		{
			name:   "Good Case: Status is true (completed)",
			status: "true",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodosByStatus", "true").Return([]models.Todo{{ID: 1, Title: "Done Task"}})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Good Case: Status is false (not completed)",
			status: "false",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodosByStatus", "false").Return([]models.Todo{{ID: 2, Title: "Pending Task"}})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Bad Case: Status is int",
			status: "123",
			mockSetup: func(m *MockTodoRepository) {
				// Should fail boolean validation in handler
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Bad Case: Status is random text",
			status: "done",
			mockSetup: func(m *MockTodoRepository) {
				// Should fail boolean validation in handler
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo, 1, "admin")

			w := performRequest(router, "GET", "/todos/status/"+tc.status, nil)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: GET /todos/search ----
func TestGetTodosBySearch(t *testing.T) {
	tests := []struct {
		name           string
		searchQuery    string
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		{
			name:        "Good Case: Valid Search Text",
			searchQuery: "?q=api",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodosBySearch", "api").Return([]models.Todo{{ID: 1, Title: "build api"}})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Bad Case: Search is empty",
			searchQuery: "?q=",
			mockSetup: func(m *MockTodoRepository) {
				// Handler should reject empty search queries
			},
			expectedStatus: http.StatusBadRequest, // Or 404 depending on your handler logic
		},
		{
			name:        "Edge Case: Search >= 10 words",
			searchQuery: "?q=this is a very long search query that exceeds the ten words limit",
			mockSetup: func(m *MockTodoRepository) {
				// Handler should count words and reject
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo, 1, "admin")

			// Notice we append the query directly to the URL
			w := performRequest(router, "GET", "/todos/search"+tc.searchQuery, nil)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: POST /todos ----
func TestCreateTodo(t *testing.T) {
	futureDate := time.Now().Add(24 * time.Hour).UTC()
	pastDate := time.Now().Add(-24 * time.Hour).UTC()

	tests := []struct {
		name           string
		reqBody        interface{}
		userID         uint
		role           string
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		// 1. RBAC & Good Cases
		{
			name:    "Good Case: Normal User Creates Own Task",
			reqBody: map[string]interface{}{"title": "Learn Go", "category": "Study", "priority": "High", "due_date": futureDate},
			userID:  1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("CountUserTodos", uint(1)).Return(int64(0))
				m.On("CreateTodo", mock.AnythingOfType("*models.Todo")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "Good Case: Admin Assigns Task to Another User",
			reqBody: map[string]interface{}{"title": "Fix Bug", "category": "Work", "priority": "High", "target_username": "seif"},
			userID:  99, role: "admin",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetUserIDByUsername", "seif").Return(uint(2), nil)
				m.On("CountUserTodos", uint(2)).Return(int64(0))
				m.On("CreateTodo", mock.AnythingOfType("*models.Todo")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "Bad Case: Normal User Tries to Assign Task",
			reqBody: map[string]interface{}{"title": "Sneaky Task", "category": "Work", "priority": "High", "target_username": "reham"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusForbidden,
		},

		// 2. Original Validations & Edge Cases (Preserved)
		{
			name:    "Bad Case: Missing Required Title",
			reqBody: map[string]interface{}{"category": "Study", "priority": "Low"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Bad Case: Missing Required Category",
			reqBody: map[string]interface{}{"title": "Task", "priority": "Low"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Bad Case: Missing Required Priority",
			reqBody: map[string]interface{}{"title": "Valid Title", "category": "Work"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Edge Case: Title is a single character",
			reqBody: map[string]interface{}{"title": "A", "category": "Study", "priority": "Low"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Edge Case: Title >= 10 words",
			reqBody: map[string]interface{}{"title": "one two three four five six seven eight nine ten", "category": "Study", "priority": "Low"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Edge Case: Title is only numbers",
			reqBody: map[string]interface{}{"title": "12345", "category": "Study", "priority": "Low"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Edge Case: Category is a single character",
			reqBody: map[string]interface{}{"title": "Valid", "category": "S", "priority": "Low"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Edge Case: Category > 3 words",
			reqBody: map[string]interface{}{"title": "Valid", "category": "my very long category", "priority": "Low"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Edge Case: Category is only numbers",
			reqBody: map[string]interface{}{"title": "Valid", "category": "999", "priority": "Low"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Bad Case: Invalid Priority",
			reqBody: map[string]interface{}{"title": "Valid", "category": "Work", "priority": "Urgent"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Edge Case: Past Due Date",
			reqBody: map[string]interface{}{"title": "Time Travel", "category": "Personal", "priority": "Medium", "due_date": pastDate},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "DB Error: Internal Server Error",
			reqBody: map[string]interface{}{"title": "Save Me", "category": "Work", "priority": "Low"},
			userID:  1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("CountUserTodos", uint(1)).Return(int64(0))
				m.On("CreateTodo", mock.AnythingOfType("*models.Todo")).Return(gorm.ErrInvalidDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo, tc.userID, tc.role)
			w := performRequest(router, "POST", "/todos", tc.reqBody)
			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: PUT /todos/:id ----
func TestEditTodo(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		reqBody        interface{}
		userID         uint   // Dynamic Auth
		role           string // Dynamic Auth
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		// 1. RBAC & Security Cases
		{
			name:    "Good Case: User Edits Own Task",
			todoID:  "1",
			reqBody: map[string]interface{}{"title": "Updated Title", "category": "Work", "priority": "High", "completed": true},
			userID:  1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 1).Return(models.Todo{ID: 1, UserID: 1, Completed: false}, nil)
				m.On("EditTodo", mock.AnythingOfType("*models.Todo")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Good Case: Admin Edits Someone Else's Task (Title only, no status change)",
			todoID:  "2",
			reqBody: map[string]interface{}{"title": "Admin Fix", "category": "Work", "priority": "High", "completed": false},
			userID:  99, role: "admin",
			mockSetup: func(m *MockTodoRepository) {
				// Task belongs to user 1, but Admin is user 99
				m.On("GetTodoByID", 2).Return(models.Todo{ID: 2, UserID: 1, Completed: false}, nil)
				m.On("EditTodo", mock.AnythingOfType("*models.Todo")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Bad Case: Admin Tries to Change Status of Someone Else's Task",
			todoID:  "2",
			reqBody: map[string]interface{}{"title": "Admin Fix", "category": "Work", "priority": "High", "completed": true}, // Trying to mark as true
			userID:  99, role: "admin",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 2).Return(models.Todo{ID: 2, UserID: 1, Completed: false}, nil)
				// Should stop here and return 403, DB update shouldn't be called
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:    "Bad Case: User Tries to Edit Someone Else's Task",
			todoID:  "2",
			reqBody: map[string]interface{}{"title": "Hacked", "category": "Work", "priority": "High"},
			userID:  2, role: "user", // User 2 trying to hack User 1's task
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 2).Return(models.Todo{ID: 2, UserID: 1}, nil)
			},
			expectedStatus: http.StatusForbidden,
		},

		// 2. Original Edge Cases & Validations (Preserved)
		{
			name:    "Edge Case: ID Not Found",
			todoID:  "999",
			reqBody: map[string]interface{}{"title": "Updated Title", "category": "Work", "priority": "High"},
			userID:  1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 999).Return(models.Todo{}, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "Bad Case: Invalid Input Data (Empty Title)",
			todoID:  "1",
			reqBody: map[string]interface{}{"category": "Work"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Bad Case: Invalid Characters in ID",
			todoID:  "ABC",
			reqBody: map[string]interface{}{"title": "Updated Title", "category": "Work", "priority": "High"},
			userID:  1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo, tc.userID, tc.role)
			w := performRequest(router, "PUT", "/todos/"+tc.todoID, tc.reqBody)
			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: PATCH /todos/:id/status ----
func TestUpdateTodoStatus(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		reqBody        interface{}
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		{
			name:    "Good Case: Update Status Successfully",
			todoID:  "1",
			reqBody: map[string]interface{}{"completed": true},
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 1).Return(models.Todo{ID: 1, UserID: 1, Completed: false}, nil)
				m.On("UpdateTodoStatus", mock.AnythingOfType("*models.Todo")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Good Case: Revert Status to False",
			todoID:  "1",
			reqBody: map[string]interface{}{"completed": false},
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 1).Return(models.Todo{ID: 1, UserID: 1, Completed: true}, nil)
				m.On("UpdateTodoStatus", mock.AnythingOfType("*models.Todo")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Edge Case: Update Non-existent ID",
			todoID:  "999",
			reqBody: map[string]interface{}{"completed": true},
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 999).Return(models.Todo{}, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "Edge Case: Update Already Deleted Todo",
			todoID:  "2",
			reqBody: map[string]interface{}{"completed": true},
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 2).Return(models.Todo{}, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Bad Case: Invalid Characters in ID",
			todoID:         "ABC",
			reqBody:        map[string]interface{}{"completed": true},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Case: Invalid Status Value (String instead of Bool)",
			todoID:         "1",
			reqBody:        map[string]interface{}{"completed": "yes"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Case: Empty Body",
			todoID:         "1",
			reqBody:        map[string]interface{}{},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo, 1, "admin")
			w := performRequest(router, "PATCH", "/todos/"+tc.todoID+"/status", tc.reqBody)
			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: DELETE /todos/:id ----
func TestDeleteTodo(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		userID         uint   // Dynamic Auth
		role           string // Dynamic Auth
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		// 1. RBAC & Security Cases
		{
			name:   "Good Case: User Deletes Own Task",
			todoID: "1",
			userID: 1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 1).Return(models.Todo{ID: 1, UserID: 1}, nil)
				m.On("DeleteTodo", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Good Case: Admin Deletes Someone Else's Task",
			todoID: "2",
			userID: 99, role: "admin",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 2).Return(models.Todo{ID: 2, UserID: 1}, nil)
				m.On("DeleteTodo", 2).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Bad Case: User Tries to Delete Someone Else's Task",
			todoID: "2",
			userID: 2, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 2).Return(models.Todo{ID: 2, UserID: 1}, nil)
			},
			expectedStatus: http.StatusForbidden,
		},

		// 2. Original Edge Cases & Validations (Preserved)
		{
			name:   "Edge Case: Delete Non-existent ID",
			todoID: "999",
			userID: 1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 999).Return(models.Todo{}, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Edge Case: Delete Already Deleted Todo",
			todoID: "3",
			userID: 1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 3).Return(models.Todo{}, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Bad Case: Invalid Characters in ID",
			todoID: "ABC",
			userID: 1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)

			router := setupTestEngine(mockRepo, tc.userID, tc.role)

			w := performRequest(router, "DELETE", "/todos/"+tc.todoID, nil)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: DELETE /todos ----
func TestDeleteAllTodos(t *testing.T) {
	tests := []struct {
		name           string
		queryString    string // To test target_username=...
		userID         uint   // Dynamic User ID
		role           string // "admin" or "user"
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		// 1. Admin Cases
		{
			name:        "Admin Case: Delete All Todos Globally",
			queryString: "",
			userID:      99,
			role:        "admin",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteAllTodosGlobal").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Admin Case: Delete Personal Todos Only",
			queryString: "?scope=personal",
			userID:      99,
			role:        "admin",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteUserTodos", uint(99)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Admin Case: Delete Specific User Todos",
			queryString: "?target_username=seif",
			userID:      99,
			role:        "admin",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetUserIDByUsername", "seif").Return(uint(2), nil)
				m.On("DeleteUserTodos", uint(2)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},

		// 2. Normal User Cases
		{
			name:        "User Case: Delete Own Todos Successfully",
			queryString: "",
			userID:      1,
			role:        "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteUserTodos", uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Bad Case (User): Tries to use Admin Filters",
			queryString: "?scope=personal",
			userID:      1,
			role:        "user",
			mockSetup: func(m *MockTodoRepository) {
				// No DB calls expected due to RBAC block
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "Bad Case (Admin): Target Username Not Found",
			queryString: "?target_username=ghost",
			userID:      99,
			role:        "admin",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetUserIDByUsername", "ghost").Return(uint(0), gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)

			router := setupTestEngine(mockRepo, tc.userID, tc.role)

			w := performRequest(router, "DELETE", "/todos"+tc.queryString, nil)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: PUT/todos/category/:category/status ----
func TestUpdateTodosByCategory(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		reqBody        interface{}
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		// 1. Good Cases
		{
			name:     "Good Case: Update Category Status to True",
			category: "Work Tasks",
			reqBody:  map[string]interface{}{"completed": true},
			mockSetup: func(m *MockTodoRepository) {
				m.On("UpdateTodosByCategory", uint(1), "Work Tasks", true).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Good Case: Update Category Status to False (Revert)",
			category: "Study",
			reqBody:  map[string]interface{}{"completed": false},
			mockSetup: func(m *MockTodoRepository) {
				m.On("UpdateTodosByCategory", uint(1), "Study", false).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},

		// 2. Edge Cases (Category Validations based on your notebook)
		{
			name:     "Edge Case: Category Not Found",
			category: "Unknown",
			reqBody:  map[string]interface{}{"completed": true},
			mockSetup: func(m *MockTodoRepository) {
				m.On("UpdateTodosByCategory", uint(1), "Unknown", true).Return(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Bad Case: Category is a single character",
			category:       "A",
			reqBody:        map[string]interface{}{"completed": true},
			mockSetup:      func(m *MockTodoRepository) {}, // DB won't be called
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Case: Category > 3 words",
			category:       "this is a very long category",
			reqBody:        map[string]interface{}{"completed": true},
			mockSetup:      func(m *MockTodoRepository) {}, // DB won't be called
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Case: Category is only numbers",
			category:       "12345",
			reqBody:        map[string]interface{}{"completed": true},
			mockSetup:      func(m *MockTodoRepository) {}, // DB won't be called
			expectedStatus: http.StatusBadRequest,
		},

		// 3. Body Validations
		{
			name:           "Bad Case: Invalid Status Value (String instead of Bool)",
			category:       "Work",
			reqBody:        map[string]interface{}{"completed": "done"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Case: Empty Body",
			category:       "Work",
			reqBody:        map[string]interface{}{},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo, 1, "admin")

			w := performRequest(router, "PUT", "/todos/category/"+tc.category, tc.reqBody)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ---- Test: DELETE /todos/category/:category ----
func TestDeleteTodosByCategory(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		queryString    string // Dynamic query string
		userID         uint
		role           string
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		// 1. RBAC & Admin Scopes
		{
			name:        "Admin Case: Valid Category Deletion (Global)",
			category:    "Personal",
			queryString: "",
			userID:      99, role: "admin",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteCategoryGlobal", "Personal").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Admin Case: Delete Category for Specific User",
			category:    "Work",
			queryString: "?target_username=reham",
			userID:      99, role: "admin",
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetUserIDByUsername", "reham").Return(uint(2), nil)
				m.On("DeleteCategoryForUser", uint(2), "Work").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "User Case: Delete Own Category Successfully",
			category:    "Study",
			queryString: "",
			userID:      1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteCategoryForUser", uint(1), "Study").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Bad Case: User Tries to Use Admin Filters",
			category:    "Study",
			queryString: "?scope=personal",
			userID:      1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusForbidden,
		},

		// 2. Original Edge Cases & Validations (Preserved)
		{
			name:        "Edge Case: Delete Non-existent Category",
			category:    "GhostCategory",
			queryString: "",
			userID:      1, role: "user",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteCategoryForUser", uint(1), "GhostCategory").Return(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "Bad Case: Category is a single character",
			category:    "X",
			queryString: "",
			userID:      1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Bad Case: Category > 3 words",
			category:    "one two three four",
			queryString: "",
			userID:      1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Bad Case: Category is only numbers",
			category:    "999",
			queryString: "",
			userID:      1, role: "user",
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo, tc.userID, tc.role)
			w := performRequest(router, "DELETE", "/todos/category/"+tc.category+tc.queryString, nil)
			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
