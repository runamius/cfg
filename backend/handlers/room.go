package handlers

import "avito/iternal/repository/service"

type AuthHandler struct {
	authSrv *service.AuthService
}

func NewAuthHandler(authSrv *service.AuthService) *AuthHandler {
	return &AuthHandler{authSrv: authSrv}
}

type dummyLoginRequest struct {
	Role string `json:"role" binding:"required, oneof=admin user"`
}
