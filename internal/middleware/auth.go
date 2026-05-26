package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"pulseroad/internal/pkg/response"
)

const CurrentUserIDKey = "current_user_id"

type TokenParser interface {
	ParseToken(token string) (uint, error)
}

func AuthRequired(parser TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		userID, err := parser.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		c.Set(CurrentUserIDKey, userID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(CurrentUserIDKey)
	if !ok {
		return 0, false
	}

	userID, ok := value.(uint)
	return userID, ok && userID != 0
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
