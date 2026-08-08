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
			router := setupTestEngine(mockRepo)

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
			router := setupTestEngine(mockRepo)

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
			router := setupTestEngine(mockRepo)

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
			router := setupTestEngine(mockRepo)

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
			router := setupTestEngine(mockRepo)

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
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		// 1. Good Paths
		{
			name:           "Good Case: Valid Full Data",
			reqBody:        models.Todo{Title: "Learn JWT", Category: "Study", Priority: "High", DueDate: &futureDate},
			mockSetup:      func(m *MockTodoRepository) { m.On("CreateTodo", mock.AnythingOfType("*models.Todo")).Return(nil) },
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Good Case: Valid Data Without Due Date",
			reqBody:        models.Todo{Title: "Learn Go", Category: "Study", Priority: "Medium"},
			mockSetup:      func(m *MockTodoRepository) { m.On("CreateTodo", mock.AnythingOfType("*models.Todo")).Return(nil) },
			expectedStatus: http.StatusCreated,
		},

		// 2. Missing Required Fields
		{
			name:           "Bad Case: Missing Required Title",
			reqBody:        models.Todo{Category: "Study", Priority: "Low"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Case: Missing Required Category",
			reqBody:        models.Todo{Title: "Task", Priority: "Low"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Bad Case: Missing Required Priority",
			reqBody: models.Todo{
				Title:    "Valid Title",
				Category: "Work",
				// Priority is missing here
			},
			mockSetup: func(m *MockTodoRepository) {
				// DB won't be called because the handler should reject the request first
			},
			expectedStatus: http.StatusBadRequest,
		},

		// 3. Edge Cases
		{
			name:           "Edge Case: Title is a single character",
			reqBody:        models.Todo{Title: "A", Category: "Study", Priority: "Low"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Edge Case: Title >= 10 words",
			reqBody:        models.Todo{Title: "one two three four five six seven eight nine ten", Category: "Study", Priority: "Low"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Edge Case: Title is only numbers",
			reqBody:        models.Todo{Title: "12345", Category: "Study", Priority: "Low"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Edge Case: Category is a single character",
			reqBody:        models.Todo{Title: "Valid Title", Category: "S", Priority: "Low"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Edge Case: Category > 3 words",
			reqBody:        models.Todo{Title: "Valid Title", Category: "my very long category", Priority: "Low"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Edge Case: Category is only numbers",
			reqBody:        models.Todo{Title: "Valid Title", Category: "999", Priority: "Low"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Case: Invalid Priority",
			reqBody:        models.Todo{Title: "Fix Bugs", Category: "Work", Priority: "Urgent"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Edge Case: Past Due Date",
			reqBody:        models.Todo{Title: "Time Travel", Category: "Personal", Priority: "Medium", DueDate: &pastDate},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Edge Case: Malformed JSON",
			reqBody:        `{ "title": "broken json" `,
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "DB Error: Internal Server Error",
			reqBody: models.Todo{Title: "Save Me", Category: "Work", Priority: "Low"},
			mockSetup: func(m *MockTodoRepository) {
				m.On("CreateTodo", mock.AnythingOfType("*models.Todo")).Return(gorm.ErrInvalidDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo)
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
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		{
			name:    "Good Case: Valid Edit",
			todoID:  "1",
			reqBody: models.Todo{Title: "Updated Title", Category: "Work", Priority: "High"},
			mockSetup: func(m *MockTodoRepository) {
				m.On("GetTodoByID", 1).Return(models.Todo{ID: 1, Title: "Old", Completed: false}, nil)
				m.On("EditTodo", mock.AnythingOfType("*models.Todo")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Edge Case: ID Not Found",
			todoID:  "999",
			reqBody: models.Todo{Title: "Updated Title", Category: "Work", Priority: "High"},
			mockSetup: func(m *MockTodoRepository) {
				// هنا الـ Get مش هيلاقيها
				m.On("GetTodoByID", 999).Return(models.Todo{}, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Bad Case: Invalid Input Data (Empty Title)",
			todoID:         "1",
			reqBody:        models.Todo{Category: "Work"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Case: Invalid Characters in ID",
			todoID:         "ABC",
			reqBody:        models.Todo{Title: "Updated Title", Category: "Work", Priority: "High"},
			mockSetup:      func(m *MockTodoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo)
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
				// Mock الـ Get
				m.On("GetTodoByID", 1).Return(models.Todo{ID: 1, Completed: false}, nil)
				m.On("UpdateTodoStatus", mock.AnythingOfType("*models.Todo")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Good Case: Revert Status to False",
			todoID:  "1",
			reqBody: map[string]interface{}{"completed": false},
			mockSetup: func(m *MockTodoRepository) {
				// Mock الـ Get
				m.On("GetTodoByID", 1).Return(models.Todo{ID: 1, Completed: true}, nil)
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
			router := setupTestEngine(mockRepo)
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
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		{
			name:   "Good Case: Valid Deletion",
			todoID: "1",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteTodo", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Edge Case: Delete Non-existent ID",
			todoID: "999",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteTodo", 999).Return(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Edge Case: Delete Already Deleted Todo",
			todoID: "2", // Assuming ID 2 was previously deleted
			mockSetup: func(m *MockTodoRepository) {
				// The database treats a deleted record as not found
				m.On("DeleteTodo", 2).Return(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Bad Case: Invalid Characters in ID",
			todoID: "ABC", // Sending string instead of int
			mockSetup: func(m *MockTodoRepository) {
				// Handler should reject this before calling the DB
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo)
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
		reqBody        interface{} // To ensure we send nil
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		{
			name:    "Good Case: Delete All Todos Successfully",
			reqBody: nil, // Ensuring the body is empty
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteAllTodos").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)

			router := setupTestEngine(mockRepo)

			// Passing exactly "/todos" with no extra parameters
			w := performRequest(router, "DELETE", "/todos", tc.reqBody)

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
				m.On("UpdateTodosByCategory", "Work Tasks", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Good Case: Update Category Status to False (Revert)",
			category: "Study",
			reqBody:  map[string]interface{}{"completed": false},
			mockSetup: func(m *MockTodoRepository) {
				m.On("UpdateTodosByCategory", "Study", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},

		// 2. Edge Cases (Category Validations based on your notebook)
		{
			name:     "Edge Case: Category Not Found",
			category: "Unknown",
			reqBody:  map[string]interface{}{"completed": true},
			mockSetup: func(m *MockTodoRepository) {
				m.On("UpdateTodosByCategory", "Unknown", mock.Anything).Return(gorm.ErrRecordNotFound)
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
			router := setupTestEngine(mockRepo)

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
		mockSetup      func(m *MockTodoRepository)
		expectedStatus int
	}{
		// 1. Good Cases
		{
			name:     "Good Case: Valid Category Deletion",
			category: "Personal",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteTodosByCategory", "Personal").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},

		// 2. Edge Cases (Not Found / Already Deleted)
		{
			name:     "Edge Case: Delete Non-existent Category (Or already deleted)",
			category: "Ghost Category",
			mockSetup: func(m *MockTodoRepository) {
				m.On("DeleteTodosByCategory", "Ghost Category").Return(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},

		// 3. Bad Paths (Validations)
		{
			name:     "Bad Case: Category is a single character",
			category: "X",
			mockSetup: func(m *MockTodoRepository) {
				// DB should not be called due to validation
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Bad Case: Category > 3 words",
			category: "one two three four",
			mockSetup: func(m *MockTodoRepository) {
				// DB should not be called due to validation
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Bad Case: Category is only numbers",
			category: "999",
			mockSetup: func(m *MockTodoRepository) {
				// DB should not be called due to validation
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTodoRepository)
			tc.mockSetup(mockRepo)
			router := setupTestEngine(mockRepo)

			// Passing nil for body since DELETE does not require one
			w := performRequest(router, "DELETE", "/todos/category/"+tc.category, nil)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
