package test

import (
	"avito/iternal/models"
	"avito/iternal/repository"
	"avito/iternal/repository/service"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── Mocks ──────────────────────────────────────────────────────────────────

type MockRoomRepo struct{ mock.Mock }

func (m *MockRoomRepo) Create(ctx context.Context, room *models.Room) error {
	return m.Called(ctx, room).Error(0)
}
func (m *MockRoomRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Room), args.Error(1)
}
func (m *MockRoomRepo) GetAll(ctx context.Context) ([]*models.Room, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Room), args.Error(1)
}

var _ repository.RoomRepo = (*MockRoomRepo)(nil)

type MockScheduleRepo struct{ mock.Mock }

func (m *MockScheduleRepo) Create(ctx context.Context, s *models.Schedule) error {
	return m.Called(ctx, s).Error(0)
}
func (m *MockScheduleRepo) GetByRoomID(ctx context.Context, roomID uuid.UUID) (*models.Schedule, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Schedule), args.Error(1)
}

var _ repository.ScheduleRepo = (*MockScheduleRepo)(nil)

// ── ScheduleService ────────────────────────────────────────────────────────

func TestScheduleService_CreateSchedule_Success(t *testing.T) {
	roomRepo := new(MockRoomRepo)
	schedRepo := new(MockScheduleRepo)
	svc := service.NewScheduleService(schedRepo, roomRepo)

	ctx := context.Background()
	roomID := uuid.New()
	room := &models.Room{ID: roomID}

	roomRepo.On("GetByID", ctx, roomID).Return(room, nil)
	schedRepo.On("GetByRoomID", ctx, roomID).Return((*models.Schedule)(nil), models.ErrNoSchedule)
	schedRepo.On("Create", ctx, mock.Anything).Return(nil)

	sched, err := svc.CreateSchedule(ctx, roomID, []int{1, 2, 3}, "09:00", "11:00")

	assert.NoError(t, err)
	assert.Equal(t, roomID, sched.RoomID)
	schedRepo.AssertExpectations(t)
}

