CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    nickname VARCHAR(64) NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'member',
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS route_books (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL REFERENCES users(id),
    title VARCHAR(128) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    visibility VARCHAR(16) NOT NULL DEFAULT 'private',
    version INT NOT NULL DEFAULT 1,
    distance_m DOUBLE PRECISION NOT NULL DEFAULT 0,
    ascent_m DOUBLE PRECISION NOT NULL DEFAULT 0,
    geometry_hash VARCHAR(64) NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS waypoints (
    id BIGSERIAL PRIMARY KEY,
    route_id BIGINT NOT NULL REFERENCES route_books(id) ON DELETE CASCADE,
    seq INT NOT NULL,
    type VARCHAR(16) NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lon DOUBLE PRECISION NOT NULL,
    elevation DOUBLE PRECISION,
    radius_m DOUBLE PRECISION,
    polygon JSONB,
    risk_weight INT NOT NULL DEFAULT 1,
    note TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_waypoints_route ON waypoints(route_id, seq);

CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    leader_id BIGINT NOT NULL REFERENCES users(id),
    route_id BIGINT REFERENCES route_books(id),
    name VARCHAR(64) NOT NULL,
    invite_code VARCHAR(8) NOT NULL UNIQUE,
    status VARCHAR(16) NOT NULL DEFAULT 'open',
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS team_members (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    role VARCHAR(16) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP NOT NULL,
    state VARCHAR(16) NOT NULL DEFAULT 'offline',
    UNIQUE (team_id, user_id)
);

CREATE TABLE IF NOT EXISTS trips (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id),
    route_id BIGINT REFERENCES route_books(id),
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS track_points (
    id BIGSERIAL PRIMARY KEY,
    trip_id BIGINT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    member_id BIGINT NOT NULL REFERENCES users(id),
    lat DOUBLE PRECISION NOT NULL,
    lon DOUBLE PRECISION NOT NULL,
    elevation DOUBLE PRECISION,
    accuracy DOUBLE PRECISION,
    speed DOUBLE PRECISION,
    recorded_at TIMESTAMP NOT NULL,
    source VARCHAR(16) NOT NULL DEFAULT 'live',
    is_noise BOOLEAN NOT NULL DEFAULT FALSE,
    fingerprint VARCHAR(64) NOT NULL,
    UNIQUE (fingerprint)
);
CREATE INDEX IF NOT EXISTS idx_track_trip_member_time ON track_points(trip_id, member_id, recorded_at);

CREATE TABLE IF NOT EXISTS track_segments (
    id BIGSERIAL PRIMARY KEY,
    trip_id BIGINT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    member_id BIGINT NOT NULL,
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP NOT NULL,
    distance_m DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_gap BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS track_batches (
    id BIGSERIAL PRIMARY KEY,
    trip_id BIGINT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    member_id BIGINT NOT NULL,
    batch_hash VARCHAR(64) NOT NULL,
    point_count INT NOT NULL,
    accepted INT NOT NULL,
    rejected INT NOT NULL,
    processed_at TIMESTAMP NOT NULL,
    UNIQUE (trip_id, member_id, batch_hash)
);

CREATE TABLE IF NOT EXISTS sos_events (
    id BIGSERIAL PRIMARY KEY,
    trip_id BIGINT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    member_id BIGINT NOT NULL,
    type VARCHAR(16) NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lon DOUBLE PRECISION NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'open',
    created_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS risk_reports (
    id BIGSERIAL PRIMARY KEY,
    trip_id BIGINT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    level VARCHAR(16) NOT NULL,
    score DOUBLE PRECISION NOT NULL,
    detail JSONB NOT NULL DEFAULT '{}'::jsonb,
    dispersion DOUBLE PRECISION NOT NULL DEFAULT 0,
    computed_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS elevation_cache (
    geometry_hash VARCHAR(64) PRIMARY KEY,
    profile JSONB NOT NULL,
    provider VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS notify_log (
    id BIGSERIAL PRIMARY KEY,
    channel VARCHAR(16) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL
);
