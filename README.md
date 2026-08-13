# 🚀 Go Todo REST API - Backend Engineering Project

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![Gin](https://img.shields.io/badge/Gin-Framework-00ADD8?style=for-the-badge&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=for-the-badge&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Containerized-2496ED?style=for-the-badge&logo=docker)
![Swagger](https://img.shields.io/badge/Swagger-API_Docs-85EA2D?style=for-the-badge&logo=swagger)

A robust, highly scalable, and fully containerized RESTful API built with **Go (Golang)**, **Gin Framework**, **GORM**, and **PostgreSQL**. This project follows **Clean Architecture** principles and is driven by **Test-Driven Development (TDD)** to ensure reliability and maintainability.

Developed as part of the Backend Engineering Internship, focusing on secure system design, database architecture, and containerization.

---

## 🌟 Key Features

### Security & Authentication
* **JWT-Based Authentication:** Secure endpoints using JSON Web Tokens.
* **Advanced Role-Based Access Control (RBAC):** Strict isolation between `user` and `admin` roles. Users can only manage their own tasks, while Admins have system-wide privileges.
* **Dynamic Role Upgrade Workflow (Bonus):** Automated secure secret key generation. Users can request to become admins, which requires explicit approval from a Master Admin.

### Core Business Logic
* **Full CRUD Operations:** Comprehensive management of Todo items.
* **Granular Updates (Bonus):** Supports `PATCH` requests for updating specific fields (e.g., status) without overwriting the entire object.
* **Filtering & Searching:** Advanced endpoints to fetch Todos by category, status, and search terms.
* **Batch Operations:** Added endpoints to delete all Todos, and update/delete Todos by specific categories.

### DevOps & Architecture
* **Clean Architecture:** Strict separation of concerns (Models, Repositories, Handlers, Routers).
* **Dockerized Environment:** Multi-stage Docker builds and `docker-compose` orchestration for zero-configuration setups.
* **Test-Driven Development (TDD):** 115/115 passing unit tests heavily validating handlers, mock repositories, and auth logic.
* **API Documentation:** Interactive Swagger UI integration for easy testing and exploration.

---

## 🛠️ Tech Stack & Tooling

* **Language:** Go (Golang)
* **Framework:** Gin (`github.com/gin-gonic/gin`)
* **Database & ORM:** PostgreSQL & GORM (`gorm.io/gorm`)
* **Authentication:** JWT (`golang-jwt/jwt`) & bcrypt for password hashing
* **Containerization:** Docker & Docker Compose
* **Testing:** Go testing package, Testify (Mocks)
* **Documentation:** Swaggo (Swagger UI)

---

## 📁 Architecture & Project Structure

The codebase is organized into isolated layers to ensure decoupling and ease of testing:

```text
ToDo/
├── cmd/api/                 # 🚀 Entry point of the application
├── database/                # 🐘 PostgreSQL connection, connection pooling, & migrations
├── handler/                 # 🌐 HTTP layer, parsing requests, and returning JSON responses
│   ├── auth_handler.go      
│   ├── todo_handler.go      
│   └── *_test.go            # 🧪 Comprehensive Unit Tests (115 passing cases)
├── middleware/              # 🛡️ Security layer (JWT validation, RBAC enforcement)
├── models/                  # 📦 Data structures and GORM schema definitions
├── repository/              # 🗄️ Data access layer (Database CRUD logic)
├── router/                  # 🛤️ API route definitions and group management
├── scripts/                 # 🛠️ Utility scripts for testing DB connections
├── utils/                   # ⚙️ Helper functions (Token generation, Hashing)
├── Dockerfile               # 🐳 Multi-stage build instructions for the Go API
├── docker-compose.yml       # 🐙 Orchestration for API and Database containers
└── .env                     # 🔐 Environment variables (Ignored in Git)
```

---

## 🚀 Getting Started

### Prerequisites
* Docker & Docker Desktop installed.
* Git installed.

### Option 1: Run via Docker Hub (Fastest)
You can pull and run the pre-built image directly from Docker Hub without needing the source code:
```bash
docker pull mariamamr286/todo-api:latest
docker run -p 8080:8080 mariamamr286/todo-api:latest
```

### Option 2: Build & Run from Source (Development)
1. **Clone the repository:**
   ```bash
   git clone [Your-GitHub-Repository-Link]
   cd ToDo
   ```

2. **Setup Environment Variables:**
   Create a `.env` file in the root directory and add the following configurations:
   ```env
   # Database Configuration
   DB_HOST=db
   DB_USER=postgres
   DB_PASSWORD=0000
   DB_NAME=todo_db
   DB_PORT=5432
   
   # JWT Secret
   JWT_SECRET=your_super_secret_key
   ```

3. **Start the application using Docker Compose:**
   ```bash
   docker-compose up --build
   ```
   *The database will initialize automatically, and GORM will handle auto-migrations. A Master Admin is seeded automatically on startup.*

---

## 📡 API Documentation & Endpoints

### 📖 Swagger Documentation
| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| `GET`  | `/swagger/*any` | Interactive API Documentation (Swagger UI) | Public |

### 🔐 Authentication & Roles
| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| `POST` | `/signup` | Register a new user | Public |
| `POST` | `/login` | Authenticate and receive JWT | Public |
| `POST` | `/user/request-upgrade` | Request admin privileges | User |
| `GET`  | `/user/my-secret-key` | View personal generated secret key | User |
| `PUT`  | `/master/approve-upgrade/:id`| Approve a user's admin request | Admin |
| `POST` | `/user/upgrade` | Finalize upgrade using approved key | User |

### 📝 Todo Management
| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| `POST` | `/todos` | Create a new Todo | User / Admin |
| `GET`  | `/todos` | Get all Todos | User / Admin |
| `GET`  | `/todos/:id` | Get a specific Todo by ID | User / Admin |
| `GET`  | `/todos/category/:category` | Get Todos filtered by category | User / Admin |
| `GET`  | `/todos/status/:status` | Get Todos filtered by completion status | User / Admin |
| `GET`  | `/todos/search` | Search Todos based on queries | User / Admin |
| `PUT`  | `/todos/:id` | Update an entire Todo item | User / Admin |
| `PUT`  | `/todos/category/:category`| Update all Todos within a specific category | User / Admin |
| `PATCH`| `/todos/:id/status` | Partially update completion status | User / Admin |
| `DELETE`| `/todos/:id` | Delete a specific Todo | User / Admin |
| `DELETE`| `/todos` | Delete all Todos | User / Admin |
| `DELETE`| `/todos/category/:category`| Delete all Todos within a specific category | User / Admin |

*(Note: Protected API calls require the `Authorization` header formatted as: `Bearer <token>`)*

---

## 🧪 Testing (TDD)
The project boasts a robust testing suite resolving handler decoupling from GORM. 
To run the full suite of unit tests:
```bash
go test ./handler/... -v
```
**Current Status:** All 115/115 tests passing.

---

## 👤 Author
**Mariam Amr Ibrahim Helal**
* GitHub: [Insert Your GitHub Link Here]
* DockerHub: [https://hub.docker.com/r/mariamamr286/todo-api](https://hub.docker.com/r/mariamamr286/todo-api)