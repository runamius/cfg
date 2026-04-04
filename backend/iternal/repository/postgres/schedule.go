package postgres

import (
	"avito/iternal/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepo struct {
	pool *pgxpool.Pool
}

func NewScheduleRepo(pool *pgxpool.Pool) *ScheduleRepo {
	return &ScheduleRepo{pool: pool}
}

func (r *ScheduleRepo) Create(ctx context.Context, schedule *models.Schedule) error {
	query := `
		INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query,
		schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.ErrScheduleExists
		}
		return fmt.Errorf("ScheduleRepo.Create: %w", err)
	}
	return nil
}

func (r *ScheduleRepo) GetByRoomID(ctx context.Context, roomID uuid.UUID) (*models.Schedule, error) {
	query := `
		SELECT id, room_id, days_of_week,
		       to_char(start_time, 'HH24:MI'),
		       to_char(end_time, 'HH24:MI')
		FROM schedules WHERE room_id = $1
	`
	row := r.pool.QueryRow(ctx, query, roomID)

	schedule := &models.Schedule{}
	err := row.Scan(&schedule.ID, &schedule.RoomID, &schedule.DaysOfWeek, &schedule.StartTime, &schedule.EndTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNoSchedule
		}
		return nil, fmt.Errorf("ScheduleRepo.GetByRoomID: %w", err)
	}
	return schedule, nil
}
