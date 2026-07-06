-- +goose Up
-- +goose StatementBegin
CREATE TYPE payment_type AS ENUM ('premium');

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type payment_type NOT NULL,
    amount_stars INT NOT NULL CHECK (amount_stars > 0),
    telegram_charge_id TEXT NOT NULL,
    provider_charge_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_telegram_charge_id_unique UNIQUE (telegram_charge_id)
);

CREATE INDEX idx_payments_user_id ON payments (user_id);
CREATE INDEX idx_payments_created_at ON payments (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payments;
DROP TYPE IF EXISTS payment_type;
-- +goose StatementEnd
