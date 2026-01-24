package middleware

import (
	"net/http"
	"strings"
	"log"
	"bhojanalya/internal/auth"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format, use 'Bearer <token>'"})
			c.Abort()
			return
		}

		userID, email, role, err := auth.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			c.Abort()
			return
		}

		log.Printf(
			"[AUTH DEBUG] userID=%v (type=%T), email=%s, role=%s",
			userID,
			userID,
			email,
			role,
	)


		// Attach user info to request context
		c.Set("userID", userID)
		c.Set("userEmail", email)
		c.Set("userRole", role)
		c.Next()
	}
}
