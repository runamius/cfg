package service

import (
	"avito/iternal/models"
	"avito/iternal/repository"
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  repository.UserRepo
	jwtSecret []byte
}

func (service *AuthService) generateToken(userID, role string) (string, error) {
	m := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(5 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, m)
	signed, err := token.SignedString(service.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("generateToken: %w", err)
	}
	return signed, nil
}

func (service *AuthService) ParseToken(tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return service.jwtSecret, nil
	})

	if err != nil {
		return "", "", fmt.Errorf("Token Parsing: %w", err)
	}

	m, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid token claims")
	}

	userID, ok := m["user_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("missing user_id claim")
	}

	role, ok := m["role"].(string)
	if !ok {
		return "", "", fmt.Errorf("missing role claim")
	}

	return userID, role, nil
}

func NewAuthService(userRepo repository.UserRepo, jwtsecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtsecret),
	}
}

func (service *AuthService) DummyLogin(role string) (string, error) {
	var userID string
	if role == "admin" {
		userID = "00000000-0000-0000-0000-000000000001"
	} else if role == "user" {
		userID = "00000000-0000-0000-0000-000000000002"
	} else {
		return "", models.ErrInvalidInput
	}
	return service.generateToken(userID, role)

}

func (service *AuthService) Register(ctx context.Context, email, password, role string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Register: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}
	err = service.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("Register create: %w", err)
	}
	return user, nil
}

func (service *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := service.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", models.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", models.ErrAlreadyExists
	}

	return service.generateToken(user.ID.String(), user.Role)
}
