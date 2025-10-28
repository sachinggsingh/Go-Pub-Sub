package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	helper "github.com/sachinggsingh/notify/internal/helpers"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		// Expecting "Bearer <token>"
		splitToken := strings.Split(authHeader, " ")
		if len(splitToken) != 2 || splitToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		clientToken := splitToken[1]

		claims, err := helper.ValidateToken(clientToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("uid", claims.UserID)
		c.Next()
	}
}
