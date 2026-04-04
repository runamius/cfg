package repository

import (
	"avito/iternal/models"
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type RoomRepo interface {
	Create(ctx context.Context, room *models.Room) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Room, error)
	GetAll(ctx context.Context) ([]*models.Room, error)
}

type ScheduleRepo interface {
	Create(ctx context.Context, shedule *models.Schedule) error
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*models.Schedule, error)
}

type SlotsRepo interface {
	CreateSlots(ctx context.Context, slots []*models.Slot) error
	GetByID(ctx context.Context, jd uuid.UUID) (*models.Slot, error)
	GetByRoomAndDate(ctx context.Context, roomID uuid.UUID, data time.Time) ([]*models.Slot, error)
}

type BookingRepo interface {
	Create(ctx context.Context, booking *models.Booking) error
	Cancel(ctx context.Context, bookingID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Booking, error)
	GetByUser(ctx context.Context, userID uuid.UUID) ([]*models.Booking, error)
	GetAll(ctx context.Context, page int, pageSize int) ([]*models.Booking, int, error)
}
