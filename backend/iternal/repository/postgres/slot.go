package postgres

import (
	"avito/iternal/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotRepo struct {
	pool *pgxpool.Pool
}

func NewSlotRepo(pool *pgxpool.Pool) *SlotRepo {
	return &SlotRepo{pool: pool}
}

func (r *SlotRepo) CreateSlots(ctx context.Context, slots []*models.Slot) error {
	if len(slots) == 0 {
		return nil
	}

	query := `
		INSERT INTO slots (id, room_id, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (room_id, start_time) DO NOTHING
	`
	for _, slot := range slots {
		if _, err := r.pool.Exec(ctx, query, slot.ID, slot.RoomID, slot.StartTime, slot.EndTime); err != nil {
			return fmt.Errorf("SlotRepo.UpsertSlots: %w", err)
		}
	}
	return nil
}

func (r *SlotRepo) GetByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]*models.Slot, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	query := `
		SELECT s.id, s.room_id, s.start_time, s.end_time
		FROM slots s
		WHERE s.room_id = $1
		  AND s.start_time >= $2
		  AND s.start_time < $3
		  AND NOT EXISTS (
		      SELECT 1 FROM bookings b
		      WHERE b.slot_id = s.id AND b.status = 'active'
		  )
		ORDER BY s.start_time ASC
	`
	rows, err := r.pool.Query(ctx, query, roomID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("SlotRepo.ListFreeByRoomAndDate: %w", err)
	}
	defer rows.Close()

	var result []*models.Slot
	for rows.Next() {
		slot := &models.Slot{}
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.StartTime, &slot.EndTime); err != nil {
			return nil, fmt.Errorf("SlotRepo.ListFreeByRoomAndDate scan: %w", err)
		}
		result = append(result, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("SlotRepo.ListFreeByRoomAndDate rows: %w", err)
	}
	return result, nil
}

func (r *SlotRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Slot, error) {
	query := `SELECT id, room_id, start_time, end_time FROM slots WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	slot := &models.Slot{}
	err := row.Scan(&slot.ID, &slot.RoomID, &slot.StartTime, &slot.EndTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("SlotRepo.GetByID: %w", err)
	}
	return slot, nil
}
