package service

import (
	"avito/iternal/models"
	"avito/iternal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo repository.BookingRepo
	slotsRepo   repository.SlotsRepo
}

func NewBookingService(bookingRepo repository.BookingRepo, slotRepo repository.SlotsRepo) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		slotsRepo:   slotRepo,
	}
}
func (service *BookingService) CreateBooking(ctx context.Context, slotID, userID uuid.UUID) (*models.Booking, error) {
	slot, err := service.slotsRepo.GetByID(ctx, slotID)
	if err != nil {
		return nil, err
	}

	if slot.StartTime.Before(time.Now().UTC()) {
		return nil, models.ErrSlotInPast
	}

	bookingID := uuid.New()
	conferenceLink := ""

	booking := &models.Booking{
		ID:             bookingID,
		SlotID:         slotID,
		UserID:         userID,
		Status:         "active",
		ConferenceLink: conferenceLink,
		CreatedAt:      time.Now().UTC(),
	}

	err = service.bookingRepo.Create(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("BookingService.CreateBooking: %w", err)
	}

	booking.Slot = slot
	return booking, nil
}

func (service *BookingService) GetBooking(ctx context.Context, id uuid.UUID) (*models.Booking, error) {
	return service.bookingRepo.GetByID(ctx, id)
}

func (service *BookingService) ListAllBookings(ctx context.Context, page, pageSize int) ([]*models.Booking, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	if pageSize > 100 {
		pageSize = 100
	}

	bookings, total, err := service.bookingRepo.GetAll(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("BookingService.ListAllBookings: %w", err)
	}

	if bookings == nil {
		bookings = []*models.Booking{}
	}
	return bookings, total, nil
}

func (s *BookingService) ListMyBookings(ctx context.Context, userID uuid.UUID) ([]*models.Booking, error) {
	bookings, err := s.bookingRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Booking: %w", err)
	}

	if bookings == nil {
		bookings = []*models.Booking{}
	}
	return bookings, nil
}

func (service *BookingService) CancelBooking(ctx context.Context, bookingID, userID uuid.UUID) (*models.Booking, error) {
	booking, err := service.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, models.ErrForbidden
	}

	if booking.Status == "canceled" {
		return booking, nil
	}

	err = service.bookingRepo.Cancel(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	booking.Status = "canceled"
	return booking, nil
}
