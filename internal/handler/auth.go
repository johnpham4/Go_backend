package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"myapp/internal/service"
)

type AuthHandler struct {
	sessions *service.SessionService
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
}

func NewAuthHandler(sessions *service.SessionService) *AuthHandler {
	return &AuthHandler{sessions: sessions}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	sessionID, err := h.sessions.CreateSession(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create session failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
	})
}
