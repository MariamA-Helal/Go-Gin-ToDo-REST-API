package handler

import (
	"fmt"
	"net/http"
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

// 1. Signup
func (h *AuthHandler) Signup(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
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
		"id":       user.UserID,
		"username": user.Username,
		"role":     user.Role,
	})
}

// 2. Login
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := h.Repo.GetUserByUsername(input.Username)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(user.UserID, user.Username, user.Role)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"token": token})
}

// 3. User requests an upgrade
func (h *AuthHandler) RequestUpgrade(c *gin.Context) {
	userID, _ := c.Get("user_id")
	repoImpl, _ := h.Repo.(*repository.UserRepositoryImpl)

	repoImpl.DB.Model(&models.User{}).Where("id = ?", userID.(uint)).Update("upgrade_status", "pending")
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Upgrade request sent to Master Admin. Please wait for approval."})
}

// 4. Master Admin approves the upgrade request
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

// 5. User views their unique secret key (only visible if approved)
func (h *AuthHandler) GetMySecretKey(c *gin.Context) {
	userID, _ := c.Get("user_id")
	repoImpl, _ := h.Repo.(*repository.UserRepositoryImpl)

	var user models.User
	repoImpl.DB.First(&user, userID.(uint))

	if user.UpgradeStatus != "approved" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Your upgrade request is either pending or not submitted"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"message":    "Here is your unique secret key. Use it to upgrade your account.",
		"secret_key": user.SecretKey,
	})
}

// 6. User views their unique secret key (only visible if approved)
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
