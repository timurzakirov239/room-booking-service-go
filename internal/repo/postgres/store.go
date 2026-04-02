package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"room-booking-service-go/internal/repo"
)

type Store struct {
	Users     repo.UsersRepository
	Rooms     repo.RoomsRepository
	Schedules repo.SchedulesRepository
	Slots     repo.SlotsRepository
	Bookings  repo.BookingsRepository
}

func NewStore(pool *pgxpool.Pool) Store {
	queries := queryRunner{pool: pool}

	return Store{
		Users:     &usersRepo{queries: queries},
		Rooms:     &roomsRepo{queries: queries},
		Schedules: &schedulesRepo{queries: queries},
		Slots:     &slotsRepo{queries: queries},
		Bookings:  &bookingsRepo{queries: queries},
	}
}
