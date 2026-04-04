package service

import (
	"avito/iternal/models"
	"avito/iternal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RoomService struct {
	roomRepo repository.RoomRepo
}

func NewRoomService(roomRepo repository.RoomRepo) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
	}
}

func (service *RoomService) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (*models.Room, error) {
	if name == "" {
		return nil, models.ErrInvalidInput
	}
	room := &models.Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
		CreatedAt:   time.Now().UTC(),
	}
	err := service.roomRepo.Create(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("Service create room: %w", err)
	}
	return room, nil
}

func (service *RoomService) GetRoom(ctx context.Context, id uuid.UUID) (*models.Room, error) {
	return service.roomRepo.GetByID(ctx, id)
}

func (service *RoomService) GetRooms(ctx context.Context) ([]*models.Room, error) {
	return service.roomRepo.GetAll(ctx)
}
