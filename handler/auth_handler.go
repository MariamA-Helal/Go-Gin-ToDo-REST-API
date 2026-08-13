package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"example/ToDo/models"
	"example/ToDo/repository"
	"example/ToDo/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Repo repository.UserRepository
}

func NewAuthHandler(repo repository.UserRepository) *AuthHandler {
	return &AuthHandler{Repo: repo}
}

// Signup godoc
// @Summary      Register a new user
// @Description  Creates a new user account with default 'user' role
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input    body      LoginInput  true  "Signup Credentials"
// @Success      201      {object}  TokenResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if strings.TrimSpace(input.Username) == "" || strings.TrimSpace(input.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password cannot be empty or just spaces"})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username:      input.Username,
		Password:      hashedPassword,
		Role:          "user",
		SecretKey:     fmt.Sprintf("KEY-%d", time.Now().UnixNano()),
		UpgradeStatus: "none",
	}

	if err := h.Repo.CreateUser(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

// Login godoc
// Login godoc
// @Summary      User Login
// @Description  Authenticates a user and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input    body      LoginInput  true  "Login Credentials"
// @Success      200      {object}  TokenResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if strings.TrimSpace(input.Username) == "" || strings.TrimSpace(input.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password cannot be empty or just spaces"})
		return
	}

	user, err := h.Repo.GetUserByUsername(input.Username)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// RequestUpgrade godoc
// @Summary      Request Role Upgrade
// @Description  Submits a request to the Master Admin to upgrade the user's role to admin
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  map[string]string
// @Router       /upgrade/request [post]
func (h *AuthHandler) RequestUpgrade(c *gin.Context) {
	userID, _ := c.Get("user_id")
	repoImpl, _ := h.Repo.(*repository.UserRepositoryImpl)

	repoImpl.DB.Model(&models.User{}).Where("id = ?", userID.(uint)).Update("upgrade_status", "pending")
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Upgrade request sent to Master Admin. Please wait for approval."})
}

// ApproveUpgrade godoc
// @Summary      Approve Role Upgrade (Admin Only)
// @Description  Approves a pending upgrade request for a specific user by their ID
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int  true  "Target User ID"
// @Success      200      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Router       /upgrade/approve/{id} [patch]
func (h *AuthHandler) ApproveUpgrade(c *gin.Context) {
	userRole, _ := c.Get("role")
	if userRole.(string) != "admin" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Only Master Admins can approve requests"})
		return
	}

	targetUserID := c.Param("id")
	repoImpl, _ := h.Repo.(*repository.UserRepositoryImpl)

	repoImpl.DB.Model(&models.User{}).Where("id = ?", targetUserID).Update("upgrade_status", "approved")
	c.IndentedJSON(http.StatusOK, gin.H{"message": "User upgrade request approved successfully"})
}

// GetMySecretKey godoc
// @Summary      Get Upgrade Secret Key
// @Description  Retrieves the unique secret key for the authenticated user if their upgrade is approved
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /upgrade/secret-key [get]
func (h *AuthHandler) GetMySecretKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	repoImpl, ok := h.Repo.(*repository.UserRepositoryImpl)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to access repository"})
		return
	}

	var user models.User
	if err := repoImpl.DB.Where("id = ?", userID.(uint)).First(&user).Error; err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.UpgradeStatus != "approved" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Your upgrade request is either pending or not submitted"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"message":    "Here is your unique secret key. Use it to upgrade your account.",
		"secret_key": user.SecretKey,
	})
}

// UpgradeToAdmin godoc
// @Summary      Complete Upgrade to Admin
// @Description  Upgrades the authenticated user to admin role using their unique secret key
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input    body      SecretKeyInput  true  "Secret Key JSON"
// @Success      200      {object}  TokenResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /upgrade/apply [post]
func (h *AuthHandler) UpgradeToAdmin(c *gin.Context) {
	var req struct {
		SecretKey string `json:"secret_key" binding:"required"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Missing secret key"})
		return
	}

	userID, _ := c.Get("user_id")
	repoImpl, _ := h.Repo.(*repository.UserRepositoryImpl)

	var user models.User
	repoImpl.DB.First(&user, userID.(uint))

	if req.SecretKey != user.SecretKey {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid unique secret key"})
		return
	}

	repoImpl.DB.Model(&models.User{}).Where("id = ?", userID.(uint)).Updates(map[string]interface{}{
		"role":           "admin",
		"upgrade_status": "done",
	})

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Congratulations! You have been successfully upgraded to Admin."})
}

// --- Swagger DTOs (Data Transfer Objects) ---

// LoginInput defines the body for login and signup requests
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse defines the success response containing the JWT token
type TokenResponse struct {
	Token string `json:"token"`
}

// ErrorResponse defines the standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SecretKeyInput defines the body for the upgrade route
type SecretKeyInput struct {
	SecretKey string `json:"secret_key" binding:"required"`
}
