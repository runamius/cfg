package test

import (
	"avito/iternal/models"
	"avito/iternal/repository/service"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthService_DummyLogin_Admin(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "test-secret")

	token, err := authService.DummyLogin("admin")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_DummyLogin_User(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "test-secret")

	token, err := authService.DummyLogin("user")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_DummyLogin_InvalidRole(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "test-secret")

	token, err := authService.DummyLogin("invalid")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, models.ErrInvalidInput))
	assert.Empty(t, token)
}

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "test-secret")

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	role := "user"

	userRepo.On("Create", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == email && u.Role == role
	})).Return(nil)

	user, err := authService.Register(ctx, email, password, role)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, role, user.Role)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "test-secret")

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	role := "user"

	userRepo.On("Create", ctx, mock.Anything).Return(models.ErrAlreadyExists)

	user, err := authService.Register(ctx, email, password, role)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, models.ErrInvalidInput))
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "test-secret")

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	existingUser := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "user",
	}

	userRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	token, err := authService.Login(ctx, email, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "test-secret")

	ctx := context.Background()
	email := "test@example.com"
	wrongPassword := "wrongpassword"

	existingUser := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // hash of "password123"
		Role:         "user",
	}

	userRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	token, err := authService.Login(ctx, email, wrongPassword)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, models.ErrInvalidCredentials))
	assert.Empty(t, token)
	userRepo.AssertExpectations(t)
}

func TestAuthService_ParseToken_Valid(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "test-secret")

	tokenString, _ := authService.DummyLogin("admin")

	userID, role, err := authService.ParseToken(tokenString)

	assert.NoError(t, err)
	assert.NotEmpty(t, userID)
	assert.Equal(t, "admin", role)
}

func TestAuthService_ParseToken_Invalid(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "test-secret")

	userID, role, err := authService.ParseToken("invalid-token")

	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Empty(t, role)
}

func TestNewAuthService(t *testing.T) {
	userRepo := new(MockUserRepo)
	authService := service.NewAuthService(userRepo, "my-secret-key")

	assert.NotNil(t, authService)
}
