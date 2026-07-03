-- +goose Up
-- +goose StatementBegin
CREATE TYPE gender AS ENUM ('male', 'female');
CREATE TYPE language AS ENUM ('ru', 'en');
CREATE TYPE nsfw_level AS ENUM ('safe', 'adult');
CREATE TYPE dialog_type AS ENUM ('ai', 'p2p');
CREATE TYPE message_role AS ENUM ('user', 'assistant', 'system');

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL,
    public_uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT users_telegram_id_unique UNIQUE (telegram_id),
    CONSTRAINT users_public_uuid_unique UNIQUE (public_uuid)
);

CREATE INDEX idx_users_telegram_id ON users (telegram_id);
CREATE INDEX idx_users_deleted_at ON users (deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE profiles (
    user_id BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    gender gender NOT NULL,
    seeking gender NOT NULL,
    age SMALLINT NOT NULL CHECK (age >= 18 AND age <= 99),
    language language NOT NULL DEFAULT 'ru'
);

CREATE TABLE premium_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_premium_subscriptions_user_expires
    ON premium_subscriptions (user_id, expires_at DESC);

CREATE TABLE personas (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    gender gender NOT NULL,
    prompt_version TEXT NOT NULL DEFAULT 'v1',
    system_prompt TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE photos (
    id BIGSERIAL PRIMARY KEY,
    persona_id BIGINT NOT NULL REFERENCES personas (id) ON DELETE CASCADE,
    tags TEXT[] NOT NULL DEFAULT '{}',
    nsfw_level nsfw_level NOT NULL DEFAULT 'safe',
    telegram_file_id TEXT NOT NULL,
    unlock_price_stars INT NOT NULL DEFAULT 0 CHECK (unlock_price_stars >= 0),
    embedding vector(1024),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT photos_telegram_file_id_unique UNIQUE (telegram_file_id)
);

CREATE INDEX idx_photos_persona_id ON photos (persona_id);
CREATE INDEX idx_photos_tags ON photos USING GIN (tags);

CREATE TABLE dialogs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type dialog_type NOT NULL,
    persona_id BIGINT REFERENCES personas (id) ON DELETE SET NULL,
    partner_user_id BIGINT REFERENCES users (id) ON DELETE SET NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    end_reason TEXT
);

CREATE INDEX idx_dialogs_user_ended ON dialogs (user_id, ended_at);

CREATE TABLE dialog_messages (
    id BIGSERIAL PRIMARY KEY,
    dialog_id BIGINT NOT NULL REFERENCES dialogs (id) ON DELETE CASCADE,
    role message_role NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dialog_messages_dialog_created
    ON dialog_messages (dialog_id, created_at);

CREATE TABLE dialog_photos_sent (
    id BIGSERIAL PRIMARY KEY,
    dialog_id BIGINT NOT NULL REFERENCES dialogs (id) ON DELETE CASCADE,
    photo_id BIGINT NOT NULL REFERENCES photos (id) ON DELETE CASCADE,
    was_blurred BOOLEAN NOT NULL DEFAULT FALSE,
    was_unlocked BOOLEAN NOT NULL DEFAULT FALSE,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT dialog_photos_sent_dialog_photo_unique UNIQUE (dialog_id, photo_id)
);

CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users (id) ON DELETE SET NULL,
    dialog_id BIGINT REFERENCES dialogs (id) ON DELETE SET NULL,
    event_type TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_type_created ON events (event_type, created_at DESC);

CREATE TABLE deletion_benefits (
    telegram_id BIGINT PRIMARY KEY,
    free_unlock_used_at TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS deletion_benefits;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS dialog_photos_sent;
DROP TABLE IF EXISTS dialog_messages;
DROP TABLE IF EXISTS dialogs;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS personas;
DROP TABLE IF EXISTS premium_subscriptions;
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS message_role;
DROP TYPE IF EXISTS dialog_type;
DROP TYPE IF EXISTS nsfw_level;
DROP TYPE IF EXISTS language;
DROP TYPE IF EXISTS gender;
-- +goose StatementEnd
