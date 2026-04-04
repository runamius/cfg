package service

import (
	"avito/iternal/models"
	"avito/iternal/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ScheduleService struct {
	scheduleRepo repository.ScheduleRepo
	roomRepo     repository.RoomRepo
}

func NewScheduleService(scheduleRepo repository.ScheduleRepo, roomRepo repository.RoomRepo) *ScheduleService {
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
	}
}

func (service *ScheduleService) CreateSchedule(ctx context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*models.Schedule, error) {
	_, err := service.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("room does not exist")
	}
	_, err = service.scheduleRepo.GetByRoomID(ctx, roomID)
	if err == nil {
		return nil, models.ErrScheduleExists
	}

	for _, d := range daysOfWeek {
		if d < 1 || d > 7 || len(daysOfWeek) == 0 {
			return nil, models.ErrInvalidInput
		}
	}
	return &models.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}, nil
}

func (service *ScheduleService) GetScheduleByRoomID(ctx context.Context, roomID uuid.UUID) (*models.Schedule, error) {
	return service.scheduleRepo.GetByRoomID(ctx, roomID)
}
