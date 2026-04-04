package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type InfoHandler struct{}

func (h *InfoHandler) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
