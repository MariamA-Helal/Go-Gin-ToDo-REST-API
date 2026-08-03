# Go Todo REST API 

A robust RESTful API built with **Go (Golang)**, **Gin Framework**, **GORM ORM**, and **PostgreSQL**. This project follows a clean, modular architecture (Models, Repositories, Handlers, Routers). 
This project is developed as part of the backend engineering internship at EBE.

## 📌 Features
- **Full CRUD Operations:** Create, Read, Update, and Delete Todo items.
- **Partial Updates (Bonus):** Includes a `PATCH` endpoint to update only the completion status.
- **In-Memory Storage:** Uses Go slices for fast, temporary data storage during server runtime.
- **Modular Architecture:** Separation of concerns by isolating business logic (Handlers/Controllers) from routing (`main.go`).
- **Robust Error Handling:** Validates inputs (e.g., empty titles, invalid IDs) and returns appropriate HTTP status codes (200, 201, 400, 404).


## 🛠️ Tech Stack
- **Language:** Go (Golang)
- **Framework:** Gin (`github.com/gin-gonic/gin`)
- **Database & ORM:** PostgreSQL & GORM
- **API Testing:** Postman

## 📁 Project Structure
```text
ToDo/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── database/
│   └── postgres.go          # Database connection & auto-migration
├── handler/
│   └── todo_handler.go      # HTTP request handlers & Gin context
├── models/
│   └── todo.go              # Database models & structs
├── repository/
│   └── todo_repo.go         # Database CRUD operations
├── router/
│   └── router.go            # API routing configurations
├── todo_postman_collection.json # Postman API testing collection
└── go.mod
```
