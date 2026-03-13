package http

import (
	"gin-quickstart/internal/domain"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(domain.ErrUnauthorized)
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Error(domain.ErrUnauthorized)
			c.Abort()
			return
		}
		tokenStr := parts[1]

		claims := &jwt.RegisteredClaims{}
		parsed, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, domain.ErrUnauthorized
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !parsed.Valid {
			c.Error(domain.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set("userID", claims.Subject)
		c.Next()
	}
}
