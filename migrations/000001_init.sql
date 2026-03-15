-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL 
);

CREATE TABLE IF NOT EXISTS events (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    userid      BIGINT NOT NULL,
    uuid        UUID NOT NULL UNIQUE,
    title       VARCHAR(200) NOT NULL,
    description TEXT,
    event_date  TIMESTAMP WITH TIME ZONE NOT NULL,
    available_seats INTEGER NOT NULL,
    booking_ttl INTEGER NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (userid) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS bookings (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id),
    event_id      BIGINT NOT NULL REFERENCES events(id),
    status        VARCHAR(20) NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX active_bookings_idx ON bookings (user_id, event_id) WHERE status = 'pending';
CREATE INDEX idx_bookings_pending_expires ON bookings (status, expires_at) WHERE status = 'pending';
CREATE INDEX idx_bookings_user_event ON bookings (user_id, event_id);
CREATE INDEX idx_events_userid ON events (userid);

-- +goose Down
DROP TABLE IF EXISTS bookings CASCADE;
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS users CASCADE;