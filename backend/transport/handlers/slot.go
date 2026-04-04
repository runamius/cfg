package handlers

import (
	"avito/iternal/models"
	"avito/iternal/repository/service"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SlotHandler struct {
	slotSvc *service.SlotService
}

func NewSlotHandler(slotsvc *service.SlotService) *SlotHandler {
	return &SlotHandler{slotSvc: slotsvc}
}

func (h *SlotHandler) ListSlots(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "invalid room ID"))
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "date query parameter is required"))
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "date must be in YYYY-MM-DD format"))
		return
	}
	date = date.UTC()

	slots, err := h.slotSvc.GetFreeSlots(c.Request.Context(), roomID, date)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorResponse("ROOM_NOT_FOUND", "room not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"slots": slots})
}
