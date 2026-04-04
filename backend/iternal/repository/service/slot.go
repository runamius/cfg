package service

import (
	"avito/iternal/models"
	"avito/iternal/repository"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SlotService struct {
	slotRepo     repository.SlotsRepo
	scheduleRepo repository.ScheduleRepo
	roomRepo     repository.RoomRepo
}

func NewSlotService(slotRepo repository.SlotsRepo, scheduleRepo repository.ScheduleRepo, roomRepo repository.RoomRepo) *SlotService {
	return &SlotService{
		slotRepo:     slotRepo,
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
	}
}

// ([01]?[0-9]|2[0-3]):[0-5][0-9]
func parseTimeStr(s string) (int, int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h > 23 || h < 0 {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m > 59 || m < 0 {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}
	return h, m, nil
}
func generateSlots(roomID uuid.UUID, date time.Time, startTime, endTime string) ([]*models.Slot, error) {
	startH, startM, err := parseTimeStr(startTime)
	if err != nil {
		return nil, err
	}

	endH, endM, err := parseTimeStr(endTime)
	if err != nil {
		return nil, err
	}

	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	start := day.Add(time.Hour*time.Duration(startH) + time.Minute*time.Duration(startM))
	end := day.Add(time.Hour*time.Duration(endH) + time.Minute*time.Duration(endM))

	var slots []*models.Slot
	for start.Before(end) {
		slotEnd := start.Add(30 * time.Minute)

		if slotEnd.After(end) {
			break
		}

		slots = append(slots, &models.Slot{
			ID:        uuid.New(),
			RoomID:    roomID,
			StartTime: start,
			EndTime:   slotEnd,
		})

		start = slotEnd
	}
	return slots, nil
}

func (service *SlotService) GetFreeSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]*models.Slot, error) {
	_, err := service.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, models.ErrInvalidInput
	}

	schedule, err := service.scheduleRepo.GetByRoomID(ctx, roomID)
	if err == models.ErrNoSchedule {
		return []*models.Slot{}, nil
	}
	if err != nil {
		return nil, err
	}

	existing, err := service.slotRepo.GetByRoomAndDate(ctx, roomID, date)
	if err != nil {
		return nil, err
	}

	if len(existing) > 0 {
		return existing, nil
	}

	slots, err := generateSlots(roomID, date, schedule.StartTime, schedule.EndTime)
	if err != nil {
		return nil, err
	}

	err = service.slotRepo.CreateSlots(ctx, slots)
	if err != nil {
		return nil, fmt.Errorf("create slots: %w", err)
	}

	return slots, nil
}
