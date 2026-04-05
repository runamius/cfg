package test

import (
	"avito/iternal/models"
	"avito/iternal/repository"
	"avito/iternal/repository/service"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBookingRepo struct {
	mock.Mock
}

func (m *MockBookingRepo) Create(ctx context.Context, booking *models.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingRepo) Cancel(ctx context.Context, bookingID uuid.UUID) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

func (m *MockBookingRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Booking), args.Error(1)
}

func (m *MockBookingRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]*models.Booking, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Booking), args.Error(1)
}

func (m *MockBookingRepo) GetAll(ctx context.Context, page int, pageSize int) ([]*models.Booking, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Booking), args.Int(1), args.Error(2)
}

var _ repository.BookingRepo = (*MockBookingRepo)(nil)

type MockSlotsRepo struct {
	mock.Mock
}

func (m *MockSlotsRepo) CreateSlots(ctx context.Context, slots []*models.Slot) error {
	args := m.Called(ctx, slots)
	return args.Error(0)
}

func (m *MockSlotsRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Slot, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Slot), args.Error(1)
}

func (m *MockSlotsRepo) GetByRoomAndDate(ctx context.Context, roomID uuid.UUID, data time.Time) ([]*models.Slot, error) {
	args := m.Called(ctx, roomID, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Slot), args.Error(1)
}

var _ repository.SlotsRepo = (*MockSlotsRepo)(nil)

func TestBookingService_CreateBooking_Success(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	slotID := uuid.New()
	userID := uuid.New()

	futureTime := time.Now().UTC().Add(24 * time.Hour)
	slot := &models.Slot{
		ID:        slotID,
		StartTime: futureTime,
		EndTime:   futureTime.Add(time.Hour),
	}

	slotsRepo.On("GetByID", ctx, slotID).Return(slot, nil)
	bookingRepo.On("Create", ctx, mock.MatchedBy(func(b *models.Booking) bool {
		return b.SlotID == slotID && b.UserID == userID && b.Status == "active"
	})).Return(nil)

	booking, err := bookingService.CreateBooking(ctx, slotID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, slotID, booking.SlotID)
	assert.Equal(t, userID, booking.UserID)
	assert.Equal(t, "active", booking.Status)
	assert.NotNil(t, booking.Slot)
	slotsRepo.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_CreateBooking_SlotNotFound(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	slotID := uuid.New()
	userID := uuid.New()

	slotsRepo.On("GetByID", ctx, slotID).Return((*models.Slot)(nil), models.ErrNotFound)

	booking, err := bookingService.CreateBooking(ctx, slotID, userID)

	assert.Error(t, err)
	assert.Nil(t, booking)
	slotsRepo.AssertExpectations(t)
}

func TestBookingService_CreateBooking_SlotInPast(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	slotID := uuid.New()
	userID := uuid.New()

	pastTime := time.Now().UTC().Add(-24 * time.Hour)
	slot := &models.Slot{
		ID:        slotID,
		StartTime: pastTime,
		EndTime:   pastTime.Add(time.Hour),
	}

	slotsRepo.On("GetByID", ctx, slotID).Return(slot, nil)

	booking, err := bookingService.CreateBooking(ctx, slotID, userID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, models.ErrSlotInPast))
	assert.Nil(t, booking)
	slotsRepo.AssertExpectations(t)
}

func TestBookingService_GetBooking_Success(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	bookingID := uuid.New()

	existingBooking := &models.Booking{
		ID:             bookingID,
		SlotID:         uuid.New(),
		UserID:         uuid.New(),
		Status:         "active",
		ConferenceLink: "https://meet.example.com",
	}

	bookingRepo.On("GetByID", ctx, bookingID).Return(existingBooking, nil)

	booking, err := bookingService.GetBooking(ctx, bookingID)

	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, bookingID, booking.ID)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_GetBooking_NotFound(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	bookingID := uuid.New()

	bookingRepo.On("GetByID", ctx, bookingID).Return((*models.Booking)(nil), models.ErrNotFound)

	booking, err := bookingService.GetBooking(ctx, bookingID)

	assert.Error(t, err)
	assert.Nil(t, booking)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_ListAllBookings_DefaultPagination(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()

	bookings := []*models.Booking{
		{ID: uuid.New(), Status: "active"},
		{ID: uuid.New(), Status: "active"},
	}

	bookingRepo.On("GetAll", ctx, 1, 20).Return(bookings, 2, nil)

	result, total, err := bookingService.ListAllBookings(ctx, 0, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_ListAllBookings_MaxPageSize(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()

	bookings := []*models.Booking{}

	bookingRepo.On("GetAll", ctx, 1, 100).Return(bookings, 0, nil)

	result, total, err := bookingService.ListAllBookings(ctx, 1, 200)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Equal(t, 0, total)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_ListMyBookings_Success(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	userID := uuid.New()

	bookings := []*models.Booking{
		{ID: uuid.New(), UserID: userID, Status: "active"},
		{ID: uuid.New(), UserID: userID, Status: "cancelled"},
	}

	bookingRepo.On("GetByUser", ctx, userID).Return(bookings, nil)

	result, err := bookingService.ListMyBookings(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_ListMyBookings_Empty(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	userID := uuid.New()

	bookingRepo.On("GetByUser", ctx, userID).Return([]*models.Booking{}, nil)

	result, err := bookingService.ListMyBookings(ctx, userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_CancelBooking_Success(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	bookingID := uuid.New()
	userID := uuid.New()

	existingBooking := &models.Booking{
		ID:     bookingID,
		UserID: userID,
		Status: "active",
	}

	bookingRepo.On("GetByID", ctx, bookingID).Return(existingBooking, nil)
	bookingRepo.On("Cancel", ctx, bookingID).Return(nil)

	booking, err := bookingService.CancelBooking(ctx, bookingID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, "cancelled", booking.Status)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_CancelBooking_Forbidden(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	bookingID := uuid.New()
	userID := uuid.New()
	wrongUserID := uuid.New()

	existingBooking := &models.Booking{
		ID:     bookingID,
		UserID: userID,
		Status: "active",
	}

	bookingRepo.On("GetByID", ctx, bookingID).Return(existingBooking, nil)

	booking, err := bookingService.CancelBooking(ctx, bookingID, wrongUserID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, models.ErrForbidden))
	assert.Nil(t, booking)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_CancelBooking_AlreadyCanceled(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	ctx := context.Background()
	bookingID := uuid.New()
	userID := uuid.New()

	existingBooking := &models.Booking{
		ID:     bookingID,
		UserID: userID,
		Status: "cancelled",
	}

	bookingRepo.On("GetByID", ctx, bookingID).Return(existingBooking, nil)

	booking, err := bookingService.CancelBooking(ctx, bookingID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, "cancelled", booking.Status)
	bookingRepo.AssertExpectations(t)
}

func TestNewBookingService(t *testing.T) {
	bookingRepo := new(MockBookingRepo)
	slotsRepo := new(MockSlotsRepo)
	bookingService := service.NewBookingService(bookingRepo, slotsRepo)

	assert.NotNil(t, bookingService)
}
