package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"myapp/internal/service"
)

const sessionHeader = "X-Session-Id"
const ctxUsernameKey = "username"

func Auth(sessions *service.SessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetHeader(sessionHeader)
		if sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing session"})
			return
		}

		username, err := sessions.GetUsername(c.Request.Context(), sessionID)
		if err == redis.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "session lookup failed"})
			return
		}

		c.Set(ctxUsernameKey, username)
		c.Next()
	}
}

func UsernameFromContext(c *gin.Context) (string, bool) {
	value, ok := c.Get(ctxUsernameKey)
	if !ok {
		return "", false
	}

	username, ok := value.(string)
	return username, ok
}
