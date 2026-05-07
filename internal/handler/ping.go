package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"myapp/internal/middleware"
	"myapp/internal/service"
)

type PingHandler struct {
	pings *service.PingService
}

func NewPingHandler(pings *service.PingService) *PingHandler {
	return &PingHandler{pings: pings}
}

func (h *PingHandler) Ping(c *gin.Context) {
	username, ok := middleware.UsernameFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user"})
		return
	}

	allowed, _, err := h.pings.CheckRateLimit(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit failed"})
		return
	}
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit"})
		return
	}

	lockToken, ok, err := h.pings.AcquireLock(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lock failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "ping busy"})
		return
	}
	defer func() {
		_ = h.pings.ReleaseLock(c.Request.Context(), lockToken)
	}()

	count, err := h.pings.IncrementCount(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
		return
	}

	if err := h.pings.UpdateStats(c.Request.Context(), username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "stats failed"})
		return
	}

	time.Sleep(5 * time.Second)

	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"count":   count,
	})
}

func (h *PingHandler) Top(c *gin.Context) {
	items, err := h.pings.GetTop(c.Request.Context(), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "top failed"})
		return
	}

	response := make([]gin.H, 0, len(items))
	for _, item := range items {
		member, ok := item.Member.(string)
		if !ok {
			continue
		}
		response = append(response, gin.H{
			"username": member,
			"score":    item.Score,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": response,
	})
}

func (h *PingHandler) Count(c *gin.Context) {
	count, err := h.pings.GetUniqueCount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}
