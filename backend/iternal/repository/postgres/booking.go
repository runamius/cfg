package postgres

import (
	"avito/iternal/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepo struct {
	pool *pgxpool.Pool
}

func NewBookingRepo(pool *pgxpool.Pool) *BookingRepo {
	return &BookingRepo{pool: pool}
}

func (r *BookingRepo) Create(ctx context.Context, booking *models.Booking) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("BookingRepo.Create begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	lockQuery := `SELECT id FROM slots WHERE id = $1 FOR UPDATE`
	var slotID uuid.UUID
	if err := tx.QueryRow(ctx, lockQuery, booking.SlotID).Scan(&slotID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ErrNotFound
		}
		return fmt.Errorf("BookingRepo.Create lock slot: %w", err)
	}

	var startTime time.Time
	if err := tx.QueryRow(ctx, `SELECT start_time FROM slots WHERE id = $1`, booking.SlotID).Scan(&startTime); err != nil {
		return fmt.Errorf("BookingRepo.Create get slot time: %w", err)
	}
	if startTime.Before(time.Now().UTC()) {
		return models.ErrSlotInPast
	}

	insertQuery := `
		INSERT INTO bookings (id, slot_id, user_id, status, conference_link, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.Exec(ctx, insertQuery,
		booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink, booking.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.ErrSlotAlreadyBooked
		}
		return fmt.Errorf("BookingRepo.Create insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("BookingRepo.Create commit: %w", err)
	}
	return nil
}

func (r *BookingRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Booking, error) {
	query := `
		SELECT b.id, b.slot_id, b.user_id, b.status, COALESCE(b.conference_link,''), b.created_at,
		       s.id, s.room_id, s.start_time, s.end_time
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		WHERE b.id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	b, err := scanBookingWithSlot(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("BookingRepo.GetByID: %w", err)
	}
	return b, nil
}

func (r *BookingRepo) GetAll(ctx context.Context, page, pageSize int) ([]*models.Booking, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM bookings`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("BookingRepo.ListAll count: %w", err)
	}

	query := `
		SELECT b.id, b.slot_id, b.user_id, b.status, COALESCE(b.conference_link,''), b.created_at,
		       s.id, s.room_id, s.start_time, s.end_time
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		ORDER BY b.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("BookingRepo.ListAll: %w", err)
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		b, err := scanBookingWithSlotFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("BookingRepo.ListAll scan: %w", err)
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("BookingRepo.ListAll rows: %w", err)
	}
	return bookings, total, nil
}

func (r *BookingRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]*models.Booking, error) {
	query := `
		SELECT b.id, b.slot_id, b.user_id, b.status, COALESCE(b.conference_link,''), b.created_at,
		       s.id, s.room_id, s.start_time, s.end_time
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		WHERE b.user_id = $1
		  AND b.status = 'active'
		  AND s.start_time > NOW()
		ORDER BY s.start_time ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("BookingRepo.ListByUserFuture: %w", err)
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		b, err := scanBookingWithSlotFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("BookingRepo.ListByUserFuture scan: %w", err)
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("BookingRepo.ListByUserFuture rows: %w", err)
	}
	return bookings, nil
}

func (r *BookingRepo) Cancel(ctx context.Context, bookingID uuid.UUID) error {
	query := `UPDATE bookings SET status = 'cancelled' WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, bookingID)
	if err != nil {
		return fmt.Errorf("BookingRepo.Cancel: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBookingWithSlot(row rowScanner) (*models.Booking, error) {
	b := &models.Booking{Slot: &models.Slot{}}
	return b, row.Scan(
		&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt,
		&b.Slot.ID, &b.Slot.RoomID, &b.Slot.StartTime, &b.Slot.EndTime,
	)
}

func scanBookingWithSlotFromRows(rows pgx.Rows) (*models.Booking, error) {
	b := &models.Booking{Slot: &models.Slot{}}
	err := rows.Scan(
		&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt,
		&b.Slot.ID, &b.Slot.RoomID, &b.Slot.StartTime, &b.Slot.EndTime,
	)
	return b, err
}
