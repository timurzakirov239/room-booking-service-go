package postgres

import (
	"context"

	"room-booking-service-go/internal/repo"
)

type bookingsRepo struct {
	queries queryRunner
}

func (r *bookingsRepo) Create(ctx context.Context, params repo.CreateBookingParams) (repo.Booking, error) {
	row := r.queries.queryRow(ctx, `
		INSERT INTO bookings (slot_id, user_id, status, conference_link)
		VALUES ($1, $2, $3, $4)
		RETURNING id, slot_id, user_id, status, conference_link, created_at
	`, params.SlotID, params.UserID, params.Status, params.ConferenceLink)

	item, err := scanBooking(row)
	if err != nil {
		return repo.Booking{}, err
	}
	return item, nil
}

func (r *bookingsRepo) GetByID(ctx context.Context, id string) (repo.Booking, error) {
	row := r.queries.queryRow(ctx, `
		SELECT id, slot_id, user_id, status, conference_link, created_at
		FROM bookings
		WHERE id = $1
	`, id)

	return scanBooking(row)
}

func (r *bookingsRepo) ListByUser(ctx context.Context, params repo.ListBookingsByUserParams) ([]repo.Booking, error) {
	baseSQL := `
		SELECT id, slot_id, user_id, status, conference_link, created_at
		FROM bookings
		WHERE user_id = $1
	`

	var (
		rows rowsScanner
		err  error
	)

	if params.From != nil {
		rows, err = r.queries.query(ctx, baseSQL+` AND created_at >= $2 ORDER BY created_at ASC, id ASC`, params.UserID, params.From.UTC())
	} else {
		rows, err = r.queries.query(ctx, baseSQL+` ORDER BY created_at ASC, id ASC`, params.UserID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]repo.Booking, 0)
	for rows.Next() {
		item, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *bookingsRepo) Cancel(ctx context.Context, id string) (repo.Booking, error) {
	row := r.queries.queryRow(ctx, `
		UPDATE bookings
		SET status = 'cancelled'
		WHERE id = $1
		RETURNING id, slot_id, user_id, status, conference_link, created_at
	`, id)

	return scanBooking(row)
}
