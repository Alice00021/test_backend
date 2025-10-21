-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS operation_commands
(
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    operation_id INTEGER NOT NULL REFERENCES operations (id) ON DELETE CASCADE,
    command_id   INTEGER NOT NULL REFERENCES commands (id) ON DELETE CASCADE
 );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS operation_commands;
-- +goose StatementEnd

