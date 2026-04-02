-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT,
    role TEXT NOT NULL CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    capacity INTEGER CHECK (capacity IS NULL OR capacity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    days_of_week SMALLINT[] NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT schedules_room_id_unique UNIQUE (room_id),
    CONSTRAINT schedules_days_of_week_not_empty CHECK (cardinality(days_of_week) > 0),
    CONSTRAINT schedules_days_of_week_valid CHECK (
        days_of_week <@ ARRAY[1, 2, 3, 4, 5, 6, 7]::SMALLINT[]
    ),
    CONSTRAINT schedules_time_range_valid CHECK (start_time < end_time)
);

CREATE TABLE slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    schedule_id UUID REFERENCES schedules(id) ON DELETE SET NULL,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT slots_time_range_valid CHECK (start_at < end_at),
    CONSTRAINT slots_unique_room_window UNIQUE (room_id, start_at, end_at)
);

CREATE INDEX slots_room_id_start_at_idx ON slots (room_id, start_at);
CREATE INDEX slots_schedule_id_start_at_idx ON slots (schedule_id, start_at);

CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_id UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status TEXT NOT NULL CHECK (status IN ('active', 'cancelled')),
    conference_link TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX bookings_active_slot_unique_idx
    ON bookings (slot_id)
    WHERE status = 'active';

CREATE INDEX bookings_user_id_created_at_idx ON bookings (user_id, created_at DESC);
CREATE INDEX bookings_status_created_at_idx ON bookings (status, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS bookings_status_created_at_idx;
DROP INDEX IF EXISTS bookings_user_id_created_at_idx;
DROP INDEX IF EXISTS bookings_active_slot_unique_idx;
DROP TABLE IF EXISTS bookings;

DROP INDEX IF EXISTS slots_schedule_id_start_at_idx;
DROP INDEX IF EXISTS slots_room_id_start_at_idx;
DROP TABLE IF EXISTS slots;

DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;
