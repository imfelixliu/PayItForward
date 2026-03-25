package middleware

import (
	"net/http"
	"os"
	"strings"
	"todo-app/apperror"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get("request_id")

		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":       apperror.ErrUnauthorized.Code,
				"message":    "missing or invalid authorization header",
				"request_id": requestID,
			})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		secret := []byte(getJWTSecret())

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":       apperror.ErrUnauthorized.Code,
				"message":    "invalid or expired token",
				"request_id": requestID,
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":       apperror.ErrUnauthorized.Code,
				"message":    "invalid token claims",
				"request_id": requestID,
			})
			return
		}

		userID := int(claims["user_id"].(float64))
		c.Set("user_id", userID)
		c.Next()
	}
}

func getJWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "default-secret-change-in-production"
}