func TestScheduleService_CreateSchedule_RoomNotFound(t *testing.T) {
	roomRepo := new(MockRoomRepo)
	schedRepo := new(MockScheduleRepo)
	svc := service.NewScheduleService(schedRepo, roomRepo)

	ctx := context.Background()
	roomID := uuid.New()

	roomRepo.On("GetByID", ctx, roomID).Return((*models.Room)(nil), models.ErrNotFound)

	_, err := svc.CreateSchedule(ctx, roomID, []int{1}, "09:00", "11:00")

	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestScheduleService_CreateSchedule_AlreadyExists(t *testing.T) {
	roomRepo := new(MockRoomRepo)
	schedRepo := new(MockScheduleRepo)
	svc := service.NewScheduleService(schedRepo, roomRepo)

	ctx := context.Background()
	roomID := uuid.New()
	room := &models.Room{ID: roomID}
	existing := &models.Schedule{RoomID: roomID}

	roomRepo.On("GetByID", ctx, roomID).Return(room, nil)
	schedRepo.On("GetByRoomID", ctx, roomID).Return(existing, nil)

	_, err := svc.CreateSchedule(ctx, roomID, []int{1}, "09:00", "11:00")

	assert.ErrorIs(t, err, models.ErrScheduleExists)
}

func TestScheduleService_CreateSchedule_InvalidDays(t *testing.T) {
	roomRepo := new(MockRoomRepo)
	schedRepo := new(MockScheduleRepo)
	svc := service.NewScheduleService(schedRepo, roomRepo)

	ctx := context.Background()
	roomID := uuid.New()
	room := &models.Room{ID: roomID}

	roomRepo.On("GetByID", ctx, roomID).Return(room, nil)
	schedRepo.On("GetByRoomID", ctx, roomID).Return((*models.Schedule)(nil), models.ErrNoSchedule)

	_, err := svc.CreateSchedule(ctx, roomID, []int{0, 8}, "09:00", "11:00")

	assert.ErrorIs(t, err, models.ErrInvalidInput)
}

// ── SlotService ────────────────────────────────────────────────────────────

func TestSlotService_GetFreeSlots_NoSchedule(t *testing.T) {
	roomRepo := new(MockRoomRepo)
	schedRepo := new(MockScheduleRepo)
	slotsRepo := new(MockSlotsRepo)
	svc := service.NewSlotService(slotsRepo, schedRepo, roomRepo)

	ctx := context.Background()
	roomID := uuid.New()
	date := time.Now().UTC().AddDate(0, 0, 1)

	roomRepo.On("GetByID", ctx, roomID).Return(&models.Room{ID: roomID}, nil)
	schedRepo.On("GetByRoomID", ctx, roomID).Return((*models.Schedule)(nil), models.ErrNoSchedule)

	slots, err := svc.GetFreeSlots(ctx, roomID, date)

	assert.NoError(t, err)
	assert.Empty(t, slots)
}

func TestSlotService_GetFreeSlots_WrongDayOfWeek(t *testing.T) {
	roomRepo := new(MockRoomRepo)
	schedRepo := new(MockScheduleRepo)
	slotsRepo := new(MockSlotsRepo)
	svc := service.NewSlotService(slotsRepo, schedRepo, roomRepo)

	ctx := context.Background()
	roomID := uuid.New()
	// Find a Monday (ISO weekday 1)
	date := time.Now().UTC().AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}
	// Schedule only has Saturday (6) and Sunday (7)
	sched := &models.Schedule{RoomID: roomID, DaysOfWeek: []int{6, 7}, StartTime: "09:00", EndTime: "11:00"}

	roomRepo.On("GetByID", ctx, roomID).Return(&models.Room{ID: roomID}, nil)
	schedRepo.On("GetByRoomID", ctx, roomID).Return(sched, nil)

	slots, err := svc.GetFreeSlots(ctx, roomID, date)

	assert.NoError(t, err)
	assert.Empty(t, slots)
}

func TestSlotService_GetFreeSlots_RoomNotFound(t *testing.T) {
	roomRepo := new(MockRoomRepo)
	schedRepo := new(MockScheduleRepo)
	slotsRepo := new(MockSlotsRepo)
	svc := service.NewSlotService(slotsRepo, schedRepo, roomRepo)

	ctx := context.Background()
	roomID := uuid.New()

	roomRepo.On("GetByID", ctx, roomID).Return((*models.Room)(nil), models.ErrNotFound)

	_, err := svc.GetFreeSlots(ctx, roomID, time.Now().UTC())

	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestSlotService_GetFreeSlots_ReturnsExisting(t *testing.T) {
	roomRepo := new(MockRoomRepo)
	schedRepo := new(MockScheduleRepo)
	slotsRepo := new(MockSlotsRepo)
	svc := service.NewSlotService(slotsRepo, schedRepo, roomRepo)

	ctx := context.Background()
	roomID := uuid.New()
	// Use a Monday
	date := time.Now().UTC().AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}
	sched := &models.Schedule{RoomID: roomID, DaysOfWeek: []int{1, 2, 3, 4, 5}, StartTime: "09:00", EndTime: "10:00"}
	existingSlots := []*models.Slot{{ID: uuid.New(), RoomID: roomID}}

	roomRepo.On("GetByID", ctx, roomID).Return(&models.Room{ID: roomID}, nil)
	schedRepo.On("GetByRoomID", ctx, roomID).Return(sched, nil)
	slotsRepo.On("GetByRoomAndDate", ctx, roomID, mock.Anything).Return(existingSlots, nil)

	slots, err := svc.GetFreeSlots(ctx, roomID, date)

	assert.NoError(t, err)
	assert.Len(t, slots, 1)
}
