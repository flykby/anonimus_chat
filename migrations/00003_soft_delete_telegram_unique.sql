-- +goose Up
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT users_telegram_id_unique;
CREATE UNIQUE INDEX users_telegram_id_active_unique ON users (telegram_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_telegram_id_active_unique;
ALTER TABLE users ADD CONSTRAINT users_telegram_id_unique UNIQUE (telegram_id);
-- +goose StatementEnd
