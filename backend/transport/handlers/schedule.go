package handlers

import (
	"avito/iternal/models"
	"avito/iternal/repository/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	scheduleSvc *service.ScheduleService
}

func NewScheduleHandler(scheduleSvc *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleSvc: scheduleSvc}
}

type createScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek" binding:"required"`
	StartTime  string `json:"startTime" binding:"required"`
	EndTime    string `json:"endTime" binding:"required"`
}

func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "invalid room ID"))
		return
	}

	var req createScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	schedule, err := h.scheduleSvc.CreateSchedule(c.Request.Context(), roomID, req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorResponse("ROOM_NOT_FOUND", "room not found"))
			return
		}
		if errors.Is(err, models.ErrScheduleExists) {
			c.JSON(http.StatusConflict, errorResponse("SCHEDULE_EXISTS", "schedule already exists for this room"))
			return
		}
		if errors.Is(err, models.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "invalid schedule parameters"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"schedule": schedule})
}
