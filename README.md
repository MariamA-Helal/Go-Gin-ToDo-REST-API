# Go Todo REST API 

A robust RESTful API built with **Go (Golang)**, **Gin Framework**, **GORM ORM**, and **PostgreSQL**. This project follows a clean, modular architecture (Models, Repositories, Handlers, Routers). 
This project is developed as part of the backend engineering internship at EBE.

## 📌 Features
- **Full CRUD Operations:** Create, Read, Update, and Delete Todo items.
- **Authentication & Authorization (JWT):** Secure user registration (Signup), login, and token-based protection using JWT.
- **Role-Based Access Control (RBAC):** Distinction between regular users and admins, ensuring users can only modify their own todos while admins have full access.
- **Secure Dynamic Upgrade Workflow (Bonus):** Automated unique secret key generation for users, requiring Master Admin approval before account elevation.
- **Partial Updates (Bonus):** Includes a `PATCH` endpoint to update only the completion status.
- **Test-Driven Development (TDD):** Comprehensive unit tests covering authentication flows, JWT validation, and role permissions.
- **Robust Error Handling:** Validates inputs (e.g., empty titles, invalid IDs, duplicate usernames) and returns appropriate HTTP status codes (200, 201, 400, 401, 403, 404).

## 🛠️ Tech Stack
- **Language:** Go (Golang)
- **Framework:** Gin (`github.com/gin-gonic/gin`)
- **Database & ORM:** PostgreSQL & GORM
- **Authentication:** JWT (`golang-jwt/jwt`)
- **API Testing:** Postman

## 📁 Project Structure
```text
ToDo/
├── cmd/
│   └── api/
│       └── main.go                  # Application entry point
├── database/
│   └── postgres.go                  # Database connection & auto-migration
├── handler/
│   ├── auth_handler.go              # Authentication & upgrade handlers
│   ├── auth_handler_test.go         # Auth unit tests
│   ├── helpers_test.go              # Test helper utilities
│   ├── mock_repository_test.go      # Mock repositories for testing
│   ├── todo_auth_test.go            # JWT & RBAC unit tests
│   ├── todo_handler.go              # Todo HTTP request handlers
│   ├── todo_handler_interface.go    # Todo handler interfaces
│   └── todo_handler_test.go         # Todo unit tests
├── middleware/
│   └── auth_middleware.go           # JWT authentication & role checking middleware
├── models/
│   ├── todo.go                      # Todo database model
│   └── user.go                      # User database model & role schema
├── repository/
│   ├── todo_repo.go                 # Todo database CRUD operations
│   └── user_repo.go                 # User database operations
├── router/
│   └── router.go                    # API routing configurations & route groups
├── scripts/
│   ├── gorm-test.go                 # GORM testing script
│   └── raw-sql-test.go              # Raw SQL testing script
├── utils/
│   └── token.go                     # JWT generation & password hashing utilities
├── .gitignore
├── go.mod
├── go.sum
├── README.md
└── ToDo.postman_collection.json     # Postman API testing collection