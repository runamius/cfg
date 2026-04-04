package handlers

import (
	"avito/iternal/repository/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomSvc *service.RoomService
}

func NewRoomHandler(roomSvc *service.RoomService) *RoomHandler {
	return &RoomHandler{roomSvc: roomSvc}
}

type createRoomRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	room, err := h.roomSvc.CreateRoom(c.Request.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"room": room})
}

func (h *RoomHandler) GetRooms(c *gin.Context) {
	rooms, err := h.roomSvc.GetRooms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}
