package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"example/ToDo/utils"

	"github.com/gin-gonic/gin"
)

// 1. RequireAuth
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := utils.VerifyToken(tokenString)
		if err != nil {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		fmt.Println("DECODED USER ID FROM TOKEN:", claims["id"])
		c.Set("user_id", uint(int(claims["id"].(float64))))
		c.Set("username", claims["username"].(string))
		c.Set("role", claims["role"].(string))

		c.Next()
	}
}

// 2. RequireRole:
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")

		if !exists || userRole.(string) != requiredRole {
			c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Forbidden: You don't have enough permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
