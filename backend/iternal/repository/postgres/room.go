package postgres

import (
	"avito/iternal/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepo struct {
	pool *pgxpool.Pool
}

func NewRoomRepo(pool *pgxpool.Pool) *RoomRepo {
	return &RoomRepo{pool: pool}
}

func (r *RoomRepo) Create(ctx context.Context, room *models.Room) error {
	query := `
		INSERT INTO rooms (id, name, description, capacity, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query,
		room.ID, room.Name, room.Description, room.Capacity, room.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("RoomRepo.Create: %w", err)
	}
	return nil
}

func (r *RoomRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Room, error) {
	query := `SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	room := &models.Room{}
	err := row.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("RoomRepo.GetByID: %w", err)
	}
	return room, nil
}

func (r *RoomRepo) GetAll(ctx context.Context) ([]*models.Room, error) {
	query := `SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("RoomRepo.List: %w", err)
	}
	defer rows.Close()

	var rooms []*models.Room
	for rows.Next() {
		room := &models.Room{}
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("RoomRepo.List scan: %w", err)
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("RoomRepo.List rows: %w", err)
	}
	return rooms, nil
}
