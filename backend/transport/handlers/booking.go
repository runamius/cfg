package handlers

import (
	"avito/iternal/models"
	"avito/iternal/repository/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BookingHandler struct {
	bookingSvc *service.BookingService
}

func NewBookingHandler(bookingSvc *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingSvc: bookingSvc}
}

type createBookingRequest struct {
	SlotID string `json:"slotId" binding:"required"`
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req createBookingRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "invalid slot ID"))
		return
	}

	slotID, err := uuid.Parse(req.SlotID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "invalid slot ID"))
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "invalid user ID in token"))
		return
	}

	booking, err := h.bookingSvc.CreateBooking(c.Request.Context(), slotID, userID)
	if err != nil {
		if errors.Is(err, models.ErrSlotInPast) {
			c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "cannot book a slot in the past"))
			return
		}
		if errors.Is(err, models.ErrSlotAlreadyBooked) {
			c.JSON(http.StatusConflict, errorResponse("SLOT_ALREADY_BOOKED", "slot is already booked"))
			return
		}
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorResponse("SLOT_NOT_FOUND", "slot not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"booking": booking})
}

func (h *BookingHandler) ListBookings(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}

	bookings, total, err := h.bookingSvc.ListAllBookings(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	})
}

func (h *BookingHandler) ListMyBookings(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "invalid user ID in token"))
		return
	}

	bookings, err := h.bookingSvc.ListMyBookings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingIDStr := c.Param("bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "invalid booking ID"))
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "invalid user ID in token"))
		return
	}

	booking, err := h.bookingSvc.CancelBooking(c.Request.Context(), bookingID, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorResponse("BOOKING_NOT_FOUND", "booking not found"))
			return
		}
		if errors.Is(err, models.ErrForbidden) {
			c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "you can only cancel your own bookings"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": booking})
}
