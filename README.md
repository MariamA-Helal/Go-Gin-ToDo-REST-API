# Go Todo REST API 

A simple, fast, and modular RESTful API for managing Todo items, built with **Go** and the **Gin** framework. 
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

## 📁 Project Structure
```text
.
├── api/
│   └── todo.go       # Business logic and handler functions
├── main.go           # Server setup and route definitions
├── go.mod            # Module dependencies
└── README.md
```
