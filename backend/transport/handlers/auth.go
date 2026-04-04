package handlers

import (
	"avito/iternal/models"
	"avito/iternal/repository/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type dummyLoginRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user"`
}

func (h *AuthHandler) DummyLogin(c *gin.Context) {
	var req dummyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	token, err := h.authSvc.DummyLogin(req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "role must be 'admin' or 'user'"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, models.ErrAlreadyExists) {
			c.JSON(http.StatusConflict, errorResponse("EMAIL_TAKEN", "email already registered"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	token, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, errorResponse("INVALID_CREDENTIALS", "invalid email or password"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
