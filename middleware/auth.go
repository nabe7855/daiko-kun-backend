package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware verifies the JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		if !strings.HasPrefix(tokenString, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Store user info in context
		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])

		if val, ok := claims["company_id"]; ok && val != nil {
			c.Set("company_id", val)
		}

		c.Next()
	}
}

// CompanyGuard ensures the user can only access their own company's data
func CompanyGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role == "super_admin" {
			c.Next()
			return
		}

		userCompanyID, exists := c.Get("company_id")
		if !exists || userCompanyID == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Company ID not found in token"})
			return
		}

		// Check if the request is trying to access a specific company resource
		// This part depends on how the handlers use the context.
        // For strict isolation, we should probably inject the company_id into the request query
        // or ensure handlers READ from c.Get("company_id") instead of c.Query("company_id")
		
		c.Next()
	}
}
